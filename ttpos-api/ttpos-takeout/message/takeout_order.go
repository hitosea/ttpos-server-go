package message

// TakeoutOrder 统一外卖订单模型
// 以 Grab 订单格式为基准，扩展支持 Lineman 字段
type TakeoutOrder struct {
	// ========== 基础字段（Grab 和 Lineman 通用） ==========

	// OrderID 订单 ID
	// Grab: orderID
	// Lineman: orderId
	OrderID string `json:"orderID"`

	// ShortOrderNumber 短订单号
	// Grab: shortOrderNumber
	// Lineman: orderShortCode
	ShortOrderNumber string `json:"shortOrderNumber"`

	// MerchantID 商户 ID（平台侧）
	// Grab: merchantID
	// Lineman: storeId (作为 provider merchant ID)
	MerchantID string `json:"merchantID"`

	// PartnerMerchantID 合作商户 ID（TTPOS 侧）
	// Grab: partnerMerchantID
	// Lineman: partnerId
	PartnerMerchantID string `json:"partnerMerchantID,omitempty"`

	// PaymentType 支付方式
	// Grab: paymentType (CASH, CASHLESS)
	// Lineman: 不直接提供，默认为 CASH
	PaymentType string `json:"paymentType"`

	// OrderTime 下单时间 (RFC3339)
	// Grab: orderTime
	// Lineman: orderAcceptedTime
	OrderTime string `json:"orderTime"`

	// Items 订单商品列表
	Items []TakeoutOrderItem `json:"items"`

	// Price 价格信息
	Price TakeoutOrderPrice `json:"price"`

	// ========== Grab 特有字段 ==========

	// Cutlery 是否需要餐具（Grab 特有）
	Cutlery *bool `json:"cutlery,omitempty"`

	// SubmitTime 提交时间（Grab 特有）
	SubmitTime *string `json:"submitTime,omitempty"`

	// CompleteTime 完成时间（Grab 特有）
	CompleteTime *string `json:"completeTime,omitempty"`

	// ScheduledTime 预约时间（Grab 特有）
	ScheduledTime *string `json:"scheduledTime,omitempty"`

	// OrderState 订单状态（Grab 特有）
	OrderState *string `json:"orderState,omitempty"`

	// Currency 货币信息（Grab 特有）
	Currency *TakeoutCurrency `json:"currency,omitempty"`

	// FeatureFlags 功能标识（Grab 特有）
	FeatureFlags *TakeoutFeatureFlags `json:"featureFlags,omitempty"`

	// Campaigns 活动信息（Grab 特有）
	Campaigns []TakeoutCampaign `json:"campaigns,omitempty"`

	// Promos 促销信息（Grab 特有）
	Promos []TakeoutPromo `json:"promos,omitempty"`

	// DineIn 堂食信息（Grab 特有）
	DineIn *TakeoutDineIn `json:"dineIn,omitempty"`

	// Receiver 收货人信息（Grab 特有）
	Receiver *TakeoutReceiver `json:"receiver,omitempty"`

	// OrderReadyEstimation 准备时间估算（Grab 特有）
	OrderReadyEstimation *TakeoutOrderReadyEstimation `json:"orderReadyEstimation,omitempty"`

	// MembershipID 会员 ID（Grab 特有）
	MembershipID *string `json:"membershipID,omitempty"`

	// ========== 扩展字段（Lineman 等其他平台需要） ==========

	// AdditionalProperties 订单附加属性（扩展字段）
	// Grab: 无此字段（但 SDK 有 AdditionalProperties map）
	// Lineman: additionalItems[] -> 映射为 AdditionalProperties
	// 如：ไม่รับช้อนส้อมพลาสติก（不需要塑料餐具）
	AdditionalProperties []TakeoutAdditionalProperty `json:"additionalProperties,omitempty"`
}

// TakeoutOrderItem 订单商品项
type TakeoutOrderItem struct {
	// ID 商品 ID
	// Grab: id
	// Lineman: id
	ID string `json:"id"`

	// GrabItemID Grab 商品 ID（Grab 特有）
	GrabItemID *string `json:"grabItemID,omitempty"`

	// Quantity 商品数量
	Quantity int `json:"quantity"`

	// Price 商品价格
	// Grab: price (int64, 最小单位，如泰铢分)
	// Lineman: unitPrice (float64, 泰铢)
	Price float64 `json:"price"`

	// Tax 税额（Grab 特有）
	Tax *float64 `json:"tax,omitempty"`

	// Specifications 规格说明（Grab 特有）
	Specifications *string `json:"specifications,omitempty"`

	// Modifiers 修改项/配料
	// Grab: modifiers
	// Lineman: properties (结构不同，需转换)
	Modifiers []TakeoutModifier `json:"modifiers,omitempty"`

	// OutOfStockInstruction 缺货处理指令（Grab 特有）
	OutOfStockInstruction *TakeoutOutOfStockInstruction `json:"outOfStockInstruction,omitempty"`
}

// TakeoutModifier 商品修改项/配料
type TakeoutModifier struct {
	// ID 修改项 ID
	ID string `json:"id"`

	// Quantity 数量
	Quantity int `json:"quantity"`

	// Price 价格
	Price float64 `json:"price"`

	// ========== Lineman 扩展（从 properties 转换） ==========

	// Values 选项值列表（Lineman properties 转换而来）
	Values []TakeoutModifierValue `json:"values,omitempty"`
}

// TakeoutModifierValue 修改项值（Lineman properties 专用）
type TakeoutModifierValue struct {
	ID    string  `json:"id"`
	Price float64 `json:"price"`
}

// TakeoutOrderPrice 价格信息
type TakeoutOrderPrice struct {
	// Subtotal 商品小计
	// Grab: subtotal (int64, 最小单位)
	// Lineman: restaurantRevenue (float64, 泰铢)
	Subtotal float64 `json:"subtotal"`

	// Tax 税额（Grab 特有）
	Tax *float64 `json:"tax,omitempty"`

	// MerchantChargeFee 商户收取费用（Grab 特有）
	MerchantChargeFee *float64 `json:"merchantChargeFee,omitempty"`

	// DeliveryFee 配送费（Grab 特有）
	DeliveryFee *float64 `json:"deliveryFee,omitempty"`

	// GrabFundPromo Grab 承担促销（Grab 特有）
	GrabFundPromo *float64 `json:"grabFundPromo,omitempty"`

	// MerchantFundPromo 商户承担促销（Grab 特有）
	MerchantFundPromo *float64 `json:"merchantFundPromo,omitempty"`

	// EaterPayment 用户实付金额（Grab 特有）
	// Lineman 用 restaurantRevenue 对应此字段
	EaterPayment *float64 `json:"eaterPayment,omitempty"`

	// Total 总金额（Grab 特有）
	Total *float64 `json:"total,omitempty"`
}

// TakeoutCurrency 货币信息（Grab 特有）
type TakeoutCurrency struct {
	Code     string `json:"code"`     // 货币代码: THB, SGD
	Symbol   string `json:"symbol"`   // 符号
	Exponent int    `json:"exponent"` // 小数位数
}

// TakeoutFeatureFlags 功能标识（Grab 特有）
type TakeoutFeatureFlags struct {
	OrderAcceptedType string `json:"orderAcceptedType"` // 接单类型: AUTO, MANUAL
	IsMexEditOrder    bool   `json:"isMexEditOrder"`    // 是否为商户编辑订单
	OrderType         string `json:"orderType"`         // 订单类型: DineIn, Delivery
}

// TakeoutCampaign 活动信息（Grab 特有）
type TakeoutCampaign struct {
	CampaignID         string  `json:"campaignID"`
	CampaignNameForMex string  `json:"campaignNameForMex"`
	Name               string  `json:"name"`
	Level              string  `json:"level"`
	Type               string  `json:"type"`
	UsedCount          int     `json:"usedCount"`
	DeductedAmount     float64 `json:"deductedAmount"`
	MexFundedRatio     int     `json:"mexFundedRatio"`
}

// TakeoutPromo 促销信息（Grab 特有）
type TakeoutPromo struct {
	Code            string  `json:"code"`
	Description     string  `json:"description"`
	PromoAmount     float64 `json:"promoAmount"`
	MexFundedRatio  int     `json:"mexFundedRatio"`
	MexFundedAmount float64 `json:"mexFundedAmount"`
}

// TakeoutDineIn 堂食信息（Grab 特有）
type TakeoutDineIn struct {
	TableID    string `json:"tableID"`
	EaterCount int    `json:"eaterCount"`
}

// TakeoutReceiver 收货人信息（Grab 特有）
type TakeoutReceiver struct {
	Name           string                  `json:"name"`
	Phone          string                  `json:"phone"`
	Address        *TakeoutDeliveryAddress `json:"address"`
	VirtualContact *TakeoutVirtualContact  `json:"virtualContact"`
}

// TakeoutDeliveryAddress 配送地址（Grab 特有）
type TakeoutDeliveryAddress struct {
	UnitNumber          string              `json:"unitNumber"`
	DeliveryInstruction string              `json:"deliveryInstruction"`
	PoiSource           string              `json:"poiSource"`
	PoiID               string              `json:"poiID"`
	Address             string              `json:"address"`
	Postcode            string              `json:"postcode"`
	Coordinates         *TakeoutCoordinates `json:"coordinates"`
}

// TakeoutCoordinates 坐标（Grab 特有）
type TakeoutCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// TakeoutVirtualContact 虚拟联系方式（Grab 特有）
type TakeoutVirtualContact struct {
	Phone      string `json:"phone"`
	ExpiryTime string `json:"expiryTime"`
}

// TakeoutOrderReadyEstimation 准备时间估算（Grab 特有）
type TakeoutOrderReadyEstimation struct {
	AllowChange        bool   `json:"allowChange"`
	EstimatedReadyTime string `json:"estimatedReadyTime"`
	MaxReadyTime       string `json:"maxReadyTime"`
	NewReadyTime       string `json:"newReadyTime,omitempty"`
}

// TakeoutOutOfStockInstruction 缺货处理指令（Grab 特有）
type TakeoutOutOfStockInstruction struct {
	Type string `json:"type"` // REMOVE_ITEM, CONTACT_EATER, CANCEL_ORDER
}

// TakeoutAdditionalProperty 订单附加属性（扩展字段）
// Lineman: additionalItems[] -> name
// Grab: 无直接对应，扩展字段
type TakeoutAdditionalProperty struct {
	Name string `json:"name"` // 如：ไม่รับช้อนส้อมพลาสติก
}
