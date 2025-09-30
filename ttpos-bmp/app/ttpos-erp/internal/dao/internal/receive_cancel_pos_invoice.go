// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveCancelPosInvoiceDao is the data access object for the table erp_receive_cancel_pos_invoice.
type ReceiveCancelPosInvoiceDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  ReceiveCancelPosInvoiceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// ReceiveCancelPosInvoiceColumns defines and stores column names for the table erp_receive_cancel_pos_invoice.
type ReceiveCancelPosInvoiceColumns struct {
	Id               string // ID
	OrderNo          string // 退款订单号，来自ttpos
	OpenPosEntryName string // POS开帐名称
	Docstatus        string // 文档状态,参考 erpnext
	CreatedAt        string // 创建时间
	UpdatedAt        string // 更新时间
	ReqMessage       string // 请求数据,base64编码
	RespMessage      string // 响应数据,base64编码
	SiteCode         string // erp_site_code, 用来区分调那个租户
	ReqBody          string // 请求文本，如果能转换
	RespBody         string // 响应文本，如果能转换
}

// receiveCancelPosInvoiceColumns holds the columns for the table erp_receive_cancel_pos_invoice.
var receiveCancelPosInvoiceColumns = ReceiveCancelPosInvoiceColumns{
	Id:               "id",
	OrderNo:          "order_no",
	OpenPosEntryName: "open_pos_entry_name",
	Docstatus:        "docstatus",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	ReqMessage:       "req_message",
	RespMessage:      "resp_message",
	SiteCode:         "site_code",
	ReqBody:          "req_body",
	RespBody:         "resp_body",
}

// NewReceiveCancelPosInvoiceDao creates and returns a new DAO object for table data access.
func NewReceiveCancelPosInvoiceDao(handlers ...gdb.ModelHandler) *ReceiveCancelPosInvoiceDao {
	return &ReceiveCancelPosInvoiceDao{
		group:    "default",
		table:    "erp_receive_cancel_pos_invoice",
		columns:  receiveCancelPosInvoiceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceiveCancelPosInvoiceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceiveCancelPosInvoiceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceiveCancelPosInvoiceDao) Columns() ReceiveCancelPosInvoiceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceiveCancelPosInvoiceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceiveCancelPosInvoiceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReceiveCancelPosInvoiceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
