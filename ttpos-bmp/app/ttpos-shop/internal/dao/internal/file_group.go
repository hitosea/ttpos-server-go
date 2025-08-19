// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// FileGroupDao is the data access object for the table ttpos_file_group.
type FileGroupDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  FileGroupColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// FileGroupColumns defines and stores column names for the table ttpos_file_group.
type FileGroupColumns struct {
	Id         string // 自增ID
	Uuid       string // 文件组ID
	GroupType  string // 文件类型
	GroupName  string // 分类名称
	Sort       string // 分类排序(数字越小越靠前)
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// fileGroupColumns holds the columns for the table ttpos_file_group.
var fileGroupColumns = FileGroupColumns{
	Id:         "id",
	Uuid:       "uuid",
	GroupType:  "group_type",
	GroupName:  "group_name",
	Sort:       "sort",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewFileGroupDao creates and returns a new DAO object for table data access.
func NewFileGroupDao(handlers ...gdb.ModelHandler) *FileGroupDao {
	return &FileGroupDao{
		group:    "default",
		table:    "ttpos_file_group",
		columns:  fileGroupColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *FileGroupDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *FileGroupDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *FileGroupDao) Columns() FileGroupColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *FileGroupDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *FileGroupDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *FileGroupDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
