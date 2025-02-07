package cashier_resp

import "ttpos-server-go/app/dto"

// 商品
type Product struct {
	ID    uint               `json:"id"`    // 商品ID
	Name  dto.LocaleResponse `json:"name"`  // 商品名称
	Image string             `json:"image"` // 商品图片
	Unit  dto.LocaleResponse `json:"unit"`  // 商品单位
}

// 商品列表响应
type ProductListWithPaginationResp struct {
	List []Product        `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}
