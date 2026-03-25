// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveReturnSalesInvoiceDao is the data access object for the table erp_receive_return_sales_invoice.
type ReceiveReturnSalesInvoiceDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  ReceiveReturnSalesInvoiceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// ReceiveReturnSalesInvoiceColumns defines and stores column names for the table erp_receive_return_sales_invoice.
type ReceiveReturnSalesInvoiceColumns struct {
	Id                string // ID
	OrderNo           string // TTPOS订单号
	SaleOrderUuid     string // TTPOS订单UUID
	PosProfile        string // POS Profile名称
	PostingDatetime   string // 过账时间戳
	Docstatus         string // 文档状态: 0=Draft 1=Submitted 2=Cancelled
	SalesInvoiceName  string // Credit Note名称
	PaymentEntryNames string // Refund Payment Entry名称(JSON)
	SiteCode          string // ERP site code
	ReqMessage        string // 请求数据(base64)
	RespMessage       string // 响应数据(base64)
	ReqBody           string // 请求文本
	RespBody          string // 响应文本
	RetryCount        string // 重试次数
	MqMsgId           string // 最后一次MQ消息ID
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// receiveReturnSalesInvoiceColumns holds the columns for the table erp_receive_return_sales_invoice.
var receiveReturnSalesInvoiceColumns = ReceiveReturnSalesInvoiceColumns{
	Id:                "id",
	OrderNo:           "order_no",
	SaleOrderUuid:     "sale_order_uuid",
	PosProfile:        "pos_profile",
	PostingDatetime:   "posting_datetime",
	Docstatus:         "docstatus",
	SalesInvoiceName:  "sales_invoice_name",
	PaymentEntryNames: "payment_entry_names",
	SiteCode:          "site_code",
	ReqMessage:        "req_message",
	RespMessage:       "resp_message",
	ReqBody:           "req_body",
	RespBody:          "resp_body",
	RetryCount:        "retry_count",
	MqMsgId:           "mq_msg_id",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewReceiveReturnSalesInvoiceDao creates and returns a new DAO object for table data access.
func NewReceiveReturnSalesInvoiceDao(handlers ...gdb.ModelHandler) *ReceiveReturnSalesInvoiceDao {
	return &ReceiveReturnSalesInvoiceDao{
		group:    "default",
		table:    "erp_receive_return_sales_invoice",
		columns:  receiveReturnSalesInvoiceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceiveReturnSalesInvoiceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceiveReturnSalesInvoiceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceiveReturnSalesInvoiceDao) Columns() ReceiveReturnSalesInvoiceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceiveReturnSalesInvoiceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceiveReturnSalesInvoiceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReceiveReturnSalesInvoiceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
