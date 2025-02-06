package resp

import "ttpos-server-go/app/dto"

type Product struct {
	ID    uint               `json:"id"`    // 商品ID
	Name  dto.LocaleResponse `json:"name"`  // 商品名称
	Image string             `json:"image"` // 商品图片
	Price float64            `json:"price"` // 商品价格
	Unit  dto.LocaleResponse `json:"unit"`  // 商品单位
}
