// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrderProductDao is the data access object for the table ttpos_production_order_product.
type ProductionOrderProductDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  ProductionOrderProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// ProductionOrderProductColumns defines and stores column names for the table ttpos_production_order_product.
type ProductionOrderProductColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 生产订单商品ID
	Name                  string // 名称
	Num                   string // 商品数量
	FlavorName            string // 规格名称,不随后台改变
	ProductAttributeNames string // 商品属性名称,多个属性名用逗号分隔,不随后台改变
	ProductSaucesNames    string // 商品加料名称,多个加料名用逗号分隔,不随后台改变
	Status                string // 状态, 0-待制作 1-制作中 2-已完成 3-已退菜
	Remark                string // 商品备注
	HasMaterial           string // 是否无原料, 0-无原料,商品没有关联原料 1-有原料
	SaleBillUuid          string // 销售账单ID
	ProductPackageUuid    string // 商品包ID
	SaleOrderProductUuid  string // 销售订单商品ID
	ProductionOrderUuid   string // 生产订单ID
	FirstCategoryUuid     string // 一级分类ID
	FinishedTime          string // 完成时间(时间戳)
	CreateTime            string // 创建时间(时间戳),送厨时间
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// productionOrderProductColumns holds the columns for the table ttpos_production_order_product.
var productionOrderProductColumns = ProductionOrderProductColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	Num:                   "num",
	FlavorName:            "flavor_name",
	ProductAttributeNames: "product_attribute_names",
	ProductSaucesNames:    "product_sauces_names",
	Status:                "status",
	Remark:                "remark",
	HasMaterial:           "has_material",
	SaleBillUuid:          "sale_bill_uuid",
	ProductPackageUuid:    "product_package_uuid",
	SaleOrderProductUuid:  "sale_order_product_uuid",
	ProductionOrderUuid:   "production_order_uuid",
	FirstCategoryUuid:     "first_category_uuid",
	FinishedTime:          "finished_time",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewProductionOrderProductDao creates and returns a new DAO object for table data access.
func NewProductionOrderProductDao(handlers ...gdb.ModelHandler) *ProductionOrderProductDao {
	return &ProductionOrderProductDao{
		group:    "default",
		table:    "ttpos_production_order_product",
		columns:  productionOrderProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductionOrderProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductionOrderProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductionOrderProductDao) Columns() ProductionOrderProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductionOrderProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductionOrderProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductionOrderProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
