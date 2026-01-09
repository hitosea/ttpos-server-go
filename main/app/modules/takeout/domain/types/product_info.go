package types

// ProductInfo 商品信息（名称、价格、分类）
type ProductInfo struct {
	Name         string  // 商品名称（多语言JSON字符串）- 用于显示，优先外卖表
	TtposName    string  // TTPOS 商品名称（多语言JSON字符串）- 用于标识，始终来自核心表
	TtposErpCode string  // TTPOS 商品 ERP 编码(来自 ProductBom.ErpCode)
	TtposPrice   float64 // TTPOS 店内价格(来自 ttpos_product_package.price)

	// 分类信息
	TtposCategoryUuid       uint64 // TTPOS分类UUID(关联 ttpos_product_category.uuid)
	TtposCategoryName       string // TTPOS分类名称
	TtposParentCategoryUuid uint64 // TTPOS父分类UUID
	TtposParentCategoryName string // TTPOS父分类名称
}
