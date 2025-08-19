// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductBomDao is the data access object for the table ttpos_product_bom.
type ProductBomDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ProductBomColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ProductBomColumns defines and stores column names for the table ttpos_product_bom.
type ProductBomColumns struct {
	Id                 string // 自增ID
	Uuid               string // 商品BOM ID
	PurchasePrice      string // 采购单价
	Price              string // 价格
	Name               string // 商品名称或小料名称(不用于业务显示)
	ProductFlavorUuid  string // 商品规格ID(仅商品使用)
	ProductSauceUuid   string // 商品小料ID(仅小料使用)
	ProductPackageUuid string // 商品包ID
	StockNum           string // 库存数量
	BarcodeValue       string // 条形码值
	IsDefaultSelect    string // 是否默认选择, 0-否 1-是
	Status             string // 状态, 0-下架 1-上架. 同步商品包的状态
	IsSoldOut          string // 是否沽清, 0-否 1-是
	ActualSaleNum      string // 实际销量。每次卖出时,实际销量增加
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// productBomColumns holds the columns for the table ttpos_product_bom.
var productBomColumns = ProductBomColumns{
	Id:                 "id",
	Uuid:               "uuid",
	PurchasePrice:      "purchase_price",
	Price:              "price",
	Name:               "name",
	ProductFlavorUuid:  "product_flavor_uuid",
	ProductSauceUuid:   "product_sauce_uuid",
	ProductPackageUuid: "product_package_uuid",
	StockNum:           "stock_num",
	BarcodeValue:       "barcode_value",
	IsDefaultSelect:    "is_default_select",
	Status:             "status",
	IsSoldOut:          "is_sold_out",
	ActualSaleNum:      "actual_sale_num",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewProductBomDao creates and returns a new DAO object for table data access.
func NewProductBomDao(handlers ...gdb.ModelHandler) *ProductBomDao {
	return &ProductBomDao{
		group:    "default",
		table:    "ttpos_product_bom",
		columns:  productBomColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductBomDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductBomDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductBomDao) Columns() ProductBomColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductBomDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductBomDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductBomDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
