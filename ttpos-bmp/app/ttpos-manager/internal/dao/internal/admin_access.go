// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AdminAccessDao is the data access object for the table ttpos_admin_access.
type AdminAccessDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AdminAccessColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AdminAccessColumns defines and stores column names for the table ttpos_admin_access.
type AdminAccessColumns struct {
	Id           string // 自增ID
	Name         string // 权限名称
	Path         string // 路由地址
	ApiPath      string // 后端路由地址
	ParentId     string // 父级ID
	Sort         string // 排序(数字越小越靠前)
	Icon         string // 菜单图标
	RedirectName string // 重定向名称
	IsRoute      string // 是否路由 0=不是1=是
	IsMenu       string // 是否菜单 0=不是1=是
	IsShow       string // 是否显示 0=不是1=是
	Remark       string // 描述
	IsSupplier   string // 是否门店菜单 0=不是1=是
	CreateTime   string // 创建时间（时间戳）
	UpdateTime   string // 更新时间（时间戳）
	DeleteTime   string // 删除时间（时间戳）
}

// adminAccessColumns holds the columns for the table ttpos_admin_access.
var adminAccessColumns = AdminAccessColumns{
	Id:           "id",
	Name:         "name",
	Path:         "path",
	ApiPath:      "api_path",
	ParentId:     "parent_id",
	Sort:         "sort",
	Icon:         "icon",
	RedirectName: "redirect_name",
	IsRoute:      "is_route",
	IsMenu:       "is_menu",
	IsShow:       "is_show",
	Remark:       "remark",
	IsSupplier:   "is_supplier",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewAdminAccessDao creates and returns a new DAO object for table data access.
func NewAdminAccessDao(handlers ...gdb.ModelHandler) *AdminAccessDao {
	return &AdminAccessDao{
		group:    "default",
		table:    "ttpos_admin_access",
		columns:  adminAccessColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AdminAccessDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AdminAccessDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AdminAccessDao) Columns() AdminAccessColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AdminAccessDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AdminAccessDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AdminAccessDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
