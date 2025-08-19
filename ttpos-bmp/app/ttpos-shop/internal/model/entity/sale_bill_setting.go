// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SaleBillSetting is the golang structure for table sale_bill_setting.
type SaleBillSetting struct {
	Id               uint    `json:"id"               orm:"id"                 description:"自增ID"`                                                                        // 自增ID
	Uuid             uint64  `json:"uuid"             orm:"uuid"               description:"销售账单设置ID"`                                                                    // 销售账单设置ID
	SaleBillUuid     uint64  `json:"saleBillUuid"     orm:"sale_bill_uuid"     description:"销售账单ID"`                                                                      // 销售账单ID
	ServiceFeeType   int     `json:"serviceFeeType"   orm:"service_fee_type"   description:"服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0"` // 服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0
	ServiceFeeValue  float64 `json:"serviceFeeValue"  orm:"service_fee_value"  description:"服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例"`                                 // 服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例
	TaxFeeType       int     `json:"taxFeeType"       orm:"tax_fee_type"       description:"税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税"`                                               // 税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税
	ZeroRule         int     `json:"zeroRule"         orm:"zero_rule"          description:"优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数"`                            // 优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
	ZeroCheckoutRule int     `json:"zeroCheckoutRule" orm:"zero_checkout_rule" description:"结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元"`                                                 // 结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元
	IsStatGift       int     `json:"isStatGift"       orm:"is_stat_gift"       description:"是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣"`                                      // 是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣
	IsStatFree       int     `json:"isStatFree"       orm:"is_stat_free"       description:"是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费"`                        // 是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费
	DiscountType     int     `json:"discountType"     orm:"discount_type"      description:"打折类型, 0-百分比打折% 1-百分比直接减免% off"`                                               // 打折类型, 0-百分比打折% 1-百分比直接减免% off
	CreateTime       uint    `json:"createTime"       orm:"create_time"        description:"创建时间(时间戳)"`                                                                   // 创建时间(时间戳)
	UpdateTime       uint    `json:"updateTime"       orm:"update_time"        description:"更新时间(时间戳)"`                                                                   // 更新时间(时间戳)
	DeleteTime       uint    `json:"deleteTime"       orm:"delete_time"        description:"删除时间(时间戳)"`                                                                   // 删除时间(时间戳)
}
