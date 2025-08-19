// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StaffDao is the data access object for the table ttpos_staff.
type StaffDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  StaffColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// StaffColumns defines and stores column names for the table ttpos_staff.
type StaffColumns struct {
	Id                  string // 自增ID
	Uuid                string // 员工ID
	CompanyUuid         string // 集团ID
	Username            string // 用户名
	Password            string // 登录密码
	Phone               string // 手机号
	PasswordChangeCount string // 修改密码次数
	PasswordChangeTime  string // 修改密码时间
	RealName            string // 姓名
	IsSuper             string // 是否为超级管理员0不是,1是
	UserType            string // 账号类型0总台1门店
	IsDisable           string // 是否禁用1禁用,0未禁用
	BindKey             string // 绑定的设备key
	CashierOnline       string // 收银员当班 0-不在线 1-在线
	CashierLoginTime    string // 收银员当班登录时间
	DutyNo              string // 当班编号
	CreateTime          string // 创建时间(时间戳)
	UpdateTime          string // 更新时间(时间戳)
	DeleteTime          string // 删除时间(时间戳)
}

// staffColumns holds the columns for the table ttpos_staff.
var staffColumns = StaffColumns{
	Id:                  "id",
	Uuid:                "uuid",
	CompanyUuid:         "company_uuid",
	Username:            "username",
	Password:            "password",
	Phone:               "phone",
	PasswordChangeCount: "password_change_count",
	PasswordChangeTime:  "password_change_time",
	RealName:            "real_name",
	IsSuper:             "is_super",
	UserType:            "user_type",
	IsDisable:           "is_disable",
	BindKey:             "bind_key",
	CashierOnline:       "cashier_online",
	CashierLoginTime:    "cashier_login_time",
	DutyNo:              "duty_no",
	CreateTime:          "create_time",
	UpdateTime:          "update_time",
	DeleteTime:          "delete_time",
}

// NewStaffDao creates and returns a new DAO object for table data access.
func NewStaffDao(handlers ...gdb.ModelHandler) *StaffDao {
	return &StaffDao{
		group:    "default",
		table:    "ttpos_staff",
		columns:  staffColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StaffDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StaffDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StaffDao) Columns() StaffColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StaffDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StaffDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StaffDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
