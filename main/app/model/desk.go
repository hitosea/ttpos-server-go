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
	DeskNo      string `gorm:"default:'';column:desk_no;comment:'桌位编号'"`
	Sort        uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	Status      uint   `gorm:"default:0;column:status;comment:'状态, 0-未开台 1-已开台'"`
	IsDisable   uint   `gorm:"default:0;column:is_disable;comment:'是否禁用, 0-否 1-是'"`
	QrcodeToken string `gorm:"default:'';column:qrcode_token;comment:'二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效'"`

	// 关联ID
	RegionUuid   uint64 `gorm:"default:0;column:region_uuid;comment:'桌台区域ID'"`
	TypeUuid     uint64 `gorm:"default:0;column:type_uuid;comment:'桌台类型ID'"`
	DeviceUuid   uint64 `gorm:"default:0;column:device_uuid;comment:'平板设备uuid, 0-未绑定'"`
	SaleBillUuid uint64 `gorm:"default:0;column:sale_bill_uuid;comment:'销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单'"`

	SaleBill *SaleBill `gorm:"foreignKey:desk_uuid;references:uuid"` // 销售账单
}
