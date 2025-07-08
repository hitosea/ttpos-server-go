// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderAbnormalRecordDao is the data access object for the table ttpos_sale_order_abnormal_record.
type SaleOrderAbnormalRecordDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  SaleOrderAbnormalRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// SaleOrderAbnormalRecordColumns defines and stores column names for the table ttpos_sale_order_abnormal_record.
type SaleOrderAbnormalRecordColumns struct {
	Id            string // 自增ID
	Uuid          string // UUID
	SaleBillUuid  string // 销售账单ID
	SaleOrderUuid string // 销售订单ID
	DutyNo        string // 当班编号
	Action        string // 行为
	SubAction     string // 自定义子行为
	Sign          string // 操作签名
	Remark        string // 备注
	CashierUuid   string // 收银员ID
	DeleteTime    string // 删除时间(时间戳)
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
}

// saleOrderAbnormalRecordColumns holds the columns for the table ttpos_sale_order_abnormal_record.
var saleOrderAbnormalRecordColumns = SaleOrderAbnormalRecordColumns{
	Id:            "id",
	Uuid:          "uuid",
	SaleBillUuid:  "sale_bill_uuid",
	SaleOrderUuid: "sale_order_uuid",
	DutyNo:        "duty_no",
	Action:        "action",
	SubAction:     "sub_action",
	Sign:          "sign",
	Remark:        "remark",
	CashierUuid:   "cashier_uuid",
	DeleteTime:    "delete_time",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
}

// NewSaleOrderAbnormalRecordDao creates and returns a new DAO object for table data access.
func NewSaleOrderAbnormalRecordDao(handlers ...gdb.ModelHandler) *SaleOrderAbnormalRecordDao {
	return &SaleOrderAbnormalRecordDao{
		group:    "default",
		table:    "ttpos_sale_order_abnormal_record",
		columns:  saleOrderAbnormalRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderAbnormalRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderAbnormalRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderAbnormalRecordDao) Columns() SaleOrderAbnormalRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderAbnormalRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderAbnormalRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderAbnormalRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
