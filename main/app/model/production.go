package model

// 生产单 `ttpos_production_order`
type ProductionOrder struct {
	BaseModel
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;comment:'销售订单ID'" json:"sale_order_uuid"`
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;comment:'销售账单ID'" json:"sale_bill_uuid"`

	ProductionOrderProducts []*ProductionOrderProduct `gorm:"foreignKey:ProductionOrderUuid;references:Uuid" json:"production_order_products"`
}

// 生产单商品 `ttpos_production_order_product`
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

	ProductionOrderMaterials []*ProductionOrderMaterial `gorm:"foreignKey:ProductionOrderProductUuid;references:Uuid" json:"production_order_materials"`
	SaleOrderProduct         SaleOrderProduct           `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid" json:"sale_order_product"`
}

// 生产单原料 `ttpos_production_order_material`
type ProductionOrderMaterial struct {
	BaseModel
	Name                       string `gorm:"column:name;default:'';comment:'原料名称,不随后台改变'" json:"name"`
	MaterialUuid               uint64 `gorm:"column:material_uuid;default:0;comment:'原料ID'" json:"material_uuid"`
	Num                        int    `gorm:"column:num;default:0;comment:'原料数量'" json:"num"`
	Unit                       string `gorm:"column:unit;default:'';comment:'单位,不随后台改变'" json:"unit"`
	IsProductBom               uint   `gorm:"column:is_product_bom;default:0;comment:'是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1'" json:"is_product_bom"`
	ProductionOrderProductUuid uint64 `gorm:"column:production_order_product_uuid;default:0;comment:'生产订单商品ID'" json:"production_order_product_uuid"`
	SaleOrderProductUuid       uint64 `gorm:"column:sale_order_product_uuid;default:0;comment:'销售订单商品ID'" json:"sale_order_product_uuid"`
}
