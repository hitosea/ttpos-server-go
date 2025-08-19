// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ClientVersionDao is the data access object for the table ttpos_client_version.
type ClientVersionDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  ClientVersionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// ClientVersionColumns defines and stores column names for the table ttpos_client_version.
type ClientVersionColumns struct {
	Id             string // 自增ID
	Type           string // 类型： 1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
	Brand          string // 品牌
	IsPublish      string // 是否发布 0-否 1-是
	Md5Hash        string // 谷歌云 md5-hash 值
	DownloadNum    string // 下载数量
	VersionNumber  string // 版本号
	VersionName    string // 版本名称
	ApkVersionCode string // Apk版本code
	ApkData        string // apk数据
	ForcedUpdate   string // 强制更新 0否 1是
	PackageUrl     string // 包地址
	OriginalName   string // 文件原名称
	UpdateLog      string // 更新日志
	CreateTime     string // 创建时间（时间戳）
	UpdateTime     string // 更新时间（时间戳）
	DeleteTime     string // 删除时间（时间戳）
}

// clientVersionColumns holds the columns for the table ttpos_client_version.
var clientVersionColumns = ClientVersionColumns{
	Id:             "id",
	Type:           "type",
	Brand:          "brand",
	IsPublish:      "is_publish",
	Md5Hash:        "md5_hash",
	DownloadNum:    "download_num",
	VersionNumber:  "version_number",
	VersionName:    "version_name",
	ApkVersionCode: "apk_version_code",
	ApkData:        "apk_data",
	ForcedUpdate:   "forced_update",
	PackageUrl:     "package_url",
	OriginalName:   "original_name",
	UpdateLog:      "update_log",
	CreateTime:     "create_time",
	UpdateTime:     "update_time",
	DeleteTime:     "delete_time",
}

// NewClientVersionDao creates and returns a new DAO object for table data access.
func NewClientVersionDao(handlers ...gdb.ModelHandler) *ClientVersionDao {
	return &ClientVersionDao{
		group:    "default",
		table:    "ttpos_client_version",
		columns:  clientVersionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ClientVersionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ClientVersionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ClientVersionDao) Columns() ClientVersionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ClientVersionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ClientVersionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ClientVersionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
