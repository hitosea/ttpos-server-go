// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseOutFormItemDao is the data access object for the table ttpos_warehouse_out_form_item.
type WarehouseOutFormItemDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  WarehouseOutFormItemColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// WarehouseOutFormItemColumns defines and stores column names for the table ttpos_warehouse_out_form_item.
type WarehouseOutFormItemColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 出库单明细uuid
	Num                  string // 数量
	Scene                string // 场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除
	Status               string // 状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示
	ReduceStock          string // 是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存
	RevokeTime           string // 撤销时间(时间戳)
	MaterialUuid         string // 材料uuid
	WarehouseOutFormUuid string // 出库单uuid
	ProductBomUuid       string // 商品BOM表uuid
	SaleOrderProductUuid string // 销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleOrderUuid        string // 销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleBillUuid         string // 销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// warehouseOutFormItemColumns holds the columns for the table ttpos_warehouse_out_form_item.
var warehouseOutFormItemColumns = WarehouseOutFormItemColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Num:                  "num",
	Scene:                "scene",
	Status:               "status",
	ReduceStock:          "reduce_stock",
	RevokeTime:           "revoke_time",
	MaterialUuid:         "material_uuid",
	WarehouseOutFormUuid: "warehouse_out_form_uuid",
	ProductBomUuid:       "product_bom_uuid",
	SaleOrderProductUuid: "sale_order_product_uuid",
	SaleOrderUuid:        "sale_order_uuid",
	SaleBillUuid:         "sale_bill_uuid",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewWarehouseOutFormItemDao creates and returns a new DAO object for table data access.
func NewWarehouseOutFormItemDao(handlers ...gdb.ModelHandler) *WarehouseOutFormItemDao {
	return &WarehouseOutFormItemDao{
		group:    "default",
		table:    "ttpos_warehouse_out_form_item",
		columns:  warehouseOutFormItemColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseOutFormItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseOutFormItemDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseOutFormItemDao) Columns() WarehouseOutFormItemColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseOutFormItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseOutFormItemDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseOutFormItemDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
