// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderItemDao is the data access object for the table takeout_order_item.
type OrderItemDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  OrderItemColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// OrderItemColumns defines and stores column names for the table takeout_order_item.
type OrderItemColumns struct {
	Id                    string // 主键
	OrderUuid             string // 关联订单UUID
	ProviderItemId        string // 平台商品ID (Grab Item ID)
	ItemName              string // 商品名称
	Quantity              string // 数量
	Price                 string // 单价
	TotalPrice            string // 总价
	Specifications        string // 规格描述(JSON格式)
	Modifiers             string // 修改项/配料 (JSON)
	Note                  string // 商品备注
	OutOfStockInstruction string // 缺货处理方式
	CreatedAt             string // 创建时间
}

// orderItemColumns holds the columns for the table takeout_order_item.
var orderItemColumns = OrderItemColumns{
	Id:                    "id",
	OrderUuid:             "order_uuid",
	ProviderItemId:        "provider_item_id",
	ItemName:              "item_name",
	Quantity:              "quantity",
	Price:                 "price",
	TotalPrice:            "total_price",
	Specifications:        "specifications",
	Modifiers:             "modifiers",
	Note:                  "note",
	OutOfStockInstruction: "out_of_stock_instruction",
	CreatedAt:             "created_at",
}

// NewOrderItemDao creates and returns a new DAO object for table data access.
func NewOrderItemDao(handlers ...gdb.ModelHandler) *OrderItemDao {
	return &OrderItemDao{
		group:    "default",
		table:    "takeout_order_item",
		columns:  orderItemColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OrderItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OrderItemDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OrderItemDao) Columns() OrderItemColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OrderItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OrderItemDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OrderItemDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
