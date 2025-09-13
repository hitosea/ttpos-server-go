package model

import "database/sql"

// StatisticsSale 销售统计表 ttpos_statistics_sale
type StatisticsSale struct {
	BaseModel
	SaleBillUuid         uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid        uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid;NOT NULL" json:"sale_order_uuid"`
	DutyNo               string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	DeskUuid             uint64  `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台uuid;NOT NULL" json:"desk_uuid"`
	MealNum              int     `gorm:"column:meal_num;type:int(11);default:0;comment:用餐人数;NOT NULL" json:"meal_num"`
	ProductPrice         float64 `gorm:"column:product_price;type:decimal(14,2);default:0.00;comment:商品原价: 不含税;NOT NULL" json:"product_price"`
	ProductOriginPrice   float64 `gorm:"column:product_origin_price;type:decimal(14,2);default:0.00;comment:原商品金额;NOT NULL" json:"product_origin_price"`
	ProductSalePrice     float64 `gorm:"column:product_sale_price;type:decimal(14,2);default:0.00;comment:商品销售价;NOT NULL" json:"product_sale_price"`
	ProductNum           float64 `gorm:"column:product_num;type:decimal(14,2);default:0.00;comment:商品数量;NOT NULL" json:"product_num"`
	ProductTax           float64 `gorm:"column:product_tax;type:decimal(14,2);default:0.00;comment:商品税;NOT NULL" json:"product_tax"`
	ServiceFee           float64 `gorm:"column:service_fee;type:decimal(14,2);default:0.00;comment:服务费;NOT NULL" json:"service_fee"`
	ServiceTax           float64 `gorm:"column:service_tax;type:decimal(14,2);default:0.00;comment:服务税;NOT NULL" json:"service_tax"`
	Discount             float64 `gorm:"column:discount;type:decimal(14,2);default:0.00;comment:优惠折扣;NOT NULL" json:"discount"`
	DiscountMember       float64 `gorm:"column:discount_member;type:decimal(14,2);default:0.00;comment:会员折扣;NOT NULL" json:"discount_member"`
	GiftAmount           float64 `gorm:"column:gift_amount;type:decimal(14,2);default:0.00;comment:赠菜金额;NOT NULL" json:"gift_amount"`
	GiftNum              float64 `gorm:"column:gift_num;type:decimal(14,2);default:0.00;comment:赠菜数量;NOT NULL" json:"gift_num"`
	FreeAmount           float64 `gorm:"column:free_amount;type:decimal(14,2);default:0.00;comment:免单金额;NOT NULL" json:"free_amount"`
	FreeNum              float64 `gorm:"column:free_num;type:decimal(14,2);default:0.00;comment:免单数量;NOT NULL" json:"free_num"`
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(14,2);default:0.00;comment:支付金额;NOT NULL" json:"payment_amount"`
	PaymentFee           float64 `gorm:"column:payment_fee;type:decimal(14,2);default:0.00;comment:支付手续费;NOT NULL" json:"payment_fee"`
	PaymentBalance       float64 `gorm:"column:payment_balance;type:decimal(14,2);default:0.00;comment:支付余额;NOT NULL" json:"payment_balance"`
	DeliveryFee          float64 `gorm:"column:delivery_fee;type:decimal(14,2);default:0.00;comment:配送费;NOT NULL" json:"delivery_fee"`
	RefundAmount         float64 `gorm:"column:refund_amount;type:decimal(14,2);default:0.00;comment:退款金额;NOT NULL" json:"refund_amount"`
	RefundPaymentBalance float64 `gorm:"column:refund_payment_balance;type:decimal(14,2);default:0.00;comment:退款支付余额;NOT NULL" json:"refund_payment_balance"`
	RefundTax            float64 `gorm:"column:refund_tax;type:decimal(14,2);default:0.00;comment:退款税费;NOT NULL" json:"refund_tax"`
	NoRefundTax          float64 `gorm:"column:no_refund_tax;type:decimal(14,2);default:0.00;comment:不退税金额;NOT NULL" json:"no_refund_tax"`
	ExtendPrice          float64 `gorm:"column:extend_price;type:decimal(14,2);default:0.00;comment:扩展价格;NOT NULL" json:"extend_price"`
	IsMeger              int     `gorm:"column:is_meger;type:int(11);default:0;comment:是否合单;NOT NULL" json:"is_meger"`
	IsSpecial            int     `gorm:"column:is_special;type:int(11);default:0;comment:是否特殊订单;NOT NULL" json:"is_special"`
	IsTakeout            int     `gorm:"column:is_takeout;type:int(11);default:0;comment:是否外送;NOT NULL" json:"is_takeout"`
	RefundServiceFee     float64 `gorm:"column:refund_service_fee;type:decimal(14,2);default:0.00;comment:退款服务费;NOT NULL" json:"refund_service_fee"`
	RefundDiscount       float64 `gorm:"column:refund_discount;type:decimal(14,2);default:0.00;comment:退款优惠折扣;NOT NULL" json:"refund_discount"`
	RefundDiscountMember float64 `gorm:"column:refund_discount_member;type:decimal(14,2);default:0.00;comment:退款会员折扣;NOT NULL" json:"refund_discount_member"`
	RefundFee            float64 `gorm:"column:refund_fee;type:decimal(14,2);default:0.00;comment:退款支付手续费;NOT NULL" json:"refund_fee"`
	CompleteTime         int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
}

// StatisticsPayment 支付统计表 ttpos_statistics_payment
type StatisticsPayment struct {
	BaseModel
	SaleBillUuid      uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid     uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid;NOT NULL" json:"sale_order_uuid"`
	DutyNo            string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	DeskUuid          uint64  `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台uuid;NOT NULL" json:"desk_uuid"`
	PaymentMethodUuid uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:支付方式uuid;NOT NULL" json:"payment_method_uuid"`
	PaymentAmount     float64 `gorm:"column:payment_amount;type:decimal(14,2);default:0.00;comment:支付金额;NOT NULL" json:"payment_amount"`
	RefundAmount      float64 `gorm:"column:refund_amount;type:decimal(14,2);default:0.00;comment:退款金额;NOT NULL" json:"refund_amount"`
	CompleteTime      int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
}

// StatisticsProduct 商品统计表 ttpos_statistics_product
type StatisticsProduct struct {
	BaseModel
	SaleBillUuid            uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid           uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid;NOT NULL" json:"sale_order_uuid"`
	DutyNo                  string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	DeskUuid                uint64  `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台uuid;NOT NULL" json:"desk_uuid"`
	ProductPackageUuid      uint64  `gorm:"column:product_package_uuid;type:bigint(20) unsigned;default:0;comment:商品包uuid;NOT NULL" json:"product_package_uuid"`
	ProductBomUuid          uint64  `gorm:"column:product_bom_uuid;type:bigint(20) unsigned;default:0;comment:商品清单uuid;NOT NULL" json:"product_bom_uuid"`
	ProductPrice            float64 `gorm:"column:product_price;type:decimal(14,2);default:0.00;comment:商品单价: 未含税;NOT NULL" json:"product_price"`
	ProductSalePrice        float64 `gorm:"column:product_sale_price;type:decimal(14,2);default:0.00;comment:商品单价: 规格+加料;NOT NULL" json:"product_sale_price"`
	ProductFinalPrice       float64 `gorm:"column:product_final_price;type:decimal(14,2);default:0.00;comment:商品最终单价;NOT NULL" json:"product_final_price"`
	FlavorPrice             float64 `gorm:"column:flavor_price;type:decimal(14,2);default:0.00;comment:商品原价(仅规格);NOT NULL" json:"flavor_price"`
	SaucePrice              float64 `gorm:"column:sauce_price;type:decimal(14,2);default:0.00;comment:加料价格;NOT NULL" json:"sauce_price"`
	ProductNum              float64 `gorm:"column:product_num;type:decimal(14,2);default:0.00;comment:商品数量;NOT NULL" json:"product_num"`
	TaxRate                 float64 `gorm:"column:tax_rate;type:decimal(14,2);default:0.00;comment:税率;NOT NULL" json:"tax_rate"`
	TaxFee                  float64 `gorm:"column:tax_fee;type:decimal(14,2);default:0.00;comment:税费;NOT NULL" json:"tax_fee"`
	ServiceFee              float64 `gorm:"column:service_fee;type:decimal(14,2);default:0.00;comment:服务费;NOT NULL" json:"service_fee"`
	ServiceTax              float64 `gorm:"column:service_tax;type:decimal(14,2);default:0.00;comment:服务税;NOT NULL" json:"service_tax"`
	GiveNum                 float64 `gorm:"column:give_num;type:decimal(14,2);default:0.00;comment:赠菜数量;NOT NULL" json:"give_num"`
	FreeNum                 float64 `gorm:"column:free_num;type:decimal(14,2);default:0.00;comment:免单数量;NOT NULL" json:"free_num"`
	RefundNum               float64 `gorm:"column:refund_num;type:decimal(14,2);default:0.00;comment:退款数量;NOT NULL" json:"refund_num"`
	CompleteTime            int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
	RefundTime              int64   `gorm:"column:refund_time;type:bigint(20);default:0;comment:退款时间;NOT NULL" json:"refund_time"`
	IsTakeout               int     `gorm:"column:is_takeout;type:int(11);default:0;comment:是否外送;NOT NULL" json:"is_takeout"`
	MemberOrderDiscountRate float64 `gorm:"column:member_order_discount_rate;type:decimal(12,2);not null;default:1.00;comment:'会员订单折扣率(1-300%)'" json:"member_order_discount_rate"`
}

// StatisticsCustomerType 客户类型统计表 statistics_customer_type
type StatisticsCustomerType struct {
	BaseModel
	SaleBillUuid                uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid               uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid;NOT NULL" json:"sale_order_uuid"`
	DutyNo                      string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	DeskUuid                    uint64  `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台uuid;NOT NULL" json:"desk_uuid"`
	BuffetPackageUuid           uint64  `gorm:"column:buffet_package_uuid;type:bigint(20) unsigned;default:0;comment:自助餐套餐ID;NOT NULL" json:"buffet_package_uuid"`
	BuffetCustomerTypePriceUuid uint64  `gorm:"column:buffet_customer_type_price_uuid;type:bigint(20) unsigned;default:0;comment:自助餐客户类型价格ID;NOT NULL" json:"buffet_customer_type_price_uuid"`
	ProductPrice                float64 `gorm:"column:product_price;type:decimal(14,2);default:0.00;comment:销售价,未含税价格（折前）;NOT NULL" json:"product_price"`
	ProductSalePrice            float64 `gorm:"column:product_sale_price;type:decimal(14,2);default:0.00;comment:原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变;NOT NULL" json:"product_sale_price"`
	ProductNum                  float64 `gorm:"column:product_num;type:decimal(14,2);default:0.00;comment:商品数量;NOT NULL" json:"product_num"`
	TaxRate                     float64 `gorm:"column:tax_rate;type:decimal(14,2);default:0.00;comment:税率;NOT NULL" json:"tax_rate"`
	TaxFee                      float64 `gorm:"column:tax_fee;type:decimal(14,2);default:0.00;comment:税费;NOT NULL" json:"tax_fee"`
	ServiceFee                  float64 `gorm:"column:service_fee;type:decimal(14,2);default:0.00;comment:服务费;NOT NULL" json:"service_fee"`
	ServiceTax                  float64 `gorm:"column:service_tax;type:decimal(14,2);default:0.00;comment:服务税;NOT NULL" json:"service_tax"`
	GiveNum                     float64 `gorm:"column:give_num;type:decimal(14,2);default:0.00;comment:赠菜数量;NOT NULL" json:"give_num"`
	FreeNum                     float64 `gorm:"column:free_num;type:decimal(14,2);default:0.00;comment:免单数量;NOT NULL" json:"free_num"`
	RefundNum                   float64 `gorm:"column:refund_num;type:decimal(14,2);default:0.00;comment:退款数量;NOT NULL" json:"refund_num"`
	CompleteTime                int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
	RefundTime                  int64   `gorm:"column:refund_time;type:bigint(20);default:0;comment:退款时间;NOT NULL" json:"refund_time"`
}

// StatisticsDelay 加钟统计表 statistics_delay
type StatisticsDelay struct {
	BaseModel
	SaleBillUuid    uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid   uint64  `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid;NOT NULL" json:"sale_order_uuid"`
	DutyNo          string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	DeskUuid        uint64  `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台uuid;NOT NULL" json:"desk_uuid"`
	BuffetDelayUuid uint64  `gorm:"column:buffet_delay_uuid;type:bigint(20) unsigned;default:0;comment:自助餐加钟价格ID;NOT NULL" json:"buffet_delay_uuid"`
	ProductPrice    float64 `gorm:"column:product_price;type:decimal(14,2);default:0.00;comment:价格,下单时固定不受后台改变，结账时再检查是否改变;NOT NULL" json:"product_price"`
	ProductNum      float64 `gorm:"column:product_num;type:decimal(14,2);default:0.00;comment:商品数量;NOT NULL" json:"product_num"`
	TaxRate         float64 `gorm:"column:tax_rate;type:decimal(14,2);default:0.00;comment:税率;NOT NULL" json:"tax_rate"`
	TaxFee          float64 `gorm:"column:tax_fee;type:decimal(14,2);default:0.00;comment:税费;NOT NULL" json:"tax_fee"`
	ServiceFee      float64 `gorm:"column:service_fee;type:decimal(14,2);default:0.00;comment:服务费;NOT NULL" json:"service_fee"`
	ServiceTax      float64 `gorm:"column:service_tax;type:decimal(14,2);default:0.00;comment:服务税;NOT NULL" json:"service_tax"`
	GiveNum         float64 `gorm:"column:give_num;type:decimal(14,2);default:0.00;comment:赠菜数量;NOT NULL" json:"give_num"`
	FreeNum         float64 `gorm:"column:free_num;type:decimal(14,2);default:0.00;comment:免单数量;NOT NULL" json:"free_num"`
	RefundNum       float64 `gorm:"column:refund_num;type:decimal(14,2);default:0.00;comment:退款数量;NOT NULL" json:"refund_num"`
	CompleteTime    int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
	RefundTime      int64   `gorm:"column:refund_time;type:bigint(20);default:0;comment:退款时间;NOT NULL" json:"refund_time"`
}

// StatisticsMember 会员统计表 ttpos_statistics_member
type StatisticsMember struct {
	BaseModel
	MemberRechargeOrderUuid uint64  `gorm:"column:member_recharge_order_uuid;type:bigint(20) unsigned;default:0;comment:会员充值订单uuid;NOT NULL" json:"member_recharge_order_uuid"`
	DutyNo                  string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	RechargeAmount          float64 `gorm:"column:recharge_amount;type:decimal(14,2);default:0.00;comment:充值金额;NOT NULL" json:"recharge_amount"`
	GiveAmount              float64 `gorm:"column:give_amount;type:decimal(14,2);default:0.00;comment:赠送金额;NOT NULL" json:"give_amount"`
	GivePoint               float64 `gorm:"column:give_point;type:decimal(14,2);default:0.00;comment:赠送积分;NOT NULL" json:"give_point"`
	PaymentAmount           float64 `gorm:"column:payment_amount;type:decimal(14,2);default:0.00;comment:支付金额;NOT NULL" json:"payment_amount"`
	PaymentFee              float64 `gorm:"column:payment_fee;type:decimal(14,2);default:0.00;comment:支付手续费;NOT NULL" json:"payment_fee"`
	RefundAmount            float64 `gorm:"column:refund_amount;type:decimal(14,2);default:0.00;comment:退款金额;NOT NULL" json:"refund_amount"`
	RefundFee               float64 `gorm:"column:refund_fee;type:decimal(14,2);default:0.00;comment:退款支付手续费;NOT NULL" json:"refund_fee"`
	CompleteTime            int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
}

// StatisticsMemberPayment 会员支付统计表 ttpos_statistics_member_payment
type StatisticsMemberPayment struct {
	BaseModel
	MemberRechargeOrderUuid uint64  `gorm:"column:member_recharge_order_uuid;type:bigint(20) unsigned;default:0;comment:会员充值订单uuid;NOT NULL" json:"member_recharge_order_uuid"`
	DutyNo                  string  `gorm:"column:duty_no;type:varchar(64);comment:当班编号;NOT NULL" json:"duty_no"`
	PaymentMethodUuid       uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:支付方式uuid;NOT NULL" json:"payment_method_uuid"`
	PaymentAmount           float64 `gorm:"column:payment_amount;type:decimal(14,2);default:0.00;comment:支付金额;NOT NULL" json:"payment_amount"`
	RefundAmount            float64 `gorm:"column:refund_amount;type:decimal(14,2);default:0.00;comment:退款金额;NOT NULL" json:"refund_amount"`
	CompleteTime            int64   `gorm:"column:complete_time;type:bigint(20);default:0;comment:完成时间;NOT NULL" json:"complete_time"`
}

// StatisticsSaleData 销售统计数据
type StatisticsSaleData struct {
	TotalSaleAmount            sql.NullFloat64 `gorm:"column:total_sale_amount;comment:总销售额"`
	TotalReceivedAmount        sql.NullFloat64 `gorm:"column:total_received_amount;comment:总实收金额"`
	TotalProductPrice          sql.NullFloat64 `gorm:"column:total_product_price;comment:总商品原价"`
	TotalProductOriginPrice    sql.NullFloat64 `gorm:"column:total_product_origin_price;comment:总原商品金额"`
	TotalProductNum            sql.NullFloat64 `gorm:"column:total_product_num;comment:总商品数量"`
	TotalDiscountMember        sql.NullFloat64 `gorm:"column:total_discount_member;comment:总会员折扣"`
	TotalBusinessAmount        sql.NullFloat64 `gorm:"column:total_business_amount;comment:总营业收入"`
	TotalServiceFee            sql.NullFloat64 `gorm:"column:total_service_fee;comment:总服务费"`
	TotalPaymentFee            sql.NullFloat64 `gorm:"column:total_payment_fee;comment:总支付手续费"`
	TotalTax                   sql.NullFloat64 `gorm:"column:total_tax;comment:总税额"`
	TotalRefundAmount          sql.NullFloat64 `gorm:"column:total_refund_amount;comment:总退款金额"`
	TotalDiscount              sql.NullFloat64 `gorm:"column:total_discount;comment:总优惠折扣"`
	TotalGiftAmount            sql.NullFloat64 `gorm:"column:total_gift_amount;comment:总赠菜金额"`
	TotalGiftNum               sql.NullFloat64 `gorm:"column:total_gift_num;comment:总赠菜数量"`
	TotalFreeAmount            sql.NullFloat64 `gorm:"column:total_free_amount;comment:总免单金额"`
	TotalFreeNum               sql.NullFloat64 `gorm:"column:total_free_num;comment:总免单数量"`
	TotalOrderNum              sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数量"`
	TotalTakeoutSaleAmount     sql.NullFloat64 `gorm:"total_takeout_sale_amount;comment:总外送销售"`
	TotalTakeoutBusinessAmount sql.NullFloat64 `gorm:"total_takeout_business_amount;comment:总外送营收"`
	TotalTakeoutRefundAmount   sql.NullFloat64 `gorm:"total_takeout_refund_amount;comment:总外送退款金额"`
	TotalTakeoutDeliveryFee    sql.NullFloat64 `gorm:"total_takeout_delivery_fee;comment:总外送配送费"`
	TotalDeskNum               sql.NullInt64   `gorm:"column:total_desk_num;comment:总桌台数量"`
	TotalDeskOrderAmount       sql.NullFloat64 `gorm:"column:total_desk_order_amount;comment:总桌台订单金额"`
	TotalMealNum               sql.NullInt64   `gorm:"column:total_meal_num;comment:总用餐人数"`
	TotalInstantOrderAmount    sql.NullFloat64 `gorm:"column:total_instant_order_amount;comment:总即时订单金额"`
	TotalInstantOrderNum       sql.NullInt64   `gorm:"column:total_instant_order_num;comment:总即时订单数量"`
	TotalTakeoutOrderAmount    sql.NullFloat64 `gorm:"column:total_takeout_order_amount;comment:总外送订单金额"`
	TotalTakeoutOrderNum       sql.NullInt64   `gorm:"column:total_takeout_order_num;comment:总外送订单数量"`
	MinOrderAmount             sql.NullFloat64 `gorm:"column:min_order_amount;comment:最小订单金额"`
	MaxOrderAmount             sql.NullFloat64 `gorm:"column:max_order_amount;comment:最大订单金额"`
	AvgOrderAmount             sql.NullFloat64 `gorm:"column:avg_order_amount;comment:平均订单金额"`
	MinDeskOrderAmount         sql.NullFloat64 `gorm:"column:min_desk_order_amount;comment:最小桌台订单金额"`
	MaxDeskOrderAmount         sql.NullFloat64 `gorm:"column:max_desk_order_amount;comment:最大桌台订单金额"`
	AvgDeskOrderAmount         sql.NullFloat64 `gorm:"column:avg_desk_order_amount;comment:平均桌台订单金额"`
	MinInstantOrderAmount      sql.NullFloat64 `gorm:"column:min_instant_order_amount;comment:最小即时订单金额"`
	MaxInstantOrderAmount      sql.NullFloat64 `gorm:"column:max_instant_order_amount;comment:最大即时订单金额"`
	AvgInstantOrderAmount      sql.NullFloat64 `gorm:"column:avg_instant_order_amount;comment:平均即时订单金额"`
	MinTakeoutOrderAmount      sql.NullFloat64 `gorm:"column:min_takeout_order_amount;comment:最小外送订单金额"`
	MaxTakeoutOrderAmount      sql.NullFloat64 `gorm:"column:max_takeout_order_amount;comment:最大外送订单金额"`
	AvgTakeoutOrderAmount      sql.NullFloat64 `gorm:"column:avg_takeout_order_amount;comment:平均外送订单金额"`
}

// StatisticsSaleDaysData 销售天数统计数据
type StatisticsSaleDaysData struct {
	StatisticsSaleData
	Day sql.NullString `gorm:"column:day;comment:日期"`
}

// StatisticsPaymentData 支付统计数据
type StatisticsPaymentData struct {
	ID                 uint64          `gorm:"column:id;comment:支付方式ID"`
	Sort               int             `gorm:"column:sort;comment:支付方式排序"`
	CreateTime         int64           `gorm:"column:create_time;comment:支付方式创建时间"`
	PaymentName        string          `gorm:"column:payment_name;comment:支付方式名称"`
	PaymentCode        int             `gorm:"column:payment_code;comment:支付方式编码"`
	ErpnextPayment     string          `gorm:"column:erpnext_payment;comment:ERPNext支付方式"`
	TotalOrderNum      sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数量"`
	TotalPaymentAmount sql.NullFloat64 `gorm:"column:total_payment_amount;comment:总支付金额"`
	TotalRefundAmount  sql.NullFloat64 `gorm:"column:total_refund_amount;comment:总退款金额"`
}

// StatisticsPaymentDaysData 支付天数统计数据
type StatisticsPaymentDaysData struct {
	StatisticsPaymentData
	Day sql.NullString `gorm:"column:day;comment:日期"`
}

// StatisticsTaxData 税类统计数据
type StatisticsTaxData struct {
	TaxRate            sql.NullFloat64 `gorm:"column:tax_rate;comment:税率"`
	TotalTaxFee        sql.NullFloat64 `gorm:"column:total_tax_fee;comment:总税费"`
	TotalProductAmount sql.NullFloat64 `gorm:"column:total_product_amount;comment:总商品金额: 含税"`
}

// StatisticsCategoryData 分类统计数据
type StatisticsCategoryData struct {
	CategoryParentUuid sql.NullInt64   `gorm:"column:category_parent_uuid;comment:分类父级uuid"`
	CategoryParentName sql.NullString  `gorm:"column:category_parent_name;comment:分类父级名称"`
	CategoryUuid       sql.NullInt64   `gorm:"column:category_uuid;comment:分类uuid"`
	CategoryName       sql.NullString  `gorm:"column:category_name;comment:分类名称"`
	SaleNum            sql.NullFloat64 `gorm:"column:sale_num;comment:销售数量"`
	SaleAmount         sql.NullFloat64 `gorm:"column:sale_amount;comment:销售金额"`
}

// StatisticsProductData 商品统计数据
type StatisticsProductData struct {
	ProductPackageUuid sql.NullInt64   `gorm:"column:product_package_uuid;comment:商品包uuid"` // 用于拿到排行榜数据后在查询商品名称
	ProductName        sql.NullString  `gorm:"column:product_name;comment:商品名称"`
	FlavorName         sql.NullString  `gorm:"column:flavor_name;comment:规格名称"`
	SalePrice          sql.NullFloat64 `gorm:"column:sale_price;comment:销售单价"`
	SaleNum            sql.NullFloat64 `gorm:"column:sale_num;comment:销售数量"`
	SaleAmount         sql.NullFloat64 `gorm:"column:sale_amount;comment:销售金额"`
	ProductType        sql.NullInt64   `gorm:"column:product_type;comment:商品类型"`
}

// StatisticsAreaData 区域统计数据
type StatisticsAreaData struct {
	AreaName           sql.NullString  `gorm:"column:area_name;comment:区域名称"`
	AreaSaleAmount     sql.NullFloat64 `gorm:"column:area_sale_amount;comment:区域销售额"`
	AreaBusinessAmount sql.NullFloat64 `gorm:"column:area_business_amount;comment:区域营业收入"`
	AreaProductNum     sql.NullFloat64 `gorm:"column:area_product_num;comment:区域商品数量"`
}

// StatisticsAreaDaysData 区域天数统计数据
type StatisticsAreaDaysData struct {
	StatisticsAreaData
	AreaID sql.NullInt64  `gorm:"column:area_id;comment:区域id"`
	Day    sql.NullString `gorm:"column:day;comment:日期"`
}

// Statistics7DaysData 7天统计数据
type Statistics7DaysData struct {
	Day                 sql.NullString  `gorm:"column:day;comment:日期"`
	TotalOrderNum       sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数"`
	TotalReceivedAmount sql.NullFloat64 `gorm:"column:total_received_amount;comment:总实收金额"`
}

// StatisticsUnpaidOrderData 未结订单统计数据
type StatisticsUnpaidOrderData struct {
	TotalOrderNum sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数"`
	TotalAmount   sql.NullFloat64 `gorm:"column:total_amount;comment:总金额"`
}

// StatisticsMemberData 会员统计数据
type StatisticsMemberData struct {
	TotalSaleAmount     sql.NullFloat64 `gorm:"column:total_sale_amount;comment:总销售额"`
	TotalRechargeAmount sql.NullFloat64 `gorm:"column:total_recharge_amount;comment:总充值金额"`
	TotalGiveAmount     sql.NullFloat64 `gorm:"column:total_give_amount;comment:总赠送金额"`
	TotalGivePoint      sql.NullFloat64 `gorm:"column:total_give_point;comment:总赠送积分"`
	TotalPaymentAmount  sql.NullFloat64 `gorm:"column:total_payment_amount;comment:总支付金额"`
	TotalPaymentFee     sql.NullFloat64 `gorm:"column:total_payment_fee;comment:总支付手续费"`
	TotalRefundAmount   sql.NullFloat64 `gorm:"column:total_refund_amount;comment:总退款金额"`
}

// StatisticsMemberDaysData 会员天数统计数据
type StatisticsMemberDaysData struct {
	StatisticsMemberData
	Day sql.NullString `gorm:"column:day;comment:日期"`
}

// StatisticsProductSaleData 商品销售统计数据
type StatisticsProductSaleData struct {
	ProductName        sql.NullString  `gorm:"column:product_name;comment:商品名称"`
	CategoryParentName sql.NullString  `gorm:"column:category_parent_name;comment:分类父级名称"`
	CategoryName       sql.NullString  `gorm:"column:category_name;comment:分类名称"`
	SaleNum            sql.NullFloat64 `gorm:"column:sale_num;comment:销售数量"`
	OriginSaleAmount   sql.NullFloat64 `gorm:"column:origin_sale_amount;comment:原价销售金额"`
	ActualSaleAmount   sql.NullFloat64 `gorm:"column:actual_sale_amount;comment:实际销售金额"`
	GiveNum            sql.NullFloat64 `gorm:"column:give_num;comment:赠菜数量"`
	BusinessAmount     sql.NullFloat64 `gorm:"column:business_amount;comment:营业收入"`
}

// StatisticsFreePaymentData 免单支付统计数据
type StatisticsFreePaymentData struct {
	TotalOrderNum   sql.NullInt64   `gorm:"column:total_order_num;comment:总订单数量"`
	TotalFreeAmount sql.NullFloat64 `gorm:"column:total_free_amount;comment:总免单金额"`
}

// StatisticsFreePaymentDaysData 免单支付天数统计数据
type StatisticsFreePaymentDaysData struct {
	StatisticsFreePaymentData
	Day sql.NullString `gorm:"column:day;comment:日期"`
}

// CountMemberNumDaysResp 会员数量天数统计数据
type CountMemberNumDaysResp struct {
	Day       sql.NullString `gorm:"column:day;comment:日期"`
	MemberNum sql.NullInt64  `gorm:"column:member_num;comment:会员数量"`
}

// StatisticsCancelOrderData 取消订单统计数据
type StatisticsCancelOrderData struct {
	TotalCancelOrderNum    sql.NullInt64   `gorm:"column:total_cancel_order_num;comment:总取消订单数"`
	TotalCancelOrderAmount sql.NullFloat64 `gorm:"column:total_cancel_order_amount;comment:总取消订单金额"`
}
