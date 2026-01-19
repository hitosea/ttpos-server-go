package model

// SaasStaff saas库保存的统一账号表 ttpos_staff
type SaasStaff struct {
	BaseModel
	Email               string `gorm:"column:email;type:varchar(255);uniqueIndex;comment:邮箱（全平台唯一）;NOT NULL" json:"email"`
	Phone               string `gorm:"column:phone;type:varchar(20);index;comment:手机号（全平台唯一，允许空字符串）" json:"phone"`
	RealName            string `gorm:"column:real_name;type:varchar(255);comment:姓名;NOT NULL" json:"real_name"`
	Password            string `gorm:"column:password;type:varchar(255);comment:登录密码（加密）;NOT NULL" json:"-"`
	PasswordChangeCount int    `gorm:"column:password_change_count;type:int;default:0;comment:修改密码次数" json:"password_change_count"`
	PasswordChangeTime  int64  `gorm:"column:password_change_time;type:int unsigned;default:0;comment:修改密码时间;NOT NULL" json:"password_change_time"`
	IsDisable           int    `gorm:"column:is_disable;type:int;default:0;comment:是否禁用1禁用,0未禁用;NOT NULL" json:"is_disable"`
	LastCompanyUuid     uint64 `gorm:"column:last_company_uuid;type:bigint unsigned;index;default:0;comment:上次登录新管理端的商家UUID;NOT NULL" json:"last_company_uuid"`
}

func (*SaasStaff) TableName() string {
	return "ttpos_staff"
}

// 注意：此表在 saas 数据库中，需要使用 saas 数据库连接
