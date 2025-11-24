package model

const DataManageTypeOrder = 0 // 数据类型 0订单

// Desk 桌台信息表,定义桌台的相关信息 ttpos_desk
type DataManage struct {
	BaseModel
	Type      uint   `gorm:"column:type;type:int(10);default:0;comment:数据类型 0订单;NOT NULL" json:"type"`
	DataUuid  uint64 `gorm:"column:data_uuid;type:bigint(20) unsigned;default:0;comment:数据UUID;NOT NULL" json:"data_uuid"`
	StaffUuid uint64 `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:员工UUID;NOT NULL" json:"staff_uuid"`
}
