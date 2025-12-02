// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceOnlineDao is the data access object for the table device_online.
type DeviceOnlineDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  DeviceOnlineColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// DeviceOnlineColumns defines and stores column names for the table device_online.
type DeviceOnlineColumns struct {
	Id                string // 主键ID
	CompanyUuid       string // 公司UUID
	StaffUuid         string // 员工UUID
	DeviceId          string // 设备ID
	SourceClient      string // 来源客户端：pos/tablet/kitchen/h5/mobile
	ConnectionKey     string // 连接唯一标识
	Status            string // 在线状态：0-离线，1-在线
	ConnectTime       string // 连接时间
	DisconnectTime    string // 断开时间
	LastHeartbeatTime string // 最后心跳时间
	IpAddress         string // IP地址
	UserAgent         string // 用户代理信息
	CreateTime        string // 创建时间
	UpdateTime        string // 更新时间
	DeleteTime        string // 删除时间（软删除）
}

// deviceOnlineColumns holds the columns for the table device_online.
var deviceOnlineColumns = DeviceOnlineColumns{
	Id:                "id",
	CompanyUuid:       "company_uuid",
	StaffUuid:         "staff_uuid",
	DeviceId:          "device_id",
	SourceClient:      "source_client",
	ConnectionKey:     "connection_key",
	Status:            "status",
	ConnectTime:       "connect_time",
	DisconnectTime:    "disconnect_time",
	LastHeartbeatTime: "last_heartbeat_time",
	IpAddress:         "ip_address",
	UserAgent:         "user_agent",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewDeviceOnlineDao creates and returns a new DAO object for table data access.
func NewDeviceOnlineDao(handlers ...gdb.ModelHandler) *DeviceOnlineDao {
	return &DeviceOnlineDao{
		group:    "default",
		table:    "device_online",
		columns:  deviceOnlineColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DeviceOnlineDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DeviceOnlineDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DeviceOnlineDao) Columns() DeviceOnlineColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DeviceOnlineDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DeviceOnlineDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DeviceOnlineDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
