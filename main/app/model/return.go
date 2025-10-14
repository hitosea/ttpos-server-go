package model

import "github.com/shopspring/decimal"

// ReturnOrder 退货单表 `ttpos_return_order`
type ReturnOrder struct {
	BaseModel
	RelatedOrderType    uint    `gorm:"column:related_order_type;type:tinyint(1) unsigned;default:0;comment:关联订单类型：0-销售订单；1-充值订单 2-会员订单;NOT NULL" json:"related_order_type"`
	RelatedOrderUuid    uint64  `gorm:"column:related_order_uuid;type:bigint(20) unsigned;default:0;comment:关联订单ID;NOT NULL" json:"related_order_uuid"`
	RelatedOrderNo      string  `gorm:"column:related_order_no;type:varchar(255);comment:关联订单号;NOT NULL" json:"related_order_no"`
	IsReverseSettlement uint    `gorm:"column:is_reverse_settlement;type:tinyint(1) unsigned;default:0;comment:是否反结账：0-否；1-是;NOT NULL" json:"is_reverse_settlement"`
	ReturnType          uint    `gorm:"column:return_type;type:tinyint(1) unsigned;default:0;comment:退货类型,1-整单退货,2-部分退货;NOT NULL" json:"return_type"`
	RefundAmount        float64 `gorm:"column:refund_amount;type:decimal(12,2);default:0.00;comment:退款金额,包括税额;NOT NULL" json:"refund_amount"`
	Unit                string  `gorm:"column:unit;type:varchar(255);comment:货币单位;NOT NULL" json:"unit"`
	RefundTaxAmount     float64 `gorm:"column:refund_tax_amount;type:decimal(12,2);default:0.00;comment:退款税额;NOT NULL" json:"refund_tax_amount"`
	RefundReason        string  `gorm:"column:refund_reason;type:varchar(255);comment:退款原因;NOT NULL" json:"refund_reason"`
	BankCode            string  `gorm:"column:bank_code;type:varchar(255);comment:银行编码 - 当存在QR PromptPay的时候需要传;NOT NULL" json:"bank_code"`
	AccountNo           string  `gorm:"column:account_no;type:varchar(255);comment:账号 - 当存在QR PromptPay的时候需要传;NOT NULL" json:"account_no"`
	AccountName         string  `gorm:"column:account_name;type:varchar(255);comment:账户名称 - 当存在QR PromptPay的时候需要传;NOT NULL" json:"account_name"`
	DutyNo              string  `gorm:"column:duty_no;type:varchar(255);comment:当班编号;NOT NULL" json:"duty_no"`
	StaffShiftLogUuid   uint64  `gorm:"column:staff_shift_log_uuid;type:bigint(20) unsigned;default:0;comment:员工班次记录ID;NOT NULL" json:"staff_shift_log_uuid"`

	// erp相关
	ErpInvoiceName string `gorm:"column:erp_invoice_name;type:varchar(255);comment:发票名称;NOT NULL" json:"erp_invoice_name"`

	ReturnOrderAmounts  []ReturnOrderAmount   `gorm:"foreignKey:ReturnOrderUuid;references:uuid"`
	ReturnOrderProducts []*ReturnOrderProduct `gorm:"foreignKey:ReturnOrderUuid;references:uuid"`
}

// SetNil 设置关联对象为nil，避免gorm创建时将关联对象也创建
func (r *ReturnOrder) SetNil() {
	r.ReturnOrderAmounts = nil
	r.ReturnOrderProducts = nil
}

// IsExistLianLianPay 是否存在连连支付
func (model *ReturnOrder) IsExistLianLianPay() bool {
	for _, amount := range model.ReturnOrderAmounts {
		if amount.PaymentMethod != nil {
			if amount.PaymentMethod.IsLianLianPay() {
				return true
			}
		}
	}
	return false
}

// GetLianLianPayCount 获取连连支付数量
func (model *ReturnOrder) GetLianLianPayCount() int {
	count := 0
	for _, amount := range model.ReturnOrderAmounts {
		if amount.PaymentMethod != nil {
			if amount.PaymentMethod.IsLianLianPay() {
				count++
			}
		}
	}
	return count
}

// IsExistQrPromptPay 是否存在QrPromptPay支付
func (model *ReturnOrder) IsExistQrPromptPay() bool {
	for _, amount := range model.ReturnOrderAmounts {
		if amount.PaymentMethod != nil {
			if amount.PaymentMethod.IsQrPromptPay() {
				return true
			}
		}
	}
	return false
}

// ReturnOrderAmount 退款金额表 ttpos_return_order_amount
type ReturnOrderAmount struct {
	BaseModel
	Amount       float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:退款金额;NOT NULL" json:"amount"`
	RefundStatus int     `gorm:"column:refund_status;type:int(11);default:0;comment:退款状态 0-退款中 1-退款成功 2-退款失败;NOT NULL" json:"refund_status"`
	// 关联字段
	ReturnOrderUuid   uint64 `gorm:"column:return_order_uuid;type:bigint(20) unsigned;default:0;comment:关联退货单ID;NOT NULL" json:"return_order_uuid"`
	PaymentMethodUuid uint64 `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:关联支付方式ID;NOT NULL" json:"payment_method_uuid"`
	PaymentOrderUuid  uint64 `gorm:"column:payment_order_uuid;type:bigint(20) unsigned;default:0;comment:关联支付单ID,用于判断支付单的钱还有多少未退;NOT NULL" json:"payment_order_uuid"`

	// 连连退款
	MerchantRefundOrderNo string `gorm:"column:merchant_refund_order_no;type:varchar(255);comment:商户退款单号;NOT NULL" json:"merchant_refund_order_no"`
	LlReturnOrderid       string `gorm:"column:ll_return_order_id;type:varchar(255);comment:连连退款订单ID, 用来重新发起退款;NOT NULL" json:"ll_return_order_id"`

	ReturnOrder      *ReturnOrder      `gorm:"foreignKey:ReturnOrderUuid;references:Uuid"`       // 关联退货单
	PaymentMethod    *PaymentMethod    `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`     // 关联支付方式
	MemberPointLog   *MemberPointLog   `gorm:"foreignKey:RelatedUuid;references:uuid"`           // 关联积分变动记录.扣减积分
	MemberBalanceLog *MemberBalanceLog `gorm:"foreignKey:RelatedUuid;references:uuid"`           // 关联余额变动记录.退回余额
	CashBoxLog       *CashBoxLog       `gorm:"foreignKey:RefundOrderAmountUuid;references:uuid"` // 关联钱箱变动记录
}

// SetNil 设置关联对象为nil，避免gorm创建时将关联对象也创建
func (r *ReturnOrderAmount) SetNil() {
	r.ReturnOrder = nil
	r.PaymentMethod = nil
	r.MemberPointLog = nil
	r.MemberBalanceLog = nil
}

// ReturnOrderProduct 退货单商品表 ttpos_return_order_product
type ReturnOrderProduct struct {
	BaseModel
	// 基本信息
	SaleOrderUuid        uint64  `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	SaleOrderProductUuid uint64  `gorm:"column:sale_order_product_uuid;comment:销售订单商品表ID" json:"sale_order_product_uuid"` // 销售订单商品表ID、自助餐顾客类型表ID、自助餐加钟表ID
	ReturnOrderUuid      uint64  `gorm:"column:return_order_uuid;comment:退货单ID" json:"return_order_uuid"`
	ProductType          uint    `gorm:"column:product_type;comment:商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct" json:"product_type"`
	ProductPackageUuid   uint64  `gorm:"column:product_package_uuid;comment:商品包ID" json:"product_package_uuid"` // 商品包ID、自助餐顾客类型ID、自助餐加钟ID
	ProductName          string  `gorm:"column:product_name;comment:商品名称" json:"product_name"`
	ProductPrice         float64 `gorm:"column:product_price;comment:商品单价" json:"product_price"`
	TaxRate              float64 `gorm:"column:tax_rate;comment:税率,根据结账时税率计算" json:"tax_rate"`
	Num                  float64 `gorm:"column:num;comment:商品数量" json:"num"`
	ProductDiscount      float64 `gorm:"column:product_discount;comment:商品折扣" json:"product_discount"`
	ProductTotalAmount   float64 `gorm:"column:product_total_amount;comment:商品总金额" json:"product_total_amount"`
	ErpCode              string  `gorm:"column:erp_code;type:varchar(255);default:'';comment:ERP系统商品编码;NOT NULL" json:"erp_code"`
}

// 退款商品的未含税价格. 当商品已含税时，需要减去税费。当商品未含税时，直接返回商品单价， taxFee 为 0。
func (r *ReturnOrderProduct) GetProductPriceNoneTax(taxFee float64, hasTax bool) float64 {
	if !hasTax {
		taxFee = 0 // 商品未含税时，不用从商品单价中减去税费
	}
	price := decimal.NewFromFloat(r.ProductPrice) // 商品最终售价（显示在订单中的价格）
	tax := decimal.NewFromFloat(taxFee)
	price = price.Sub(tax) // 商品已含税时，需要减去税费. 商品未含税时，直接返回商品单价， taxFee 为 0。
	return price.Round(2).InexactFloat64()
}

// 不含税费的商品总金额
func (r *ReturnOrderProduct) GetProductTotalAmountNoneTax(taxFee float64, hasTax bool) float64 {
	return decimal.NewFromFloat(r.GetProductPriceNoneTax(taxFee, hasTax)).Mul(decimal.NewFromFloat(r.Num)).Round(2).InexactFloat64()
}

// SetNil 设置关联对象为nil，避免gorm创建时将关联对象也创建
func (r *ReturnOrderProduct) SetNil() {
}
