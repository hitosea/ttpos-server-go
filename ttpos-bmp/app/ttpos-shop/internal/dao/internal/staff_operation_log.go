// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StaffOperationLogDao is the data access object for the table ttpos_staff_operation_log.
type StaffOperationLogDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  StaffOperationLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// StaffOperationLogColumns defines and stores column names for the table ttpos_staff_operation_log.
type StaffOperationLogColumns struct {
	Id           string // 自增ID
	Uuid         string // 操作日志ID
	StaffUuid    string // 员工ID
	Title        string // 标题
	Url          string // 操作URL
	RequestData  string // 请求数据
	ResponseData string // 响应数据
	Type         string // 操作类型
	Ip           string // 操作IP
	Source       string // 操作来源
	Agent        string // 操作用户代理
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// staffOperationLogColumns holds the columns for the table ttpos_staff_operation_log.
var staffOperationLogColumns = StaffOperationLogColumns{
	Id:           "id",
	Uuid:         "uuid",
	StaffUuid:    "staff_uuid",
	Title:        "title",
	Url:          "url",
	RequestData:  "request_data",
	ResponseData: "response_data",
	Type:         "type",
	Ip:           "ip",
	Source:       "source",
	Agent:        "agent",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewStaffOperationLogDao creates and returns a new DAO object for table data access.
func NewStaffOperationLogDao(handlers ...gdb.ModelHandler) *StaffOperationLogDao {
	return &StaffOperationLogDao{
		group:    "default",
		table:    "ttpos_staff_operation_log",
		columns:  staffOperationLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StaffOperationLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StaffOperationLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StaffOperationLogDao) Columns() StaffOperationLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StaffOperationLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StaffOperationLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StaffOperationLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
