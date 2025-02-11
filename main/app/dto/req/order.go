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

// 创建桌位订单
type CreateDeskOrderReq struct {
	DeskUuid uint64 `json:"desk_uuid" binding:"required"`                               // 桌位uuid
	IsBuffet bool   `json:"is_buffet" binding:"required"`                               // 是否自助餐
	MealNum  uint   `json:"meal_num" binding:"required_if=IsBuffet true,min=1,max=999"` // 用餐人数: 自助餐必填, 范围1-999
	Remark   string `json:"remark" binding:"max=50"`                                    // 备注: 非必填, 最多50字符
}

// 创建桌位订单错误信息
var CreateDeskOrderReqMessage = map[string]string{
	"desk_uuid.required":   "桌位uuid不能为空",
	"is_buffet.required":   "是否自助餐不能为空",
	"meal_num.required_if": "用餐人数不能为空",
	"meal_num.min":         "用餐人数不能小于1",
	"meal_num.max":         "用餐人数不能大于999",
	"remark.max":           "备注不能超过50个字符",
}
