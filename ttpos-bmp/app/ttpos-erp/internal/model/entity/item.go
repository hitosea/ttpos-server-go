// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Item 商品实体
type Item struct {
	ID           uint   `orm:"id,primary" json:"id"`               // 主键ID
	Uuid         uint64 `orm:"uuid" json:"uuid"`                   // UUID
	Name         string `orm:"name" json:"name"`                   // 商品名称
	ImageName    string `orm:"image_name" json:"image_name"`       // 图片名称
	Status       uint   `orm:"status" json:"status"`               // 状态, 0-上架 1-下架
	CategoryUuid uint64 `orm:"category_uuid" json:"category_uuid"` // 分类UUID
	CreateTime   int64  `orm:"create_time" json:"create_time"`     // 创建时间
	UpdateTime   int64  `orm:"update_time" json:"update_time"`     // 更新时间
	DeleteTime   int64  `orm:"delete_time" json:"delete_time"`     // 删除时间
}
