// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentMethod is the golang structure of table ttpos_payment_method for DAO operations like Where/Data.
type PaymentMethod struct {
	g.Meta               `orm:"table:ttpos_payment_method, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 支付方式ID
	Name                 interface{} // 支付方式名称
	Code                 interface{} // 支付方式代号
	PaymentName          interface{} // 支付名称
	Source               interface{} // 来源 0-系统 1-手动 2-LianLianPay
	LogoFileUuid         interface{} // logo图片ID
	QrcodeFileUuid       interface{} // 二维码图片ID
	FeePercent           interface{} // 手续费百分比,取值范围0-1
	IsShowCashier        interface{} // 0-不显示 1-收银机结账显示
	IsShowAssistant      interface{} // 0-不显示 1-点餐助手结账显示
	IsShowMemberRecharge interface{} // 0-不显示 1-收银机会员充值显示
	Status               interface{} // 状态 0-禁用 1-启用
	Sort                 interface{} // 排序
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
