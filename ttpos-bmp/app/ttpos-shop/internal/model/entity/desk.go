// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Desk is the golang structure for table desk.
type Desk struct {
	Id             uint   `json:"id"             orm:"id"               description:"自增ID"`                                          // 自增ID
	Uuid           uint64 `json:"uuid"           orm:"uuid"             description:"桌台ID"`                                          // 桌台ID
	DeskNo         string `json:"deskNo"         orm:"desk_no"          description:"桌位编号"`                                          // 桌位编号
	RegionUuid     uint64 `json:"regionUuid"     orm:"region_uuid"      description:"桌台区域ID"`                                        // 桌台区域ID
	TypeUuid       uint64 `json:"typeUuid"       orm:"type_uuid"        description:"桌台类型ID"`                                        // 桌台类型ID
	Sort           int    `json:"sort"           orm:"sort"             description:"排序序号"`                                          // 排序序号
	Status         int    `json:"status"         orm:"status"           description:"状态, 0-未开台 1-已开台"`                               // 状态, 0-未开台 1-已开台
	IsDisable      int    `json:"isDisable"      orm:"is_disable"       description:"是否禁用, 0-否 1-是"`                                 // 是否禁用, 0-否 1-是
	NeedServiceFee int    `json:"needServiceFee" orm:"need_service_fee" description:"是否需要服务费, 0-否 1-是.标记该桌台收取服务费"`                   // 是否需要服务费, 0-否 1-是.标记该桌台收取服务费
	QrcodeToken    string `json:"qrcodeToken"    orm:"qrcode_token"     description:"二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效"`    // 二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效
	SaleBillUuid   uint64 `json:"saleBillUuid"   orm:"sale_bill_uuid"   description:"销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单"` // 销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单
	DeviceUuid     uint64 `json:"deviceUuid"     orm:"device_uuid"      description:"平板设备uuid, 0-未绑定"`                               // 平板设备uuid, 0-未绑定
	CreateTime     uint   `json:"createTime"     orm:"create_time"      description:"创建时间(时间戳)"`                                     // 创建时间(时间戳)
	UpdateTime     uint   `json:"updateTime"     orm:"update_time"      description:"更新时间(时间戳)"`                                     // 更新时间(时间戳)
	DeleteTime     uint   `json:"deleteTime"     orm:"delete_time"      description:"删除时间(时间戳)"`                                     // 删除时间(时间戳)
}
