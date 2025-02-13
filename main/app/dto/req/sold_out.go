package req

import "ttpos-server-go/app/dto"

// SoldOutListReq 沽清列表
type SoldOutListReq struct {
	dto.PageReq // 分页参数
}

type SoldOutItem struct {
	ProductBomUuid uint64 `json:"product_bom_uuid" binding:"required"` // 商品规格Uuid
	IsSoldOut      *bool  `json:"is_sold_out" binding:"required"`      // 是否售罄：true-是；false-否
}

type AddSoldOutReq struct {
	SoldOutData []SoldOutItem `json:"sold_out_data" binding:"required,gt=0,dive,required"` // 设置售罄数据
}

type CancelSoldOutReq struct {
	ProductBomUuid uint64 `json:"product_bom_uuid" binding:"required"` // 商品规格Uuid
}
