// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUserOptLogDao is the data access object for the table ttpos_admin_user_opt_log.
type AdminUserOptLogDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  AdminUserOptLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// AdminUserOptLogColumns defines and stores column names for the table ttpos_admin_user_opt_log.
type AdminUserOptLogColumns struct {
	Id          string // 自增ID
	AdminUserId string // 用户ID
	Title       string // 标题
	Url         string // 访问url
	RequestType string // 请求类型
	Browser     string // 浏览器
	Agent       string // 浏览器信息
	Content     string // 操作内容
	Ip          string // 登录ip
	CreateTime  string // 创建时间（时间戳）
	UpdateTime  string // 更新时间（时间戳）
	DeleteTime  string // 删除时间（时间戳）
}

// adminUserOptLogColumns holds the columns for the table ttpos_admin_user_opt_log.
var adminUserOptLogColumns = AdminUserOptLogColumns{
	Id:          "id",
	AdminUserId: "admin_user_id",
	Title:       "title",
	Url:         "url",
	RequestType: "request_type",
	Browser:     "browser",
	Agent:       "agent",
	Content:     "content",
	Ip:          "ip",
	CreateTime:  "create_time",
	UpdateTime:  "update_time",
	DeleteTime:  "delete_time",
}

// NewAdminUserOptLogDao creates and returns a new DAO object for table data access.
func NewAdminUserOptLogDao(handlers ...gdb.ModelHandler) *AdminUserOptLogDao {
	return &AdminUserOptLogDao{
		group:    "default",
		table:    "ttpos_admin_user_opt_log",
		columns:  adminUserOptLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AdminUserOptLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AdminUserOptLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AdminUserOptLogDao) Columns() AdminUserOptLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AdminUserOptLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AdminUserOptLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AdminUserOptLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
