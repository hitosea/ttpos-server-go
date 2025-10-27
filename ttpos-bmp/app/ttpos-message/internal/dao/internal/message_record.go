// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MessageRecordDao is the data access object for the table message_record.
type MessageRecordDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MessageRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MessageRecordColumns defines and stores column names for the table message_record.
type MessageRecordColumns struct {
	Id           string // 消息ID
	Uuid         string // 消息UUID
	TemplateId   string // 模板ID
	MessageType  string // 消息类型(email/sms)
	Recipient    string // 接收人
	Subject      string // 消息主题
	Content      string // 消息内容(渲染后)
	MessageArgs  string // 消息参数
	Status       string // 状态(0-待发送,1-发送中,2-发送成功,3-发送失败)
	ErrorMessage string // 错误信息
	RetryCount   string // 重试次数
	CompanyUuid  string // 公司UUID
	OperatorUuid string // 操作人UUID
	SendTime     string // 发送时间
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 删除时间
}

// messageRecordColumns holds the columns for the table message_record.
var messageRecordColumns = MessageRecordColumns{
	Id:           "id",
	Uuid:         "uuid",
	TemplateId:   "template_id",
	MessageType:  "message_type",
	Recipient:    "recipient",
	Subject:      "subject",
	Content:      "content",
	MessageArgs:  "message_args",
	Status:       "status",
	ErrorMessage: "error_message",
	RetryCount:   "retry_count",
	CompanyUuid:  "company_uuid",
	OperatorUuid: "operator_uuid",
	SendTime:     "send_time",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewMessageRecordDao creates and returns a new DAO object for table data access.
func NewMessageRecordDao(handlers ...gdb.ModelHandler) *MessageRecordDao {
	return &MessageRecordDao{
		group:    "default",
		table:    "message_record",
		columns:  messageRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MessageRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MessageRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MessageRecordDao) Columns() MessageRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MessageRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MessageRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MessageRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
