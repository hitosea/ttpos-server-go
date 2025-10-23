// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MessageTemplate is the golang structure of table message_template for DAO operations like Where/Data.
type MessageTemplate struct {
	g.Meta          `orm:"table:message_template, do:true"`
	Id              interface{} // 模板ID
	Uuid            interface{} // 模板UUID
	TemplateName    interface{} // 模板名称
	TemplateType    interface{} // 模板类型(email/sms)
	TemplateSubject interface{} // 模板主题(邮件用)
	TemplateContent interface{} // 模板内容(支持变量)
	TemplateArgs    interface{} // 模板参数定义
	Status          interface{} // 状态(0-禁用,1-启用)
	Remark          interface{} // 备注
	CreatedAt       interface{} // 创建时间
	UpdatedAt       interface{} // 更新时间
	DeletedAt       interface{} // 删除时间
}
