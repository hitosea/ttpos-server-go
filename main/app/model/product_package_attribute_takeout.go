package model

// ProductPackageAttributeTakeout 外卖属性价格表 `ttpos_product_package_attribute_takeout`
type ProductPackageAttributeTakeout struct {
	BaseModel
	ProductPackageTakeoutUuid   uint64  `gorm:"default:0;column:product_package_takeout_uuid;comment:'外卖商品UUID，关联 ttpos_product_package_takeout.uuid'"`
	ProductPackageAttributeUuid uint64  `gorm:"default:0;column:product_package_attribute_uuid;comment:'店内商品属性 UUID，关联 ttpos_product_package_attribute.uuid'"`
	HeadquarterUuid             uint64  `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID,0表示不是总部商品'"`
	Price                       float64 `gorm:"type:decimal(22,4);default:0.0000;column:price;comment:'外卖属性价格'"`

	// 关联关系
	ProductPackageTakeout   ProductPackageTakeout   `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"`   // 外卖商品
	ProductPackageAttribute ProductPackageAttribute `gorm:"foreignKey:product_package_attribute_uuid;references:uuid" json:"-"` // 店内商品属性
}

// TableName 指定表名
func (ProductPackageAttributeTakeout) TableName() string {
	return "ttpos_product_package_attribute_takeout"
}
