// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// JobDao is the data access object for the table takeout_job.
type JobDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  JobColumns         // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// JobColumns defines and stores column names for the table takeout_job.
type JobColumns struct {
	Id                   string //
	Uuid                 string // 外送订单UUID
	ShopRefNo            string // 餐馆订单参考，如UUID
	CustomerMobile       string // 下单客户电话
	CustomerEmail        string // 下单客户联系邮件
	ProviderName         string // 外送供应商： skootar,grab
	TakeoutRefNo         string // 外送系统订单号
	ShopLocationUuid     string // 餐馆位置信息
	ConsumerLocationUuid string // 消费者位置信息
	JobDate              string // 订单日期:'YYYY-MM-DD'
	StartTime            string // Start time. Format in 24 hr (00:00 to 23:59) or "now" for immediate job
	FinishTime           string // 订单结束时间
	PaymentType          string // 支付类型 Payment solution is 3 choice is ""invoice"", ""cash"", ""creditcard"",""prepaid""
	JobStatus            string // 外送订单状态
	Remark               string // 订单备注
	Reserved1            string // 保留字段1
	Reserved2            string // 保留字段2
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
	DeletedAt            string // 软删除
	CallbackUrl          string // 订单状态更新回调
	SkootarId            string // 骑手Id
	SkootarName          string // 骑手名称
	SkootarPhone         string // 骑手电话
	SkootarImageUrl      string // 骑手头像
	SkootarRating        string // 骑手评分
}

// jobColumns holds the columns for the table takeout_job.
var jobColumns = JobColumns{
	Id:                   "id",
	Uuid:                 "uuid",
	ShopRefNo:            "shop_ref_no",
	CustomerMobile:       "customer_mobile",
	CustomerEmail:        "customer_email",
	ProviderName:         "provider_name",
	TakeoutRefNo:         "takeout_ref_no",
	ShopLocationUuid:     "shop_location_uuid",
	ConsumerLocationUuid: "consumer_location_uuid",
	JobDate:              "job_date",
	StartTime:            "start_time",
	FinishTime:           "finish_time",
	PaymentType:          "payment_type",
	JobStatus:            "job_status",
	Remark:               "remark",
	Reserved1:            "reserved1",
	Reserved2:            "reserved2",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	DeletedAt:            "deleted_at",
	CallbackUrl:          "callback_url",
	SkootarId:            "skootar_id",
	SkootarName:          "skootar_name",
	SkootarPhone:         "skootar_phone",
	SkootarImageUrl:      "skootar_image_url",
	SkootarRating:        "skootar_rating",
}

// NewJobDao creates and returns a new DAO object for table data access.
func NewJobDao(handlers ...gdb.ModelHandler) *JobDao {
	return &JobDao{
		group:    "default",
		table:    "takeout_job",
		columns:  jobColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *JobDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *JobDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *JobDao) Columns() JobColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *JobDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *JobDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *JobDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
