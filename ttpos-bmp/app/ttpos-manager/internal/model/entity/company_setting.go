// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CompanySetting is the golang structure for table company_setting.
type CompanySetting struct {
	Id               int    `json:"id"               orm:"id"                  description:"自增ID"`                      // 自增ID
	Uuid             uint64 `json:"uuid"             orm:"uuid"                description:"集团设置ID"`                    // 集团设置ID
	CompanyUuid      int64  `json:"companyUuid"      orm:"company_uuid"        description:"集团ID"`                      // 集团ID
	RealName         string `json:"realName"         orm:"real_name"           description:"真实姓名"`                      // 真实姓名
	LinkName         string `json:"linkName"         orm:"link_name"           description:"联系人"`                       // 联系人
	LinkPhone        string `json:"linkPhone"        orm:"link_phone"          description:"联系电话"`                      // 联系电话
	SaleStock        int    `json:"saleStock"        orm:"sale_stock"          description:"进销存: 0不开启, 1开启"`            // 进销存: 0不开启, 1开启
	IsOpenCoupon     int    `json:"isOpenCoupon"     orm:"is_open_coupon"      description:"是否开启优惠券"`                   // 是否开启优惠券
	IsOpenMarketing  int    `json:"isOpenMarketing"  orm:"is_open_marketing"   description:"是否开启营销活动"`                  // 是否开启营销活动
	IsOpenTax        int    `json:"isOpenTax"        orm:"is_open_tax"         description:"是否开启税务对接: 0不开启, 1奥地利 2-其他"` // 是否开启税务对接: 0不开启, 1奥地利 2-其他
	IsOpenMember     int    `json:"isOpenMember"     orm:"is_open_member"      description:"是否开启会员: 0不开启, 1开启"`         // 是否开启会员: 0不开启, 1开启
	IsOpenTablet     int    `json:"isOpenTablet"     orm:"is_open_tablet"      description:"是否开启平板: 0不开启, 1开启"`         // 是否开启平板: 0不开启, 1开启
	IsOpenH5         int    `json:"isOpenH5"         orm:"is_open_h5"          description:"是否开启扫码H5: 0不开启, 1开启"`       // 是否开启扫码H5: 0不开启, 1开启
	IsOpenAssistant  int    `json:"isOpenAssistant"  orm:"is_open_assistant"   description:"是否开启点餐助手: 0不开启, 1开启"`       // 是否开启点餐助手: 0不开启, 1开启
	IsOpenKitchenKds int    `json:"isOpenKitchenKds" orm:"is_open_kitchen_kds" description:"是否开启后厨KDS: 0不开启, 1开启"`      // 是否开启后厨KDS: 0不开启, 1开启
	IsOpenBuffet     int    `json:"isOpenBuffet"     orm:"is_open_buffet"      description:"是否开启自助餐: 0不开启, 1开启"`        // 是否开启自助餐: 0不开启, 1开启
	EnableSms        int    `json:"enableSms"        orm:"enable_sms"          description:"是否启用短信功能：0-否；1-是"`          // 是否启用短信功能：0-否；1-是
	SmsQuota         int    `json:"smsQuota"         orm:"sms_quota"           description:"短信配额"`                      // 短信配额
	IsOpenH5Order    int    `json:"isOpenH5Order"    orm:"is_open_h5_order"    description:"是否开启扫码点餐接单 0不开启, 1开启"`      // 是否开启扫码点餐接单 0不开启, 1开启
	IsOpenLocalPrint int    `json:"isOpenLocalPrint" orm:"is_open_local_print" description:"是否开启本地打印服务 0不开启, 1开启"`      // 是否开启本地打印服务 0不开启, 1开启
	CashLimit        int    `json:"cashLimit"        orm:"cash_limit"          description:"收银机上限"`                     // 收银机上限
	KitchenLimit     int    `json:"kitchenLimit"     orm:"kitchen_limit"       description:"厨显上限"`                      // 厨显上限
	TabletLimit      int    `json:"tabletLimit"      orm:"tablet_limit"        description:"平板上限"`                      // 平板上限
	AssistantLimit   int    `json:"assistantLimit"   orm:"assistant_limit"     description:"点餐助手上限"`                    // 点餐助手上限
	TableLimit       int    `json:"tableLimit"       orm:"table_limit"         description:"桌台上限"`                      // 桌台上限
	PrinterLimit     int    `json:"printerLimit"     orm:"printer_limit"       description:"打印机上限"`                     // 打印机上限
	Timezone         string `json:"timezone"         orm:"timezone"            description:"时区"`                        // 时区
	Languages        string `json:"languages"        orm:"languages"           description:"支持语言"`                      // 支持语言
	Address          string `json:"address"          orm:"address"             description:"联系地址"`                      // 联系地址
	CreateTime       int    `json:"createTime"       orm:"create_time"         description:"创建时间（时间戳）"`                 // 创建时间（时间戳）
	UpdateTime       int    `json:"updateTime"       orm:"update_time"         description:"更新时间（时间戳）"`                 // 更新时间（时间戳）
	DeleteTime       int    `json:"deleteTime"       orm:"delete_time"         description:"删除时间（时间戳）"`                 // 删除时间（时间戳）
}
