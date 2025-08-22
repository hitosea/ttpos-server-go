// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ShopCashierDao is the data access object for the table erp_shop_cashier.
type ShopCashierDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ShopCashierColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ShopCashierColumns defines and stores column names for the table erp_shop_cashier.
type ShopCashierColumns struct {
	Id           string //
	ShopUuid     string // 商店UUID
	AdminUuid    string // 商店管理员UUID
	CashierEmail string // 收银员邮箱
	ApiKey       string //
	ApiSecret    string //
}

// shopCashierColumns holds the columns for the table erp_shop_cashier.
var shopCashierColumns = ShopCashierColumns{
	Id:           "id",
	ShopUuid:     "shop_uuid",
	AdminUuid:    "admin_uuid",
	CashierEmail: "cashier_email",
	ApiKey:       "api_key",
	ApiSecret:    "api_secret",
}

// NewShopCashierDao creates and returns a new DAO object for table data access.
func NewShopCashierDao(handlers ...gdb.ModelHandler) *ShopCashierDao {
	return &ShopCashierDao{
		group:    "default",
		table:    "erp_shop_cashier",
		columns:  shopCashierColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ShopCashierDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ShopCashierDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ShopCashierDao) Columns() ShopCashierColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ShopCashierDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ShopCashierDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ShopCashierDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
