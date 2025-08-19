// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductReasonDao is the data access object for the table ttpos_sale_order_product_reason.
type SaleOrderProductReasonDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  SaleOrderProductReasonColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// SaleOrderProductReasonColumns defines and stores column names for the table ttpos_sale_order_product_reason.
type SaleOrderProductReasonColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 自增UUID
	SaleOrderUuid         string // 销售订单ID
	SaleOrderProductUuid  string // 销售订单商品ID，如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0
	ReturnFoodReasonUuid  string // 退菜原因ID
	FreeReasonUuid        string // 免单原因ID
	GiftReasonUuid        string // 赠菜原因ID
	MultiLanguageNameUuid string // 原因-多语言名称ID
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// saleOrderProductReasonColumns holds the columns for the table ttpos_sale_order_product_reason.
var saleOrderProductReasonColumns = SaleOrderProductReasonColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	SaleOrderUuid:         "sale_order_uuid",
	SaleOrderProductUuid:  "sale_order_product_uuid",
	ReturnFoodReasonUuid:  "return_food_reason_uuid",
	FreeReasonUuid:        "free_reason_uuid",
	GiftReasonUuid:        "gift_reason_uuid",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewSaleOrderProductReasonDao creates and returns a new DAO object for table data access.
func NewSaleOrderProductReasonDao(handlers ...gdb.ModelHandler) *SaleOrderProductReasonDao {
	return &SaleOrderProductReasonDao{
		group:    "default",
		table:    "ttpos_sale_order_product_reason",
		columns:  saleOrderProductReasonColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderProductReasonDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderProductReasonDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderProductReasonDao) Columns() SaleOrderProductReasonColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderProductReasonDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderProductReasonDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderProductReasonDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
