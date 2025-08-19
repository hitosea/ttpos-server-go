// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CashBoxDao is the data access object for the table ttpos_cash_box.
type CashBoxDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CashBoxColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CashBoxColumns defines and stores column names for the table ttpos_cash_box.
type CashBoxColumns struct {
	Id              string // 自增ID
	Uuid            string // 钱箱ID
	Name            string // 名称
	Balance         string // 钱箱余额
	FrozenBalance   string // 冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额
	PreviousBalance string // 上一班遗留备用金
	CashWithdrawal  string // 中途取出金额
	CashDeposit     string // 中途存入金额
	CreateTime      string // 创建时间(时间戳)
	UpdateTime      string // 更新时间(时间戳)
	DeleteTime      string // 删除时间(时间戳)
}

// cashBoxColumns holds the columns for the table ttpos_cash_box.
var cashBoxColumns = CashBoxColumns{
	Id:              "id",
	Uuid:            "uuid",
	Name:            "name",
	Balance:         "balance",
	FrozenBalance:   "frozen_balance",
	PreviousBalance: "previous_balance",
	CashWithdrawal:  "cash_withdrawal",
	CashDeposit:     "cash_deposit",
	CreateTime:      "create_time",
	UpdateTime:      "update_time",
	DeleteTime:      "delete_time",
}

// NewCashBoxDao creates and returns a new DAO object for table data access.
func NewCashBoxDao(handlers ...gdb.ModelHandler) *CashBoxDao {
	return &CashBoxDao{
		group:    "default",
		table:    "ttpos_cash_box",
		columns:  cashBoxColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CashBoxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CashBoxDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CashBoxDao) Columns() CashBoxColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CashBoxDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CashBoxDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CashBoxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
