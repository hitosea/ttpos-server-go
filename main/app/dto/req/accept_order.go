package req

import "ttpos-server-go/app/dto"

// H5OrderListReq 接单列表
type H5OrderListReq struct {
	dto.PageReq           // 分页参数
	Status         *uint  `form:"status" binding:"required,oneof=0 1"` // 状态:0-待处理；1-已处理
	DeskRegionUuid uint64 `form:"desk_region_uuid"`                    // 桌台区域Uuid
	OrderType      int    `form:"order_type,default=-1"`               // 订单类型:-1=全部、0=桌台扫码订单、1=门店扫码堂食订单(会员端堂食订单)
}

type H5OrderDetailReq struct {
	H5OrderUuid uint64 `form:"h5_order_uuid" binding:"required"` // h5订单Uuid
}
