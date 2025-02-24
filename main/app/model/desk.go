package model

// DeskRegion 餐桌区域表,定义餐桌的区域信息 ttpos_desk_region
type DeskRegion struct {
	BaseModel
	Name string `gorm:"default:'';column:name;comment:'餐桌区域名称'"`
	Sort uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
}

// DeskType 餐桌类型表,定义餐桌的类型信息 ttpos_desk_type
type DeskType struct {
	BaseModel
	Name     string `gorm:"default:'';column:name;comment:'餐桌类型名称'"`
	Sort     uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	RangeMin uint   `gorm:"default:0;column:range_min;comment:'最少人数'"`
	RangeMax uint   `gorm:"default:0;column:range_max;comment:'最多人数'"`
}

// Desk 桌台信息表,定义桌台的相关信息 ttpos_desk
type Desk struct {
	BaseModel
	DeskNo       string `gorm:"column:desk_no;type:varchar(255);comment:桌位编号;NOT NULL" json:"desk_no"`
	RegionUuid   uint64 `gorm:"column:region_uuid;type:bigint(20) unsigned;default:0;comment:桌台区域ID;NOT NULL" json:"region_uuid"`
	TypeUuid     uint64 `gorm:"column:type_uuid;type:bigint(20) unsigned;default:0;comment:桌台类型ID;NOT NULL" json:"type_uuid"`
	Sort         uint   `gorm:"column:sort;type:int(11);default:0;comment:排序序号;NOT NULL" json:"sort"`
	Status       uint   `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-未开台 1-已开台;NOT NULL" json:"status"`
	IsDisable    uint   `gorm:"column:is_disable;type:tinyint(1);default:1;comment:是否禁用, 0-否 1-是;NOT NULL" json:"is_disable"`
	QrcodeToken  string `gorm:"column:qrcode_token;type:varchar(255);comment:二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效;NOT NULL" json:"qrcode_token"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单;NOT NULL" json:"sale_bill_uuid"`
	DeviceUuid   uint64 `gorm:"column:device_uuid;type:bigint(20) unsigned;default:0;comment:平板设备uuid, 0-未绑定;NOT NULL" json:"device_uuid"`

	SaleBill *SaleBill `gorm:"foreignKey:SaleBillUuid;references:uuid"` // 销售账单
	Device   *Device   `gorm:"foreignKey:DeviceUuid;references:uuid"`   // 关联绑定设备
}
