package model

import "ttpos-server-go/app/constant"

// Staff 员工表 ttpos_staff
type Staff struct {
	BaseModel
	CompanyUuid         uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	Username            string `gorm:"column:username;type:varchar(255);comment:用户名;NOT NULL" json:"username"`
	Password            string `gorm:"column:password;type:varchar(255);comment:登录密码;NOT NULL" json:"password"`
	Phone               string `gorm:"column:phone;type:varchar(20);comment:手机号" json:"phone"`
	PasswordChangeCount int    `gorm:"column:password_change_count;type:int(11);default:0;comment:修改密码次数" json:"password_change_count"`
	PasswordChangeTime  int64  `gorm:"column:password_change_time;type:int(10) unsigned;default:0;comment:修改密码时间;NOT NULL" json:"password_change_time"`
	RealName            string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	IsSuper             int    `gorm:"column:is_super;type:tinyint(3);default:0;comment:是否为超级管理员0不是,1是;NOT NULL" json:"is_super"`
	UserType            int    `gorm:"column:user_type;type:tinyint(1);default:0;comment:账号类型0总台1门店;NOT NULL" json:"user_type"`
	IsDisable           int    `gorm:"column:is_disable;type:tinyint(3);default:0;comment:是否禁用1禁用,0未禁用;NOT NULL" json:"is_disable"`
	BindKey             string `gorm:"column:bind_key;type:varchar(255);comment:绑定的设备key" json:"bind_key"`
	CashierOnline       int    `gorm:"column:cashier_online;type:tinyint(1);default:0;comment:收银员当班 0-不在线 1-在线;NOT NULL" json:"cashier_online"`
	CashierLoginTime    int64  `gorm:"column:cashier_login_time;type:int(11) unsigned;default:0;comment:收银员当班登录时间;NOT NULL" json:"cashier_login_time"`
	DutyNo              string `gorm:"column:duty_no;type:varchar(64);comment:当班编号" json:"duty_no"`

	Company *Company `gorm:"foreignKey:CompanyUuid;references:Uuid"`
	Device  *Device  `gorm:"foreignKey:BindKey;references:DeviceId"`
}

// StaffRole 员工角色关系表 ttpos_staff_role
type StaffRole struct {
	BaseModel
	StaffUuid int64 `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:超管用户ID;NOT NULL" json:"staff_uuid"`
	RoleUuid  int64 `gorm:"column:role_uuid;type:bigint(20);default:0;comment:角色ID;NOT NULL" json:"role_uuid"`
}

// StaffShiftLog 员工交班记录表 ttpos_staff_shift_log
type StaffShiftLog struct {
	BaseModel
	StaffUuid         uint64  `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"staff_uuid"`
	ShiftNo           string  `gorm:"column:shift_no;type:varchar(64);comment:交班编号;NOT NULL" json:"shift_no"`
	Status            int     `gorm:"column:status;type:int(11);default:0;comment:状态： 0未交班,1已交班;NOT NULL" json:"status"`
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
	ShiftStartTime    int64   `gorm:"column:shift_start_time;type:int(10);default:0;comment:当班开始时间;NOT NULL" json:"shift_start_time"`
	ShiftEndTime      int64   `gorm:"column:shift_end_time;type:int(10);default:0;comment:当班结束时间;NOT NULL" json:"shift_end_time"`
}

// StaffLoginLog 管理员登录记录表 `ttpos_staff_login_log`
type StaffLoginLog struct {
	BaseModel
	StaffUuid uint64 `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名;NOT NULL" json:"username"`
	Ip        string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	Result    string `gorm:"column:result;type:varchar(128);comment:登录结果;NOT NULL" json:"result"`
	// CreateTime 签到时间
}

// StaffOperationLog 员工操作日志表 ttpos_staff_operation_log
type StaffOperationLog struct {
	BaseModel
	StaffUuid    uint64 `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"staff_uuid"`
	Title        string `gorm:"column:title;type:varchar(255);comment:标题;NOT NULL" json:"title"`
	Url          string `gorm:"column:url;type:varchar(255);comment:操作URL;NOT NULL" json:"url"`
	RequestData  string `gorm:"column:request_data;type:varchar(255);comment:请求数据;NOT NULL" json:"request_data"`
	ResponseData string `gorm:"column:response_data;type:varchar(255);comment:响应数据;NOT NULL" json:"response_data"`
	Type         string `gorm:"column:type;type:varchar(255);comment:操作类型;NOT NULL" json:"type"`
	Ip           string `gorm:"column:ip;type:varchar(255);comment:操作IP;NOT NULL" json:"ip"`
	Source       string `gorm:"column:source;type:varchar(255);comment:操作来源;NOT NULL" json:"source"`
	Agent        string `gorm:"column:agent;type:varchar(255);comment:操作用户代理;NOT NULL" json:"agent"`
}

// IsHandedOver 是否已交班
func (model *StaffShiftLog) IsHandedOver() bool {
	return model.Status == constant.StaffHandedOver
}

// IsReported 是否已经报备
func (model *StaffShiftLog) IsReported() bool {
	return model.ExceptionRemark != ""
}
