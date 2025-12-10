// Package grab 定义 GrabFood API v1.1.3 交互的数据结构
//
// 仅保留常量定义（OrderState*）以兼容现有代码
package grab

// ============================================================================
// Submit Order Webhook (POST /partner/v1/orders)
// Deprecated: 使用 SDK grabfood.SubmitOrderRequest
// ============================================================================

// SubmitOrderRequest Grab 提交订单请求
// Deprecated: 使用 SDK grabfood.SubmitOrderRequest
type SubmitOrderRequest struct {
	OrderID              string                `json:"orderID"`                        // Grab 订单 ID
	ShortOrderNumber     string                `json:"shortOrderNumber"`               // 短单号
	MerchantID           string                `json:"merchantID"`                     // 商户 ID
	PartnerMerchantID    string                `json:"partnerMerchantID"`              // 合作商户 ID
	PaymentType          string                `json:"paymentType"`                    // 支付方式: CASH, CASHLESS
	Cutlery              bool                  `json:"cutlery"`                        // 是否需要餐具
	OrderTime            string                `json:"orderTime"`                      // 下单时间 (RFC3339)
	SubmitTime           string                `json:"submitTime"`                     // 提交时间
	CompleteTime         string                `json:"completeTime"`                   // 预计完成时间
	ScheduledTime        string                `json:"scheduledTime"`                  // 预约时间
	OrderState           string                `json:"orderState"`                     // 订单状态
	Currency             Currency              `json:"currency"`                       // 货币信息
	FeatureFlags         FeatureFlags          `json:"featureFlags"`                   // 功能标识
	Items                []OrderItem           `json:"items"`                          // 订单项
	Campaigns            []Campaign            `json:"campaigns"`                      // 活动信息
	Promos               []Promo               `json:"promos"`                         // 促销信息
	Price                OrderPrice            `json:"price"`                          // 价格信息
	DineIn               *DineInInfo           `json:"dineIn,omitempty"`               // 堂食信息
	Receiver             *Receiver             `json:"receiver"`                       // 收货人
	OrderReadyEstimation *OrderReadyEstimation `json:"orderReadyEstimation,omitempty"` // 准备时间估算
	MembershipID         string                `json:"membershipID,omitempty"`         // 会员 ID
	IsMexEditOrder       bool                  `json:"isMexEditOrder"`                 // 是否为商户编辑订单
}

// OrderItem 订单项
type OrderItem struct {
	ID                    string                 `json:"id"`                              // 商品 ID
	GrabItemID            string                 `json:"grabItemID"`                      // Grab 商品 ID
	Quantity              int                    `json:"quantity"`                        // 数量
	Price                 int64                  `json:"price"`                           // 价格 (最小单位)
	Tax                   int64                  `json:"tax"`                             // 税额
	Specifications        string                 `json:"specifications"`                  // 规格说明
	Modifiers             []Modifier             `json:"modifiers"`                       // 修改项
	OutOfStockInstruction *OutOfStockInstruction `json:"outOfStockInstruction,omitempty"` // 缺货处理
}

// Modifier 修改项/配料
type Modifier struct {
	ID       string `json:"id"`       // 修改项 ID
	Quantity int    `json:"quantity"` // 数量
	Price    int64  `json:"price"`    // 价格
}

// OutOfStockInstruction 缺货处理指令
type OutOfStockInstruction struct {
	Type string `json:"type"` // 处理类型: REMOVE_ITEM, CONTACT_EATER, CANCEL_ORDER
}

// Currency 货币信息
type Currency struct {
	Code     string `json:"code"`     // 货币代码: THB, SGD, etc.
	Symbol   string `json:"symbol"`   // 符号
	Exponent int    `json:"exponent"` // 小数位数
}

// FeatureFlags 功能标识
type FeatureFlags struct {
	OrderAcceptedType string `json:"orderAcceptedType"` // 接单类型: AUTO, MANUAL
	IsMexEditOrder    bool   `json:"isMexEditOrder"`    // 是否为商户编辑订单
	IsSplitOrder      bool   `json:"isSplitOrder"`      // 是否为拆单
}

// Campaign 活动信息
type Campaign struct {
	CampaignID         string `json:"campaignID"`         // 活动 ID
	CampaignNameForMex string `json:"campaignNameForMex"` // 商户端活动名
	Name               string `json:"name"`               // 活动名称
	Level              string `json:"level"`              // 级别: ORDER, ITEM
	Type               string `json:"type"`               // 类型
	UsedCount          int    `json:"usedCount"`          // 使用次数
	DeductedAmount     int64  `json:"deductedAmount"`     // 扣减金额
	MexFundedRatio     int    `json:"mexFundedRatio"`     // 商户承担比例
}

// Promo 促销信息
type Promo struct {
	Code            string `json:"code"`            // 促销码
	Description     string `json:"description"`     // 描述
	PromoAmount     int64  `json:"promoAmount"`     // 促销金额
	MexFundedRatio  int    `json:"mexFundedRatio"`  // 商户承担比例
	MexFundedAmount int64  `json:"mexFundedAmount"` // 商户承担金额
}

// OrderPrice 价格信息
type OrderPrice struct {
	Subtotal          int64 `json:"subtotal"`          // 商品小计
	Tax               int64 `json:"tax"`               // 税额
	MerchantChargeFee int64 `json:"merchantChargeFee"` // 商户收取费用
	GrabFundPromo     int64 `json:"grabFundPromo"`     // Grab 承担促销
	MerchantFundPromo int64 `json:"merchantFundPromo"` // 商户承担促销
	Total             int64 `json:"total"`             // 总金额
}

// DineInInfo 堂食信息
type DineInInfo struct {
	TableID    string `json:"tableID"`    // 桌号
	EaterCount int    `json:"eaterCount"` // 用餐人数
}

// Receiver 收货人信息
type Receiver struct {
	Name           string           `json:"name"`           // 姓名
	Phone          string           `json:"phone"`          // 电话
	Address        *DeliveryAddress `json:"address"`        // 地址
	VirtualContact *VirtualContact  `json:"virtualContact"` // 虚拟联系方式
}

// DeliveryAddress 配送地址
type DeliveryAddress struct {
	UnitNumber          string       `json:"unitNumber"`          // 单元号
	DeliveryInstruction string       `json:"deliveryInstruction"` // 配送说明
	PoiSource           string       `json:"poiSource"`           // POI 来源
	PoiID               string       `json:"poiID"`               // POI ID
	Address             string       `json:"address"`             // 地址
	Postcode            string       `json:"postcode"`            // 邮编
	Coordinates         *Coordinates `json:"coordinates"`         // 坐标
}

// Coordinates 坐标
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// VirtualContact 虚拟联系方式
type VirtualContact struct {
	Phone      string `json:"phone"`      // 虚拟电话
	ExpiryTime string `json:"expiryTime"` // 过期时间
}

// OrderReadyEstimation 准备时间估算
type OrderReadyEstimation struct {
	AllowChange        bool   `json:"allowChange"`        // 是否允许修改
	EstimatedReadyTime string `json:"estimatedReadyTime"` // 预计准备完成时间
	MaxReadyTime       string `json:"maxReadyTime"`       // 最晚准备时间
	NewReadyTime       string `json:"newReadyTime"`       // 新准备时间
}

// SubmitOrderResponse 提交订单响应
type SubmitOrderResponse struct {
	// Grab 期望返回 HTTP 200 表示成功接收
}

// ============================================================================
// Push Order State Webhook (PUT /partner/v1/orders/{orderID}/state)
// ============================================================================

// OrderStateRequest 订单状态变更请求
type OrderStateRequest struct {
	OrderID           string `json:"orderID"`           // 订单 ID
	MerchantID        string `json:"merchantID"`        // 商户 ID
	PartnerMerchantID string `json:"partnerMerchantID"` // 合作商户 ID
	OrderState        string `json:"state"`             // 新状态
	DriverETA         *int   `json:"driverETA"`         // 骑手预计到达时间(分钟)
	Code              int    `json:"code"`              // 状态码
	Message           string `json:"message"`           // 消息
}

// OrderState 订单状态枚举
const (
	OrderStatePending   = "PENDING"
	OrderStateAccepted  = "ACCEPTED"
	OrderStatePreparing = "PREPARING"
	OrderStateReady     = "READY"
	OrderStateCollected = "COLLECTED"
	OrderStateDelivered = "DELIVERED"
	OrderStateCompleted = "COMPLETED"
	OrderStateCancelled = "CANCELLED"
	OrderStateFailed    = "FAILED"
	OrderStateBillPaid  = "BILL_PAID" // STO
	OrderStateRefunded  = "REFUNDED"  // STO
)
