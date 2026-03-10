// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceivePaymentEntryDao is the data access object for the table erp_receive_payment_entry.
type ReceivePaymentEntryDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  ReceivePaymentEntryColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// ReceivePaymentEntryColumns defines and stores column names for the table erp_receive_payment_entry.
type ReceivePaymentEntryColumns struct {
	Id               string // ID
	SaleOrderUuid    string // TTPOS订单UUID
	ModeOfPayment    string // 支付方式
	PaymentEntryName string // ERP Payment Entry名称
	Docstatus        string // 文档状态
	PaidAmount       string // 支付金额
	SiteCode         string // ERP site code
	ReqBody          string // 请求文本
	RespBody         string // 响应文本
	RetryCount       string // 重试次数
	CreatedAt        string // 创建时间
	UpdatedAt        string // 更新时间
}

// receivePaymentEntryColumns holds the columns for the table erp_receive_payment_entry.
var receivePaymentEntryColumns = ReceivePaymentEntryColumns{
	Id:               "id",
	SaleOrderUuid:    "sale_order_uuid",
	ModeOfPayment:    "mode_of_payment",
	PaymentEntryName: "payment_entry_name",
	Docstatus:        "docstatus",
	PaidAmount:       "paid_amount",
	SiteCode:         "site_code",
	ReqBody:          "req_body",
	RespBody:         "resp_body",
	RetryCount:       "retry_count",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewReceivePaymentEntryDao creates and returns a new DAO object for table data access.
func NewReceivePaymentEntryDao(handlers ...gdb.ModelHandler) *ReceivePaymentEntryDao {
	return &ReceivePaymentEntryDao{
		group:    "default",
		table:    "erp_receive_payment_entry",
		columns:  receivePaymentEntryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceivePaymentEntryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceivePaymentEntryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceivePaymentEntryDao) Columns() ReceivePaymentEntryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceivePaymentEntryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceivePaymentEntryDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ReceivePaymentEntryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
