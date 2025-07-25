// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// JobLocation is the golang structure for table job_location.
type JobLocation struct {
	Id           int64       `json:"id"           orm:"id"            description:"主键"`               // 主键
	Uuid         string      `json:"uuid"         orm:"uuid"          description:"全局唯一uuid"`         // 全局唯一uuid
	LocationType int         `json:"locationType" orm:"location_type" description:"位置类型： 0 餐馆，1 消费者"` // 位置类型： 0 餐馆，1 消费者
	AddressName  string      `json:"addressName"  orm:"address_name"  description:"地址说明"`             // 地址说明
	Address      string      `json:"address"      orm:"address"       description:"详细地址"`             // 详细地址
	Lat          string      `json:"lat"          orm:"lat"           description:"纬度"`               // 纬度
	Lng          string      `json:"lng"          orm:"lng"           description:"经度"`               // 经度
	ContactName  string      `json:"contactName"  orm:"contact_name"  description:"联系人名称"`            // 联系人名称
	ContactPhone string      `json:"contactPhone" orm:"contact_phone" description:"联系人号码"`            // 联系人号码
	Seq          int         `json:"seq"          orm:"seq"           description:"地址序列，1开始"`         // 地址序列，1开始
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`               // 备注
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`             // 创建时间
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`             // 更新时间
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"软删除"`              // 软删除
}
