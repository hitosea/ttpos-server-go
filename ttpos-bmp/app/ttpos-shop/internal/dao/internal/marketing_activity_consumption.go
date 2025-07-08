// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityConsumptionDao is the data access object for the table ttpos_marketing_activity_consumption.
type MarketingActivityConsumptionDao struct {
	table    string                              // table is the underlying table name of the DAO.
	group    string                              // group is the database configuration group name of the current DAO.
	columns  MarketingActivityConsumptionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                  // handlers for customized model modification.
}

// MarketingActivityConsumptionColumns defines and stores column names for the table ttpos_marketing_activity_consumption.
type MarketingActivityConsumptionColumns struct {
	Id                string //
	Uuid              string // 消费记录唯一ID
	ActivityUuid      string // 活动uuid
	ReferrerUuid      string // 推荐人uuid
	ConsumerUuid      string // 消费人uuid
	ConsumptionAmount string // 消费金额
	RewardAmount      string // 奖励金额
	RewardStatus      string // 奖励状态 0未发放 1已发放
	CreateTime        string // 创建时间
	UpdateTime        string // 更新时间
	DeleteTime        string // 删除时间
}

// marketingActivityConsumptionColumns holds the columns for the table ttpos_marketing_activity_consumption.
var marketingActivityConsumptionColumns = MarketingActivityConsumptionColumns{
	Id:                "id",
	Uuid:              "uuid",
	ActivityUuid:      "activity_uuid",
	ReferrerUuid:      "referrer_uuid",
	ConsumerUuid:      "consumer_uuid",
	ConsumptionAmount: "consumption_amount",
	RewardAmount:      "reward_amount",
	RewardStatus:      "reward_status",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewMarketingActivityConsumptionDao creates and returns a new DAO object for table data access.
func NewMarketingActivityConsumptionDao(handlers ...gdb.ModelHandler) *MarketingActivityConsumptionDao {
	return &MarketingActivityConsumptionDao{
		group:    "default",
		table:    "ttpos_marketing_activity_consumption",
		columns:  marketingActivityConsumptionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingActivityConsumptionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingActivityConsumptionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingActivityConsumptionDao) Columns() MarketingActivityConsumptionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingActivityConsumptionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingActivityConsumptionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingActivityConsumptionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
