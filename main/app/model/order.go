package model

// 销售账单 SaleBill
type SaleBill struct {
	ID              uint    `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid            uint64  `gorm:"column:uuid;type:bigint(20);default:0;comment:销售账单ID" json:"uuid"`
	Sn              string  `gorm:"column:sn;type:varchar(255);default:'';comment:订单编号" json:"sn"`
	BillType        uint    `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-Desk桌台订单、1-OrderingFood点餐订单" json:"bill_type"`
	DiningMethod    uint    `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式, 0-Takeout打包、1-DineIn堂食" json:"dining_method"`
	IsBuffet        uint    `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	IsLock          uint    `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`
	MealNum         uint    `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Status          uint    `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-Pending待处理、1-Processing处理中、2-Completed已完成、3-Cancelled已取消、4-Failed失败" json:"status"`
	Reason          string  `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`
	Remark          string  `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	OrderAmount     float64 `gorm:"column:order_amount;type:decimal(12,2);default:0;comment:订单总金额" json:"order_amount"`
	ProductAmount   float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	PaymentAmount   float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额" json:"payment_amount"`
	ConsumerUuid    uint64  `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid     uint64  `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	BuffetOrderUUID uint64  `gorm:"column:buffet_order_uuid;type:bigint(20);default:0;comment:自助餐订单ID" json:"buffet_order_uuid"`
	TableUuid       uint64  `gorm:"column:table_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"table_uuid"`
	BuffetDuration  uint    `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒）" json:"buffet_duration"`
	HideBillTime    uint    `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`
	FinishTime      uint    `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	CreateTime      uint    `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime      uint    `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime      uint    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`

	SaleOrders []SaleOrder `gorm:"foreignKey:SaleBillUUID;references:UUID"`
}

// 销售订单 SaleOrder
type SaleOrder struct {
	ID                    uint    `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid                  uint64  `gorm:"column:uuid;type:bigint(20);default:0;comment:销售订单ID" json:"uuid"`
	OrderNo               string  `gorm:"column:order_no;type:varchar(255);default:'';comment:订单编号" json:"order_no"`
	IsBuffet              bool    `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	Type                  uint    `gorm:"column:type;type:tinyint(1);default:0;comment:销售订单类型, 0-桌台订单 1-扫码订单" json:"type"`
	Status                uint    `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-未结账 1-已结账" json:"status"`
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:商品原始金额" json:"product_original_amount"`
	ServiceFee            float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费" json:"service_fee"`
	TaxFee                float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费" json:"tax_fee"`
	DiscountFee           float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用" json:"discount_fee"`
	MemberDiscountFee     float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用" json:"member_discount_fee"`
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:总金额" json:"amount"`
	IsGift                uint    `gorm:"column:is_gift;type:tinyint(1);default:0;comment:是否免单, 0-否 1-是" json:"is_gift"`
	ConsumerUuid          uint64  `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid           uint64  `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	SaleBillUuid          uint64  `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	HandleTime            uint    `gorm:"column:handle_time;type:int(10);default:0;comment:接单时间（时间戳）" json:"handle_time"`
	FinishTime            uint    `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	CreateTime            uint    `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime            uint    `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime            uint    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`
}
