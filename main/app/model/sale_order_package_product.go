package model

// `ttpos_sale_order_package_product`
type SaleOrderPackageProduct struct {
	BaseModel
	SaleOrderUuid               uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单ID;NOT NULL" json:"sale_order_uuid"`
	SaleBillUuid                uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单ID;NOT NULL" json:"sale_bill_uuid"`
	RelatedUuid                 uint64 `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联订单套餐UUID, sale_order_product_uuid;NOT NULL" json:"related_uuid"`
	SaleOrderProductUuid        uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;default:0;comment:销售订单商品ID,套餐子商品的订单商品;NOT NULL" json:"sale_order_product_uuid"`
	ProductPackageGroupItemUuid uint64 `gorm:"column:product_package_group_item_uuid;type:bigint(20) unsigned;default:0;comment:套餐分组商品ID;NOT NULL" json:"product_package_group_item_uuid"`

	SaleOrderProduct *SaleOrderProduct `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
}
