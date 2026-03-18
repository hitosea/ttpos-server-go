package tasks

import (
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

const (
	stockEntryMaxRetries = 3
	stockEntryRetryDelay = 5 * time.Minute
)

// ErpStockEntryTask 0 点 Stock Entry 合并扣减任务
// 每日执行，查询所有 ERP 商家的未出库订单，合并后提交 Stock Entry
type ErpStockEntryTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewErpStockEntryTask(dbm *database.DBManager, cache cache.Cache) *ErpStockEntryTask {
	return &ErpStockEntryTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行定时任务
func (t *ErpStockEntryTask) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("ERP Stock Entry合并扣减任务panic", zap.Any("error", err))
		}
	}()

	logger.Logger.Info("开始执行ERP Stock Entry合并扣减任务")

	saasDB := t.dbm.GetDB(constant.DefaultDB)
	if saasDB == nil {
		logger.Logger.Error("无法获取saas主库连接")
		return
	}

	// 通过 company_setting 查询所有已开启 ERP 的商家 UUID
	companySettingRepo := repository.NewCompanySettingRepo(saasDB)
	abbrUuidMap, err := companySettingRepo.GetErpnextCompanyAbbrUuidMap(
		companySettingRepo.WhereErpnextCompanyAbbrNotEmpty(),
	)
	if err != nil {
		logger.Logger.Error("查询ERP商家失败", zap.Error(err))
		return
	}
	if len(abbrUuidMap) == 0 {
		return
	}

	companyRepo := repository.NewCompanyRepo(saasDB)
	erpStockEntrySrv := service.NewErpStockEntrySrv(t.dbm)
	successCount := 0
	failCount := 0

	for _, companyUuid := range abbrUuidMap {
		company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
		if err != nil || company == nil || company.CompanySetting == nil {
			continue
		}
		if company.CompanySetting.ErpnextSiteCode == "" {
			continue
		}
		// 仅 Sales Invoice 模式需要延迟扣库存
		if !company.CompanySetting.IsErpSalesInvoiceMode() {
			continue
		}

		ctx := context.NewDefaultContext()
		ctx.SetCompanyUuid(company.Uuid)
		ctx.SetCompanySetting(*company.CompanySetting)

		var lastErr error
		for attempt := 1; attempt <= stockEntryMaxRetries; attempt++ {
			lastErr = erpStockEntrySrv.TriggerStockEntryDeduction(ctx, company.Uuid)
			if lastErr == nil {
				break
			}
			logger.Logger.Warn("Stock Entry合并扣减失败，准备重试",
				zap.Uint64("company_uuid", company.Uuid),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", stockEntryMaxRetries),
				zap.Error(lastErr),
			)
			if attempt < stockEntryMaxRetries {
				time.Sleep(stockEntryRetryDelay)
			}
		}

		if lastErr != nil {
			logger.Logger.Error("[告警] Stock Entry合并扣减最终失败，已耗尽重试次数",
				zap.Uint64("company_uuid", company.Uuid),
				zap.Int("max_retries", stockEntryMaxRetries),
				zap.Error(lastErr),
			)
			failCount++
		} else {
			successCount++
		}
	}

	logger.Logger.Info("ERP Stock Entry合并扣减任务完成",
		zap.Int("total", len(abbrUuidMap)),
		zap.Int("success", successCount),
		zap.Int("fail", failCount),
	)
}
