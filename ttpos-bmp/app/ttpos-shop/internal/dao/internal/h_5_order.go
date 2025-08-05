// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// H5OrderDao is the data access object for the table ttpos_h5_order.
type H5OrderDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  H5OrderColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// H5OrderColumns defines and stores column names for the table ttpos_h5_order.
type H5OrderColumns struct {
	Id                     string // 自增ID
	Uuid                   string // 扫码订单ID
	DeskUuid               string // 桌台uuid
	DeskNo                 string // 桌台编号
	Status                 string // 状态, 0-未下单 1-未接单 2-已接单 3-已拒单
	IsAutoAccept           string // 是否自动接单, 0-否 1-是
	IsBuffet               string // 是否是自助餐, 0-非自助餐 1-自助餐
	MemberDiscountRate     string // 会员折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	MemberCardDiscountRate string // 会员卡折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	CustomDiscountRate     string // 自定义折扣率(0-100%).接单和拒单后从sale_order_product表获取，不再改变
	ProductTotalPrice      string // 商品总价。接单和拒单后从sale_order_product表获取，不再改变
	TotalAmount            string // 订单金额. 订单金额=商品总价*折扣率。接单和拒单后从sale_order_product表获取，不再改变
	StaffUuid              string // 接单或拒单员工ID
	HandleTime             string // 接单或拒单时间(时间戳)
	OrderTime              string // 下单时间(时间戳)
	SaleOrderUuid          string // 销售订单uuid
	SaleBillUuid           string // 销售账单uuid
	CreateTime             string // 创建时间(时间戳)，扫码下单时间
	UpdateTime             string // 更新时间(时间戳)
	DeleteTime             string // 删除时间(时间戳)
}

// h5OrderColumns holds the columns for the table ttpos_h5_order.
var h5OrderColumns = H5OrderColumns{
	Id:                     "id",
	Uuid:                   "uuid",
	DeskUuid:               "desk_uuid",
	DeskNo:                 "desk_no",
	Status:                 "status",
	IsAutoAccept:           "is_auto_accept",
	IsBuffet:               "is_buffet",
	MemberDiscountRate:     "member_discount_rate",
	MemberCardDiscountRate: "member_card_discount_rate",
	CustomDiscountRate:     "custom_discount_rate",
	ProductTotalPrice:      "product_total_price",
	TotalAmount:            "total_amount",
	StaffUuid:              "staff_uuid",
	HandleTime:             "handle_time",
	OrderTime:              "order_time",
	SaleOrderUuid:          "sale_order_uuid",
	SaleBillUuid:           "sale_bill_uuid",
	CreateTime:             "create_time",
	UpdateTime:             "update_time",
	DeleteTime:             "delete_time",
}

// NewH5OrderDao creates and returns a new DAO object for table data access.
func NewH5OrderDao(handlers ...gdb.ModelHandler) *H5OrderDao {
	return &H5OrderDao{
		group:    "default",
		table:    "ttpos_h5_order",
		columns:  h5OrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *H5OrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *H5OrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *H5OrderDao) Columns() H5OrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *H5OrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *H5OrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *H5OrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
