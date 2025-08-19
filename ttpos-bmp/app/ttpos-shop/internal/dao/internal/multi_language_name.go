// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MultiLanguageNameDao is the data access object for the table ttpos_multi_language_name.
type MultiLanguageNameDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  MultiLanguageNameColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// MultiLanguageNameColumns defines and stores column names for the table ttpos_multi_language_name.
type MultiLanguageNameColumns struct {
	Id         string // 自增ID
	Uuid       string // 多语言名称ID
	EnName     string // 英文名称
	ZhName     string // 中文名称
	ZhTwName   string // 繁体中文名称
	ThName     string // 泰语名称
	MyName     string // 缅甸语名称
	JaName     string // 日语名称
	KoName     string // 韩语名称
	TrName     string // 土耳其语名称
	CreateTime string // 创建时间(时间戳)
	UpdateTime string // 更新时间(时间戳)
	DeleteTime string // 删除时间(时间戳)
}

// multiLanguageNameColumns holds the columns for the table ttpos_multi_language_name.
var multiLanguageNameColumns = MultiLanguageNameColumns{
	Id:         "id",
	Uuid:       "uuid",
	EnName:     "en_name",
	ZhName:     "zh_name",
	ZhTwName:   "zh_tw_name",
	ThName:     "th_name",
	MyName:     "my_name",
	JaName:     "ja_name",
	KoName:     "ko_name",
	TrName:     "tr_name",
	CreateTime: "create_time",
	UpdateTime: "update_time",
	DeleteTime: "delete_time",
}

// NewMultiLanguageNameDao creates and returns a new DAO object for table data access.
func NewMultiLanguageNameDao(handlers ...gdb.ModelHandler) *MultiLanguageNameDao {
	return &MultiLanguageNameDao{
		group:    "default",
		table:    "ttpos_multi_language_name",
		columns:  multiLanguageNameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MultiLanguageNameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MultiLanguageNameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MultiLanguageNameDao) Columns() MultiLanguageNameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MultiLanguageNameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MultiLanguageNameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MultiLanguageNameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
