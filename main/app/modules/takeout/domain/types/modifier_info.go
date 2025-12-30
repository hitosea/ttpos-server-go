package types

// ModifierInfo 修饰符信息（名称和数量）
type ModifierInfo struct {
	Name             string  // 修饰符名称（多语言JSON字符串）- 用于显示（commodity: 外卖表优先, 其他: 核心表）
	Num              float64 // TTPOS 数量（仅针对 commodity 类型）
	TtposName        string  // TTPOS 修饰符名称（多语言JSON字符串）- 用于标识，始终来自核心表
	TtposProductUuid uint64  // TTPOS 商品UUID（关联ttpos_product_package.uuid）
	TtposFlavorUuid  uint64  // TTPOS 规格UUID（commodity类型对应product_package_group_item.product_bom_uuid）
	TtposFlavorName  string  // TTPOS 规格名称（commodity类型使用）
	TtposErpCode     string  // TTPOS 修饰符 ERP 编码(来自 ProductBom.ErpCode 或 ProductSauce.ErpCode)
}
