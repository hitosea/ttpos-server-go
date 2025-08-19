// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductCategoryDao is the data access object for the table ttpos_product_category.
type ProductCategoryDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ProductCategoryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ProductCategoryColumns defines and stores column names for the table ttpos_product_category.
type ProductCategoryColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 商品类别ID
	Name                  string // 名称
	MultiLanguageNameUuid string // 多语言名称ID
	Status                string // 状态, 1-开启 0-关闭
	ParentUuid            string // 父级ID
	IsSpecial             string // 特殊分类, 1-是 0-否
	CategoryKey           string // 关键字
	Sort                  string // 排序
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// productCategoryColumns holds the columns for the table ttpos_product_category.
var productCategoryColumns = ProductCategoryColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	Status:                "status",
	ParentUuid:            "parent_uuid",
	IsSpecial:             "is_special",
	CategoryKey:           "category_key",
	Sort:                  "sort",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewProductCategoryDao creates and returns a new DAO object for table data access.
func NewProductCategoryDao(handlers ...gdb.ModelHandler) *ProductCategoryDao {
	return &ProductCategoryDao{
		group:    "default",
		table:    "ttpos_product_category",
		columns:  productCategoryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductCategoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductCategoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductCategoryDao) Columns() ProductCategoryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductCategoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductCategoryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductCategoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
