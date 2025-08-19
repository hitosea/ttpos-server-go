// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductMustPlanRegion is the golang structure for table product_must_plan_region.
type ProductMustPlanRegion struct {
	Id                  uint   `json:"id"                  orm:"id"                     description:"自增ID"`           // 自增ID
	Uuid                uint64 `json:"uuid"                orm:"uuid"                   description:"商品必选商品计划区域明细ID"` // 商品必选商品计划区域明细ID
	ProductMustPlanUuid uint64 `json:"productMustPlanUuid" orm:"product_must_plan_uuid" description:"商品必选商品计划ID"`     // 商品必选商品计划ID
	DeskRegionUuid      uint64 `json:"deskRegionUuid"      orm:"desk_region_uuid"       description:"桌台区域ID"`         // 桌台区域ID
	CreateTime          uint   `json:"createTime"          orm:"create_time"            description:"创建时间(时间戳)"`      // 创建时间(时间戳)
	UpdateTime          uint   `json:"updateTime"          orm:"update_time"            description:"更新时间(时间戳)"`      // 更新时间(时间戳)
	DeleteTime          uint   `json:"deleteTime"          orm:"delete_time"            description:"删除时间(时间戳)"`      // 删除时间(时间戳)
}
