package model

// DeskRegion 餐桌区域表，定义餐桌的区域信息
type DeskRegion struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint   `gorm:"default:0;comment:'餐桌区域ID'"`
	Name       string `gorm:"default:'';comment:'餐桌区域名称'"`
	OrderBy    uint   `gorm:"default:0;comment:'排序序号'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// DeskType 餐桌类型表，定义餐桌的类型信息
type DeskType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint   `gorm:"default:0;comment:'餐桌类型ID'"`
	Name       string `gorm:"default:'';comment:'餐桌类型名称'"`
	OrderBy    uint   `gorm:"default:0;comment:'排序序号'"`
	RangeMin   uint   `gorm:"default:0;comment:'最少人数'"`
	RangeMax   uint   `gorm:"default:0;comment:'最多人数'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

// Desk 桌台信息表，定义桌台的相关信息
type Desk struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint   `gorm:"default:0;comment:'桌台ID'"`
	TableNo    string `gorm:"default:'';comment:'桌位编号'"`
	RegionUuid uint   `gorm:"default:0;comment:'桌台区域ID'"`
	TypeUuid   uint   `gorm:"default:0;comment:'桌台类型ID'"`
	OrderBy    uint   `gorm:"default:0;comment:'排序序号'"`
	Status     uint   `gorm:"default:0;comment:'状态, 0-未开台 1-已开台'"`
	IsDisable  uint   `gorm:"default:0;comment:'是否禁用, 0-否 1-是'"`
	QrcodeUrl  string `gorm:"default:'';comment:'二维码图片URL'"`
	IsBind     uint   `gorm:"default:0;comment:'平板绑定状态 0-否 1-是'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
	// order      SaleOrder `gorm:"foreignKey:multi_language_name_uuid;references:uuid"` // 多语言名称
}
