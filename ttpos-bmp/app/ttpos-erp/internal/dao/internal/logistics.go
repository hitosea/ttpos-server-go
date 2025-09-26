// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LogisticsDao is the data access object for the table erp_logistics.
type LogisticsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  LogisticsColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// LogisticsColumns defines and stores column names for the table erp_logistics.
type LogisticsColumns struct {
	Id           string // ID
	Uuid         string // UUID
	Vendor       string // 供应商，如 JT:极兔
	VendorUserId string // 供应商用户id,如极兔的货主编码
	InfConf      string // 接口连接信息。如ak/sk 根据不同供应商有所不同
	Remarks      string // 备注信息
	Reserve1     string // 保留字段1
	Reserve2     string // 保留字段2
}

// logisticsColumns holds the columns for the table erp_logistics.
var logisticsColumns = LogisticsColumns{
	Id:           "id",
	Uuid:         "uuid",
	Vendor:       "vendor",
	VendorUserId: "vendor_user_id",
	InfConf:      "inf_conf",
	Remarks:      "remarks",
	Reserve1:     "reserve1",
	Reserve2:     "reserve2",
}

// NewLogisticsDao creates and returns a new DAO object for table data access.
func NewLogisticsDao(handlers ...gdb.ModelHandler) *LogisticsDao {
	return &LogisticsDao{
		group:    "default",
		table:    "erp_logistics",
		columns:  logisticsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LogisticsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LogisticsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LogisticsDao) Columns() LogisticsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LogisticsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LogisticsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *LogisticsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
