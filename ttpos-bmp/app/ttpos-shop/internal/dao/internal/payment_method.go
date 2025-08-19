// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentMethodDao is the data access object for the table ttpos_payment_method.
type PaymentMethodDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  PaymentMethodColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// PaymentMethodColumns defines and stores column names for the table ttpos_payment_method.
type PaymentMethodColumns struct {
	Id                   string // 自增ID
	Uuid                 string // 支付方式ID
	Name                 string // 支付方式名称
	Code                 string // 支付方式代号
	PaymentName          string // 支付名称
	Source               string // 来源 0-系统 1-手动 2-LianLianPay
	LogoFileUuid         string // logo图片ID
	QrcodeFileUuid       string // 二维码图片ID
	FeePercent           string // 手续费百分比,取值范围0-1
	IsShowCashier        string // 0-不显示 1-收银机结账显示
	IsShowAssistant      string // 0-不显示 1-点餐助手结账显示
	IsShowMemberRecharge string // 0-不显示 1-收银机会员充值显示
	Status               string // 状态 0-禁用 1-启用
	Sort                 string // 排序
	CreateTime           string // 创建时间(时间戳)
	UpdateTime           string // 更新时间(时间戳)
	DeleteTime           string // 删除时间(时间戳)
}

// paymentMethodColumns holds the columns for the table ttpos_payment_method.
var paymentMethodColumns = PaymentMethodColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	Name:                 "name",
	Code:                 "code",
	PaymentName:          "payment_name",
	Source:               "source",
	LogoFileUuid:         "logo_file_uuid",
	QrcodeFileUuid:       "qrcode_file_uuid",
	FeePercent:           "fee_percent",
	IsShowCashier:        "is_show_cashier",
	IsShowAssistant:      "is_show_assistant",
	IsShowMemberRecharge: "is_show_member_recharge",
	Status:               "status",
	Sort:                 "sort",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewPaymentMethodDao creates and returns a new DAO object for table data access.
func NewPaymentMethodDao(handlers ...gdb.ModelHandler) *PaymentMethodDao {
	return &PaymentMethodDao{
		group:    "default",
		table:    "ttpos_payment_method",
		columns:  paymentMethodColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PaymentMethodDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PaymentMethodDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PaymentMethodDao) Columns() PaymentMethodColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PaymentMethodDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PaymentMethodDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PaymentMethodDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
