package req

import "ttpos-server-go/app/dto"

var DeskReqMessage = map[string]string{
	"uuid.required":            "桌台uuid不能为空",
	"sale_bill_uuid.required":  "销售账单UUID不能为空",
	"sale_order_uuid.required": "销售订单UUID不能为空",
}

// 桌台列表查询
type DeskListReq struct {
	dto.PageReq // 分页参数
}

// 桌台信息
type DeskInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 桌台uuid
}

// 关闭桌台
type DeskCloseReq struct {
	Uuid     uint64 `form:"uuid" binding:"required"` // 桌台uuid
	Reason   string `form:"reason"`                  // 关闭原因
	Password string `form:"password"`                // 密码 后台开启的时候才传
}

// 关闭桌台订单
type DeskOrderCloseReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"  binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"  binding:"required"` // 销售订单UUID
	Reason        string `form:"reason"`                              // 关闭原因
	Password      string `form:"password"`                            // 密码 后台开启的时候才传
}

// 自助餐顾客类型
type DeskBuffetCustomerType struct {
	Uuid    uint64 `json:"uuid"`     // 自助餐顾客类型uuid
	MealNum *uint  `json:"meal_num"` // 就餐人数
}

// 桌台订单创建
type DeskOrderCreateReq struct {
	DeskUuid            uint64                   `json:"desk_uuid"`             // 桌台uuid, 必填
	IsBuffet            *bool                    `json:"is_buffet"`             // 是否是自助餐: false-否, true-是
	MealNum             *uint                    `json:"meal_num"`              // 就餐人数: 非自助餐时, 最小为0, 最大为999, 自助餐时为0
	BuffetUuids         []uint64                 `json:"buffet_uuids"`          // 自助餐uuid列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1, 最大为2
	BuffetCustomerTypes []DeskBuffetCustomerType `json:"buffet_customer_types"` // 自助餐顾客类型列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1
	Remark              string                   `json:"remark"`                // 备注: 最小空字符串,最大50字符
}
