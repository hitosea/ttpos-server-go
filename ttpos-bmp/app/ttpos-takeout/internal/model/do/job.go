// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Job is the golang structure of table takeout_job for DAO operations like Where/Data.
type Job struct {
	g.Meta               `orm:"table:takeout_job, do:true"`
	Id                   any //
	Uuid                 any // 外送订单UUID
	ShopRefNo            any // 餐馆订单参考，如UUID
	CustomerMobile       any // 下单客户电话
	CustomerEmail        any // 下单客户联系邮件
	ProviderName         any // 外送供应商： skootar,grab
	TakeoutRefNo         any // 外送系统订单号
	ShopLocationUuid     any // 餐馆位置信息
	ConsumerLocationUuid any // 消费者位置信息
	JobDate              any // 订单日期:'YYYY-MM-DD'
	StartTime            any // Start time. Format in 24 hr (00:00 to 23:59) or "now" for immediate job
	FinishTime           any // 订单结束时间
	PaymentType          any // 支付类型 Payment solution is 3 choice is ""invoice"", ""cash"", ""creditcard"",""prepaid""
	JobStatus            any // 外送订单状态
	Remark               any // 订单备注
	Reserved1            any // 保留字段1
	Reserved2            any // 保留字段2
	CallbackUrl          any // 订单状态更新回调
	SkootarId            any // 骑手Id
	SkootarName          any // 骑手名称
	SkootarPhone         any // 骑手电话
	SkootarImageUrl      any // 骑手头像
	SkootarRating        any // 骑手评分
	CreatedAt            any // 创建时间
	UpdatedAt            any // 更新时间
	DeletedAt            any // 软删除
}
