// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderOperationRecordDao is the data access object for the table ttpos_sale_order_operation_record.
type SaleOrderOperationRecordDao struct {
	table    string                          // table is the underlying table name of the DAO.
	group    string                          // group is the database configuration group name of the current DAO.
	columns  SaleOrderOperationRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler              // handlers for customized model modification.
}

// SaleOrderOperationRecordColumns defines and stores column names for the table ttpos_sale_order_operation_record.
type SaleOrderOperationRecordColumns struct {
	Id            string // 自增ID
	Uuid          string // 桌台账单记录ID
	Source        string // 操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐
	Action        string // 操作行为
	Data          string // 数据
	Remark        string // 备注
	SaleBillUuid  string // 销售账单ID
	SaleOrderUuid string // 销售订单ID
	H5OrderUuid   string // h5订单Uuid
	OperatorUuid  string // 操作员ID
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// saleOrderOperationRecordColumns holds the columns for the table ttpos_sale_order_operation_record.
var saleOrderOperationRecordColumns = SaleOrderOperationRecordColumns{
	Id:            "id",
	Uuid:          "uuid",
	Source:        "source",
	Action:        "action",
	Data:          "data",
	Remark:        "remark",
	SaleBillUuid:  "sale_bill_uuid",
	SaleOrderUuid: "sale_order_uuid",
	H5OrderUuid:   "h5_order_uuid",
	OperatorUuid:  "operator_uuid",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewSaleOrderOperationRecordDao creates and returns a new DAO object for table data access.
func NewSaleOrderOperationRecordDao(handlers ...gdb.ModelHandler) *SaleOrderOperationRecordDao {
	return &SaleOrderOperationRecordDao{
		group:    "default",
		table:    "ttpos_sale_order_operation_record",
		columns:  saleOrderOperationRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderOperationRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderOperationRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderOperationRecordDao) Columns() SaleOrderOperationRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderOperationRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderOperationRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderOperationRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
