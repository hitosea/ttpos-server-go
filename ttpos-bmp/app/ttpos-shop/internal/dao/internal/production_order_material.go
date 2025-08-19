// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrderMaterialDao is the data access object for the table ttpos_production_order_material.
type ProductionOrderMaterialDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  ProductionOrderMaterialColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// ProductionOrderMaterialColumns defines and stores column names for the table ttpos_production_order_material.
type ProductionOrderMaterialColumns struct {
	Id                         string // 自增ID
	Uuid                       string // 生产订单原料ID
	Name                       string // 原料名称,不随后台改变
	MaterialUuid               string // 原料ID
	Num                        string // 原料数量
	IsProductBom               string // 是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1
	Unit                       string // 单位,不随后台改变
	ProductionOrderProductUuid string // 生产订单商品ID
	SaleOrderProductUuid       string // 销售订单商品ID
	CreateTime                 string // 创建时间(时间戳)
	UpdateTime                 string // 更新时间(时间戳)
	DeleteTime                 string // 删除时间(时间戳)
}

// productionOrderMaterialColumns holds the columns for the table ttpos_production_order_material.
var productionOrderMaterialColumns = ProductionOrderMaterialColumns{
	Id:                         "id",
	Uuid:                       "uuid",
	Name:                       "name",
	MaterialUuid:               "material_uuid",
	Num:                        "num",
	IsProductBom:               "is_product_bom",
	Unit:                       "unit",
	ProductionOrderProductUuid: "production_order_product_uuid",
	SaleOrderProductUuid:       "sale_order_product_uuid",
	CreateTime:                 "create_time",
	UpdateTime:                 "update_time",
	DeleteTime:                 "delete_time",
}

// NewProductionOrderMaterialDao creates and returns a new DAO object for table data access.
func NewProductionOrderMaterialDao(handlers ...gdb.ModelHandler) *ProductionOrderMaterialDao {
	return &ProductionOrderMaterialDao{
		group:    "default",
		table:    "ttpos_production_order_material",
		columns:  productionOrderMaterialColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductionOrderMaterialDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductionOrderMaterialDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductionOrderMaterialDao) Columns() ProductionOrderMaterialColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductionOrderMaterialDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductionOrderMaterialDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductionOrderMaterialDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
