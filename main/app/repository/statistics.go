package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IStatisticsRepo interface {
	CountSale(opts ...DBOption) model.StatisticsSaleData                                                       // 统计销售
	CountSaleDays(opts ...DBOption) []model.StatisticsSaleDaysData                                             // 统计销售天数
	CountPayment(opts ...DBOption) []model.StatisticsPaymentData                                               // 统计支付
	CountPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData                                       // 统计支付天数
	CountTax(opts ...DBOption) []model.StatisticsTaxData                                                       // 统计税类
	CountCategory(categoryType int, language string, opts ...DBOption) []model.StatisticsCategoryData          // 统计分类
	CountProduct(language string, opts ...DBOption) []model.StatisticsProductData                              // 统计商品
	CountArea(opts ...DBOption) []model.StatisticsAreaData                                                     // 统计区域
	CountAreaDays(opts ...DBOption) []model.StatisticsAreaDaysData                                             // 统计区域
	Count7Days(opts ...DBOption) []model.Statistics7DaysData                                                   // 统计销售天数
	CountUnpaidOrder(opts ...DBOption) model.StatisticsUnpaidOrderData                                         // 统计未结订单
	CountMemberNum(opts ...DBOption) int64                                                                     // 统计会员数量
	CountMemberNumDays(opts ...DBOption) []model.CountMemberNumDaysResp                                        // 统计会员数量天数
	CountMember(opts ...DBOption) model.StatisticsMemberData                                                   // 统计会员
	CountMemberDays(opts ...DBOption) []model.StatisticsMemberDaysData                                         // 统计会员天数
	CountMemberPayment(opts ...DBOption) []model.StatisticsPaymentData                                         // 统计会员支付
	CountMemberPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData                                 // 统计会员支付天数
	CountProductSale(req CountProductSaleRepoReq, opts ...DBOption) ([]model.StatisticsProductSaleData, int64) // 统计商品销售
	CountFreePayment(opts ...DBOption) model.StatisticsFreePaymentData                                         // 统计免单支付
	CountFreePaymentDays(opts ...DBOption) []model.StatisticsFreePaymentDaysData                               // 统计免单支付天数
	RankProduct(rankType int, language string, opts ...DBOption) []model.StatisticsProductData                 // 统计商品排行
	SaveSale(sales []model.StatisticsSale) error                                                               // 保存销售
	SavePayment(payments []model.StatisticsPayment) error                                                      // 保存支付
	SaveProduct(products []model.StatisticsProduct) error                                                      // 保存商品
	DeleteSale(saleBillUuid uint64) error                                                                      // 删除销售
	DeletePayment(saleBillUuid uint64) error                                                                   // 删除支付
	DeleteProduct(saleBillUuid uint64) error                                                                   // 删除商品
	SaveMember(member model.StatisticsMember) error                                                            // 保存会员
	SaveMemberPayment(payments []model.StatisticsMemberPayment) error                                          // 保存会员支付
	DeleteMember(memberRechargeOrderUuid uint64) error                                                         // 删除会员
	DeleteMemberPayment(memberRechargeOrderUuid uint64) error                                                  // 删除会员支付
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

var (
	// 统计销售子查询
	countSaleSubQuerySelect = []string{
		"sale_bill_uuid",
		"desk_uuid",
		"SUM(product_price + product_tax + service_fee + service_tax + no_refund_tax + payment_fee - refund_tax - refund_service_fee - refund_fee) AS sale_amount",
		"SUM(payment_amount - refund_amount - payment_balance) AS received_amount",
		"SUM(product_price) AS product_price",
		"SUM(product_origin_price) AS product_origin_price",
		"SUM(product_num) AS product_num",
		"SUM(discount_member) AS discount_member",
		"SUM(payment_amount - refund_amount - refund_payment_balance - product_tax - service_tax + refund_tax) AS business_amount",
		"SUM(payment_fee - refund_fee) AS payment_fee",
		"SUM(service_fee - refund_service_fee) AS service_fee",
		"SUM(product_tax + service_tax - refund_tax) AS tax",
		"SUM(refund_amount + refund_payment_balance) AS refund_amount",
		"SUM(discount - refund_discount) AS discount",
		"SUM(gift_amount) AS gift_amount",
		"SUM(gift_num) AS gift_num",
		"SUM(free_amount) AS free_amount",
		"SUM(free_num) AS free_num",
		"SUM(IF(desk_uuid > 0, meal_num, 0)) AS meal_num",
		"SUM(payment_amount - refund_amount - refund_payment_balance) AS order_amount",
		"SUM(IF(desk_uuid > 0, payment_amount - refund_amount - refund_payment_balance, NULL)) AS desk_order_amount",
		"SUM(IF(desk_uuid = 0, payment_amount - refund_amount - refund_payment_balance, NULL)) AS instant_order_amount",
		"complete_time",
	}
	// 统计销售
	countSaleSelect = []string{
		"SUM(t.sale_amount) AS total_sale_amount",
		"SUM(t.received_amount) AS total_received_amount",
		"SUM(t.product_price) AS total_product_price",
		"SUM(t.product_origin_price) AS total_product_origin_price",
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
	}
)

// CountSale 统计销售数据
func (r *StatisticsRepo) CountSale(opts ...DBOption) model.StatisticsSaleData {
	var result model.StatisticsSaleData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	subQuery := db.Model(&model.StatisticsSale{}).
		Select(countSaleSubQuerySelect).
		Group("sale_bill_uuid")

	r.db.Table("(?) AS t", subQuery).
		Select(countSaleSelect).
		Find(&result)

	return result
}

// CountSaleDays 统计销售天数
func (r *StatisticsRepo) CountSaleDays(opts ...DBOption) []model.StatisticsSaleDaysData {
	var result []model.StatisticsSaleDaysData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	subQuery := db.Model(&model.StatisticsSale{}).
		Select(countSaleSubQuerySelect).
		Group("sale_bill_uuid").
		Group("FROM_UNIXTIME(complete_time, '%Y-%m-%d')")

	r.db.Table("(?) AS t", subQuery).
		Select(countSaleSelect, "FROM_UNIXTIME(complete_time, '%Y-%m-%d') AS day").
		Group("DAY").
		Order("DAY ASC").
		Find(&result)

	return result
}

var (
	countPaymentSelect = []string{
		"sp.payment_method_uuid",
		"pm.name AS payment_name",
		"pm.code AS payment_code",
		"COUNT(sp.payment_method_uuid) AS total_order_num",
		"SUM(sp.payment_amount-sp.refund_amount) AS total_payment_amount",
		"SUM(sp.refund_amount) AS total_refund_amount",
	}
)

// CountPayment 统计支付
func (r *StatisticsRepo) CountPayment(opts ...DBOption) []model.StatisticsPaymentData {
	var result []model.StatisticsPaymentData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsPaymentTable := prefix + "statistics_payment sp"
	paymentMethodTable := prefix + "payment_method pm"

	db.Table(statisticsPaymentTable).
		Select(countPaymentSelect).
		Joins("LEFT JOIN " + paymentMethodTable + " ON sp.payment_method_uuid = pm.uuid").
		Group("sp.payment_method_uuid").
		Find(&result)

	return result
}

// CountPaymentDays 统计支付天数
func (r *StatisticsRepo) CountPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData {
	var result []model.StatisticsPaymentDaysData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsPaymentTable := prefix + "statistics_payment sp"
	paymentMethodTable := prefix + "payment_method pm"

	db.Table(statisticsPaymentTable).
		Select(countPaymentSelect, "FROM_UNIXTIME(sp.complete_time, '%Y-%m-%d') AS day").
		Joins("LEFT JOIN " + paymentMethodTable + " ON sp.payment_method_uuid = pm.uuid").
		Group("sp.payment_method_uuid").
		Group("day").
		Order("day ASC").
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
			"SUM((product_price + tax_fee) * (product_num - refund_num)) AS total_product_amount",
			"SUM(tax_fee * (product_num - refund_num)) AS total_tax_fee",
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

	if categoryType != 2 {
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
	} else {
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

var (
	countAreaSelect = []string{
		"dr.name AS area_name",
		"SUM(ss.product_price + ss.product_tax + ss.service_fee + ss.service_tax + ss.payment_fee - ss.refund_tax - ss.refund_service_fee) AS area_sale_amount",
		"SUM(ss.payment_amount - ss.refund_amount - ss.refund_payment_balance - ss.product_tax - ss.service_tax + ss.refund_tax) AS area_business_amount",
		"SUM(ss.product_num) AS area_product_num",
	}
)

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
		Select(countAreaSelect, "dr.uuid AS area_id").
		Joins("LEFT JOIN " + deskTable + " ON ss.desk_uuid = d.uuid").
		Joins("LEFT JOIN " + deskRegionTable + " ON d.region_uuid = dr.uuid").
		Where("ss.desk_uuid > 0").
		Group("dr.uuid").
		Find(&result)

	return result
}

// CountAreaDays 统计区域
func (r *StatisticsRepo) CountAreaDays(opts ...DBOption) []model.StatisticsAreaDaysData {
	var result []model.StatisticsAreaDaysData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsSaleTable := prefix + "statistics_sale as ss"
	deskTable := prefix + "desk as d"
	deskRegionTable := prefix + "desk_region as dr"
	db.Table(statisticsSaleTable).
		Select(countAreaSelect, "FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d') AS day").
		Joins("LEFT JOIN " + deskTable + " ON ss.desk_uuid = d.uuid").
		Joins("LEFT JOIN " + deskRegionTable + " ON d.region_uuid = dr.uuid").
		Where("ss.desk_uuid > 0").
		Group("dr.uuid").
		Group("day").
		Order("day ASC").
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

	query := db.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+language+"')) AS product_name",
			"sp.product_sale_price AS sale_price",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.product_sale_price * sp.product_num) AS sale_amount",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Where("sp.refund_time = 0").
		Group("sp.product_package_uuid")

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

// CountMemberNumDays 统计会员数量天数
func (r *StatisticsRepo) CountMemberNumDays(opts ...DBOption) []model.CountMemberNumDaysResp {
	var result []model.CountMemberNumDaysResp
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.Member{}).
		Select("COUNT(uuid) AS member_num", "FROM_UNIXTIME(create_time, '%Y-%m-%d') AS day").
		Group("day").
		Order("day ASC").
		Find(&result)

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

// CountUnpaidOrder 统计未结订单
func (r *StatisticsRepo) CountUnpaidOrder(opts ...DBOption) model.StatisticsUnpaidOrderData {
	var result model.StatisticsUnpaidOrderData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.SaleBill{}).
		Select("COUNT(uuid) AS total_order_num", "SUM(amount) AS total_amount").
		Where("status = ?", constant.SaleBillStatusPending).
		Where("production_time > 0").
		Find(&result)

	return result
}

// SaveMember 保存会员
func (r *StatisticsRepo) SaveMember(member model.StatisticsMember) error {
	return r.db.Create(&member).Error
}

// DeleteMember 删除会员
func (r *StatisticsRepo) DeleteMember(memberRechargeOrderUuid uint64) error {
	return r.db.Where("member_recharge_order_uuid = ?", memberRechargeOrderUuid).Delete(&model.StatisticsMember{}).Error
}

// SaveMemberPayment 保存会员支付
func (r *StatisticsRepo) SaveMemberPayment(payments []model.StatisticsMemberPayment) error {
	return r.db.Create(&payments).Error
}

// DeleteMemberPayment 删除会员支付
func (r *StatisticsRepo) DeleteMemberPayment(memberRechargeOrderUuid uint64) error {
	return r.db.Where("member_recharge_order_uuid = ?", memberRechargeOrderUuid).Delete(&model.StatisticsMemberPayment{}).Error
}

var (
	// 统计会员查询
	countMemberSelect = []string{
		"SUM(payment_fee) AS total_sale_amount",
		"SUM(IF(payment_amount - refund_amount = 0, 0, recharge_amount - refund_amount)) AS total_recharge_amount",
		"SUM(give_amount) AS total_give_amount",
		"SUM(give_point) AS total_give_point",
		"SUM(payment_amount - refund_amount - refund_fee) AS total_payment_amount",
		"SUM(payment_fee) AS total_payment_fee",
		"SUM(refund_amount) AS total_refund_amount",
	}
)

// CountMember 统计会员
func (r *StatisticsRepo) CountMember(opts ...DBOption) model.StatisticsMemberData {
	var result model.StatisticsMemberData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsMember{}).
		Select(countMemberSelect).
		Find(&result)

	return result
}

// CountMemberDays 统计会员天数
func (r *StatisticsRepo) CountMemberDays(opts ...DBOption) []model.StatisticsMemberDaysData {
	var result []model.StatisticsMemberDaysData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsMember{}).
		Select(countMemberSelect, "FROM_UNIXTIME(complete_time, '%Y-%m-%d') AS day").
		Group("FROM_UNIXTIME(complete_time, '%Y-%m-%d')").
		Find(&result)

	return result
}

var (
	countMemberPaymentSelect = []string{
		"smp.payment_method_uuid",
		"pm.name AS payment_name",
		"pm.code AS payment_code",
		"COUNT(smp.payment_method_uuid) AS total_order_num",
		"SUM(smp.payment_amount-smp.refund_amount) AS total_payment_amount",
		"SUM(smp.refund_amount) AS total_refund_amount",
	}
)

// CountMemberPayment 统计会员支付
func (r *StatisticsRepo) CountMemberPayment(opts ...DBOption) []model.StatisticsPaymentData {
	var result []model.StatisticsPaymentData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsMemberPaymentTable := prefix + "statistics_member_payment smp"
	paymentMethodTable := prefix + "payment_method pm"

	db.Table(statisticsMemberPaymentTable).
		Select(countMemberPaymentSelect).
		Joins("LEFT JOIN " + paymentMethodTable + " ON smp.payment_method_uuid = pm.uuid").
		Group("smp.payment_method_uuid").
		Find(&result)

	return result
}

// CountMemberPaymentDays 统计会员支付天数
func (r *StatisticsRepo) CountMemberPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData {
	var result []model.StatisticsPaymentDaysData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsMemberPaymentTable := prefix + "statistics_member_payment smp"
	paymentMethodTable := prefix + "payment_method pm"

	db.Table(statisticsMemberPaymentTable).
		Select(countMemberPaymentSelect, "FROM_UNIXTIME(smp.complete_time, '%Y-%m-%d') AS day").
		Joins("LEFT JOIN " + paymentMethodTable + " ON smp.payment_method_uuid = pm.uuid").
		Group("smp.payment_method_uuid").
		Group("day").
		Order("day ASC").
		Find(&result)

	return result
}

type CountProductSaleRepoReq struct {
	PageNo        int
	PageSize      int
	RankType      int
	RankDirection int
	Language      string
	AreaUuid      uint64
	CategoryUuid  uint64
	ProductName   string
}

// CountProductSale 统计商品销售
func (r *StatisticsRepo) CountProductSale(req CountProductSaleRepoReq, opts ...DBOption) ([]model.StatisticsProductSaleData, int64) {
	var result []model.StatisticsProductSaleData
	db := r.db
	db2 := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	for _, opt := range opts {
		db2 = opt(db2)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"
	saleBillTable := prefix + "sale_bill as sb"
	deskTable := prefix + "desk as d"

	var total int64
	db.Table(statisticsProductTable).
		Select("COUNT(DISTINCT sp.product_package_uuid) AS total").
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
		Joins("LEFT JOIN " + saleBillTable + " ON sp.sale_bill_uuid = sb.uuid")

	if req.AreaUuid > 0 {
		db.Where("sb.desk_uuid IN (?)", r.db.Table(deskTable).Select("d.uuid").Where("d.region_uuid = ?", req.AreaUuid))
	}
	if req.CategoryUuid > 0 {
		db.Where("pp.category_uuid = ? OR pp.category_uuid IN (?)", req.CategoryUuid, r.db.Table(productCategoryTable).Select("pc.uuid").Where("pc.parent_uuid = ?", req.CategoryUuid))
	}
	if req.ProductName != "" {
		db.Where("JSON_UNQUOTE(JSON_EXTRACT(pp.name, ?)) LIKE ?", "$."+req.Language, "%"+req.ProductName+"%")
	}
	db.Find(&total)

	listQuery := db2.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+req.Language+"')) AS product_name",
			"JSON_UNQUOTE(JSON_EXTRACT(pc.name, '$."+req.Language+"')) AS category_name",
			"JSON_UNQUOTE(JSON_EXTRACT(ppc.name, '$."+req.Language+"')) AS category_parent_name",
			"SUM(sp.product_num) AS sale_num",
			"SUM((sp.product_price + sp.tax_fee + sp.service_fee + service_tax) * sp.product_num) AS origin_sale_amount",
			"SUM(sp.product_final_price * sp.product_num) AS actual_sale_amount",
			"SUM((sp.product_final_price - sp.tax_fee - sp.service_tax) * sp.product_num) AS business_amount",
			"SUM(IF(sp.free_num > 0,sp.free_num,sp.give_num)) AS give_num",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
		Joins("LEFT JOIN " + saleBillTable + " ON sp.sale_bill_uuid = sb.uuid")

	if req.AreaUuid > 0 {
		listQuery.Where("sb.desk_uuid IN (?)", r.db.Table(deskTable).Select("d.uuid").Where("d.region_uuid = ?", req.AreaUuid))
	}
	if req.CategoryUuid > 0 {
		listQuery.Where("pp.category_uuid = ? OR pp.category_uuid IN (?)", req.CategoryUuid, r.db.Table(productCategoryTable).Select("pc.uuid").Where("pc.parent_uuid = ?", req.CategoryUuid))
	}
	if req.ProductName != "" {
		listQuery.Where("JSON_UNQUOTE(JSON_EXTRACT(pp.name, ?)) LIKE ?", "$."+req.Language, "%"+req.ProductName+"%")
	}
	listQuery.Group("sp.product_package_uuid")

	direction := "DESC"
	if req.RankDirection == 1 {
		direction = "ASC"
	}

	if req.RankType == 1 {
		listQuery = listQuery.Order("sale_num " + direction)
	}

	if req.RankType == 2 {
		listQuery = listQuery.Order("origin_sale_amount " + direction)
	}

	listQuery.Limit(req.PageSize).Offset((req.PageNo - 1) * req.PageSize).Find(&result)

	return result, total
}

// CountFreePayment 统计免单支付
func (r *StatisticsRepo) CountFreePayment(opts ...DBOption) model.StatisticsFreePaymentData {
	var result model.StatisticsFreePaymentData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsSale{}).
		Select(
			"COUNT(sale_order_uuid) AS total_order_num",
			"SUM(free_amount) AS total_free_amount",
		).
		Where("free_num > 0").
		Find(&result)

	return result
}

// CountFreePaymentDays 统计免单支付天数
func (r *StatisticsRepo) CountFreePaymentDays(opts ...DBOption) []model.StatisticsFreePaymentDaysData {
	var result []model.StatisticsFreePaymentDaysData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsSale{}).
		Select(
			"COUNT(sale_order_uuid) AS total_order_num",
			"SUM(free_amount) AS total_free_amount",
			"FROM_UNIXTIME(complete_time, '%Y-%m-%d') AS day",
		).
		Group("day").
		Order("day ASC").
		Find(&result)

	return result
}
