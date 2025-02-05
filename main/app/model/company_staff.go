package model

// CompanyStaff saas库保存的集团员工关联表
type CompanyStaff struct {
	StaffId    uint   `gorm:"primaryKey;not null;default:0;comment:员工id" json:"staff_id"`
	CompanyId  uint   `gorm:"not null;default:0;comment:集团id" json:"company_id"`
	Username   string `gorm:"not null;default:'';comment:员工名称" json:"username"`
	Phone      string `gorm:"not null;default:'';comment:员工手机号" json:"phone"`
	CreateTime int    `gorm:"autoCreateTime;not null;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime int    `gorm:"autoUpdateTime;not null;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime int    `gorm:"not null;default:0;comment:删除时间（时间戳）" json:"delete_time"`

	Company *Company `gorm:"foreignKey:company_id;references:id"`
}
