// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityRecordDao is the data access object for the table ttpos_marketing_activity_record.
type MarketingActivityRecordDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  MarketingActivityRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// MarketingActivityRecordColumns defines and stores column names for the table ttpos_marketing_activity_record.
type MarketingActivityRecordColumns struct {
	Id             string //
	Uuid           string // 记录唯一ID
	ActivityUuid   string // 活动uuid
	PrizeUuid      string // 奖品uuid
	MemberUuid     string // 会员uuid
	RewardCount    string // 已获得奖励次数
	LastRewardTime string // 最后一次获得奖励时间
	CreateTime     string // 创建时间
	UpdateTime     string // 更新时间
	DeleteTime     string // 删除时间
}

// marketingActivityRecordColumns holds the columns for the table ttpos_marketing_activity_record.
var marketingActivityRecordColumns = MarketingActivityRecordColumns{
	Id:             "id",
	Uuid:           "uuid",
	ActivityUuid:   "activity_uuid",
	PrizeUuid:      "prize_uuid",
	MemberUuid:     "member_uuid",
	RewardCount:    "reward_count",
	LastRewardTime: "last_reward_time",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewMarketingActivityRecordDao creates and returns a new DAO object for table data access.
func NewMarketingActivityRecordDao(handlers ...gdb.ModelHandler) *MarketingActivityRecordDao {
	return &MarketingActivityRecordDao{
		group:    "default",
		table:    "ttpos_marketing_activity_record",
		columns:  marketingActivityRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingActivityRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingActivityRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingActivityRecordDao) Columns() MarketingActivityRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingActivityRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingActivityRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingActivityRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
