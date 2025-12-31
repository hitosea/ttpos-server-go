// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MenuLogDao is the data access object for the table takeout_menu_log.
type MenuLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MenuLogColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MenuLogColumns defines and stores column names for the table takeout_menu_log.
type MenuLogColumns struct {
	Id           string // 主键
	Uuid         string // 唯一ID
	MerchantId   string // 商户ID
	ProviderName string // 渠道: grab
	SyncType     string // 同步类型: FULL, PARTIAL, NOTIFY
	RequestId    string // 请求ID (来自Grab)
	Status       string // 同步状态: QUEUED, SUCCESS, FAIL, PROCESSING
	MenuSnapshot string // 菜单快照(JSON)
	ErrorCode    string // 错误代码
	ErrorMsg     string // 错误信息
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// menuLogColumns holds the columns for the table takeout_menu_log.
var menuLogColumns = MenuLogColumns{
	Id:           "id",
	Uuid:         "uuid",
	MerchantId:   "merchant_id",
	ProviderName: "provider_name",
	SyncType:     "sync_type",
	RequestId:    "request_id",
	Status:       "status",
	MenuSnapshot: "menu_snapshot",
	ErrorCode:    "error_code",
	ErrorMsg:     "error_msg",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewMenuLogDao creates and returns a new DAO object for table data access.
func NewMenuLogDao(handlers ...gdb.ModelHandler) *MenuLogDao {
	return &MenuLogDao{
		group:    "default",
		table:    "takeout_menu_log",
		columns:  menuLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MenuLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MenuLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MenuLogDao) Columns() MenuLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MenuLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MenuLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MenuLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
