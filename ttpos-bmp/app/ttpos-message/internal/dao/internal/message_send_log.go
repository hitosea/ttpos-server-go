// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MessageSendLogDao is the data access object for the table message_send_log.
type MessageSendLogDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MessageSendLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MessageSendLogColumns defines and stores column names for the table message_send_log.
type MessageSendLogColumns struct {
	Id           string // 日志ID
	MessageUuid  string // 消息UUID
	SendTime     string // 发送时间
	SendResult   string // 发送结果(0-失败,1-成功)
	ErrorMessage string // 错误信息
	RequestData  string // 请求数据
	ResponseData string // 响应数据
	CreatedAt    string // 创建时间
}

// messageSendLogColumns holds the columns for the table message_send_log.
var messageSendLogColumns = MessageSendLogColumns{
	Id:           "id",
	MessageUuid:  "message_uuid",
	SendTime:     "send_time",
	SendResult:   "send_result",
	ErrorMessage: "error_message",
	RequestData:  "request_data",
	ResponseData: "response_data",
	CreatedAt:    "created_at",
}

// NewMessageSendLogDao creates and returns a new DAO object for table data access.
func NewMessageSendLogDao(handlers ...gdb.ModelHandler) *MessageSendLogDao {
	return &MessageSendLogDao{
		group:    "default",
		table:    "message_send_log",
		columns:  messageSendLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MessageSendLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MessageSendLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MessageSendLogDao) Columns() MessageSendLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MessageSendLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MessageSendLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MessageSendLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
