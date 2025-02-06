package model

// CompanyStaff saas库保存的集团员工关联表
type CompanyStaff struct {
	StaffId    uint   `gorm:"primaryKey;default:0;comment:员工id"`
	CompanyId  uint   `gorm:"default:0;comment:集团id"`
	Username   string `gorm:"default:'';comment:员工名称"`
	Phone      string `gorm:"default:'';comment:员工手机号"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）"`
	DeleteTime int64  `gorm:"default:0;comment:删除时间（时间戳）"`

	Company *Company `gorm:"foreignKey:company_id;references:id"`
}
