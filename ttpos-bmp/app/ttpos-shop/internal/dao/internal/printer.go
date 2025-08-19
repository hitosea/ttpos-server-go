// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterDao is the data access object for the table ttpos_printer.
type PrinterDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PrinterColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PrinterColumns defines and stores column names for the table ttpos_printer.
type PrinterColumns struct {
	Id              string // 自增ID
	Uuid            string // 打印机ID
	Name            string // 打印机名称
	PrinterTypeUuid string // 打印机类型ID
	ConfigJson      string // 打印机json配置
	Copies          string // 打印份数
	Sort            string // 排序
	CreateTime      string // 创建时间(时间戳)
	UpdateTime      string // 更新时间(时间戳)
	DeleteTime      string // 删除时间(时间戳)
}

// printerColumns holds the columns for the table ttpos_printer.
var printerColumns = PrinterColumns{
	Id:              "id",
	Uuid:            "uuid",
	Name:            "name",
	PrinterTypeUuid: "printer_type_uuid",
	ConfigJson:      "config_json",
	Copies:          "copies",
	Sort:            "sort",
	CreateTime:      "create_time",
	UpdateTime:      "update_time",
	DeleteTime:      "delete_time",
}

// NewPrinterDao creates and returns a new DAO object for table data access.
func NewPrinterDao(handlers ...gdb.ModelHandler) *PrinterDao {
	return &PrinterDao{
		group:    "default",
		table:    "ttpos_printer",
		columns:  printerColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PrinterDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PrinterDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PrinterDao) Columns() PrinterColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PrinterDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PrinterDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PrinterDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
