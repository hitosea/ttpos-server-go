// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StaffRoleDao is the data access object for the table ttpos_staff_role.
type StaffRoleDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  StaffRoleColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// StaffRoleColumns defines and stores column names for the table ttpos_staff_role.
type StaffRoleColumns struct {
	Id         string // 自增ID
	Uuid       string // 员工角色关系ID
	StaffUuid  string // 超管用户ID
	RoleUuid   string // 角色ID
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// staffRoleColumns holds the columns for the table ttpos_staff_role.
var staffRoleColumns = StaffRoleColumns{
	Id:         "id",
	Uuid:       "uuid",
	StaffUuid:  "staff_uuid",
	RoleUuid:   "role_uuid",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewStaffRoleDao creates and returns a new DAO object for table data access.
func NewStaffRoleDao(handlers ...gdb.ModelHandler) *StaffRoleDao {
	return &StaffRoleDao{
		group:    "default",
		table:    "ttpos_staff_role",
		columns:  staffRoleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StaffRoleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StaffRoleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StaffRoleDao) Columns() StaffRoleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StaffRoleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StaffRoleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StaffRoleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
