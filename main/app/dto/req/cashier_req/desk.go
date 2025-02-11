package cashier_req

import "ttpos-server-go/app/dto"

// 桌台列表查询
type DeskListReq struct {
	dto.PageReq // 分页参数
}

type DeskInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 桌台uuid
}

// 桌台订单创建
type DeskOrderCreateReq struct {
	DeskUuid uint64 `json:"desk_uuid" binding:"required"`                                // 桌台uuid
	IsBuffet bool   `json:"is_buffet" binding:"required"`                                // 是否是自助餐: false-否, true-是
	MealNum  uint   `json:"meal_num" binding:"required_if=IsBuffet false,min=1,max=999"` // 餐数: 非自助餐时必填, 最小1, 最大999
	Remark   string `json:"remark" binding:"max=50"`                                     // 备注: 非必填, 最大50字符
}

// 桌台订单创建错误信息
var DeskOrderCreateReqMessage = map[string]string{
	"desk_uuid.required":   "桌台uuid不能为空",
	"is_buffet.required":   "是否是自助餐不能为空",
	"meal_num.required_if": "餐数不能为空",
	"meal_num.min":         "餐数不能小于1",
	"meal_num.max":         "餐数不能大于999",
	"remark.max":           "备注不能大于50字符",
}
