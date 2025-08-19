// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseFormDao is the data access object for the table ttpos_purchase_form.
type PurchaseFormDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  PurchaseFormColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// PurchaseFormColumns defines and stores column names for the table ttpos_purchase_form.
type PurchaseFormColumns struct {
	Id            string // 自增ID
	Uuid          string // 采购单ID
	FormNo        string // 编号
	Name          string // 采购单名称
	ApplicantUuid string // 申请人ID
	Remark        string // 备注
	Num           string // 总数量
	Amount        string // 总金额
	Status        string // 状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
	ArrivalTime   string // 到达时间(时间戳)
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// purchaseFormColumns holds the columns for the table ttpos_purchase_form.
var purchaseFormColumns = PurchaseFormColumns{
	Id:            "id",
	Uuid:          "uuid",
	FormNo:        "form_no",
	Name:          "name",
	ApplicantUuid: "applicant_uuid",
	Remark:        "remark",
	Num:           "num",
	Amount:        "amount",
	Status:        "status",
	ArrivalTime:   "arrival_time",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewPurchaseFormDao creates and returns a new DAO object for table data access.
func NewPurchaseFormDao(handlers ...gdb.ModelHandler) *PurchaseFormDao {
	return &PurchaseFormDao{
		group:    "default",
		table:    "ttpos_purchase_form",
		columns:  purchaseFormColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PurchaseFormDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PurchaseFormDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PurchaseFormDao) Columns() PurchaseFormColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PurchaseFormDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PurchaseFormDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PurchaseFormDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
