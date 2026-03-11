package command

import (
	stdCtx "context"
	"fmt"
	"log"

	"ttpos-server-go/app/constant"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/otlp"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	triggerStockEntryCmd.Flags().Uint64Var(&triggerStockEntryCompanyUuidFlag, "company-uuid", 0, "门店UUID")
	triggerStockEntryCmd.MarkFlagRequired("company-uuid")
	rootCommand.AddCommand(triggerStockEntryCmd)
}

var triggerStockEntryCompanyUuidFlag uint64

var triggerStockEntryCmd = &cobra.Command{
	Use:   "trigger-stock-entry",
	Short: "手动触发指定商家的 ERP Stock Entry 合并扣减任务，用于测试或修复",
	Long:  `手动触发指定商家的 ERP Stock Entry 合并扣减任务，用于测试或修复`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		config.Server.Mode = "release"

		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		i18n.Init()
		validator.Init()
		utils.InitIdGenerator()

		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		sms.InitClient(config.SMS.APIKey, config.SMS.BaseURL, config.SMS.ProjectName)
		if err := sms.GetSMSClient().CheckConfig(); err != nil {
			fmt.Printf("[FATAL] Failed to check SMS client config: %v\n", err)
			logger.Logger.Info("Failed to check SMS client config", zap.Error(err))
		}

		cloud.Init()

		if err := otlp.Init(stdCtx.Background(), config.Otlp); err != nil {
			fmt.Printf("[FATAL] Failed to initialize OpenTelemetry: %v\n", err)
			logger.Logger.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		if triggerStockEntryCompanyUuidFlag == 0 {
			fmt.Printf("%s错误: 门店UUID不能为空%s\n", redColor, resetColor)
			return
		}

		dbm := database.GetDBManager(config.Database)

		saasDB := dbm.GetDB(constant.DefaultDB)
		if saasDB == nil {
			fmt.Printf("%s错误: 无法获取saas主库连接%s\n", redColor, resetColor)
			return
		}

		companySettingRepo := repository.NewCompanySettingRepo(saasDB)
		companyRepo := repository.NewCompanyRepo(saasDB)

		company, err := companyRepo.GetCompanyInfoByUuid(triggerStockEntryCompanyUuidFlag)
		if err != nil || company == nil || company.CompanySetting == nil {
			fmt.Printf("%s错误: 找不到商家信息或商家设置%s\n", redColor, resetColor)
			return
		}

		_ = companySettingRepo

		if company.CompanySetting.ErpnextSiteCode == "" {
			fmt.Printf("%s错误: 商家未配置ERP%s\n", redColor, resetColor)
			return
		}

		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s手动触发 Stock Entry 合并扣减%s\n", blueColor, resetColor)
		fmt.Printf("%s门店UUID: %d%s\n", blueColor, triggerStockEntryCompanyUuidFlag, resetColor)
		fmt.Printf("%sERP Site: %s%s\n", blueColor, company.CompanySetting.ErpnextSiteCode, resetColor)
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		ctx := appContext.NewDefaultContext()
		ctx.SetCompanyUuid(company.Uuid)
		ctx.SetCompanySetting(*company.CompanySetting)

		erpStockEntrySrv := service.NewErpStockEntrySrv(dbm)
		if err := erpStockEntrySrv.TriggerStockEntryDeduction(ctx, company.Uuid); err != nil {
			fmt.Printf("%sStock Entry 合并扣减执行失败: %v%s\n", redColor, err, resetColor)
			logger.Logger.Error("Stock Entry合并扣减执行失败",
				zap.Uint64("company_uuid", company.Uuid),
				zap.Error(err),
			)
			return
		}

		fmt.Printf("%sStock Entry 合并扣减执行成功！%s\n", greenColor, resetColor)
	},
}
