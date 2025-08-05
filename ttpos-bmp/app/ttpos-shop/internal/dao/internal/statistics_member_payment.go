// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsMemberPaymentDao is the data access object for the table ttpos_statistics_member_payment.
type StatisticsMemberPaymentDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  StatisticsMemberPaymentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// StatisticsMemberPaymentColumns defines and stores column names for the table ttpos_statistics_member_payment.
type StatisticsMemberPaymentColumns struct {
	Id                      string //
	Uuid                    string // uuid
	MemberRechargeOrderUuid string // 会员充值订单uuid
	DutyNo                  string // 当班编号
	PaymentMethodUuid       string // 支付方式uuid
	PaymentAmount           string // 支付金额
	RefundAmount            string // 退款金额
	CompleteTime            string // 完成时间
	CreateTime              string // 创建时间
	UpdateTime              string // 更新时间
	DeleteTime              string // 删除时间
}

// statisticsMemberPaymentColumns holds the columns for the table ttpos_statistics_member_payment.
var statisticsMemberPaymentColumns = StatisticsMemberPaymentColumns{
	Id:                      "id",
	Uuid:                    "uuid",
	MemberRechargeOrderUuid: "member_recharge_order_uuid",
	DutyNo:                  "duty_no",
	PaymentMethodUuid:       "payment_method_uuid",
	PaymentAmount:           "payment_amount",
	RefundAmount:            "refund_amount",
	CompleteTime:            "complete_time",
	CreateTime:              "create_time",
	UpdateTime:              "update_time",
	DeleteTime:              "delete_time",
}

// NewStatisticsMemberPaymentDao creates and returns a new DAO object for table data access.
func NewStatisticsMemberPaymentDao(handlers ...gdb.ModelHandler) *StatisticsMemberPaymentDao {
	return &StatisticsMemberPaymentDao{
		group:    "default",
		table:    "ttpos_statistics_member_payment",
		columns:  statisticsMemberPaymentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StatisticsMemberPaymentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StatisticsMemberPaymentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StatisticsMemberPaymentDao) Columns() StatisticsMemberPaymentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StatisticsMemberPaymentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StatisticsMemberPaymentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StatisticsMemberPaymentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
