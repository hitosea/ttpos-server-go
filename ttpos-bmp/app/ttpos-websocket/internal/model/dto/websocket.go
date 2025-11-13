// Package dto 定义数据传输对象
package dto

// PushMessageInput 推送消息输入参数
type PushMessageInput struct {
	CompanyUuid  uint64 `json:"company_uuid"`   // 公司UUID
	StaffUuid    uint64 `json:"staff_uuid"`     // 员工UUID，0表示推送给所有员工
	NotStaffUuid uint64 `json:"not_staff_uuid"` // 排除的员工UUID
	SourceClient string `json:"source_client"`  // 来源客户端
	DeviceId     string `json:"device_id"`      // 设备ID
	NotDeviceId  string `json:"not_device_id"`  // 排除的设备ID
	MessageType  string `json:"message_type"`   // 消息类型
	Data         string `json:"data"`           // 消息数据（JSON格式）
}

// PushMessageOutput 推送消息输出参数
type PushMessageOutput struct {
	Success   bool   `json:"success"`    // 是否成功
	Message   string `json:"message"`    // 响应消息
	PushCount int32  `json:"push_count"` // 成功推送的连接数
}

// GetConnectionStatsInput 获取连接统计输入参数
type GetConnectionStatsInput struct {
	CompanyUuid uint64 `json:"company_uuid"` // 公司UUID，可选
}

// GetConnectionStatsOutput 获取连接统计输出参数
type GetConnectionStatsOutput struct {
	Success          bool             `json:"success"`           // 是否成功
	Message          string           `json:"message"`           // 响应消息
	TotalConnections int32            `json:"total_connections"` // 总连接数
	ByCompany        map[uint64]int32 `json:"by_company"`        // 按公司统计
	BySource         map[string]int32 `json:"by_source"`         // 按来源统计
	ByDevice         map[string]int32 `json:"by_device"`         // 按设备统计
}

// CheckDeviceOnlineInput 检查设备在线输入参数
type CheckDeviceOnlineInput struct {
	CompanyUuid  uint64 `json:"company_uuid"`  // 公司UUID
	SourceClient string `json:"source_client"` // 来源客户端
	DeviceId     string `json:"device_id"`     // 设备ID
}

// CheckDeviceOnlineOutput 检查设备在线输出参数
type CheckDeviceOnlineOutput struct {
	Success       bool   `json:"success"`        // 是否成功
	Message       string `json:"message"`        // 响应消息
	IsOnline      bool   `json:"is_online"`      // 是否在线
	LastHeartbeat string `json:"last_heartbeat"` // 最后心跳时间
}

// CloseConnectionInput 关闭连接输入参数
type CloseConnectionInput struct {
	CompanyUuid  uint64 `json:"company_uuid"`  // 公司UUID
	SourceClient string `json:"source_client"` // 来源客户端，可选
	DeviceId     string `json:"device_id"`     // 设备ID，可选
}

// CloseConnectionOutput 关闭连接输出参数
type CloseConnectionOutput struct {
	Success     bool   `json:"success"`      // 是否成功
	Message     string `json:"message"`      // 响应消息
	ClosedCount int32  `json:"closed_count"` // 关闭的连接数
}
