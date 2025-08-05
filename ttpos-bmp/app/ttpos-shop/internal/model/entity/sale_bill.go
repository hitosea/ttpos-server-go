// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleBill is the golang structure for table sale_bill.
type SaleBill struct {
	Id                    uint    `json:"id"                    orm:"id"                      description:"自增ID"`                                       // 自增ID
	Uuid                  int64   `json:"uuid"                  orm:"uuid"                    description:"销售账单ID"`                                     // 销售账单ID
	OrderNo               string  `json:"orderNo"               orm:"order_no"                description:"销售账单编号"`                                     // 销售账单编号
	DutyNo                string  `json:"dutyNo"                orm:"duty_no"                 description:"当班编号,用于标记该账单属于哪个当班"`                         // 当班编号,用于标记该账单属于哪个当班
	SerialNo              string  `json:"serialNo"              orm:"serial_no"               description:"桌位编号 (点餐流水号)"`                               // 桌位编号 (点餐流水号)
	BillType              int     `json:"billType"              orm:"bill_type"               description:"账单类型, 0-桌台订单、1-点餐订单"`                        // 账单类型, 0-桌台订单、1-点餐订单
	DiningMethod          int     `json:"diningMethod"          orm:"dining_method"           description:"用餐方式,0-堂食(店内就餐) 1-打包"`                       // 用餐方式,0-堂食(店内就餐) 1-打包
	IsBuffet              int     `json:"isBuffet"              orm:"is_buffet"               description:"是否自助餐, 0-否 1-是"`                             // 是否自助餐, 0-否 1-是
	Reason                string  `json:"reason"                orm:"reason"                  description:"取消原因"`                                       // 取消原因
	IsLock                int     `json:"isLock"                orm:"is_lock"                 description:"是否锁单, 0-否 1-是"`                              // 是否锁单, 0-否 1-是
	MealNum               int     `json:"mealNum"               orm:"meal_num"                description:"就餐人数"`                                       // 就餐人数
	Status                int     `json:"status"                orm:"status"                  description:"订单状态, 0-待付款、1-已完成、2-已取消。"`                   // 订单状态, 0-待付款、1-已完成、2-已取消。
	Remark                string  `json:"remark"                orm:"remark"                  description:"备注(开台备注)"`                                   // 备注(开台备注)
	CashierName           string  `json:"cashierName"           orm:"cashier_name"            description:"收银员名称"`                                      // 收银员名称
	ConsumerUuid          int64   `json:"consumerUuid"          orm:"consumer_uuid"           description:"消费者ID"`                                      // 消费者ID
	CashierUuid           int64   `json:"cashierUuid"           orm:"cashier_uuid"            description:"收银员ID。系统自动创建的销售账单，收银员ID为0"`                  // 收银员ID。系统自动创建的销售账单，收银员ID为0
	DeskUuid              int64   `json:"deskUuid"              orm:"desk_uuid"               description:"餐桌ID"`                                       // 餐桌ID
	BuffetPackage1Uuid    int64   `json:"buffetPackage1Uuid"    orm:"buffet_package1_uuid"    description:"自助餐套餐1的uuid"`                                // 自助餐套餐1的uuid
	BuffetPackage2Uuid    int64   `json:"buffetPackage2Uuid"    orm:"buffet_package2_uuid"    description:"自助餐套餐2的uuid"`                                // 自助餐套餐2的uuid
	DeviceUuid            int64   `json:"deviceUuid"            orm:"device_uuid"             description:"设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询"`      // 设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询
	Amount                float64 `json:"amount"                orm:"amount"                  description:"订单金额,关联销售订单的总金额之和"`                          // 订单金额,关联销售订单的总金额之和
	ProductAmount         float64 `json:"productAmount"         orm:"product_amount"          description:"商品金额,关联销售订单的商品金额之和"`                         // 商品金额,关联销售订单的商品金额之和
	ServiceFee            float64 `json:"serviceFee"            orm:"service_fee"             description:"服务费,关联销售订单的服务费之和"`                           // 服务费,关联销售订单的服务费之和
	TaxFee                float64 `json:"taxFee"                orm:"tax_fee"                 description:"税费,关联销售订单的税费之和"`                             // 税费,关联销售订单的税费之和
	CustomDiscountFee     float64 `json:"customDiscountFee"     orm:"custom_discount_fee"     description:"自定义折扣费用,关联销售订单的会员折扣费用之和"`                    // 自定义折扣费用,关联销售订单的会员折扣费用之和
	MemberDiscountFee     float64 `json:"memberDiscountFee"     orm:"member_discount_fee"     description:"会员折扣费用,关联销售订单的会员折扣费用之和"`                     // 会员折扣费用,关联销售订单的会员折扣费用之和
	GiftAmount            float64 `json:"giftAmount"            orm:"gift_amount"             description:"赠菜金额,关联销售订单的赠菜金额之和"`                         // 赠菜金额,关联销售订单的赠菜金额之和
	FreeAmount            float64 `json:"freeAmount"            orm:"free_amount"             description:"免单金额,关联销售订单的免单金额之和"`                         // 免单金额,关联销售订单的免单金额之和
	PaymentCommissionFee  float64 `json:"paymentCommissionFee"  orm:"payment_commission_fee"  description:"支付手续费,多次支付的支付手续费之和"`                         // 支付手续费,多次支付的支付手续费之和
	PaymentAmount         float64 `json:"paymentAmount"         orm:"payment_amount"          description:"支付金额,支付金额-订单总金额=支付手续费"`                      // 支付金额,支付金额-订单总金额=支付手续费
	ProductOriginalAmount float64 `json:"productOriginalAmount" orm:"product_original_amount" description:"原始商品金额。 商品原始金额=(订单.原始商品金额)之和。"`              // 原始商品金额。 商品原始金额=(订单.原始商品金额)之和。
	ShowMustPlan          int     `json:"showMustPlan"          orm:"show_must_plan"          description:"是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0"`       // 是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0
	AutoAddMustProduct    int     `json:"autoAddMustProduct"    orm:"auto_add_must_product"   description:"是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0"` // 是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0
	TaxType               int     `json:"taxType"               orm:"tax_type"                description:"税费类型, 0-商品未含税 1-商品已含税,下单后不变"`                // 税费类型, 0-商品未含税 1-商品已含税,下单后不变
	BuffetDuration        int     `json:"buffetDuration"        orm:"buffet_duration"         description:"自助餐可用时长(秒)"`                                 // 自助餐可用时长(秒)
	NonOrderingTime       int     `json:"nonOrderingTime"       orm:"non_ordering_time"       description:"自助餐结束前x分钟时不可下单，用于助手端、平板端和h5"`                // 自助餐结束前x分钟时不可下单，用于助手端、平板端和h5
	ReminderOrderTime     int     `json:"reminderOrderTime"     orm:"reminder_order_time"     description:"自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5"`              // 自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5
	BuffetStartTime       int     `json:"buffetStartTime"       orm:"buffet_start_time"       description:"自助餐开始时间(秒)"`                                 // 自助餐开始时间(秒)
	DelayDuration         int     `json:"delayDuration"         orm:"delay_duration"          description:"总延迟时长(秒)"`                                   // 总延迟时长(秒)
	DelayStartTime        int     `json:"delayStartTime"        orm:"delay_start_time"        description:"总延迟时长开始时间(秒)"`                               // 总延迟时长开始时间(秒)
	HideBillTime          int     `json:"hideBillTime"          orm:"hide_bill_time"          description:"隐藏账单(挂单)时间(时间戳)"`                            // 隐藏账单(挂单)时间(时间戳)
	ProductionTime        int     `json:"productionTime"        orm:"production_time"         description:"首次送厨时间(时间戳)"`                                // 首次送厨时间(时间戳)
	FinishTime            int     `json:"finishTime"            orm:"finish_time"             description:"完成时间(时间戳),结账时间"`                             // 完成时间(时间戳),结账时间
	CreateTime            uint    `json:"createTime"            orm:"create_time"             description:"创建时间(时间戳),开台时间"`                             // 创建时间(时间戳),开台时间
	UpdateTime            uint    `json:"updateTime"            orm:"update_time"             description:"更新时间(时间戳)"`                                  // 更新时间(时间戳)
	DeleteTime            uint    `json:"deleteTime"            orm:"delete_time"             description:"删除时间(时间戳)"`                                  // 删除时间(时间戳)
}
