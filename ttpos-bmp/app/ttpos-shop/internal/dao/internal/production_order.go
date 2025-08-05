// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrderDao is the data access object for the table ttpos_production_order.
type ProductionOrderDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ProductionOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ProductionOrderColumns defines and stores column names for the table ttpos_production_order.
type ProductionOrderColumns struct {
	Id            string // 自增ID
	Uuid          string // 生产订单ID
	DeskUuid      string // 桌台ID
	SaleOrderUuid string // 销售订单ID
	SaleBillUuid  string // 销售账单ID
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// productionOrderColumns holds the columns for the table ttpos_production_order.
var productionOrderColumns = ProductionOrderColumns{
	Id:            "id",
	Uuid:          "uuid",
	DeskUuid:      "desk_uuid",
	SaleOrderUuid: "sale_order_uuid",
	SaleBillUuid:  "sale_bill_uuid",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewProductionOrderDao creates and returns a new DAO object for table data access.
func NewProductionOrderDao(handlers ...gdb.ModelHandler) *ProductionOrderDao {
	return &ProductionOrderDao{
		group:    "default",
		table:    "ttpos_production_order",
		columns:  productionOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductionOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductionOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductionOrderDao) Columns() ProductionOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductionOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductionOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductionOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
