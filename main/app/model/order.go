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

	// 关联对象
	PaymentOrders     []PaymentOrder     `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
	Member            Member             `gorm:"foreignKey:ConsumerUuid;references:uuid"`
	SaleOrderProducts []SaleOrderProduct `gorm:"foreignKey:SaleOrderUuid;references:uuid"`
}

// TableName 指定表名
func (SaleOrder) TableName() string {
	return "ttpos_sale_order"
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

// SaleOrderProduct 销售订单产品 `ttpos_sale_order_product`
type SaleOrderProduct struct {
	// 主键和标识字段
	BaseModel

	// 产品基本信息
	Name       string `gorm:"column:name;type:varchar(255);default:'';comment:产品名称" json:"name"`
	FlavorName string `gorm:"column:flavor_name;type:varchar(255);default:'';comment:口味名称" json:"flavor_name"`
	Sign       string `gorm:"column:sign;type:varchar(255);default:'';comment:商品签名" json:"sign"`

	// 产品数量和价格
	Num         uint    `gorm:"column:num;type:int(11);default:0;comment:数量" json:"num"`
	CustomPrice float64 `gorm:"column:custom_price;type:decimal(12,2);default:0.00;comment:自定义价格" json:"custom_price"`
	UnitPrice   float64 `gorm:"column:unit_price;type:decimal(12,2);default:0.00;comment:单价" json:"unit_price"`
	Price       float64 `gorm:"column:price;type:decimal(12,2);default:0.00;comment:最终单价" json:"price"`
	ServiceFee  float64 `gorm:"column:service_fee;type:decimal(12,2);default:0.00;comment:服务费" json:"service_fee"`

	// 产品税率和原价
	TaxRate               uint    `gorm:"column:tax_rate;type:tinyint(1);default:0;comment:税率,单位%.下单时单税率,结账时再重新核算" json:"tax_rate"`
	TaxFee                float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0.00;comment:税费" json:"tax_fee"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0.00;comment:原价销售额.包含加料、税费." json:"product_original_amount"`

	// 产品状态和属性
	Status          uint   `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-正常 1-退菜" json:"status"`
	IsRequire       uint   `gorm:"column:is_require;type:tinyint(1);default:0;comment:是否必点商品 0-否 1-是" json:"is_require"`
	IsGift          uint   `gorm:"column:is_gift;type:tinyint(1);default:0;comment:是否赠品, 0-否 1-是" json:"is_gift"`
	DeductStockType uint   `gorm:"column:deduct_stock_type;type:tinyint(1);default:0;comment:库存计算方式,0-下单减库存 1-付款减库存。创建商品后不变" json:"deduct_stock_type"`
	Remark          string `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
	GiftReason      string `gorm:"column:gift_reason;type:varchar(255);default:'';comment:赠品原因" json:"gift_reason"`

	// 关联ID
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20);default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	ImageFileUuid         uint64 `gorm:"column:image_file_uuid;type:bigint(20);default:0;comment:图片ID" json:"image_file_uuid"`
	ProductionOrderUuid   uint64 `gorm:"column:production_order_uuid;type:bigint(20);default:0;comment:生产订单ID" json:"production_order_uuid"`
	ProductPackageUuid    uint64 `gorm:"column:product_package_uuid;type:bigint(20);default:0;comment:产品包ID" json:"product_package_uuid"`
	QrcodeOrderUuid       uint64 `gorm:"column:qrcode_order_uuid;type:bigint(20);default:0;comment:扫码订单ID" json:"qrcode_order_uuid"`
	SaleBillUuid          uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_order_uuid"`

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

// GenerateProductSign 生成商品包签名. 商品包签名,规格、属性、加料、`改价`相同的商品签名相同,用于取消拆单时合并商品。改价销售订单商品价格后要重新生成签名
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
	return fmt.Sprintf("%s-%s-%d", bomIdListStr, attributeIdListStr, model.CustomPrice)
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
	// 主键和标识字段
	BaseModel

	// 关联字段
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 服务费相关设置
	ServiceFeeType  uint    `gorm:"column:service_fee_type;type:tinyint(1);default:0;comment:服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费" json:"service_fee_type"`
	ServiceFeeValue float64 `gorm:"column:service_fee_value;type:decimal(12,2);default:0;comment:服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例" json:"service_fee_value"`

	// 税费设置
	TaxFeeType uint `gorm:"column:tax_fee_type;type:tinyint(1);default:0;comment:税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税" json:"tax_fee_type"`

	// 抹零设置
	Zero         uint `gorm:"column:zero;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero"`
	ZeroCheckout uint `gorm:"column:zero_checkout;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout"`
}

// SaleOrderBuffetCustomerType 销售订单自助餐顾客类型
type SaleOrderBuffetCustomerType struct {
	// 主键字段
	BaseModel

	// 关联ID字段
	SaleOrderUuid          uint64 `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`
	BuffetPackageUuid      uint64 `gorm:"column:buffet_package_uuid;type:bigint(20);default:0;comment:自助餐套餐ID" json:"buffet_package_uuid"`
	BuffetCustomerTypeUuid uint64 `gorm:"column:buffet_customer_type_uuid;type:bigint(20);default:0;comment:自助餐客户类型ID" json:"buffet_customer_type_uuid"`

	// 数值字段
	Num uint `gorm:"column:num;type:int(11);default:0;comment:人数" json:"num"`
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
	Source  string `gorm:"column:source;type:varchar(255);not null;default:'';comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"source"`
	Action  string `gorm:"column:action;type:varchar(150);not null;default:'';comment:操作行为" json:"action"`
	Message string `gorm:"column:message;type:varchar(255);not null;default:'';comment:消息内容" json:"message"`
	Remark  string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
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
