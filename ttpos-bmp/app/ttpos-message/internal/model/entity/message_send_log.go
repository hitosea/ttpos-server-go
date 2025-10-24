// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MessageSendLog is the golang structure for table message_send_log.
type MessageSendLog struct {
	Id           uint64 `json:"id"           orm:"id"            description:"日志ID"`            // 日志ID
	MessageUuid  string `json:"messageUuid"  orm:"message_uuid"  description:"消息UUID"`          // 消息UUID
	SendTime     int    `json:"sendTime"     orm:"send_time"     description:"发送时间"`            // 发送时间
	SendResult   int    `json:"sendResult"   orm:"send_result"   description:"发送结果(0-失败,1-成功)"` // 发送结果(0-失败,1-成功)
	ErrorMessage string `json:"errorMessage" orm:"error_message" description:"错误信息"`            // 错误信息
	RequestData  string `json:"requestData"  orm:"request_data"  description:"请求数据"`            // 请求数据
	ResponseData string `json:"responseData" orm:"response_data" description:"响应数据"`            // 响应数据
	CreatedAt    int    `json:"createdAt"    orm:"created_at"    description:"创建时间"`            // 创建时间
}
