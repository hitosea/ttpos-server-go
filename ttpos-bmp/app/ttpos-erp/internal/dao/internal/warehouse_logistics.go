// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseLogisticsDao is the data access object for the table erp_warehouse_logistics.
type WarehouseLogisticsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  WarehouseLogisticsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// WarehouseLogisticsColumns defines and stores column names for the table erp_warehouse_logistics.
type WarehouseLogisticsColumns struct {
	Id            string // ID
	SiteCode      string // 站点编码。 关联 erp_site.site_code
	ShopUuid      string // ttpos商铺ID
	WarehouseCode string // 仓库编码. erpnext warehouse
	LogisticsId   string // 物流ID
}

// warehouseLogisticsColumns holds the columns for the table erp_warehouse_logistics.
var warehouseLogisticsColumns = WarehouseLogisticsColumns{
	Id:            "id",
	SiteCode:      "site_code",
	ShopUuid:      "shop_uuid",
	WarehouseCode: "warehouse_code",
	LogisticsId:   "logistics_id",
}

// NewWarehouseLogisticsDao creates and returns a new DAO object for table data access.
func NewWarehouseLogisticsDao(handlers ...gdb.ModelHandler) *WarehouseLogisticsDao {
	return &WarehouseLogisticsDao{
		group:    "default",
		table:    "erp_warehouse_logistics",
		columns:  warehouseLogisticsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseLogisticsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseLogisticsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseLogisticsDao) Columns() WarehouseLogisticsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseLogisticsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseLogisticsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseLogisticsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
