package req

import "ttpos-server-go/app/dto"

// ProductionListReq 送厨商品列表
type ProductionListReq struct {
	dto.PageReq // 分页参数
}

// ProductionListByCategoryReq 送厨商品列表
type ProductionListByCategoryReq struct {
	CategoryUuid uint64 `form:"category_uuid,default=0"  json:"category_uuid"` // 分类Uuid，0-全部
	dto.PageReq         // 分页参数
}

// ProductUuid 送厨商品Uuid
type ProductUuid struct {
	ProductUuid uint64 `json:"product_uuid"` // 送厨商品ID
}

// SaleBillUuid 按订单查看送厨商品，确认整单取消时传递销售账单Uuid
type SaleBillUuid struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单Uuid
}
