// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Supplier is the golang structure for table supplier.
type Supplier struct {
	Id           uint   `json:"id"           orm:"id"            description:"自增ID"`        // 自增ID
	Uuid         uint64 `json:"uuid"         orm:"uuid"          description:"供应商ID"`       // 供应商ID
	Name         string `json:"name"         orm:"name"          description:"供应商名称"`       // 供应商名称
	Address      string `json:"address"      orm:"address"       description:"供应商地址"`       // 供应商地址
	ContactName  string `json:"contactName"  orm:"contact_name"  description:"联系人姓名"`       // 联系人姓名
	ContactPhone string `json:"contactPhone" orm:"contact_phone" description:"联系人电话"`       // 联系人电话
	Position     string `json:"position"     orm:"position"      description:"职位"`          // 职位
	StaffUuid    uint64 `json:"staffUuid"    orm:"staff_uuid"    description:"员工ID, 采购负责人"` // 员工ID, 采购负责人
	CreateTime   uint   `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"`   // 创建时间(时间戳)
	UpdateTime   uint   `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"`   // 更新时间(时间戳)
	DeleteTime   uint   `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"`   // 删除时间(时间戳)
}
