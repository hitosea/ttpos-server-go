// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CallbackMsgDao is the data access object for the table takeout_callback_msg.
type CallbackMsgDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CallbackMsgColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CallbackMsgColumns defines and stores column names for the table takeout_callback_msg.
type CallbackMsgColumns struct {
	Id             string // 主键
	Uuid           string // 全局唯一ID
	TakeoutRefNo   string // 外送系统订单号，如skootar.jobId
	Content        string // 消息内容
	StatusDatetime string // 状态变更时间
	CreatedAt      string // 创建时间
	UpdatedAt      string // 修改时间
	DeletedAt      string // 软删除
}

// callbackMsgColumns holds the columns for the table takeout_callback_msg.
var callbackMsgColumns = CallbackMsgColumns{
	Id:             "id",
	Uuid:           "uuid",
	TakeoutRefNo:   "takeout_ref_no",
	Content:        "content",
	StatusDatetime: "status_datetime",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewCallbackMsgDao creates and returns a new DAO object for table data access.
func NewCallbackMsgDao(handlers ...gdb.ModelHandler) *CallbackMsgDao {
	return &CallbackMsgDao{
		group:    "default",
		table:    "takeout_callback_msg",
		columns:  callbackMsgColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CallbackMsgDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CallbackMsgDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CallbackMsgDao) Columns() CallbackMsgColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CallbackMsgDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CallbackMsgDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CallbackMsgDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
