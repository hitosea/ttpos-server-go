// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrderDao is the data access object for the table ttpos_return_order.
type ReturnOrderDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ReturnOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ReturnOrderColumns defines and stores column names for the table ttpos_return_order.
type ReturnOrderColumns struct {
	Id                  string // 自增ID
	Uuid                string // 退货单唯一标识符
	RelatedOrderType    string // 关联订单类型：0-销售订单；1-充值订单
	RelatedOrderUuid    string // 关联订单ID
	RelatedOrderNo      string // 关联订单号
	LlReturnOrderId     string // 连连退款订单ID, 用来重新发起退款
	IsReverseSettlement string // 是否反结账：0-否；1-是
	ReturnType          string // 退货类型,1-整单退货,2-部分退货
	RefundAmount        string // 退款金额,包括税额
	Unit                string // 货币单位
	RefundTaxAmount     string // 退款税额
	RefundReason        string // 退款原因
	BankCode            string // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo           string // 账号 - 当存在QR PromptPay的时候需要传
	AccountName         string // 账户名称 - 当存在QR PromptPay的时候需要传
	CreateTime          string // 创建时间(时间戳)
	UpdateTime          string // 更新时间(时间戳)
	DeleteTime          string // 删除时间(时间戳)
}

// returnOrderColumns holds the columns for the table ttpos_return_order.
var returnOrderColumns = ReturnOrderColumns{
	Id:                  "id",
	Uuid:                "uuid",
	RelatedOrderType:    "related_order_type",
	RelatedOrderUuid:    "related_order_uuid",
	RelatedOrderNo:      "related_order_no",
	LlReturnOrderId:     "ll_return_order_id",
	IsReverseSettlement: "is_reverse_settlement",
	ReturnType:          "return_type",
	RefundAmount:        "refund_amount",
	Unit:                "unit",
	RefundTaxAmount:     "refund_tax_amount",
	RefundReason:        "refund_reason",
	BankCode:            "bank_code",
	AccountNo:           "account_no",
	AccountName:         "account_name",
	CreateTime:          "create_time",
	UpdateTime:          "update_time",
	DeleteTime:          "delete_time",
}

// NewReturnOrderDao creates and returns a new DAO object for table data access.
func NewReturnOrderDao(handlers ...gdb.ModelHandler) *ReturnOrderDao {
	return &ReturnOrderDao{
		group:    "default",
		table:    "ttpos_return_order",
		columns:  returnOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReturnOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReturnOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReturnOrderDao) Columns() ReturnOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReturnOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReturnOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReturnOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
