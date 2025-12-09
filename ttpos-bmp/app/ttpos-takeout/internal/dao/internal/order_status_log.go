// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderStatusLogDao is the data access object for the table takeout_order_status_log.
type OrderStatusLogDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  OrderStatusLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// OrderStatusLogColumns defines and stores column names for the table takeout_order_status_log.
type OrderStatusLogColumns struct {
	Id           string // 主键
	Uuid         string // 唯一ID
	OrderUuid    string // 关联订单UUID
	ProviderName string // 渠道: grab
	StatusBefore string // 变更前状态
	StatusAfter  string // 变更后状态
	ChangeSource string // 变更来源: WEBHOOK, API, SYSTEM
	DriverEta    string // 骑手预计到达时间(分钟)
	Remark       string // 备注或原因
	RawData      string // 原始JSON数据
	CreatedAt    string // 创建时间
}

// orderStatusLogColumns holds the columns for the table takeout_order_status_log.
var orderStatusLogColumns = OrderStatusLogColumns{
	Id:           "id",
	Uuid:         "uuid",
	OrderUuid:    "order_uuid",
	ProviderName: "provider_name",
	StatusBefore: "status_before",
	StatusAfter:  "status_after",
	ChangeSource: "change_source",
	DriverEta:    "driver_eta",
	Remark:       "remark",
	RawData:      "raw_data",
	CreatedAt:    "created_at",
}

// NewOrderStatusLogDao creates and returns a new DAO object for table data access.
func NewOrderStatusLogDao(handlers ...gdb.ModelHandler) *OrderStatusLogDao {
	return &OrderStatusLogDao{
		group:    "default",
		table:    "takeout_order_status_log",
		columns:  orderStatusLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OrderStatusLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OrderStatusLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OrderStatusLogDao) Columns() OrderStatusLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OrderStatusLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OrderStatusLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OrderStatusLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
