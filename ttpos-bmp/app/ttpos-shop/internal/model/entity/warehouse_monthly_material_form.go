// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WarehouseMonthlyMaterialForm is the golang structure for table warehouse_monthly_material_form.
type WarehouseMonthlyMaterialForm struct {
	Id           uint    `json:"id"           orm:"id"            description:"自增ID"`           // 自增ID
	Uuid         uint64  `json:"uuid"         orm:"uuid"          description:"月度报表uuid"`       // 月度报表uuid
	Year         int     `json:"year"         orm:"year"          description:"年"`              // 年
	Month        int     `json:"month"        orm:"month"         description:"月"`              // 月
	Scene        int     `json:"scene"        orm:"scene"         description:"记录类型,0-月初 1-月末"` // 记录类型,0-月初 1-月末
	MaterialUuid uint64  `json:"materialUuid" orm:"material_uuid" description:"物料uuid"`         // 物料uuid
	Stock        float64 `json:"stock"        orm:"stock"         description:"库存"`             // 库存
	CreateTime   uint    `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"`      // 创建时间(时间戳)
	UpdateTime   uint    `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"`      // 更新时间(时间戳)
	DeleteTime   uint    `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"`      // 删除时间(时间戳)
}
