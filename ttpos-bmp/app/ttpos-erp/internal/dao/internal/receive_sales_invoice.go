// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveSalesInvoiceDao is the data access object for the table erp_receive_sales_invoice.
type ReceiveSalesInvoiceDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  ReceiveSalesInvoiceColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// ReceiveSalesInvoiceColumns defines and stores column names for the table erp_receive_sales_invoice.
type ReceiveSalesInvoiceColumns struct {
	Id                string // ID
	OrderNo           string // TTPOS订单号
	SaleOrderUuid     string // TTPOS订单UUID（幂等键）
	PosProfile        string // POS Profile名称
	PostingDatetime   string // 过账时间戳
	Docstatus         string // 文档状态: 0=Draft 1=Submitted 2=Cancelled
	SalesInvoiceName  string // ERP Sales Invoice名称
	PaymentEntryNames string // ERP Payment Entry名称(JSON)
	SiteCode          string // ERP site code
	ReqMessage        string // 请求数据(base64)
	RespMessage       string // 响应数据(base64)
	ReqBody           string // 请求文本
	RespBody          string // 响应文本
	RetryCount        string // 重试次数
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// receiveSalesInvoiceColumns holds the columns for the table erp_receive_sales_invoice.
var receiveSalesInvoiceColumns = ReceiveSalesInvoiceColumns{
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
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewReceiveSalesInvoiceDao creates and returns a new DAO object for table data access.
func NewReceiveSalesInvoiceDao(handlers ...gdb.ModelHandler) *ReceiveSalesInvoiceDao {
	return &ReceiveSalesInvoiceDao{
		group:    "default",
		table:    "erp_receive_sales_invoice",
		columns:  receiveSalesInvoiceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceiveSalesInvoiceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceiveSalesInvoiceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceiveSalesInvoiceDao) Columns() ReceiveSalesInvoiceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceiveSalesInvoiceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceiveSalesInvoiceDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ReceiveSalesInvoiceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
