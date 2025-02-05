package model

// Company 集团表
type Company struct {
	ID            uint   `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT;comment:集团id" json:"id"`
	Name          string `gorm:"column:name;type:varchar(255);comment:集团名称" json:"name"`
	Logo          string `gorm:"column:logo;type:varchar(255);comment:logo" json:"logo"`
	IsRecycle     uint   `gorm:"column:is_recycle;type:tinyint(3);default:0;comment:是否回收;NOT NULL" json:"is_recycle"`
	IsChain       int    `gorm:"column:is_chain;type:tinyint(3);default:1;comment:是否连锁0否1是" json:"is_chain"`
	ExpireTime    int    `gorm:"column:expire_time;type:int(11);default:0;comment:过期时间;NOT NULL" json:"expire_time"`
	AuthDay       int    `gorm:"column:auth_day;type:int(11);default:0;comment:授权时间(天) 0为永不过期" json:"auth_day"`
	AuthStartTime int    `gorm:"column:auth_start_time;type:int(11);default:0;comment:授权开始时间" json:"auth_start_time"`
	Status        int    `gorm:"column:status;type:tinyint(1);default:1;comment:状态1=》启用0禁用;NOT NULL" json:"status"`
	IsDelete      uint   `gorm:"column:is_delete;type:tinyint(3);default:0;comment:是否删除;NOT NULL" json:"is_delete"`
	CreateTime    uint   `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime    uint   `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime    uint   `gorm:"column:delete_time;type:int(11);default:0;comment:删除时间;NOT NULL" json:"delete_time"`

	CompanySetting *CompanySetting
}
