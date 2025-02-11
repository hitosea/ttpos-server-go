package req

import "ttpos-server-go/app/dto"

// 商品列表查询
type ProductListReq struct {
	dto.PageReq // 分页参数
}
