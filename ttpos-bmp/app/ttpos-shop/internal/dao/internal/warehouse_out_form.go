// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseOutFormDao is the data access object for the table ttpos_warehouse_out_form.
type WarehouseOutFormDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  WarehouseOutFormColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// WarehouseOutFormColumns defines and stores column names for the table ttpos_warehouse_out_form.
type WarehouseOutFormColumns struct {
	Id                  string // 自增ID
	Uuid                string // 出库单uuid
	FormNo              string // 编号
	Scene               string // 出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库
	Remark              string // 备注
	Status              string // 状态,0-success已出库 1-canceled已撤销
	OperatorUuid        string // 操作员uuid
	AssociatedOrderUuid string // 关联订单uuid
	RevokeTime          string // 撤销时间(时间戳)
	CreateTime          string // 创建时间(时间戳)
	UpdateTime          string // 更新时间(时间戳)
	DeleteTime          string // 删除时间(时间戳)
}

// warehouseOutFormColumns holds the columns for the table ttpos_warehouse_out_form.
var warehouseOutFormColumns = WarehouseOutFormColumns{
	Id:                  "id",
	Uuid:                "uuid",
	FormNo:              "form_no",
	Scene:               "scene",
	Remark:              "remark",
	Status:              "status",
	OperatorUuid:        "operator_uuid",
	AssociatedOrderUuid: "associated_order_uuid",
	RevokeTime:          "revoke_time",
	CreateTime:          "create_time",
	UpdateTime:          "update_time",
	DeleteTime:          "delete_time",
}

// NewWarehouseOutFormDao creates and returns a new DAO object for table data access.
func NewWarehouseOutFormDao(handlers ...gdb.ModelHandler) *WarehouseOutFormDao {
	return &WarehouseOutFormDao{
		group:    "default",
		table:    "ttpos_warehouse_out_form",
		columns:  warehouseOutFormColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseOutFormDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseOutFormDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseOutFormDao) Columns() WarehouseOutFormColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseOutFormDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseOutFormDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseOutFormDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
