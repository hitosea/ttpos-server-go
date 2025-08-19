// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPrinterRegionDao is the data access object for the table ttpos_product_printer_region.
type ProductPrinterRegionDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  ProductPrinterRegionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// ProductPrinterRegionColumns defines and stores column names for the table ttpos_product_printer_region.
type ProductPrinterRegionColumns struct {
	Id                 string // 自增ID
	Uuid               string // 商品打印机区域ID
	ProductPrinterUuid string // 商品打印机ID
	DeskRegionUuid     string // 桌台区域ID
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// productPrinterRegionColumns holds the columns for the table ttpos_product_printer_region.
var productPrinterRegionColumns = ProductPrinterRegionColumns{
	Id:                 "id",
	Uuid:               "uuid",
	ProductPrinterUuid: "product_printer_uuid",
	DeskRegionUuid:     "desk_region_uuid",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewProductPrinterRegionDao creates and returns a new DAO object for table data access.
func NewProductPrinterRegionDao(handlers ...gdb.ModelHandler) *ProductPrinterRegionDao {
	return &ProductPrinterRegionDao{
		group:    "default",
		table:    "ttpos_product_printer_region",
		columns:  productPrinterRegionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductPrinterRegionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductPrinterRegionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductPrinterRegionDao) Columns() ProductPrinterRegionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductPrinterRegionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductPrinterRegionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductPrinterRegionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
