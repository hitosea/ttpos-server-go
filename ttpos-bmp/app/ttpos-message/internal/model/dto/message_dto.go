// Package dto 定义数据传输对象
package dto

// MessageTemplateDTO 消息模板数据传输对象
type MessageTemplateDTO struct {
	Id              uint64 `json:"id" description:"模板ID"`
	Uuid            string `json:"uuid" description:"模板UUID"`
	TemplateName    string `json:"template_name" description:"模板名称"`
	TemplateType    string `json:"template_type" description:"模板类型"`
	TemplateSubject string `json:"template_subject" description:"模板主题"`
	TemplateContent string `json:"template_content" description:"模板内容"`
	TemplateArgs    string `json:"template_args" description:"模板参数定义"`
	Status          int    `json:"status" description:"状态"`
	Remark          string `json:"remark" description:"备注"`
	CreatedAt       int64  `json:"created_at" description:"创建时间"`
	UpdatedAt       int64  `json:"updated_at" description:"更新时间"`
}

// MessageRecordDTO 消息记录数据传输对象
type MessageRecordDTO struct {
	Id           uint64 `json:"id" description:"消息ID"`
	Uuid         string `json:"uuid" description:"消息UUID"`
	TemplateId   uint64 `json:"template_id" description:"模板ID"`
	TemplateUuid string `json:"template_uuid" description:"模板UUID"`
	MessageType  string `json:"message_type" description:"消息类型"`
	Recipient    string `json:"recipient" description:"接收人"`
	Subject      string `json:"subject" description:"消息主题"`
	Content      string `json:"content" description:"消息内容"`
	MessageArgs  string `json:"message_args" description:"消息参数"`
	Status       int    `json:"status" description:"状态"`
	ErrorMessage string `json:"error_message" description:"错误信息"`
	RetryCount   int    `json:"retry_count" description:"重试次数"`
	CompanyUuid  string `json:"company_uuid" description:"公司UUID"`
	OperatorUuid string `json:"operator_uuid" description:"操作人UUID"`
	SendTime     int64  `json:"send_time" description:"发送时间"`
	CreatedAt    int64  `json:"created_at" description:"创建时间"`
	UpdatedAt    int64  `json:"updated_at" description:"更新时间"`
}

// MessageSendLogDTO 消息发送日志数据传输对象
type MessageSendLogDTO struct {
	Id           uint64 `json:"id" description:"日志ID"`
	MessageUuid  string `json:"message_uuid" description:"消息UUID"`
	SendTime     int64  `json:"send_time" description:"发送时间"`
	SendResult   int    `json:"send_result" description:"发送结果"`
	ErrorMessage string `json:"error_message" description:"错误信息"`
	RequestData  string `json:"request_data" description:"请求数据"`
	ResponseData string `json:"response_data" description:"响应数据"`
	CreatedAt    int64  `json:"created_at" description:"创建时间"`
}

// SendMessageInput 发送消息输入参数
type SendMessageInput struct {
	MessageUuid  string `json:"message_uuid" v:"required#消息UUID不能为空"`
	TemplateId   uint64 `json:"template_id"`
	TemplateUuid string `json:"template_uuid" v:"required#模板UUID不能为空"`
	MessageArgs  string `json:"message_args" v:"json#消息参数必须是有效的JSON格式"`
	MessageType  string `json:"message_type" v:"required|in:email,sms#消息类型不能为空|消息类型只能是email或sms"`
	Recipient    string `json:"recipient" v:"required#接收人不能为空"`
	Subject      string `json:"subject"`
	CompanyUuid  string `json:"company_uuid"`
	OperatorUuid string `json:"operator_uuid"`
}

// SendMessageOutput 发送消息输出参数
type SendMessageOutput struct {
	Success     bool   `json:"success" description:"是否成功"`
	Message     string `json:"message" description:"响应消息"`
	MessageUuid string `json:"message_uuid" description:"消息UUID"`
}

// GetMessageStatusInput 查询消息状态输入参数
type GetMessageStatusInput struct {
	MessageUuid string `json:"message_uuid" v:"required#消息UUID不能为空"`
}

// GetMessageStatusOutput 查询消息状态输出参数
type GetMessageStatusOutput struct {
	Success     bool              `json:"success" description:"是否成功"`
	Message     string            `json:"message" description:"响应消息"`
	MessageInfo *MessageRecordDTO `json:"message_info" description:"消息信息"`
}

// ResendMessageInput 重发消息输入参数
type ResendMessageInput struct {
	MessageUuid string `json:"message_uuid" v:"required#消息UUID不能为空"`
}

// ResendMessageOutput 重发消息输出参数
type ResendMessageOutput struct {
	Success bool   `json:"success" description:"是否成功"`
	Message string `json:"message" description:"响应消息"`
}

// MailgunSendRequest Mailgun 发送请求
type MailgunSendRequest struct {
	From    string `json:"from" description:"发件人"`
	To      string `json:"to" description:"收件人"`
	Subject string `json:"subject" description:"邮件主题"`
	HTML    string `json:"html" description:"HTML内容"`
	Text    string `json:"text" description:"纯文本内容"`
}

// MailgunSendResponse Mailgun 发送响应
type MailgunSendResponse struct {
	Id      string `json:"id" description:"消息ID"`
	Message string `json:"message" description:"响应消息"`
}

// RocketMQMessage RocketMQ 消息体（优化版：只传递消息UUID，详细内容从数据库获取）
type RocketMQMessage struct {
	MessageUuid string `json:"message_uuid" description:"消息UUID"`
	MessageType string `json:"message_type" description:"消息类型"`
}
