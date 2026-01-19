package resp

import "ttpos-server-go/app/dto"

type SoldOut struct {
	LocaleProductName    dto.LocaleResponse `json:"locale_product_name"`     // 商品名称
	ProductBomUuid       uint64             `json:"product_bom_uuid"`        // 商品规格Uuid
	LocaleProductBomName dto.LocaleResponse `json:"locale_product_bom_name"` // 商品规格名称
}

// SoldOutPaginationResp 售罄商品列表响应
type SoldOutPaginationResp struct {
	List []SoldOut        `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// SoldOutSettingsResp 沽清设置响应
type SoldOutSettingsResp struct {
	Settings []SoldOutSetting `json:"settings"`
}

// SoldOutSetting 单个规格的沽清设置
type SoldOutSetting struct {
	ProductBomUuid   uint64  `json:"product_bom_uuid"`   // 商品规格UUID
	HasBomCard       bool    `json:"has_bom_card"`       // 是否关联了成本卡
	UseBomCardStock  bool    `json:"use_bom_card_stock"` // 是否使用成本卡库存
	BomCardStockNum  float64 `json:"bom_card_stock_num"` // 成本卡库存数量
	IsSoldOut        bool    `json:"is_sold_out"`        // 是否售罄
	IsOpenStock      bool    `json:"is_open_stock"`      // 是否开启可售库存
	SellableQuantity float64 `json:"sellable_quantity"`  // 可售数量
}
