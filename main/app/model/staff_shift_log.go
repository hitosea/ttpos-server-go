package model

// StaffShiftLog 员工交班记录表
type StaffShiftLog struct {
	ID                uint    `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid              uint64  `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:交班记录ID;NOT NULL" json:"uuid"`
	StaffUuid         uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"staff_uuid"`
	ShiftNo           string  `gorm:"column:shift_no;type:varchar(64);comment:交班编号;NOT NULL" json:"shift_no"`
	Status            int     `gorm:"column:status;type:int(11);default:1;comment:状态： 0未交班，1已交班;NOT NULL" json:"status"`
	PreviousShiftCash float64 `gorm:"column:previous_shift_cash;type:decimal(12,2);default:0.00;comment:上一班遗留备用金;NOT NULL" json:"previous_shift_cash"`
	CurrentCashTotal  float64 `gorm:"column:current_cash_total;type:decimal(12,2);default:0.00;comment:当前钱箱现金总计;NOT NULL" json:"current_cash_total"`
	Incomes           string  `gorm:"column:incomes;type:varchar(255);comment:收入详情" json:"incomes"`
	TotalIncome       float64 `gorm:"column:total_income;type:decimal(12,2);default:0.00;comment:总收入;NOT NULL" json:"total_income"`
	CashTakenOut      float64 `gorm:"column:cash_taken_out;type:decimal(12,2);default:0.00;comment:本班取出现金;NOT NULL" json:"cash_taken_out"`
	CashLeft          float64 `gorm:"column:cash_left;type:decimal(12,2);default:0.00;comment:本班遗留备用金;NOT NULL" json:"cash_left"`
	CashIncome        float64 `gorm:"column:cash_income;type:decimal(12,2);default:0.00;comment:本班收入现金;NOT NULL" json:"cash_income"`
	TotalBusiness     float64 `gorm:"column:total_business;type:decimal(12,2);default:0.00;comment:本班营业总额（不包含退款）;NOT NULL" json:"total_business"`
	IsPrinted         int     `gorm:"column:is_printed;type:tinyint(1);default:0;comment:是否打印 0-未打印 1-已打印;NOT NULL" json:"is_printed"`
	Remark            string  `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	WithdrawCash      float64 `gorm:"column:withdraw_cash;type:decimal(12,2);default:0.00;comment:中途取出现金;NOT NULL" json:"withdraw_cash"`
	DepositCash       float64 `gorm:"column:deposit_cash;type:decimal(12,2);default:0.00;comment:中途存入现金;NOT NULL" json:"deposit_cash"`
	ExceptionRemark   string  `gorm:"column:exception_remark;type:varchar(255);comment:异常报备;NOT NULL" json:"exception_remark"`
	Abnormal          string  `gorm:"column:abnormal;type:varchar(255);comment:异常信息-json字符串" json:"abnormal"`
	ShiftStartTime    int     `gorm:"column:shift_start_time;type:int(10);default:0;comment:当班开始时间;NOT NULL" json:"shift_start_time"`
	ShiftEndTime      int     `gorm:"column:shift_end_time;type:int(10);default:0;comment:当班结束时间;NOT NULL" json:"shift_end_time"`
	CreateTime        int     `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime        int     `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime        int     `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`
}
