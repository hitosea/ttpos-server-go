package grab

// ==================== Grab 订单相关结构体 ====================

// GrabOrderWebhook Grab 订单 Webhook 数据结构
//
// 说明：Grab 的订单数据通常是直接的 GrabOrder 对象。
// 此结构用于统一处理不同场景下的订单数据：
// 1. 直接订单格式：GrabOrder（最常见）
// 2. Webhook 包装格式：包含 order 字段（某些特定通知）
//
// 转换器会自动识别格式并填充此结构
type GrabOrderWebhook struct {
	MerchantID        string     `json:"merchantID"`        // 商户ID（Grab分配）
	PartnerMerchantID string     `json:"partnerMerchantID"` // 合作商户ID（商家自己的ID）
	OrderID           string     `json:"orderID"`           // 订单ID
	State             string     `json:"state"`             // 订单状态
	DriverETA         *string    `json:"driverETA"`         // 骑手预计到达时间（仅配送订单）
	Code              string     `json:"code,omitempty"`    // 错误码（如果有）
	Message           string     `json:"message,omitempty"` // 错误消息（如果有）
	Order             *GrabOrder `json:"order"`             // 完整订单数据
}

// GrabOrder Grab 订单详细信息
type GrabOrder struct {
	OrderID              string                    `json:"orderID"`
	ShortOrderNumber     string                    `json:"shortOrderNumber"`
	MerchantID           string                    `json:"merchantID"`
	PartnerMerchantID    string                    `json:"partnerMerchantID"`
	PaymentType          string                    `json:"paymentType"`
	Cutlery              bool                      `json:"cutlery"`
	OrderTime            string                    `json:"orderTime"`
	SubmitTime           string                    `json:"submitTime"`
	CompleteTime         string                    `json:"completeTime,omitempty"`
	ScheduledTime        string                    `json:"scheduledTime,omitempty"`
	OrderState           string                    `json:"orderState"`
	Currency             GrabCurrency              `json:"currency"`
	FeatureFlags         GrabFeatureFlags          `json:"featureFlags"`
	Items                []GrabOrderItem           `json:"items"`
	Campaigns            []GrabCampaign            `json:"campaigns,omitempty"`
	Promos               []GrabPromo               `json:"promos,omitempty"`
	Price                GrabPrice                 `json:"price"`
	DineIn               *GrabDineIn               `json:"dineIn,omitempty"`
	Receiver             *GrabReceiver             `json:"receiver,omitempty"`
	OrderReadyEstimation *GrabOrderReadyEstimation `json:"orderReadyEstimation,omitempty"`
	MembershipID         string                    `json:"membershipID,omitempty"`
	Discounts            []GrabDiscount            `json:"discounts,omitempty"`
	Payments             []GrabPayment             `json:"payments,omitempty"`
}

// GrabFeatureFlags Grab 订单特性标记
type GrabFeatureFlags struct {
	OrderAcceptedType string `json:"orderAcceptedType"`
	OrderType         string `json:"orderType"`
	IsMexEditOrder    bool   `json:"isMexEditOrder"`
}

// GrabOrderItem Grab 订单商品
type GrabOrderItem struct {
	ID                    string                     `json:"id"`
	GrabItemID            string                     `json:"grabItemID"`
	Quantity              int                        `json:"quantity"`
	Price                 int64                      `json:"price"`
	Tax                   int64                      `json:"tax"`
	Specifications        string                     `json:"specifications,omitempty"`
	OutOfStockInstruction *GrabOutOfStockInstruction `json:"outOfStockInstruction,omitempty"`
	Modifiers             []GrabOrderModifier        `json:"modifiers,omitempty"`
}

// GrabOutOfStockInstruction Grab 商品缺货处理指示
type GrabOutOfStockInstruction struct {
	Title                 string `json:"title"`
	InstructionType       string `json:"instructionType"`
	ReplacementItemID     string `json:"replacementItemID,omitempty"`
	ReplacementGrabItemID string `json:"replacementGrabItemID,omitempty"`
}

// GrabOrderModifier Grab 订单修饰符（区别于菜单修饰符）
type GrabOrderModifier struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Tax      int64  `json:"tax"`
	Quantity int    `json:"quantity"`
}

// GrabCampaign Grab 营销活动
type GrabCampaign struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	CampaignNameForMex string        `json:"campaignNameForMex,omitempty"`
	Level              string        `json:"level"`
	Type               string        `json:"type"`
	UsageCount         int           `json:"usageCount"`
	MexFundedRatio     int           `json:"mexFundedRatio"`
	DeductedAmount     int64         `json:"deductedAmount"`
	DeductedPart       string        `json:"deductedPart"`
	AppliedItemIDs     []string      `json:"appliedItemIDs,omitempty"`
	FreeItem           *GrabFreeItem `json:"freeItem,omitempty"`
}

// GrabFreeItem Grab 赠品
type GrabFreeItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}

// GrabPromo Grab 促销优惠
type GrabPromo struct {
	Code             string `json:"code"`
	Description      string `json:"description"`
	Name             string `json:"name"`
	PromoAmount      int64  `json:"promoAmount"`
	MexFundedRatio   int    `json:"mexFundedRatio"`
	MexFundedAmount  int64  `json:"mexFundedAmount"`
	TargetedPrice    int64  `json:"targetedPrice"`
	PromoAmountInMin int64  `json:"promoAmountInMin"`
}

// GrabPrice Grab 订单价格信息
type GrabPrice struct {
	Subtotal          int64                `json:"subtotal"`
	Tax               int64                `json:"tax"`
	MerchantChargeFee int64                `json:"merchantChargeFee"`
	GrabFundPromo     int64                `json:"grabFundPromo"`
	MerchantFundPromo int64                `json:"merchantFundPromo"`
	BasketPromo       int64                `json:"basketPromo"`
	DeliveryFee       int64                `json:"deliveryFee"`
	SmallOrderFee     int64                `json:"smallOrderFee"`
	EaterPayment      int64                `json:"eaterPayment"`
	Total             int64                `json:"total"`
	MerchantEarning   *GrabMerchantEarning `json:"merchantEarning,omitempty"`
}

// GrabMerchantEarning Grab 商家收益信息
type GrabMerchantEarning struct {
	Revenue         int64 `json:"revenue"`
	NetEarning      int64 `json:"netEarning"`
	MexFundDiscount int64 `json:"mexFundDiscount"`
	Commission      int64 `json:"commission"`
}

// GrabDineIn Grab 堂食信息
type GrabDineIn struct {
	TableID    string `json:"tableID"`
	EaterCount int    `json:"eaterCount"`
}

// GrabReceiver Grab 收货人信息
type GrabReceiver struct {
	Name           string              `json:"name"`
	Phones         string              `json:"phones"`
	Address        GrabAddress         `json:"address"`
	VirtualContact *GrabVirtualContact `json:"virtualContact,omitempty"`
}

// GrabAddress Grab 地址信息
type GrabAddress struct {
	UnitNumber          string           `json:"unitNumber,omitempty"`
	DeliveryInstruction string           `json:"deliveryInstruction,omitempty"`
	PoiSource           string           `json:"poiSource,omitempty"`
	PoiID               string           `json:"poiID,omitempty"`
	Address             string           `json:"address"`
	Postcode            int              `json:"postcode"`
	Coordinates         *GrabCoordinates `json:"coordinates,omitempty"`
}

// GrabCoordinates Grab 地理坐标
type GrabCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GrabVirtualContact Grab 虚拟联系方式
type GrabVirtualContact struct {
	PhoneNumber string `json:"phoneNumber"`
	PIN         int    `json:"PIN"`
	ExpiredAt   string `json:"expiredAt"`
	Status      string `json:"status"`
}

// GrabOrderReadyEstimation Grab 订单准备时间预估
type GrabOrderReadyEstimation struct {
	AllowChange             bool   `json:"allowChange"`
	EstimatedOrderReadyTime string `json:"estimatedOrderReadyTime"`
	MaxOrderReadyTime       string `json:"maxOrderReadyTime"`
	NewOrderReadyTime       string `json:"newOrderReadyTime,omitempty"`
}

// GrabDiscount Grab 折扣信息
type GrabDiscount struct {
	Code                 string   `json:"code"`
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	DeductAmountInMin    int64    `json:"deductAmountInMin"`
	Level                string   `json:"level"`
	Type                 string   `json:"type"`
	MexFundedAmountInMin int64    `json:"mexFundedAmountInMin"`
	AppliedItemIDs       []string `json:"appliedItemIDs,omitempty"`
}

// GrabPayment Grab 支付信息
type GrabPayment struct {
	Method      string `json:"method"`
	FundingType string `json:"fundingType"`
	AmountInMin int64  `json:"amountInMin"`
}
