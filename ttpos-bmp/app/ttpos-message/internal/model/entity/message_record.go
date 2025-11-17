// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MessageRecord is the golang structure for table message_record.
type MessageRecord struct {
	Id           uint64 `json:"id"           orm:"id"            description:"消息ID"`                          // 消息ID
	Uuid         string `json:"uuid"         orm:"uuid"          description:"消息UUID"`                        // 消息UUID
	TemplateId   uint64 `json:"templateId"   orm:"template_id"   description:"模板ID"`                          // 模板ID
	TemplateUuid string `json:"templateUuid" orm:"template_uuid" description:"模板UUID"`                        // 模板UUID
	MessageType  string `json:"messageType"  orm:"message_type"  description:"消息类型(email/sms)"`               // 消息类型(email/sms)
	Recipient    string `json:"recipient"    orm:"recipient"     description:"接收人"`                           // 接收人
	Subject      string `json:"subject"      orm:"subject"       description:"消息主题"`                          // 消息主题
	Content      string `json:"content"      orm:"content"       description:"消息内容(渲染后)"`                     // 消息内容(渲染后)
	MessageArgs  string `json:"messageArgs"  orm:"message_args"  description:"消息参数"`                          // 消息参数
	Status       int    `json:"status"       orm:"status"        description:"状态(0-待发送,1-发送中,2-发送成功,3-发送失败)"` // 状态(0-待发送,1-发送中,2-发送成功,3-发送失败)
	ErrorMessage string `json:"errorMessage" orm:"error_message" description:"错误信息"`                          // 错误信息
	RetryCount   int    `json:"retryCount"   orm:"retry_count"   description:"重试次数"`                          // 重试次数
	CompanyUuid  string `json:"companyUuid"  orm:"company_uuid"  description:"公司UUID"`                        // 公司UUID
	OperatorUuid string `json:"operatorUuid" orm:"operator_uuid" description:"操作人UUID"`                       // 操作人UUID
	SendTime     int    `json:"sendTime"     orm:"send_time"     description:"发送时间"`                          // 发送时间
	CreatedAt    int    `json:"createdAt"    orm:"created_at"    description:"创建时间"`                          // 创建时间
	UpdatedAt    int    `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`                          // 更新时间
	DeletedAt    int    `json:"deletedAt"    orm:"deleted_at"    description:"删除时间"`                          // 删除时间
}
