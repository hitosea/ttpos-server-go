// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductCategory is the golang structure for table product_category.
type ProductCategory struct {
	Id                    uint   `json:"id"                    orm:"id"                       description:"自增ID"`          // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                     description:"商品类别ID"`        // 商品类别ID
	Name                  string `json:"name"                  orm:"name"                     description:"名称"`            // 名称
	MultiLanguageNameUuid uint64 `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`       // 多语言名称ID
	Status                int    `json:"status"                orm:"status"                   description:"状态, 1-开启 0-关闭"` // 状态, 1-开启 0-关闭
	ParentUuid            uint64 `json:"parentUuid"            orm:"parent_uuid"              description:"父级ID"`          // 父级ID
	IsSpecial             int    `json:"isSpecial"             orm:"is_special"               description:"特殊分类, 1-是 0-否"` // 特殊分类, 1-是 0-否
	CategoryKey           string `json:"categoryKey"           orm:"category_key"             description:"关键字"`           // 关键字
	Sort                  int    `json:"sort"                  orm:"sort"                     description:"排序"`            // 排序
	CreateTime            uint   `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`     // 创建时间(时间戳)
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`     // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`     // 删除时间(时间戳)
}
