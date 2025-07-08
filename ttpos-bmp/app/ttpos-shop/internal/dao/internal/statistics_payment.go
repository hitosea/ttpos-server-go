// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsPaymentDao is the data access object for the table ttpos_statistics_payment.
type StatisticsPaymentDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  StatisticsPaymentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// StatisticsPaymentColumns defines and stores column names for the table ttpos_statistics_payment.
type StatisticsPaymentColumns struct {
	Id                string // 自增ID
	Uuid              string // UUID
	SaleBillUuid      string // 销售单UUID
	SaleOrderUuid     string // 销售订单UUID
	DutyNo            string // 当班编号
	DeskUuid          string // 桌台UUID
	PaymentMethodUuid string // 支付方式UUID
	PaymentAmount     string // 支付金额
	RefundAmount      string // 退款金额
	CompleteTime      string // 完成时间
	CreateTime        string // 创建时间
	UpdateTime        string // 更新时间
	DeleteTime        string // 删除时间
}

// statisticsPaymentColumns holds the columns for the table ttpos_statistics_payment.
var statisticsPaymentColumns = StatisticsPaymentColumns{
	Id:                "id",
	Uuid:              "uuid",
	SaleBillUuid:      "sale_bill_uuid",
	SaleOrderUuid:     "sale_order_uuid",
	DutyNo:            "duty_no",
	DeskUuid:          "desk_uuid",
	PaymentMethodUuid: "payment_method_uuid",
	PaymentAmount:     "payment_amount",
	RefundAmount:      "refund_amount",
	CompleteTime:      "complete_time",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewStatisticsPaymentDao creates and returns a new DAO object for table data access.
func NewStatisticsPaymentDao(handlers ...gdb.ModelHandler) *StatisticsPaymentDao {
	return &StatisticsPaymentDao{
		group:    "default",
		table:    "ttpos_statistics_payment",
		columns:  statisticsPaymentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StatisticsPaymentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StatisticsPaymentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StatisticsPaymentDao) Columns() StatisticsPaymentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StatisticsPaymentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StatisticsPaymentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StatisticsPaymentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
