package model

import (
	"encoding/json"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/errors"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/i18n"
)

// TakeoutOrder 外卖订单表（多平台）
type TakeoutOrder struct {
	BaseModel

	// 关联外卖平台订单
	TakeoutOrderUuid string `gorm:"column:takeout_order_uuid" json:"takeout_order_uuid"` // 外卖平台订单UUID

	// 平台信息
	Platform           string `gorm:"column:platform" json:"platform"`                         // grab, lineman
	PlatformOrderId    string `gorm:"column:platform_order_id" json:"platform_order_id"`       // 平台订单ID
	PlatformOrderState string `gorm:"column:platform_order_state" json:"platform_order_state"` // 平台订单状态 (Grab: NEW 待接单, ACCEPTED 已接单, ALLOCATING 待骑手接单, DRIVER_ALLOCATED 骑手已分配, DRIVER_ARRIVED 骑手已到店, COLLECTED 已取餐, DELIVERING 配送中, DELIVERED 已送达, COMPLETED 已完成, CANCELLED 已取消, REJECTED 已拒单, FAILED 失败, REFUNDED 已退款)
	ShortOrderNumber   string `gorm:"column:short_order_number" json:"short_order_number"`     // 短单号
	MerchantId         string `gorm:"column:merchant_id" json:"merchant_id"`                   // 商户ID	(Grab: merchantID)
	PartnerMerchantId  string `gorm:"column:partner_merchant_id" json:"partner_merchant_id"`   // 合作伙伴商户ID (Grab: partnerMerchantID)

	// 订单状态
	OrderState     int    `gorm:"column:order_state" json:"order_state"`                   // : 0=待接单, 10=已接单配餐中, 20=待骑手接单, 30=骑手配送中, 40=已完成, 50=已拒单, 60=已取消
	IsAbnormal     int    `gorm:"column:is_abnormal" json:"is_abnormal"`                   // 是否异常: 0=正常, 1=异常
	AbnormalDetail string `gorm:"column:abnormal_detail;type:text" json:"abnormal_detail"` // 异常详情
	StockStatus    int    `gorm:"column:stock_status" json:"stock_status"`                 // 库存状态: 1=充足, 2=不足

	// 价格信息（单位：元）
	Subtotal          float64 `gorm:"column:subtotal" json:"subtotal"`                       // 小计金额
	DeliveryFee       float64 `gorm:"column:delivery_fee" json:"delivery_fee"`               // 配送费
	SmallOrderFee     float64 `gorm:"column:small_order_fee" json:"small_order_fee"`         // 小单费用
	EaterPayment      float64 `gorm:"column:eater_payment" json:"eater_payment"`             // 顾客实付
	PlatformDiscount  float64 `gorm:"column:platform_discount" json:"platform_discount"`     // 平台优惠
	MerchantDiscount  float64 `gorm:"column:merchant_discount" json:"merchant_discount"`     // 商户优惠
	BasketPromo       float64 `gorm:"column:basket_promo" json:"basket_promo"`               // 购物车优惠
	Tax               float64 `gorm:"column:tax" json:"tax"`                                 // 税费
	MerchantChargeFee float64 `gorm:"column:merchant_charge_fee" json:"merchant_charge_fee"` // 商户收取费用
	PlatformTotal     float64 `gorm:"column:platform_total" json:"platform_total"`           // 平台结算总额 (subtotal + merchant_charge_fee - merchant_discount)

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
	OrderType         string `gorm:"column:order_type" json:"order_type"`                   // 订单类型: DELIVERY, RESTAURANT_DELIVERY, TAKEAWAY，DINEIN
	OrderAcceptedType string `gorm:"column:order_accepted_type" json:"order_accepted_type"` // 接单类型: AUTO, MANUAL
	IsMexEditOrder    int    `gorm:"column:is_mex_edit_order" json:"is_mex_edit_order"`
	MembershipId      string `gorm:"column:membership_id" json:"membership_id"`
	DriverEta         int64  `gorm:"column:driver_eta" json:"driver_eta"`

	// 完整原始数据（JSON 格式）
	RawData string `gorm:"column:raw_data;type:mediumtext" json:"raw_data"`

	// 订单额外属性信息（JSON 格式）
	AdditionalProperties string `gorm:"column:additional_properties" json:"additional_properties"`

	// 操作信息
	AcceptedBy        uint64 `gorm:"column:accepted_by" json:"accepted_by"`
	StaffShiftLogUuid uint64 `gorm:"column:staff_shift_log_uuid" json:"staff_shift_log_uuid"`
	ErpPosInvoiceResp string `gorm:"column:erp_pos_invoice_resp;type:text" json:"erp_pos_invoice_resp"` // ERP POS Invoice响应数据(JSON格式)
	RejectedBy        uint64 `gorm:"column:rejected_by" json:"rejected_by"`
	RejectReasonCode  string `gorm:"column:reject_reason_code" json:"reject_reason_code"`
	RejectReason      string `gorm:"column:reject_reason" json:"reject_reason"`

	// 关联字表结构
	TakeoutOrderItems      []TakeoutOrderItem      `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderReceiver   *TakeoutOrderReceiver   `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderCampaigns  []TakeoutOrderCampaign  `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderMaterials  []TakeoutOrderMaterial  `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderPromos     []TakeoutOrderPromo     `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
	TakeoutOrderUpdateLogs []TakeoutOrderUpdateLog `gorm:"foreignKey:TakeoutOrderUuid;references:Uuid"`
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

func (o *TakeoutOrder) SetTakeoutOrderPromos(promos []*TakeoutOrderPromo) {
	if promos == nil {
		o.TakeoutOrderPromos = nil
		return
	}
	if len(promos) == 0 {
		o.TakeoutOrderPromos = nil
		return
	}
	o.TakeoutOrderPromos = make([]TakeoutOrderPromo, len(promos))
	for i, promo := range promos {
		o.TakeoutOrderPromos[i] = *promo
	}
}

// 是否异常订单
func (o *TakeoutOrder) IsAbnormalOrder() bool {
	return o.IsAbnormal == 1
}

// 是否待接单订单
func (o *TakeoutOrder) IsPendingOrder() error {
	if o == nil {
		return errors.New("订单不存在")
	}
	// 检查订单状态
	if o.OrderState != valueobject.TakeoutOrderStatePending {
		return errors.New("订单状态不正确")
	}
	return nil
}

// 是否是打包订单
func (o *TakeoutOrder) IsTakeawayOrder() bool {
	if o == nil {
		return false
	}
	return o.OrderType != valueobject.TakeoutOrderTypeDineIn
}

// 是否商家配送
func (o *TakeoutOrder) IsRestaurantDeliveryOrder() bool {
	if o == nil {
		return false
	}
	return o.OrderType == valueobject.TakeoutOrderTypeRestaurantDelivery
}

// 判断是否删除或者已经取消
func (o *TakeoutOrder) IsDeletedOrCanceled() bool {
	if o == nil {
		return true
	}
	return o.DeleteTime > 0 || o.OrderState == valueobject.TakeoutOrderStateRejected || o.OrderState == valueobject.TakeoutOrderStateCanceled
}

// 是否店内就餐订单
func (o *TakeoutOrder) IsDineInOrder() bool {
	if o == nil {
		return false
	}
	return o.OrderType == valueobject.TakeoutOrderTypeDineIn
}

// 是否自动接单
func (o *TakeoutOrder) IsAutoAcceptOrder() bool {
	if o == nil {
		return false
	}
	return o.OrderAcceptedType == valueobject.TakeoutOrderAcceptedTypeAuto
}

// 是否grab订单
func (o *TakeoutOrder) IsGrabOrder() bool {
	if o == nil {
		return false
	}
	return o.Platform == valueobject.TakeoutPlatformGrab
}

// 是否lineman订单
func (o *TakeoutOrder) IsLinemanOrder() bool {
	if o == nil {
		return false
	}
	return o.Platform == valueobject.TakeoutPlatformLineman
}

// 是否存在班次
func (o *TakeoutOrder) IsExistShiftLog() bool {
	if o == nil {
		return false
	}
	return o.StaffShiftLogUuid > 0
}

// 是否 ERP发票已同步
func (o *TakeoutOrder) IsErpInvoiceSynced() bool {
	if o == nil {
		return false
	}
	return len(o.ErpPosInvoiceResp) > 0
}

// 获取ERP POS Invoice响应数据
func (o *TakeoutOrder) GetErpPosInvoiceResp() *selling.SavePosInvoiceResp {
	if len(o.ErpPosInvoiceResp) == 0 {
		return nil
	}
	var erpPosInvoiceResp selling.SavePosInvoiceResp
	if err := json.Unmarshal([]byte(o.ErpPosInvoiceResp), &erpPosInvoiceResp); err != nil {
		return nil
	}
	return &erpPosInvoiceResp
}

// 获取外卖订单编号
func (o *TakeoutOrder) GetKdsTakeoutPlatformAndOrderNumber() string {
	if o == nil {
		return ""
	}
	return o.ShortOrderNumber
}

// 获取外卖平台 全小写
func (o *TakeoutOrder) GetToLowerPlatform() string {
	if o == nil {
		return ""
	}
	return strings.ToLower(o.Platform)
}

// 获取外卖平台 首字母大写
func (o *TakeoutOrder) GetCapitalPlatform() string {
	if o == nil || len(o.Platform) == 0 {
		return ""
	}
	return strings.ToUpper(o.Platform[0:1]) + strings.ToLower(o.Platform[1:])
}

// 获取外卖平台 无空格名称
func (o *TakeoutOrder) GetNoSpacePlatformName() string {
	if o == nil {
		return ""
	}
	return valueobject.GetNoSpacePlatformName(o.Platform)
}

// 获取外卖平台 有空格名称
func (o *TakeoutOrder) GetSpacePlatformName() string {
	if o == nil {
		return ""
	}
	return valueobject.GetPlatformName(o.Platform)
}

// 获取lineman服务费
func (o *TakeoutOrder) GetLinemanServiceFee() float64 {
	if o == nil {
		return 0
	}
	if o.IsLinemanOrder() {
		return o.Subtotal - o.PlatformTotal
	}
	return 0
}

// 获取骑手状态
func (o *TakeoutOrder) GetRiderStatus() string {
	if o.PlatformOrderState == "DRIVER_ARRIVED" {
		return "rider_accepted"
	}
	return "rider_pending"
}

// 获取附加信息
func (o *TakeoutOrder) GetAdditionalProperties(language string) string {
	if o == nil {
		return ""
	}
	if o.IsGrabOrder() {
		if o.Cutlery == 1 {
			return i18n.Translate(language, "需要餐具")
		} else {
			return i18n.Translate(language, "不需要餐具")
		}
	}
	// 替换泰语餐具文案为 i18n 翻译
	result := o.AdditionalProperties
	result = strings.ReplaceAll(result, "ไม่รับช้อนส้อมพลาสติก", i18n.Translate(language, "不需要餐具"))
	result = strings.ReplaceAll(result, "รับช้อนส้อมพลาสติก", i18n.Translate(language, "需要餐具"))
	return result
}
