// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductPackageAttribute is the golang structure for table product_package_attribute.
type ProductPackageAttribute struct {
	Id                               uint   `json:"id"                               orm:"id"                                   description:"自增ID"`            // 自增ID
	Uuid                             uint64 `json:"uuid"                             orm:"uuid"                                 description:"商品包属性ID"`         // 商品包属性ID
	ProductPackageAttributeGroupUuid uint64 `json:"productPackageAttributeGroupUuid" orm:"product_package_attribute_group_uuid" description:"商品包属性组ID"`        // 商品包属性组ID
	AttributeUuid                    uint64 `json:"attributeUuid"                    orm:"attribute_uuid"                       description:"商品属性ID"`          // 商品属性ID
	IsDefaultSelected                int    `json:"isDefaultSelected"                orm:"is_default_selected"                  description:"是否默认选中, 0-否 1-是"` // 是否默认选中, 0-否 1-是
	CreateTime                       uint   `json:"createTime"                       orm:"create_time"                          description:"创建时间(时间戳)"`       // 创建时间(时间戳)
	UpdateTime                       uint   `json:"updateTime"                       orm:"update_time"                          description:"更新时间(时间戳)"`       // 更新时间(时间戳)
	DeleteTime                       uint   `json:"deleteTime"                       orm:"delete_time"                          description:"删除时间(时间戳)"`       // 删除时间(时间戳)
}
