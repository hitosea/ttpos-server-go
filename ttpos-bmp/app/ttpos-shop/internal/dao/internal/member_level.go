// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberLevelDao is the data access object for the table ttpos_member_level.
type MemberLevelDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MemberLevelColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MemberLevelColumns defines and stores column names for the table ttpos_member_level.
type MemberLevelColumns struct {
	Id           string // 自增ID
	Uuid         string // 会员等级ID
	Name         string // 等级名称
	OpenMoney    string // 是否开放累计消费额升级，0-否 1-是
	UpgradeMoney string // 升级条件，累计消费额
	OpenPoint    string // 是否开放累计积分升级，0-否 1-是
	UpgradePoint string // 升级条件，累计积分
	Discount     string // 等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8
	Priority     string // 等级权重，越大等级越高
	IsDefault    string // 是否默认, 1-是 0-否
	Remark       string // 备注
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// memberLevelColumns holds the columns for the table ttpos_member_level.
var memberLevelColumns = MemberLevelColumns{
	Id:           "id",
	Uuid:         "uuid",
	Name:         "name",
	OpenMoney:    "open_money",
	UpgradeMoney: "upgrade_money",
	OpenPoint:    "open_point",
	UpgradePoint: "upgrade_point",
	Discount:     "discount",
	Priority:     "priority",
	IsDefault:    "is_default",
	Remark:       "remark",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewMemberLevelDao creates and returns a new DAO object for table data access.
func NewMemberLevelDao(handlers ...gdb.ModelHandler) *MemberLevelDao {
	return &MemberLevelDao{
		group:    "default",
		table:    "ttpos_member_level",
		columns:  memberLevelColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberLevelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberLevelDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberLevelDao) Columns() MemberLevelColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberLevelDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberLevelDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberLevelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
