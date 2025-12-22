package model

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// PaymentMethod 支付方式 `ttpos_payment_method`
type PaymentMethod struct {
	BaseModel
	Name                 string  `gorm:"column:name;type:varchar(255);comment:支付方式名称;NOT NULL" json:"name"`
	Code                 int     `gorm:"column:code;type:int(11);default:0;comment:支付方式代号;NOT NULL" json:"code"`
	PaymentName          string  `gorm:"column:payment_name;type:varchar(255);comment:支付名称;NOT NULL" json:"payment_name"`
	Source               int     `gorm:"column:source;type:tinyint(1);default:0;comment:来源 0-系统 1-手动 2-LianLianPay;NOT NULL" json:"source"`
	LogoFileUuid         uint64  `gorm:"column:logo_file_uuid;type:bigint(20) unsigned;default:0;comment:logo图片ID;NOT NULL" json:"logo_file_uuid"`
	QrcodeFileUuid       uint64  `gorm:"column:qrcode_file_uuid;type:bigint(20) unsigned;default:0;comment:二维码图片ID;NOT NULL" json:"qrcode_file_uuid"`
	FeePercent           float64 `gorm:"column:fee_percent;type:decimal(5,4);default:0;comment:手续费百分比，取值范围0-1;NOT NULL" json:"fee_percent"`
	IsShowCashier        int     `gorm:"column:is_show_cashier;type:tinyint(1);default:0;comment:0-不显示 1-收银机结账显示;NOT NULL" json:"is_show_cashier"`
	IsShowAssistant      int     `gorm:"column:is_show_assistant;type:tinyint(1);default:0;comment:0-不显示 1-点餐助手结账显示;NOT NULL" json:"is_show_assistant"`
	IsShowKiosk          int     `gorm:"column:is_show_kiosk;type:tinyint(1);default:0;comment:0-不显示 1-自助点餐机结账显示;NOT NULL" json:"is_show_kiosk"`
	IsShowMemberRecharge int     `gorm:"column:is_show_member_recharge;type:tinyint(1);default:0;comment:0-不显示 1-收银机会员充值显示;NOT NULL" json:"is_show_member_recharge"`
	Status               int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态 0-禁用 1-启用;NOT NULL" json:"status"`
	Sort                 int     `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
	DefaultImg           string  `gorm:"column:default_img;type:varchar(255);comment:默认图片;NOT NULL" json:"default_img"`
	ErpnextPayment       string  `gorm:"column:erpnext_payment;type:varchar(255);comment:ERPNext支付方式;NOT NULL" json:"erpnext_payment"`
	HeadquarterUuid      uint64  `gorm:"column:headquarter_uuid;type:bigint(20) unsigned;default:0;comment:总部ID;NOT NULL" json:"headquarter_uuid"`

	QrcodeFile *File `gorm:"foreignKey:QrcodeFileUuid;references:Uuid"` // 关联文件
	LogoFile   *File `gorm:"foreignKey:LogoFileUuid;references:Uuid"`   // 关联文件
}

func (model *PaymentMethod) SetNil() {
	model.QrcodeFile = nil
	model.LogoFile = nil
}

// 系统默认支付方式code
const (
	PaymentMethodBalance = 10 // 余额
	PaymentMethodCash    = 40 // 现金
)

// 渠道提供方
const (
	PaymentSourceSystem      = 0 // 系统默认
	PaymentSourceDefault     = 1 // 自行添加
	PaymentSourceLianlianpay = 2 // LianLianPay
)

// LianLian渠道支付方式code（特殊code）
const (
	PaymentCodeLianlianWechat      = 90111 // LIANLIAN_WECHAT_PAY
	PaymentCodeLianlianAli         = 90222 // LIANLIAN_ALI_PAY
	PaymentCodeLianlianQrPromptPay = 90333 // LIANLIAN_QR_PROMPT_PAY
)

// GetPaymentName 获取支付名称
func (model *PaymentMethod) GetPaymentName() string {
	return utils.IfString(model.PaymentName != "", model.PaymentName, "-")
}

// GetName 获取支付方式名称
func (model *PaymentMethod) GetName() string {
	return utils.IfString(model.Name != "", model.Name, "-")
}

// GetFeePercent 获取手续费率
func (model *PaymentMethod) GetFeePercent() float64 {
	// 兼容按1-100取值范围。将1-100转位0-1
	if model.FeePercent > 1 {
		model.FeePercent = decimal.NewFromFloat(model.FeePercent).Div(decimal.NewFromUint64(100)).InexactFloat64()
	}
	return model.FeePercent
}

// CalculatePaymentCommissionFee 计算手续费
func (model *PaymentMethod) CalculatePaymentCommissionFee(paymentAmount float64) float64 {
	// 将 paymentAmount 和 fee/100 转换为 decimal
	decimalPrice := decimal.NewFromFloat(paymentAmount)
	feeRate := decimal.NewFromFloat(model.GetFeePercent())
	// 计算费用，先保留3位小数，然后四舍五入到2位
	feeMoney := decimalPrice.Mul(feeRate).Truncate(3).Round(2)
	// 转换回 float64
	return feeMoney.InexactFloat64()
}

// CalculatePaymentAmount 包含手续费
func (model *PaymentMethod) CalculatePaymentAmount(paymentAmount float64) float64 {
	commissionFee := model.CalculatePaymentCommissionFee(paymentAmount)
	return decimal.NewFromFloat(paymentAmount).Add(decimal.NewFromFloat(commissionFee)).InexactFloat64()
}

// HasCommission 判断是否含手续费
func (model *PaymentMethod) HasCommission() bool {
	return model.FeePercent > 0
}

// 是否总部支付方式
func (model *PaymentMethod) IsHeadquarterPayment() bool {
	return model.HeadquarterUuid > 0
}

// IsBalance 是否是余额支付
func (model *PaymentMethod) IsBalance() bool {
	return model.Code == constant.PaymentMethodCodeBalance
}

// IsCash 是否是现金支付
func (model *PaymentMethod) IsCash() bool {
	return model.Code == constant.PaymentMethodCodeCash
}

// IsLianLianPay 是否连连支付
func (model *PaymentMethod) IsLianLianPay() bool {
	return slices.Contains([]int{
		constant.PaymentMethodCodeLianLianWechatPay,
		constant.PaymentMethodCodeLianLianAliPay,
		constant.PaymentMethodCodeLianLianQRPromptPay,
	}, model.Code)
}

// IsWechatPay 是否微信支付
func (model *PaymentMethod) IsWechatPay() bool {
	return model.Code == constant.PaymentMethodCodeLianLianWechatPay
}

// IsAliPay 是否支付宝支付
func (model *PaymentMethod) IsAliPay() bool {
	return model.Code == constant.PaymentMethodCodeLianLianAliPay
}

// IsQrPromptPay 是否QrPromptPay 支付
func (model *PaymentMethod) IsQrPromptPay() bool {
	return model.Code == constant.PaymentMethodCodeLianLianQRPromptPay
}

// IsDisabledCancel 判断是否不允许取消支付
func (model *PaymentMethod) IsDisabledCancel() bool {
	return model.IsLianLianPay()
}

// IsDraft 判断是否草稿状态
func (model *PaymentMethod) IsDraft() bool {
	return model.IsHeadquarterPayment() && model.ErpnextPayment == ""
}

// GetSourceText 获取来源文本
func (model *PaymentMethod) GetSourceText(language string) string {
	if model.Source == 0 {
		return i18n.Translate(language, "系统默认")
	} else if model.Source == 1 {
		return i18n.Translate(language, "自行添加")
	} else if model.Source == 2 {
		return i18n.Translate(language, "LianLianPay")
	}
	return ""
}

// PaymentOrder 支付订单 ttpos_payment_order
type PaymentOrder struct {
	BaseModel
	PaymentMethodName    string  `gorm:"column:payment_method_name;type:varchar(255);comment:支付类型名称;NOT NULL" json:"payment_method_name"`
	PaymentMethodUuid    uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:支付类型ID;NOT NULL" json:"payment_method_uuid"`
	PaymentFeePercent    float64 `gorm:"column:payment_fee_percent;type:decimal(5,4);default:0;comment:支付手续费百分比;NOT NULL" json:"payment_fee_percent"`
	RelatedType          int     `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_type"`
	RelatedUuid          uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联的充值订单、销售订单ID;NOT NULL" json:"related_uuid"`
	CurrencyUnit         string  `gorm:"column:currency_unit;type:varchar(10);comment:货币单位;NOT NULL" json:"currency_unit"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额;NOT NULL" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,支付金额*支付手续费百分比;NOT NULL" json:"payment_commission_fee"`
	Amount               float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:实收金额,实收金额=支付金额+支付手续费;NOT NULL" json:"amount"`
	TransactionNumber    string  `gorm:"column:transaction_number;type:varchar(255);comment:交易号;NOT NULL" json:"transaction_number"`
	Status               int     `gorm:"column:status;type:tinyint(1);default:0;comment:支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常;NOT NULL" json:"status"`
	StatusReason         string  `gorm:"column:status_reason;type:text;default:'';comment:支付状态原因;NOT NULL" json:"status_reason"`

	// 余额支付相关
	BalanceAmount     float64 `gorm:"column:balance_amount;type:decimal(12,2);default:0;comment:主账户金额,用于反结账时退款;NOT NULL" json:"balance_amount"`
	GiftBalanceAmount float64 `gorm:"column:gift_balance_amount;type:decimal(12,2);default:0;comment:赠送帐户金额,用于反结账时退款;NOT NULL" json:"gift_balance_amount"`

	// 关联字段
	PaymentMethod       *PaymentMethod       `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
	MemberRechargeOrder *MemberRechargeOrder `gorm:"foreignKey:RelatedUuid;references:Uuid"`
	ReturnOrderAmounts  []*ReturnOrderAmount `gorm:"foreignKey:PaymentOrderUuid;references:Uuid"`
	RefundOrder         *RefundOrder         `gorm:"foreignKey:PaymentOrderUuid;references:Uuid"`
}

// NewRefundOrder 创建退款单
func (model *PaymentOrder) NewRefundOrder() *RefundOrder {
	return &RefundOrder{
		SaleOrderUuid:    model.RelatedUuid,
		PaymentOrderUuid: model.Uuid,
		RefundType:       1, // 反结账
		Amount:           model.Amount,
		Reason:           "反结账",
	}
}

// RefundOrder 退款单表 `ttpos_refund_order`
type RefundOrder struct {
	BaseModel
	SaleOrderUuid    uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单ID;NOT NULL" json:"sale_order_uuid"` // 也可能是充值订单ID
	SaleOrderNo      string  `gorm:"column:sale_order_no;type:varchar(255);default:'';comment:销售订单号;NOT NULL" json:"sale_order_no"`            // 废弃，暂时用不到
	PaymentOrderUuid uint64  `gorm:"column:payment_order_uuid;type:bigint(20) unsigned;default:0;comment:支付单ID;NOT NULL" json:"payment_order_uuid"`
	RefundType       uint    `gorm:"column:refund_type;type:int(11);default:0;comment:退款类型,1-反结账;NOT NULL" json:"refund_type"`
	Amount           float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:退款金额;NOT NULL" json:"amount"`
	Reason           string  `gorm:"column:reason;type:varchar(255);default:'';comment:退款原因;NOT NULL" json:"reason"`
	Status           uint    `gorm:"column:status;type:int(11);default:0;comment:退款状态;NOT NULL" json:"status"`

	// erp相关
	ErpInvoiceName string `gorm:"column:erp_invoice_name;type:varchar(255);comment:发票名称;NOT NULL" json:"erp_invoice_name"`
}

// GetCanReturnAmount 获取支付单的可退款金额. 可退款金额=支付金额-已退款金额
func (model *PaymentOrder) GetCanReturnAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, returnOrderAmount := range model.ReturnOrderAmounts {
		amount = amount.Add(decimal.NewFromFloat(returnOrderAmount.Amount))
	}
	// 可退款金额=支付金额-已退款金额
	return decimal.NewFromFloat(model.Amount).Sub(amount).Round(2).InexactFloat64()
}

func (model *PaymentOrder) GetSource() int {
	if model.PaymentMethod == nil {
		return 0
	}
	return model.PaymentMethod.Source
}

func (model *PaymentOrder) GetSourceText(language string) string {
	if model.PaymentMethod == nil {
		return ""
	}
	return model.PaymentMethod.GetSourceText(language)
}

// SetBaseModel 设置基础模型,当同一个支付方式已经存在付款单时，更新付款单
func (model *PaymentOrder) SetBaseModel(baseModel BaseModel) {
	defer model.SetUpdate()
	model.BaseModel = baseModel
}

func (model *PaymentOrder) SetNil() {
	model.PaymentMethod = nil
	model.MemberRechargeOrder = nil
}

func (model *PaymentOrder) Cancel() {
	model.Status = constant.PaymentOrderStatusRefund
	model.SetDelete()
}
