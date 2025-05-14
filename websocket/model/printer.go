package model

type Printer struct {
	ID                uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid              uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:打印机UUID;NOT NULL" json:"uuid"`
	Name              string `gorm:"column:name;type:varchar(255);comment:打印机名称;NOT NULL" json:"name"`
	PrinterTypeUuid   uint64 `gorm:"column:printer_type_uuid;type:bigint(20) unsigned;default:0;comment:打印机类型ID;NOT NULL" json:"printer_type_uuid"`
	ConfigJson        string `gorm:"column:config_json;type:text;comment:打印机json配置" json:"config_json"`
	Copies            uint   `gorm:"column:copies;type:int(11) unsigned;default:0;comment:打印份数;NOT NULL" json:"copies"`
	Sort              uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序;NOT NULL" json:"sort"`
	IsUsb             int    `gorm:"column:is_usb;type:int(1);default:0;comment:是否usb;NOT NULL" json:"is_usb"`
	Status            int    `gorm:"column:status;type:int(1);default:0;comment:状态 0-离线 1-在线;NOT NULL" json:"status"`
	LastHeartbeatTime uint   `gorm:"column:last_heartbeat_time;type:int(10) unsigned;default:0;comment:最后心跳时间;NOT NULL" json:"last_heartbeat_time"`
	CreateTime        uint   `gorm:"autoCreateTime;column:create_time;type:int(10) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime        uint   `gorm:"autoUpdateTime;column:update_time;type:int(10) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime        uint   `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

// PrinterType 打印机类型信息表 ttpos_printer_type
type PrinterType struct {
	ID                    uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid                  uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:打印机UUID;NOT NULL" json:"uuid"`
	Name                  string `gorm:"column:name;type:varchar(255);comment:打印机类型名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	Key                   string `gorm:"column:key;type:varchar(255);comment:打印机类型key;NOT NULL" json:"key"`
	ConfigJson            string `gorm:"column:config_json;type:text;comment:打印机类型json配置,描述需要填写的字段" json:"config_json"`
	CreateTime            uint   `gorm:"autoCreateTime;column:create_time;type:int(10) unsigned;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime            uint   `gorm:"autoUpdateTime;column:update_time;type:int(10) unsigned;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime            uint   `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}
