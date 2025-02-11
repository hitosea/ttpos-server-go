package req

import "ttpos-server-go/app/dto"

// 订单列表查询
type GetOrderListReq struct {
	dto.PageReq // 分页参数
}

// 订单信息查询
type GetOrderInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 订单uuid
}
