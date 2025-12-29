package resp

import (
	"ttpos-server-go/app/dto"
)

// StockReconciliationListResp 盘点单列表响应
type StockReconciliationListResp struct {
	List []*StockReconciliationInfo `json:"list"` // 盘点单列表
	Meta dto.PageResponse           `json:"meta"` // 分页信息
}

// StockReconciliationInfo 盘点单信息
type StockReconciliationInfo struct {
	Uuid                uint64             `json:"uuid"`                  // 盘点单UUID
	OrderNo             string             `json:"order_no"`              // 单据编号
	ErpCode             string             `json:"erp_code"`              // ERP盘点单号
	Type                int                `json:"type"`                  // 盘点类型 1-指定物品盘点 2-全部物品盘点
	WarehouseUuid       uint64             `json:"warehouse_uuid"`        // 仓库UUID
	WarehouseLocaleName dto.LocaleResponse `json:"warehouse_locale_name"` // 仓库多语言名称
	Status              int                `json:"status"`                // 状态 0-已保存 1-已提交 2-已审核 3-已驳回
	ItemsCount          int                `json:"items_count"`           // 物品明细数量
	SubmitTime          int                `json:"submit_time"`           // 提交时间
	CreateTime          int                `json:"create_time"`           // 创建时间
}

// StockReconciliationDetailResp 盘点单详情响应
type StockReconciliationDetailResp struct {
	Uuid          uint64                         `json:"uuid"`           // 盘点单UUID
	OrderNo       string                         `json:"order_no"`       // 单据编号
	ErpCode       string                         `json:"erp_code"`       // ERP盘点单号
	Type          int                            `json:"type"`           // 盘点类型 1-指定物品盘点 2-全部物品盘点
	WarehouseUuid uint64                         `json:"warehouse_uuid"` // 仓库UUID
	WarehouseName dto.LocaleResponse             `json:"warehouse_name"` // 仓库名称
	Purpose       int                            `json:"purpose"`        // 盘点目的 1-库存盘点 2-期初盘点
	Status        int                            `json:"status"`         // 状态 0-已保存 1-已提交 2-已审核 3-已驳回
	Items         []*StockReconciliationItemInfo `json:"items"`          // 盘点单物品明细
	SubmitTime    int                            `json:"submit_time"`    // 提交时间
	CreateTime    int                            `json:"create_time"`    // 创建时间
	UpdateTime    int                            `json:"update_time"`    // 更新时间
}

// StockReconciliationItemInfo 盘点单物品明细信息
type StockReconciliationItemInfo struct {
	MaterialUuid               uint64                             `json:"material_uuid"`                 // 物品UUID
	MaterialBarcode            string                             `json:"material_barcode"`              // 物品条码
	InternalCode               string                             `json:"internal_code"`                 // 内部编码
	MaterialCode               string                             `json:"material_code"`                 // 物品编码
	LocaleName                 dto.LocaleResponse                 `json:"locale_name"`                   // 物品名称
	BookedQuantity             float64                            `json:"booked_quantity"`               // 账面库存数量
	CountedQuantity            float64                            `json:"counted_quantity"`              // 实盘库存数量
	ItemUnits                  []*StockReconciliationItemUnitInfo `json:"item_units"`                    // 盘点单物品单位明细
	Units                      []MaterialUnitInfo                 `json:"units"`                         // 所有单位（包含基准单位）
	CreateTime                 int                                `json:"create_time"`                   // 创建时间
	InventoryStatus            int                                `json:"inventory_status"`              // 库存状态 1-盘盈 2-盘亏 3-正常
	DiffQuantity               float64                            `json:"diff_quantity"`                 // 盈亏差值数量
	IsInventoryStatusException bool                               `json:"is_inventory_status_exception"` // 是否盘盈盘亏异常（差值大于20%）
}

// StockReconciliationItemUnitInfo 盘点单物品单位明细信息
type StockReconciliationItemUnitInfo struct {
	MaterialUnitUuid uint64             `json:"material_unit_uuid"` // 单位UUID
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 单位名称
	Quantity         *float64           `json:"quantity"`           // 单位数量
	ConversionRate   float64            `json:"conversion_rate"`    // 转换率（相对于基准单位）
}

// StockReconciliationUuidResp 保存、提交盘点单响应
type StockReconciliationUuidResp struct {
	Uuid uint64 `json:"uuid"` // 盘点单UUID
}

type StockReconciliationApproveResp struct {
	List []dto.LocaleResponse `json:"list"` // 禁用物品列表
}

type StockReconciliationCheckMaterialsResp struct {
	LocaleName                 dto.LocaleResponse `json:"locale_name"`                   // 物品名称
	IsInventoryStatusException bool               `json:"is_inventory_status_exception"` // 是否盈亏异常
	Status                     bool               `json:"status"`                        // 物品状态,true上架 false下架
	IsDeleted                  bool               `json:"is_deleted"`                    // 是否已删除
	UnitCount                  uint               `json:"unit_count"`                    // 单位数量
	ExistsInWarehouse          bool               `json:"exists_in_warehouse"`           // 是否在仓库中
}

type StockReconciliationCheckMaterialsListResp struct {
	List              []StockReconciliationCheckMaterialsResp `json:"list"`               // 物品列表
	WarehouseDisabled bool                                    `json:"warehouse_disabled"` // 仓库是否禁用 true-被禁用；false-正常
}

// StockReconciliationTemplateResp 盘点单模板响应
type StockReconciliationTemplateResp struct {
	LastUpdated string                          `json:"last_updated"` // 最后更新时间
	Count       int                             `json:"count"`        // 总数量
	Data        StockReconciliationTemplateData `json:"data"`         // 模板数据
}

// StockReconciliationTemplateData 盘点单模板数据
type StockReconciliationTemplateData struct {
	Daily   []string `json:"daily"`   // 日盘物品编号列表
	Weekly  []string `json:"weekly"`  // 周盘物品编号列表
	Monthly []string `json:"monthly"` // 月盘物品编号列表
}
