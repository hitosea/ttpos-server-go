// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Device is the golang structure for table device.
type Device struct {
	Id                 uint   `json:"id"                 orm:"id"                   description:"自增ID"`                                                 // 自增ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"绑定记录ID"`                                               // 绑定记录ID
	FinallyLoginUuid   uint64 `json:"finallyLoginUuid"   orm:"finally_login_uuid"   description:"最后一个登录id, 退出会清为0"`                                     // 最后一个登录id, 退出会清为0
	FinallyLoginTime   uint   `json:"finallyLoginTime"   orm:"finally_login_time"   description:"最后登录时间"`                                               // 最后登录时间
	Source             string `json:"source"             orm:"source"               description:"来源 cashier-收银机 tablet-平板端 kitchen-厨显端"`                // 来源 cashier-收银机 tablet-平板端 kitchen-厨显端
	DeviceId           string `json:"deviceId"           orm:"device_id"            description:"唯一设备标识id"`                                             // 唯一设备标识id
	IsMain             int    `json:"isMain"             orm:"is_main"              description:"是否主设备 0-常规 1-主"`                                       // 是否主设备 0-常规 1-主
	ProductPrinterUuid int64  `json:"productPrinterUuid" orm:"product_printer_uuid" description:"打印档口Uuid"`                                             // 打印档口Uuid
	Address            string `json:"address"            orm:"address"              description:"绑定地址"`                                                 // 绑定地址
	Port               int    `json:"port"               orm:"port"                 description:"绑定端口"`                                                 // 绑定端口
	DeviceIp           string `json:"deviceIp"           orm:"device_ip"            description:"设备ip"`                                                 // 设备ip
	Remark             string `json:"remark"             orm:"remark"               description:"备注"`                                                   // 备注
	Brand              string `json:"brand"              orm:"brand"                description:"品牌名称"`                                                 // 品牌名称
	Platform           int    `json:"platform"           orm:"platform"             description:"平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端"` // 平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端
	UserAgent          string `json:"userAgent"          orm:"user_agent"           description:"请求头信息"`                                                // 请求头信息
	CashSign           string `json:"cashSign"           orm:"cash_sign"            description:"收银终端标识"`                                               // 收银终端标识
	CashBoxId          string `json:"cashBoxId"          orm:"cash_box_id"          description:"现金箱ID"`                                                // 现金箱ID
	AccessToken        string `json:"accessToken"        orm:"access_token"         description:"访问令牌"`                                                 // 访问令牌
	QueueUrl           string `json:"queueUrl"           orm:"queue_url"            description:"关联队列url"`                                              // 关联队列url
	CreateTime         uint   `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`                                            // 创建时间(时间戳)
	UpdateTime         uint   `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`                                            // 更新时间(时间戳)
	DeleteTime         uint   `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`                                            // 删除时间(时间戳)
}
