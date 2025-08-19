// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberDao is the data access object for the table ttpos_member.
type MemberDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MemberColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MemberColumns defines and stores column names for the table ttpos_member.
type MemberColumns struct {
	Id                           string // 自增ID
	Uuid                         string // 会员ID
	MemberNo                     string // 会员编号
	Nickname                     string // 昵称
	Gender                       string // 性别,0-女 1-男 2-未知
	Phone                        string // 电话号码
	Password                     string // 密码
	Birthday                     string // 生日,时间戳
	Point                        string // 积分
	FrozenPoint                  string // 冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分
	AccumulatedConsumptionAmount string // 累计消费金额
	ConsumptionCount             string // 消费次数
	Balance                      string // 余额
	FrozenBalance                string // 冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额
	GiftBalance                  string // 赠送账户余额
	FrozenGiftBalance            string // 冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额
	AccumulatedRechargeAmount    string // 累计充值金额
	MemberLevelUuid              string // 会员等级ID
	MemberCardUuid               string // 会员卡片ID
	CreateTime                   string // 创建时间(时间戳)
	UpdateTime                   string // 更新时间(时间戳)
	DeleteTime                   string // 删除时间(时间戳)
}

// memberColumns holds the columns for the table ttpos_member.
var memberColumns = MemberColumns{
	Id:                           "id",
	Uuid:                         "uuid",
	MemberNo:                     "member_no",
	Nickname:                     "nickname",
	Gender:                       "gender",
	Phone:                        "phone",
	Password:                     "password",
	Birthday:                     "birthday",
	Point:                        "point",
	FrozenPoint:                  "frozen_point",
	AccumulatedConsumptionAmount: "accumulated_consumption_amount",
	ConsumptionCount:             "consumption_count",
	Balance:                      "balance",
	FrozenBalance:                "frozen_balance",
	GiftBalance:                  "gift_balance",
	FrozenGiftBalance:            "frozen_gift_balance",
	AccumulatedRechargeAmount:    "accumulated_recharge_amount",
	MemberLevelUuid:              "member_level_uuid",
	MemberCardUuid:               "member_card_uuid",
	CreateTime:                   "create_time",
	UpdateTime:                   "update_time",
	DeleteTime:                   "delete_time",
}

// NewMemberDao creates and returns a new DAO object for table data access.
func NewMemberDao(handlers ...gdb.ModelHandler) *MemberDao {
	return &MemberDao{
		group:    "default",
		table:    "ttpos_member",
		columns:  memberColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberDao) Columns() MemberColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
