package service

import (
	"time"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/websocket"
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
	// companySetting := ctx.GetCompanySetting()
	// companyUuid := ctx.GetCompanyUuid()

	// 连锁子店，同步ttpos总部
	{
		// 商品分类
		// 物品分类
		// 税类
	}
	// 连锁子店和总部、散户
	{
		// uom
		// 物品
		// 规格
		// 属性
		// 加料
		// 商品
		// 成本卡
		// 供应商
		// 仓库
		// 仓库物品库存
	}

	s.supplierSrv.Sync(ctx)
	s.warehouseSrv.Sync(ctx)

	// TODO 推送同步总部数据配置更新
	go websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceShop, websocket.UPDATE_CONFIG, map[string]any{
		"update_time": time.Now().Unix(),
	})
	return nil
}
