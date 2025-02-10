package model

type BuffetCustomerType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	UUID       uint   `gorm:"default:0;comment:'自助餐客户类型ID'"`
	Name       string `gorm:"default:'';comment:'自助餐客户类型名称'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}
