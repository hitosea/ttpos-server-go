package model

// ProductPackageGroupItem 商品套餐组商品模型 `ttpos_product_package_group_item`
type ProductPackageGroupItem struct {
	BaseModel
	ProductPackageGroupUuid uint64  `json:"product_package_group_uuid" gorm:"index:idx_product_package_group_uuid;not null;default:0;comment:商品套餐组ID"`
	RelatedUuid             uint64  `json:"related_uuid" gorm:"index:idx_related_uuid;not null;default:0;comment:关联商品UUID, product_package_uuid"`
	ProductBomUuid          uint64  `json:"product_bom_uuid" gorm:"index:idx_product_bom_uuid;not null;default:0;comment:商品BOM UUID,商品规格uuid"`
	Num                     float64 `json:"num" gorm:"type:decimal(12,4);not null;default:0;comment:数量"`
	Sort                    int     `json:"sort" gorm:"index:idx_sort;not null;default:0;comment:排序"`

	ProductBom          *ProductBom          `gorm:"foreignKey:product_bom_uuid;references:uuid"`           // 商品BOM规格
	ProductPackage      *ProductPackage      `gorm:"foreignKey:related_uuid;references:uuid"`               // 商品套餐
	ProductPackageGroup *ProductPackageGroup `gorm:"foreignKey:product_package_group_uuid;references:uuid"` // 商品套餐组
}
