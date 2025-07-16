// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobLocationDao is the data access object for the table takeout_job_location.
type JobLocationDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  JobLocationColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// JobLocationColumns defines and stores column names for the table takeout_job_location.
type JobLocationColumns struct {
	Id           string // 主键
	Uuid         string // 全局唯一uuid
	LocationType string // 位置类型： 0 餐馆，1 消费者
	AddressName  string // 地址说明
	Address      string // 详细地址
	Lat          string // 纬度
	Lng          string // 经度
	ContactName  string // 联系人名称
	ContactPhone string // 联系人号码
	Seq          string // 地址序列，1开始
	Remark       string // 备注
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 软删除
}

// jobLocationColumns holds the columns for the table takeout_job_location.
var jobLocationColumns = JobLocationColumns{
	Id:           "id",
	Uuid:         "uuid",
	LocationType: "location_type",
	AddressName:  "address_name",
	Address:      "address",
	Lat:          "lat",
	Lng:          "lng",
	ContactName:  "contact_name",
	ContactPhone: "contact_phone",
	Seq:          "seq",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewJobLocationDao creates and returns a new DAO object for table data access.
func NewJobLocationDao(handlers ...gdb.ModelHandler) *JobLocationDao {
	return &JobLocationDao{
		group:    "default",
		table:    "takeout_job_location",
		columns:  jobLocationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobLocationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobLocationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobLocationDao) Columns() JobLocationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobLocationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobLocationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *JobLocationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
