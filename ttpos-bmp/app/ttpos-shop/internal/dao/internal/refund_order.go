// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RefundOrderDao is the data access object for the table ttpos_refund_order.
type RefundOrderDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  RefundOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// RefundOrderColumns defines and stores column names for the table ttpos_refund_order.
type RefundOrderColumns struct {
	Id               string // 自增ID
	Uuid             string // 退款单唯一标识符
	SaleOrderUuid    string // 销售订单ID
	SaleOrderNo      string // 销售订单号
	PaymentOrderUuid string // 支付单ID
	RefundType       string // 退款类型,1-反结账,2-取消付款
	Amount           string // 退款金额
	Reason           string // 退款原因
	Status           string // 退款状态
	CreateTime       string // 创建时间(时间戳)
	UpdateTime       string // 更新时间(时间戳)
	DeleteTime       string // 删除时间(时间戳)
}

// refundOrderColumns holds the columns for the table ttpos_refund_order.
var refundOrderColumns = RefundOrderColumns{
	Id:               "id",
	Uuid:             "uuid",
	SaleOrderUuid:    "sale_order_uuid",
	SaleOrderNo:      "sale_order_no",
	PaymentOrderUuid: "payment_order_uuid",
	RefundType:       "refund_type",
	Amount:           "amount",
	Reason:           "reason",
	Status:           "status",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewRefundOrderDao creates and returns a new DAO object for table data access.
func NewRefundOrderDao(handlers ...gdb.ModelHandler) *RefundOrderDao {
	return &RefundOrderDao{
		group:    "default",
		table:    "ttpos_refund_order",
		columns:  refundOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RefundOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RefundOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RefundOrderDao) Columns() RefundOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RefundOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RefundOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RefundOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
