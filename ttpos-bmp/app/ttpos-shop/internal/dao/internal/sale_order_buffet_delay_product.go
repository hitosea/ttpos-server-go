// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderBuffetDelayProductDao is the data access object for the table ttpos_sale_order_buffet_delay_product.
type SaleOrderBuffetDelayProductDao struct {
	table    string                             // table is the underlying table name of the DAO.
	group    string                             // group is the database configuration group name of the current DAO.
	columns  SaleOrderBuffetDelayProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                 // handlers for customized model modification.
}

// SaleOrderBuffetDelayProductColumns defines and stores column names for the table ttpos_sale_order_buffet_delay_product.
type SaleOrderBuffetDelayProductColumns struct {
	Id              string // 自增ID
	Uuid            string // 自助餐加钟价格ID
	SaleOrderUuid   string // 销售订单ID
	BuffetDelayUuid string // 自助餐加钟价格ID
	Name            string // 自助餐加钟商品名称，下单时固定不受后台改变
	Num             string // 数量
	Price           string // 价格,下单时固定不受后台改变，结账时再检查是否改变
	DelayTime       string // 加钟时间(分钟)
	Sign            string // 加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并
	CreateTime      string // 创建时间(时间戳)
	UpdateTime      string // 更新时间(时间戳)
	DeleteTime      string // 删除时间(时间戳)
}

// saleOrderBuffetDelayProductColumns holds the columns for the table ttpos_sale_order_buffet_delay_product.
var saleOrderBuffetDelayProductColumns = SaleOrderBuffetDelayProductColumns{
	Id:              "id",
	Uuid:            "uuid",
	SaleOrderUuid:   "sale_order_uuid",
	BuffetDelayUuid: "buffet_delay_uuid",
	Name:            "name",
	Num:             "num",
	Price:           "price",
	DelayTime:       "delay_time",
	Sign:            "sign",
	CreateTime:      "create_time",
	UpdateTime:      "update_time",
	DeleteTime:      "delete_time",
}

// NewSaleOrderBuffetDelayProductDao creates and returns a new DAO object for table data access.
func NewSaleOrderBuffetDelayProductDao(handlers ...gdb.ModelHandler) *SaleOrderBuffetDelayProductDao {
	return &SaleOrderBuffetDelayProductDao{
		group:    "default",
		table:    "ttpos_sale_order_buffet_delay_product",
		columns:  saleOrderBuffetDelayProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderBuffetDelayProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderBuffetDelayProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderBuffetDelayProductDao) Columns() SaleOrderBuffetDelayProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderBuffetDelayProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderBuffetDelayProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderBuffetDelayProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
