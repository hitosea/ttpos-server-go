// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductMustPlanDao is the data access object for the table ttpos_product_must_plan.
type ProductMustPlanDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ProductMustPlanColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ProductMustPlanColumns defines and stores column names for the table ttpos_product_must_plan.
type ProductMustPlanColumns struct {
	Id           string // 自增ID
	Uuid         string // 商品必选商品计划ID
	Name         string // 方案名称
	UseChannel   string // 使用渠道 10-点餐方式 20-桌台方式
	MustType     string // 必点类型 0-每笔订单必点1份 1-每人必点1份
	MustRule     string // 必点规则 0-固定商品 1-可选商品
	Status       string // 状态,1-开启 0-关闭
	AutoCart     string // 自动加入购物车 1-是 0-否
	AutoChange   string // 顾客可修改必点数量 1-是 0-否
	AutoCheck    string // 下单时检查必点商品 1-是 0-否
	AutoCheckout string // 结账时检查必点商品 1-是 0-否
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// productMustPlanColumns holds the columns for the table ttpos_product_must_plan.
var productMustPlanColumns = ProductMustPlanColumns{
	Id:           "id",
	Uuid:         "uuid",
	Name:         "name",
	UseChannel:   "use_channel",
	MustType:     "must_type",
	MustRule:     "must_rule",
	Status:       "status",
	AutoCart:     "auto_cart",
	AutoChange:   "auto_change",
	AutoCheck:    "auto_check",
	AutoCheckout: "auto_checkout",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewProductMustPlanDao creates and returns a new DAO object for table data access.
func NewProductMustPlanDao(handlers ...gdb.ModelHandler) *ProductMustPlanDao {
	return &ProductMustPlanDao{
		group:    "default",
		table:    "ttpos_product_must_plan",
		columns:  productMustPlanColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductMustPlanDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductMustPlanDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductMustPlanDao) Columns() ProductMustPlanColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductMustPlanDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductMustPlanDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductMustPlanDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
