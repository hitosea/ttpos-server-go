package model

// BindRecord 商家设备绑定记录表
type BindRecord struct {
	ID               uint   `gorm:"primary_key;AUTO_INCREMENT"`
	FinallyLoginId   int    `gorm:"default:0;comment:最后一个登录id, 退出会清为0"`
	FinallyLoginTime int    `gorm:"default:0;comment:最后登录时间"`
	Source           string `gorm:"default:'';comment:来源 cashier-收银机 tablet-平板端 kitchen-厨显端"`
	DeviceId         string `gorm:"default:'';comment:唯一设备标识key"`
	IsMain           int    `gorm:"default:0;comment:是否主设备 0-常规 1-主"`
	PrintPortId      int    `gorm:"default:0;comment:打印档口ID"`
	Address          string `gorm:"default:'';comment:绑定地址"`
	Port             int    `gorm:"default:0;comment:绑定端口"`
	DeviceIp         string `gorm:"default:'';comment:设备ip"`
	Remark           string `gorm:"default:'';comment:备注"`
	Brand            string `gorm:"default:'';comment:品牌名称"`
	Platform         string `gorm:"default:'';comment:平台（Web Android iPhone Mobile）"`
	UserAgent        string `gorm:"default:'';comment:请求头信息"`
	CreateTime       uint   `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime       uint   `gorm:"autoUpdateTime;comment:更新时间"`
	DeleteTime       uint   `gorm:"default:0;comment:删除时间"`
}
