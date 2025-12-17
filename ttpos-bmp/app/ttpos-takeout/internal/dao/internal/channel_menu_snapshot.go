// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ChannelMenuSnapshotDao is the data access object for the table takeout_channel_menu_snapshot.
type ChannelMenuSnapshotDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  ChannelMenuSnapshotColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// ChannelMenuSnapshotColumns defines and stores column names for the table takeout_channel_menu_snapshot.
type ChannelMenuSnapshotColumns struct {
	Id             string // 主键ID
	Uuid           string // 唯一标识
	ShopUuid       string // 商户UUID
	ProviderName   string // 渠道名称 (grab, lineman)
	MenuData       string // 菜单数据快照 (JSON)
	CreatedAt      string // 创建时间
	UpdatedAt      string // 更新时间
	DeletedAt      string // 删除时间
	SyncState      string // 同步状态: QUEUEING, PROCESSING, SUCCESS, FAILED
	TtposMenuData  string // TTPOS 侧菜单原始数据 (JSON)
	TtposUpdatedAt string // TTPOS 侧菜单数据更新时间
}

// channelMenuSnapshotColumns holds the columns for the table takeout_channel_menu_snapshot.
var channelMenuSnapshotColumns = ChannelMenuSnapshotColumns{
	Id:             "id",
	Uuid:           "uuid",
	ShopUuid:       "shop_uuid",
	ProviderName:   "provider_name",
	MenuData:       "menu_data",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
	SyncState:      "sync_state",
	TtposMenuData:  "ttpos_menu_data",
	TtposUpdatedAt: "ttpos_updated_at",
}

// NewChannelMenuSnapshotDao creates and returns a new DAO object for table data access.
func NewChannelMenuSnapshotDao(handlers ...gdb.ModelHandler) *ChannelMenuSnapshotDao {
	return &ChannelMenuSnapshotDao{
		group:    "default",
		table:    "takeout_channel_menu_snapshot",
		columns:  channelMenuSnapshotColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ChannelMenuSnapshotDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ChannelMenuSnapshotDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ChannelMenuSnapshotDao) Columns() ChannelMenuSnapshotColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ChannelMenuSnapshotDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ChannelMenuSnapshotDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ChannelMenuSnapshotDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
