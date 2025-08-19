// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCardDao is the data access object for the table ttpos_member_card.
type MemberCardDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MemberCardColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MemberCardColumns defines and stores column names for the table ttpos_member_card.
type MemberCardColumns struct {
	Id           string // 自增ID
	Uuid         string // 会员卡ID
	CardTypeUuid string // 会员卡类型ID
	MemberUuid   string // 会员ID
	ExpireTime   string // 截止日期(时间戳)
	Discount     string // 折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段
	CreateTime   string // 创建时间(时间戳),领取时间
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// memberCardColumns holds the columns for the table ttpos_member_card.
var memberCardColumns = MemberCardColumns{
	Id:           "id",
	Uuid:         "uuid",
	CardTypeUuid: "card_type_uuid",
	MemberUuid:   "member_uuid",
	ExpireTime:   "expire_time",
	Discount:     "discount",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewMemberCardDao creates and returns a new DAO object for table data access.
func NewMemberCardDao(handlers ...gdb.ModelHandler) *MemberCardDao {
	return &MemberCardDao{
		group:    "default",
		table:    "ttpos_member_card",
		columns:  memberCardColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberCardDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberCardDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberCardDao) Columns() MemberCardColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberCardDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberCardDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberCardDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
