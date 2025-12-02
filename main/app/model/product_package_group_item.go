package model

// ProductPackageGroupItem 商品套餐组商品模型 `ttpos_product_package_group_item`
type ProductPackageGroupItem struct {
	BaseModel
	ProductPackageGroupUuid uint64  `json:"product_package_group_uuid" gorm:"index:idx_product_package_group_uuid;not null;default:0;comment:商品套餐组ID"`
	RelatedUuid             uint64  `json:"related_uuid" gorm:"index:idx_related_uuid;not null;default:0;comment:关联商品UUID, product_package_uuid"`
	ProductBomUuid          uint64  `json:"product_bom_uuid" gorm:"index:idx_product_bom_uuid;not null;default:0;comment:商品BOM UUID,商品规格uuid"`
	Num                     float64 `json:"num" gorm:"type:decimal(12,4);not null;default:0;comment:数量"`
	Sort                    int     `json:"sort" gorm:"index:idx_sort;not null;default:0;comment:排序"`
	AddPrice                float64 `json:"add_price" gorm:"type:decimal(22,4);not null;default:0.00;comment:加价金额，表示该商品需要加价多少钱"`
	IsRequired              int     `json:"is_required" gorm:"type:int(10);not null;default:0;comment:必选 0-不必选 1-必选"`
	IsDefault               int     `json:"is_default" gorm:"type:int(10);not null;default:0;comment:默认选中 0-默认不选中 1-默认选中"`

	ProductBom          *ProductBom          `gorm:"foreignKey:product_bom_uuid;references:uuid"`           // 商品BOM规格
	ProductPackage      *ProductPackage      `gorm:"foreignKey:related_uuid;references:uuid"`               // 商品套餐
	ProductPackageGroup *ProductPackageGroup `gorm:"foreignKey:product_package_group_uuid;references:uuid"` // 商品套餐组
}
