// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// OrderSkootar is the golang structure for table order_skootar.
type OrderSkootar struct {
	Id              int64   `json:"id"              orm:"id"                description:"主键ID"`                           // 主键ID
	Uuid            string  `json:"uuid"            orm:"uuid"              description:"唯一标识"`                           // 唯一标识
	OrderUuid       string  `json:"orderUuid"       orm:"order_uuid"        description:"关联主订单UUID (takeout_order.uuid)"` // 关联主订单UUID (takeout_order.uuid)
	SkootarId       string  `json:"skootarId"       orm:"skootar_id"        description:"骑手ID"`                           // 骑手ID
	SkootarName     string  `json:"skootarName"     orm:"skootar_name"      description:"骑手名称"`                           // 骑手名称
	SkootarPhone    string  `json:"skootarPhone"    orm:"skootar_phone"     description:"骑手电话"`                           // 骑手电话
	SkootarRating   float64 `json:"skootarRating"   orm:"skootar_rating"    description:"骑手评分"`                           // 骑手评分
	SkootarImageUrl string  `json:"skootarImageUrl" orm:"skootar_image_url" description:"骑手头像"`                           // 骑手头像
	CreatedAt       int     `json:"createdAt"       orm:"created_at"        description:"创建时间"`                           // 创建时间
	UpdatedAt       int     `json:"updatedAt"       orm:"updated_at"        description:"更新时间"`                           // 更新时间
	DeletedAt       int     `json:"deletedAt"       orm:"deleted_at"        description:"软删除"`                            // 软删除
}
