package model

import (
	"github.com/shopspring/decimal"
)

// StockLoss 报损单主表
type StockLoss struct {
	BaseModel
	Code          string `gorm:"column:code;type:varchar(50);not null;default:'';index:idx_code;comment:单据编号 SL+时间戳+序列号" json:"code"`
	ErpCode       string `gorm:"column:erp_code;type:varchar(50);not null;default:'';index:idx_erp_code;comment:ERP单据编号" json:"erp_code"`
	LossType      int    `gorm:"column:loss_type;not null;default:1;comment:报损类型 1:物品损坏 2:物品报废 3:物品过期" json:"loss_type"`
	WarehouseUuid uint64 `gorm:"column:warehouse_uuid;not null;default:0;index:idx_warehouse_uuid;comment:报损仓库UUID" json:"warehouse_uuid"`
	Reason        string `gorm:"column:reason;type:text;comment:报损原因" json:"reason"`
	Status        int    `gorm:"column:status;not null;default:0;index:idx_status;comment:状态 0:已保存 1:已提交 2:已审核通过 3:已驳回" json:"status"`
	SubmitTime    int    `gorm:"column:submit_time;not null;default:0;comment:提交时间" json:"submit_time"`
	ApproveTime   int    `gorm:"column:approve_time;not null;default:0;comment:审核通过时间" json:"approve_time"`
	RejectTime    int    `gorm:"column:reject_time;not null;default:0;comment:驳回时间" json:"reject_time"`
	SubmitterUuid uint64 `gorm:"column:submitter_uuid;not null;default:0;comment:提交人UUID" json:"submitter_uuid"`
	ApproverUuid  uint64 `gorm:"column:approver_uuid;not null;default:0;comment:审核人UUID" json:"approver_uuid"`

	// 关联仓库
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid"`
	// 关联物品明细
	StockLossItems []*StockLossItem `gorm:"foreignKey:StockLossUuid;references:Uuid"`
	// 关联附件
	StockLossFiles []*StockLossFile `gorm:"foreignKey:StockLossUuid;references:Uuid"`
	// 关联批注
	StockLossAnnotations []*StockLossAnnotation `gorm:"foreignKey:StockLossUuid;references:Uuid"`
}

// TableName 指定表名
func (StockLoss) TableName() string {
	return "ttpos_stock_loss"
}

// StockLossItem 报损单明细表
type StockLossItem struct {
	BaseModel
	StockLossUuid uint64          `gorm:"column:stock_loss_uuid;not null;default:0;index:idx_stock_loss_uuid;comment:报损单UUID" json:"stock_loss_uuid"`
	MaterialUuid  uint64          `gorm:"column:material_uuid;not null;default:0;index:idx_material_uuid;comment:物料UUID" json:"material_uuid"`
	MaterialName  string          `gorm:"column:material_name;type:text;comment:物料名称（多语言JSON备份）" json:"material_name"`
	BaseQuantity  decimal.Decimal `gorm:"column:base_quantity;type:decimal(14,4);not null;default:0.0000;comment:基准单位数量（所有单位转换后的总量）" json:"base_quantity"`

	// 关联物品
	Material *Material `gorm:"foreignKey:MaterialUuid;references:Uuid"`
	// 关联单位明细
	StockLossItemUnits []*StockLossItemUnit `gorm:"foreignKey:StockLossItemUuid;references:Uuid"`
}

// TableName 指定表名
func (StockLossItem) TableName() string {
	return "ttpos_stock_loss_item"
}

// StockLossItemUnit 报损单物品单位明细表
type StockLossItemUnit struct {
	BaseModel
	StockLossItemUuid uint64   `gorm:"column:stock_loss_item_uuid;not null;default:0;index:idx_stock_loss_item_uuid;comment:报损单物品明细UUID" json:"stock_loss_item_uuid"`
	MaterialUnitUuid  uint64   `gorm:"column:material_unit_uuid;not null;default:0;index:idx_material_unit_uuid;comment:物料单位UUID" json:"material_unit_uuid"`
	MaterialUnitName  string   `gorm:"column:material_unit_name;type:text;comment:物料单位名称（多语言JSON备份）" json:"material_unit_name"`
	Quantity          *float64 `gorm:"column:quantity;type:decimal(22,4);comment:单位数量" json:"quantity"`

	// 关联单位
	MaterialUnit *MaterialUnit `gorm:"foreignKey:MaterialUnitUuid;references:Uuid"`
}

// TableName 指定表名
func (StockLossItemUnit) TableName() string {
	return "ttpos_stock_loss_item_unit"
}

// 报损单状态常量
const (
	StockLossStatusSaved    = 0 // 已保存
	StockLossStatusSubmit   = 1 // 已提交
	StockLossStatusApproved = 2 // 已审核通过
	StockLossStatusRejected = 3 // 已驳回
)

// 报损类型常量
const (
	StockLossTypeDamage  = 1 // 物品损坏
	StockLossTypeScrap   = 2 // 物品报废
	StockLossTypeExpired = 3 // 物品过期
)
