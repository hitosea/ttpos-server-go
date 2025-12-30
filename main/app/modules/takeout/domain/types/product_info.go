package types

// ProductInfo 商品信息（名称）
type ProductInfo struct {
	Name         string // 商品名称（多语言JSON字符串）- 用于显示，优先外卖表
	TtposName    string // TTPOS 商品名称（多语言JSON字符串）- 用于标识，始终来自核心表
	TtposErpCode string // TTPOS 商品 ERP 编码(来自 ProductBom.ErpCode)
}
