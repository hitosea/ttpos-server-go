// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleBillSetting is the golang structure of table ttpos_sale_bill_setting for DAO operations like Where/Data.
type SaleBillSetting struct {
	g.Meta           `orm:"table:ttpos_sale_bill_setting, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 销售账单设置ID
	SaleBillUuid     interface{} // 销售账单ID
	ServiceFeeType   interface{} // 服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费。如果服务费收费应用范围不包括该账单，则该账单的服务费类型为0
	ServiceFeeValue  interface{} // 服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例
	TaxFeeType       interface{} // 税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税
	ZeroRule         interface{} // 优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
	ZeroCheckoutRule interface{} // 结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元
	IsStatGift       interface{} // 是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣
	IsStatFree       interface{} // 是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费
	DiscountType     interface{} // 打折类型, 0-百分比打折% 1-百分比直接减免% off
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
