// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductAttributeDao is the data access object for the table ttpos_product_attribute.
type ProductAttributeDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  ProductAttributeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// ProductAttributeColumns defines and stores column names for the table ttpos_product_attribute.
type ProductAttributeColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 商品属性ID
	Name                  string // 名称
	MultiLanguageNameUuid string // 多语言名称ID
	AttributeGroupUuid    string // 属性组ID
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// productAttributeColumns holds the columns for the table ttpos_product_attribute.
var productAttributeColumns = ProductAttributeColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	AttributeGroupUuid:    "attribute_group_uuid",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewProductAttributeDao creates and returns a new DAO object for table data access.
func NewProductAttributeDao(handlers ...gdb.ModelHandler) *ProductAttributeDao {
	return &ProductAttributeDao{
		group:    "default",
		table:    "ttpos_product_attribute",
		columns:  productAttributeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductAttributeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductAttributeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductAttributeDao) Columns() ProductAttributeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductAttributeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductAttributeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductAttributeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
