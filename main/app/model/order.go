package model

// 销售账单 SaleBill
type SaleBill struct {
	// 主键和标识字段
	ID      uint   `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid    uint64 `gorm:"column:uuid;type:bigint(20);default:0;comment:销售账单ID" json:"uuid"`
	OrderNo string `gorm:"column:order_no;type:varchar(255);default:'';comment:销售账单编号" json:"order_no"`

	// 基本信息字段
	BillType       uint   `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 账单类型, 0-桌台订单、1-点餐订单" json:"bill_type"`
	DiningMethod   uint   `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	IsBuffet       uint   `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	MealNum        uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	TaxType        uint   `gorm:"column:tax_type;type:tinyint(1);default:0;comment:税费类型, 0-商品未含税 1-商品已含税,下单后不变" json:"tax_type"`
	BuffetDuration uint   `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒）" json:"buffet_duration"`
	Remark         string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	Reason         string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 状态相关字段
	Status uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-待付款、1-已完成、2-已取消" json:"status"`
	IsLock uint `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`

	// 金额相关字段
	OrderAmount          float64 `gorm:"column:order_amount;type:decimal(12,2);default:0;comment:订单总金额，关联销售订单的总金额之和" json:"order_amount"`
	ProductAmount        float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额，关联销售订单的商品金额之和" json:"product_amount"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额，支付金额-订单总金额=支付手续费" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费，多次支付的支付手续费之和" json:"payment_commission_fee"`

	// 关联ID字段
	ConsumerUuid    uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid     uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	BuffetOrderUuid uint64 `gorm:"column:buffet_order_uuid;type:bigint(20);default:0;comment:自助餐订单ID" json:"buffet_order_uuid"`
	DeskUuid        uint64 `gorm:"column:desk_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"desk_uuid"`
	SerialNo        string `gorm:"column:serial_no;type:varchar(255);default:'';comment:桌位编号 (点餐流水号)" json:"serial_no"`

	// 时间相关字段
	HideBillTime uint `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`
	FinishTime   uint `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	CreateTime   uint `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime   uint `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime   uint `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`

	// 关联字段
	SaleOrders []SaleOrder `gorm:"foreignKey:SaleBillUuid;references:uuid"`
}

// 销售订单 SaleOrder
type SaleOrder struct {
	// 主键和标识字段
	ID      uint   `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid    uint64 `gorm:"column:uuid;type:bigint(20);default:0;comment:销售订单ID" json:"uuid"`
	OrderNo string `gorm:"column:order_no;type:varchar(255);default:'';comment:订单编号" json:"order_no"`

	// 状态字段
	Status uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	IsGift uint `gorm:"column:is_gift;type:tinyint(1);default:0;comment:是否免单, 0-否 1-是" json:"is_gift"`

	// 金额相关字段
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:商品原始金额" json:"product_original_amount"`
	ServiceFee            float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费" json:"service_fee"`
	TaxFee                float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费" json:"tax_fee"`
	DiscountFee           float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用" json:"discount_fee"`
	MemberDiscountFee     float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用" json:"member_discount_fee"`
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:总金额" json:"amount"`

	// 关联ID字段
	ConsumerUuid uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid  uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 时间相关字段
	FinishTime uint `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	CreateTime uint `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime uint `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime uint `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`
}
