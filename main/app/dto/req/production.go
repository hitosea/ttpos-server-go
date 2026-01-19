package req

import "ttpos-server-go/app/dto"

// ProductionListReq 送厨商品列表
type ProductionListReq struct {
	Mode        uint `form:"mode" json:"mode"` // 状态 0-传菜模式，查看待传菜的列表 ; 1-制作模式，查看待制作的列表
	dto.PageReq      // 分页参数
}

// ProductionListByCategoryReq 送厨商品列表
type ProductionListByCategoryReq struct {
	Mode         uint   `form:"mode" json:"mode"`                              // 状态 0-传菜模式，查看待传菜的列表 ; 1-制作模式，查看待制作的列表
	CategoryUuid uint64 `form:"category_uuid,default=0"  json:"category_uuid"` // 分类Uuid，0-全部
	dto.PageReq         // 分页参数
}

// ProductUuid 送厨商品Uuid
type ProductUuid struct {
	ProductUuid uint64 `json:"product_uuid"` // 送厨商品ID
}

type HistoryReq struct {
	Mode uint `form:"mode" json:"mode"` // 模式 0-传菜历史 ; 1-制作历史
}

type FinishReq struct {
	ProductUuid uint64 `json:"product_uuid"` // 送厨商品ID
	Mode        uint   `json:"mode"`         // 模式 0-传菜完成 ; 1-制作完成

	ProductUuids              []uint64 `json:"product_uuids"`                // 送厨商品ID列表 v2.9新增"一键完成"
	ConfirmReturnProductUuids []uint64 `json:"confirm_return_product_uuids"` // 确认退菜送厨商品ID列表 v2.9新增"一键完成"
}

type RecoveryReq struct {
	ProductUuid uint64 `json:"product_uuid"` // 送厨商品ID
	Mode        uint   `json:"mode"`         // 模式 0-传菜历史恢复 ; 1-制作历史恢复
}

// ConfirmReturnAllReq 按订单查看送厨商品，确认整单取消时传递销售账单Uuid和外卖订单Uuid
type ConfirmReturnAllReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid"`     // 销售账单Uuid
	TakeoutOrderUuid uint64 `json:"takeout_order_uuid"` // 外卖订单Uuid
}
