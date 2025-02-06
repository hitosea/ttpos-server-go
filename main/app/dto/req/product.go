package req

import "ttpos-server-go/app/dto"

// 商品列表查询
type ProductListReq struct {
	dto.PageReq        // 分页参数
	Keyword     string `form:"product_id" binding:"omitempty"`  // 搜索关键字
	CategoryID  uint   `form:"category_id" binding:"omitempty"` // 搜索分类ID
}
