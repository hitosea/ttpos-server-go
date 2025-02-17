package model

import (
	"ttpos-server-go/i18n"
)

// PaymentMethod 支付方式 `ttpos_payment_method`
type PaymentMethod struct {
	BaseModel
	Name                 string  `gorm:"column:name;type:varchar(255);comment:支付方式名称;NOT NULL" json:"name"`
	Code                 int     `gorm:"column:code;type:int(11);default:0;comment:支付方式代号;NOT NULL" json:"code"`
	PaymentName          string  `gorm:"column:payment_name;type:varchar(255);comment:支付名称;NOT NULL" json:"payment_name"`
	Source               int     `gorm:"column:source;type:tinyint(1);default:1;comment:来源 0-系统 1-手动 2-LianLianPay;NOT NULL" json:"source"`
	LogoFileUuid         uint64  `gorm:"column:logo_file_uuid;type:bigint(20) unsigned;default:0;comment:logo图片ID;NOT NULL" json:"logo_file_uuid"`
	QrcodeFileUuid       uint64  `gorm:"column:qrcode_file_uuid;type:bigint(20) unsigned;default:0;comment:二维码图片ID;NOT NULL" json:"qrcode_file_uuid"`
	FeePercent           float64 `gorm:"column:fee_percent;type:decimal(5,2);default:0;comment:手续费百分比;NOT NULL" json:"fee_percent"`
	IsShowCashier        int     `gorm:"column:is_show_cashier;type:tinyint(1);default:0;comment:0-不显示 1-收银机结账显示;NOT NULL" json:"is_show_cashier"`
	IsShowAssistant      int     `gorm:"column:is_show_assistant;type:tinyint(1);default:0;comment:0-不显示 1-点餐助手结账显示;NOT NULL" json:"is_show_assistant"`
	IsShowMemberRecharge int     `gorm:"column:is_show_member_recharge;type:tinyint(1);default:0;comment:0-不显示 1-收银机会员充值显示;NOT NULL" json:"is_show_member_recharge"`
	Status               int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态 0-禁用 1-启用;NOT NULL" json:"status"`
	Sort                 int     `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
}

// GetSourceText 获取来源文本
func (model *PaymentMethod) GetSourceText(language string) string {
	if model.Source == 0 {
		return i18n.Translate(language, "系统")
	} else if model.Source == 1 {
		return i18n.Translate(language, "手动")
	} else if model.Source == 2 {
		return i18n.Translate(language, "LianLianPay")
	}
	return ""
}

/**
PaymentAmount:
要充值的金额，现金支付可能大于这个值；比如要充值28块，但是会员给了30，则payment_amount=28；charge_due=2，且此时不可修改充值订单为29

如果要充值18元，使用非现金支付方式，且支付方式手续费为1%，充值10元，则payment_amount=10; payment_commission_fee=0.1，则amount=10.1（此时可以修改充值金额为10）
剩余8元使用现金支付，会员给了10元现金，需要找零2元，则payment_amount=8; charge_du=2（此时不可修改充值订单金额为19元）

得出结论，修改充值金额时，充值金额不能小于所有有支付订单的 payment_amount + charge_due
*/

// PaymentOrder 支付订单 ttpos_payment_order
type PaymentOrder struct {
	BaseModel
	PaymentMethodName    string  `gorm:"column:payment_method_name;type:varchar(255);comment:支付类型名称;NOT NULL" json:"payment_method_name"`
	PaymentMethodUuid    uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:支付类型ID;NOT NULL" json:"payment_method_uuid"`
	PaymentFeePercent    float64 `gorm:"column:payment_fee_percent;type:decimal(5,2);default:0;comment:支付手续费百分比;NOT NULL" json:"payment_fee_percent"`
	RelatedUuid          uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联的充值订单、销售订单ID;NOT NULL" json:"related_uuid"`
	CurrencyUnit         string  `gorm:"column:currency_unit;type:varchar(10);comment:货币单位;NOT NULL" json:"currency_unit"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额;NOT NULL" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,支付金额*支付手续费百分比;NOT NULL" json:"payment_commission_fee"`
	ChargeDue            float64 `gorm:"column:charge_due;type:decimal(12,2);default:0;comment:找零;NOT NULL" json:"charge_due"`
	Amount               float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:金额：支付金额+支付手续费;NOT NULL" json:"amount"`
	TransactionNumber    string  `gorm:"column:transaction_number;type:varchar(255);comment:交易号;NOT NULL" json:"transaction_number"`
	Status               int     `gorm:"column:status;type:tinyint(1);default:0;comment:支付状态, 0-未支付 1-已支付 2-已退款;NOT NULL" json:"status"`

	// 关联字段
	PaymentMethod       *PaymentMethod       `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
	MemberRechargeOrder *MemberRechargeOrder `gorm:"foreignKey:RelatedUuid;references:Uuid"`
}
