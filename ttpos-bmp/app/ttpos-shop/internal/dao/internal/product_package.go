// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductPackageDao is the data access object for the table ttpos_product_package.
type ProductPackageDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  ProductPackageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// ProductPackageColumns defines and stores column names for the table ttpos_product_package.
type ProductPackageColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 商品包ID
	Name                  string // 商品包名称
	MultiLanguageNameUuid string // 多语言名称ID
	ImageName             string // 图片名称
	ImageFileUuid         string // 图片ID
	DeductStockType       string // 库存计算方法, 0-付款减库存 1-下单减库存
	UnitUuid              string // 单位UUID
	DineTaxUuid           string // 堂食税UUID
	CategoryUuid          string // 类别UUID
	SpecialCategoryUuid   string // 特殊类别UUID
	TakeoutTaxUuid        string // 外卖税UUID
	PrinterTagUuid        string // 打印机标签UUID
	SupplierUuid          string // 供应商UUID
	Status                string // 状态,0-下架 1-上架
	IsShowCashier         string // 是否在收银设备显示, 0-否 1-是
	IsShowTablet          string // 是否在平板设备显示, 0-否 1-是
	IsShowKitchen         string // 是否在厨房设备显示, 0-否 1-是
	IsShowAssistant       string // 是否在助手设备显示, 0-否 1-是
	IsShowH5              string // 是否在H5设备显示, 0-否 1-是
	Sort                  string // 排序
	LimitNum              string // 限购数量
	SauceRequired         string // 是否必选小料, 0-否 1-是
	SauceMaxSelection     string // 小料最大选择数量
	Describe              string // 卖点描述
	OpenDiscount          string // 是否开启会员折扣, 0-否 1-是
	ActualSaleNum         string // 实际销量。每次卖出时,实际销量增加
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// productPackageColumns holds the columns for the table ttpos_product_package.
var productPackageColumns = ProductPackageColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	ImageName:             "image_name",
	ImageFileUuid:         "image_file_uuid",
	DeductStockType:       "deduct_stock_type",
	UnitUuid:              "unit_uuid",
	DineTaxUuid:           "dine_tax_uuid",
	CategoryUuid:          "category_uuid",
	SpecialCategoryUuid:   "special_category_uuid",
	TakeoutTaxUuid:        "takeout_tax_uuid",
	PrinterTagUuid:        "printer_tag_uuid",
	SupplierUuid:          "supplier_uuid",
	Status:                "status",
	IsShowCashier:         "is_show_cashier",
	IsShowTablet:          "is_show_tablet",
	IsShowKitchen:         "is_show_kitchen",
	IsShowAssistant:       "is_show_assistant",
	IsShowH5:              "is_show_h5",
	Sort:                  "sort",
	LimitNum:              "limit_num",
	SauceRequired:         "sauce_required",
	SauceMaxSelection:     "sauce_max_selection",
	Describe:              "describe",
	OpenDiscount:          "open_discount",
	ActualSaleNum:         "actual_sale_num",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewProductPackageDao creates and returns a new DAO object for table data access.
func NewProductPackageDao(handlers ...gdb.ModelHandler) *ProductPackageDao {
	return &ProductPackageDao{
		group:    "default",
		table:    "ttpos_product_package",
		columns:  productPackageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductPackageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductPackageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductPackageDao) Columns() ProductPackageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductPackageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductPackageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductPackageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
