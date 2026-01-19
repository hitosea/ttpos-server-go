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
	regenerateOrderMaterialCmd.Flags().Uint64Var(&regenerateOrderMaterialCompanyUuidFlag, "company-uuid", 0, "门店UUID")
	regenerateOrderMaterialCmd.Flags().Uint64Var(&regenerateOrderMaterialSaleOrderUuidFlag, "sale-order-uuid", 0, "订单UUID")
	regenerateOrderMaterialCmd.Flags().BoolVar(&regenerateOrderMaterialDryRunFlag, "dry-run", false, "仅预览，不实际执行")
	regenerateOrderMaterialCmd.MarkFlagRequired("company-uuid")
	regenerateOrderMaterialCmd.MarkFlagRequired("sale-order-uuid")
	rootCommand.AddCommand(regenerateOrderMaterialCmd)
}

var (
	regenerateOrderMaterialCompanyUuidFlag   uint64
	regenerateOrderMaterialSaleOrderUuidFlag uint64
	regenerateOrderMaterialDryRunFlag        bool
)

// regenerate-order-material 重新生成订单材料用料记录
var regenerateOrderMaterialCmd = &cobra.Command{
	Use:   "regenerate-order-material",
	Short: "重新生成指定订单的材料用料记录",
	Long:  `删除指定订单的旧材料记录，并重新计算生成新的材料记录`,
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
		if regenerateOrderMaterialCompanyUuidFlag == 0 {
			fmt.Printf("%s错误: 门店UUID不能为空%s\n", redColor, resetColor)
			return
		}
		if regenerateOrderMaterialSaleOrderUuidFlag == 0 {
			fmt.Printf("%s错误: 订单UUID不能为空%s\n", redColor, resetColor)
			return
		}

		// 初始化数据库管理器和服务
		dbm := database.GetDBManager(config.Database)
		settingSrv := setting.NewSrv(dbm, cache.Global)
		salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache.Global)

		// 显示操作信息
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s重新生成订单材料用料记录%s\n", blueColor, resetColor)
		fmt.Printf("%s门店UUID: %d%s\n", blueColor, regenerateOrderMaterialCompanyUuidFlag, resetColor)
		fmt.Printf("%s订单UUID: %d%s\n", blueColor, regenerateOrderMaterialSaleOrderUuidFlag, resetColor)
		if regenerateOrderMaterialDryRunFlag {
			fmt.Printf("%s模式: 预览模式（不会实际执行）%s\n", yellowColor, resetColor)
		}
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		// 预览模式：获取订单信息用于预览
		if regenerateOrderMaterialDryRunFlag {
			db := dbm.GetDB(regenerateOrderMaterialCompanyUuidFlag)
			orderRepo := repository.NewOrderRepo(db)

			// 先通过 sale_order_uuid 获取 sale_bill_uuid
			saleOrder, err := orderRepo.GetSaleBillSaleOrderRecord(regenerateOrderMaterialSaleOrderUuidFlag)
			if err != nil {
				fmt.Printf("%s错误: 获取订单信息失败: %s%s\n", redColor, err.Error(), resetColor)
				logger.Logger.Error("获取订单信息失败", zap.Uint64("saleOrderUuid", regenerateOrderMaterialSaleOrderUuidFlag), zap.Error(err))
				return
			}
			if saleOrder == nil {
				fmt.Printf("%s错误: 订单不存在%s\n", redColor, resetColor)
				return
			}

			saleBillUuid := saleOrder.SaleBillUuid

			// 获取订单完整信息（包含商品、BOM、材料关联等）
			saleBill, err := orderRepo.GetSaleBillAllInfo(saleBillUuid)
			if err != nil {
				fmt.Printf("%s错误: 获取订单完整信息失败: %s%s\n", redColor, err.Error(), resetColor)
				logger.Logger.Error("获取订单完整信息失败", zap.Uint64("saleBillUuid", saleBillUuid), zap.Error(err))
				return
			}
			if saleBill == nil {
				fmt.Printf("%s错误: 订单账单不存在%s\n", redColor, resetColor)
				return
			}

			// 获取指定的订单
			saleOrderFromBill := saleBill.GetSaleOrder(regenerateOrderMaterialSaleOrderUuidFlag)
			if saleOrderFromBill == nil {
				fmt.Printf("%s错误: 订单不存在于账单中%s\n", redColor, resetColor)
				return
			}

			// 检查订单是否完成
			if saleOrderFromBill.FinishTime == 0 {
				fmt.Printf("%s警告: 订单未完成（finish_time=0），材料用量可能不准确%s\n", yellowColor, resetColor)
			}

			// 计算材料用量（用于预览）
			materialStocks := saleOrderFromBill.GetValidSaleOrderProductMaterialList()

			fmt.Printf("%s预览模式：将执行以下操作：%s\n", yellowColor, resetColor)
			fmt.Printf("  1. 删除订单 %d 的旧材料记录\n", regenerateOrderMaterialSaleOrderUuidFlag)
			fmt.Printf("  2. 重新计算材料用量（预计 %d 种材料）\n", len(materialStocks))
			fmt.Printf("  3. 插入新材料记录（预计 %d 条）\n", len(materialStocks))
			fmt.Printf("%s预览模式结束，未实际执行操作%s\n", yellowColor, resetColor)
			return
		}

		// 确认操作
		var confirmation string
		fmt.Printf("%s警告：此操作将删除并重新生成订单的材料用料记录%s\n", redColor, resetColor)
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

		result, err := salesOutboundSummarySrv.RegenerateOrderMaterial(ctx, regenerateOrderMaterialCompanyUuidFlag, regenerateOrderMaterialSaleOrderUuidFlag)
		if err != nil {
			fmt.Printf("%s操作失败: %s%s\n", redColor, err.Error(), resetColor)
			logger.Logger.Error("重新生成订单材料记录失败", zap.Uint64("saleOrderUuid", regenerateOrderMaterialSaleOrderUuidFlag), zap.Error(err))
			return
		}

		duration := time.Since(start)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s删除记录数: %d%s\n", blueColor, result.DeletedCount, resetColor)
		fmt.Printf("%s新增记录数: %d%s\n", blueColor, result.InsertedCount, resetColor)
		fmt.Printf("%s耗时: %v (API返回: %dms)%s\n", blueColor, duration, result.DurationMs, resetColor)
	},
}
