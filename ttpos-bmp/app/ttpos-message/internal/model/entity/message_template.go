// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MessageTemplate is the golang structure for table message_template.
type MessageTemplate struct {
	Id              uint64 `json:"id"              orm:"id"               description:"模板ID"`            // 模板ID
	Uuid            string `json:"uuid"            orm:"uuid"             description:"模板UUID"`          // 模板UUID
	TemplateName    string `json:"templateName"    orm:"template_name"    description:"模板名称"`            // 模板名称
	TemplateType    string `json:"templateType"    orm:"template_type"    description:"模板类型(email/sms)"` // 模板类型(email/sms)
	TemplateSubject string `json:"templateSubject" orm:"template_subject" description:"模板主题(邮件用)"`       // 模板主题(邮件用)
	TemplateContent string `json:"templateContent" orm:"template_content" description:"模板内容(支持变量)"`      // 模板内容(支持变量)
	TemplateArgs    string `json:"templateArgs"    orm:"template_args"    description:"模板参数定义"`          // 模板参数定义
	Status          int    `json:"status"          orm:"status"           description:"状态(0-禁用,1-启用)"`   // 状态(0-禁用,1-启用)
	Remark          string `json:"remark"          orm:"remark"           description:"备注"`              // 备注
	CreatedAt       int    `json:"createdAt"       orm:"created_at"       description:"创建时间"`            // 创建时间
	UpdatedAt       int    `json:"updatedAt"       orm:"updated_at"       description:"更新时间"`            // 更新时间
	DeletedAt       int    `json:"deletedAt"       orm:"deleted_at"       description:"删除时间"`            // 删除时间
}
