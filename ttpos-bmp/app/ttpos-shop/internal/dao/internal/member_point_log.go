// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberPointLogDao is the data access object for the table ttpos_member_point_log.
type MemberPointLogDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MemberPointLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MemberPointLogColumns defines and stores column names for the table ttpos_member_point_log.
type MemberPointLogColumns struct {
	Id          string // 自增ID
	Uuid        string // 积分变动记录ID
	MemberUuid  string // 会员ID
	Scene       string // 场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减
	Value       string // 数值,负数:减积分 正数:加积分
	Describe    string // 变动描述
	RelatedUuid string // 关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额
	Processed   string // 是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分
	CreateTime  string // 创建时间(时间戳)
	UpdateTime  string // 更新时间(时间戳)
	DeleteTime  string // 删除时间(时间戳)
}

// memberPointLogColumns holds the columns for the table ttpos_member_point_log.
var memberPointLogColumns = MemberPointLogColumns{
	Id:          "id",
	Uuid:        "uuid",
	MemberUuid:  "member_uuid",
	Scene:       "scene",
	Value:       "value",
	Describe:    "describe",
	RelatedUuid: "related_uuid",
	Processed:   "processed",
	CreateTime:  "create_time",
	UpdateTime:  "update_time",
	DeleteTime:  "delete_time",
}

// NewMemberPointLogDao creates and returns a new DAO object for table data access.
func NewMemberPointLogDao(handlers ...gdb.ModelHandler) *MemberPointLogDao {
	return &MemberPointLogDao{
		group:    "default",
		table:    "ttpos_member_point_log",
		columns:  memberPointLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberPointLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberPointLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberPointLogDao) Columns() MemberPointLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberPointLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberPointLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberPointLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
