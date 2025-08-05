// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberBalanceLogDao is the data access object for the table ttpos_member_balance_log.
type MemberBalanceLogDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MemberBalanceLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MemberBalanceLogColumns defines and stores column names for the table ttpos_member_balance_log.
type MemberBalanceLogColumns struct {
	Id            string // 自增ID
	Uuid          string // 余额变动记录ID
	MemberUuid    string // 会员ID
	Scene         string // 场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减
	SaleOrderUuid string // 销售订单ID
	Money         string // 变动金额,负数:减余额 整数:加余额
	GiftMoney     string // 变动赠送金额
	Describe      string // 变动描述
	Processed     string // 是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额
	RelatedUuid   string // 关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// memberBalanceLogColumns holds the columns for the table ttpos_member_balance_log.
var memberBalanceLogColumns = MemberBalanceLogColumns{
	Id:            "id",
	Uuid:          "uuid",
	MemberUuid:    "member_uuid",
	Scene:         "scene",
	SaleOrderUuid: "sale_order_uuid",
	Money:         "money",
	GiftMoney:     "gift_money",
	Describe:      "describe",
	Processed:     "processed",
	RelatedUuid:   "related_uuid",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewMemberBalanceLogDao creates and returns a new DAO object for table data access.
func NewMemberBalanceLogDao(handlers ...gdb.ModelHandler) *MemberBalanceLogDao {
	return &MemberBalanceLogDao{
		group:    "default",
		table:    "ttpos_member_balance_log",
		columns:  memberBalanceLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberBalanceLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberBalanceLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberBalanceLogDao) Columns() MemberBalanceLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberBalanceLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberBalanceLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberBalanceLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
