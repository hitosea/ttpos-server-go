// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CompanySetting is the golang structure of table ttpos_company_setting for DAO operations like Where/Data.
type CompanySetting struct {
	g.Meta           `orm:"table:ttpos_company_setting, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 集团设置ID
	CompanyUuid      interface{} // 集团ID
	RealName         interface{} // 真实姓名
	LinkName         interface{} // 联系人
	LinkPhone        interface{} // 联系电话
	SaleStock        interface{} // 进销存: 0不开启, 1开启
	IsOpenCoupon     interface{} // 是否开启优惠券
	IsOpenMarketing  interface{} // 是否开启营销活动
	IsOpenTax        interface{} // 是否开启税务对接: 0不开启, 1奥地利 2-其他
	IsOpenMember     interface{} // 是否开启会员: 0不开启, 1开启
	IsOpenTablet     interface{} // 是否开启平板: 0不开启, 1开启
	IsOpenH5         interface{} // 是否开启扫码H5: 0不开启, 1开启
	IsOpenAssistant  interface{} // 是否开启点餐助手: 0不开启, 1开启
	IsOpenKitchenKds interface{} // 是否开启后厨KDS: 0不开启, 1开启
	IsOpenBuffet     interface{} // 是否开启自助餐: 0不开启, 1开启
	EnableSms        interface{} // 是否启用短信功能：0-否；1-是
	SmsQuota         interface{} // 短信配额
	IsOpenH5Order    interface{} // 是否开启扫码点餐接单 0不开启, 1开启
	IsOpenLocalPrint interface{} // 是否开启本地打印服务 0不开启, 1开启
	CashLimit        interface{} // 收银机上限
	KitchenLimit     interface{} // 厨显上限
	TabletLimit      interface{} // 平板上限
	AssistantLimit   interface{} // 点餐助手上限
	TableLimit       interface{} // 桌台上限
	PrinterLimit     interface{} // 打印机上限
	Timezone         interface{} // 时区
	Languages        interface{} // 支持语言
	Address          interface{} // 联系地址
	CreateTime       interface{} // 创建时间（时间戳）
	UpdateTime       interface{} // 更新时间（时间戳）
	DeleteTime       interface{} // 删除时间（时间戳）
}
