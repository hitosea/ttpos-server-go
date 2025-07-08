// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsMemberDao is the data access object for the table ttpos_statistics_member.
type StatisticsMemberDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  StatisticsMemberColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// StatisticsMemberColumns defines and stores column names for the table ttpos_statistics_member.
type StatisticsMemberColumns struct {
	Id                      string // 自增ID
	Uuid                    string // UUID
	MemberRechargeOrderUuid string // 会员充值订单uuid
	DutyNo                  string // 当班编号
	RechargeAmount          string // 充值金额
	GiveAmount              string // 赠送金额
	GivePoint               string // 赠送积分
	PaymentAmount           string // 支付金额
	PaymentFee              string // 支付手续费
	RefundAmount            string // 退款金额
	RefundFee               string // 退款手续费
	CompleteTime            string // 完成时间
	RefundTime              string // 完成时间
	CreateTime              string // 创建时间
	UpdateTime              string // 更新时间
	DeleteTime              string // 删除时间
}

// statisticsMemberColumns holds the columns for the table ttpos_statistics_member.
var statisticsMemberColumns = StatisticsMemberColumns{
	Id:                      "id",
	Uuid:                    "uuid",
	MemberRechargeOrderUuid: "member_recharge_order_uuid",
	DutyNo:                  "duty_no",
	RechargeAmount:          "recharge_amount",
	GiveAmount:              "give_amount",
	GivePoint:               "give_point",
	PaymentAmount:           "payment_amount",
	PaymentFee:              "payment_fee",
	RefundAmount:            "refund_amount",
	RefundFee:               "refund_fee",
	CompleteTime:            "complete_time",
	RefundTime:              "refund_time",
	CreateTime:              "create_time",
	UpdateTime:              "update_time",
	DeleteTime:              "delete_time",
}

// NewStatisticsMemberDao creates and returns a new DAO object for table data access.
func NewStatisticsMemberDao(handlers ...gdb.ModelHandler) *StatisticsMemberDao {
	return &StatisticsMemberDao{
		group:    "default",
		table:    "ttpos_statistics_member",
		columns:  statisticsMemberColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StatisticsMemberDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StatisticsMemberDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StatisticsMemberDao) Columns() StatisticsMemberColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StatisticsMemberDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StatisticsMemberDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StatisticsMemberDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
