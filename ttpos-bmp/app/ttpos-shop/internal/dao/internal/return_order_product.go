// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrderProductDao is the data access object for the table ttpos_return_order_product.
type ReturnOrderProductDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  ReturnOrderProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// ReturnOrderProductColumns defines and stores column names for the table ttpos_return_order_product.
type ReturnOrderProductColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 退货单商品唯一标识符
	SaleOrderUuid        string // 销售订单ID
	SaleOrderProductUuid string // 销售订单商品表ID
	ReturnOrderUuid      string // 退货单ID
	ProductType          string // 商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct
	ProductPackageUuid   string // 商品包ID
	ProductName          string // 商品名称
	ProductPrice         string // 商品单价
	TaxRate              string // 税率,根据结账时税率计算
	Num                  string // 商品数量,退货的商品数量
	ProductDiscount      string // 商品折扣
	ProductTotalAmount   string // 商品总金额（退款总金额）
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// returnOrderProductColumns holds the columns for the table ttpos_return_order_product.
var returnOrderProductColumns = ReturnOrderProductColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	SaleOrderUuid:        "sale_order_uuid",
	SaleOrderProductUuid: "sale_order_product_uuid",
	ReturnOrderUuid:      "return_order_uuid",
	ProductType:          "product_type",
	ProductPackageUuid:   "product_package_uuid",
	ProductName:          "product_name",
	ProductPrice:         "product_price",
	TaxRate:              "tax_rate",
	Num:                  "num",
	ProductDiscount:      "product_discount",
	ProductTotalAmount:   "product_total_amount",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewReturnOrderProductDao creates and returns a new DAO object for table data access.
func NewReturnOrderProductDao(handlers ...gdb.ModelHandler) *ReturnOrderProductDao {
	return &ReturnOrderProductDao{
		group:    "default",
		table:    "ttpos_return_order_product",
		columns:  returnOrderProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReturnOrderProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReturnOrderProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReturnOrderProductDao) Columns() ReturnOrderProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReturnOrderProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReturnOrderProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReturnOrderProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
