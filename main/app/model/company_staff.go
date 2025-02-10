package model

// CompanyStaff saas库保存的集团员工关联表
type CompanyStaff struct {
	ID          int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:自增ID" json:"id"`
	Uuid        uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:员工ID;NOT NULL" json:"uuid"`
	CompanyUuid uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"company_uuid"`
	Username    string `gorm:"column:username;type:varchar(255);comment:员工账号;NOT NULL" json:"username"`
	Phone       string `gorm:"column:phone;type:varchar(255);comment:员工手机号;NOT NULL" json:"phone"`
	CreateTime  int64  `gorm:"autoCreateTime;comment:创建时间（时间戳）"`
	UpdateTime  int64  `gorm:"autoUpdateTime;comment:更新时间（时间戳）"`
	DeleteTime  int64  `gorm:"default:0;comment:删除时间（时间戳）"`

	Company *Company `gorm:"foreignKey:company_id;references:id"`
}
