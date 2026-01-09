package types

// ModifierInfo 修饰符信息（名称、数量、价格、分类）
type ModifierInfo struct {
	Name                      string  // 修饰符名称（多语言JSON字符串）- 用于显示（commodity: 外卖表优先, 其他: 核心表）
	Num                       float64 // TTPOS 数量（仅针对 commodity 类型）
	TtposName                 string  // TTPOS 修饰符名称（多语言JSON字符串）- 用于标识，始终来自核心表
	TtposProductPackageUuid   uint64  // TTPOS 商品套餐UUID（关联ttpos_product_package.uuid）
	TtposFlavorProductBomUuid uint64  // TTPOS 规格商品物料UUID（commodity类型对应product_package_group_item.product_bom_uuid）
	TtposFlavorName           string  // TTPOS 规格名称（commodity类型使用）
	TtposErpCode              string  // TTPOS 修饰符 ERP 编码(来自 ProductBom.ErpCode 或 ProductSauce.ErpCode)
	TtposPrice                float64 // TTPOS 店内价格(来自对应表的price字段)

	// 分类信息
	TtposCategoryUuid       uint64 // TTPOS分类UUID(关联 ttpos_product_category.uuid)
	TtposCategoryName       string // TTPOS分类名称
	TtposParentCategoryUuid uint64 // TTPOS父分类UUID
	TtposParentCategoryName string // TTPOS父分类名称
}
