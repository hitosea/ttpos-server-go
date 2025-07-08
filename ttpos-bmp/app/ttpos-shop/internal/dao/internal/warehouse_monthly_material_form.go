// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseMonthlyMaterialFormDao is the data access object for the table ttpos_warehouse_monthly_material_form.
type WarehouseMonthlyMaterialFormDao struct {
	table    string                              // table is the underlying table name of the DAO.
	group    string                              // group is the database configuration group name of the current DAO.
	columns  WarehouseMonthlyMaterialFormColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                  // handlers for customized model modification.
}

// WarehouseMonthlyMaterialFormColumns defines and stores column names for the table ttpos_warehouse_monthly_material_form.
type WarehouseMonthlyMaterialFormColumns struct {
	Id           string // 自增ID
	Uuid         string // 月度报表uuid
	Year         string // 年
	Month        string // 月
	Scene        string // 记录类型,0-月初 1-月末
	MaterialUuid string // 物料uuid
	Stock        string // 库存
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// warehouseMonthlyMaterialFormColumns holds the columns for the table ttpos_warehouse_monthly_material_form.
var warehouseMonthlyMaterialFormColumns = WarehouseMonthlyMaterialFormColumns{
	Id:           "id",
	Uuid:         "uuid",
	Year:         "year",
	Month:        "month",
	Scene:        "scene",
	MaterialUuid: "material_uuid",
	Stock:        "stock",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewWarehouseMonthlyMaterialFormDao creates and returns a new DAO object for table data access.
func NewWarehouseMonthlyMaterialFormDao(handlers ...gdb.ModelHandler) *WarehouseMonthlyMaterialFormDao {
	return &WarehouseMonthlyMaterialFormDao{
		group:    "default",
		table:    "ttpos_warehouse_monthly_material_form",
		columns:  warehouseMonthlyMaterialFormColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseMonthlyMaterialFormDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseMonthlyMaterialFormDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseMonthlyMaterialFormDao) Columns() WarehouseMonthlyMaterialFormColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseMonthlyMaterialFormDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseMonthlyMaterialFormDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseMonthlyMaterialFormDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
