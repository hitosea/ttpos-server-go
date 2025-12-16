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
	regenerateSaleBillMaterialOutboundCmd.Flags().Uint64Var(&regenerateSaleBillMaterialOutboundCompanyUuidFlag, "company-uuid", 0, "门店UUID")
	regenerateSaleBillMaterialOutboundCmd.Flags().Uint64Var(&regenerateSaleBillMaterialOutboundSaleBillUuidFlag, "sale-bill-uuid", 0, "销售账单UUID")
	regenerateSaleBillMaterialOutboundCmd.Flags().BoolVar(&regenerateSaleBillMaterialOutboundDryRunFlag, "dry-run", false, "仅预览，不实际执行")
	regenerateSaleBillMaterialOutboundCmd.MarkFlagRequired("company-uuid")
	regenerateSaleBillMaterialOutboundCmd.MarkFlagRequired("sale-bill-uuid")
	rootCommand.AddCommand(regenerateSaleBillMaterialOutboundCmd)
}

var (
	regenerateSaleBillMaterialOutboundCompanyUuidFlag uint64
	regenerateSaleBillMaterialOutboundSaleBillUuidFlag uint64
	regenerateSaleBillMaterialOutboundDryRunFlag      bool
)

// regenerate-sale-bill-material-outbound 重新生成销售账单材料出库记录
var regenerateSaleBillMaterialOutboundCmd = &cobra.Command{
	Use:   "regenerate-sale-bill-material-outbound",
	Short: "重新生成指定销售账单的材料出库记录",
	Long:  `删除指定销售账单的旧材料出库记录，并重新计算生成新的材料出库记录`,
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
		if regenerateSaleBillMaterialOutboundCompanyUuidFlag == 0 {
			fmt.Printf("%s错误: 门店UUID不能为空%s\n", redColor, resetColor)
			return
		}
		if regenerateSaleBillMaterialOutboundSaleBillUuidFlag == 0 {
			fmt.Printf("%s错误: 销售账单UUID不能为空%s\n", redColor, resetColor)
			return
		}

		// 初始化数据库管理器和服务
		dbm := database.GetDBManager(config.Database)
		settingSrv := setting.NewSrv(dbm, cache.Global)
		salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache.Global)

		// 显示操作信息
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s重新生成销售账单材料出库记录%s\n", blueColor, resetColor)
		fmt.Printf("%s门店UUID: %d%s\n", blueColor, regenerateSaleBillMaterialOutboundCompanyUuidFlag, resetColor)
		fmt.Printf("%s销售账单UUID: %d%s\n", blueColor, regenerateSaleBillMaterialOutboundSaleBillUuidFlag, resetColor)
		if regenerateSaleBillMaterialOutboundDryRunFlag {
			fmt.Printf("%s模式: 预览模式（不会实际执行）%s\n", yellowColor, resetColor)
		}
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		// 预览模式：获取销售账单信息用于预览
		if regenerateSaleBillMaterialOutboundDryRunFlag {
			db := dbm.GetDB(regenerateSaleBillMaterialOutboundCompanyUuidFlag)
			orderRepo := repository.NewOrderRepo(db)
			warehouseFormRepo := repository.NewWarehouseFormRepo(db)

			// 获取销售账单完整信息
			saleBill, err := orderRepo.GetSaleBillAllInfo(regenerateSaleBillMaterialOutboundSaleBillUuidFlag)
			if err != nil {
				fmt.Printf("%s错误: 获取销售账单信息失败: %s%s\n", redColor, err.Error(), resetColor)
				logger.Logger.Error("获取销售账单信息失败", zap.Uint64("saleBillUuid", regenerateSaleBillMaterialOutboundSaleBillUuidFlag), zap.Error(err))
				return
			}
			if saleBill == nil {
				fmt.Printf("%s错误: 销售账单不存在%s\n", redColor, resetColor)
				return
			}

			// 查询原记录
			warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItemBySaleBillUuid(regenerateSaleBillMaterialOutboundSaleBillUuidFlag)
			if err != nil {
				fmt.Printf("%s错误: 查询材料出库记录失败: %s%s\n", redColor, err.Error(), resetColor)
				logger.Logger.Error("查询材料出库记录失败", zap.Uint64("saleBillUuid", regenerateSaleBillMaterialOutboundSaleBillUuidFlag), zap.Error(err))
				return
			}

			// 过滤材料出库记录
			materialItemCount := 0
			for _, item := range warehouseOutFormItems {
				if item.MaterialUuid != 0 && item.DeleteTime == 0 {
					materialItemCount++
				}
			}

			// 计算材料消耗（用于预览）
			materialStocksMap := make(map[uint64]bool)
			for _, saleOrder := range saleBill.SaleOrders {
				materialStocks := saleOrder.GetValidSaleOrderProductMaterialList()
				for _, materialStock := range materialStocks {
					materialStocksMap[materialStock.MaterialUuid] = true
				}
			}

			fmt.Printf("%s预览模式：将执行以下操作：%s\n", yellowColor, resetColor)
			fmt.Printf("  1. 删除销售账单 %d 的旧材料出库记录（预计 %d 条）\n", regenerateSaleBillMaterialOutboundSaleBillUuidFlag, materialItemCount)
			fmt.Printf("  2. 重新计算材料消耗（预计 %d 种材料）\n", len(materialStocksMap))
			fmt.Printf("  3. 创建新材料出库记录（预计 %d 条）\n", materialItemCount)
			fmt.Printf("%s预览模式结束，未实际执行操作%s\n", yellowColor, resetColor)
			return
		}

		// 确认操作
		var confirmation string
		fmt.Printf("%s警告：此操作将删除并重新生成销售账单的材料出库记录%s\n", redColor, resetColor)
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

		result, err := salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound(ctx, regenerateSaleBillMaterialOutboundCompanyUuidFlag, regenerateSaleBillMaterialOutboundSaleBillUuidFlag)
		if err != nil {
			fmt.Printf("%s操作失败: %s%s\n", redColor, err.Error(), resetColor)
			logger.Logger.Error("重新生成销售账单材料出库记录失败", zap.Uint64("saleBillUuid", regenerateSaleBillMaterialOutboundSaleBillUuidFlag), zap.Error(err))
			return
		}

		duration := time.Since(start)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s删除记录数: %d%s\n", blueColor, result.DeletedCount, resetColor)
		fmt.Printf("%s新增记录数: %d%s\n", blueColor, result.InsertedCount, resetColor)
		fmt.Printf("%s耗时: %v (API返回: %dms)%s\n", blueColor, duration, result.DurationMs, resetColor)
	},
}

