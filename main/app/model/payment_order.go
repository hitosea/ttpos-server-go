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

	QrcodeFile *File `gorm:"foreignKey:QrcodeFileUuid;references:Uuid"` // 关联文件
	LogoFile   *File `gorm:"foreignKey:LogoFileUuid;references:Uuid"`   // 关联文件
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

// PaymentOrder 支付订单 ttpos_payment_order
type PaymentOrder struct {
	BaseModel
	PaymentMethodName    string  `gorm:"column:payment_method_name;type:varchar(255);comment:支付类型名称;NOT NULL" json:"payment_method_name"`
	PaymentMethodUuid    uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:支付类型ID;NOT NULL" json:"payment_method_uuid"`
	PaymentFeePercent    float64 `gorm:"column:payment_fee_percent;type:decimal(5,2);default:0;comment:支付手续费百分比;NOT NULL" json:"payment_fee_percent"`
	RelatedType          int     `gorm:"column:related_type;type:tinyint(1);default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_type"`
	RelatedUuid          uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联的充值订单、销售订单ID;NOT NULL" json:"related_uuid"`
	CurrencyUnit         string  `gorm:"column:currency_unit;type:varchar(10);comment:货币单位;NOT NULL" json:"currency_unit"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额;NOT NULL" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,支付金额*支付手续费百分比;NOT NULL" json:"payment_commission_fee"`
	Amount               float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:实收金额;NOT NULL" json:"amount"`
	TransactionNumber    string  `gorm:"column:transaction_number;type:varchar(255);comment:交易号;NOT NULL" json:"transaction_number"`
	Status               int     `gorm:"column:status;type:tinyint(1);default:0;comment:支付状态, 0-未支付 1-已支付 2-已退款;NOT NULL" json:"status"`

	// 关联字段
	PaymentMethod       *PaymentMethod       `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
	MemberRechargeOrder *MemberRechargeOrder `gorm:"foreignKey:RelatedUuid;references:Uuid"`
}
