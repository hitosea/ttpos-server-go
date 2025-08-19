// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingCouponDao is the data access object for the table ttpos_marketing_coupon.
type MarketingCouponDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  MarketingCouponColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// MarketingCouponColumns defines and stores column names for the table ttpos_marketing_coupon.
type MarketingCouponColumns struct {
	Id             string //
	Uuid           string // 优惠券唯一ID
	Name           string // 优惠券名称
	Sort           string // 排序, 1-99
	Type           string // 优惠券类型: deduction - 抵扣券
	DeductionType  string // 抵扣类型: taxed - 税后抵扣
	Amount         string // 优惠券金额
	Count          string // 优惠券数量, 最大999999
	DayStartTime   string // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     string // 每日适用时段结束时间, hh:mm 格式
	Requirement    string // 获得优惠券所需条件: none - 都可以获取; marketing - 营销活动
	ValidStartTime string // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   string // 优惠券有效结束时间, requirement = none 时有效
	ValidDays      string // 领取优惠券后n天内有效, requirement = marketing 时有效
	CreateTime     string // 创建时间
	UpdateTime     string // 更新时间
	DeleteTime     string // 删除时间
}

// marketingCouponColumns holds the columns for the table ttpos_marketing_coupon.
var marketingCouponColumns = MarketingCouponColumns{
	Id:             "id",
	Uuid:           "uuid",
	Name:           "name",
	Sort:           "sort",
	Type:           "type",
	DeductionType:  "deduction_type",
	Amount:         "amount",
	Count:          "count",
	DayStartTime:   "day_start_time",
	DayEndTime:     "day_end_time",
	Requirement:    "requirement",
	ValidStartTime: "valid_start_time",
	ValidEndTime:   "valid_end_time",
	ValidDays:      "valid_days",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewMarketingCouponDao creates and returns a new DAO object for table data access.
func NewMarketingCouponDao(handlers ...gdb.ModelHandler) *MarketingCouponDao {
	return &MarketingCouponDao{
		group:    "default",
		table:    "ttpos_marketing_coupon",
		columns:  marketingCouponColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingCouponDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingCouponDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingCouponDao) Columns() MarketingCouponColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingCouponDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingCouponDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingCouponDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
