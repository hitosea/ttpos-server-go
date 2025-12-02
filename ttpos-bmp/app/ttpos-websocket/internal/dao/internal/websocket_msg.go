// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WebsocketMsgDao is the data access object for the table websocket_msg.
type WebsocketMsgDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  WebsocketMsgColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// WebsocketMsgColumns defines and stores column names for the table websocket_msg.
type WebsocketMsgColumns struct {
	Id           string // 主键ID
	CompanyUuid  string // 公司UUID
	Uid          string // 用户/设备标识
	Msg          string // 消息内容（JSON格式）
	Type         string // 消息类型：heartbeat/order/notification/system/broadcast
	SourceClient string // 来源客户端：pos/tablet/kitchen/h5/mobile
	Status       string // 消息状态：0-待发送，1-发送中，2-发送成功，3-发送失败
	IsOffline    string // 是否离线消息：0-在线消息，1-离线消息
	CreateTime   string // 创建时间
	UpdateTime   string // 更新时间
	DeleteTime   string // 删除时间（软删除）
}

// websocketMsgColumns holds the columns for the table websocket_msg.
var websocketMsgColumns = WebsocketMsgColumns{
	Id:           "id",
	CompanyUuid:  "company_uuid",
	Uid:          "uid",
	Msg:          "msg",
	Type:         "type",
	SourceClient: "source_client",
	Status:       "status",
	IsOffline:    "is_offline",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewWebsocketMsgDao creates and returns a new DAO object for table data access.
func NewWebsocketMsgDao(handlers ...gdb.ModelHandler) *WebsocketMsgDao {
	return &WebsocketMsgDao{
		group:    "default",
		table:    "websocket_msg",
		columns:  websocketMsgColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WebsocketMsgDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WebsocketMsgDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WebsocketMsgDao) Columns() WebsocketMsgColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WebsocketMsgDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WebsocketMsgDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WebsocketMsgDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
