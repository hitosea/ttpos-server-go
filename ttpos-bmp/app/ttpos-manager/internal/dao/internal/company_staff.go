// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyStaffDao is the data access object for the table ttpos_company_staff.
type CompanyStaffDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  CompanyStaffColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// CompanyStaffColumns defines and stores column names for the table ttpos_company_staff.
type CompanyStaffColumns struct {
	Id          string // 自增ID
	Uuid        string // 员工ID
	CompanyUuid string // 集团ID
	Username    string // 员工账号
	Phone       string // 员工手机号
	IsSuper     string // 是否超级管理员
	CreateTime  string // 创建时间（时间戳）
	UpdateTime  string // 更新时间（时间戳）
	DeleteTime  string // 删除时间（时间戳）
}

// companyStaffColumns holds the columns for the table ttpos_company_staff.
var companyStaffColumns = CompanyStaffColumns{
	Id:          "id",
	Uuid:        "uuid",
	CompanyUuid: "company_uuid",
	Username:    "username",
	Phone:       "phone",
	IsSuper:     "is_super",
	CreateTime:  "create_time",
	UpdateTime:  "update_time",
	DeleteTime:  "delete_time",
}

// NewCompanyStaffDao creates and returns a new DAO object for table data access.
func NewCompanyStaffDao(handlers ...gdb.ModelHandler) *CompanyStaffDao {
	return &CompanyStaffDao{
		group:    "default",
		table:    "ttpos_company_staff",
		columns:  companyStaffColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CompanyStaffDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CompanyStaffDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CompanyStaffDao) Columns() CompanyStaffColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CompanyStaffDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CompanyStaffDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CompanyStaffDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
