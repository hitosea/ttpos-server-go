package model

import "ttpos-server-go/config"

// LoginLog 管理员登录记录表
type LoginLog struct {
	Id         int    `gorm:"column:id;type:int(11);AUTO_INCREMENT;primary_key;comment:自增ID" json:"id"`
	Uuid       int64  `gorm:"column:uuid;type:bigint(20);default:0;comment:UUID;NOT NULL" json:"uuid"`
	StaffUuid  int64  `gorm:"column:staff_uuid;type:bigint(20);default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
	Username   string `gorm:"column:username;type:varchar(50);comment:用户名;NOT NULL" json:"username"`
	Ip         string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	Result     string `gorm:"column:result;type:varchar(128);comment:登录结果;NOT NULL" json:"result"`
	CreateTime uint   `gorm:"autoCreateTime;column:create_time;type:int(10) unsigned;comment:签到时间;NOT NULL" json:"create_time"`
}

func (LoginLog) TableName() string {
	return config.Database.TablePrefix + "shop_login_log"
}
