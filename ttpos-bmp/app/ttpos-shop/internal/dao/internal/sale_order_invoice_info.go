// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderInvoiceInfoDao is the data access object for the table ttpos_sale_order_invoice_info.
type SaleOrderInvoiceInfoDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  SaleOrderInvoiceInfoColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// SaleOrderInvoiceInfoColumns defines and stores column names for the table ttpos_sale_order_invoice_info.
type SaleOrderInvoiceInfoColumns struct {
	Id               string // 自增ID
	Uuid             string // 唯一ID
	SaleOrderUuid    string // 销售订单ID
	CompanyName      string // 公司名称
	CompanyAddr      string // 公司地址
	CompanyTaxNumber string // 公司税号
	CompanyPhone     string // 公司电话
	PrintNum         string // 打印次数
	CreateTime       string // 创建时间
	UpdateTime       string // 更新时间
	DeleteTime       string // 删除时间
}

// saleOrderInvoiceInfoColumns holds the columns for the table ttpos_sale_order_invoice_info.
var saleOrderInvoiceInfoColumns = SaleOrderInvoiceInfoColumns{
	Id:               "id",
	Uuid:             "uuid",
	SaleOrderUuid:    "sale_order_uuid",
	CompanyName:      "company_name",
	CompanyAddr:      "company_addr",
	CompanyTaxNumber: "company_tax_number",
	CompanyPhone:     "company_phone",
	PrintNum:         "print_num",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewSaleOrderInvoiceInfoDao creates and returns a new DAO object for table data access.
func NewSaleOrderInvoiceInfoDao(handlers ...gdb.ModelHandler) *SaleOrderInvoiceInfoDao {
	return &SaleOrderInvoiceInfoDao{
		group:    "default",
		table:    "ttpos_sale_order_invoice_info",
		columns:  saleOrderInvoiceInfoColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderInvoiceInfoDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderInvoiceInfoDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderInvoiceInfoDao) Columns() SaleOrderInvoiceInfoColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderInvoiceInfoDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderInvoiceInfoDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderInvoiceInfoDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
