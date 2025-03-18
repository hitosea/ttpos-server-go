package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IStatisticsRepo interface {
	CountShiftSaleRefundAmount(shiftNo string) model.StatisticsShiftSaleRefundAmount         // 统计当班用餐订单退款金额
	CountShiftRechargeRefundAmount(shiftNo string) model.StatisticsShiftRechargeRefundAmount // 统计当班充值订单退款金额
	CountShiftPaymentMethodAmount(shiftNo string) []model.StatisticsPaymentMethodAmount      // 统计当前班次支付方式收入
	CountShiftSaleFreeAmount(shiftNo string) model.StatisticsSaleFreeAmount                  // 统计当班用餐订单免单金额
}

func NewStatisticsRepo(db *gorm.DB) IStatisticsRepo {
	return NewStatisticsRepoImpl(db)
}

type StatisticsRepo struct {
	db *gorm.DB
}

func NewStatisticsRepoImpl(db *gorm.DB) *StatisticsRepo {
	return &StatisticsRepo{db: db}
}

// CountShiftSaleRefundAmount 统计当班用餐订单退款金额
func (r *StatisticsRepo) CountShiftSaleRefundAmount(shiftNo string) model.StatisticsShiftSaleRefundAmount {
	var result model.StatisticsShiftSaleRefundAmount

	db := r.db
	prefix := config.Database.TablePrefix

	saleBillTable := prefix + "sale_bill"       // 销售账单表
	saleOrderTable := prefix + "sale_order"     // 销售订单表
	returnOrderTable := prefix + "return_order" // 退款订单表

	db.Table(saleBillTable).
		Select("duty_no AS shift_no, SUM(refund_amount) AS refund_amount, SUM(refund_tax_amount) AS refund_tax_amount").
		Joins("LEFT JOIN "+saleOrderTable+" ON "+saleBillTable+".uuid = "+saleOrderTable+".sale_bill_uuid").
		Joins("LEFT JOIN "+returnOrderTable+" ON "+saleOrderTable+".uuid = "+returnOrderTable+".related_order_uuid").
		Where(saleBillTable+".duty_no = ?", shiftNo).
		Where(saleBillTable+".delete_time = 0").
		Where(saleOrderTable+".status = ?", constant.SaleOrderStatusFinish).
		Where(saleOrderTable+".delete_time = 0").
		Where(returnOrderTable+".related_order_type = ?", constant.ReturnOrderRelatedOrderTypeSaleOrder).
		Where(returnOrderTable+".is_reverse_settlement = ?", constant.ReturnOrderIsReverseSettlementNo).
		Where(returnOrderTable + ".delete_time = 0").
		Find(&result)

	return result
}

// CountShiftRechargeRefundAmount 统计当班充值订单退款金额
func (r *StatisticsRepo) CountShiftRechargeRefundAmount(shiftNo string) model.StatisticsShiftRechargeRefundAmount {
	var result model.StatisticsShiftRechargeRefundAmount

	r.db.Model(&model.MemberRechargeOrder{}).
		Select("duty_no AS shift_no, SUM(refund_amount) AS refund_amount").
		Where("duty_no = ?", shiftNo).
		Where("delete_time = 0").
		Find(&result)

	return result
}

// CountShiftPaymentMethodAmount 统计当前班次支付方式收入
func (r *StatisticsRepo) CountShiftPaymentMethodAmount(shiftNo string) []model.StatisticsPaymentMethodAmount {
	var list []model.StatisticsPaymentMethodAmount

	db := r.db
	prefix := config.Database.TablePrefix
	paymentOrderTable := prefix + "payment_order"
	paymentMethodTable := prefix + "payment_method"
	saleBillTable := prefix + "sale_bill"
	saleOrderTable := prefix + "sale_order"
	returnOrderAmountTable := prefix + "return_order_amount"

	subQuery := db.Raw(
		"(SELECT "+saleOrderTable+".uuid AS uuid FROM "+saleBillTable+
			" LEFT JOIN "+saleOrderTable+" ON "+saleBillTable+".uuid = "+saleOrderTable+".sale_bill_uuid AND "+saleOrderTable+".delete_time = 0 "+
			"WHERE "+saleBillTable+".duty_no = ? AND "+saleBillTable+".delete_time = 0 AND "+saleOrderTable+".status = ?) UNION ALL "+
			"(SELECT uuid FROM "+prefix+"member_recharge_order "+
			"WHERE duty_no = ? AND status = 1 AND delete_time = 0)",
		shiftNo,
		constant.SaleOrderStatusFinish,
		shiftNo,
	)
	db.Model(&model.PaymentOrder{}).
		Select(
			paymentOrderTable+".payment_method_uuid AS payment_method_uuid, "+
				paymentMethodTable+".name AS payment_name, "+
				paymentMethodTable+".code AS payment_code, "+
				"SUM("+paymentOrderTable+".amount) AS pay_amount, "+
				"SUM("+returnOrderAmountTable+".amount) AS refund_amount",
		).
		Joins("LEFT JOIN "+paymentMethodTable+" ON "+paymentOrderTable+".payment_method_uuid = "+paymentMethodTable+".uuid AND "+paymentMethodTable+".delete_time = 0").
		Joins("LEFT JOIN "+returnOrderAmountTable+" ON "+paymentOrderTable+".uuid = "+returnOrderAmountTable+".payment_order_uuid AND "+returnOrderAmountTable+".delete_time = 0").
		Where(paymentOrderTable+".related_uuid IN (?)", subQuery).
		Where(paymentOrderTable+".status = ?", constant.PaymentOrderStatusPaid).
		Where(paymentOrderTable + ".delete_time = 0").
		Group(paymentOrderTable + ".payment_method_uuid").
		Find(&list)

	return list
}

// CountShiftSaleFreeAmount 统计当班用餐订单免单金额
func (r *StatisticsRepo) CountShiftSaleFreeAmount(shiftNo string) model.StatisticsSaleFreeAmount {
	var result model.StatisticsSaleFreeAmount

	prefix := config.Database.TablePrefix
	saleBillTable := prefix + "sale_bill"
	saleOrderTable := prefix + "sale_order"

	r.db.Table(saleBillTable).
		Select(saleBillTable+".duty_no AS shift_no, SUM("+saleOrderTable+".amount) AS free_amount").
		Joins("LEFT JOIN "+saleOrderTable+" ON "+saleBillTable+".uuid = "+saleOrderTable+".sale_bill_uuid").
		Where(saleBillTable+".duty_no = ?", shiftNo).
		Where(saleBillTable+".delete_time = 0").
		Where(saleOrderTable+".status = ?", constant.SaleOrderStatusFinish).
		Where(saleOrderTable+".is_free = ?", constant.SaleOrderIsFreeYes).
		Where(saleOrderTable + ".delete_time = 0").
		Find(&result)

	return result
}
