package command

import (
	"fmt"
	"log"
	"os"
	"strings"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	rootCommand.AddCommand(syncErpDataCmd)
}

// 数据迁移
var syncErpDataCmd = &cobra.Command{
	Use:   "sync-erp-data",
	Short: "run sync",
	Long:  `run sync`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 为命令行环境创建同时输出到控制台和文件的logger
		setupConsoleLogger()

		// 初始化id生成器
		utils.InitIdGenerator()

		// 目标的公司UUID
		fmt.Printf("%s 输入要同步的公司UUID: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 数据同步已取消 %s\n", redColor, resetColor)
			return
		}
		if _, err := fmt.Sscanf(companyIdStr, "%d", &companyUuid); err != nil {
			fmt.Printf("%s 错误: 公司UUID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
			return
		}

		// 初始化数据库
		companyDB, err := database.NewMySQLConnection(config.Database, fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid))
		if err != nil {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 连接数据库失败, 检查是否正确配置了数据库信息，以及 -c 参数是否正确 %s\n", redColor, resetColor)
			return
		}
		fmt.Printf("%s 连接数据库成功 %s\n", greenColor, resetColor)

		// 将公司数据库实例保存到全局变量中，供 Run 函数使用
		targetDB = companyDB

		//初始化服务发现
		cloud.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if targetDB == nil {
			fmt.Printf("%s 错误: 数据库连接无效 %s\n", redColor, resetColor)
			return
		}

		// 获取公司信息
		companyRepo := repository.NewCompanyRepo(targetDB)
		company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
		if err != nil && !strings.Contains(err.Error(), "record not found") {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 获取新数据库公司信息失败 %s\n", redColor, resetColor)
			return
		}
		if !company.IsOpenErp() {
			fmt.Printf("%s 错误: 公司未开启ERP %s\n", redColor, resetColor)
			return
		}

		// 二次确认
		fmt.Printf("%s 将要从数据库 %s 同步最新的ERP数据 %s \n", redColor, fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid), resetColor)
		fmt.Printf("%s 数据库的商家名称为：%s %s\n", redColor, company.Name, resetColor)

		// 读取用户输入
		var confirmation string
		fmt.Printf("%s 输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s 数据迁移已取消 %s\n", redColor, resetColor)
			return
		}

		// 读取用户输入
		// verificationCode := fmt.Sprintf("%05d", rand.Intn(90000)+10000) // 生成五位数字验证码
		// fmt.Printf("%s 请输入验证码 '%s' 继续，输入其他内容取消: %s", yellowColor, verificationCode, resetColor)
		// fmt.Scanln(&confirmation)
		// if confirmation != verificationCode {
		// 	fmt.Printf("%s 验证码错误，数据迁移已取消 %s\n", redColor, resetColor)
		// 	return
		// }

		// 开始数据迁移
		fmt.Printf("%s 开始数据同步... %s\n", blueColor, resetColor)

		// 数据迁移
		var dbm *database.DBManager = database.GetDBManager(config.Database)
		cache := cache.Global
		localeSrv := service.NewLocaleSrv()
		settingSrv := setting.NewSrv(dbm, cache)
		translateSrv := service.NewTranslateSrv(dbm, cache)
		supplierSrv := service.NewSupplierSrv(dbm)
		productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
		materialSrv := service.NewMaterialSrv(dbm, localeSrv, settingSrv, translateSrv)
		warehouseSrv := service.NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)
		syncSrv := service.NewSyncSrv(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv)

		// 设置上下文
		ctx := ttposContext.NewContext()
		ctx.SetDB(targetDB)
		ctx.SetCompanyUuid(companyUuid)
		ctx.SetCompany(*company)
		ctx.SetCompanySetting(*company.CompanySetting)
		ctx.SetLanguage("zh")

		// 同步数据
		_, err = syncSrv.Sync(ctx, req.SyncReq{
			IsSyncExecute: true,
		})
		if err != nil {
			fmt.Printf("%s 数据同步失败 : %s %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 数据同步完成 %s\n", greenColor, resetColor)
	},
}

// setupConsoleLogger 设置同时输出到控制台和文件的logger
func setupConsoleLogger() {
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色输出
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出核心
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 控制台编码器
		zapcore.AddSync(os.Stdout),               // 输出到标准输出
		zap.InfoLevel,                            // 日志级别
	)

	// 文件输出核心（保持原有的文件日志）
	fileCore := logger.Logger.Core()

	// 合并两个核心
	combinedCore := zapcore.NewTee(consoleCore, fileCore)

	// 替换全局logger
	logger.Logger = zap.New(combinedCore,
		zap.AddCaller(),                   // 调用文件和行号
		zap.AddStacktrace(zap.ErrorLevel), // Error级别日志才会显示堆栈信息
	)
}
