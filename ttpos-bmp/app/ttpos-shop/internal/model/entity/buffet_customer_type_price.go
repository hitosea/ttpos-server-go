// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BuffetCustomerTypePrice is the golang structure for table buffet_customer_type_price.
type BuffetCustomerTypePrice struct {
	Id                uint    `json:"id"                orm:"id"                  description:"自增ID"`        // 自增ID
	Uuid              uint64  `json:"uuid"              orm:"uuid"                description:"自助餐顾客类型价格ID"` // 自助餐顾客类型价格ID
	BuffetPackageUuid uint64  `json:"buffetPackageUuid" orm:"buffet_package_uuid" description:"自助餐套餐ID"`     // 自助餐套餐ID
	CustomerTypeUuid  uint64  `json:"customerTypeUuid"  orm:"customer_type_uuid"  description:"客户类型ID"`      // 客户类型ID
	Price             float64 `json:"price"             orm:"price"               description:"价格"`          // 价格
	CreateTime        uint    `json:"createTime"        orm:"create_time"         description:"创建时间(时间戳)"`   // 创建时间(时间戳)
	UpdateTime        uint    `json:"updateTime"        orm:"update_time"         description:"更新时间(时间戳)"`   // 更新时间(时间戳)
	DeleteTime        uint    `json:"deleteTime"        orm:"delete_time"         description:"删除时间(时间戳)"`   // 删除时间(时间戳)
}
