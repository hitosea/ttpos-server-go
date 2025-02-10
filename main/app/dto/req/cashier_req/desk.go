package cashier_req

import "ttpos-server-go/app/dto"

// 桌台列表查询
type DeskListReq struct {
	dto.PageReq // 分页参数
}

type DeskInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 桌台uuid
}
