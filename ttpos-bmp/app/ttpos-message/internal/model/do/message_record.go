// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MessageRecord is the golang structure of table message_record for DAO operations like Where/Data.
type MessageRecord struct {
	g.Meta       `orm:"table:message_record, do:true"`
	Id           interface{} // 消息ID
	Uuid         interface{} // 消息UUID
	TemplateId   interface{} // 模板ID
	TemplateUuid interface{} // 模板UUID
	MessageType  interface{} // 消息类型(email/sms)
	Recipient    interface{} // 接收人
	Subject      interface{} // 消息主题
	Content      interface{} // 消息内容(渲染后)
	MessageArgs  interface{} // 消息参数
	Status       interface{} // 状态(0-待发送,1-发送中,2-发送成功,3-发送失败)
	ErrorMessage interface{} // 错误信息
	RetryCount   interface{} // 重试次数
	CompanyUuid  interface{} // 公司UUID
	OperatorUuid interface{} // 操作人UUID
	SendTime     interface{} // 发送时间
	CreatedAt    interface{} // 创建时间
	UpdatedAt    interface{} // 更新时间
	DeletedAt    interface{} // 删除时间
}
