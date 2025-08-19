// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RelatedMaterial is the golang structure for table related_material.
type RelatedMaterial struct {
	Id           uint    `json:"id"           orm:"id"            description:"自增ID"`       // 自增ID
	Uuid         uint64  `json:"uuid"         orm:"uuid"          description:"关联材料ID"`     // 关联材料ID
	RelatedUuid  uint64  `json:"relatedUuid"  orm:"related_uuid"  description:"物料清单BOM的ID"` // 物料清单BOM的ID
	MaterialUuid uint64  `json:"materialUuid" orm:"material_uuid" description:"原料ID"`       // 原料ID
	Num          float64 `json:"num"          orm:"num"           description:"材料用量,可小数"`   // 材料用量,可小数
	CreateTime   uint    `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"`  // 创建时间(时间戳)
	UpdateTime   uint    `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"`  // 更新时间(时间戳)
	DeleteTime   uint    `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"`  // 删除时间(时间戳)
}
