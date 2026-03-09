package model

// StocktakeSnapshot 盘点快照-未出库订单快照
// 记录盘点时尚未通过 Stock Entry 扣减的订单商品数量
// 用于盘点差异计算：预期库存 = ERP账面 - 快照扣减量
type StocktakeSnapshot struct {
	BaseModel
	StockReconciliationUuid uint64  `gorm:"column:stock_reconciliation_uuid;type:bigint(20);default:0;comment:关联盘点单UUID;NOT NULL" json:"stock_reconciliation_uuid"`
	ItemCode                string  `gorm:"column:item_code;type:varchar(255);default:'';comment:ERP物品编码;NOT NULL" json:"item_code"`
	WarehouseErpCode        string  `gorm:"column:warehouse_erp_code;type:varchar(255);default:'';comment:ERP仓库编码;NOT NULL" json:"warehouse_erp_code"`
	PendingQty              float64 `gorm:"column:pending_qty;type:decimal(12,3);default:0;comment:待扣减数量;NOT NULL" json:"pending_qty"`
	OrderCount              int     `gorm:"column:order_count;type:int(11);default:0;comment:涉及订单数;NOT NULL" json:"order_count"`
}

func (StocktakeSnapshot) TableName() string {
	return "ttpos_stocktake_snapshot"
}
