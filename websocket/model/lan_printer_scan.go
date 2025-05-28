package model

type LanPrinterScan struct {
	ID             uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid           uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:uuid;NOT NULL" json:"uuid"`
	Ip             string `gorm:"column:ip;type:varchar(255);comment:ip;NOT NULL" json:"ip"`
	Port           int    `gorm:"column:port;type:int(11) unsigned;default:0;comment:端口;NOT NULL" json:"port"`
	Status         int    `gorm:"column:status;type:int(1);default:0;comment:状态 0-离线 1-在线;NOT NULL" json:"status"`
	Remark         string `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	SourceDeviceSn string `gorm:"column:source_device_sn;type:varchar(255);comment:来源设备SN;NOT NULL" json:"source_device_sn"`
	CreateTime     uint   `gorm:"autoCreateTime;column:create_time;type:int(10) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     uint   `gorm:"autoUpdateTime;column:update_time;type:int(10) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime     uint   `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}
