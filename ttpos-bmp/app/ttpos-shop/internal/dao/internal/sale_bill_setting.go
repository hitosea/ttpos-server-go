// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleBillSettingDao is the data access object for the table ttpos_sale_bill_setting.
type SaleBillSettingDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  SaleBillSettingColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// SaleBillSettingColumns defines and stores column names for the table ttpos_sale_bill_setting.
type SaleBillSettingColumns struct {
	Id               string // 自增ID
	Uuid             string // 销售账单设置ID
	SaleBillUuid     string // 销售账单ID
	ServiceFeeType   string // 服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0
	ServiceFeeValue  string // 服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例
	TaxFeeType       string // 税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税
	ZeroRule         string // 优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
	ZeroCheckoutRule string // 结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元
	IsStatGift       string // 是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣
	IsStatFree       string // 是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费
	DiscountType     string // 打折类型, 0-百分比打折% 1-百分比直接减免% off
	CreateTime       string // 创建时间(时间戳)
	UpdateTime       string // 更新时间(时间戳)
	DeleteTime       string // 删除时间(时间戳)
}

// saleBillSettingColumns holds the columns for the table ttpos_sale_bill_setting.
var saleBillSettingColumns = SaleBillSettingColumns{
	Id:               "id",
	Uuid:             "uuid",
	SaleBillUuid:     "sale_bill_uuid",
	ServiceFeeType:   "service_fee_type",
	ServiceFeeValue:  "service_fee_value",
	TaxFeeType:       "tax_fee_type",
	ZeroRule:         "zero_rule",
	ZeroCheckoutRule: "zero_checkout_rule",
	IsStatGift:       "is_stat_gift",
	IsStatFree:       "is_stat_free",
	DiscountType:     "discount_type",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewSaleBillSettingDao creates and returns a new DAO object for table data access.
func NewSaleBillSettingDao(handlers ...gdb.ModelHandler) *SaleBillSettingDao {
	return &SaleBillSettingDao{
		group:    "default",
		table:    "ttpos_sale_bill_setting",
		columns:  saleBillSettingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleBillSettingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleBillSettingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleBillSettingDao) Columns() SaleBillSettingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleBillSettingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleBillSettingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleBillSettingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
