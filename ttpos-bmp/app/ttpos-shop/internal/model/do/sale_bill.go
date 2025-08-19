// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleBill is the golang structure of table ttpos_sale_bill for DAO operations like Where/Data.
type SaleBill struct {
	g.Meta                `orm:"table:ttpos_sale_bill, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 销售账单ID
	OrderNo               interface{} // 销售账单编号
	DutyNo                interface{} // 当班编号,用于标记该账单属于哪个当班
	SerialNo              interface{} // 桌位编号 (点餐流水号)
	BillType              interface{} // 账单类型, 0-桌台订单、1-点餐订单
	DiningMethod          interface{} // 用餐方式,0-堂食(店内就餐) 1-打包
	IsBuffet              interface{} // 是否自助餐, 0-否 1-是
	Reason                interface{} // 取消原因
	IsLock                interface{} // 是否锁单, 0-否 1-是
	MealNum               interface{} // 就餐人数
	Status                interface{} // 订单状态, 0-待付款、1-已完成、2-已取消。
	Remark                interface{} // 备注(开台备注)
	CashierName           interface{} // 收银员名称
	ConsumerUuid          interface{} // 消费者ID
	CashierUuid           interface{} // 收银员ID。系统自动创建的销售账单，收银员ID为0
	DeskUuid              interface{} // 餐桌ID
	BuffetPackage1Uuid    interface{} // 自助餐套餐1的uuid
	BuffetPackage2Uuid    interface{} // 自助餐套餐2的uuid
	DeviceUuid            interface{} // 设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询
	Amount                interface{} // 订单金额,关联销售订单的总金额之和
	ProductAmount         interface{} // 商品金额,关联销售订单的商品金额之和
	ServiceFee            interface{} // 服务费,关联销售订单的服务费之和
	TaxFee                interface{} // 税费,关联销售订单的税费之和
	CustomDiscountFee     interface{} // 自定义折扣费用,关联销售订单的会员折扣费用之和
	MemberDiscountFee     interface{} // 会员折扣费用,关联销售订单的会员折扣费用之和
	GiftAmount            interface{} // 赠菜金额,关联销售订单的赠菜金额之和
	FreeAmount            interface{} // 免单金额,关联销售订单的免单金额之和
	PaymentCommissionFee  interface{} // 支付手续费,多次支付的支付手续费之和
	PaymentAmount         interface{} // 支付金额,支付金额-订单总金额=支付手续费
	ProductOriginalAmount interface{} // 原始商品金额。 商品原始金额=(订单.原始商品金额)之和。
	ShowMustPlan          interface{} // 是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0
	AutoAddMustProduct    interface{} // 是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0
	TaxType               interface{} // 税费类型, 0-商品未含税 1-商品已含税,下单后不变
	BuffetDuration        interface{} // 自助餐可用时长(秒)
	NonOrderingTime       interface{} // 自助餐结束前x分钟时不可下单，用于助手端、平板端和h5
	ReminderOrderTime     interface{} // 自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5
	BuffetStartTime       interface{} // 自助餐开始时间(秒)
	DelayDuration         interface{} // 总延迟时长(秒)
	DelayStartTime        interface{} // 总延迟时长开始时间(秒)
	HideBillTime          interface{} // 隐藏账单(挂单)时间(时间戳)
	ProductionTime        interface{} // 首次送厨时间(时间戳)
	FinishTime            interface{} // 完成时间(时间戳),结账时间
	CreateTime            interface{} // 创建时间(时间戳),开台时间
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
