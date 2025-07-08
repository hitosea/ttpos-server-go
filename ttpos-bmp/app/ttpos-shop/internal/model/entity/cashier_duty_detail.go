// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CashierDutyDetail is the golang structure for table cashier_duty_detail.
type CashierDutyDetail struct {
	Id                           uint    `json:"id"                           orm:"id"                               description:"自增ID"`      // 自增ID
	Uuid                         uint64  `json:"uuid"                         orm:"uuid"                             description:"收银交班详情ID"`  // 收银交班详情ID
	StaffUuid                    uint64  `json:"staffUuid"                    orm:"staff_uuid"                       description:"员工ID"`      // 员工ID
	DutyNo                       string  `json:"dutyNo"                       orm:"duty_no"                          description:"当班编号"`      // 当班编号
	DutyStartTime                uint    `json:"dutyStartTime"                orm:"duty_start_time"                  description:"当班开始时间"`    // 当班开始时间
	DutyEndTime                  uint    `json:"dutyEndTime"                  orm:"duty_end_time"                    description:"当班结束时间"`    // 当班结束时间
	TotalSales                   float64 `json:"totalSales"                   orm:"total_sales"                      description:"总销售额"`      // 总销售额
	TotalServiceFee              float64 `json:"totalServiceFee"              orm:"total_service_fee"                description:"总服务费"`      // 总服务费
	TotalPaymentCommissionFee    float64 `json:"totalPaymentCommissionFee"    orm:"total_payment_commission_fee"     description:"总支付手续费"`    // 总支付手续费
	TotalTaxFee                  float64 `json:"totalTaxFee"                  orm:"total_tax_fee"                    description:"总税费"`       // 总税费
	TotalProductQuantity         int     `json:"totalProductQuantity"         orm:"total_product_quantity"           description:"商品数量"`      // 商品数量
	TotalDiscountFee             float64 `json:"totalDiscountFee"             orm:"total_discount_fee"               description:"总优惠折扣"`     // 总优惠折扣
	TotalRefundFee               float64 `json:"totalRefundFee"               orm:"total_refund_fee"                 description:"总退款"`       // 总退款
	TotalRevenue                 float64 `json:"totalRevenue"                 orm:"total_revenue"                    description:"总营业收入"`     // 总营业收入
	TotalActualAmount            float64 `json:"totalActualAmount"            orm:"total_actual_amount"              description:"总实收金额"`     // 总实收金额
	TotalRechargeAmount          float64 `json:"totalRechargeAmount"          orm:"total_recharge_amount"            description:"充值金额"`      // 充值金额
	TotalGiftAmount              float64 `json:"totalGiftAmount"              orm:"total_gift_amount"                description:"赠送金额"`      // 赠送金额
	TotalGiftPoint               int     `json:"totalGiftPoint"               orm:"total_gift_point"                 description:"赠送积分"`      // 赠送积分
	PreviousBalance              float64 `json:"previousBalance"              orm:"previous_balance"                 description:"上一班遗留备用金"`  // 上一班遗留备用金
	TotalOffCashWithdrawal       float64 `json:"totalOffCashWithdrawal"       orm:"total_off_cash_withdrawal"        description:"下班取出现金"`    // 下班取出现金
	TotalCashBalance             float64 `json:"totalCashBalance"             orm:"total_cash_balance"               description:"本班遗留备用金"`   // 本班遗留备用金
	CashDeposit                  float64 `json:"cashDeposit"                  orm:"cash_deposit"                     description:"中途存入现金"`    // 中途存入现金
	CashWithdrawal               float64 `json:"cashWithdrawal"               orm:"cash_withdrawal"                  description:"中途取出现金"`    // 中途取出现金
	ExceptionReport              string  `json:"exceptionReport"              orm:"exception_report"                 description:"异常报备"`      // 异常报备
	TotalReturnFoodCount         int     `json:"totalReturnFoodCount"         orm:"total_return_food_count"          description:"退菜次数"`      // 退菜次数
	TotalRefundCount             int     `json:"totalRefundCount"             orm:"total_refund_count"               description:"退款次数"`      // 退款次数
	TotalReconciliationCount     int     `json:"totalReconciliationCount"     orm:"total_reconciliation_count"       description:"反结账次数"`     // 反结账次数
	TotalGiftProductCount        int     `json:"totalGiftProductCount"        orm:"total_gift_product_count"         description:"赠菜次数"`      // 赠菜次数
	TotalFreeOrderCount          int     `json:"totalFreeOrderCount"          orm:"total_free_order_count"           description:"免单次数"`      // 免单次数
	TotalTransferProductCount    int     `json:"totalTransferProductCount"    orm:"total_transfer_product_count"     description:"转菜次数"`      // 转菜次数
	TotalSinglePriceChangeCount  int     `json:"totalSinglePriceChangeCount"  orm:"total_single_price_change_count"  description:"单品改价次数"`    // 单品改价次数
	TotalOrderPriceChangeCount   int     `json:"totalOrderPriceChangeCount"   orm:"total_order_price_change_count"   description:"整单改价次数"`    // 整单改价次数
	TotalOrderDiscoutCount       int     `json:"totalOrderDiscoutCount"       orm:"total_order_discout_count"        description:"整单折扣次数"`    // 整单折扣次数
	TotalZeroCheckoutCount       int     `json:"totalZeroCheckoutCount"       orm:"total_zero_checkout_count"        description:"整单结账抹零次数"`  // 整单结账抹零次数
	TotalOrderCount              int     `json:"totalOrderCount"              orm:"total_order_count"                description:"所有订单数"`     // 所有订单数
	TotalTableCount              int     `json:"totalTableCount"              orm:"total_table_count"                description:"桌数"`        // 桌数
	TotalCustomerCount           int     `json:"totalCustomerCount"           orm:"total_customer_count"             description:"人数"`        // 人数
	TotalMinOrderAmount          float64 `json:"totalMinOrderAmount"          orm:"total_min_order_amount"           description:"最小订单金额"`    // 最小订单金额
	TotalMaxOrderAmount          float64 `json:"totalMaxOrderAmount"          orm:"total_max_order_amount"           description:"最大订单金额"`    // 最大订单金额
	TotalAverageOrderAmount      float64 `json:"totalAverageOrderAmount"      orm:"total_average_order_amount"       description:"平均订单金额"`    // 平均订单金额
	TotalTableCustomerCount      int     `json:"totalTableCustomerCount"      orm:"total_table_customer_count"       description:"桌台人数"`      // 桌台人数
	TotalTableMinOrderAmount     float64 `json:"totalTableMinOrderAmount"     orm:"total_table_min_order_amount"     description:"桌台最小订单金额"`  // 桌台最小订单金额
	TotalTableMaxOrderAmount     float64 `json:"totalTableMaxOrderAmount"     orm:"total_table_max_order_amount"     description:"桌台最大订单金额"`  // 桌台最大订单金额
	TotalTableAverageOrderAmount float64 `json:"totalTableAverageOrderAmount" orm:"total_table_average_order_amount" description:"桌台人均消费金额"`  // 桌台人均消费金额
	TotalScanOrderCount          int     `json:"totalScanOrderCount"          orm:"total_scan_order_count"           description:"点餐订单数"`     // 点餐订单数
	TotalScanMinOrderAmount      float64 `json:"totalScanMinOrderAmount"      orm:"total_scan_min_order_amount"      description:"点餐最小订单金额"`  // 点餐最小订单金额
	TotalScanMaxOrderAmount      float64 `json:"totalScanMaxOrderAmount"      orm:"total_scan_max_order_amount"      description:"点餐最大订单金额"`  // 点餐最大订单金额
	TotalScanAverageOrderAmount  float64 `json:"totalScanAverageOrderAmount"  orm:"total_scan_average_order_amount"  description:"点餐平均订单金额"`  // 点餐平均订单金额
	TotalGiftProductAmount       float64 `json:"totalGiftProductAmount"       orm:"total_gift_product_amount"        description:"赠菜金额"`      // 赠菜金额
	TotalGiftProductPoint        int     `json:"totalGiftProductPoint"        orm:"total_gift_product_point"         description:"赠菜积分"`      // 赠菜积分
	CreateTime                   uint    `json:"createTime"                   orm:"create_time"                      description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime                   uint    `json:"updateTime"                   orm:"update_time"                      description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime                   uint    `json:"deleteTime"                   orm:"delete_time"                      description:"删除时间(时间戳)"` // 删除时间(时间戳)
}
