// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyDao is the data access object for the table ttpos_company.
type CompanyDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CompanyColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CompanyColumns defines and stores column names for the table ttpos_company.
type CompanyColumns struct {
	Id            string // 自增ID
	Uuid          string // 集团ID
	Name          string // 集团名称
	Logo          string // logo
	ExpireTime    string // 过期时间;not null
	AuthDay       string // 授权时间(天) 0为永不过期
	Status        string // 状态 1-启用 0-禁用;not null
	AuthStartTime string // 授权开始时间(时间戳)
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// companyColumns holds the columns for the table ttpos_company.
var companyColumns = CompanyColumns{
	Id:            "id",
	Uuid:          "uuid",
	Name:          "name",
	Logo:          "logo",
	ExpireTime:    "expire_time",
	AuthDay:       "auth_day",
	Status:        "status",
	AuthStartTime: "auth_start_time",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewCompanyDao creates and returns a new DAO object for table data access.
func NewCompanyDao(handlers ...gdb.ModelHandler) *CompanyDao {
	return &CompanyDao{
		group:    "default",
		table:    "ttpos_company",
		columns:  companyColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CompanyDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CompanyDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CompanyDao) Columns() CompanyColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CompanyDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CompanyDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CompanyDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
