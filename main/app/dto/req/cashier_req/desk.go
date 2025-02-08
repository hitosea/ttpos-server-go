package cashier_req

import "ttpos-server-go/app/dto"

// 桌台列表查询
type DeskListReq struct {
	dto.PageReq // 分页参数
}
