// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsProductDao is the data access object for the table ttpos_statistics_product.
type StatisticsProductDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  StatisticsProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// StatisticsProductColumns defines and stores column names for the table ttpos_statistics_product.
type StatisticsProductColumns struct {
	Id                 string // 自增ID
	Uuid               string // UUID
	SaleBillUuid       string // 销售单UUID
	SaleOrderUuid      string // 销售订单UUID
	DutyNo             string // 当班编号
	DeskUuid           string // 桌台UUID
	ProductPackageUuid string // 商品包uuid
	ProductBomUuid     string // 商品清单uuid
	ProductPrice       string // 商品单价: 未含税
	ProductSalePrice   string // 商品销售价: 规格+加料
	ProductFinalPrice  string // 商品最终价
	ProductNum         string // 商品数量
	TaxRate            string // 税率
	TaxFee             string // 税费
	ServiceFee         string // 服务费
	ServiceTax         string // 服务税
	GiveNum            string // 赠菜数量
	FreeNum            string // 免单数量
	RefundNum          string // 退款数量
	CompleteTime       string // 完成时间
	RefundTime         string // 完成时间
	CreateTime         string // 创建时间
	UpdateTime         string // 更新时间
	DeleteTime         string // 删除时间
}

// statisticsProductColumns holds the columns for the table ttpos_statistics_product.
var statisticsProductColumns = StatisticsProductColumns{
	Id:                 "id",
	Uuid:               "uuid",
	SaleBillUuid:       "sale_bill_uuid",
	SaleOrderUuid:      "sale_order_uuid",
	DutyNo:             "duty_no",
	DeskUuid:           "desk_uuid",
	ProductPackageUuid: "product_package_uuid",
	ProductBomUuid:     "product_bom_uuid",
	ProductPrice:       "product_price",
	ProductSalePrice:   "product_sale_price",
	ProductFinalPrice:  "product_final_price",
	ProductNum:         "product_num",
	TaxRate:            "tax_rate",
	TaxFee:             "tax_fee",
	ServiceFee:         "service_fee",
	ServiceTax:         "service_tax",
	GiveNum:            "give_num",
	FreeNum:            "free_num",
	RefundNum:          "refund_num",
	CompleteTime:       "complete_time",
	RefundTime:         "refund_time",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewStatisticsProductDao creates and returns a new DAO object for table data access.
func NewStatisticsProductDao(handlers ...gdb.ModelHandler) *StatisticsProductDao {
	return &StatisticsProductDao{
		group:    "default",
		table:    "ttpos_statistics_product",
		columns:  statisticsProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StatisticsProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StatisticsProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StatisticsProductDao) Columns() StatisticsProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StatisticsProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StatisticsProductDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StatisticsProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
