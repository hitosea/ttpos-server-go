// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingCouponRecord is the golang structure for table marketing_coupon_record.
type MarketingCouponRecord struct {
	Id           uint   `json:"id"           orm:"id"            description:""`                                                                      //
	Uuid         int64  `json:"uuid"         orm:"uuid"          description:"优惠券记录唯一ID"`                                                             // 优惠券记录唯一ID
	CouponUuid   int64  `json:"couponUuid"   orm:"coupon_uuid"   description:"优惠券唯一ID"`                                                               // 优惠券唯一ID
	ActivityUuid int64  `json:"activityUuid" orm:"activity_uuid" description:"活动uuid"`                                                                // 活动uuid
	SerialNo     string `json:"serialNo"     orm:"serial_no"     description:"记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用"` // 记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用
	Type         int    `json:"type"         orm:"type"          description:"记录类型：1-首次添加、2-调整添加、3-调整扣减"`                                             // 记录类型：1-首次添加、2-调整添加、3-调整扣减
	Count        int    `json:"count"        orm:"count"         description:"变动数量"`                                                                  // 变动数量
	LeftCount    int    `json:"leftCount"    orm:"left_count"    description:"剩余有效张数"`                                                                // 剩余有效张数
	CreateTime   int    `json:"createTime"   orm:"create_time"   description:"创建时间"`                                                                  // 创建时间
	UpdateTime   int    `json:"updateTime"   orm:"update_time"   description:"更新时间"`                                                                  // 更新时间
	DeleteTime   int    `json:"deleteTime"   orm:"delete_time"   description:"删除时间"`                                                                  // 删除时间
	MemberUuid   int64  `json:"memberUuid"   orm:"member_uuid"   description:"会员uuid"`                                                                // 会员uuid
}
