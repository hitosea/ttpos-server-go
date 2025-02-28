package req

import (
	"errors"
	"ttpos-server-go/app/dto"
)

var DeskReqMessage = map[string]string{
	"uuid.required":            "桌台uuid不能为空",
	"sale_bill_uuid.required":  "销售账单UUID不能为空",
	"sale_order_uuid.required": "销售订单UUID不能为空",
}

// DeskListReq 桌台列表查询
type DeskListReq struct {
	dto.PageReq // 分页参数
}

// DeskInfoReq 桌台信息
type DeskInfoReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 桌台uuid
}

// DeskCloseReq 关闭桌台
type DeskCloseReq struct {
	Uuid     uint64 `form:"uuid" binding:"required"` // 桌台uuid
	Reason   string `form:"reason"`                  // 关闭原因
	Password string `form:"password"`                // 密码 后台开启的时候才传
}

// DeskOrderCloseReq 关闭桌台订单
type DeskOrderCloseReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"  binding:"required"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"  binding:"required"` // 销售订单UUID
	Reason        string `form:"reason"`                              // 关闭原因
	Password      string `form:"password"`                            // 密码 后台开启的时候才传
}

// DeskBuffetCustomerType 自助餐顾客类型
type DeskBuffetCustomerType struct {
	Uuid    uint64 `json:"uuid"`     // 自助餐顾客类型uuid
	MealNum *uint  `json:"meal_num"` // 就餐人数
}

// DeskOrderCreateReq 桌台订单创建
type DeskOrderCreateReq struct {
	DeskUuid            uint64                   `json:"desk_uuid"`             // 桌台uuid, 必填
	IsBuffet            *bool                    `json:"is_buffet"`             // 是否是自助餐: false-否, true-是
	MealNum             *uint                    `json:"meal_num"`              // 就餐人数: 非自助餐时, 最小为0, 最大为999, 自助餐时为0
	BuffetUuids         []uint64                 `json:"buffet_uuids"`          // 自助餐uuid列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1, 最大为2
	BuffetCustomerTypes []DeskBuffetCustomerType `json:"buffet_customer_types"` // 自助餐顾客类型列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1
	Remark              string                   `json:"remark"`                // 备注: 最小空字符串,最大50字符
}

func (req *DeskOrderCreateReq) ValidateCreateDeskOrderReq() error {
	if req.DeskUuid == 0 {
		return errors.New("桌台uuid不能为0")
	}
	if req.IsBuffet == nil {
		return errors.New("是否是自助餐不能为空")
	}
	if req.MealNum == nil {
		return errors.New("就餐人数不能为空")
	}
	if !*req.IsBuffet {
		if *req.MealNum < 1 || *req.MealNum > 999 {
			return errors.New("就餐人数不能小于1或大于999")
		}
	}
	if *req.IsBuffet {
		if len(req.BuffetUuids) < 1 || len(req.BuffetUuids) > 2 {
			return errors.New("自助餐uuid列表不能小于1或大于2")
		}
		if len(req.BuffetCustomerTypes) == 0 {
			return errors.New("自助餐顾客类型列表不能为空")
		}
	}
	return nil
}

// BindDeskReq 平板绑定/换绑桌台请求参数
type BindDeskReq struct {
	DeviceId    string `json:"device_id" binding:"required"`
	Brand       string `json:"brand"`
	DeskUuid    uint64 `json:"desk_uuid" binding:"required"`
	OldDeskUuid uint64 `json:"old_desk_uuid"`
	Remark      string `json:"remark"`
}
