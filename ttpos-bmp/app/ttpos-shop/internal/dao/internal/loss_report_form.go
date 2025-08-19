// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LossReportFormDao is the data access object for the table ttpos_loss_report_form.
type LossReportFormDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  LossReportFormColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// LossReportFormColumns defines and stores column names for the table ttpos_loss_report_form.
type LossReportFormColumns struct {
	Id             string // 自增ID
	Uuid           string // 报损单ID
	FormNo         string // 编号
	Scene          string // 报损类型,0-loss损耗 1-lost丢失
	Num            string // 数量
	Remark         string // 备注
	ProductBomUuid string // 商品清单bom uuid
	MaterialUuid   string // 物料ID
	ApplicantUuid  string // 申请人ID
	RejectReason   string // 拒绝原因
	Status         string // 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
	OperatorUuid   string // 操作员ID
	ApprovedTime   string // 通过时间(时间戳)
	RevokeTime     string // 撤销时间(时间戳)
	CreateTime     string // 创建时间(时间戳)
	UpdateTime     string // 更新时间(时间戳)
	DeleteTime     string // 删除时间(时间戳)
}

// lossReportFormColumns holds the columns for the table ttpos_loss_report_form.
var lossReportFormColumns = LossReportFormColumns{
	Id:             "id",
	Uuid:           "uuid",
	FormNo:         "form_no",
	Scene:          "scene",
	Num:            "num",
	Remark:         "remark",
	ProductBomUuid: "product_bom_uuid",
	MaterialUuid:   "material_uuid",
	ApplicantUuid:  "applicant_uuid",
	RejectReason:   "reject_reason",
	Status:         "status",
	OperatorUuid:   "operator_uuid",
	ApprovedTime:   "approved_time",
	RevokeTime:     "revoke_time",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewLossReportFormDao creates and returns a new DAO object for table data access.
func NewLossReportFormDao(handlers ...gdb.ModelHandler) *LossReportFormDao {
	return &LossReportFormDao{
		group:    "default",
		table:    "ttpos_loss_report_form",
		columns:  lossReportFormColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LossReportFormDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LossReportFormDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LossReportFormDao) Columns() LossReportFormColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LossReportFormDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LossReportFormDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *LossReportFormDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
