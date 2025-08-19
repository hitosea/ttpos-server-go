// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingCouponRecordDao is the data access object for the table ttpos_marketing_coupon_record.
type MarketingCouponRecordDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  MarketingCouponRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// MarketingCouponRecordColumns defines and stores column names for the table ttpos_marketing_coupon_record.
type MarketingCouponRecordColumns struct {
	Id           string //
	Uuid         string // 优惠券记录唯一ID
	CouponUuid   string // 优惠券唯一ID
	ActivityUuid string // 活动uuid
	SerialNo     string // 记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用
	Type         string // 记录类型：1-首次添加、2-调整添加、3-调整扣减
	Count        string // 变动数量
	LeftCount    string // 剩余有效张数
	CreateTime   string // 创建时间
	UpdateTime   string // 更新时间
	DeleteTime   string // 删除时间
	MemberUuid   string // 会员uuid
}

// marketingCouponRecordColumns holds the columns for the table ttpos_marketing_coupon_record.
var marketingCouponRecordColumns = MarketingCouponRecordColumns{
	Id:           "id",
	Uuid:         "uuid",
	CouponUuid:   "coupon_uuid",
	ActivityUuid: "activity_uuid",
	SerialNo:     "serial_no",
	Type:         "type",
	Count:        "count",
	LeftCount:    "left_count",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
	MemberUuid:   "member_uuid",
}

// NewMarketingCouponRecordDao creates and returns a new DAO object for table data access.
func NewMarketingCouponRecordDao(handlers ...gdb.ModelHandler) *MarketingCouponRecordDao {
	return &MarketingCouponRecordDao{
		group:    "default",
		table:    "ttpos_marketing_coupon_record",
		columns:  marketingCouponRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingCouponRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingCouponRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingCouponRecordDao) Columns() MarketingCouponRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingCouponRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingCouponRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingCouponRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
