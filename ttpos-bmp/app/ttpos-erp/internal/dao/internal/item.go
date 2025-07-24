// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ItemDao is the data access object for table item.
type ItemDao struct {
	table   string      // table is the underlying table name of the DAO.
	group   string      // group is the database configuration group name of current DAO.
	columns ItemColumns // columns contains all the column names of Table for convenient usage.
}

// ItemColumns defines and stores column names for table item.
type ItemColumns struct {
	ID           string // 主键ID
	Uuid         string // UUID
	Name         string // 商品名称
	ImageName    string // 图片名称
	Status       string // 状态
	CategoryUuid string // 分类UUID
	CreateTime   string // 创建时间
	UpdateTime   string // 更新时间
	DeleteTime   string // 删除时间
}

// itemColumns holds the columns for table item.
var itemColumns = ItemColumns{
	ID:           "id",
	Uuid:         "uuid",
	Name:         "name",
	ImageName:    "image_name",
	Status:       "status",
	CategoryUuid: "category_uuid",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewItemDao creates and returns a new DAO object for table data access.
func NewItemDao() *ItemDao {
	return &ItemDao{
		group:   "default",
		table:   "ttpos_product_package", // 使用已存在的商品包表
		columns: itemColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ItemDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ItemDao) Table() string {
	return dao.table
}

// Columns returns the columns of current dao.
func (dao *ItemDao) Columns() ItemColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ItemDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ItemDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
