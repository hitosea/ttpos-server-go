// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceivePosInvoiceDao is the data access object for the table erp_receive_pos_invoice.
type ReceivePosInvoiceDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ReceivePosInvoiceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ReceivePosInvoiceColumns defines and stores column names for the table erp_receive_pos_invoice.
type ReceivePosInvoiceColumns struct {
	Id                  string // ID
	OrderNo             string // 销售订单号，来自ttpos
	OpenPosEntryName    string // POS开帐名称
	PostingDatetime     string // 过账日期时间
	Branch              string // 门店分支，可选
	Docstatus           string // 文档状态,参考 erpnext
	CreatedAt           string // 创建时间
	UpdatedAt           string // 更新时间
	ReqMessage          string // 请求数据,base64编码
	RespMessage         string // 响应数据,base64编码
	ProductsInvoiceName string // 商品销售发票
	MaterialInvoiceName string // 物品销售发票
	SiteCode            string // erp_site_code, 用来区分调那个租户
	ReqBody             string // 请求文本，如果能转换
	RespBody            string // 响应文本，如果能转换
}

// receivePosInvoiceColumns holds the columns for the table erp_receive_pos_invoice.
var receivePosInvoiceColumns = ReceivePosInvoiceColumns{
	Id:                  "id",
	OrderNo:             "order_no",
	OpenPosEntryName:    "open_pos_entry_name",
	PostingDatetime:     "posting_datetime",
	Branch:              "branch",
	Docstatus:           "docstatus",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	ReqMessage:          "req_message",
	RespMessage:         "resp_message",
	ProductsInvoiceName: "products_invoice_name",
	MaterialInvoiceName: "material_invoice_name",
	SiteCode:            "site_code",
	ReqBody:             "req_body",
	RespBody:            "resp_body",
}

// NewReceivePosInvoiceDao creates and returns a new DAO object for table data access.
func NewReceivePosInvoiceDao(handlers ...gdb.ModelHandler) *ReceivePosInvoiceDao {
	return &ReceivePosInvoiceDao{
		group:    "default",
		table:    "erp_receive_pos_invoice",
		columns:  receivePosInvoiceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceivePosInvoiceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceivePosInvoiceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceivePosInvoiceDao) Columns() ReceivePosInvoiceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceivePosInvoiceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceivePosInvoiceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReceivePosInvoiceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
