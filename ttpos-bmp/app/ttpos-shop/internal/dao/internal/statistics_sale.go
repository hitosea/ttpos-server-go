// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsSaleDao is the data access object for the table ttpos_statistics_sale.
type StatisticsSaleDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  StatisticsSaleColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// StatisticsSaleColumns defines and stores column names for the table ttpos_statistics_sale.
type StatisticsSaleColumns struct {
	Id                   string // 自增ID
	Uuid                 string // UUID
	SaleBillUuid         string // 销售单UUID
	SaleOrderUuid        string // 销售订单UUID
	DutyNo               string // 当班编号
	DeskUuid             string // 桌台UUID
	MealNum              string // 用餐人数
	ProductPrice         string // 商品原价: 不含税
	ProductSalePrice     string // 商品销售价
	ProductNum           string // 商品数量
	ProductTax           string // 商品税
	ServiceFee           string // 服务费
	ServiceTax           string // 服务税
	Discount             string // 优惠折扣
	DiscountMember       string // 会员折扣
	GiftAmount           string // 赠菜金额
	GiftNum              string // 赠菜数量
	FreeAmount           string // 免单金额
	FreeNum              string // 免单数量
	PaymentAmount        string // 支付金额
	PaymentFee           string // 支付手续费
	PaymentBalance       string // 支付余额
	RefundAmount         string // 退款金额
	RefundTax            string // 退款税额
	RefundServiceFee     string // 退款服务费
	RefundDiscount       string // 退款优惠折扣
	RefundDiscountMember string // 退款会员折扣
	RefundFee            string // 退款支付手续费
	CompleteTime         string // 完成时间
	RefundTime           string // 退款时间
	CreateTime           string // 创建时间
	UpdateTime           string // 更新时间
	DeleteTime           string // 删除时间
}

// statisticsSaleColumns holds the columns for the table ttpos_statistics_sale.
var statisticsSaleColumns = StatisticsSaleColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	SaleBillUuid:         "sale_bill_uuid",
	SaleOrderUuid:        "sale_order_uuid",
	DutyNo:               "duty_no",
	DeskUuid:             "desk_uuid",
	MealNum:              "meal_num",
	ProductPrice:         "product_price",
	ProductSalePrice:     "product_sale_price",
	ProductNum:           "product_num",
	ProductTax:           "product_tax",
	ServiceFee:           "service_fee",
	ServiceTax:           "service_tax",
	Discount:             "discount",
	DiscountMember:       "discount_member",
	GiftAmount:           "gift_amount",
	GiftNum:              "gift_num",
	FreeAmount:           "free_amount",
	FreeNum:              "free_num",
	PaymentAmount:        "payment_amount",
	PaymentFee:           "payment_fee",
	PaymentBalance:       "payment_balance",
	RefundAmount:         "refund_amount",
	RefundTax:            "refund_tax",
	RefundServiceFee:     "refund_service_fee",
	RefundDiscount:       "refund_discount",
	RefundDiscountMember: "refund_discount_member",
	RefundFee:            "refund_fee",
	CompleteTime:         "complete_time",
	RefundTime:           "refund_time",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewStatisticsSaleDao creates and returns a new DAO object for table data access.
func NewStatisticsSaleDao(handlers ...gdb.ModelHandler) *StatisticsSaleDao {
	return &StatisticsSaleDao{
		group:    "default",
		table:    "ttpos_statistics_sale",
		columns:  statisticsSaleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StatisticsSaleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StatisticsSaleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StatisticsSaleDao) Columns() StatisticsSaleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StatisticsSaleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StatisticsSaleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StatisticsSaleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
