package resp

import (
	"ttpos-server-go/app/dto"

	"github.com/shopspring/decimal"
)

// StockReconciliationListResp 盘点单列表响应
type StockReconciliationListResp struct {
	List []*StockReconciliationInfo `json:"list"` // 盘点单列表
	Meta dto.PageResponse           `json:"meta"` // 分页信息
}

// StockReconciliationInfo 盘点单信息
type StockReconciliationInfo struct {
	Uuid          uint64 `json:"uuid"`           // 盘点单UUID
	OrderNo       string `json:"order_no"`       // 单据编号
	ErpCode       string `json:"erp_code"`       // ERP盘点单号
	Type          int    `json:"type"`           // 盘点类型 1-指定物品盘点 2-全部物品盘点
	WarehouseUuid uint64 `json:"warehouse_uuid"` // 仓库UUID
	Purpose       int    `json:"purpose"`        // 盘点目的 1-库存盘点 2-期初盘点
	Status        int    `json:"status"`         // 状态 0-已保存 1-已提交 2-已审核 3-已驳回
	CreateTime    int    `json:"create_time"`    // 创建时间
	UpdateTime    int    `json:"update_time"`    // 更新时间
}

// StockReconciliationDetailResp 盘点单详情响应
type StockReconciliationDetailResp struct {
	Uuid          uint64                         `json:"uuid"`           // 盘点单UUID
	OrderNo       string                         `json:"order_no"`       // 单据编号
	ErpCode       string                         `json:"erp_code"`       // ERP盘点单号
	Type          int                            `json:"type"`           // 盘点类型 1-指定物品盘点 2-全部物品盘点
	WarehouseUuid uint64                         `json:"warehouse_uuid"` // 仓库UUID
	Purpose       int                            `json:"purpose"`        // 盘点目的 1-库存盘点 2-期初盘点
	Status        int                            `json:"status"`         // 状态 0-已保存 1-已提交 2-已审核 3-已驳回
	Items         []*StockReconciliationItemInfo `json:"items"`          // 盘点单物品明细
	CreateTime    int                            `json:"create_time"`    // 创建时间
	UpdateTime    int                            `json:"update_time"`    // 更新时间
}

// StockReconciliationItemInfo 盘点单物品明细信息
type StockReconciliationItemInfo struct {
	MaterialUuid               uint64                             `json:"material_uuid"`                 // 物品UUID
	MaterialCode               string                             `json:"material_code"`                 // 物品编码
	LocaleName                 dto.LocaleResponse                 `json:"locale_name"`                   // 物品名称
	BookedQuantity             decimal.Decimal                    `json:"booked_quantity"`               // 账面库存数量
	CountedQuantity            decimal.Decimal                    `json:"counted_quantity"`              // 实盘库存数量
	Units                      []*StockReconciliationItemUnitInfo `json:"units"`                         // 盘点单物品单位明细
	CreateTime                 int                                `json:"create_time"`                   // 创建时间
	InventoryStatus            int                                `json:"inventory_status"`              // 库存状态 1-盘盈 2-盘亏 3-正常
	IsInventoryStatusException bool                               `json:"is_inventory_status_exception"` // 是否盘盈盘亏异常（差值大于20%）
}

// StockReconciliationItemUnitInfo 盘点单物品单位明细信息
type StockReconciliationItemUnitInfo struct {
	MaterialUnitUuid uint64             `json:"material_unit_uuid"` // 单位UUID
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 单位名称
	Quantity         *decimal.Decimal   `json:"quantity"`           // 单位数量
}

// StockReconciliationCreateResp 创建盘点单响应
type StockReconciliationCreateResp struct {
	Uuid uint64 `json:"uuid"` // 盘点单UUID
}
