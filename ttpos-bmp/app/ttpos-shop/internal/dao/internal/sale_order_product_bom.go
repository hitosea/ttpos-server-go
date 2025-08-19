// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductBomDao is the data access object for the table ttpos_sale_order_product_bom.
type SaleOrderProductBomDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  SaleOrderProductBomColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// SaleOrderProductBomColumns defines and stores column names for the table ttpos_sale_order_product_bom.
type SaleOrderProductBomColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 销售订单商品规格或小料ID
	Name                 string // 规格或小料名称,不随后台更新
	Price                string // 单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动
	SaleOrderUuid        string // 销售订单ID
	SaleOrderProductUuid string // 销售订单商品ID
	ProductBomUuid       string // 商品BOM ID
	IsFlavorBom          string // 是否为规格商品BOM, 0-否 1-是
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// saleOrderProductBomColumns holds the columns for the table ttpos_sale_order_product_bom.
var saleOrderProductBomColumns = SaleOrderProductBomColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Name:                 "name",
	Price:                "price",
	SaleOrderUuid:        "sale_order_uuid",
	SaleOrderProductUuid: "sale_order_product_uuid",
	ProductBomUuid:       "product_bom_uuid",
	IsFlavorBom:          "is_flavor_bom",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewSaleOrderProductBomDao creates and returns a new DAO object for table data access.
func NewSaleOrderProductBomDao(handlers ...gdb.ModelHandler) *SaleOrderProductBomDao {
	return &SaleOrderProductBomDao{
		group:    "default",
		table:    "ttpos_sale_order_product_bom",
		columns:  saleOrderProductBomColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderProductBomDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderProductBomDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderProductBomDao) Columns() SaleOrderProductBomColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderProductBomDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderProductBomDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderProductBomDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
