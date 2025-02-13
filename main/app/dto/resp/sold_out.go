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
