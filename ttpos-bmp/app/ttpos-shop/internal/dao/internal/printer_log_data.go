// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterLogDataDao is the data access object for the table ttpos_printer_log_data.
type PrinterLogDataDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  PrinterLogDataColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// PrinterLogDataColumns defines and stores column names for the table ttpos_printer_log_data.
type PrinterLogDataColumns struct {
	Id         string //
	Uuid       string // 唯一ID
	LogUuid    string // 打印日志UUID
	Data       string // 打印数据
	CreateTime string // 创建时间
	UpdateTime string // 更新时间
	DeleteTime string // 删除时间
}

// printerLogDataColumns holds the columns for the table ttpos_printer_log_data.
var printerLogDataColumns = PrinterLogDataColumns{
	Id:         "id",
	Uuid:       "uuid",
	LogUuid:    "log_uuid",
	Data:       "data",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewPrinterLogDataDao creates and returns a new DAO object for table data access.
func NewPrinterLogDataDao(handlers ...gdb.ModelHandler) *PrinterLogDataDao {
	return &PrinterLogDataDao{
		group:    "default",
		table:    "ttpos_printer_log_data",
		columns:  printerLogDataColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PrinterLogDataDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PrinterLogDataDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PrinterLogDataDao) Columns() PrinterLogDataColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PrinterLogDataDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PrinterLogDataDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PrinterLogDataDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
