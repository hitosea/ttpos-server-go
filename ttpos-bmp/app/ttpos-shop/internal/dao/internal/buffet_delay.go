// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetDelayDao is the data access object for the table ttpos_buffet_delay.
type BuffetDelayDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BuffetDelayColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BuffetDelayColumns defines and stores column names for the table ttpos_buffet_delay.
type BuffetDelayColumns struct {
	Id         string // 自增ID
	Uuid       string // 自助餐加钟价格ID
	Name       string // 自助餐加钟商品名称
	DelayTime  string // 加钟时间(分钟)
	Price      string // 价格
	Status     string // 状态 0-禁用 1-启用
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// buffetDelayColumns holds the columns for the table ttpos_buffet_delay.
var buffetDelayColumns = BuffetDelayColumns{
	Id:         "id",
	Uuid:       "uuid",
	Name:       "name",
	DelayTime:  "delay_time",
	Price:      "price",
	Status:     "status",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewBuffetDelayDao creates and returns a new DAO object for table data access.
func NewBuffetDelayDao(handlers ...gdb.ModelHandler) *BuffetDelayDao {
	return &BuffetDelayDao{
		group:    "default",
		table:    "ttpos_buffet_delay",
		columns:  buffetDelayColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BuffetDelayDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BuffetDelayDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BuffetDelayDao) Columns() BuffetDelayColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BuffetDelayDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BuffetDelayDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BuffetDelayDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
