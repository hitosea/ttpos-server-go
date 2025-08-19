package erp

// Bom 结构体，表示物料清单（Bill of Materials）
type Bom struct {
	Item      string    `json:"item,omitempty"`       // 商品编码
	Company   string    `json:"company,omitempty"`    // 公司名称
	Quantity  float64   `json:"quantity,omitempty"`   // 数量
	Uom       string    `json:"uom,omitempty"`        // 计量单位
	IsActive  bool      `json:"is_active,omitempty"`  // 是否激活
	IsDefault bool      `json:"is_default,omitempty"` // 是否默认
	Items     []BomItem `json:"items,omitempty"`      // BOM项目列表
}

// BomItem 结构体，表示BOM项目明细
type BomItem struct {
	ItemCode string  `json:"item_code,omitempty"` // 商品编码
	Rate     float64 `json:"rate,omitempty"`      // 比率
	Qty      float64 `json:"qty,omitempty"`       // 数量
	Uom      string  `json:"uom,omitempty"`       // 计量单位
}
