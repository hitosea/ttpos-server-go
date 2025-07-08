// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinterItemDao is the data access object for the table ttpos_product_printer_item.
type ProductPrinterItemDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  ProductPrinterItemColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// ProductPrinterItemColumns defines and stores column names for the table ttpos_product_printer_item.
type ProductPrinterItemColumns struct {
	Id                 string // 自增ID
	Uuid               string // 商品打印(档口)打印机ID
	ProductPrinterUuid string // 商品打印(档口)ID
	PrinterUuid        string // 打印机ID
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// productPrinterItemColumns holds the columns for the table ttpos_product_printer_item.
var productPrinterItemColumns = ProductPrinterItemColumns{
	Id:                 "id",
	Uuid:               "uuid",
	ProductPrinterUuid: "product_printer_uuid",
	PrinterUuid:        "printer_uuid",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewProductPrinterItemDao creates and returns a new DAO object for table data access.
func NewProductPrinterItemDao(handlers ...gdb.ModelHandler) *ProductPrinterItemDao {
	return &ProductPrinterItemDao{
		group:    "default",
		table:    "ttpos_product_printer_item",
		columns:  productPrinterItemColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductPrinterItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductPrinterItemDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductPrinterItemDao) Columns() ProductPrinterItemColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductPrinterItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductPrinterItemDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductPrinterItemDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
