// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceDao is the data access object for the table ttpos_device.
type DeviceDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  DeviceColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// DeviceColumns defines and stores column names for the table ttpos_device.
type DeviceColumns struct {
	Id                 string // 自增ID
	Uuid               string // 绑定记录ID
	FinallyLoginUuid   string // 最后一个登录id, 退出会清为0
	FinallyLoginTime   string // 最后登录时间
	Source             string // 来源 cashier-收银机 tablet-平板端 kitchen-厨显端
	DeviceId           string // 唯一设备标识id
	IsMain             string // 是否主设备 0-常规 1-主
	ProductPrinterUuid string // 打印档口Uuid
	Address            string // 绑定地址
	Port               string // 绑定端口
	DeviceIp           string // 设备ip
	Remark             string // 备注
	Brand              string // 品牌名称
	Platform           string // 平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端
	UserAgent          string // 请求头信息
	CashSign           string // 收银终端标识
	CashBoxId          string // 现金箱ID
	AccessToken        string // 访问令牌
	QueueUrl           string // 关联队列url
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// deviceColumns holds the columns for the table ttpos_device.
var deviceColumns = DeviceColumns{
	Id:                 "id",
	Uuid:               "uuid",
	FinallyLoginUuid:   "finally_login_uuid",
	FinallyLoginTime:   "finally_login_time",
	Source:             "source",
	DeviceId:           "device_id",
	IsMain:             "is_main",
	ProductPrinterUuid: "product_printer_uuid",
	Address:            "address",
	Port:               "port",
	DeviceIp:           "device_ip",
	Remark:             "remark",
	Brand:              "brand",
	Platform:           "platform",
	UserAgent:          "user_agent",
	CashSign:           "cash_sign",
	CashBoxId:          "cash_box_id",
	AccessToken:        "access_token",
	QueueUrl:           "queue_url",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewDeviceDao creates and returns a new DAO object for table data access.
func NewDeviceDao(handlers ...gdb.ModelHandler) *DeviceDao {
	return &DeviceDao{
		group:    "default",
		table:    "ttpos_device",
		columns:  deviceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DeviceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DeviceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DeviceDao) Columns() DeviceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DeviceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DeviceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DeviceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
