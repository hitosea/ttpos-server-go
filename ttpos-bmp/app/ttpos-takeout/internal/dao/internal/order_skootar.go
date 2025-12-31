// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderSkootarDao is the data access object for the table takeout_order_skootar.
type OrderSkootarDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  OrderSkootarColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// OrderSkootarColumns defines and stores column names for the table takeout_order_skootar.
type OrderSkootarColumns struct {
	Id              string // 主键ID
	Uuid            string // 唯一标识
	OrderUuid       string // 关联主订单UUID (takeout_order.uuid)
	SkootarId       string // 骑手ID
	SkootarName     string // 骑手名称
	SkootarPhone    string // 骑手电话
	SkootarRating   string // 骑手评分
	SkootarImageUrl string // 骑手头像
	CreatedAt       string // 创建时间
	UpdatedAt       string // 更新时间
	DeletedAt       string // 软删除
}

// orderSkootarColumns holds the columns for the table takeout_order_skootar.
var orderSkootarColumns = OrderSkootarColumns{
	Id:              "id",
	Uuid:            "uuid",
	OrderUuid:       "order_uuid",
	SkootarId:       "skootar_id",
	SkootarName:     "skootar_name",
	SkootarPhone:    "skootar_phone",
	SkootarRating:   "skootar_rating",
	SkootarImageUrl: "skootar_image_url",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewOrderSkootarDao creates and returns a new DAO object for table data access.
func NewOrderSkootarDao(handlers ...gdb.ModelHandler) *OrderSkootarDao {
	return &OrderSkootarDao{
		group:    "default",
		table:    "takeout_order_skootar",
		columns:  orderSkootarColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OrderSkootarDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OrderSkootarDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OrderSkootarDao) Columns() OrderSkootarColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OrderSkootarDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OrderSkootarDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OrderSkootarDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
