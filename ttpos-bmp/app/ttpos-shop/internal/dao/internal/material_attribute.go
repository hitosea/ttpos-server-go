// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MaterialAttributeDao is the data access object for the table ttpos_material_attribute.
type MaterialAttributeDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  MaterialAttributeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// MaterialAttributeColumns defines and stores column names for the table ttpos_material_attribute.
type MaterialAttributeColumns struct {
	Id                           string // 自增ID
	Uuid                         string // 原料属性ID
	MaterialUuid                 string // 原料ID
	HistoricalPurchaseQuantity   string // 历史采购数量
	HistoricalLossReportQuantity string // 历史报损数量
	HistoricalSaleQuantity       string // 历史销售数量
	CreateTime                   string // 创建时间(时间戳)
	UpdateTime                   string // 更新时间(时间戳)
	DeleteTime                   string // 删除时间(时间戳)
}

// materialAttributeColumns holds the columns for the table ttpos_material_attribute.
var materialAttributeColumns = MaterialAttributeColumns{
	Id:                           "id",
	Uuid:                         "uuid",
	MaterialUuid:                 "material_uuid",
	HistoricalPurchaseQuantity:   "historical_purchase_quantity",
	HistoricalLossReportQuantity: "historical_loss_report_quantity",
	HistoricalSaleQuantity:       "historical_sale_quantity",
	CreateTime:                   "create_time",
	UpdateTime:                   "update_time",
	DeleteTime:                   "delete_time",
}

// NewMaterialAttributeDao creates and returns a new DAO object for table data access.
func NewMaterialAttributeDao(handlers ...gdb.ModelHandler) *MaterialAttributeDao {
	return &MaterialAttributeDao{
		group:    "default",
		table:    "ttpos_material_attribute",
		columns:  materialAttributeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MaterialAttributeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MaterialAttributeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MaterialAttributeDao) Columns() MaterialAttributeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MaterialAttributeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MaterialAttributeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MaterialAttributeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
