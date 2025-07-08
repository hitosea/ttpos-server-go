// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCouponUseRecordDao is the data access object for the table ttpos_member_coupon_use_record.
type MemberCouponUseRecordDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  MemberCouponUseRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// MemberCouponUseRecordColumns defines and stores column names for the table ttpos_member_coupon_use_record.
type MemberCouponUseRecordColumns struct {
	Id             string //
	Uuid           string // 唯一ID
	MemberUuid     string // 会员uuid
	CouponUuid     string // 优惠券uuid
	UseOrderUuid   string // 优惠券使用订单uuid
	UseOrderAmount string // 优惠券使用订单金额
	CreateTime     string // 创建时间
	UpdateTime     string // 更新时间
	DeleteTime     string // 删除时间
}

// memberCouponUseRecordColumns holds the columns for the table ttpos_member_coupon_use_record.
var memberCouponUseRecordColumns = MemberCouponUseRecordColumns{
	Id:             "id",
	Uuid:           "uuid",
	MemberUuid:     "member_uuid",
	CouponUuid:     "coupon_uuid",
	UseOrderUuid:   "use_order_uuid",
	UseOrderAmount: "use_order_amount",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewMemberCouponUseRecordDao creates and returns a new DAO object for table data access.
func NewMemberCouponUseRecordDao(handlers ...gdb.ModelHandler) *MemberCouponUseRecordDao {
	return &MemberCouponUseRecordDao{
		group:    "default",
		table:    "ttpos_member_coupon_use_record",
		columns:  memberCouponUseRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberCouponUseRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberCouponUseRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberCouponUseRecordDao) Columns() MemberCouponUseRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberCouponUseRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberCouponUseRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberCouponUseRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
