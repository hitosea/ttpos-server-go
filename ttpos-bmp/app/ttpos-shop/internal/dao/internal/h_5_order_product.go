// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// H5OrderProductDao is the data access object for the table ttpos_h5_order_product.
type H5OrderProductDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  H5OrderProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// H5OrderProductColumns defines and stores column names for the table ttpos_h5_order_product.
type H5OrderProductColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 扫码订单商品uuid
	Name                 string // 商品名称.接单和拒单后从sale_order_product表获取，不再改变
	Price                string // 最终单价（折后价）。接单和拒单后从sale_order_product表获取，不再改变
	SalePrice            string // 销售价（折前价）。接单和拒单后从sale_order_product表获取，不再改变
	Num                  string // 最终商品数量.接单和拒单后从sale_order_product表获取，不再改变
	AttributeText        string // 商品属性文本。接单和拒单后从sale_order_product表获取，不再改变
	Remark               string // 备注。接单和拒单后从sale_order_product表获取，不再改变
	SaleOrderProductUuid string // 销售订单商品uuid
	H5OrderUuid          string // 扫码订单uuid
	SaleBillUuid         string // 销售账单uuid
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// h5OrderProductColumns holds the columns for the table ttpos_h5_order_product.
var h5OrderProductColumns = H5OrderProductColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Name:                 "name",
	Price:                "price",
	SalePrice:            "sale_price",
	Num:                  "num",
	AttributeText:        "attribute_text",
	Remark:               "remark",
	SaleOrderProductUuid: "sale_order_product_uuid",
	H5OrderUuid:          "h5_order_uuid",
	SaleBillUuid:         "sale_bill_uuid",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewH5OrderProductDao creates and returns a new DAO object for table data access.
func NewH5OrderProductDao(handlers ...gdb.ModelHandler) *H5OrderProductDao {
	return &H5OrderProductDao{
		group:    "default",
		table:    "ttpos_h5_order_product",
		columns:  h5OrderProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *H5OrderProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *H5OrderProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *H5OrderProductDao) Columns() H5OrderProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *H5OrderProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *H5OrderProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *H5OrderProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
