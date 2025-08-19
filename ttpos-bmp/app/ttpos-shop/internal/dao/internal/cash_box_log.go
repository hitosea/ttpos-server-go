// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CashBoxLogDao is the data access object for the table ttpos_cash_box_log.
type CashBoxLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CashBoxLogColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CashBoxLogColumns defines and stores column names for the table ttpos_cash_box_log.
type CashBoxLogColumns struct {
	Id                    string // 自增ID
	Uuid                  string // 钱箱ID
	Scene                 string // 场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零
	Amount                string // 金额
	Remark                string // 备注
	Processed             string // 是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额
	RelatedUuid           string // 关联的充值订单、销售订单ID,场景为1、6时必填
	ReturnOrderUuid       string // 退货单ID,场景为2时必填
	RefundOrderAmountUuid string // 退款单金额ID,场景为3时必填
	CreateTime            string // 创建时间(时间戳)
	UpdateTime            string // 更新时间(时间戳)
	DeleteTime            string // 删除时间(时间戳)
}

// cashBoxLogColumns holds the columns for the table ttpos_cash_box_log.
var cashBoxLogColumns = CashBoxLogColumns{
	Id:                    "id",
	Uuid:                  "uuid",
	Scene:                 "scene",
	Amount:                "amount",
	Remark:                "remark",
	Processed:             "processed",
	RelatedUuid:           "related_uuid",
	ReturnOrderUuid:       "return_order_uuid",
	RefundOrderAmountUuid: "refund_order_amount_uuid",
	CreateTime:            "create_time",
	UpdateTime:            "update_time",
	DeleteTime:            "delete_time",
}

// NewCashBoxLogDao creates and returns a new DAO object for table data access.
func NewCashBoxLogDao(handlers ...gdb.ModelHandler) *CashBoxLogDao {
	return &CashBoxLogDao{
		group:    "default",
		table:    "ttpos_cash_box_log",
		columns:  cashBoxLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CashBoxLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CashBoxLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CashBoxLogDao) Columns() CashBoxLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CashBoxLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CashBoxLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CashBoxLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
