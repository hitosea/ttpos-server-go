package model

import (
	"github.com/shopspring/decimal"
)

// StockReconciliation 盘点单表
type StockReconciliation struct {
	BaseModel
	OrderNo            string `gorm:"column:order_no;type:varchar(255);not null;default:'';index:idx_order_no;comment:单据编号" json:"order_no"`
	ErpCode            string `gorm:"column:erp_code;type:varchar(255);not null;default:'';comment:ERP盘点单号" json:"erp_code"`
	Type               int    `gorm:"column:type;not null;default:1;comment:盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘 6-固定资产盘点" json:"type"`
	WarehouseUuid      uint64 `gorm:"column:warehouse_uuid;not null;default:0;index:idx_warehouse_uuid;comment:仓库ID" json:"warehouse_uuid"`
	Purpose            int    `gorm:"column:purpose;not null;default:1;comment:盘点目的 1-库存盘点 2-期初盘点" json:"purpose"`
	Status             int    `gorm:"column:status;not null;default:0;index:idx_status;comment:状态 0-已保存 1-已提交 2-已审核 3-已驳回" json:"status"`
	SubmitTime         int    `gorm:"column:submit_time;not null;default:0;comment:提交时间(时间戳)" json:"submit_time"` // 为了能按照这个时间排序,保存时先记录一个时间,最后真正提交时会再更新提交时间
	SubmitterStaffUuid uint64 `gorm:"column:submitter_staff_uuid;not null;default:0;comment:发起人员工UUID" json:"submitter_staff_uuid"`

	// 关联仓库
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid"`
	// 关联物品明细
	StockReconciliationItems []*StockReconciliationItem `gorm:"foreignKey:StockReconciliationUuid;references:Uuid"`
}

// StockReconciliationItem 盘点单物品明细表
type StockReconciliationItem struct {
	BaseModel
	StockReconciliationUuid uint64          `gorm:"column:stock_reconciliation_uuid;not null;default:0;index:idx_stock_reconciliation_uuid;comment:盘点单ID" json:"stock_reconciliation_uuid"`
	MaterialUuid            uint64          `gorm:"column:material_uuid;not null;default:0;index:idx_material_uuid;comment:物品ID" json:"material_uuid"`
	MaterialName            string          `gorm:"column:material_name;type:text;comment:物品名称，用于备份多语言" json:"material_name"`
	BookedQuantity          decimal.Decimal `gorm:"column:booked_quantity;type:decimal(22,4);not null;default:0.0000;comment:账面库存数量，基准单位后的数量" json:"booked_quantity"`
	CountedQuantity         decimal.Decimal `gorm:"column:counted_quantity;type:decimal(22,4);not null;default:0.0000;comment:实盘库存数量，物品所有单位换算成基准单位后的数量" json:"counted_quantity"`

	// 关联物品
	Material *Material `gorm:"foreignKey:MaterialUuid;references:Uuid"`
	// 关联StockReconciliationItemUnit
	StockReconciliationItemUnits []*StockReconciliationItemUnit `gorm:"foreignKey:StockReconciliationItemUuid;references:Uuid"`
}

// StockReconciliationItemUnit 盘点单物品单位明细表
type StockReconciliationItemUnit struct {
	BaseModel
	StockReconciliationItemUuid uint64   `gorm:"column:stock_reconciliation_item_uuid;not null;default:0;index:idx_stock_reconciliation_item_uuid;comment:盘点单物品明细ID" json:"stock_reconciliation_item_uuid"`
	MaterialUnitUuid            uint64   `gorm:"column:material_unit_uuid;not null;default:0;index:idx_material_unit_uuid;comment:单位ID" json:"material_unit_uuid"`
	MaterialUnitName            string   `gorm:"column:material_unit_name;type:text;comment:物品单位名称，用于备份多语言" json:"material_unit_name"`
	Quantity                    *float64 `gorm:"column:quantity;type:decimal(22,4);comment:单位数量" json:"quantity"`

	// 关联 ttpos_material_unit 表
	MaterialUnit *MaterialUnit `gorm:"foreignKey:MaterialUnitUuid;references:Uuid"`
}
