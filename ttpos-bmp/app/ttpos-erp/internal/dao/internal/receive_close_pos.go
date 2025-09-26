// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveClosePosDao is the data access object for the table erp_receive_close_pos.
type ReceiveClosePosDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ReceiveClosePosColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ReceiveClosePosColumns defines and stores column names for the table erp_receive_close_pos.
type ReceiveClosePosColumns struct {
	Id               string // ID
	PosOpenEntryName string // 开帐名称
	PeriodEndDate    string // 结账时间
	Docstatus        string // 文档状态，参考erpnext
	CreatedAt        string // 创建时间
	UpdatedAt        string // 更新时间
	ReqMessage       string // 请求数据,base64编码
	RespMessage      string // 响应数据,base64编码
	SiteCode         string // erp_site_code, 用来区分调那个租户
}

// receiveClosePosColumns holds the columns for the table erp_receive_close_pos.
var receiveClosePosColumns = ReceiveClosePosColumns{
	Id:               "id",
	PosOpenEntryName: "pos_open_entry_name",
	PeriodEndDate:    "period_end_date",
	Docstatus:        "docstatus",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	ReqMessage:       "req_message",
	RespMessage:      "resp_message",
	SiteCode:         "site_code",
}

// NewReceiveClosePosDao creates and returns a new DAO object for table data access.
func NewReceiveClosePosDao(handlers ...gdb.ModelHandler) *ReceiveClosePosDao {
	return &ReceiveClosePosDao{
		group:    "default",
		table:    "erp_receive_close_pos",
		columns:  receiveClosePosColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReceiveClosePosDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReceiveClosePosDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReceiveClosePosDao) Columns() ReceiveClosePosColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReceiveClosePosDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReceiveClosePosDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReceiveClosePosDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
