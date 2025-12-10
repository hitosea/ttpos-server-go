package model

// ProductBomTakeout 外卖规格价格表 `ttpos_product_bom_takeout`
type ProductBomTakeout struct {
	BaseModel
	ProductPackageTakeoutUuid uint64  `gorm:"default:0;column:product_package_takeout_uuid;comment:'外卖商品UUID，关联 ttpos_product_package_takeout.uuid'"`
	ProductBomUuid            uint64  `gorm:"default:0;column:product_bom_uuid;comment:'店内商品BOM UUID，关联 ttpos_product_bom.uuid'"`
	HeadquarterUuid           uint64  `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID,0表示不是总部商品'"`
	Price                     float64 `gorm:"type:decimal(22,4);default:0.0000;column:price;comment:'外卖规格价格'"`

	// 关联关系
	ProductPackageTakeout ProductPackageTakeout `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"` // 外卖商品
	ProductBom            ProductBom            `gorm:"foreignKey:product_bom_uuid;references:uuid" json:"-"`             // 店内商品BOM
}

// TableName 指定表名
func (ProductBomTakeout) TableName() string {
	return "ttpos_product_bom_takeout"
}
