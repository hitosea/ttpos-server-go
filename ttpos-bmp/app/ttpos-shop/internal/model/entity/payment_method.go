// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PaymentMethod is the golang structure for table payment_method.
type PaymentMethod struct {
	Id                   int     `json:"id"                   orm:"id"                      description:"自增ID"`                       // 自增ID
	Uuid                 int64   `json:"uuid"                 orm:"uuid"                    description:"支付方式ID"`                     // 支付方式ID
	Name                 string  `json:"name"                 orm:"name"                    description:"支付方式名称"`                     // 支付方式名称
	Code                 int     `json:"code"                 orm:"code"                    description:"支付方式代号"`                     // 支付方式代号
	PaymentName          string  `json:"paymentName"          orm:"payment_name"            description:"支付名称"`                       // 支付名称
	Source               int     `json:"source"               orm:"source"                  description:"来源 0-系统 1-手动 2-LianLianPay"` // 来源 0-系统 1-手动 2-LianLianPay
	LogoFileUuid         uint64  `json:"logoFileUuid"         orm:"logo_file_uuid"          description:"logo图片ID"`                   // logo图片ID
	QrcodeFileUuid       uint64  `json:"qrcodeFileUuid"       orm:"qrcode_file_uuid"        description:"二维码图片ID"`                    // 二维码图片ID
	FeePercent           float64 `json:"feePercent"           orm:"fee_percent"             description:"手续费百分比,取值范围0-1"`             // 手续费百分比,取值范围0-1
	IsShowCashier        int     `json:"isShowCashier"        orm:"is_show_cashier"         description:"0-不显示 1-收银机结账显示"`            // 0-不显示 1-收银机结账显示
	IsShowAssistant      int     `json:"isShowAssistant"      orm:"is_show_assistant"       description:"0-不显示 1-点餐助手结账显示"`           // 0-不显示 1-点餐助手结账显示
	IsShowMemberRecharge int     `json:"isShowMemberRecharge" orm:"is_show_member_recharge" description:"0-不显示 1-收银机会员充值显示"`          // 0-不显示 1-收银机会员充值显示
	Status               int     `json:"status"               orm:"status"                  description:"状态 0-禁用 1-启用"`               // 状态 0-禁用 1-启用
	Sort                 int     `json:"sort"                 orm:"sort"                    description:"排序"`                         // 排序
	CreateTime           uint    `json:"createTime"           orm:"create_time"             description:"创建时间(时间戳)"`                  // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"             description:"更新时间(时间戳)"`                  // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"             description:"删除时间(时间戳)"`                  // 删除时间(时间戳)
}
