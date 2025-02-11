package resp

import "ttpos-server-go/app/dto"

type Buffet struct {
	Uuid               uint64             `json:"uuid"`                 // 自助餐UUID
	Price              float64            `json:"price"`                // 价格
	IsLimitTime        bool               `json:"is_limit_time"`        // 是否限时
	CanCombined        bool               `json:"can_combined"`         // 是否可组合
	NonOrderingTime    uint               `json:"non_ordering_time"`    // 不可下单时间（分钟）
	ReminderOrderTime  uint               `json:"reminder_order_time"`  // 提醒下单时间（分钟）
	LocaleName         dto.LocaleResponse `json:"locale_name"`          // 自助餐名称
	BuffetCustomerType BuffetCustomerType `json:"buffet_customer_type"` // 自助餐客户类型
}

type BuffetCustomerType struct {
	Uuid       uint64             `json:"uuid"`        // 自助餐客户类型UUID
	Price      float64            `json:"price"`       // 价格
	LocaleName dto.LocaleResponse `json:"locale_name"` // 自助餐客户类型名称
}

// BuffetListPaginationResp 自助餐列表响应
type BuffetListPaginationResp struct {
	List []Buffet         `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}
