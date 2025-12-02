// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceOnline is the golang structure of table device_online for DAO operations like Where/Data.
type DeviceOnline struct {
	g.Meta            `orm:"table:device_online, do:true"`
	Id                any // 主键ID
	CompanyUuid       any // 公司UUID
	StaffUuid         any // 员工UUID
	DeviceId          any // 设备ID
	SourceClient      any // 来源客户端：pos/tablet/kitchen/h5/mobile
	ConnectionKey     any // 连接唯一标识
	Status            any // 在线状态：0-离线，1-在线
	ConnectTime       any // 连接时间
	DisconnectTime    any // 断开时间
	LastHeartbeatTime any // 最后心跳时间
	IpAddress         any // IP地址
	UserAgent         any // 用户代理信息
	CreateTime        any // 创建时间
	UpdateTime        any // 更新时间
	DeleteTime        any // 删除时间（软删除）
}
