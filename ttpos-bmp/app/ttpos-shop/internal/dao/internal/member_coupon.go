// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCouponDao is the data access object for the table ttpos_member_coupon.
type MemberCouponDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  MemberCouponColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// MemberCouponColumns defines and stores column names for the table ttpos_member_coupon.
type MemberCouponColumns struct {
	Id             string //
	Uuid           string // 唯一ID
	MemberUuid     string // 会员uuid
	CouponUuid     string // 优惠券uuid
	Name           string // 优惠券名称
	DeductionType  string // 抵扣类型: taxed - 税后抵扣
	Type           string // 优惠券类型: deduction - 抵扣券
	DayStartTime   string // 每日适用时段开始时间, hh:mm 格式
	DayEndTime     string // 每日适用时段结束时间, hh:mm 格式
	ValidStartTime string // 优惠券有效开始时间, requirement = none 时有效
	ValidEndTime   string // 优惠券有效结束时间, requirement = none 时有效
	Amount         string // 优惠券面值
	Status         string // 优惠券状态 0未使用 1已使用
	StartTime      string // 优惠券开始时间
	EndTime        string // 优惠券结束时间
	UseTime        string // 优惠券使用时间
	DeleteTime     string // 删除时间
	CreateTime     string // 创建时间
	UpdateTime     string // 更新时间
}

// memberCouponColumns holds the columns for the table ttpos_member_coupon.
var memberCouponColumns = MemberCouponColumns{
	Id:             "id",
	Uuid:           "uuid",
	MemberUuid:     "member_uuid",
	CouponUuid:     "coupon_uuid",
	Name:           "name",
	DeductionType:  "deduction_type",
	Type:           "type",
	DayStartTime:   "day_start_time",
	DayEndTime:     "day_end_time",
	ValidStartTime: "valid_start_time",
	ValidEndTime:   "valid_end_time",
	Amount:         "amount",
	Status:         "status",
	StartTime:      "start_time",
	EndTime:        "end_time",
	UseTime:        "use_time",
	DeleteTime:     "delete_time",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
}

// NewMemberCouponDao creates and returns a new DAO object for table data access.
func NewMemberCouponDao(handlers ...gdb.ModelHandler) *MemberCouponDao {
	return &MemberCouponDao{
		group:    "default",
		table:    "ttpos_member_coupon",
		columns:  memberCouponColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberCouponDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberCouponDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberCouponDao) Columns() MemberCouponColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberCouponDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberCouponDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberCouponDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
