package model

// 供应商表
type Supplier struct {
	Id           uint   `gorm:"primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid         uint64 `gorm:"default:0;comment:'供应商ID'"`
	Name         string `gorm:"default:'';comment:'供应商名称'"`
	Address      string `gorm:"default:'';comment:'供应商地址'"`
	ContactName  string `gorm:"default:'';comment:'联系人姓名'"`
	ContactPhone string `gorm:"default:'';comment:'联系人电话'"`
	Position     string `gorm:"default:'';comment:'职位'"`
	StaffUuid    uint64 `gorm:"default:0;comment:'员工ID, 采购负责人'"`
	CreateTime   int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime   int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime   int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}
