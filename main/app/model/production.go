package model

type ProductionOrder struct {
	BaseModel
	DeskUuid      uint64 `gorm:"column:desk_uuid;comment:'桌台ID'" json:"desk_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;comment:'销售订单ID'" json:"sale_order_uuid"`
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;comment:'销售账单ID'" json:"sale_bill_uuid"`
}

type ProductionOrderProduct struct {
	BaseModel
	ProductionOrderUuid   uint64 `gorm:"column:production_order_uuid;comment:'生产订单ID'" json:"production_order_uuid"`
	SaleOrderProductUuid  uint64 `gorm:"column:sale_order_product_uuid;comment:'销售订单商品ID'" json:"sale_order_product_uuid"`
	FirstCategoryUuid     uint64 `gorm:"column:first_category_uuid;comment:'一级分类ID'" json:"first_category_uuid"`
	Num                   uint   `gorm:"column:num;comment:'商品数量'" json:"num"`
	FlavorName            string `gorm:"column:flavor_name;comment:'规格名称,不随后台改变'" json:"flavor_name"`
	ProductAttributeNames string `gorm:"column:product_attribute_names;comment:'商品属性名称,多个属性名用逗号分隔,不随后台改变'" json:"product_attribute_names"`
	ProductSaucesNames    string `gorm:"column:product_sauces_names;comment:'商品加料名称,多个加料名用逗号分隔,不随后台改变'" json:"product_sauces_names"`
	Status                uint   `gorm:"column:status;comment:'状态, 0-待制作 1-制作中 2-已完成 3-已退菜'" json:"status"`
	Remark                string `gorm:"column:remark;comment:'商品备注'" json:"remark"`
	HasMaterial           uint   `gorm:"column:has_material;comment:'是否无原料, 0-无原料,商品没有关联原料 1-有原料'" json:"has_material"`
	FinishedTime          int64  `gorm:"column:finished_time;comment:'完成时间(时间戳)'" json:"finished_time"`
}
