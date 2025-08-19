// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityDao is the data access object for the table ttpos_marketing_activity.
type MarketingActivityDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  MarketingActivityColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// MarketingActivityColumns defines and stores column names for the table ttpos_marketing_activity.
type MarketingActivityColumns struct {
	Id                    string //
	Uuid                  string // 活动唯一ID
	Name                  string // 活动名称
	Type                  string // 活动类型 0邀请有礼 1积分商城
	MultiLanguageNameUuid string // 活动名称多语言uuid
	Description           string // 活动描述
	MultiLanguageDescUuid string // 活动文案多语言uuid
	StartTime             string // 活动开始时间
	EndTime               string // 活动结束时间
	RewardConditionAmount string // 奖励条件金额
	IsOpenRewardLimit     string // 是否开启奖励次数限制 0否 1是
	RewardLimit           string // 奖励次数限制
	IsInvalid             string // 是否失效 0否 1是
	ImageBase64           string // 活动图片base64
	CreateTime            string // 创建时间
	UpdateTime            string // 更新时间
	DeleteTime            string // 删除时间
}

// marketingActivityColumns holds the columns for the table ttpos_marketing_activity.
var marketingActivityColumns = MarketingActivityColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Name:                  "name",
	Type:                  "type",
	MultiLanguageNameUuid: "multi_language_name_uuid",
	Description:           "description",
	MultiLanguageDescUuid: "multi_language_desc_uuid",
	StartTime:             "start_time",
	EndTime:               "end_time",
	RewardConditionAmount: "reward_condition_amount",
	IsOpenRewardLimit:     "is_open_reward_limit",
	RewardLimit:           "reward_limit",
	IsInvalid:             "is_invalid",
	ImageBase64:           "image_base64",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewMarketingActivityDao creates and returns a new DAO object for table data access.
func NewMarketingActivityDao(handlers ...gdb.ModelHandler) *MarketingActivityDao {
	return &MarketingActivityDao{
		group:    "default",
		table:    "ttpos_marketing_activity",
		columns:  marketingActivityColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingActivityDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingActivityDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingActivityDao) Columns() MarketingActivityColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingActivityDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingActivityDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingActivityDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
