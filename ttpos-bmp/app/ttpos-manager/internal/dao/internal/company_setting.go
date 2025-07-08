// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CompanySettingDao is the data access object for the table ttpos_company_setting.
type CompanySettingDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  CompanySettingColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// CompanySettingColumns defines and stores column names for the table ttpos_company_setting.
type CompanySettingColumns struct {
	Id               string // 自增ID
	Uuid             string // 集团设置ID
	CompanyUuid      string // 集团ID
	RealName         string // 真实姓名
	LinkName         string // 联系人
	LinkPhone        string // 联系电话
	SaleStock        string // 进销存: 0不开启, 1开启
	IsOpenCoupon     string // 是否开启优惠券
	IsOpenMarketing  string // 是否开启营销活动
	IsOpenTax        string // 是否开启税务对接: 0不开启, 1奥地利 2-其他
	IsOpenMember     string // 是否开启会员: 0不开启, 1开启
	IsOpenTablet     string // 是否开启平板: 0不开启, 1开启
	IsOpenH5         string // 是否开启扫码H5: 0不开启, 1开启
	IsOpenAssistant  string // 是否开启点餐助手: 0不开启, 1开启
	IsOpenKitchenKds string // 是否开启后厨KDS: 0不开启, 1开启
	IsOpenBuffet     string // 是否开启自助餐: 0不开启, 1开启
	EnableSms        string // 是否启用短信功能：0-否；1-是
	SmsQuota         string // 短信配额
	IsOpenH5Order    string // 是否开启扫码点餐接单 0不开启, 1开启
	IsOpenLocalPrint string // 是否开启本地打印服务 0不开启, 1开启
	CashLimit        string // 收银机上限
	KitchenLimit     string // 厨显上限
	TabletLimit      string // 平板上限
	AssistantLimit   string // 点餐助手上限
	TableLimit       string // 桌台上限
	PrinterLimit     string // 打印机上限
	Timezone         string // 时区
	Languages        string // 支持语言
	Address          string // 联系地址
	CreateTime       string // 创建时间（时间戳）
	UpdateTime       string // 更新时间（时间戳）
	DeleteTime       string // 删除时间（时间戳）
}

// companySettingColumns holds the columns for the table ttpos_company_setting.
var companySettingColumns = CompanySettingColumns{
	Id:               "id",
	Uuid:             "uuid",
	CompanyUuid:      "company_uuid",
	RealName:         "real_name",
	LinkName:         "link_name",
	LinkPhone:        "link_phone",
	SaleStock:        "sale_stock",
	IsOpenCoupon:     "is_open_coupon",
	IsOpenMarketing:  "is_open_marketing",
	IsOpenTax:        "is_open_tax",
	IsOpenMember:     "is_open_member",
	IsOpenTablet:     "is_open_tablet",
	IsOpenH5:         "is_open_h5",
	IsOpenAssistant:  "is_open_assistant",
	IsOpenKitchenKds: "is_open_kitchen_kds",
	IsOpenBuffet:     "is_open_buffet",
	EnableSms:        "enable_sms",
	SmsQuota:         "sms_quota",
	IsOpenH5Order:    "is_open_h5_order",
	IsOpenLocalPrint: "is_open_local_print",
	CashLimit:        "cash_limit",
	KitchenLimit:     "kitchen_limit",
	TabletLimit:      "tablet_limit",
	AssistantLimit:   "assistant_limit",
	TableLimit:       "table_limit",
	PrinterLimit:     "printer_limit",
	Timezone:         "timezone",
	Languages:        "languages",
	Address:          "address",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewCompanySettingDao creates and returns a new DAO object for table data access.
func NewCompanySettingDao(handlers ...gdb.ModelHandler) *CompanySettingDao {
	return &CompanySettingDao{
		group:    "default",
		table:    "ttpos_company_setting",
		columns:  companySettingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CompanySettingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CompanySettingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CompanySettingDao) Columns() CompanySettingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CompanySettingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CompanySettingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CompanySettingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
