// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderDiscountStrategyDao is the data access object for the table ttpos_sale_order_discount_strategy.
type SaleOrderDiscountStrategyDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  SaleOrderDiscountStrategyColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// SaleOrderDiscountStrategyColumns defines and stores column names for the table ttpos_sale_order_discount_strategy.
type SaleOrderDiscountStrategyColumns struct {
	Id            string // 自增ID
	Uuid          string // 销售订单优惠策略ID
	Type          string // 优惠策略类型,0-整单折扣、1-会员折扣
	Name          string // 优惠策略名称
	Value         string // 优惠策略值
	JsonField     string // JSON字段
	SaleOrderUuid string // 销售订单ID
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// saleOrderDiscountStrategyColumns holds the columns for the table ttpos_sale_order_discount_strategy.
var saleOrderDiscountStrategyColumns = SaleOrderDiscountStrategyColumns{
	Id:            "id",
	Uuid:          "uuid",
	Type:          "type",
	Name:          "name",
	Value:         "value",
	JsonField:     "json_field",
	SaleOrderUuid: "sale_order_uuid",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewSaleOrderDiscountStrategyDao creates and returns a new DAO object for table data access.
func NewSaleOrderDiscountStrategyDao(handlers ...gdb.ModelHandler) *SaleOrderDiscountStrategyDao {
	return &SaleOrderDiscountStrategyDao{
		group:    "default",
		table:    "ttpos_sale_order_discount_strategy",
		columns:  saleOrderDiscountStrategyColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderDiscountStrategyDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderDiscountStrategyDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderDiscountStrategyDao) Columns() SaleOrderDiscountStrategyColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderDiscountStrategyDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderDiscountStrategyDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderDiscountStrategyDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
