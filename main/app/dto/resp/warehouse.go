package resp

import "ttpos-server-go/app/dto"

// WarehouseResp 仓库响应
type WarehouseResp struct {
	Uuid       uint64             `json:"uuid"`
	LocalName  dto.LocaleResponse `json:"local_name"`
	Type       string             `json:"type"`
	Code       string             `json:"code"`
	Status     int                `json:"status"`
	Contact    string             `json:"contact"`
	Phone      string             `json:"phone"`
	Address    string             `json:"address"`
	IsDefault  int                `json:"is_default"`
	IsEditable bool               `json:"is_editable"` // 是否可编辑
	HasItem    bool               `json:"has_item"`    // 是否存在物品
}

// WarehouseListResp 仓库列表响应
type WarehouseListResp struct {
	List []WarehouseResp  `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// WarehouseInOutListResp 仓库出入库明细列表响应
type WarehouseInOutListResp struct {
	List []WarehouseInOutResp `json:"list"`
	Meta dto.PageResponse     `json:"meta"`
}

// WarehouseInOutResp 仓库出入库明细响应
type WarehouseInOutResp struct {
	Uuid    uint64  `json:"uuid"`     // 出入库明细ID
	OrderNo string  `json:"order_no"` // 单据编号
	Type    string  `json:"type"`     // 类型：采购入库-purchase、销售出库-sale、发货出库-delivery
	Date    string  `json:"date"`     // 日期
	Num     float64 `json:"num"`      // 数量
	Amount  float64 `json:"amount"`   // 金额
	// 物品信息
	MaterialUuid         uint64             `json:"material_uuid"`          // 物品ID
	MaterialName         dto.LocaleResponse `json:"material_name"`          // 物品名称
	MaterialCode         string             `json:"material_code"`          // 物品编码
	MaterialBarcode      string             `json:"material_barcode"`       // 物品条形码
	MaterialCategoryUuid uint64             `json:"material_category_uuid"` // 物料分类ID
	// 供应商信息
	SupplierUuid    uint64             `json:"supplier_uuid"`     // 供应商ID
	SupplierErpCode string             `json:"supplier_erp_code"` // 供应商ERP编码
	SupplierName    dto.LocaleResponse `json:"supplier_name"`     // 供应商名称
	// 仓库信息
	WarehouseUuid uint64             `json:"warehouse_uuid"` // 仓库ID
	WarehouseName dto.LocaleResponse `json:"warehouse_name"` // 仓库名称
	// 对方机构
	OtherOrgUuid uint64 `json:"other_org_uuid"` // 对方机构ID
	OtherOrgType uint64 `json:"other_org_type"` // 对方机构类型
	OtherOrgName string `json:"other_org_name"` // 对方机构名称
}
