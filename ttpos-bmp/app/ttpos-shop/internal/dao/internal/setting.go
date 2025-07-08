// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SettingDao is the data access object for the table ttpos_setting.
type SettingDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SettingColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SettingColumns defines and stores column names for the table ttpos_setting.
type SettingColumns struct {
	Key        string // 设置项标示
	Describe   string // 设置项描述
	Values     string // 设置内容(json格式)
	CreateTime string // 创建时间
	UpdateTime string // 更新时间
	DeleteTime string // 删除时间
}

// settingColumns holds the columns for the table ttpos_setting.
var settingColumns = SettingColumns{
	Key:        "key",
	Describe:   "describe",
	Values:     "values",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewSettingDao creates and returns a new DAO object for table data access.
func NewSettingDao(handlers ...gdb.ModelHandler) *SettingDao {
	return &SettingDao{
		group:    "default",
		table:    "ttpos_setting",
		columns:  settingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SettingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SettingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SettingDao) Columns() SettingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SettingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SettingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SettingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
