// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrderAmountDao is the data access object for the table ttpos_return_order_amount.
type ReturnOrderAmountDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ReturnOrderAmountColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ReturnOrderAmountColumns defines and stores column names for the table ttpos_return_order_amount.
type ReturnOrderAmountColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 退货金额唯一标识符
	ReturnOrderUuid       string // 关联退货单ID
	PaymentMethodUuid     string // 关联支付方式ID
	PaymentOrderUuid      string // 关联支付单ID,用于判断支付单的钱还有多少未退
	Amount                string // 退款金额
	RefundStatus          string // 退款状态 0-退款中 1-退款成功 2-退款失败
	LlReturnOrderId       string // 连连退款订单ID, 用来重新发起退款
	MerchantRefundOrderNo string // 商户退款单号
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// returnOrderAmountColumns holds the columns for the table ttpos_return_order_amount.
var returnOrderAmountColumns = ReturnOrderAmountColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	ReturnOrderUuid:       "return_order_uuid",
	PaymentMethodUuid:     "payment_method_uuid",
	PaymentOrderUuid:      "payment_order_uuid",
	Amount:                "amount",
	RefundStatus:          "refund_status",
	LlReturnOrderId:       "ll_return_order_id",
	MerchantRefundOrderNo: "merchant_refund_order_no",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewReturnOrderAmountDao creates and returns a new DAO object for table data access.
func NewReturnOrderAmountDao(handlers ...gdb.ModelHandler) *ReturnOrderAmountDao {
	return &ReturnOrderAmountDao{
		group:    "default",
		table:    "ttpos_return_order_amount",
		columns:  returnOrderAmountColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReturnOrderAmountDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReturnOrderAmountDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReturnOrderAmountDao) Columns() ReturnOrderAmountColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReturnOrderAmountDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReturnOrderAmountDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReturnOrderAmountDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
