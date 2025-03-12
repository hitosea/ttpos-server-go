package req

import "ttpos-server-go/app/dto"

// ProductionListReq 送厨商品列表
type ProductionListReq struct {
	dto.PageReq // 分页参数
}

// ProductUuid 送厨商品Uuid
type ProductUuid struct {
	ProductUuid uint64 `json:"product_uuid"` // 送厨商品ID
}
