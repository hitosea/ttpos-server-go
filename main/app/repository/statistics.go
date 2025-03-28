package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IStatisticsRepo interface {
	CountShiftSaleRefundAmount(shiftNo string) model.StatisticsShiftSaleRefundAmount                  // 统计当班用餐订单退款金额
	CountShiftRechargeRefundAmount(shiftNo string) model.StatisticsShiftRechargeRefundAmount          // 统计当班充值订单退款金额
	CountShiftPaymentMethodAmount(shiftNo string) []model.StatisticsPaymentMethodAmount               // 统计当前班次支付方式收入
	CountShiftSaleFreeAmount(shiftNo string) model.StatisticsSaleFreeAmount                           // 统计当班用餐订单免单金额
	CountBusinessAmount(shiftNo string) float64                                                       // 统计营业额
	CountSale(opts ...DBOption) model.StatisticsSaleData                                              // 统计销售
	CountPayment(opts ...DBOption) []model.StatisticsPaymentData                                      // 统计支付
	CountTax(opts ...DBOption) []model.StatisticsTaxData                                              // 统计税类
	CountCategory(categoryType int, language string, opts ...DBOption) []model.StatisticsCategoryData // 统计分类
	CountProduct(language string, opts ...DBOption) []model.StatisticsProductData                     // 统计商品
	CountArea(opts ...DBOption) []model.StatisticsAreaData                                            // 统计区域
	Count7Days(opts ...DBOption) []model.Statistics7DaysData                                          // 统计销售天数
	CountMemberNum(opts ...DBOption) int64                                                            // 统计会员数量
	RankProduct(rankType int, language string, opts ...DBOption) []model.StatisticsProductData        // 统计商品排行
	SaveSale(sales []model.StatisticsSale) error                                                      // 保存销售
	SavePayment(payments []model.StatisticsPayment) error                                             // 保存支付
	SaveProduct(products []model.StatisticsProduct) error                                             // 保存商品
	DeleteSale(saleBillUuid uint64) error                                                             // 删除销售
	DeletePayment(saleBillUuid uint64) error                                                          // 删除支付
	DeleteProduct(saleBillUuid uint64) error                                                          // 删除商品
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
	memberRechargeOrderTable := prefix + "member_recharge_order"

	subQuery := db.Raw(
		"(SELECT "+saleOrderTable+".uuid AS uuid FROM "+saleBillTable+
			" LEFT JOIN "+saleOrderTable+" ON "+saleBillTable+".uuid = "+saleOrderTable+".sale_bill_uuid AND "+saleOrderTable+".delete_time = 0 "+
			"WHERE "+saleBillTable+".duty_no = ? AND "+saleBillTable+".delete_time = 0 AND "+saleOrderTable+".status = ?) UNION ALL "+
			"(SELECT uuid FROM "+memberRechargeOrderTable+" "+
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

// CountBusinessAmount 统计营业额
func (r *StatisticsRepo) CountBusinessAmount(shiftNo string) float64 {
	var result float64

	return result
}

// CountSale 统计销售数据
func (r *StatisticsRepo) CountSale(opts ...DBOption) model.StatisticsSaleData {
	var result model.StatisticsSaleData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	subQuery := db.Model(&model.StatisticsSale{}).
		Select(
			"sale_bill_uuid",
			"desk_uuid",
			"SUM(product_price + product_tax + service_fee + service_tax + payment_fee) AS sale_amount",
			"SUM(payment_amount - payment_balance) AS received_amount",
			"SUM(product_price) AS product_price",
			"SUM(product_num) AS product_num",
			"SUM(discount_member) AS discount_member",
			"SUM(payment_amount - product_tax - service_tax) AS business_amount",
			"SUM(payment_fee) AS payment_fee",
			"SUM(service_fee) AS service_fee",
			"SUM(product_tax + service_tax) AS tax",
			"SUM(refund_amount) AS refund_amount",
			"SUM(discount) AS discount",
			"SUM(gift_amount) AS gift_amount",
			"SUM(gift_num) AS gift_num",
			"SUM(free_amount) AS free_amount",
			"SUM(free_num) AS free_num",
			"SUM(IF(desk_uuid > 0, meal_num, 0)) AS meal_num",
			"SUM(payment_amount) AS order_amount",
			"SUM(IF(desk_uuid > 0, payment_amount, NULL)) AS desk_order_amount",
			"SUM(IF(desk_uuid = 0, payment_amount, NULL)) AS instant_order_amount",
		).Group("sale_bill_uuid")

	r.db.Table("(?) AS t", subQuery).
		Select(
			"SUM(t.sale_amount) AS total_sale_amount",
			"SUM(t.received_amount) AS total_received_amount",
			"SUM(t.product_price) AS total_product_price",
			"SUM(t.product_num) AS total_product_num",
			"SUM(t.discount_member) AS total_discount_member",
			"SUM(t.business_amount) AS total_business_amount",
			"SUM(t.payment_fee) AS total_payment_fee",
			"SUM(t.service_fee) AS total_service_fee",
			"SUM(t.tax) AS total_tax",
			"SUM(t.refund_amount) AS total_refund_amount",
			"SUM(t.discount) AS total_discount",
			"SUM(t.gift_amount) AS total_gift_amount",
			"SUM(t.gift_num) AS total_gift_num",
			"SUM(t.free_amount) AS total_free_amount",
			"SUM(t.free_num) AS total_free_num",
			"COUNT(t.sale_bill_uuid) AS total_order_num",
			"COUNT(CASE WHEN t.desk_uuid > 0 THEN 1 END) AS total_desk_num",
			"SUM(t.desk_order_amount) AS total_desk_order_amount",
			"SUM(t.meal_num) AS total_meal_num",
			"SUM(t.instant_order_amount) AS total_instant_order_amount",
			"COUNT(CASE WHEN t.desk_uuid = 0 THEN 1 END) AS total_instant_order_num",
			"MIN(t.order_amount) AS min_order_amount",
			"MAX(t.order_amount) AS max_order_amount",
			"AVG(t.order_amount) AS avg_order_amount",
			"MIN(CASE WHEN t.desk_order_amount >= 0 THEN t.desk_order_amount ELSE NULL END) AS min_desk_order_amount",
			"MAX(CASE WHEN t.desk_order_amount >= 0 THEN t.desk_order_amount ELSE NULL END) AS max_desk_order_amount",
			"AVG(CASE WHEN t.desk_order_amount >= 0 THEN t.desk_order_amount ELSE NULL END) AS avg_desk_order_amount",
			"MIN(CASE WHEN t.instant_order_amount >= 0 THEN t.instant_order_amount ELSE NULL END) AS min_instant_order_amount",
			"MAX(CASE WHEN t.instant_order_amount >= 0 THEN t.instant_order_amount ELSE NULL END) AS max_instant_order_amount",
			"AVG(CASE WHEN t.instant_order_amount >= 0 THEN t.instant_order_amount ELSE NULL END) AS avg_instant_order_amount",
		).
		Find(&result)

	return result
}

// CountPayment 统计支付
func (r *StatisticsRepo) CountPayment(opts ...DBOption) []model.StatisticsPaymentData {
	var result []model.StatisticsPaymentData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsPaymentTable := prefix + "statistics_payment"
	paymentMethodTable := prefix + "payment_method"

	db.Model(&model.StatisticsPayment{}).
		Select(
			statisticsPaymentTable+".payment_method_uuid",
			paymentMethodTable+".name AS payment_name",
			paymentMethodTable+".code AS payment_code",
			"COUNT("+statisticsPaymentTable+".payment_method_uuid) AS total_order_num",
			"SUM("+statisticsPaymentTable+".payment_amount) AS total_payment_amount",
			"SUM("+statisticsPaymentTable+".refund_amount) AS total_refund_amount",
		).
		Joins("LEFT JOIN " + paymentMethodTable + " ON " + statisticsPaymentTable + ".payment_method_uuid = " + paymentMethodTable + ".uuid").
		Group(statisticsPaymentTable + ".payment_method_uuid").
		Find(&result)

	return result
}

// CountTax 统计税类
func (r *StatisticsRepo) CountTax(opts ...DBOption) []model.StatisticsTaxData {
	var result []model.StatisticsTaxData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsProduct{}).
		Select(
			"tax_rate",
			"SUM((product_price + tax_fee) * product_num) AS total_product_amount",
			"SUM(tax_fee) AS total_tax_fee",
		).Group("tax_rate").
		Find(&result)

	return result
}

// CountCategory 统计分类
func (r *StatisticsRepo) CountCategory(categoryType int, language string, opts ...DBOption) []model.StatisticsCategoryData {
	var result []model.StatisticsCategoryData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productBomTable := prefix + "product_bom as pb"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"

	if categoryType == 1 {
		db.Table(statisticsProductTable).
			Select(
				"IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid) AS category_parent_uuid",
				"IF(pc.parent_uuid = 0, JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')), JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"'))) AS category_parent_name",
				"0 AS category_uuid",
				"'' AS category_name",
				"SUM(sp.product_num) AS sale_num",
				"SUM(sp.product_sale_price * sp.product_num) AS sale_amount",
			).
			Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
			Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
			Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
			Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
			Group("IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid)").
			Find(&result)
	}

	if categoryType == 2 {
		db.Table(statisticsProductTable).
			Select(
				"pc.parent_uuid AS category_parent_uuid",
				"JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"')) AS category_parent_name",
				"pc.uuid AS category_uuid",
				"JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')) AS category_name",
				"SUM(sp.product_num) AS sale_num",
				"SUM(sp.product_sale_price * sp.product_num) AS sale_amount",
			).
			Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
			Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
			Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
			Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
			Where("pc.parent_uuid > 0").
			Group("pc.uuid").
			Find(&result)
	}

	return result
}

// CountProduct 统计商品
func (r *StatisticsRepo) CountProduct(language string, opts ...DBOption) []model.StatisticsProductData {
	var result []model.StatisticsProductData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productBomTable := prefix + "product_bom as pb"

	db.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+language+"')) AS product_name",
			"JSON_UNQUOTE(JSON_EXTRACT(pb.name, '$."+language+"')) AS flavor_name",
			"sp.product_sale_price AS sale_price",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.product_sale_price * sp.product_num) AS sale_amount",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
		Group("sp.product_bom_uuid").
		Find(&result)

	return result
}

// CountArea 统计区域
func (r *StatisticsRepo) CountArea(opts ...DBOption) []model.StatisticsAreaData {
	var result []model.StatisticsAreaData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsSaleTable := prefix + "statistics_sale as ss"
	deskTable := prefix + "desk as d"
	deskRegionTable := prefix + "desk_region as dr"

	db.Table(statisticsSaleTable).
		Select(
			"dr.name AS area_name",
			"SUM(ss.product_price + ss.product_tax + ss.service_fee + ss.service_tax + ss.payment_fee) AS area_sale_amount",
			"SUM(ss.payment_amount - ss.product_tax - ss.service_tax) AS area_business_amount",
			"SUM(ss.product_num) AS area_product_num",
		).
		Joins("LEFT JOIN " + deskTable + " ON ss.desk_uuid = d.uuid").
		Joins("LEFT JOIN " + deskRegionTable + " ON d.region_uuid = dr.uuid").
		Where("ss.desk_uuid > 0").
		Group("dr.uuid").
		Find(&result)

	return result
}

// RankProduct 统计商品排行
func (r *StatisticsRepo) RankProduct(rankType int, language string, opts ...DBOption) []model.StatisticsProductData {
	var result []model.StatisticsProductData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productBomTable := prefix + "product_bom as pb"

	query := db.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+language+"')) AS product_name",
			"JSON_UNQUOTE(JSON_EXTRACT(pb.name, '$."+language+"')) AS flavor_name",
			"sp.product_sale_price AS sale_price",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.product_sale_price * sp.product_num) AS sale_amount",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
		Where("sp.refund_time = 0").
		Group("sp.product_bom_uuid")

	if rankType == 1 {
		query = query.Order("sale_num DESC")
	}

	if rankType == 2 {
		query = query.Order("sale_amount DESC")
	}

	query = query.Limit(10)
	query.Find(&result)

	return result
}

// Count7Days 统计销售天数
func (r *StatisticsRepo) Count7Days(opts ...DBOption) []model.Statistics7DaysData {
	var result []model.Statistics7DaysData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	subQuery := db.Model(&model.StatisticsSale{}).
		Select(
			"complete_time",
			"COUNT(sale_bill_uuid) AS order_num",
			"SUM(payment_amount - payment_balance) AS received_amount",
		).
		Group("sale_bill_uuid")

	r.db.Table("(?) AS t", subQuery).
		Select(
			"FROM_UNIXTIME(t.complete_time, '%Y-%m-%d') AS day",
			"SUM(t.order_num) AS total_order_num",
			"SUM(t.received_amount) AS total_received_amount",
		).
		Group("FROM_UNIXTIME(t.complete_time, '%Y-%m-%d')").
		Find(&result)

	return result
}

// CountMemberNum 统计会员数量
func (r *StatisticsRepo) CountMemberNum(opts ...DBOption) int64 {
	var result int64
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.Member{}).Count(&result)

	return result
}

// SaveSale 保存销售
func (r *StatisticsRepo) SaveSale(sales []model.StatisticsSale) error {
	return r.db.Create(&sales).Error
}

// DeleteSale 删除销售
func (r *StatisticsRepo) DeleteSale(saleBillUuid uint64) error {
	return r.db.Where("sale_bill_uuid = ?", saleBillUuid).Delete(&model.StatisticsSale{}).Error
}

// SavePayment 保存支付
func (r *StatisticsRepo) SavePayment(payments []model.StatisticsPayment) error {
	return r.db.Create(&payments).Error
}

// DeletePayment 删除支付
func (r *StatisticsRepo) DeletePayment(saleBillUuid uint64) error {
	return r.db.Where("sale_bill_uuid = ?", saleBillUuid).Delete(&model.StatisticsPayment{}).Error
}

// SaveProduct 保存商品
func (r *StatisticsRepo) SaveProduct(products []model.StatisticsProduct) error {
	return r.db.Create(&products).Error
}

// DeleteProduct 删除商品
func (r *StatisticsRepo) DeleteProduct(saleBillUuid uint64) error {
	return r.db.Where("sale_bill_uuid = ?", saleBillUuid).Delete(&model.StatisticsProduct{}).Error
}
