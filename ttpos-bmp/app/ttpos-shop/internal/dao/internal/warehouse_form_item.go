// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseFormItemDao is the data access object for the table ttpos_warehouse_form_item.
type WarehouseFormItemDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  WarehouseFormItemColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// WarehouseFormItemColumns defines and stores column names for the table ttpos_warehouse_form_item.
type WarehouseFormItemColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 入库单明细uuid
	Num                  string // 入库数量
	Scene                string // 场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面
	AddStock             string // 是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存
	MaterialUuid         string // 材料uuid
	ProductBomUuid       string // 商品BOM表uuid
	WarehouseFormUuid    string // 入库单uuid
	SaleOrderProductUuid string // 销售订单商品uuid,用于退菜入库
	SaleBillUuid         string // 销售账单uuid,用于退菜入库
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// warehouseFormItemColumns holds the columns for the table ttpos_warehouse_form_item.
var warehouseFormItemColumns = WarehouseFormItemColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Num:                  "num",
	Scene:                "scene",
	AddStock:             "add_stock",
	MaterialUuid:         "material_uuid",
	ProductBomUuid:       "product_bom_uuid",
	WarehouseFormUuid:    "warehouse_form_uuid",
	SaleOrderProductUuid: "sale_order_product_uuid",
	SaleBillUuid:         "sale_bill_uuid",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewWarehouseFormItemDao creates and returns a new DAO object for table data access.
func NewWarehouseFormItemDao(handlers ...gdb.ModelHandler) *WarehouseFormItemDao {
	return &WarehouseFormItemDao{
		group:    "default",
		table:    "ttpos_warehouse_form_item",
		columns:  warehouseFormItemColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseFormItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseFormItemDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseFormItemDao) Columns() WarehouseFormItemColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseFormItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseFormItemDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseFormItemDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
