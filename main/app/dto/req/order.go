package req

import "ttpos-server-go/app/dto"

// 订单列表查询
type GetOrderListReq struct {
	dto.PageReq          // 分页参数
	OrderNo       string `form:"order_no"`          // 订单编号
	DateType      int    `form:"date_type"`         // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	QueryTimeType []uint `form:"query_time_type[]"` // 查询时间类型 1-开台时间、2-支付时间
	QueryTimes    []uint `form:"query_times[]"`     // 日期范围 [开始时间戳, 结束时间戳]
	Status        int    `form:"status"`            // 账单状态, -1=全都、 0=待付款、1=已完成、2=已取消
	BillType      int    `form:"bill_type"`         // 账单类型, -1=全都、 0=Desk桌台订单、1=OrderingFood点餐订单
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
