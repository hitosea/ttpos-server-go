// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ShopProviderCfgDao is the data access object for the table shop_provider_cfg.
type ShopProviderCfgDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ShopProviderCfgColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ShopProviderCfgColumns defines and stores column names for the table shop_provider_cfg.
type ShopProviderCfgColumns struct {
	Id                 string // 主键ID
	Uuid               string // 唯一标识
	ShopUuid           string // 门店UUID
	ProviderName       string // 第三方名称，如 grab
	ProviderMerchantId string // 第三方商户ID
	ProviderShopStatus string // 门店集成状态
	CreatedAt          string // 创建时间
	UpdatedAt          string // 更新时间
	DeletedAt          string // 删除时间
}

// shopProviderCfgColumns holds the columns for the table shop_provider_cfg.
var shopProviderCfgColumns = ShopProviderCfgColumns{
	Id:                 "id",
	Uuid:               "uuid",
	ShopUuid:           "shop_uuid",
	ProviderName:       "provider_name",
	ProviderMerchantId: "provider_merchant_id",
	ProviderShopStatus: "provider_shop_status",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	DeletedAt:          "deleted_at",
}

// NewShopProviderCfgDao creates and returns a new DAO object for table data access.
func NewShopProviderCfgDao(handlers ...gdb.ModelHandler) *ShopProviderCfgDao {
	return &ShopProviderCfgDao{
		group:    "default",
		table:    "shop_provider_cfg",
		columns:  shopProviderCfgColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ShopProviderCfgDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ShopProviderCfgDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ShopProviderCfgDao) Columns() ShopProviderCfgColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ShopProviderCfgDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ShopProviderCfgDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ShopProviderCfgDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
