// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MaterialDao is the data access object for the table ttpos_material.
type MaterialDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MaterialColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MaterialColumns defines and stores column names for the table ttpos_material.
type MaterialColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 原料ID
	Name                  string // 原料名称
	MultiLanguageNameUuid string // 多语言名称ID
	CategoryUuid          string // 类别ID
	SupplierUuid          string // 供应商ID
	ImageUuid             string // 图片ID
	ImageName             string // 图片名称
	UnitUuid              string // 单位ID
	Price                 string // 采购单价
	StockNum              string // 库存数量
	BarcodeValue          string // 条形码值
	Status                string // 状态, 1-上架 0-下架
	ActualSaleNum         string // 实际销量。每次卖出时,实际销量增加
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// materialColumns holds the columns for the table ttpos_material.
var materialColumns = MaterialColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	CategoryUuid:          "category_uuid",
	SupplierUuid:          "supplier_uuid",
	ImageUuid:             "image_uuid",
	ImageName:             "image_name",
	UnitUuid:              "unit_uuid",
	Price:                 "price",
	StockNum:              "stock_num",
	BarcodeValue:          "barcode_value",
	Status:                "status",
	ActualSaleNum:         "actual_sale_num",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewMaterialDao creates and returns a new DAO object for table data access.
func NewMaterialDao(handlers ...gdb.ModelHandler) *MaterialDao {
	return &MaterialDao{
		group:    "default",
		table:    "ttpos_material",
		columns:  materialColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MaterialDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MaterialDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MaterialDao) Columns() MaterialColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MaterialDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MaterialDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MaterialDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
