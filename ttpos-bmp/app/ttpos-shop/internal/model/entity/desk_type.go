// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DeskType is the golang structure for table desk_type.
type DeskType struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"餐桌类型ID"`    // 餐桌类型ID
	Name       string `json:"name"       orm:"name"        description:"餐桌类型名称"`    // 餐桌类型名称
	Sort       int    `json:"sort"       orm:"sort"        description:"排序序号"`      // 排序序号
	RangeMin   int    `json:"rangeMin"   orm:"range_min"   description:"最少人数"`      // 最少人数
	RangeMax   int    `json:"rangeMax"   orm:"range_max"   description:"最多人数"`      // 最多人数
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
