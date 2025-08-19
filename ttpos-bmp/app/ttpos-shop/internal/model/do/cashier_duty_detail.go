// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CashierDutyDetail is the golang structure of table ttpos_cashier_duty_detail for DAO operations like Where/Data.
type CashierDutyDetail struct {
	g.Meta                       `orm:"table:ttpos_cashier_duty_detail, do:true"`
	Id                           interface{} // 自增ID
	Uuid                         interface{} // 收银交班详情ID
	StaffUuid                    interface{} // 员工ID
	DutyNo                       interface{} // 当班编号
	DutyStartTime                interface{} // 当班开始时间
	DutyEndTime                  interface{} // 当班结束时间
	TotalSales                   interface{} // 总销售额
	TotalServiceFee              interface{} // 总服务费
	TotalPaymentCommissionFee    interface{} // 总支付手续费
	TotalTaxFee                  interface{} // 总税费
	TotalProductQuantity         interface{} // 商品数量
	TotalDiscountFee             interface{} // 总优惠折扣
	TotalRefundFee               interface{} // 总退款
	TotalRevenue                 interface{} // 总营业收入
	TotalActualAmount            interface{} // 总实收金额
	TotalRechargeAmount          interface{} // 充值金额
	TotalGiftAmount              interface{} // 赠送金额
	TotalGiftPoint               interface{} // 赠送积分
	PreviousBalance              interface{} // 上一班遗留备用金
	TotalOffCashWithdrawal       interface{} // 下班取出现金
	TotalCashBalance             interface{} // 本班遗留备用金
	CashDeposit                  interface{} // 中途存入现金
	CashWithdrawal               interface{} // 中途取出现金
	ExceptionReport              interface{} // 异常报备
	TotalReturnFoodCount         interface{} // 退菜次数
	TotalRefundCount             interface{} // 退款次数
	TotalReconciliationCount     interface{} // 反结账次数
	TotalGiftProductCount        interface{} // 赠菜次数
	TotalFreeOrderCount          interface{} // 免单次数
	TotalTransferProductCount    interface{} // 转菜次数
	TotalSinglePriceChangeCount  interface{} // 单品改价次数
	TotalOrderPriceChangeCount   interface{} // 整单改价次数
	TotalOrderDiscoutCount       interface{} // 整单折扣次数
	TotalZeroCheckoutCount       interface{} // 整单结账抹零次数
	TotalOrderCount              interface{} // 所有订单数
	TotalTableCount              interface{} // 桌数
	TotalCustomerCount           interface{} // 人数
	TotalMinOrderAmount          interface{} // 最小订单金额
	TotalMaxOrderAmount          interface{} // 最大订单金额
	TotalAverageOrderAmount      interface{} // 平均订单金额
	TotalTableCustomerCount      interface{} // 桌台人数
	TotalTableMinOrderAmount     interface{} // 桌台最小订单金额
	TotalTableMaxOrderAmount     interface{} // 桌台最大订单金额
	TotalTableAverageOrderAmount interface{} // 桌台人均消费金额
	TotalScanOrderCount          interface{} // 点餐订单数
	TotalScanMinOrderAmount      interface{} // 点餐最小订单金额
	TotalScanMaxOrderAmount      interface{} // 点餐最大订单金额
	TotalScanAverageOrderAmount  interface{} // 点餐平均订单金额
	TotalGiftProductAmount       interface{} // 赠菜金额
	TotalGiftProductPoint        interface{} // 赠菜积分
	CreateTime                   interface{} // 创建时间(时间戳)
	UpdateTime                   interface{} // 更新时间(时间戳)
	DeleteTime                   interface{} // 删除时间(时间戳)
}
