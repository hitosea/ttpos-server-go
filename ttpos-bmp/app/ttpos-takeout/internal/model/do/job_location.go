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
	Id           interface{} // 主键
	Uuid         interface{} // 全局唯一uuid
	LocationType interface{} // 位置类型： 0 餐馆，1 消费者
	AddressName  interface{} // 地址说明
	Address      interface{} // 详细地址
	Lat          interface{} // 纬度
	Lng          interface{} // 经度
	ContactName  interface{} // 联系人名称
	ContactPhone interface{} // 联系人号码
	Seq          interface{} // 地址序列，1开始
	Remark       interface{} // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 软删除
}
