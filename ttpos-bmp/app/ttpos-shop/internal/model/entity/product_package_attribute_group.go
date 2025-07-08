// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductPackageAttributeGroup is the golang structure for table product_package_attribute_group.
type ProductPackageAttributeGroup struct {
	Id                        uint   `json:"id"                        orm:"id"                           description:"自增ID"`          // 自增ID
	Uuid                      uint64 `json:"uuid"                      orm:"uuid"                         description:"商品包属性组ID"`      // 商品包属性组ID
	IsMust                    int    `json:"isMust"                    orm:"is_must"                      description:"是否必选, 0-否 1-是"` // 是否必选, 0-否 1-是
	MaxSelection              int    `json:"maxSelection"              orm:"max_selection"                description:"最大选择数量"`        // 最大选择数量
	ProductPackageUuid        uint64 `json:"productPackageUuid"        orm:"product_package_uuid"         description:"商品包ID"`         // 商品包ID
	ProductAttributeGroupUuid uint64 `json:"productAttributeGroupUuid" orm:"product_attribute_group_uuid" description:"商品属性组ID"`       // 商品属性组ID
	CreateTime                uint   `json:"createTime"                orm:"create_time"                  description:"创建时间(时间戳)"`     // 创建时间(时间戳)
	UpdateTime                uint   `json:"updateTime"                orm:"update_time"                  description:"更新时间(时间戳)"`     // 更新时间(时间戳)
	DeleteTime                uint   `json:"deleteTime"                orm:"delete_time"                  description:"删除时间(时间戳)"`     // 删除时间(时间戳)
}
