// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetPackageDao is the data access object for the table ttpos_buffet_package.
type BuffetPackageDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  BuffetPackageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// BuffetPackageColumns defines and stores column names for the table ttpos_buffet_package.
type BuffetPackageColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 自助餐套餐ID
	Name                  string // 自助餐套餐名称
	MultiLanguageNameUuid string // 多语言名称ID
	Sort                  string // 排序顺序
	TaxUuid               string // 税收ID
	IsLimitTime           string // 是否限时, 0-否 1-是
	LimitTime             string // 限时时间(分钟)
	CanCombined           string // 是否可合并, 0-否 1-是
	NonOrderingTime       string // 平板不可下单时间(分钟)
	ReminderOrderTime     string // 平板提醒不可下单时间(分钟)
	Status                string // 状态 0-禁用 1-启用
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// buffetPackageColumns holds the columns for the table ttpos_buffet_package.
var buffetPackageColumns = BuffetPackageColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	Sort:                  "sort",
	TaxUuid:               "tax_uuid",
	IsLimitTime:           "is_limit_time",
	LimitTime:             "limit_time",
	CanCombined:           "can_combined",
	NonOrderingTime:       "non_ordering_time",
	ReminderOrderTime:     "reminder_order_time",
	Status:                "status",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewBuffetPackageDao creates and returns a new DAO object for table data access.
func NewBuffetPackageDao(handlers ...gdb.ModelHandler) *BuffetPackageDao {
	return &BuffetPackageDao{
		group:    "default",
		table:    "ttpos_buffet_package",
		columns:  buffetPackageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BuffetPackageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BuffetPackageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BuffetPackageDao) Columns() BuffetPackageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BuffetPackageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BuffetPackageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BuffetPackageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
