// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductUnit is the golang structure for table product_unit.
type ProductUnit struct {
	Id                    uint   `json:"id"                    orm:"id"                       description:"自增ID"`      // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                     description:"商品单位ID"`    // 商品单位ID
	Name                  string `json:"name"                  orm:"name"                     description:"单位名称"`      // 单位名称
	MultiLanguageNameUuid uint64 `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`   // 多语言名称ID
	CreateTime            uint   `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
