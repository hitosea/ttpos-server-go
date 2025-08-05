// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleBillDao is the data access object for the table ttpos_sale_bill.
type SaleBillDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SaleBillColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SaleBillColumns defines and stores column names for the table ttpos_sale_bill.
type SaleBillColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 销售账单ID
	OrderNo               string // 销售账单编号
	DutyNo                string // 当班编号,用于标记该账单属于哪个当班
	SerialNo              string // 桌位编号 (点餐流水号)
	BillType              string // 账单类型, 0-桌台订单、1-点餐订单
	DiningMethod          string // 用餐方式,0-堂食(店内就餐) 1-打包
	IsBuffet              string // 是否自助餐, 0-否 1-是
	Reason                string // 取消原因
	IsLock                string // 是否锁单, 0-否 1-是
	MealNum               string // 就餐人数
	Status                string // 订单状态, 0-待付款、1-已完成、2-已取消。
	Remark                string // 备注(开台备注)
	CashierName           string // 收银员名称
	ConsumerUuid          string // 消费者ID
	CashierUuid           string // 收银员ID。系统自动创建的销售账单，收银员ID为0
	DeskUuid              string // 餐桌ID
	BuffetPackage1Uuid    string // 自助餐套餐1的uuid
	BuffetPackage2Uuid    string // 自助餐套餐2的uuid
	DeviceUuid            string // 设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询
	Amount                string // 订单金额,关联销售订单的总金额之和
	ProductAmount         string // 商品金额,关联销售订单的商品金额之和
	ServiceFee            string // 服务费,关联销售订单的服务费之和
	TaxFee                string // 税费,关联销售订单的税费之和
	CustomDiscountFee     string // 自定义折扣费用,关联销售订单的会员折扣费用之和
	MemberDiscountFee     string // 会员折扣费用,关联销售订单的会员折扣费用之和
	GiftAmount            string // 赠菜金额,关联销售订单的赠菜金额之和
	FreeAmount            string // 免单金额,关联销售订单的免单金额之和
	PaymentCommissionFee  string // 支付手续费,多次支付的支付手续费之和
	PaymentAmount         string // 支付金额,支付金额-订单总金额=支付手续费
	ProductOriginalAmount string // 原始商品金额。 商品原始金额=(订单.原始商品金额)之和。
	ShowMustPlan          string // 是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0
	AutoAddMustProduct    string // 是否自动加购必点商品, 0-不自动加购 1-自动加购.自动将商品加入购物车后改值为0
	TaxType               string // 税费类型, 0-商品未含税 1-商品已含税,下单后不变
	BuffetDuration        string // 自助餐可用时长(秒)
	NonOrderingTime       string // 自助餐结束前x分钟时不可下单，用于助手端、平板端和h5
	ReminderOrderTime     string // 自助餐结束前x分钟时提醒不可下单，用于助手端、平板端和h5
	BuffetStartTime       string // 自助餐开始时间(秒)
	DelayDuration         string // 总延迟时长(秒)
	DelayStartTime        string // 总延迟时长开始时间(秒)
	HideBillTime          string // 隐藏账单(挂单)时间(时间戳)
	ProductionTime        string // 首次送厨时间(时间戳)
	FinishTime            string // 完成时间(时间戳),结账时间
	CreateTime            string // 创建时间(时间戳),开台时间
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// saleBillColumns holds the columns for the table ttpos_sale_bill.
var saleBillColumns = SaleBillColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	OrderNo:               "order_no",
	DutyNo:                "duty_no",
	SerialNo:              "serial_no",
	BillType:              "bill_type",
	DiningMethod:          "dining_method",
	IsBuffet:              "is_buffet",
	Reason:                "reason",
	IsLock:                "is_lock",
	MealNum:               "meal_num",
	Status:                "status",
	Remark:                "remark",
	CashierName:           "cashier_name",
	ConsumerUuid:          "consumer_uuid",
	CashierUuid:           "cashier_uuid",
	DeskUuid:              "desk_uuid",
	BuffetPackage1Uuid:    "buffet_package1_uuid",
	BuffetPackage2Uuid:    "buffet_package2_uuid",
	DeviceUuid:            "device_uuid",
	Amount:                "amount",
	ProductAmount:         "product_amount",
	ServiceFee:            "service_fee",
	TaxFee:                "tax_fee",
	CustomDiscountFee:     "custom_discount_fee",
	MemberDiscountFee:     "member_discount_fee",
	GiftAmount:            "gift_amount",
	FreeAmount:            "free_amount",
	PaymentCommissionFee:  "payment_commission_fee",
	PaymentAmount:         "payment_amount",
	ProductOriginalAmount: "product_original_amount",
	ShowMustPlan:          "show_must_plan",
	AutoAddMustProduct:    "auto_add_must_product",
	TaxType:               "tax_type",
	BuffetDuration:        "buffet_duration",
	NonOrderingTime:       "non_ordering_time",
	ReminderOrderTime:     "reminder_order_time",
	BuffetStartTime:       "buffet_start_time",
	DelayDuration:         "delay_duration",
	DelayStartTime:        "delay_start_time",
	HideBillTime:          "hide_bill_time",
	ProductionTime:        "production_time",
	FinishTime:            "finish_time",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewSaleBillDao creates and returns a new DAO object for table data access.
func NewSaleBillDao(handlers ...gdb.ModelHandler) *SaleBillDao {
	return &SaleBillDao{
		group:    "default",
		table:    "ttpos_sale_bill",
		columns:  saleBillColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleBillDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleBillDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleBillDao) Columns() SaleBillColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleBillDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleBillDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SaleBillDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
