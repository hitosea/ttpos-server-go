package model

// WarehouseInOutLog 仓库出入库记录表 `ttpos_warehouse_in_out_log`
type WarehouseInOutLog struct {
	BaseModel
	LogType              int     `json:"log_type" gorm:"type:int;default:0;comment:日志类型,0-入库 1-出库"`
	Scene                int     `json:"scene" gorm:"type:int;default:0;comment:场景,0-采购入库 1-销售出库 2-发货出库"`
	WarehouseUuid        uint64  `json:"warehouse_uuid" gorm:"type:bigint;default:0;comment:仓库ID"`
	MaterialUuid         uint64  `json:"material_uuid" gorm:"type:bigint;default:0;comment:物品ID"`
	MaterialName         string  `json:"material_name" gorm:"type:text;default:'';comment:物品名称JSON,记录当时物品名称"`
	MaterialBaseUnitUuid uint64  `json:"material_base_unit_uuid" gorm:"type:bigint;default:0;comment:物品基准单位ID"`
	MaterialBaseUnitName string  `json:"material_base_unit_name" gorm:"type:text;default:'';comment:物品基准单位名称"`
	Num                  float64 `json:"num" gorm:"type:decimal(22,4);default:0;comment:数量"`
	Price                float64 `json:"price" gorm:"type:decimal(22,4);default:0;comment:单价，物品基准单位单价"`
	Amount               float64 `json:"amount" gorm:"type:decimal(22,4);default:0;comment:金额,单价*数量"`
	SupplierUuid         uint64  `json:"supplier_uuid" gorm:"type:bigint;default:0;comment:供应商ID"`
	SupplierErpCode      string  `json:"supplier_erp_code" gorm:"type:varchar(500);default:'';comment:供应商ERP编码"`
	SupplierName         string  `json:"supplier_name" gorm:"type:text;comment:供应商名称"`
	OrderNo              string  `json:"order_no" gorm:"type:varchar(255);default:'';comment:单据编号"`
	OtherOrgUuid         uint64  `json:"other_org_uuid" gorm:"type:bigint;default:0;comment:对方机构ID"`
	OtherOrgType         uint64  `json:"other_org_type" gorm:"type:bigint;default:0;comment:对方机构类型 0:供应商 1:客户"`
	OtherOrgName         string  `json:"other_org_name" gorm:"type:text;comment:对方机构名称"`
	OpeningHours         string  `json:"opening_hours" gorm:"type:varchar(255);default:'';comment:营业时段,仅用于Scene销售出库的场景"`

	// 关联模型
	Material  *Material  `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
	Supplier  *Supplier  `gorm:"foreignKey:SupplierUuid;references:Uuid" json:"supplier,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid" json:"warehouse,omitempty"`
}

func (w *WarehouseInOutLog) TableName() string {
	return "ttpos_warehouse_in_out_log"
}

func (w *WarehouseInOutLog) SetNil() {
	w.Material = nil
	w.Supplier = nil
	w.Warehouse = nil
}
