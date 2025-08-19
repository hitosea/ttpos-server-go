// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// FileDao is the data access object for the table ttpos_file.
type FileDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  FileColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// FileColumns defines and stores column names for the table ttpos_file.
type FileColumns struct {
	Id            string // 自增ID
	Uuid          string // 文件ID
	Storage       string // 存储方式
	GroupUuid     string // 文件分组UUID
	FileUrl       string // 存储域名
	SaveName      string // 保存路径
	FileName      string // 文件路径
	FileSize      string // 文件大小(字节)
	FileType      string // 文件类型
	RealName      string // 文件真实名
	UrlParam      string // 签名参数
	IndexFileName string // 文件唯一名
	Extension     string // 文件扩展名
	IsUser        string // 是否为c端用户上传
	IsRecycle     string // 是否已回收
	CreateTime    string // 创建时间(时间戳)
	UpdateTime    string // 更新时间(时间戳)
	DeleteTime    string // 删除时间(时间戳)
}

// fileColumns holds the columns for the table ttpos_file.
var fileColumns = FileColumns{
	Id:            "id",
	Uuid:          "uuid",
	Storage:       "storage",
	GroupUuid:     "group_uuid",
	FileUrl:       "file_url",
	SaveName:      "save_name",
	FileName:      "file_name",
	FileSize:      "file_size",
	FileType:      "file_type",
	RealName:      "real_name",
	UrlParam:      "url_param",
	IndexFileName: "index_file_name",
	Extension:     "extension",
	IsUser:        "is_user",
	IsRecycle:     "is_recycle",
	CreateTime:    "create_time",
	UpdateTime:    "update_time",
	DeleteTime:    "delete_time",
}

// NewFileDao creates and returns a new DAO object for table data access.
func NewFileDao(handlers ...gdb.ModelHandler) *FileDao {
	return &FileDao{
		group:    "default",
		table:    "ttpos_file",
		columns:  fileColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *FileDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *FileDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *FileDao) Columns() FileColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *FileDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *FileDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *FileDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
