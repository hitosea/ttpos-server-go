// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseFormLogDao is the data access object for the table ttpos_purchase_form_log.
type PurchaseFormLogDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  PurchaseFormLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// PurchaseFormLogColumns defines and stores column names for the table ttpos_purchase_form_log.
type PurchaseFormLogColumns struct {
	Id               string // 自增ID
	Uuid             string // 采购单日志UUID
	PurchaseFormUuid string // 采购单uuid
	OperatorUuid     string // 操作人uuid
	Username         string // 操作人员
	Status           string // 操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	Operation        string // 操作动作
	Remark           string // 备注
	CreateTime       string // 创建时间(时间戳)
	UpdateTime       string // 更新时间(时间戳)
	DeleteTime       string // 删除时间(时间戳)
}

// purchaseFormLogColumns holds the columns for the table ttpos_purchase_form_log.
var purchaseFormLogColumns = PurchaseFormLogColumns{
	Id:               "id",
	Uuid:             "uuid",
	PurchaseFormUuid: "purchase_form_uuid",
	OperatorUuid:     "operator_uuid",
	Username:         "username",
	Status:           "status",
	Operation:        "operation",
	Remark:           "remark",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewPurchaseFormLogDao creates and returns a new DAO object for table data access.
func NewPurchaseFormLogDao(handlers ...gdb.ModelHandler) *PurchaseFormLogDao {
	return &PurchaseFormLogDao{
		group:    "default",
		table:    "ttpos_purchase_form_log",
		columns:  purchaseFormLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PurchaseFormLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PurchaseFormLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PurchaseFormLogDao) Columns() PurchaseFormLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PurchaseFormLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PurchaseFormLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PurchaseFormLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
