// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingCouponRecord is the golang structure of table ttpos_marketing_coupon_record for DAO operations like Where/Data.
type MarketingCouponRecord struct {
	g.Meta       `orm:"table:ttpos_marketing_coupon_record, do:true"`
	Id           interface{} //
	Uuid         interface{} // 优惠券记录唯一ID
	CouponUuid   interface{} // 优惠券唯一ID
	ActivityUuid interface{} // 活动uuid
	SerialNo     interface{} // 记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用
	Type         interface{} // 记录类型：1-首次添加、2-调整添加、3-调整扣减
	Count        interface{} // 变动数量
	LeftCount    interface{} // 剩余有效张数
	CreateTime   interface{} // 创建时间
	UpdateTime   interface{} // 更新时间
	DeleteTime   interface{} // 删除时间
	MemberUuid   interface{} // 会员uuid
}
