// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetProductDao is the data access object for the table ttpos_buffet_product.
type BuffetProductDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  BuffetProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// BuffetProductColumns defines and stores column names for the table ttpos_buffet_product.
type BuffetProductColumns struct {
	Id                 string // 自增ID
	Uuid               string // 自助餐商品ID
	BuffetPackageUuid  string // 自助餐套餐ID
	ProductPackageUuid string // 商品包ID
	IsShowCashier      string // 是否在收银台显示, 0-否 1-是
	IsShowTablet       string // 是否在平板显示, 0-否 1-是
	IsShowKitchen      string // 是否在厨房显示, 0-否 1-是
	IsShowAssistant    string // 是否在助手显示, 0-否 1-是
	Limit              string // 限购数量
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// buffetProductColumns holds the columns for the table ttpos_buffet_product.
var buffetProductColumns = BuffetProductColumns{
	Id:                 "id",
	Uuid:               "uuid",
	BuffetPackageUuid:  "buffet_package_uuid",
	ProductPackageUuid: "product_package_uuid",
	IsShowCashier:      "is_show_cashier",
	IsShowTablet:       "is_show_tablet",
	IsShowKitchen:      "is_show_kitchen",
	IsShowAssistant:    "is_show_assistant",
	Limit:              "limit",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewBuffetProductDao creates and returns a new DAO object for table data access.
func NewBuffetProductDao(handlers ...gdb.ModelHandler) *BuffetProductDao {
	return &BuffetProductDao{
		group:    "default",
		table:    "ttpos_buffet_product",
		columns:  buffetProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BuffetProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BuffetProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BuffetProductDao) Columns() BuffetProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BuffetProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BuffetProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BuffetProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
