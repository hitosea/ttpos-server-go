// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StaffOperationLog is the golang structure for table staff_operation_log.
type StaffOperationLog struct {
	Id           uint   `json:"id"           orm:"id"            description:"自增ID"`      // 自增ID
	Uuid         uint64 `json:"uuid"         orm:"uuid"          description:"操作日志ID"`    // 操作日志ID
	StaffUuid    uint64 `json:"staffUuid"    orm:"staff_uuid"    description:"员工ID"`      // 员工ID
	Title        string `json:"title"        orm:"title"         description:"标题"`        // 标题
	Url          string `json:"url"          orm:"url"           description:"操作URL"`     // 操作URL
	RequestData  string `json:"requestData"  orm:"request_data"  description:"请求数据"`      // 请求数据
	ResponseData string `json:"responseData" orm:"response_data" description:"响应数据"`      // 响应数据
	Type         string `json:"type"         orm:"type"          description:"操作类型"`      // 操作类型
	Ip           string `json:"ip"           orm:"ip"            description:"操作IP"`      // 操作IP
	Source       string `json:"source"       orm:"source"        description:"操作来源"`      // 操作来源
	Agent        string `json:"agent"        orm:"agent"         description:"操作用户代理"`    // 操作用户代理
	CreateTime   uint   `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime   uint   `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime   uint   `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
