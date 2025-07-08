// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseFormDao is the data access object for the table ttpos_warehouse_form.
type WarehouseFormDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  WarehouseFormColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// WarehouseFormColumns defines and stores column names for the table ttpos_warehouse_form.
type WarehouseFormColumns struct {
	Id                string // 自增ID
	Uuid              string // 库存入库单ID
	FormNo            string // 编号
	Scene             string // 交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库
	Num               string // 数量
	Remark            string // 备注
	Status            string // 状态,0-success已入库 1-canceled已撤销
	ProductBomUuid    string // 商品BOM表uuid
	MaterialUuid      string // 材料uuid
	PurchaseOrderUuid string // 采购订单uuid
	OperatorUuid      string // 操作员uuid
	RevokeTime        string // 撤销时间(时间戳)
	CreateTime        string // 创建时间(时间戳)
	UpdateTime        string // 更新时间(时间戳)
	DeleteTime        string // 删除时间(时间戳)
}

// warehouseFormColumns holds the columns for the table ttpos_warehouse_form.
var warehouseFormColumns = WarehouseFormColumns{
	Id:                "id",
	Uuid:              "uuid",
	FormNo:            "form_no",
	Scene:             "scene",
	Num:               "num",
	Remark:            "remark",
	Status:            "status",
	ProductBomUuid:    "product_bom_uuid",
	MaterialUuid:      "material_uuid",
	PurchaseOrderUuid: "purchase_order_uuid",
	OperatorUuid:      "operator_uuid",
	RevokeTime:        "revoke_time",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewWarehouseFormDao creates and returns a new DAO object for table data access.
func NewWarehouseFormDao(handlers ...gdb.ModelHandler) *WarehouseFormDao {
	return &WarehouseFormDao{
		group:    "default",
		table:    "ttpos_warehouse_form",
		columns:  warehouseFormColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WarehouseFormDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WarehouseFormDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WarehouseFormDao) Columns() WarehouseFormColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WarehouseFormDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WarehouseFormDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WarehouseFormDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
