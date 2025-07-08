// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LanPrinterScanDao is the data access object for the table ttpos_lan_printer_scan.
type LanPrinterScanDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  LanPrinterScanColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// LanPrinterScanColumns defines and stores column names for the table ttpos_lan_printer_scan.
type LanPrinterScanColumns struct {
	Id             string //
	Uuid           string // uuid
	Ip             string // ip
	Port           string // 端口
	Status         string // 状态 0: 离线 1: 在线
	Remark         string // 备注
	SourceDeviceSn string // 来源设备SN
	CreateTime     string // 创建时间
	UpdateTime     string // 更新时间
	DeleteTime     string // 删除时间
}

// lanPrinterScanColumns holds the columns for the table ttpos_lan_printer_scan.
var lanPrinterScanColumns = LanPrinterScanColumns{
	Id:             "id",
	Uuid:           "uuid",
	Ip:             "ip",
	Port:           "port",
	Status:         "status",
	Remark:         "remark",
	SourceDeviceSn: "source_device_sn",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewLanPrinterScanDao creates and returns a new DAO object for table data access.
func NewLanPrinterScanDao(handlers ...gdb.ModelHandler) *LanPrinterScanDao {
	return &LanPrinterScanDao{
		group:    "default",
		table:    "ttpos_lan_printer_scan",
		columns:  lanPrinterScanColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LanPrinterScanDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LanPrinterScanDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LanPrinterScanDao) Columns() LanPrinterScanColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LanPrinterScanDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LanPrinterScanDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *LanPrinterScanDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
