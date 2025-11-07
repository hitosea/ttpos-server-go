package tasks

import (
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// KitchenEfficiencyAnalysisTask 后厨效率分析定时任务
// 每小时执行一次，检查门店是否到达营业结束时间，如果是则统计当天销售出库记录并写入汇总表
type KitchenEfficiencyAnalysisTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewKitchenEfficiencyAnalysisTask 创建后厨效率分析定时任务
func NewKitchenEfficiencyAnalysisTask(dbm *database.DBManager, cache cache.Cache) *KitchenEfficiencyAnalysisTask {
	return &KitchenEfficiencyAnalysisTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行定时任务
func (t *KitchenEfficiencyAnalysisTask) Execute() {
	logger.Logger.Info("开始执行后厨效率分析定时任务")

	start := time.Now()
	// 分布式锁,避免多个节点同时执行
	lock.NewSystemLock().LockUuid(lock.KitchenEfficiencyAnalysisLock)
	defer lock.NewSystemLock().UnlockUuid(lock.KitchenEfficiencyAnalysisLock)
	spend := time.Since(start)
	logger.Logger.Info("厨效率分析定时任务,分布式锁耗时: %v", zap.Duration("spend", spend))
	if spend > 1*time.Second {
		logger.Logger.Warn("厨效率分析定时任务.其他节点已经处理该任务，本节点跳过")
		return
	}

	// 获取所有门店
	companies, err := t.getAllCompanies()
	if err != nil {
		logger.Logger.Error("厨效率分析定时任务,获取门店列表失败: %v", zap.Error(err))
		return
	}

	logger.Logger.Info("厨效率分析定时任务,找到 %d 个门店,开始统计后厨效率分析", zap.Int("company_count", len(companies)))

	for _, company := range companies {
		if err := t.ProcessCompany(company); err != nil {
			logger.Logger.Error("厨效率分析定时任务,处理门店失败", zap.String("company_name", company.Name), zap.Uint64("company_uuid", company.Uuid), zap.Error(err))
			continue
		}
	}

	logger.Logger.Info("厨效率分析定时任务执行完成")
}

// getAllCompanies 获取所有门店
func (t *KitchenEfficiencyAnalysisTask) getAllCompanies() ([]*model.Company, error) {
	var companies []*model.Company
	db := t.dbm.GetDB(0)

	// 查询所有未删除且启用的门店，包含设置信息
	err := db.Model(&model.Company{}).
		Where("ttpos_company.delete_time = 0 AND ttpos_company.status = 1").
		Preload("CompanySetting").
		Find(&companies).Error

	return companies, err
}

// processCompany 处理单个门店
func (t *KitchenEfficiencyAnalysisTask) ProcessCompany(company *model.Company) error {
	statisticsSrv := service.NewStatisticsSrv()
	uploadFileSrv := service.NewUploadFileSrv(t.dbm)
	businessSrv := service.NewBusinessSrv(statisticsSrv, uploadFileSrv)
	ctx := context.NewDefaultContext()
	ctx.SetCompanySetting(*company.CompanySetting)
	ctx.SetCompanyUuid(company.Uuid)
	ctx.SetCompany(*company)
	ctx.SetDB(t.dbm.GetDB(company.Uuid))
	_, err := businessSrv.StatsKitchenEfficiencyAnalysis(ctx)
	if err != nil {
		return err
	}
	logger.Logger.Info("统计后厨效率分析成功", zap.String("company_name", company.Name), zap.Uint64("company_uuid", company.Uuid))
	return nil
}
