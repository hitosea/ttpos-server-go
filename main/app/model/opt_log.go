package model

// ShopOptLog 商家操作日志 ttpos_shop_opt_log
type ShopOptLog struct {
	OptLogId    uint   `gorm:"column:opt_log_id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"opt_log_id"`
	ShopUserId  uint   `gorm:"column:shop_user_id;type:int(10);default:0;comment:用户id;NOT NULL" json:"shop_user_id"`
	Title       string `gorm:"column:title;type:varchar(255);comment:标题" json:"title"`
	Url         string `gorm:"column:url;type:varchar(255);comment:访问url" json:"url"`
	RequestType string `gorm:"column:request_type;type:varchar(50);comment:请求类型" json:"request_type"`
	Browser     string `gorm:"column:browser;type:varchar(255);comment:浏览器" json:"browser"`
	Agent       string `gorm:"column:agent;type:varchar(500);comment:浏览器信息" json:"agent"`
	Content     string `gorm:"column:content;type:longtext;comment:操作内容" json:"content"`
	Ip          string `gorm:"column:ip;type:varchar(128);comment:登录ip;NOT NULL" json:"ip"`
	AppId       uint   `gorm:"column:app_id;type:int(10);default:0;comment:小程序id;NOT NULL" json:"app_id"`
	CreateTime  int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:签到时间;NOT NULL" json:"create_time"`
}
