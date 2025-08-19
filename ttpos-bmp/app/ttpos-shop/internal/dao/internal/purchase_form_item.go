// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PurchaseFormItemDao is the data access object for the table ttpos_purchase_form_item.
type PurchaseFormItemDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  PurchaseFormItemColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// PurchaseFormItemColumns defines and stores column names for the table ttpos_purchase_form_item.
type PurchaseFormItemColumns struct {
	Id               string // 自增ID
	Uuid             string // 采购单明细ID
	PurchaseFormUuid string // 采购单ID
	MaterialType     string // 物料类型,0-商品 1-原料
	MaterialUuid     string // 物料ID
	EstimateNum      string // 预计数量
	EstimatePrice    string // 预计单价
	EstimateAmount   string // 预计金额
	Num              string // 数量
	Price            string // 单价
	Amount           string // 金额
	CreateTime       string // 创建时间(时间戳)
	UpdateTime       string // 更新时间(时间戳)
	DeleteTime       string // 删除时间(时间戳)
}

// purchaseFormItemColumns holds the columns for the table ttpos_purchase_form_item.
var purchaseFormItemColumns = PurchaseFormItemColumns{
	Id:               "id",
	Uuid:             "uuid",
	PurchaseFormUuid: "purchase_form_uuid",
	MaterialType:     "material_type",
	MaterialUuid:     "material_uuid",
	EstimateNum:      "estimate_num",
	EstimatePrice:    "estimate_price",
	EstimateAmount:   "estimate_amount",
	Num:              "num",
	Price:            "price",
	Amount:           "amount",
	CreateTime:       "create_time",
	UpdateTime:       "update_time",
	DeleteTime:       "delete_time",
}

// NewPurchaseFormItemDao creates and returns a new DAO object for table data access.
func NewPurchaseFormItemDao(handlers ...gdb.ModelHandler) *PurchaseFormItemDao {
	return &PurchaseFormItemDao{
		group:    "default",
		table:    "ttpos_purchase_form_item",
		columns:  purchaseFormItemColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PurchaseFormItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PurchaseFormItemDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PurchaseFormItemDao) Columns() PurchaseFormItemColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PurchaseFormItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PurchaseFormItemDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PurchaseFormItemDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
