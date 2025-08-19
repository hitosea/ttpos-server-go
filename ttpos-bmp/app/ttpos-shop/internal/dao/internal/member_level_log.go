// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MemberLevelLogDao is the data access object for the table ttpos_member_level_log.
type MemberLevelLogDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MemberLevelLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MemberLevelLogColumns defines and stores column names for the table ttpos_member_level_log.
type MemberLevelLogColumns struct {
	Id         string // 自增ID
	Uuid       string // 日志ID
	MemberUuid string // 会员ID
	OldLevelId string // 变更前的等级id
	NewLevelId string // 变更后的等级id
	ChangeType string // 变更类型(10后台管理员设置 20自动升级)
	Remark     string // 管理员备注
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// memberLevelLogColumns holds the columns for the table ttpos_member_level_log.
var memberLevelLogColumns = MemberLevelLogColumns{
	Id:         "id",
	Uuid:       "uuid",
	MemberUuid: "member_uuid",
	OldLevelId: "old_level_id",
	NewLevelId: "new_level_id",
	ChangeType: "change_type",
	Remark:     "remark",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewMemberLevelLogDao creates and returns a new DAO object for table data access.
func NewMemberLevelLogDao(handlers ...gdb.ModelHandler) *MemberLevelLogDao {
	return &MemberLevelLogDao{
		group:    "default",
		table:    "ttpos_member_level_log",
		columns:  memberLevelLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MemberLevelLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MemberLevelLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MemberLevelLogDao) Columns() MemberLevelLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MemberLevelLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MemberLevelLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MemberLevelLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
