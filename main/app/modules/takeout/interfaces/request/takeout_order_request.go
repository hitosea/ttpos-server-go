package request

// TakeoutOrderListReq 订单列表请求
type TakeoutOrderListReq struct {
	PageNo    int    `json:"page_no" binding:"required"`
	PageSize  int    `json:"page_size" binding:"required"`
	Platform  string `json:"platform"`   // 平台筛选: grab,foodpanda,lineman (空=全部)
	Status    int    `json:"status"`     // 0=全部,1=待接单,2=已接单,3=制作中,4=已完成,5=已拒单
	StartTime int64  `json:"start_time"` // 开始时间
	EndTime   int64  `json:"end_time"`   // 结束时间
	Search    string `json:"search"`     // 搜索关键词
}

// TakeoutOrderDetailReq 订单详情请求
type TakeoutOrderDetailReq struct {
	OrderUuid uint64 `json:"order_uuid" binding:"required"`
}

// TakeoutOrderSyncReq 同步订单请求
type TakeoutOrderSyncReq struct {
	Platform string                 `json:"platform" binding:"required"` // grab,foodpanda,lineman
	RawData  map[string]interface{} `json:"raw_data" binding:"required"` // 原始订单数据
}

// TakeoutOrderAcceptReq 接单请求
type TakeoutOrderAcceptReq struct {
	OrderUuid uint64 `json:"order_uuid" binding:"required"`
}

// TakeoutOrderRejectReq 拒单请求
type TakeoutOrderRejectReq struct {
	OrderUuid        uint64 `json:"order_uuid" binding:"required"`
	RejectReasonCode string `json:"reject_reason_code" binding:"required"`
}

// TakeoutSettingsGetReq 获取配置请求
type TakeoutSettingsGetReq struct {
	Platform string `json:"platform" binding:"required"` // grab,foodpanda,lineman
}

// TakeoutSettingsSaveReq 配置保存请求
type TakeoutSettingsSaveReq struct {
	Platform   string `json:"platform" binding:"required"` // grab,foodpanda,lineman
	AutoAccept bool   `json:"auto_accept"`
	MaxAmount  int64  `json:"max_amount"` // 单位：分
}
