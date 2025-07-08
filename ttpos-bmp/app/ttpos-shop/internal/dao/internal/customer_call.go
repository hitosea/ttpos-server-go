// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CustomerCallDao is the data access object for the table ttpos_customer_call.
type CustomerCallDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  CustomerCallColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// CustomerCallColumns defines and stores column names for the table ttpos_customer_call.
type CustomerCallColumns struct {
	Id         string // 自增ID
	Uuid       string // 客户呼叫记录ID
	DeskUuid   string // 桌台ID
	DeskNo     string // 桌台编号,不随后台改变
	CallType   string // 呼叫类型(1服务员,2结账)
	Status     string // 状态,0-unhandled未处理 1-handled已处理
	IsSend     string // 消息发送状态 0-否 1-是
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// customerCallColumns holds the columns for the table ttpos_customer_call.
var customerCallColumns = CustomerCallColumns{
	Id:         "id",
	Uuid:       "uuid",
	DeskUuid:   "desk_uuid",
	DeskNo:     "desk_no",
	CallType:   "call_type",
	Status:     "status",
	IsSend:     "is_send",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewCustomerCallDao creates and returns a new DAO object for table data access.
func NewCustomerCallDao(handlers ...gdb.ModelHandler) *CustomerCallDao {
	return &CustomerCallDao{
		group:    "default",
		table:    "ttpos_customer_call",
		columns:  customerCallColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CustomerCallDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CustomerCallDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CustomerCallDao) Columns() CustomerCallColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CustomerCallDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CustomerCallDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CustomerCallDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
