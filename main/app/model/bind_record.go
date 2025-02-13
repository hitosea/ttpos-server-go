package model

// Device 商家设备绑定记录表 `ttpos_device`
type Device struct {
	BaseModel
	FinallyLoginUuid uint64 `gorm:"column:finally_login_uuid;type:bigint(20) unsigned;default:0;comment:最后一个登录id, 退出会清为0;NOT NULL" json:"finally_login_uuid"`
	FinallyLoginTime int64  `gorm:"column:finally_login_time;type:int(10) unsigned;default:0;comment:最后登录时间;NOT NULL" json:"finally_login_time"`
	Source           string `gorm:"column:source;type:varchar(255);comment:来源 cashier-收银机 tablet-平板端 kitchen-厨显端;NOT NULL" json:"source"`
	DeviceKey        string `gorm:"column:device_key;type:varchar(255);comment:唯一设备标识key" json:"device_id"`
	IsMain           int    `gorm:"column:is_main;type:tinyint(1);default:0;comment:是否主设备 0-常规 1-主" json:"is_main"`
	PrintPortUuid    uint64 `gorm:"column:print_port_uuid;type:bigint(20) unsigned;default:0;comment:打印档口ID" json:"print_port_uuid"`
	Address          string `gorm:"column:address;type:varchar(255);comment:绑定地址" json:"address"`
	Port             int    `gorm:"column:port;type:int(11);default:0;comment:绑定端口" json:"port"`
	DeviceIp         string `gorm:"column:device_ip;type:varchar(50);comment:设备ip" json:"device_ip"`
	Remark           string `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	Brand            string `gorm:"column:brand;type:varchar(255);comment:品牌名称" json:"brand"`
	Platform         int    `gorm:"column:platform;type:tinyint(1);default:0;comment:平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端" json:"platform"`
	UserAgent        string `gorm:"column:user_agent;type:longtext;comment:请求头信息" json:"user_agent"`
}
