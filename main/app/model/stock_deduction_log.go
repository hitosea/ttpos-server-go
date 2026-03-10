package model

// StockDeductionLog 库存扣减日志
// 记录每笔订单每个 item_code 的 Stock Entry 扣减状态
// 用于防止合并扣减时部分成功导致的重复扣减
type StockDeductionLog struct {
	BaseModel
	SaleOrderUuid  uint64  `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:关联订单UUID;NOT NULL" json:"sale_order_uuid"`
	ErpCode        string  `gorm:"column:erp_code;type:varchar(255);default:'';comment:ERP物品编码(item_code);NOT NULL" json:"erp_code"`
	Qty            float64 `gorm:"column:qty;type:decimal(14,4);default:0;comment:扣减数量;NOT NULL" json:"qty"`
	StockEntryName string  `gorm:"column:stock_entry_name;type:varchar(255);default:'';comment:关联的Stock Entry单据名;NOT NULL" json:"stock_entry_name"`
}

func (StockDeductionLog) TableName() string {
	return "ttpos_stock_deduction_log"
}
