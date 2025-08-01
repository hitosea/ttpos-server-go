// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// JobDriver is the golang structure for table job_driver.
type JobDriver struct {
	Id             int64       `json:"id"             orm:"id"               description:"主键"`       // 主键
	Uuid           string      `json:"uuid"           orm:"uuid"             description:"全局唯一uuid"` // 全局唯一uuid
	JobUuid        string      `json:"jobUuid"        orm:"job_uuid"         description:"外送订单uuid"` // 外送订单uuid
	DriverId       string      `json:"driverId"       orm:"driver_id"        description:"骑手ID"`     // 骑手ID
	DriverName     string      `json:"driverName"     orm:"driver_name"      description:"骑手名称"`     // 骑手名称
	DriverPhone    string      `json:"driverPhone"    orm:"driver_phone"     description:"骑手电话"`     // 骑手电话
	DriverImageUrl string      `json:"driverImageUrl" orm:"driver_image_url" description:"骑手头像"`     // 骑手头像
	Lat            string      `json:"lat"            orm:"lat"              description:"骑手当前纬度"`   // 骑手当前纬度
	Lng            string      `json:"lng"            orm:"lng"              description:"骑手当前经度"`   // 骑手当前经度
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"       description:"创建时间"`     // 创建时间
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"       description:"更新时间"`     // 更新时间
	DeletedAt      *gtime.Time `json:"deletedAt"      orm:"deleted_at"       description:"软删除"`      // 软删除
}
