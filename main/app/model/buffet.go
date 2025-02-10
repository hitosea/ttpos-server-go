package model

type BuffetCustomerType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID       uint   `gorm:"default:0;comment:'自助餐客户类型ID'"`
	Name       string `gorm:"default:'';comment:'自助餐客户类型名称'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

type BuffetCustomerTypePrice struct {
	ID                uint    `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID              uint64  `gorm:"default:0;comment:'自助餐顾客类型价格ID'"`
	BuffetPackageUUID uint64  `gorm:"default:0;comment:'自助餐套餐ID'"`
	CustomerTypeUUID  uint64  `gorm:"default:0;comment:'客户类型ID'"`
	Price             float64 `gorm:"type:decimal(12,2);default:0;comment:'价格'"`
	CreateTime        int64   `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime        int64   `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime        int64   `gorm:"default:0;comment:'删除时间（时间戳）'"`
}
