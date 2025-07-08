// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MigrationsDao is the data access object for the table ttpos_migrations.
type MigrationsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MigrationsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MigrationsColumns defines and stores column names for the table ttpos_migrations.
type MigrationsColumns struct {
	Version       string //
	MigrationName string //
	StartTime     string //
	EndTime       string //
	Breakpoint    string //
}

// migrationsColumns holds the columns for the table ttpos_migrations.
var migrationsColumns = MigrationsColumns{
	Version:       "version",
	MigrationName: "migration_name",
	StartTime:     "start_time",
	EndTime:       "end_time",
	Breakpoint:    "breakpoint",
}

// NewMigrationsDao creates and returns a new DAO object for table data access.
func NewMigrationsDao(handlers ...gdb.ModelHandler) *MigrationsDao {
	return &MigrationsDao{
		group:    "default",
		table:    "ttpos_migrations",
		columns:  migrationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MigrationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MigrationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MigrationsDao) Columns() MigrationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MigrationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MigrationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MigrationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
