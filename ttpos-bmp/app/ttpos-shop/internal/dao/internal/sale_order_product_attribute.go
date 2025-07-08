// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductAttributeDao is the data access object for the table ttpos_sale_order_product_attribute.
type SaleOrderProductAttributeDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  SaleOrderProductAttributeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// SaleOrderProductAttributeColumns defines and stores column names for the table ttpos_sale_order_product_attribute.
type SaleOrderProductAttributeColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 商品属性ID
	Name                 string // 商品属性名称,不随后台更新
	SaleOrderUuid        string // 销售订单ID
	SaleOrderProductUuid string // 销售订单商品ID
	ProductAttributeUuid string // 商品属性ID
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// saleOrderProductAttributeColumns holds the columns for the table ttpos_sale_order_product_attribute.
var saleOrderProductAttributeColumns = SaleOrderProductAttributeColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Name:                 "name",
	SaleOrderUuid:        "sale_order_uuid",
	SaleOrderProductUuid: "sale_order_product_uuid",
	ProductAttributeUuid: "product_attribute_uuid",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewSaleOrderProductAttributeDao creates and returns a new DAO object for table data access.
func NewSaleOrderProductAttributeDao(handlers ...gdb.ModelHandler) *SaleOrderProductAttributeDao {
	return &SaleOrderProductAttributeDao{
		group:    "default",
		table:    "ttpos_sale_order_product_attribute",
		columns:  saleOrderProductAttributeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderProductAttributeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderProductAttributeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderProductAttributeDao) Columns() SaleOrderProductAttributeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderProductAttributeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderProductAttributeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderProductAttributeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
