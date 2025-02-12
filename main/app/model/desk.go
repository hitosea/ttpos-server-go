package model

// DeskRegion 餐桌区域表,定义餐桌的区域信息 ttpos_desk_region
type DeskRegion struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"default:0;column:uuid;comment:'餐桌区域ID'"`
	Name       string `gorm:"default:'';column:name;comment:'餐桌区域名称'"`
	Sort       uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

// DeskType 餐桌类型表,定义餐桌的类型信息 ttpos_desk_type
type DeskType struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"default:0;column:uuid;comment:'餐桌类型ID'"`
	Name       string `gorm:"default:'';column:name;comment:'餐桌类型名称'"`
	Sort       uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	RangeMin   uint   `gorm:"default:0;column:range_min;comment:'最少人数'"`
	RangeMax   uint   `gorm:"default:0;column:range_max;comment:'最多人数'"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}

// Desk 桌台信息表,定义桌台的相关信息 ttpos_desk
type Desk struct {
	ID          uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid        uint64 `gorm:"default:0;column:uuid;comment:'桌台ID'"`
	DeskNo      string `gorm:"default:'';column:desk_no;comment:'桌位编号'"`
	RegionUuid  uint64 `gorm:"default:0;column:region_uuid;comment:'桌台区域ID'"`
	TypeUuid    uint64 `gorm:"default:0;column:type_uuid;comment:'桌台类型ID'"`
	Sort        uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	Status      uint   `gorm:"default:0;column:status;comment:'状态, 0-未开台 1-已开台'"`
	IsDisable   uint   `gorm:"default:0;column:is_disable;comment:'是否禁用, 0-否 1-是'"`
	QrcodeToken string `gorm:"default:'';column:qrcode_token;comment:'二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效'"`
	DeviceUuid  uint64 `gorm:"default:0;column:device_uuid;comment:'平板设备uuid, 0-未绑定'"`
	CreateTime  int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime  int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime  int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`

	SaleBill SaleBill `gorm:"foreignKey:desk_uuid;references:uuid"` // 销售账单
}
