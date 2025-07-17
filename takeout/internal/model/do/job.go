// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Job is the golang structure of table takeout_job for DAO operations like Where/Data.
type Job struct {
	g.Meta               `orm:"table:takeout_job, do:true"`
	Id                   interface{} //
	Uuid                 interface{} // 外送订单UUID
	ShopRefNo            interface{} // 餐馆订单参考，如UUID
	CustomerMobile       interface{} // 下单客户电话
	CustomerEmail        interface{} // 下单客户联系邮件
	ProviderName         interface{} // 外送供应商： skootar,grab
	TakeoutRefNo         interface{} // 外送系统订单号
	ShopLocationUuid     interface{} // 餐馆位置信息
	ConsumerLocationUuid interface{} // 消费者位置信息
	JobDate              interface{} // 订单日期:'YYYY-MM-DD'
	StartTime            interface{} // Start time. Format in 24 hr (00:00 to 23:59) or "now" for immediate job
	FinishTime           interface{} // 订单结束时间
	PaymentType          interface{} // 支付类型 Payment solution is 3 choice is ""invoice"", ""cash"", ""creditcard"",""prepaid""
	JobStatus            interface{} // 外送订单状态
	Remark               interface{} // 订单备注
	Reserved1            interface{} // 保留字段1
	Reserved2            interface{} // 保留字段2
	CreatedAt            *gtime.Time // 创建时间
	UpdatedAt            *gtime.Time // 更新时间
	DeletedAt            *gtime.Time // 软删除
	CallbackUrl          interface{} // 订单状态更新回调
	SkootarId            interface{} // 骑手Id
	SkootarName          interface{} // 骑手名称
	SkootarPhone         interface{} // 骑手电话
	SkootarImageUrl      interface{} // 骑手头像
	SkootarRating        interface{} // 骑手评分
}
