package model

// TakeoutOrder 外卖订单表（多平台）
type TakeoutOrder struct {
	BaseModel

	// 关联外卖平台订单
	TakeoutOrderUuid string `gorm:"column:takeout_order_uuid" json:"takeout_order_uuid"`

	// 平台信息
	Platform          string `gorm:"column:platform" json:"platform"`
	PlatformOrderId   string `gorm:"column:platform_order_id" json:"platform_order_id"`
	ShortOrderNumber  string `gorm:"column:short_order_number" json:"short_order_number"`
	MerchantId        string `gorm:"column:merchant_id" json:"merchant_id"`
	PartnerMerchantId string `gorm:"column:partner_merchant_id" json:"partner_merchant_id"`

	// 订单状态
	OrderState     int    `gorm:"column:order_state" json:"order_state"`
	IsAbnormal     int    `gorm:"column:is_abnormal" json:"is_abnormal"`
	AbnormalDetail string `gorm:"column:abnormal_detail;type:text" json:"abnormal_detail"`
	StockStatus    int    `gorm:"column:stock_status" json:"stock_status"`

	// 价格信息（单位：分）
	Subtotal          int64 `gorm:"column:subtotal" json:"subtotal"`
	DeliveryFee       int64 `gorm:"column:delivery_fee" json:"delivery_fee"`
	SmallOrderFee     int64 `gorm:"column:small_order_fee" json:"small_order_fee"`
	TotalAmount       int64 `gorm:"column:total_amount" json:"total_amount"`
	EaterPayment      int64 `gorm:"column:eater_payment" json:"eater_payment"`
	PlatformDiscount  int64 `gorm:"column:platform_discount" json:"platform_discount"`
	MerchantDiscount  int64 `gorm:"column:merchant_discount" json:"merchant_discount"`
	BasketPromo       int64 `gorm:"column:basket_promo" json:"basket_promo"`
	Tax               int64 `gorm:"column:tax" json:"tax"`
	MerchantChargeFee int64 `gorm:"column:merchant_charge_fee" json:"merchant_charge_fee"`

	// 货币信息
	CurrencyCode     string `gorm:"column:currency_code" json:"currency_code"`
	CurrencySymbol   string `gorm:"column:currency_symbol" json:"currency_symbol"`
	CurrencyExponent int    `gorm:"column:currency_exponent" json:"currency_exponent"`

	// 支付信息
	PaymentType string `gorm:"column:payment_type" json:"payment_type"`

	// 订单时间
	OrderTime          int64 `gorm:"column:order_time" json:"order_time"`
	SubmitTime         int64 `gorm:"column:submit_time" json:"submit_time"`
	ScheduledTime      int64 `gorm:"column:scheduled_time" json:"scheduled_time"`
	AcceptedTime       int64 `gorm:"column:accepted_time" json:"accepted_time"`
	CompletedTime      int64 `gorm:"column:completed_time" json:"completed_time"`
	RejectedTime       int64 `gorm:"column:rejected_time" json:"rejected_time"`
	EstimatedReadyTime int64 `gorm:"column:estimated_ready_time" json:"estimated_ready_time"`
	MaxReadyTime       int64 `gorm:"column:max_ready_time" json:"max_ready_time"`

	// 其他通用信息
	Cutlery           int    `gorm:"column:cutlery" json:"cutlery"`
	OrderType         string `gorm:"column:order_type" json:"order_type"`
	OrderAcceptedType string `gorm:"column:order_accepted_type" json:"order_accepted_type"`
	IsMexEditOrder    int    `gorm:"column:is_mex_edit_order" json:"is_mex_edit_order"`
	MembershipId      string `gorm:"column:membership_id" json:"membership_id"`
	DriverEta         int64  `gorm:"column:driver_eta" json:"driver_eta"`

	// 平台特定数据（JSON 格式）
	PlatformData string `gorm:"column:platform_data;type:mediumtext" json:"platform_data"`

	// 完整原始数据（JSON 格式）
	RawData string `gorm:"column:raw_data;type:mediumtext" json:"raw_data"`

	// 操作信息
	AcceptedBy       uint64 `gorm:"column:accepted_by" json:"accepted_by"`
	RejectedBy       uint64 `gorm:"column:rejected_by" json:"rejected_by"`
	RejectReasonCode string `gorm:"column:reject_reason_code" json:"reject_reason_code"`
	RejectReason     string `gorm:"column:reject_reason" json:"reject_reason"`

	// 关联字表结构
	TakeoutOrderItems     []TakeoutOrderItem     `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderReceiver  *TakeoutOrderReceiver  `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderCampaigns []TakeoutOrderCampaign `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
}

func (*TakeoutOrder) TableName() string {
	return "ttpos_takeout_order"
}

func (o *TakeoutOrder) SetTakeoutOrderReceiver(receiver *TakeoutOrderReceiver) {
	o.TakeoutOrderReceiver = receiver
}

func (o *TakeoutOrder) SetTakeoutOrderCampaigns(campaigns []*TakeoutOrderCampaign) {
	if campaigns == nil {
		o.TakeoutOrderCampaigns = nil
		return
	}
	if len(campaigns) == 0 {
		o.TakeoutOrderCampaigns = nil
		return
	}
	o.TakeoutOrderCampaigns = make([]TakeoutOrderCampaign, len(campaigns))
	for i, campaign := range campaigns {
		o.TakeoutOrderCampaigns[i] = *campaign
	}
}
