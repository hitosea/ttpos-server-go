// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterLogDao is the data access object for the table ttpos_printer_log.
type PrinterLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  PrinterLogColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// PrinterLogColumns defines and stores column names for the table ttpos_printer_log.
type PrinterLogColumns struct {
	Id                 string // 自增ID
	Uuid               string // 打印日志ID
	PrinterUuid        string // 打印机id
	ProductPrinterUuid string // 商品打印机id
	CashierDeviceId    string // 收银机绑定的id
	ReadDeviceId       string // 读取设备id
	RelatedType        string // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid        string // 销售账单、充值订单id
	Data               string // 打印数据
	Type               string // 类型:0系统默认队列,1云上服务下放
	DataType           string // 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单
	PrintMethod        string // 打印方式 1文本打印, 2图片打印
	PrinterType        string // 打印机类型
	Num                string // 打印次数
	Status             string // 状态(0结束,1进行中,2成功)
	Reason             string // 原因
	PrinterTime        string // 打印时间
	FirstExecution     string // 是否首次执行打印 1-是 0-否
	CreateTime         string // 创建时间(时间戳)
	UpdateTime         string // 更新时间(时间戳)
	DeleteTime         string // 删除时间(时间戳)
}

// printerLogColumns holds the columns for the table ttpos_printer_log.
var printerLogColumns = PrinterLogColumns{
	Id:                 "id",
	Uuid:               "uuid",
	PrinterUuid:        "printer_uuid",
	ProductPrinterUuid: "product_printer_uuid",
	CashierDeviceId:    "cashier_device_id",
	ReadDeviceId:       "read_device_id",
	RelatedType:        "related_type",
	RelatedUuid:        "related_uuid",
	Data:               "data",
	Type:               "type",
	DataType:           "data_type",
	PrintMethod:        "print_method",
	PrinterType:        "printer_type",
	Num:                "num",
	Status:             "status",
	Reason:             "reason",
	PrinterTime:        "printer_time",
	FirstExecution:     "first_execution",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
	DeleteTime:         "delete_time",
}

// NewPrinterLogDao creates and returns a new DAO object for table data access.
func NewPrinterLogDao(handlers ...gdb.ModelHandler) *PrinterLogDao {
	return &PrinterLogDao{
		group:    "default",
		table:    "ttpos_printer_log",
		columns:  printerLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PrinterLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PrinterLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PrinterLogDao) Columns() PrinterLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PrinterLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PrinterLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PrinterLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
