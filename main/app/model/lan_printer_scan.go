package model

type LanPrinterScan struct {
	BaseModel
	Ip             string `gorm:"column:ip;type:varchar(255);comment:ip;NOT NULL" json:"ip"`
	Port           int    `gorm:"column:port;type:int(11) unsigned;default:0;comment:端口;NOT NULL" json:"port"`
	Status         int    `gorm:"column:status;type:int(1);default:0;comment:状态 0-离线 1-在线;NOT NULL" json:"status"`
	Remark         string `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	SourceDeviceSn string `gorm:"column:source_device_sn;type:varchar(255);comment:来源设备SN;NOT NULL" json:"source_device_sn"`
}
