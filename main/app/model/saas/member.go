package saas_model

import (
	"ttpos-server-go/app/model"
)

// Member 会员信息表 `ttpos_member`
type Member struct {
	model.BaseModel
	CompanyUuid uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;comment:公司ID;NOT NULL" json:"company_uuid"`
	Nickname    string `gorm:"column:nickname;type:varchar(255);comment:昵称;NOT NULL" json:"nickname"`
	Phone       string `gorm:"column:phone;type:varchar(20);comment:电话号码;NOT NULL" json:"phone"`
	IsVisitor   bool   `gorm:"column:is_visitor;type:tinyint(1);default:0;comment:是否游客,0-否 1-是;NOT NULL" json:"is_visitor"`
	DeviceId    string `gorm:"column:device_id;type:varchar(100);comment:设备ID,用于标识游客;NOT NULL" json:"device_id"`
}
