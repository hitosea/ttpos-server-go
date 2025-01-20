package model

// App 应用表
type App struct {
	AppId         uint   `gorm:"column:app_id;type:int(11);primary_key;AUTO_INCREMENT;comment:小程序id" json:"app_id"`
	AppName       string `gorm:"column:app_name;type:varchar(255);comment:应用名称" json:"app_name"`
	Logo          int    `gorm:"column:logo;type:int(11);default:0;comment:logo" json:"logo"`
	IsRecycle     uint   `gorm:"column:is_recycle;type:tinyint(3);default:0;comment:是否回收;NOT NULL" json:"is_recycle"`
	IsChain       int    `gorm:"column:is_chain;type:tinyint(3);default:1;comment:是否连锁0否1是" json:"is_chain"`
	ExpireTime    int    `gorm:"column:expire_time;type:int(11);default:0;comment:过期时间;NOT NULL" json:"expire_time"`
	AuthDay       int    `gorm:"column:auth_day;type:int(11);default:0;comment:授权时间(天) 0为永不过期" json:"auth_day"`
	AuthStartTime int    `gorm:"column:auth_start_time;type:int(11);default:0;comment:授权开始时间" json:"auth_start_time"`
	Status        int    `gorm:"column:status;type:tinyint(1);default:1;comment:状态1=》启用0禁用;NOT NULL" json:"status"`
	IsDelete      uint   `gorm:"column:is_delete;type:tinyint(3);default:0;comment:是否删除;NOT NULL" json:"is_delete"`
	CreateTime    uint   `gorm:"autoCreateTime;column:create_time;type:int(11);comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime    uint   `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`

	Supplier *Supplier `gorm:"foreignKey:shop_supplier_id;references:app_id"`
}
