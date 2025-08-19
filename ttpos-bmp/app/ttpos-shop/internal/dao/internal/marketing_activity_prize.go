// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityPrizeDao is the data access object for the table ttpos_marketing_activity_prize.
type MarketingActivityPrizeDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  MarketingActivityPrizeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// MarketingActivityPrizeColumns defines and stores column names for the table ttpos_marketing_activity_prize.
type MarketingActivityPrizeColumns struct {
	Id           string //
	Uuid         string // 礼品唯一ID
	ActivityUuid string // 活动uuid
	PrizeType    string // 奖品类型
	PrizeUuid    string // 奖品uuid
	CreateTime   string // 创建时间
	UpdateTime   string // 更新时间
	DeleteTime   string // 删除时间
}

// marketingActivityPrizeColumns holds the columns for the table ttpos_marketing_activity_prize.
var marketingActivityPrizeColumns = MarketingActivityPrizeColumns{
	Id:           "id",
	Uuid:         "uuid",
	ActivityUuid: "activity_uuid",
	PrizeType:    "prize_type",
	PrizeUuid:    "prize_uuid",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewMarketingActivityPrizeDao creates and returns a new DAO object for table data access.
func NewMarketingActivityPrizeDao(handlers ...gdb.ModelHandler) *MarketingActivityPrizeDao {
	return &MarketingActivityPrizeDao{
		group:    "default",
		table:    "ttpos_marketing_activity_prize",
		columns:  marketingActivityPrizeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MarketingActivityPrizeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MarketingActivityPrizeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MarketingActivityPrizeDao) Columns() MarketingActivityPrizeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MarketingActivityPrizeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MarketingActivityPrizeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MarketingActivityPrizeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
