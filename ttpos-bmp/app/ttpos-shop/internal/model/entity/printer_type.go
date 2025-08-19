// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterType is the golang structure for table printer_type.
type PrinterType struct {
	Id                    uint   `json:"id"                    orm:"id"                       description:"自增ID"`                  // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                     description:"打印机类型ID"`               // 打印机类型ID
	Name                  string `json:"name"                  orm:"name"                     description:"打印机类型名称"`               // 打印机类型名称
	MultiLanguageNameUuid uint64 `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`               // 多语言名称ID
	Key                   string `json:"key"                   orm:"key"                      description:"打印机类型key"`              // 打印机类型key
	ConfigJson            string `json:"configJson"            orm:"config_json"              description:"打印机类型json配置,描述需要填写的字段"` // 打印机类型json配置,描述需要填写的字段
	CreateTime            uint   `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`             // 创建时间(时间戳)
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`             // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`             // 删除时间(时间戳)
}
