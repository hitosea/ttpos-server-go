package command

import (
	"context"
	"fmt"
	"log"
	"time"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/repository"
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
	regenerateOrderPosInvoiceCmd.Flags().Uint64Var(&regenerateOrderPosInvoiceCompanyUuidFlag, "company-uuid", 0, "门店UUID")
	regenerateOrderPosInvoiceCmd.Flags().Uint64Var(&regenerateOrderPosInvoiceSaleOrderUuidFlag, "sale-order-uuid", 0, "销售订单UUID")
	regenerateOrderPosInvoiceCmd.Flags().StringVar(&regenerateOrderPosInvoiceOpenPosEntryNameFlag, "open-pos-entry-name", "", "OpenPosEntryName")
	regenerateOrderPosInvoiceCmd.Flags().BoolVar(&regenerateOrderPosInvoiceDryRunFlag, "dry-run", false, "仅预览，不实际执行")
	regenerateOrderPosInvoiceCmd.MarkFlagRequired("company-uuid")
	regenerateOrderPosInvoiceCmd.MarkFlagRequired("sale-order-uuid")
	regenerateOrderPosInvoiceCmd.MarkFlagRequired("open-pos-entry-name")
	rootCommand.AddCommand(regenerateOrderPosInvoiceCmd)
}

var (
	regenerateOrderPosInvoiceCompanyUuidFlag      uint64
	regenerateOrderPosInvoiceSaleOrderUuidFlag     uint64
	regenerateOrderPosInvoiceOpenPosEntryNameFlag  string
	regenerateOrderPosInvoiceDryRunFlag           bool
)

// regenerate-order-pos-invoice 重新生成订单POS发票
var regenerateOrderPosInvoiceCmd = &cobra.Command{
	Use:   "regenerate-order-pos-invoice",
	Short: "重新生成指定销售订单的POS发票",
	Long:  `重新生成指定销售订单的POS发票，调用SavePosInvoice方法保存发票到ERP系统`,
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
		if regenerateOrderPosInvoiceCompanyUuidFlag == 0 {
			fmt.Printf("%s错误: 门店UUID不能为空%s\n", redColor, resetColor)
			return
		}
		if regenerateOrderPosInvoiceSaleOrderUuidFlag == 0 {
			fmt.Printf("%s错误: 销售订单UUID不能为空%s\n", redColor, resetColor)
			return
		}
		if regenerateOrderPosInvoiceOpenPosEntryNameFlag == "" {
			fmt.Printf("%s错误: OpenPosEntryName不能为空%s\n", redColor, resetColor)
			return
		}

		// 初始化数据库管理器和服务
		dbm := database.GetDBManager(config.Database)
		settingSrv := setting.NewSrv(dbm, cache.Global)
		salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache.Global)

		// 显示操作信息
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s重新生成订单POS发票%s\n", blueColor, resetColor)
		fmt.Printf("%s门店UUID: %d%s\n", blueColor, regenerateOrderPosInvoiceCompanyUuidFlag, resetColor)
		fmt.Printf("%s销售订单UUID: %d%s\n", blueColor, regenerateOrderPosInvoiceSaleOrderUuidFlag, resetColor)
		if regenerateOrderPosInvoiceDryRunFlag {
			fmt.Printf("%s模式: 预览模式（不会实际执行）%s\n", yellowColor, resetColor)
		}
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		// 预览模式：显示预览信息
		if regenerateOrderPosInvoiceDryRunFlag {
			// 获取数据库连接用于预览
			db := dbm.GetDB(regenerateOrderPosInvoiceCompanyUuidFlag)
			saleOrderRepo := repository.NewSaleOrderRepo(db)
			saleOrder, err := saleOrderRepo.GetSaleOrderByUuid(regenerateOrderPosInvoiceSaleOrderUuidFlag)
			if err != nil {
				fmt.Printf("%s错误: 获取订单信息失败: %s%s\n", redColor, err.Error(), resetColor)
				logger.Logger.Error("获取订单信息失败", zap.Uint64("saleOrderUuid", regenerateOrderPosInvoiceSaleOrderUuidFlag), zap.Error(err))
				return
			}
			if saleOrder == nil || saleOrder.Uuid == 0 {
				fmt.Printf("%s错误: 订单不存在%s\n", redColor, resetColor)
				return
			}

			// 获取公司信息用于预览
			companyRepo := repository.NewCompanyRepo(db)
			company, err := companyRepo.GetCompanyInfoByUuid(regenerateOrderPosInvoiceCompanyUuidFlag)
			if err != nil {
				fmt.Printf("%s错误: 获取公司信息失败: %s%s\n", redColor, err.Error(), resetColor)
				return
			}

			fmt.Printf("%s预览模式：将执行以下操作：%s\n", yellowColor, resetColor)
			fmt.Printf("  订单号: %s\n", saleOrder.OrderNo)
			fmt.Printf("  订单金额: %.2f\n", saleOrder.Amount)
			if company != nil && company.CompanySetting != nil {
				fmt.Printf("  ERP SiteCode: %s\n", company.CompanySetting.ErpnextSiteCode)
				fmt.Printf("  ERP CompanyAbbr: %s\n", company.CompanySetting.ErpnextCompanyAbbr)
			}
			fmt.Printf("  将调用 SavePosInvoice 方法生成发票\n")
			fmt.Printf("  将更新订单发票名称字段\n")
			fmt.Printf("%s预览模式结束，未实际执行操作%s\n", yellowColor, resetColor)
			return
		}

		// 确认操作
		var confirmation string
		fmt.Printf("%s警告：此操作将重新生成订单的POS发票%s\n", redColor, resetColor)
		fmt.Printf("%s输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s操作已取消%s\n", yellowColor, resetColor)
			return
		}

		// 执行操作
		fmt.Printf("%s开始执行重新生成发票操作...%s\n", blueColor, resetColor)
		start := time.Now()

		// 创建空的 gin.Context（命令行环境）
		ginCtx := &gin.Context{}

		// 调用服务方法
		result, err := salesOutboundSummarySrv.RegenerateOrderPosInvoice(
			ginCtx,
			regenerateOrderPosInvoiceCompanyUuidFlag,
			regenerateOrderPosInvoiceSaleOrderUuidFlag,
			regenerateOrderPosInvoiceOpenPosEntryNameFlag,
		)
		if err != nil {
			fmt.Printf("%s操作失败: %s%s\n", redColor, err.Error(), resetColor)
			logger.Logger.Error("重新生成订单POS发票失败", zap.Uint64("saleOrderUuid", regenerateOrderPosInvoiceSaleOrderUuidFlag), zap.Error(err))
			return
		}

		duration := time.Since(start)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s商品发票名称: %s%s\n", blueColor, result.ProductsInvoiceName, resetColor)
		fmt.Printf("%s材料发票名称: %s%s\n", blueColor, result.MaterialInvoiceName, resetColor)
		fmt.Printf("%s耗时: %v (API返回: %dms)%s\n", blueColor, duration, result.DurationMs, resetColor)
	},
}
