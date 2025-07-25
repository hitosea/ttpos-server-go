// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobStatusLogDao is the data access object for the table takeout_job_status_log.
type JobStatusLogDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  JobStatusLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// JobStatusLogColumns defines and stores column names for the table takeout_job_status_log.
type JobStatusLogColumns struct {
	Id           string // 主键
	Uuid         string // 全局唯一ID
	JobUuid      string // 外送订单uuid
	StatusBefore string // 变更前状态
	StatusAfter  string // 变更后状态
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 软删除
}

// jobStatusLogColumns holds the columns for the table takeout_job_status_log.
var jobStatusLogColumns = JobStatusLogColumns{
	Id:           "id",
	Uuid:         "uuid",
	JobUuid:      "job_uuid",
	StatusBefore: "status_before",
	StatusAfter:  "status_after",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewJobStatusLogDao creates and returns a new DAO object for table data access.
func NewJobStatusLogDao(handlers ...gdb.ModelHandler) *JobStatusLogDao {
	return &JobStatusLogDao{
		group:    "default",
		table:    "takeout_job_status_log",
		columns:  jobStatusLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobStatusLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobStatusLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobStatusLogDao) Columns() JobStatusLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobStatusLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobStatusLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *JobStatusLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
