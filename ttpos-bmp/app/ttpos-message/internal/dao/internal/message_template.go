// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MessageTemplateDao is the data access object for the table message_template.
type MessageTemplateDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  MessageTemplateColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// MessageTemplateColumns defines and stores column names for the table message_template.
type MessageTemplateColumns struct {
	Id              string // 模板ID
	Uuid            string // 模板UUID
	TemplateName    string // 模板名称
	TemplateType    string // 模板类型(email/sms)
	TemplateSubject string // 模板主题(邮件用)
	TemplateContent string // 模板内容(支持变量)
	TemplateArgs    string // 模板参数定义
	Status          string // 状态(0-禁用,1-启用)
	Remark          string // 备注
	CreatedAt       string // 创建时间
	UpdatedAt       string // 更新时间
	DeletedAt       string // 删除时间
}

// messageTemplateColumns holds the columns for the table message_template.
var messageTemplateColumns = MessageTemplateColumns{
	Id:              "id",
	Uuid:            "uuid",
	TemplateName:    "template_name",
	TemplateType:    "template_type",
	TemplateSubject: "template_subject",
	TemplateContent: "template_content",
	TemplateArgs:    "template_args",
	Status:          "status",
	Remark:          "remark",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewMessageTemplateDao creates and returns a new DAO object for table data access.
func NewMessageTemplateDao(handlers ...gdb.ModelHandler) *MessageTemplateDao {
	return &MessageTemplateDao{
		group:    "default",
		table:    "message_template",
		columns:  messageTemplateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MessageTemplateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MessageTemplateDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MessageTemplateDao) Columns() MessageTemplateColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MessageTemplateDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MessageTemplateDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MessageTemplateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
