package model

// ProductPackageGroupItemTakeout 外卖套餐子商品价格表 `ttpos_product_package_group_item_takeout`
// 存储套餐子商品在外卖平台的加价信息
type ProductPackageGroupItemTakeout struct {
	BaseModel
	ProductPackageTakeoutUuid   uint64  `gorm:"default:0;column:product_package_takeout_uuid;comment:'外卖商品UUID，关联 ttpos_product_package_takeout.uuid'"`
	ProductPackageGroupItemUuid uint64  `gorm:"default:0;column:product_package_group_item_uuid;comment:'套餐子商品UUID，关联 ttpos_product_package_group_item.uuid'"`
	ProductPackageGroupUuid     uint64  `gorm:"default:0;column:product_package_group_uuid;comment:'套餐分组UUID，关联 ttpos_product_package_group.uuid'"`
	HeadquarterUuid             uint64  `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID,0表示不是总部商品'"`
	AddPrice                    float64 `gorm:"type:decimal(22,4);default:0.0000;column:add_price;comment:'外卖平台的加价金额（覆盖店内加价）'"`

	// 关联关系
	ProductPackageTakeout   ProductPackageTakeout   `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"`    // 外卖商品
	ProductPackageGroupItem ProductPackageGroupItem `gorm:"foreignKey:product_package_group_item_uuid;references:uuid" json:"-"` // 套餐子商品
	ProductPackageGroup     ProductPackageGroup     `gorm:"foreignKey:product_package_group_uuid;references:uuid" json:"-"`      // 套餐分组
}

// TableName 指定表名
func (ProductPackageGroupItemTakeout) TableName() string {
	return "ttpos_product_package_group_item_takeout"
}
