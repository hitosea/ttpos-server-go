package service

import (
	"sync"
	"time"
	"ttpos-server-go/pkg/context"
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
	warehouseSrv IWarehouseSrv
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

// NewSyncSrv 创建新同步服务
func NewSyncSrv(warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv) ISyncSrv {
	return NewSyncSrvImpl(warehouseSrv, supplierSrv, productSrv)
}

// NewSyncSrvImpl 创建新同步服务实现
func NewSyncSrvImpl(warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv) ISyncSrv {
	return &SyncSrv{
		warehouseSrv: warehouseSrv,
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
		logger.Logger.Info("同步任务已在运行，跳过本次同步", zap.Uint64("companyUuid", companyUuid))
		return nil
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
		// 确保任务完成时清理状态
		defer func() {
			syncTaskManager.finishTask(companyUuid)
			logger.Logger.Info("同步任务完成", zap.Uint64("companyUuid", companyUuid))

			// 推送websocket
			go websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
				"sync_time": time.Now().Unix(),
			})
		}()

		logger.Logger.Info("开始同步任务", zap.Uint64("companyUuid", companyUuid))

		// 1 uom
		if err := s.productSrv.SyncUnit(ctx); err != nil {
			logger.Logger.Error("单位同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		}

		// 3 规格
		if err := s.productSrv.SyncProductFlavor(ctx); err != nil {
			logger.Logger.Error("规格同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		}

		// TODO 等待erp接口 4.1 属性组
		// if err := s.productSrv.SyncAttributeGroup(ctx); err != nil {
		// 	logger.Logger.Error("属性组同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		// }

		// TODO 等待erp接口 5.1 加料
		// if err := s.productSrv.SyncSauce(ctx); err != nil {
		// 	logger.Logger.Error("加料同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		// }

		// 8 供应商
		if err := s.supplierSrv.SyncSupplier(ctx); err != nil {
			logger.Logger.Error("供应商同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		}

		// 9 仓库
		firstSync, err := s.warehouseSrv.SyncWarehouse(ctx)
		if err != nil {
			logger.Logger.Error("仓库同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
		}

		// 10 仓库物品
		if firstSync {
			if err := s.warehouseSrv.SyncDefaultWarehouseStock(ctx); err != nil {
				logger.Logger.Error("默认仓库库存同步失败", zap.Uint64("companyUuid", companyUuid), zap.Error(err))
			}
		}
	}(ctx)

	return nil
}
