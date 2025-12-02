// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DeviceOnline is the golang structure for table device_online.
type DeviceOnline struct {
	Id                int64  `json:"id"                orm:"id"                  description:"主键ID"`                               // 主键ID
	CompanyUuid       int64  `json:"companyUuid"       orm:"company_uuid"        description:"公司UUID"`                             // 公司UUID
	StaffUuid         int64  `json:"staffUuid"         orm:"staff_uuid"          description:"员工UUID"`                             // 员工UUID
	DeviceId          string `json:"deviceId"          orm:"device_id"           description:"设备ID"`                               // 设备ID
	SourceClient      string `json:"sourceClient"      orm:"source_client"       description:"来源客户端：pos/tablet/kitchen/h5/mobile"` // 来源客户端：pos/tablet/kitchen/h5/mobile
	ConnectionKey     string `json:"connectionKey"     orm:"connection_key"      description:"连接唯一标识"`                             // 连接唯一标识
	Status            int    `json:"status"            orm:"status"              description:"在线状态：0-离线，1-在线"`                     // 在线状态：0-离线，1-在线
	ConnectTime       int    `json:"connectTime"       orm:"connect_time"        description:"连接时间"`                               // 连接时间
	DisconnectTime    int    `json:"disconnectTime"    orm:"disconnect_time"     description:"断开时间"`                               // 断开时间
	LastHeartbeatTime int    `json:"lastHeartbeatTime" orm:"last_heartbeat_time" description:"最后心跳时间"`                             // 最后心跳时间
	IpAddress         string `json:"ipAddress"         orm:"ip_address"          description:"IP地址"`                               // IP地址
	UserAgent         string `json:"userAgent"         orm:"user_agent"          description:"用户代理信息"`                             // 用户代理信息
	CreateTime        int    `json:"createTime"        orm:"create_time"         description:"创建时间"`                               // 创建时间
	UpdateTime        int    `json:"updateTime"        orm:"update_time"         description:"更新时间"`                               // 更新时间
	DeleteTime        int    `json:"deleteTime"        orm:"delete_time"         description:"删除时间（软删除）"`                          // 删除时间（软删除）
}
