// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCardTypeDao is the data access object for the table ttpos_member_card_type.
type MemberCardTypeDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MemberCardTypeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MemberCardTypeColumns defines and stores column names for the table ttpos_member_card_type.
type MemberCardTypeColumns struct {
	Id           string // 自增ID
	Uuid         string // 会员卡类型ID
	Name         string // 会员卡类型名称
	Expire       string // 有效期限,单位:月, 0为永久有效
	Price        string // 价格
	Discount     string // 折扣,单位%
	Sort         string // 排序
	Status       string // 状态, 0-开启 1-关闭
	OpenPoint    string // 开卡赠送积分,0-否 1-是
	OpenPointNum string // 开卡赠送积分数
	OpenMoney    string // 开卡赠送余额,0-否 1-是
	OpenMoneyNum string // 开卡赠送余额数
	Describe     string // 使用须知
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// memberCardTypeColumns holds the columns for the table ttpos_member_card_type.
var memberCardTypeColumns = MemberCardTypeColumns{
	Id:           "id",
	Uuid:         "uuid",
	Name:         "name",
	Expire:       "expire",
	Price:        "price",
	Discount:     "discount",
	Sort:         "sort",
	Status:       "status",
	OpenPoint:    "open_point",
	OpenPointNum: "open_point_num",
	OpenMoney:    "open_money",
	OpenMoneyNum: "open_money_num",
	Describe:     "describe",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewMemberCardTypeDao creates and returns a new DAO object for table data access.
func NewMemberCardTypeDao(handlers ...gdb.ModelHandler) *MemberCardTypeDao {
	return &MemberCardTypeDao{
		group:    "default",
		table:    "ttpos_member_card_type",
		columns:  memberCardTypeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberCardTypeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberCardTypeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberCardTypeDao) Columns() MemberCardTypeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberCardTypeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberCardTypeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberCardTypeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
