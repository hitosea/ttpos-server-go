// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DeskDao is the data access object for the table ttpos_desk.
type DeskDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  DeskColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// DeskColumns defines and stores column names for the table ttpos_desk.
type DeskColumns struct {
	Id             string // 自增ID
	Uuid           string // 桌台ID
	DeskNo         string // 桌位编号
	RegionUuid     string // 桌台区域ID
	TypeUuid       string // 桌台类型ID
	Sort           string // 排序序号
	Status         string // 状态, 0-未开台 1-已开台
	IsDisable      string // 是否禁用, 0-否 1-是
	NeedServiceFee string // 是否需要服务费, 0-否 1-是.标记该桌台收取服务费
	QrcodeToken    string // 二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效
	SaleBillUuid   string // 销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单
	DeviceUuid     string // 平板设备uuid, 0-未绑定
	CreateTime     string // 创建时间(时间戳)
	UpdateTime     string // 更新时间(时间戳)
	DeleteTime     string // 删除时间(时间戳)
}

// deskColumns holds the columns for the table ttpos_desk.
var deskColumns = DeskColumns{
	Id:             "id",
	Uuid:           "uuid",
	DeskNo:         "desk_no",
	RegionUuid:     "region_uuid",
	TypeUuid:       "type_uuid",
	Sort:           "sort",
	Status:         "status",
	IsDisable:      "is_disable",
	NeedServiceFee: "need_service_fee",
	QrcodeToken:    "qrcode_token",
	SaleBillUuid:   "sale_bill_uuid",
	DeviceUuid:     "device_uuid",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewDeskDao creates and returns a new DAO object for table data access.
func NewDeskDao(handlers ...gdb.ModelHandler) *DeskDao {
	return &DeskDao{
		group:    "default",
		table:    "ttpos_desk",
		columns:  deskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DeskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DeskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DeskDao) Columns() DeskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DeskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DeskDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DeskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
