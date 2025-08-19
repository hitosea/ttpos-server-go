// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UploadFileDao is the data access object for the table ttpos_upload_file.
type UploadFileDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UploadFileColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UploadFileColumns defines and stores column names for the table ttpos_upload_file.
type UploadFileColumns struct {
	FileId     string // 自增ID
	Storage    string // 存储方式
	GroupId    string // 文件分组ID
	FileUrl    string // 存储域名
	SaveName   string // 保存路径
	FileName   string // 文件路径
	FileSize   string // 文件大小(字节)
	FileType   string // 文件类型
	RealName   string // 文件真实名
	UrlParam   string // 签名参数
	Extension  string // 文件扩展名
	IsUser     string // 是否为c端用户上传
	IsRecycle  string // 是否已回收
	CreateTime string // 创建时间（时间戳）
	UpdateTime string // 更新时间（时间戳）
	DeleteTime string // 删除时间（时间戳）
}

// uploadFileColumns holds the columns for the table ttpos_upload_file.
var uploadFileColumns = UploadFileColumns{
	FileId:     "file_id",
	Storage:    "storage",
	GroupId:    "group_id",
	FileUrl:    "file_url",
	SaveName:   "save_name",
	FileName:   "file_name",
	FileSize:   "file_size",
	FileType:   "file_type",
	RealName:   "real_name",
	UrlParam:   "url_param",
	Extension:  "extension",
	IsUser:     "is_user",
	IsRecycle:  "is_recycle",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewUploadFileDao creates and returns a new DAO object for table data access.
func NewUploadFileDao(handlers ...gdb.ModelHandler) *UploadFileDao {
	return &UploadFileDao{
		group:    "default",
		table:    "ttpos_upload_file",
		columns:  uploadFileColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UploadFileDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UploadFileDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UploadFileDao) Columns() UploadFileColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UploadFileDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UploadFileDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UploadFileDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
