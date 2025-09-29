package service

import (
	"sync"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
)

// ISyncSrv同步服务接口
type ISyncSrv interface {
	Sync(context.Context) error // 同步
}

// SyncSrv同步服务结构体
type SyncSrv struct {
	dbm          *database.DBManager
	warehouseSrv IWarehouseSrv
	materialSrv  IMaterialSrv
	supplierSrv  ISupplierSrv
	productSrv   IProductSrv
}

// 全局同步任务管理器
var (
	syncTaskManager = &SyncTaskManager{
		runningTasks: sync.Map{},
	}
)

// SyncTaskManager 同步任务管理器
type SyncTaskManager struct {
	runningTasks sync.Map // key: companyUuid, value: bool
}

// tryStartTask 尝试启动任务，如果已有任务在运行则返回false
func (m *SyncTaskManager) tryStartTask(companyUuid uint64) bool {
	_, loaded := m.runningTasks.LoadOrStore(companyUuid, true)
	return !loaded // 如果之前没有值则返回true，表示成功启动任务
}

// finishTask 完成任务
func (m *SyncTaskManager) finishTask(companyUuid uint64) {
	m.runningTasks.Delete(companyUuid)
}

// getRunningCompanyUuids 获取当前正在执行同步任务的所有companyUuid
func (m *SyncTaskManager) getRunningCompanyUuids() []uint64 {
	var companyUuids []uint64
	m.runningTasks.Range(func(key, value any) bool {
		if companyUuid, ok := key.(uint64); ok {
			companyUuids = append(companyUuids, companyUuid)
		}
		return true
	})
	return companyUuids
}

// NewSyncSrv 创建新同步服务
func NewSyncSrv(dbm *database.DBManager, warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv, materialSrv IMaterialSrv) ISyncSrv {
	return NewSyncSrvImpl(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv)
}

// NewSyncSrvImpl 创建新同步服务实现
func NewSyncSrvImpl(dbm *database.DBManager, warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv, materialSrv IMaterialSrv) ISyncSrv {
	return &SyncSrv{
		dbm:          dbm,
		warehouseSrv: warehouseSrv,
		materialSrv:  materialSrv,
		supplierSrv:  supplierSrv,
		productSrv:   productSrv,
	}
}

// Sync 同步
func (s *SyncSrv) Sync(ctx context.Context) error {

	company := ctx.GetCompany()
	companyUuid := company.Uuid

	// 检查是否已有同步任务在运行
	if !syncTaskManager.tryStartTask(companyUuid) {
		return errors.New("数据同步中，请稍后再试")
	}

	// 连锁子店，同步ttpos总部

	// 01商品分类
	// 02物品分类
	// 03税类

	// 连锁子店和总部、散户

	// 1 uom
	// 2 物品
	// 3 规格
	// 4 属性
	// 5 加料
	// 6 商品
	// 7 成本卡
	// 8 供应商
	// 9 仓库
	// 10 仓库物品库存

	go func(ctx context.Context) {
		var isExceptionOccurred bool
		// 确保任务完成时清理状态
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Error("同步任务发生panic", zap.Uint64("companyUuid", companyUuid), zap.Any("panic", r))
				isExceptionOccurred = true
			}

			syncTaskManager.finishTask(companyUuid)
			logger.Logger.Info("同步任务完成", zap.Uint64("companyUuid", companyUuid))

			lastSyncTime := time.Now().Unix()
			// 未报错才记录上次同步完成时间
			if !isExceptionOccurred {
				s.dbm.GetDB(companyUuid).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
				s.dbm.GetDB(0).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
			}

			// 推送websocket
			go websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
				"sync_time":             lastSyncTime,
				"is_exception_occurred": isExceptionOccurred,
			})
		}()

		logger.Logger.Info("开始同步任务", zap.Uint64("companyUuid", companyUuid))

		logger.Logger.Info("开始同步商品分类", zap.Uint64("companyUuid", companyUuid))
		// 01 商品分类
		if err := s.productSrv.SyncProductShopCategory(ctx); err != nil {
			logger.Logger.Error("商品分类同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		logger.Logger.Info("开始同步物品分类", zap.Uint64("companyUuid", companyUuid))
		// 02 物品分类
		if err := s.materialSrv.SyncMaterialCategory(ctx); err != nil {
			logger.Logger.Error("物品分类同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 03 商品税类
		logger.Logger.Info("开始同步商品税类", zap.Uint64("companyUuid", companyUuid))
		if err := s.productSrv.SyncProductTax(ctx); err != nil {
			logger.Logger.Error("商品税类同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 1 uom
		logger.Logger.Info("开始同步单位", zap.Uint64("companyUuid", companyUuid))
		if err := s.productSrv.SyncUnit(ctx); err != nil {
			logger.Logger.Error("单位同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 9 仓库
		logger.Logger.Info("开始同步仓库", zap.Uint64("companyUuid", companyUuid))
		err := s.warehouseSrv.SyncWarehouse(ctx)
		if err != nil {
			logger.Logger.Error("仓库同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 2-1 总部物品
		logger.Logger.Info("开始同步物品", zap.Uint64("companyUuid", companyUuid))
		if err := s.materialSrv.SyncHeadquarterMaterial(ctx); err != nil {
			logger.Logger.Error("物品同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}
		// 2-2 子店物品
		if err := s.materialSrv.SyncSubShopMaterial(ctx); err != nil {
			logger.Logger.Error("物品同步子店失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 3,4,5,6,7 暂不同步

		logger.Logger.Info("开始同步供应商", zap.Uint64("companyUuid", companyUuid))
		// 8 供应商
		if err := s.supplierSrv.SyncSupplier(ctx); err != nil {
			logger.Logger.Error("供应商同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			isExceptionOccurred = true || isExceptionOccurred
		}

		// 10 仓库物品，仅限首次同步
		if company.LastSyncTime == 0 {
			logger.Logger.Info("开始同步默认仓库库存", zap.Uint64("companyUuid", companyUuid))
			if err := s.warehouseSrv.SyncDefaultWarehouseStock(ctx); err != nil {
				logger.Logger.Error("默认仓库库存同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
				isExceptionOccurred = true || isExceptionOccurred
			}
		}

		isExceptionOccurred = false || isExceptionOccurred
	}(ctx)

	return nil
}
