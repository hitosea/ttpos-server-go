package command

import (
	"fmt"
	"log"
	"strings"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
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
)

func init() {
	rootCommand.AddCommand(resyncProductDataToErpCmd)
}

// 同步到ERP物品数据
var resyncProductDataToErpCmd = &cobra.Command{
	Use:   "resync-product-data-to-erp",
	Short: "run resync to erp product data",
	Long:  `run resync product data to erp, 将本地数据库的物品数据同步到ERP系统`,
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
		fmt.Printf("%s 输入要同步商品到ERP的公司UUID: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 同步商品到ERP已取消 %s\n", redColor, resetColor)
			return
		}
		if _, err := fmt.Sscanf(companyIdStr, "%d", &companyUuid); err != nil {
			fmt.Printf("%s 错误: 同步商品到ERP的公司UUID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
			return
		}

		// 初始化目标数据库连接
		targetSaas, err := database.NewMySQLConnection(config.Database, "saas")
		if err != nil {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 连接数据库失败, 检查是否正确配置了数据库信息，以及 -c 参数是否正确 %s\n", redColor, resetColor)
			return
		}
		targetSaasDB = targetSaas

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
		if !company.CompanySetting.IsHeadquarter() {
			fmt.Printf("%s 错误: 公司不是ERP总部，不允许执行 %s\n", redColor, resetColor)
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

		// 开始数据迁移
		fmt.Printf("%s 开始数据同步... %s\n", blueColor, resetColor)

		// 数据迁移
		var dbm *database.DBManager = database.GetDBManager(config.Database)
		cache := cache.Global
		localeSrv := service.NewLocaleSrv()
		settingSrv := setting.NewSrv(dbm, cache)
		translateSrv := service.NewTranslateSrv(dbm, cache)
		productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
		syncProductToDataSrv := service.NewSyncProductToErpSrv(dbm, productSrv, settingSrv)

		// 设置上下文
		ctx := ttposContext.NewContext()
		ctx.SetDB(targetDB)
		ctx.SetCompanyUuid(companyUuid)
		ctx.SetCompany(*company)
		ctx.SetCompanySetting(*company.CompanySetting)
		ctx.SetLanguage("zh")

		// 同步数据
		err = syncProductToDataSrv.Sync(ctx)
		if err != nil {
			fmt.Printf("%s 同步本地商品数据到ERP失败 : %s %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 数据同步完成 %s\n", greenColor, resetColor)
	},
}
