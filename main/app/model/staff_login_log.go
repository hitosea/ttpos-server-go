package model

// StaffLoginLog 管理员登录记录表 `ttpos_staff_login_log`
type StaffLoginLog struct {
	BaseModel
	StaffUuid uint64 `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名;NOT NULL" json:"username"`
	Ip        string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	Result    string `gorm:"column:result;type:varchar(128);comment:登录结果;NOT NULL" json:"result"`
	// CreateTime 签到时间
}
