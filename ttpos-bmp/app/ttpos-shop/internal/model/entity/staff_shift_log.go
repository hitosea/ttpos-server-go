// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StaffShiftLog is the golang structure for table staff_shift_log.
type StaffShiftLog struct {
	Id                uint    `json:"id"                orm:"id"                  description:"自增ID"`             // 自增ID
	Uuid              uint64  `json:"uuid"              orm:"uuid"                description:"交班记录ID"`           // 交班记录ID
	StaffUuid         uint64  `json:"staffUuid"         orm:"staff_uuid"          description:"员工ID"`             // 员工ID
	ShiftNo           string  `json:"shiftNo"           orm:"shift_no"            description:"交班编号"`             // 交班编号
	Status            int     `json:"status"            orm:"status"              description:"状态: 0未交班,1已交班"`    // 状态: 0未交班,1已交班
	PreviousShiftCash float64 `json:"previousShiftCash" orm:"previous_shift_cash" description:"上一班遗留备用金"`         // 上一班遗留备用金
	CurrentCashTotal  float64 `json:"currentCashTotal"  orm:"current_cash_total"  description:"当前钱箱现金总计"`         // 当前钱箱现金总计
	Incomes           string  `json:"incomes"           orm:"incomes"             description:"收入详情"`             // 收入详情
	TotalIncome       float64 `json:"totalIncome"       orm:"total_income"        description:"总收入"`              // 总收入
	CashTakenOut      float64 `json:"cashTakenOut"      orm:"cash_taken_out"      description:"本班取出现金"`           // 本班取出现金
	CashLeft          float64 `json:"cashLeft"          orm:"cash_left"           description:"本班遗留备用金"`          // 本班遗留备用金
	CashIncome        float64 `json:"cashIncome"        orm:"cash_income"         description:"本班收入现金"`           // 本班收入现金
	TotalBusiness     float64 `json:"totalBusiness"     orm:"total_business"      description:"本班营业总额(不包含退款)"`    // 本班营业总额(不包含退款)
	IsPrinted         int     `json:"isPrinted"         orm:"is_printed"          description:"是否打印 0-未打印 1-已打印"` // 是否打印 0-未打印 1-已打印
	Remark            string  `json:"remark"            orm:"remark"              description:"备注"`               // 备注
	WithdrawCash      float64 `json:"withdrawCash"      orm:"withdraw_cash"       description:"中途取出现金"`           // 中途取出现金
	DepositCash       float64 `json:"depositCash"       orm:"deposit_cash"        description:"中途存入现金"`           // 中途存入现金
	ExceptionRemark   string  `json:"exceptionRemark"   orm:"exception_remark"    description:"异常报备"`             // 异常报备
	Abnormal          string  `json:"abnormal"          orm:"abnormal"            description:"异常信息-json字符串"`     // 异常信息-json字符串
	ShiftStartTime    uint    `json:"shiftStartTime"    orm:"shift_start_time"    description:"当班开始时间"`           // 当班开始时间
	ShiftEndTime      uint    `json:"shiftEndTime"      orm:"shift_end_time"      description:"当班结束时间"`           // 当班结束时间
	CreateTime        uint    `json:"createTime"        orm:"create_time"         description:"创建时间(时间戳)"`        // 创建时间(时间戳)
	UpdateTime        uint    `json:"updateTime"        orm:"update_time"         description:"更新时间(时间戳)"`        // 更新时间(时间戳)
	DeleteTime        uint    `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"`        // 删除时间(时间戳)
}
