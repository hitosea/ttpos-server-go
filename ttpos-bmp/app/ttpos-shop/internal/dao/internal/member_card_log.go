// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCardLogDao is the data access object for the table ttpos_member_card_log.
type MemberCardLogDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MemberCardLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MemberCardLogColumns defines and stores column names for the table ttpos_member_card_log.
type MemberCardLogColumns struct {
	Id                 string // 自增ID
	Uuid               string // 会员卡领取记录ID
	Price              string // 价格,会员卡价格,不随后台改变,记录领取时的价格
	Discount           string // 折扣,单位%,不随后台改变,记录领取时的折扣
	Expire             string // 有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限
	MemberName         string // 会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberPhone        string // 会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberNo           string // 会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberCardTypeName string // 会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段
	MemberCardTypeUuid string // 会员卡类型ID
	MemberUuid         string // 会员ID
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// memberCardLogColumns holds the columns for the table ttpos_member_card_log.
var memberCardLogColumns = MemberCardLogColumns{
	Id:                 "id",
	Uuid:               "uuid",
	Price:              "price",
	Discount:           "discount",
	Expire:             "expire",
	MemberName:         "member_name",
	MemberPhone:        "member_phone",
	MemberNo:           "member_no",
	MemberCardTypeName: "member_card_type_name",
	MemberCardTypeUuid: "member_card_type_uuid",
	MemberUuid:         "member_uuid",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewMemberCardLogDao creates and returns a new DAO object for table data access.
func NewMemberCardLogDao(handlers ...gdb.ModelHandler) *MemberCardLogDao {
	return &MemberCardLogDao{
		group:    "default",
		table:    "ttpos_member_card_log",
		columns:  memberCardLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberCardLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberCardLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberCardLogDao) Columns() MemberCardLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberCardLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberCardLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberCardLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
