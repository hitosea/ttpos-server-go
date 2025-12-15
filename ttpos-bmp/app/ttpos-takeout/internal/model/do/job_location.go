// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// JobLocation is the golang structure of table takeout_job_location for DAO operations like Where/Data.
type JobLocation struct {
	g.Meta       `orm:"table:takeout_job_location, do:true"`
	Id           any         // 主键
	Uuid         any         // 全局唯一uuid
	LocationType any         // 位置类型： 0 餐馆，1 消费者
	AddressName  any         // 地址说明
	Address      any         // 详细地址
	Lat          any         // 纬度
	Lng          any         // 经度
	ContactName  any         // 联系人名称
	ContactPhone any         // 联系人号码
	Seq          any         // 地址序列，1开始
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 软删除
}
