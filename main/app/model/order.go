package model

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
)

// SaleBill 销售账单 `ttpos_sale_bill`
type SaleBill struct {
	BaseModel
	// 主键和标识字段
	OrderNo  string `gorm:"column:order_no;type:varchar(255);default:'';comment:销售账单编号" json:"order_no"`
	DutyNo   string `gorm:"column:duty_no;type:varchar(255);default:'';comment:当班编号,用于标记该账单属于哪个当班" json:"duty_no"`
	SerialNo string `gorm:"column:serial_no;type:varchar(255);default:'';comment:桌位编号 (点餐流水号)" json:"serial_no"`

	// 状态相关字段
	Status uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-待付款、1-已完成、2-已取消" json:"status"`
	IsLock uint `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`

	// 订单类型字段
	BillType       uint `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-桌台订单、1-点餐订单" json:"bill_type"`
	DiningMethod   uint `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	IsBuffet       uint `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	BuffetDuration uint `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒）" json:"buffet_duration"`

	// 订单基本信息
	MealNum uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Remark  string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	Reason  string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 关联ID字段
	ConsumerUuid    uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid     uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	BuffetOrderUuid uint64 `gorm:"column:buffet_order_uuid;type:bigint(20);default:0;comment:自助餐订单ID" json:"buffet_order_uuid"`
	DeskUuid        uint64 `gorm:"column:desk_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"desk_uuid"`

	// 金额字段 - 主要金额
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
	ProductAmount float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额,关联销售订单的商品金额之和" json:"product_amount"`

	// 金额字段 - 支付相关
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,多次支付的支付手续费之和" json:"payment_commission_fee"`

	// 金额字段 - 费用相关
	ServiceFee float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费,关联销售订单的服务费之和" json:"service_fee"`
	TaxFee     float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费,关联销售订单的税费之和" json:"tax_fee"`

	// 金额字段 - 优惠相关
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用,关联销售订单的折扣费用之和" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用,关联销售订单的会员折扣费用之和" json:"member_discount_fee"`
	GiftAmount        float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,关联销售订单的赠菜金额之和" json:"gift_amount"`
	FreeAmount        float64 `gorm:"column:free_amount;type:decimal(12,2);default:0;comment:免单金额,关联销售订单的免单金额之和" json:"free_amount"`

	// 时间相关字段
	FinishTime   int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	HideBillTime int64 `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`

	// 关联字段
	SaleOrders      []SaleOrder     `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	SaleBillSetting SaleBillSetting `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	Cashier         Staff           `gorm:"foreignKey:CashierUuid;references:uuid"`
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleBill) ValidateOrderStatus(operation string, saleOrderUuid ...uint64) error {
	if operation != constant.OrderSettle && model.IsLock == 1 {
		return errors.New("订单已被锁定，请解锁后重新操作")
	}
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	if len(model.SaleOrders) > 0 {
		// 拆单没有取消权限
		if operation == constant.OrderOrderCancel && len(model.SaleOrders) > 1 {
			return errors.New("拆单不可操作")
		}
		// 单个订单不能操作
		for _, so := range model.SaleOrders {
			if len(saleOrderUuid) == 0 || slices.Contains(saleOrderUuid, so.Uuid) {
				if err := so.ValidateOrderStatus(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 获取总的退款金额
func (model *SaleBill) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, saleOrder := range model.SaleOrders {
		for _, refundOrder := range saleOrder.ReturnOrders {
			refundAmount += refundOrder.RefundAmount
		}
	}
	return refundAmount
}

// 获取所有自助餐名称
func (model *SaleBill) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, order := range model.SaleOrders {
		for _, buffet := range order.Buffets {
			name := buffet.BuffetPackageMultiLanguageName.GetNameByLang(language)
			if !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
			}
		}
	}
	return strings.Join(buffets, "+")
}

// SaleOrder 销售订单 ttpos_sale_order
type SaleOrder struct {
	BaseModel
	// 基础标识字段
	OrderNo string `gorm:"column:order_no;type:varchar(255);default:'';comment:订单编号" json:"order_no"`
	Status  uint   `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	IsFree  uint   `gorm:"column:is_free;type:tinyint(1);default:0;comment:是否免单, 0-否 1-是" json:"is_gift"`

	// 关联ID字段
	ConsumerUuid uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid  uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 商品金额相关字段
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:商品原始金额" json:"product_original_amount"`

	// 费用相关字段
	ServiceFee        float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费" json:"service_fee"`
	TaxFee            float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费" json:"tax_fee"`
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用" json:"member_discount_fee"`

	// 订单总额相关字段
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
	PaymentAmount float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`

	// 时间相关字段
	FinishTime int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`

	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);default:100;comment:会员折扣率(0-100%)" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);default:100;comment:会员卡折扣率(0-100%)" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);default:100;comment:自定义折扣率(0-100%)" json:"custom_discount_rate"`

	// 关联对象
	PaymentOrders     []PaymentOrder                `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	Member            Member                        `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts []SaleOrderProduct            `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	ReturnOrders      []ReturnOrder                 `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	Buffets           []SaleOrderBuffetCustomerType `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
}

// TableName 指定表名
func (SaleOrder) TableName() string {
	return "ttpos_sale_order"
}

// 获取总的退款金额
func (model *SaleOrder) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, refundOrder := range model.ReturnOrders {
		refundAmount += refundOrder.RefundAmount
	}
	return refundAmount
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleOrder) ValidateOrderStatus() error {
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	return nil
}

// 获取所有自助餐名称
func (model *SaleOrder) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, buffet := range model.Buffets {
		buffets = append(buffets, buffet.BuffetPackageMultiLanguageName.GetNameByLang(language))
	}
	return strings.Join(buffets, "+")
}

// SaleOrderProduct 销售订单产品 `ttpos_sale_order_product`
type SaleOrderProduct struct {
	// 基础字段
	BaseModel

	// 基本信息字段
	Name       string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品名称'" json:"name"`
	FlavorName string `gorm:"column:flavor_name;type:varchar(255);not null;default:'';comment:'规格名称'" json:"flavor_name"`
	Num        uint   `gorm:"column:num;type:int(11);not null;default:0;comment:'商品数量。不能减为0，当数量为1再减时，标记删除'" json:"num"`
	Remark     string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'备注，顾客对商品的备注信息'" json:"remark"`

	// 状态相关字段
	Status        uint `gorm:"column:status;type:tinyint(1);not null;default:0;comment:'状态, 0-未送厨 1-已送厨 2-已退'" json:"status"`
	IsRequire     uint `gorm:"column:is_require;type:tinyint(1);not null;default:0;comment:'是否必点商品 0-否 1-是。用于在前端显示必点图标'" json:"is_require"`
	IsAcceptOrder uint `gorm:"column:is_accept_order;type:tinyint(1);not null;default:0;comment:'是否已接单, 0-否 1-是'" json:"is_accept_order"`

	// 价格相关字段
	FlavorPrice  float64 `gorm:"column:flavor_price;type:decimal(12,2);not null;default:0.00;comment:'规格原价（单商品）,仅某规格商品的原价'" json:"flavor_price"`
	SaucePrice   float64 `gorm:"column:sauce_price;type:decimal(12,2);not null;default:0.00;comment:'小料价（单商品）,所有小料的价格之和'" json:"sauce_price"`
	ProductPrice float64 `gorm:"column:product_price;type:decimal(12,2);not null;default:0.00;comment:'原始单价（单商品）,规格原价+小料价'" json:"product_price"`
	SalePrice    float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0.00;comment:'销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价'" json:"sale_price"`
	Price        float64 `gorm:"column:price;type:decimal(12,2);not null;default:0.00;comment:'最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率'" json:"price"`
	TotalPrice   float64 `gorm:"column:total_price;type:decimal(12,2);not null;default:0.00;comment:'应收金额(单商品)=最终单价+服务费+总税费'" json:"total_price"`

	// 折扣相关字段
	IsCustomPrice          uint    `gorm:"column:is_custom_price;type:tinyint(1);not null;default:0;comment:'是否自定义价格（单商品）, 0-否 1-是'" json:"is_custom_price"`
	IsOpenMemberDiscount   uint    `gorm:"column:is_open_member_discount;type:tinyint(1);not null;default:0;comment:'是否开启会员折扣, 0-否 1-是'" json:"is_open_member_discount"`
	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员折扣率(0-100%)'" json:"member_discount_rate"`
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员卡折扣率(0-100%)'" json:"member_card_discount_rate"`
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣率(0-100%)'" json:"custom_discount_rate"`

	// 折扣金额字段
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);not null;default:0.00;comment:'打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额'" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);not null;default:0.00;comment:'会员折扣金额（单商品）=销售价*会员折扣率*会员卡折扣率'" json:"member_discount_fee"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣金额（单商品）=销售价-最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)'" json:"custom_discount_fee"`

	// 税费和服务费字段
	TaxRate       float64 `gorm:"column:tax_rate;type:decimal(12,2);not null;default:0;comment:'税率,单位%.加购时记录税率,结账时再重新核算'" json:"tax_rate"`
	ServiceTaxFee float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率'" json:"service_tax_fee"`
	TaxFee        float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0.00;comment:'商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率'" json:"tax_fee"`
	ServiceFee    float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例'" json:"service_fee"`

	// 库存相关字段
	DeductStockType uint  `gorm:"column:deduct_stock_type;type:tinyint(1);not null;default:0;comment:'库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数'" json:"deduct_stock_type"`
	DeductStockTime int64 `gorm:"column:deduct_stock_time;type:int(10);not null;default:0;comment:'减库存的时间(时间戳)，0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存'" json:"deduct_stock_time"`

	// 赠品相关字段
	IsGift     uint   `gorm:"column:is_gift;type:tinyint(1);not null;default:0;comment:'是否赠菜, 0-否 1-是'" json:"is_gift"`
	GiftReason string `gorm:"column:gift_reason;type:varchar(255);not null;default:'';comment:'赠菜原因'" json:"gift_reason"`

	// 关联ID字段
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称UUID'" json:"multi_language_name_uuid"`
	ImageFileUuid         uint64 `gorm:"column:image_file_uuid;not null;default:0;comment:'商品图片ID'" json:"image_file_uuid"`
	ProductionOrderUuid   uint64 `gorm:"column:production_order_uuid;type:bigint(20);not null;default:0;comment:'生产订单ID'" json:"production_order_uuid"`
	ProductPackageUuid    uint64 `gorm:"column:product_package_uuid;type:bigint(20);not null;default:0;comment:'商品包ID'" json:"product_package_uuid"`
	SaleBillUuid          uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);not null;default:0;comment:'销售账单ID'" json:"sale_bill_uuid"`
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:'销售订单ID'" json:"sale_order_uuid"`
	QrcodeOrderUuid       uint64 `gorm:"column:qrcode_order_uuid;type:bigint(20);not null;default:0;comment:'扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品'" json:"qrcode_order_uuid"`

	// 其他字段
	Sign                 string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:'商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品'" json:"sign"`
	IsQrcodeOrderProduct uint   `gorm:"column:is_qrcode_order_product;type:tinyint(1);not null;default:0;comment:'是否为扫码订单商品, 0-否 1-是'" json:"is_qrcode_order_product"`

	// 关联对象
	MultiLanguageName          MultiLanguageName           `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	ImageFile                  File                        `gorm:"foreignKey:image_file_uuid;references:uuid"`
	SaleOrderProductBoms       []SaleOrderProductBom       `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
	SaleOrderProductAttributes []SaleOrderProductAttribute `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
}

// SaleOrderProductAttribute 销售订单产品属性 `ttpos_sale_order_product_attribute`
type SaleOrderProductAttribute struct {
	BaseModel
	Name                 string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品属性名称,不随后台更新'"`
	SaleOrderUuid        uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductAttributeUuid uint64 `gorm:"column:product_attribute_uuid;not null;default:0;comment:'商品属性ID'"`
}

// GetAttributeNames 获取属性名称字符串
func (model *SaleOrderProduct) GetAttributeNames() string {
	attributeNames := []string{}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom == 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	for _, attribute := range model.SaleOrderProductAttributes {
		attributeNames = append(attributeNames, attribute.Name)
	}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom != 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	return strings.Join(attributeNames, "; ")
}

// SaleOrderProductBom 销售订单产品原料 `ttpos_sale_order_product_bom`
type SaleOrderProductBom struct {
	BaseModel
	Name                 string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:'规格或小料规格名称,不随后台更新'"`
	Price                float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:'单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动'"`
	IsFlavorBom          uint    `gorm:"column:is_flavor_bom;type:tinyint(1);not null;default:0;comment:'是否为规格商品BOM, 0-否,加料商品 1-是,规格商品'"`
	SaleOrderUuid        uint64  `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64  `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductBomUuid       uint64  `gorm:"column:product_bom_uuid;not null;default:0;comment:'商品BOM ID'"`

	ProductBom       ProductBom       `gorm:"foreignKey:product_bom_uuid;references:uuid"`
	SaleOrderProduct SaleOrderProduct `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
}

// SaleBillSetting 销售账单设置 ttpos_sale_bill_setting
type SaleBillSetting struct {
	// 基础字段
	BaseModel
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 费用计算设置
	ServiceFeeType  uint    `gorm:"column:service_fee_type;type:tinyint(1);default:0;comment:服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费" json:"service_fee_type"`
	ServiceFeeValue float64 `gorm:"column:service_fee_value;type:decimal(12,2);default:0;comment:服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例" json:"service_fee_value"`
	TaxFeeType      uint    `gorm:"column:tax_fee_type;type:tinyint(1);default:0;comment:税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税" json:"tax_fee_type"`

	// 优惠和抹零设置
	DiscountType uint `gorm:"column:discount_type;type:tinyint(1);default:0;comment:打折类型, 0-百分比打折% 1-百分比直接减免% off" json:"discount_type"`
	Zero         uint `gorm:"column:zero;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero"`
	ZeroCheckout uint `gorm:"column:zero_checkout;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout"`

	// 统计设置
	IsStatGift uint `gorm:"column:is_stat_gift;type:tinyint(1);default:0;comment:是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣" json:"is_stat_gift"`
	IsStatFree uint `gorm:"column:is_stat_free;type:tinyint(1);default:0;comment:是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费" json:"is_stat_free"`
}

// SaleOrderBuffetCustomerType 销售订单自助餐顾客类型
type SaleOrderBuffetCustomerType struct {
	// 主键字段
	BaseModel

	// 关联ID字段
	SaleOrderUuid                           uint64 `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	BuffetPackageUuid                       uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
	BuffetPackageMultiLanguageNameUuid      uint64 `gorm:"column:buffet_package_multi_language_name_uuid;comment:自助餐套餐多语言ID" json:"buffet_package_multi_language_name_uuid"`
	BuffetCustomerTypeUuid                  uint64 `gorm:"column:buffet_customer_type_uuid;comment:自助餐客户类型ID" json:"buffet_customer_type_uuid"`
	BuffetCustomerTypeMultiLanguageNameUuid uint64 `gorm:"column:buffet_customer_type_multi_language_name_uuid;comment:自助餐客户类型多语言ID" json:"buffet_customer_type_multi_language_name_uuid"`

	// 数值字段
	Num uint `gorm:"column:num;type:int(11);default:0;comment:人数" json:"num"`

	// 关联字段
	BuffetPackageMultiLanguageName      MultiLanguageName `gorm:"foreignKey:BuffetPackageMultiLanguageNameUuid;references:uuid"`
	BuffetCustomerTypeMultiLanguageName MultiLanguageName `gorm:"foreignKey:BuffetCustomerTypeMultiLanguageNameUuid;references:uuid"`
}

// SaleOrderBuffetDelayProduct 销售订单加钟价格商品表 `ttpos_sale_order_buffet_delay_product`
type SaleOrderBuffetDelayProduct struct {
	BaseModel
	SaleOrderUuid   uint64 `gorm:"default:0;column:sale_order_uuid;comment:'销售订单ID'"`
	BuffetDelayUuid uint64 `gorm:"default:0;column:buffet_delay_uuid;comment:'自助餐加钟价格ID'"`
	Num             uint   `gorm:"default:0;column:num;comment:'数量'"`
}

// SaleOrderProductMaterial 销售订单产品原料
type SaleOrderProductMaterial struct {
	BaseModel
	// 关联ID字段
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20);default:0;comment:销售订单产品ID" json:"sale_order_product_uuid"`
	BomUuid              uint64 `gorm:"column:bom_uuid;type:bigint(20);default:0;comment:BOM ID" json:"bom_uuid"`
}

// SaleBillOperationRecord 桌台账单操作记录
type SaleBillOperationRecord struct {
	BaseModel
	// 基本信息
	Data   string `gorm:"column:data;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"data"`
	Source string `gorm:"column:source;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"source"`
	Action string `gorm:"column:action;comment:操作行为" json:"action"`
	Remark string `gorm:"column:remark;comment:备注" json:"remark"`
	// 关联ID字段
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作员ID" json:"operator_uuid"`
}

// 销售订单优惠策略表
type SaleOrderDiscountStrategy struct {
	BaseModel
	// 基本信息
	Type      uint    `gorm:"column:type;type:tinyint(2);not null;default:0;comment:优惠策略类型,0-整单折扣、1-会员折扣" json:"type"`
	Name      string  `gorm:"column:name;type:varchar(50);not null;default:'';comment:优惠策略名称" json:"name"`
	Value     float64 `gorm:"column:value;type:decimal(12,2);not null;default:0.00;comment:优惠策略值" json:"value"`
	JsonField string  `gorm:"column:json_field;type:text;default:null;comment:JSON字段" json:"json_field"`
	// 关联ID字段
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
}

// GenerateProductSign 生成商品包签名. 商品包签名,规格、属性、加料、备注、`改价`相同的商品签名相同,用于取消拆单时合并商品。改价销售订单商品价格后要重新生成签名
func (model *SaleOrderProduct) GenerateProductSign() string {
	bomIdList := make([]string, 0)
	attributeIdList := make([]string, 0)

	// 物料ID列表
	for _, bom := range model.SaleOrderProductBoms {
		bomIdList = append(bomIdList, strconv.FormatUint(bom.ProductBomUuid, 10))
	}
	// 属性ID列表
	for _, attributeGroup := range model.SaleOrderProductAttributes {
		attributeIdList = append(attributeIdList, strconv.FormatUint(attributeGroup.ProductAttributeUuid, 10))
	}
	// 物料ID列表和属性ID列表排序
	sort.Slice(bomIdList, func(i, j int) bool {
		return bomIdList[i] < bomIdList[j]
	})
	sort.Slice(attributeIdList, func(i, j int) bool {
		return attributeIdList[i] < attributeIdList[j]
	})
	// 物料ID列表和属性ID列表拼接。格式：物料,物料,物料-属性,属性,属性-改价
	bomIdListStr := strings.Join(bomIdList, ",")
	attributeIdListStr := strings.Join(attributeIdList, ",")
	return fmt.Sprintf("%s-%s-%s-%.2f", bomIdListStr, attributeIdListStr, model.Remark, model.SalePrice)
}
