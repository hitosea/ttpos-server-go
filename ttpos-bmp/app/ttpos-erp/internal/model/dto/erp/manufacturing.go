package erp

// Bom 结构体，表示物料清单（Bill of Materials）
type Bom struct {
	Item      string    `json:"item"`       // 商品编码
	Company   string    `json:"company"`    // 公司名称
	Quantity  float32   `json:"quantity"`   // 数量
	Uom       string    `json:"uom"`        // 计量单位
	IsActive  bool      `json:"is_active"`  // 是否激活
	IsDefault bool      `json:"is_default"` // 是否默认
	Items     []BomItem `json:"items"`      // BOM项目列表
}

// BomItem 结构体，表示BOM项目明细
type BomItem struct {
	ItemCode string  `json:"item_code"` // 商品编码
	Rate     float32 `json:"rate"`      // 比率
	Qty      float32 `json:"qty"`       // 数量
	Uom      string  `json:"uom"`       // 计量单位
}
