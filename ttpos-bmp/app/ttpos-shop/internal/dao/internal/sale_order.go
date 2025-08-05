// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderDao is the data access object for the table ttpos_sale_order.
type SaleOrderDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SaleOrderColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SaleOrderColumns defines and stores column names for the table ttpos_sale_order.
type SaleOrderColumns struct {
	Id                     string // 自增ID
	Uuid                   string // 销售订单ID
	OrderNo                string // 订单编号
	Status                 string // 订单状态, 0-未结账 1-已结账
	MemberDiscountFee      string // 总会员折扣金额。总会员折扣金额=(订单商品.会员折扣金额)之和
	CustomDiscountFee      string // 总自定义折扣金额。总自定义折扣金额=(订单商品.自定义折扣金额)之和
	ZeroFee                string // 优惠折扣抹零金额。
	ProductAmount          string // 商品金额，订单商品的最终单价(折后价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费
	ProductOriginalAmount  string // 原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。
	ServiceFee             string // 服务费固定服务费时，服务费=固定服务费；按比例收服务费时，服务费=(订单商品.总服务费)之和
	TaxFee                 string // 税费。税费=(订单商品.总税费)之和
	Amount                 string // 应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）
	OriginAmount           string // 原始应收金额。原始应收金额=商品金额+服务费+消费税。商品未含税时，原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。商品已含税时，原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。
	IsFree                 string // 是否免单, 0-否 1-是
	FreeReason             string // 免单原因
	MemberDiscountRate     string // 会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	MemberCardDiscountRate string // 会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	CustomDiscountRate     string // 自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	CustomAmount           string // 整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额
	ZeroRule               string // 优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
	ZeroCheckoutRule       string // 结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元
	PaymentAmount          string // 已支付金额,关联付款单的支付金额之和。
	ChangeAmount           string // 找零金额,结账完成后才记录
	ZeroCheckoutFee        string // 结账抹零金额。
	FinalPrice             string // 最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额
	PaymentCommissionFee   string // 支付手续费,关联付款单的支付手续费之和
	GiftAmount             string // 赠菜金额,(销售订单赠菜商品.总最终单价)之和
	GiftPoints             string // 赠送积分. 赠送积分=应收金额amount*积分赠送比例.
	GiftPointsRate         string // 赠送积分比例. 取值范围0-1。结账后记录，不受后台改变
	MemberBalance          string // 会员余额.会员消费本单后剩余的余额
	CashierName            string // 收银员名称
	ConsumerUuid           string // 消费者ID
	CashierUuid            string // 收银员ID
	SaleBillUuid           string // 销售账单ID
	FinishTime             string // 完成时间(时间戳),结账时间
	CreateTime             string // 创建时间(时间戳)
	UpdateTime             string // 更新时间(时间戳)
	DeleteTime             string // 删除时间(时间戳)
}

// saleOrderColumns holds the columns for the table ttpos_sale_order.
var saleOrderColumns = SaleOrderColumns{
	Id:                     "id",
	Uuid:                   "uuid",
	OrderNo:                "order_no",
	Status:                 "status",
	MemberDiscountFee:      "member_discount_fee",
	CustomDiscountFee:      "custom_discount_fee",
	ZeroFee:                "zero_fee",
	ProductAmount:          "product_amount",
	ProductOriginalAmount:  "product_original_amount",
	ServiceFee:             "service_fee",
	TaxFee:                 "tax_fee",
	Amount:                 "amount",
	OriginAmount:           "origin_amount",
	IsFree:                 "is_free",
	FreeReason:             "free_reason",
	MemberDiscountRate:     "member_discount_rate",
	MemberCardDiscountRate: "member_card_discount_rate",
	CustomDiscountRate:     "custom_discount_rate",
	CustomAmount:           "custom_amount",
	ZeroRule:               "zero_rule",
	ZeroCheckoutRule:       "zero_checkout_rule",
	PaymentAmount:          "payment_amount",
	ChangeAmount:           "change_amount",
	ZeroCheckoutFee:        "zero_checkout_fee",
	FinalPrice:             "final_price",
	PaymentCommissionFee:   "payment_commission_fee",
	GiftAmount:             "gift_amount",
	GiftPoints:             "gift_points",
	GiftPointsRate:         "gift_points_rate",
	MemberBalance:          "member_balance",
	CashierName:            "cashier_name",
	ConsumerUuid:           "consumer_uuid",
	CashierUuid:            "cashier_uuid",
	SaleBillUuid:           "sale_bill_uuid",
	FinishTime:             "finish_time",
	CreateTime:             "create_time",
	UpdateTime:             "update_time",
	DeleteTime:             "delete_time",
}

// NewSaleOrderDao creates and returns a new DAO object for table data access.
func NewSaleOrderDao(handlers ...gdb.ModelHandler) *SaleOrderDao {
	return &SaleOrderDao{
		group:    "default",
		table:    "ttpos_sale_order",
		columns:  saleOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderDao) Columns() SaleOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
