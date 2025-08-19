package model

// WarehouseMonthlyForm 月度报表表 `ttpos_warehouse_monthly_form`
type WarehouseMonthlyForm struct {
	BaseModel
	Year  int     `gorm:"default:0;comment:'年'"`
	Month int     `gorm:"default:0;comment:'月'"`
	Scene int     `gorm:"default:0;comment:'记录类型,0-月初 1-月末'"`
	Stock float64 `gorm:"type:decimal(20,4);default:0.0000;comment:'库存'"`
}

// WarehouseMonthlyProductBomForm 月度商品bom报表表 `ttpos_warehouse_monthly_product_bom_form`
type WarehouseMonthlyProductBomForm struct {
	BaseModel
	Year           int     `gorm:"default:0;comment:'年'"`
	Month          int     `gorm:"default:0;comment:'月'"`
	Scene          int     `gorm:"default:0;comment:'记录类型,0-月初 1-月末'"`
	ProductBomUuid uint64  `gorm:"default:0;comment:'商品bom uuid'"`
	Stock          float64 `gorm:"type:decimal(20,4);default:0.0000;comment:'库存'"`
}

// WarehouseMonthlyMaterialForm 月度材料报表表 `ttpos_warehouse_monthly_material_form`
type WarehouseMonthlyMaterialForm struct {
	BaseModel
	Year         int     `gorm:"default:0;comment:'年'"`
	Month        int     `gorm:"default:0;comment:'月'"`
	Scene        int     `gorm:"default:0;comment:'记录类型,0-月初 1-月末'"`
	MaterialUuid uint64  `gorm:"default:0;comment:'材料uuid'"`
	Stock        float64 `gorm:"type:decimal(20,4);default:0.0000;comment:'库存'"`
}
