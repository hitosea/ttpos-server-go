// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterTypeDao is the data access object for the table ttpos_printer_type.
type PrinterTypeDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PrinterTypeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PrinterTypeColumns defines and stores column names for the table ttpos_printer_type.
type PrinterTypeColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 打印机类型ID
	Name                  string // 打印机类型名称
	MultiLanguageNameUuid string // 多语言名称ID
	Key                   string // 打印机类型key
	ConfigJson            string // 打印机类型json配置,描述需要填写的字段
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// printerTypeColumns holds the columns for the table ttpos_printer_type.
var printerTypeColumns = PrinterTypeColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	Key:                   "key",
	ConfigJson:            "config_json",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewPrinterTypeDao creates and returns a new DAO object for table data access.
func NewPrinterTypeDao(handlers ...gdb.ModelHandler) *PrinterTypeDao {
	return &PrinterTypeDao{
		group:    "default",
		table:    "ttpos_printer_type",
		columns:  printerTypeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PrinterTypeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PrinterTypeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PrinterTypeDao) Columns() PrinterTypeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PrinterTypeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PrinterTypeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PrinterTypeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
