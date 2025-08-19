// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// StaffShiftLogDao is the data access object for the table ttpos_staff_shift_log.
type StaffShiftLogDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  StaffShiftLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// StaffShiftLogColumns defines and stores column names for the table ttpos_staff_shift_log.
type StaffShiftLogColumns struct {
	Id                string // 自增ID
	Uuid              string // 交班记录ID
	StaffUuid         string // 员工ID
	ShiftNo           string // 交班编号
	Status            string // 状态: 0未交班,1已交班
	PreviousShiftCash string // 上一班遗留备用金
	CurrentCashTotal  string // 当前钱箱现金总计
	Incomes           string // 收入详情
	TotalIncome       string // 总收入
	CashTakenOut      string // 本班取出现金
	CashLeft          string // 本班遗留备用金
	CashIncome        string // 本班收入现金
	TotalBusiness     string // 本班营业总额(不包含退款)
	IsPrinted         string // 是否打印 0-未打印 1-已打印
	Remark            string // 备注
	WithdrawCash      string // 中途取出现金
	DepositCash       string // 中途存入现金
	ExceptionRemark   string // 异常报备
	Abnormal          string // 异常信息-json字符串
	ShiftStartTime    string // 当班开始时间
	ShiftEndTime      string // 当班结束时间
	CreateTime        string // 创建时间(时间戳)
	UpdateTime        string // 更新时间(时间戳)
	DeleteTime        string // 删除时间(时间戳)
}

// staffShiftLogColumns holds the columns for the table ttpos_staff_shift_log.
var staffShiftLogColumns = StaffShiftLogColumns{
	Id:                "id",
	Uuid:              "uuid",
	StaffUuid:         "staff_uuid",
	ShiftNo:           "shift_no",
	Status:            "status",
	PreviousShiftCash: "previous_shift_cash",
	CurrentCashTotal:  "current_cash_total",
	Incomes:           "incomes",
	TotalIncome:       "total_income",
	CashTakenOut:      "cash_taken_out",
	CashLeft:          "cash_left",
	CashIncome:        "cash_income",
	TotalBusiness:     "total_business",
	IsPrinted:         "is_printed",
	Remark:            "remark",
	WithdrawCash:      "withdraw_cash",
	DepositCash:       "deposit_cash",
	ExceptionRemark:   "exception_remark",
	Abnormal:          "abnormal",
	ShiftStartTime:    "shift_start_time",
	ShiftEndTime:      "shift_end_time",
	CreateTime:        "create_time",
	UpdateTime:        "update_time",
	DeleteTime:        "delete_time",
}

// NewStaffShiftLogDao creates and returns a new DAO object for table data access.
func NewStaffShiftLogDao(handlers ...gdb.ModelHandler) *StaffShiftLogDao {
	return &StaffShiftLogDao{
		group:    "default",
		table:    "ttpos_staff_shift_log",
		columns:  staffShiftLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *StaffShiftLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *StaffShiftLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *StaffShiftLogDao) Columns() StaffShiftLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *StaffShiftLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *StaffShiftLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *StaffShiftLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
