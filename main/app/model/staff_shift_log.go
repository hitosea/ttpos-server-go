package model

// StaffShiftLog 员工交班记录表
type StaffShiftLog struct {
	ID                uint    `gorm:"primary_key;AUTO_INCREMENT""`
	ShiftUserId       uint    `gorm:"default:0;comment:收银员id""`
	ShiftNo           string  `gorm:"default:'';comment:交班编号""`
	Status            uint    `gorm:"default:1;comment:状态： 0未交班，1已交班""`
	PreviousShiftCash float64 `gorm:"default:0.00;comment:上一班遗留备用金""`
	CurrentCashTotal  float64 `gorm:"default:0.00;comment:当前钱箱现金总计""`
	Incomes           string  `gorm:"type:text;comment:收入详情""`
	TotalIncome       float64 `gorm:"default:0.00;comment:总收入""`
	CashTakenOut      float64 `gorm:"default:0.00;comment:本班取出现金""`
	CashLeft          float64 `gorm:"default:0.00;comment:本班遗留备用金""`
	CashIncome        float64 `gorm:"default:0.00;comment:本班收入现金""`
	TotalBusiness     float64 `gorm:"default:0.00;comment:本班营业总额（不包含退款）""`
	IsPrinted         uint    `gorm:"default:0;comment:是否打印 0-未打印 1-已打印""`
	Remark            string  `gorm:"default:'';comment:备注""`
	WithdrawCash      float64 `gorm:"default:0.00;comment:中途取出现金""`
	DepositCash       float64 `gorm:"default:0.00;comment:中途存入现金""`
	ExceptionRemark   string  `gorm:"default:'';comment:异常报备""`
	Abnormal          string  `gorm:"type:text;comment:异常信息-json字符串""`
	ShiftStartTime    uint    `gorm:"default:0;comment:当班开始时间""`
	ShiftEndTime      uint    `gorm:"default:0;comment:当班结束时间""`
	CreateTime        int64   `gorm:"autoCreateTime;comment:创建时间""`
	UpdateTime        int64   `gorm:"autoUpdateTime;comment:更新时间""`
	DeleteTime        int64   `gorm:"default:0;comment:删除时间""`
}
