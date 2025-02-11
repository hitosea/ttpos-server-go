package model

// 供应商表 ttpos_supplier
type Supplier struct {
	ID           uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid         uint64 `gorm:"default:0;column:uuid;comment:'供应商ID'"`
	Name         string `gorm:"default:'';column:name;comment:'供应商名称'"`
	Address      string `gorm:"default:'';column:address;comment:'供应商地址'"`
	ContactName  string `gorm:"default:'';column:contact_name;comment:'联系人姓名'"`
	ContactPhone string `gorm:"default:'';column:contact_phone;comment:'联系人电话'"`
	Position     string `gorm:"default:'';column:position;comment:'职位'"`
	StaffUuid    uint64 `gorm:"default:0;column:staff_uuid;comment:'员工ID, 采购负责人'"`
	CreateTime   int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间（时间戳）'"`
	UpdateTime   int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间（时间戳）'"`
	DeleteTime   int64  `gorm:"default:0;column:delete_time;comment:'删除时间（时间戳）'"`
}
