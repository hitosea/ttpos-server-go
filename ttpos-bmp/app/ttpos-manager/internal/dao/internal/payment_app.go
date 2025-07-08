// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentAppDao is the data access object for the table ttpos_payment_app.
type PaymentAppDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PaymentAppColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PaymentAppColumns defines and stores column names for the table ttpos_payment_app.
type PaymentAppColumns struct {
	Id                   string // 自增ID
	CompanyUuid          string // 集团ID
	LlWhiteIp            string // 白名单IP
	LlMerchantId         string // 商户号
	LlStoreId            string // 站点ID
	LlPublicKey          string // LianLianpay公钥
	LlMerchantPrivateKey string // 商户私钥
	LlToken              string // Token
	LlSignSalt           string // 签名盐
	CreateTime           string // 创建时间（时间戳）
	UpdateTime           string // 更新时间（时间戳）
	DeleteTime           string // 删除时间（时间戳）
}

// paymentAppColumns holds the columns for the table ttpos_payment_app.
var paymentAppColumns = PaymentAppColumns{
	Id:                   "id",
	CompanyUuid:          "company_uuid",
	LlWhiteIp:            "ll_white_ip",
	LlMerchantId:         "ll_merchant_id",
	LlStoreId:            "ll_store_id",
	LlPublicKey:          "ll_public_key",
	LlMerchantPrivateKey: "ll_merchant_private_key",
	LlToken:              "ll_token",
	LlSignSalt:           "ll_sign_salt",
	CreateTime:           "create_time",
	UpdateTime:           "update_time",
	DeleteTime:           "delete_time",
}

// NewPaymentAppDao creates and returns a new DAO object for table data access.
func NewPaymentAppDao(handlers ...gdb.ModelHandler) *PaymentAppDao {
	return &PaymentAppDao{
		group:    "default",
		table:    "ttpos_payment_app",
		columns:  paymentAppColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PaymentAppDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PaymentAppDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PaymentAppDao) Columns() PaymentAppColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PaymentAppDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PaymentAppDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PaymentAppDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
