// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberRechargeOrderDao is the data access object for the table ttpos_member_recharge_order.
type MemberRechargeOrderDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  MemberRechargeOrderColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// MemberRechargeOrderColumns defines and stores column names for the table ttpos_member_recharge_order.
type MemberRechargeOrderColumns struct {
	Id               string // 自增ID
	Uuid             string // 充值订单ID
	OrderNo          string // 充值订单编号
	DutyNo           string // 当班编号
	Status           string // 状态,0-pending待支付 1-paid已支付 2-canceled已取消
	Amount           string // 交易金额=充值金额+手续费
	RefundMoney      string // 退款金额，不大于amount
	ChargeDue        string // 找零
	RechargeAmount   string // 充值金额
	RefundAmount     string // 退款充值金额，不大于recharge_amount
	GiftAmount       string // 赠送金额
	GiftPoint        string // 赠送积分
	MemberUuid       string // 会员ID
	StaffUuid        string // 员工ID
	PaymentTime      string // 支付时间(时间戳)
	Balance          string // 充值前会员余额
	BalanceRecharged string // 充值后会员余额
	CreateTime       string // 创建时间(时间戳)
	UpdateTime       string // 更新时间(时间戳)
	DeleteTime       string // 删除时间(时间戳)
}

// memberRechargeOrderColumns holds the columns for the table ttpos_member_recharge_order.
var memberRechargeOrderColumns = MemberRechargeOrderColumns{
	Id:               "id",
	Uuid:             "uuid",
	OrderNo:          "order_no",
	DutyNo:           "duty_no",
	Status:           "status",
	Amount:           "amount",
	RefundMoney:      "refund_money",
	ChargeDue:        "charge_due",
	RechargeAmount:   "recharge_amount",
	RefundAmount:     "refund_amount",
	GiftAmount:       "gift_amount",
	GiftPoint:        "gift_point",
	MemberUuid:       "member_uuid",
	StaffUuid:        "staff_uuid",
	PaymentTime:      "payment_time",
	Balance:          "balance",
	BalanceRecharged: "balance_recharged",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewMemberRechargeOrderDao creates and returns a new DAO object for table data access.
func NewMemberRechargeOrderDao(handlers ...gdb.ModelHandler) *MemberRechargeOrderDao {
	return &MemberRechargeOrderDao{
		group:    "default",
		table:    "ttpos_member_recharge_order",
		columns:  memberRechargeOrderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberRechargeOrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberRechargeOrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberRechargeOrderDao) Columns() MemberRechargeOrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberRechargeOrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberRechargeOrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberRechargeOrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
