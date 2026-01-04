package response

import (
	"ttpos-server-go/app/dto"
)

// TakeoutOrderCancelCheckResp 检查订单是否可取消响应
type TakeoutOrderCancelCheckResp struct {
	CanCancel             bool                       `json:"can_cancel"`              // 是否可以取消
	NonCancellationReason string                     `json:"non_cancellation_reason"` // 不可取消原因（当 can_cancel=false 时）
	CancelReasons         []TakeoutOrderCancelReason `json:"cancel_reasons"`          // 可选的取消原因列表
}

// TakeoutOrderCancelReason 取消原因
type TakeoutOrderCancelReason struct {
	Code   string `json:"code"`   // 原因代码
	Reason string `json:"reason"` // 原因描述
}

// TakeoutOrderResp 订单响应
type TakeoutOrderResp struct {
	Uuid             uint64 `json:"uuid"`               // 订单UUID
	Platform         string `json:"platform"`           // 平台名称
	ShortOrderNumber string `json:"short_order_number"` // 短订单号
	OrderState       int    `json:"order_state"`        // 订单状态 0=待接单,1=已接单配餐中, 2=待骑手接单, 3=骑手配送中, 4=已完成, 5=已拒单
	Cutlery          int    `json:"cutlery"`            // 是否需要餐具
	// 订单时间
	OrderTimes TakeoutOrderTimesResp `json:"order_times"` // 订单时间
	// 订单异常
	IsAbnormal     int    `json:"is_abnormal"`     // 是否异常
	AbnormalDetail string `json:"abnormal_detail"` // 异常详情
	// 订单金额
	Price TakeoutOrderPriceResp `json:"price"` // 订单金额
	// 货币信息
	Currencies TakeoutOrderCurrencyResp `json:"currencies"` // 货币信息
	// 订单优惠
	Discounts TakeoutOrderDiscountsResp `json:"discounts"` // 订单优惠
	// 支付类型
	PaymentType string                     `json:"payment_type"` // 支付类型
	OrderType   string                     `json:"order_type"`   // 订单类型  DELIVERY - 平台配送，RESTAURANT_DELIVERY - 商家配送，TAKEAWAY - 自提/打包，DINEIN - 店内就餐
	TotalItems  int                        `json:"total_items"`  // 总商品数量
	Items       []TakeoutOrderItemResp     `json:"items"`        // 订单商品列表
	Campaigns   []TakeoutOrderCampaignResp `json:"campaigns"`    // 活动信息
	Promos      []TakeoutOrderPromoResp    `json:"promos"`       // 促销信息
	Receiver    TakeoutOrderReceiverResp   `json:"receiver"`     // 收货人信息 (联系人信息)
}

type TakeoutOrderDiscountsResp struct {
	PlatformDiscount float64 `json:"platform_discount"` // 平台优惠
	MerchantDiscount float64 `json:"merchant_discount"` // 商户优惠
	BasketPromo      float64 `json:"basket_promo"`      // 篮子优惠
	Tax              float64 `json:"tax"`               // 税费
}

type TakeoutOrderTimesResp struct {
	OrderTime          int64 `json:"order_time"`           // 下单时间
	SubmitTime         int64 `json:"submit_time"`          // 提交时间 (提交时间，支付时间)
	AcceptedTime       int64 `json:"accepted_time"`        // 接单时间
	CompletedTime      int64 `json:"completed_time"`       // 完成时间
	EstimatedReadyTime int64 `json:"estimated_ready_time"` // 预计完成时间
	MaxReadyTime       int64 `json:"max_ready_time"`       // 最大准备时间
}

type TakeoutOrderPriceResp struct {
	Subtotal          float64 `json:"subtotal"`            // 商品小计
	DeliveryFee       float64 `json:"delivery_fee"`        // 配送费
	SmallOrderFee     float64 `json:"small_order_fee"`     // 小订单费
	EaterPayment      float64 `json:"eater_payment"`       // 顾客实付
	PlatformDiscount  float64 `json:"platform_discount"`   // 平台优惠
	MerchantDiscount  float64 `json:"merchant_discount"`   // 商户优惠
	BasketPromo       float64 `json:"basket_promo"`        // 购物车促销
	Tax               float64 `json:"tax"`                 // 税费
	MerchantChargeFee float64 `json:"merchant_charge_fee"` // 商户收取费用
}

// CurrencyResp 货币响应
type TakeoutOrderCurrencyResp struct {
	CurrencyCode     string `json:"currency_code"`     // 货币代码
	CurrencySymbol   string `json:"currency_symbol"`   // 货币符号
	CurrencyExponent int    `json:"currency_exponent"` // 货币小数位数
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
	CampaignName   string `json:"campaign_name"`   // 活动名称
	CampaignType   string `json:"campaign_type"`   // 活动类型
	DeductedAmount int64  `json:"deducted_amount"` // 折扣金额(元)
}

// TakeoutOrderPromoResp 订单促销响应
type TakeoutOrderPromoResp struct {
	Uuid             uint64 `json:"uuid"`
	PromoCode        string `json:"promo_code"`        // 促销代码
	PromoName        string `json:"promo_name"`        // 促销名称
	PromoDescription string `json:"promo_description"` // 促销描述
	PromoAmount      int64  `json:"promo_amount"`      // 促销金额
	MexFundedRatio   int    `json:"mex_funded_ratio"`  // 商户承担比例
}

// TakeoutOrderItemResp 订单商品响应
type TakeoutOrderItemResp struct {
	Uuid           uint64                 `json:"uuid"`
	ItemName       dto.LocaleResponse     `json:"item_name"`
	Quantity       int                    `json:"quantity"`
	Price          float64                `json:"price"`
	Tax            float64                `json:"tax"`
	Specifications string                 `json:"specifications"`
	Modifiers      string                 `json:"modifiers"`
	IsPackage      bool                   `json:"is_package"` // 是否套餐
	SubItems       []TakeoutOrderItemResp `json:"sub_items"`  // 子商品列表
}

type TakeoutOrderItemModifierResp struct {
	Uuid         uint64             `json:"uuid"`
	ModifierName dto.LocaleResponse `json:"modifier_name"`
	Quantity     int                `json:"quantity"`
	Price        float64            `json:"price"`
}

// TakeoutOrderListItemResp 订单响应
type TakeoutOrderListItemResp struct {
	Uuid             uint64  `json:"uuid"`               // 订单UUID
	Platform         string  `json:"platform"`           // 平台名称
	ShortOrderNumber string  `json:"short_order_number"` // 短订单号
	OrderState       int     `json:"order_state"`        // 订单状态: 0=待接单,10=已接单配餐中, 20=待骑手接单, 30=骑手配送中, 40=已完成, 50=已拒单, 60=已取消
	IsAbnormal       int     `json:"is_abnormal"`        // 是否异常
	Subtotal         float64 `json:"subtotal"`           // 小计金额
	TotalItems       int     `json:"total_items"`        // 总商品数量
}

// TakeoutOrderListResp 订单列表响应
type TakeoutOrderListResp struct {
	List []*TakeoutOrderListItemResp `json:"list"`
	Meta Meta                        `json:"meta"`
}

// TakeoutSettingsResp 配置响应
type TakeoutSettingsResp struct {
	Uuid       uint64  `json:"uuid"`
	Platform   string  `json:"platform"`
	AutoAccept bool    `json:"auto_accept"`
	MaxAmount  float64 `json:"max_amount"`
}
