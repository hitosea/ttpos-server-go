// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RelatedMaterialDao is the data access object for the table ttpos_related_material.
type RelatedMaterialDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  RelatedMaterialColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// RelatedMaterialColumns defines and stores column names for the table ttpos_related_material.
type RelatedMaterialColumns struct {
	Id           string // 自增ID
	Uuid         string // 关联材料ID
	RelatedUuid  string // 物料清单BOM的ID
	MaterialUuid string // 原料ID
	Num          string // 材料用量,可小数
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// relatedMaterialColumns holds the columns for the table ttpos_related_material.
var relatedMaterialColumns = RelatedMaterialColumns{
	Id:           "id",
	Uuid:         "uuid",
	RelatedUuid:  "related_uuid",
	MaterialUuid: "material_uuid",
	Num:          "num",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewRelatedMaterialDao creates and returns a new DAO object for table data access.
func NewRelatedMaterialDao(handlers ...gdb.ModelHandler) *RelatedMaterialDao {
	return &RelatedMaterialDao{
		group:    "default",
		table:    "ttpos_related_material",
		columns:  relatedMaterialColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RelatedMaterialDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RelatedMaterialDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RelatedMaterialDao) Columns() RelatedMaterialColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RelatedMaterialDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RelatedMaterialDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RelatedMaterialDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
