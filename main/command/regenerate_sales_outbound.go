package command

import (
	"context"
	"fmt"
	"log"
	"time"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/otlp"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	regenerateSalesOutboundCmd.Flags().Uint64Var(&companyUuidFlag, "company-uuid", 0, "门店UUID")
	regenerateSalesOutboundCmd.Flags().StringVar(&dateFlag, "date", "", "日期，格式：YYYY-MM-DD")
	regenerateSalesOutboundCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "仅预览，不实际执行")
	regenerateSalesOutboundCmd.MarkFlagRequired("company-uuid")
	regenerateSalesOutboundCmd.MarkFlagRequired("date")
	rootCommand.AddCommand(regenerateSalesOutboundCmd)
}

var (
	companyUuidFlag uint64
	dateFlag        string
	dryRunFlag      bool
)

// regenerate-sales-outbound 重新生成销售出库汇总记录
var regenerateSalesOutboundCmd = &cobra.Command{
	Use:   "regenerate-sales-outbound",
	Short: "重新生成指定日期的销售出库汇总记录",
	Long:  `删除指定日期的旧销售出库汇总记录，并重新生成新的汇总记录`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		config.Server.Mode = "release"

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化国际化
		i18n.Init()

		// 自定义验证规则
		validator.Init()

		// 初始化id生成器
		utils.InitIdGenerator()

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁（必须在调用 NewSystemLock 之前）
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化短信客户端
		sms.InitClient(config.SMS.APIKey, config.SMS.BaseURL, config.SMS.ProjectName)
		// 检查短信客户端配置
		if err := sms.GetSMSClient().CheckConfig(); err != nil {
			fmt.Printf("[FATAL] Failed to check SMS client config: %v\n", err)
			logger.Logger.Info("Failed to check SMS client config", zap.Error(err))
		}

		//初始化服务发现
		cloud.Init()

		// 初始化 OTLP 调用链跟踪
		if err := otlp.Init(context.Background(), config.Otlp); err != nil {
			fmt.Printf("[FATAL] Failed to initialize OpenTelemetry: %v\n", err)
			logger.Logger.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 验证参数
		if companyUuidFlag == 0 {
			fmt.Printf("%s错误: 门店UUID不能为空%s\n", redColor, resetColor)
			return
		}
		if dateFlag == "" {
			fmt.Printf("%s错误: 日期不能为空%s\n", redColor, resetColor)
			return
		}

		// 初始化数据库管理器
		dbm := database.GetDBManager(config.Database)

		// 初始化服务
		settingSrv := setting.NewSrv(dbm, cache.Global)
		salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache.Global)

		// 显示操作信息
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s重新生成销售出库汇总记录%s\n", blueColor, resetColor)
		fmt.Printf("%s门店UUID: %d%s\n", blueColor, companyUuidFlag, resetColor)
		fmt.Printf("%s日期: %s%s\n", blueColor, dateFlag, resetColor)
		if dryRunFlag {
			fmt.Printf("%s模式: 预览模式（不会实际执行）%s\n", yellowColor, resetColor)
		}
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		if dryRunFlag {
			fmt.Printf("%s预览模式：将执行以下操作：%s\n", yellowColor, resetColor)
			fmt.Printf("  1. 删除日期 %s 的所有销售出库汇总记录\n", dateFlag)
			fmt.Printf("  2. 重新生成日期 %s 的销售出库汇总记录\n", dateFlag)
			fmt.Printf("%s预览模式结束，未实际执行操作%s\n", yellowColor, resetColor)
			return
		}

		// 确认操作
		var confirmation string
		fmt.Printf("%s警告：此操作将删除并重新生成指定日期的销售出库汇总记录%s\n", redColor, resetColor)
		fmt.Printf("%s输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s操作已取消%s\n", yellowColor, resetColor)
			return
		}

		// 执行操作
		fmt.Printf("%s开始执行重新生成操作...%s\n", blueColor, resetColor)
		start := time.Now()

		// 创建空的 gin.Context（命令行环境）
		ctx := &gin.Context{}

		result, err := salesOutboundSummarySrv.RegenerateSalesOutboundSummary(ctx, companyUuidFlag, dateFlag)
		if err != nil {
			fmt.Printf("%s操作失败: %s%s\n", redColor, err.Error(), resetColor)
			return
		}

		duration := time.Since(start)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s删除记录数: %d%s\n", blueColor, result.DeletedCount, resetColor)
		fmt.Printf("%s生成记录数: %d%s\n", blueColor, result.GeneratedCount, resetColor)
		fmt.Printf("%s耗时: %v (API返回: %dms)%s\n", blueColor, duration, result.DurationMs, resetColor)
	},
}
