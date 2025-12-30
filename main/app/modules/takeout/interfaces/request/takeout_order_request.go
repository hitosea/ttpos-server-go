package request

// TakeoutIntegrationEvent 门店集成状态变更事件
type TakeoutIntegrationEvent struct {
	ShopUuid           uint64 `json:"shopUuid"`           // TTPOS 店铺 UUID
	ProviderName       string `json:"providerName"`       // 供应商名称 grab,foodpanda,lineman
	ProviderShopStatus string `json:"providerShopStatus"` // 供应商门店状态 SYNCING, ACTIVE, INACTIVE
	ProviderMerchantId string `json:"providerMerchantId"` // 供应商商户 ID
	UpdatedAt          int64  `json:"updatedAt"`          // 更新时间戳
}

// TakeoutOrderEvent 供应商订单更新事件
type TakeoutOrderEvent struct {
	Action       string `json:"action"`       // create, status_update, cancel
	ProviderName string `json:"providerName"` // grab
	ShopUuid     string `json:"shopUuid"`     // TTPOS 店铺 UUID
	OrderUuid    string `json:"orderUuid"`    // 订单 UUID
	OrderId      string `json:"orderId"`      // 平台订单 ID
	MerchantId   string `json:"merchantId"`   // 商户 ID
	Status       string `json:"status"`       // 当前状态 PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED
	Timestamp    int64  `json:"timestamp"`    // 事件时间戳
}

// TakeoutOrderListReq 订单列表请求
type TakeoutOrderListReq struct {
	PageReq
	Platform  string `json:"platform" form:"platform"`          // 平台筛选: grab,lineman (空=全部) (默认: 空)
	Status    int    `json:"status" form:"status" default:"-1"` // -1=全部, 0=待接单,1=已接单配餐中, 2=待骑手接单, 3=骑手配送中, 4=已完成, 5=已拒单 (默认: -1)
	StartTime int64  `json:"start_time" form:"start_time"`      // 开始时间 (默认: 0)
	EndTime   int64  `json:"end_time" form:"end_time"`          // 结束时间 (默认: 0)
	Search    string `json:"search" form:"search"`              // 搜索关键词 (默认: 空)
	IsHistory bool   `json:"is_history" form:"is_history"`      // 是否历史订单: true=历史订单, false=未完成订单 (默认: false)
}

// TakeoutOrderDetailReq 订单详情请求
type TakeoutOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"`
}

// TakeoutOrderSyncReq 同步订单请求
type TakeoutOrderSyncReq struct {
	Platform string                 `json:"platform" binding:"required"` // grab,foodpanda,lineman
	RawData  map[string]interface{} `json:"raw_data" binding:"required"` // 原始订单数据
}

// TakeoutOrderAcceptReq 接单请求
type TakeoutOrderAcceptReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

// TakeoutOrderRejectReq 拒单请求
type TakeoutOrderRejectReq struct {
	Uuid             uint64 `json:"uuid" binding:"required"`
	RejectReasonCode string `json:"reject_reason_code"`
}

// TakeoutOrderCallRiderReq 呼叫骑手请求
type TakeoutOrderCallRiderReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 订单UUID
}

// TakeoutOrderCheckCancelableReq 检查订单是否可取消请求
type TakeoutOrderCheckCancelableReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 订单UUID
}

// TakeoutOrderCancelReq 取消订单请求
type TakeoutOrderCancelReq struct {
	Uuid       uint64 `json:"uuid" binding:"required"`        // 订单UUID
	ReasonCode string `json:"reason_code" binding:"required"` // 取消原因码
}

// TakeoutOrderPrintReq 打印订单请求
type TakeoutOrderPrintReq struct {
	Uuid      uint64 `json:"uuid" binding:"required"`       // 订单UUID
	PrintLang string `json:"print_lang" binding:"required"` // 语言
}

// TakeoutSettingsGetReq 获取配置请求
type TakeoutSettingsGetReq struct {
	Platform string `json:"platform" binding:"required"` // grab,foodpanda,lineman
}

// TakeoutSettingsSaveReq 配置保存请求
type TakeoutSettingsSaveReq struct {
	Platform   string  `json:"platform" binding:"required"` // grab,foodpanda,lineman
	AutoAccept bool    `json:"auto_accept"`
	MaxAmount  float64 `json:"max_amount"` // 单位：分
}
