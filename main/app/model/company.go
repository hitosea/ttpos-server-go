package model

// Company 集团表
type Company struct {
	ID            uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid          uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:集团ID;NOT NULL" json:"uuid"`
	Name          string `gorm:"column:name;type:varchar(255);comment:集团名称;NOT NULL" json:"name"`
	Logo          string `gorm:"column:logo;type:varchar(255);comment:logo;NOT NULL" json:"logo"`
	ExpireTime    int    `gorm:"column:expire_time;type:int(10);default:0;comment:过期时间;not null;NOT NULL" json:"expire_time"`
	AuthDay       int    `gorm:"column:auth_day;type:int(11);default:0;comment:授权时间(天) 0为永不过期;NOT NULL" json:"auth_day"`
	Status        int    `gorm:"column:status;type:tinyint(1);default:1;comment:状态 1-启用 0-禁用;not null;NOT NULL" json:"status"`
	AuthStartTime int    `gorm:"column:auth_start_time;type:int(10);default:0;comment:授权开始时间（时间戳）;NOT NULL" json:"auth_start_time"`
	CreateTime    int    `gorm:"autoCreateTime;column:create_time;type:int(10);comment:创建时间(时间戳);NOT NULL" json:"create_time"`
	UpdateTime    int    `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:更新时间(时间戳);NOT NULL" json:"update_time"`
	DeleteTime    int    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间(时间戳);NOT NULL" json:"delete_time"`

	CompanySetting *CompanySetting `gorm:"foreignKey:CompanyUuid;references:Uuid" json:"company_setting,omitempty"`
}
