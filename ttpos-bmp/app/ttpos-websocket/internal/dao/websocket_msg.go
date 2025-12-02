// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// websocketMsgDao DAO管理对象
type websocketMsgDao struct {
	table   string                 // 表名
	group   string                 // 数据库配置组
	columns websocketMsgDaoColumns // 字段信息
}

// websocketMsgDaoColumns WebsocketMsg表字段
type websocketMsgDaoColumns struct {
	Id           string //
	Uuid         string //
	CompanyUuid  string //
	Uid          string //
	Msg          string //
	Type         string //
	SourceClient string //
	Status       string //
	IsOffline    string //
	CreateTime   string //
	UpdateTime   string //
	DeleteTime   string //
}

// websocketMsgColumns WebsocketMsg表字段
var websocketMsgColumns = websocketMsgDaoColumns{
	Id:           "id",
	Uuid:         "uuid",
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

// newWebsocketMsgDao 创建WebsocketMsg DAO
func newWebsocketMsgDao() *websocketMsgDao {
	return &websocketMsgDao{
		group:   "default",
		table:   "websocket_msg",
		columns: websocketMsgColumns,
	}
}

// DB 获取当前DAO操作的数据库连接
func (dao *websocketMsgDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table 返回表名
func (dao *websocketMsgDao) Table() string {
	return dao.table
}

// Columns 返回字段信息
func (dao *websocketMsgDao) Columns() websocketMsgDaoColumns {
	return dao.columns
}

// Group 返回数据库配置组
func (dao *websocketMsgDao) Group() string {
	return dao.group
}

// Ctx 创建并返回Model，自动设置上下文和表名
func (dao *websocketMsgDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *websocketMsgDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.DB().Transaction(ctx, f)
}
