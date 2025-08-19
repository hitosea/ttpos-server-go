// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SupplierDao is the data access object for the table ttpos_supplier.
type SupplierDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SupplierColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SupplierColumns defines and stores column names for the table ttpos_supplier.
type SupplierColumns struct {
	Id           string // 自增ID
	Uuid         string // 供应商ID
	Name         string // 供应商名称
	Address      string // 供应商地址
	ContactName  string // 联系人姓名
	ContactPhone string // 联系人电话
	Position     string // 职位
	StaffUuid    string // 员工ID, 采购负责人
	CreateTime   string // 创建时间(时间戳)
	UpdateTime   string // 更新时间(时间戳)
	DeleteTime   string // 删除时间(时间戳)
}

// supplierColumns holds the columns for the table ttpos_supplier.
var supplierColumns = SupplierColumns{
	Id:           "id",
	Uuid:         "uuid",
	Name:         "name",
	Address:      "address",
	ContactName:  "contact_name",
	ContactPhone: "contact_phone",
	Position:     "position",
	StaffUuid:    "staff_uuid",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
	DeleteTime:   "delete_time",
}

// NewSupplierDao creates and returns a new DAO object for table data access.
func NewSupplierDao(handlers ...gdb.ModelHandler) *SupplierDao {
	return &SupplierDao{
		group:    "default",
		table:    "ttpos_supplier",
		columns:  supplierColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SupplierDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SupplierDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SupplierDao) Columns() SupplierColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SupplierDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SupplierDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SupplierDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
