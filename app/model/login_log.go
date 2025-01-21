package model

import "ttpos-server-go/config"

// LoginLog 管理员登录记录表
type LoginLog struct {
	LoginLogId uint   `gorm:"column:login_log_id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"login_log_id"`
	Username   string `gorm:"column:username;type:varchar(50);comment:用户名;NOT NULL" json:"username"`
	Ip         string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	Result     string `gorm:"column:result;type:varchar(128);comment:登录结果;NOT NULL" json:"result"`
	AppId      uint   `gorm:"column:app_id;type:int(10);default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime uint   `gorm:"autoCreateTime;column:create_time;type:int(10);comment:签到时间;NOT NULL" json:"create_time"`
}

func (LoginLog) TableName() string {
	return config.Database.TablePrefix + "shop_login_log"
}
