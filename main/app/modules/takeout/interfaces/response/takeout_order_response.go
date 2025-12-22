package response

// TakeoutOrderResp 订单响应
type TakeoutOrderResp struct {
	Uuid             uint64 `json:"uuid"`               // 订单UUID
	Platform         string `json:"platform"`           // 平台名称
	ShortOrderNumber string `json:"short_order_number"` // 短订单号
	OrderState       int    `json:"order_state"`        // 订单状态
	Cutlery          int    `json:"cutlery"`            // 是否需要餐具
	// 订单时间
	OrderTime          int64 `json:"order_time"`           // 订单时间 (创建时间)
	SubmitTime         int64 `json:"submit_time"`          // 提交时间 (提交时间，支付时间)
	AcceptedTime       int64 `json:"accepted_time"`        // 接单时间
	CompletedTime      int64 `json:"completed_time"`       // 完成时间
	EstimatedReadyTime int64 `json:"estimated_ready_time"` // 预计完成时间
	MaxReadyTime       int64 `json:"max_ready_time"`       // 最大准备时间
	// 订单异常
	IsAbnormal     int    `json:"is_abnormal"`     // 是否异常
	AbnormalDetail string `json:"abnormal_detail"` // 异常详情
	// 订单金额
	Subtotal      int64 `json:"subtotal"`        // 商品小计
	DeliveryFee   int64 `json:"delivery_fee"`    // 配送费
	SmallOrderFee int64 `json:"small_order_fee"` // 小订单费
	EaterPayment  int64 `json:"eater_payment"`   // 顾客实付
	TotalAmount   int64 `json:"total_amount"`    // 订单总金额
	// 货币信息
	CurrencyCode   string `json:"currency_code"`   // 货币代码
	CurrencySymbol string `json:"currency_symbol"` // 货币符号
	PaymentType    string `json:"payment_type"`    // 支付类型
	OrderType      string `json:"order_type"`      // 订单类型
	// 订单商品
	TotalItems int                        `json:"total_items"`         // 总商品数量
	Items      []TakeoutOrderItemResp     `json:"items"`               // 订单商品列表
	Receiver   TakeoutOrderReceiverResp   `json:"receiver,omitempty"`  // 收货人信息 (联系人信息)
	Campaigns  []TakeoutOrderCampaignResp `json:"campaigns,omitempty"` // 活动信息
}

// TakeoutOrderReceiverResp 订单收货人响应
type TakeoutOrderReceiverResp struct {
	ReceiverName        string `json:"receiver_name,omitempty"`        // 收货人姓名
	ReceiverPhones      string `json:"receiver_phones,omitempty"`      // 收货人电话
	UnitNumber          string `json:"unit_number,omitempty"`          // 单元号/门牌号
	DeliveryInstruction string `json:"delivery_instruction,omitempty"` // 配送说明
	Address             string `json:"address,omitempty"`              // 详细地址
}

// TakeoutOrderCampaignResp 订单活动响应
type TakeoutOrderCampaignResp struct {
	Uuid           uint64 `json:"uuid"`
	CampaignName   string `json:"campaign_name"`
	CampaignType   string `json:"campaign_type"`
	DeductedAmount int64  `json:"deducted_amount"` // 折扣金额(元)
}

// TakeoutOrderItemResp 订单商品响应
type TakeoutOrderItemResp struct {
	Uuid             uint64 `json:"uuid"`
	PlatformItemId   string `json:"platform_item_id"`
	PlatformItemName string `json:"platform_item_name"`
	Quantity         int    `json:"quantity"`
	Price            int64  `json:"price"`
	Tax              int64  `json:"tax"`
	Specifications   string `json:"specifications"`
	IsMapped         int    `json:"is_mapped"`
}

// TakeoutOrderListItemResp 订单响应
type TakeoutOrderListItemResp struct {
	Uuid             uint64 `json:"uuid"`               // 订单UUID
	Platform         string `json:"platform"`           // 平台名称
	ShortOrderNumber string `json:"short_order_number"` // 短订单号
	OrderState       int    `json:"order_state"`        // 订单状态
	IsAbnormal       int    `json:"is_abnormal"`        // 是否异常
	TotalItems       int    `json:"total_items"`        // 总商品数量
}

// TakeoutOrderListResp 订单列表响应
type TakeoutOrderListResp struct {
	List []*TakeoutOrderListItemResp `json:"list"`
	Meta Meta                        `json:"meta"`
}

// TakeoutSettingsResp 配置响应
type TakeoutSettingsResp struct {
	Uuid       uint64 `json:"uuid"`
	Platform   string `json:"platform"`
	AutoAccept bool   `json:"auto_accept"`
	MaxAmount  int64  `json:"max_amount"`
}
