package model

// BindRecord 商家设备绑定记录表
type BindRecord struct {
	ID               uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	FinallyLoginId   int    `gorm:"column:finally_login_id;type:int(11);default:0;comment:最后一个登录id, 退出会清为0" json:"finally_login_id"`
	FinallyLoginTime int    `gorm:"column:finally_login_time;type:int(11);default:0;comment:最后登录时间" json:"finally_login_time"`
	Source           string `gorm:"column:source;type:varchar(255);comment:来源 cashier-收银机 tablet-平板端 kitchen-厨显端" json:"source"`
	DeviceId         string `gorm:"column:key;type:varchar(255);comment:唯一设备标识key" json:"device_id"`
	IsMain           int    `gorm:"column:is_main;type:int(11);default:0;comment:是否主设备 0-常规 1-主" json:"is_main"`
	PrintPortId      int    `gorm:"column:print_port_id;type:int(11);default:0;comment:打印档口ID" json:"print_port_id"`
	Address          string `gorm:"column:address;type:varchar(255);comment:绑定地址" json:"address"`
	Port             int    `gorm:"column:port;type:int(11);default:0;comment:绑定端口" json:"port"`
	DeviceIp         string `gorm:"column:device_ip;type:varchar(50);comment:设备ip" json:"device_ip"`
	Remark           string `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	Brand            string `gorm:"column:brand;type:varchar(255);comment:品牌名称" json:"brand"`
	Platform         string `gorm:"column:platform;type:varchar(50);comment:平台（Web Android iPhone Mobile）" json:"platform"`
	UserAgent        string `gorm:"column:user_agent;type:longtext;comment:请求头信息" json:"user_agent"`
	CreateTime       uint   `gorm:"autoCreateTime;column:create_time;type:int(10) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime       uint   `gorm:"autoUpdateTime;column:update_time;type:int(10) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime       uint   `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}
