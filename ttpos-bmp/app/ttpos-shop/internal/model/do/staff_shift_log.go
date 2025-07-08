// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StaffShiftLog is the golang structure of table ttpos_staff_shift_log for DAO operations like Where/Data.
type StaffShiftLog struct {
	g.Meta            `orm:"table:ttpos_staff_shift_log, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // 交班记录ID
	StaffUuid         interface{} // 员工ID
	ShiftNo           interface{} // 交班编号
	Status            interface{} // 状态: 0未交班,1已交班
	PreviousShiftCash interface{} // 上一班遗留备用金
	CurrentCashTotal  interface{} // 当前钱箱现金总计
	Incomes           interface{} // 收入详情
	TotalIncome       interface{} // 总收入
	CashTakenOut      interface{} // 本班取出现金
	CashLeft          interface{} // 本班遗留备用金
	CashIncome        interface{} // 本班收入现金
	TotalBusiness     interface{} // 本班营业总额(不包含退款)
	IsPrinted         interface{} // 是否打印 0-未打印 1-已打印
	Remark            interface{} // 备注
	WithdrawCash      interface{} // 中途取出现金
	DepositCash       interface{} // 中途存入现金
	ExceptionRemark   interface{} // 异常报备
	Abnormal          interface{} // 异常信息-json字符串
	ShiftStartTime    interface{} // 当班开始时间
	ShiftEndTime      interface{} // 当班结束时间
	CreateTime        interface{} // 创建时间(时间戳)
	UpdateTime        interface{} // 更新时间(时间戳)
	DeleteTime        interface{} // 删除时间(时间戳)
}
