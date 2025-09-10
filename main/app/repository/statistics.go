package repository

import (
	"database/sql"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IStatisticsRepo interface {
	CountSale(opts ...DBOption) model.StatisticsSaleData                                                                       // 统计销售
	CountSaleDays(opts ...DBOption) []model.StatisticsSaleDaysData                                                             // 统计销售天数
	CountPayment(opts ...DBOption) []model.StatisticsPaymentData                                                               // 统计支付
	CountPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData                                                       // 统计支付天数
	CountTax(opts ...DBOption) []model.StatisticsTaxData                                                                       // 统计税类
	CountBuffetTax(opts ...DBOption) []model.StatisticsTaxData                                                                 // 统计自助餐税类
	CountBuffetDelayTax(opts ...DBOption) []model.StatisticsTaxData                                                            // 统计自助餐加钟税类
	CountCategory(categoryType int, language string, opts ...DBOption) (orderNum int64, result []model.StatisticsCategoryData) // 统计分类
	CountProduct(language string, opts ...DBOption) []model.StatisticsProductData                                              // 统计商品
	CountArea(opts ...DBOption) []model.StatisticsAreaData                                                                     // 统计区域
	CountAreaDays(opts ...DBOption) []model.StatisticsAreaDaysData                                                             // 统计区域
	Count7Days(opts ...DBOption) model.Statistics7DaysData                                                                     // 统计销售天数
	CountUnpaidOrder(opts ...DBOption) model.StatisticsUnpaidOrderData                                                         // 统计未结订单
	CountMemberNum(opts ...DBOption) int64                                                                                     // 统计会员数量
	CountMemberNumDays(opts ...DBOption) []model.CountMemberNumDaysResp                                                        // 统计会员数量天数
	CountMember(opts ...DBOption) model.StatisticsMemberData                                                                   // 统计会员
	CountMemberDays(opts ...DBOption) []model.StatisticsMemberDaysData                                                         // 统计会员天数
	CountMemberPayment(opts ...DBOption) []model.StatisticsPaymentData                                                         // 统计会员支付
	CountMemberPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData                                                 // 统计会员支付天数
	CountProductSale(req CountProductSaleRepoReq, opts ...DBOption) ([]model.StatisticsProductSaleData, int64)                 // 统计商品销售
	CountFreePayment(opts ...DBOption) model.StatisticsFreePaymentData                                                         // 统计免单支付
	CountFreePaymentDays(opts ...DBOption) []model.StatisticsFreePaymentDaysData                                               // 统计免单支付天数
	CountCancelOrder(opts ...DBOption) model.StatisticsCancelOrderData                                                         // 统计取消订单
	RankProduct(rankType int, language string, opts ...DBOption) []model.StatisticsProductData                                 // 统计商品排行
	SaveSale(sales []model.StatisticsSale) error                                                                               // 保存销售
	SavePayment(payments []model.StatisticsPayment) error                                                                      // 保存支付
	SaveProduct(products []model.StatisticsProduct) error                                                                      // 保存商品
	SaveCustomerType(customerTypes []model.StatisticsCustomerType) error                                                       // 保存客户类型
	SaveDelay(delays []model.StatisticsDelay) error                                                                            // 保存加钟
	DeleteSale(saleBillUuid uint64) error                                                                                      // 删除销售
	DeletePayment(saleBillUuid uint64) error                                                                                   // 删除支付
	DeleteProduct(saleBillUuid uint64) error                                                                                   // 删除商品
	DeleteCustomerType(saleBillUuid uint64) error                                                                              // 删除客户类型
	DeleteDelay(saleBillUuid uint64) error                                                                                     // 删除加钟
	SaveMember(member model.StatisticsMember) error                                                                            // 保存会员
	SaveMembers(members []model.StatisticsMember) error                                                                        // 保存会员
	SaveMemberPayment(payments []model.StatisticsMemberPayment) error                                                          // 保存会员支付
	DeleteMember(memberRechargeOrderUuid uint64) error                                                                         // 删除会员
	DeleteMemberPayment(memberRechargeOrderUuid uint64) error                                                                  // 删除会员支付
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
		"is_meger",
		"is_special",
		"is_takeout",
		"SUM(product_price + product_tax + service_fee + service_tax + payment_fee + extend_price) AS sale_amount",
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
		"SUM(IF(is_takeout = 1, payment_amount, 0)) AS takeout_sale_amount",
		"SUM(IF(is_takeout = 1, payment_amount - refund_amount - delivery_fee, 0)) AS takeout_business_amount",
		"SUM(IF(is_takeout = 1, refund_amount, 0)) AS takeout_refund_amount",
		"SUM(IF(is_takeout = 1, delivery_fee, 0)) AS takeout_delivery_fee",
		"SUM(IF(desk_uuid > 0, meal_num, 0)) AS meal_num",
		"SUM(IF(is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS order_amount",
		"SUM(payment_amount - refund_amount - refund_payment_balance) AS avg_order_amount",
		"SUM(IF(desk_uuid > 0 AND is_takeout = 0 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS desk_order_amount",
		"SUM(IF(desk_uuid > 0 AND is_takeout = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_desk_order_amount",
		"SUM(IF(desk_uuid = 0 AND is_takeout = 0 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS instant_order_amount",
		"SUM(IF(desk_uuid = 0 AND is_takeout = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_instant_order_amount",
		"SUM(IF(is_takeout = 1 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS takeout_order_amount",
		"SUM(IF(is_takeout = 1, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_takeout_order_amount",
		"complete_time",
	}
	// 统计销售
	countSaleSelect = []string{
		"SUM(t.sale_amount) AS total_sale_amount",                                                                        // 总销售额
		"SUM(t.received_amount) AS total_received_amount",                                                                // 总实收金额
		"SUM(t.product_price) AS total_product_price",                                                                    // 总商品原价
		"SUM(t.product_origin_price) AS total_product_origin_price",                                                      // 总原商品金额
		"SUM(t.product_num) AS total_product_num",                                                                        // 总商品数量
		"SUM(t.discount_member) AS total_discount_member",                                                                // 总会员折扣
		"SUM(t.business_amount) AS total_business_amount",                                                                // 总营业收入
		"SUM(t.payment_fee) AS total_payment_fee",                                                                        // 总支付手续费
		"SUM(t.service_fee) AS total_service_fee",                                                                        // 总服务费
		"SUM(t.tax) AS total_tax",                                                                                        // 总税额
		"SUM(t.refund_amount) AS total_refund_amount",                                                                    // 总退款金额
		"SUM(t.discount) AS total_discount",                                                                              // 总优惠折扣
		"SUM(t.gift_amount) AS total_gift_amount",                                                                        // 总赠菜金额
		"SUM(t.gift_num) AS total_gift_num",                                                                              // 总赠菜数量
		"SUM(t.free_amount) AS total_free_amount",                                                                        // 总免单金额
		"SUM(t.free_num) AS total_free_num",                                                                              // 总免单数量
		"SUM(IF(t.is_meger = 0, 1, 0)) AS total_order_num",                                                               // 总订单数量
		"SUM(t.takeout_sale_amount) AS total_takeout_sale_amount",                                                        // 总外送销售
		"SUM(t.takeout_business_amount) AS total_takeout_business_amount",                                                // 总外送营收
		"SUM(t.takeout_refund_amount) AS total_takeout_refund_amount",                                                    // 总外送退款金额
		"SUM(t.takeout_delivery_fee) AS total_takeout_delivery_fee",                                                      // 总外送配送费
		"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_meger = 0 THEN 1 END) AS total_desk_num",                               // 总桌台数量
		"SUM(t.desk_order_amount) AS total_desk_order_amount",                                                            // 总桌台订单金额
		"SUM(t.meal_num) AS total_meal_num",                                                                              // 总用餐人数
		"SUM(t.instant_order_amount) AS total_instant_order_amount",                                                      // 总即时订单金额
		"COUNT(CASE WHEN t.desk_uuid = 0 AND t.is_meger = 0 THEN 1 END) AS total_instant_order_num",                      // 总即时订单数量
		"SUM(t.takeout_order_amount) AS total_takeout_order_amount",                                                      // 总外送订单金额
		"COUNT(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END) AS total_takeout_order_num", // 总外送订单数量
		"MIN(CASE WHEN t.order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS min_order_amount",                                                                  // 最小订单金额
		"MAX(CASE WHEN t.order_amount > 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS max_order_amount",                                                                                        // 最大订单金额
		"SUM(t.avg_order_amount) / SUM(IF(t.is_meger = 0, 1, 0)) AS avg_order_amount",                                                                                                                       // 平均订单金额
		"MIN(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS min_desk_order_amount",          // 最小桌台订单金额
		"MAX(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount > 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS max_desk_order_amount",                                // 最大桌台订单金额
		"SUM(t.avg_desk_order_amount) / COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS avg_desk_order_amount",                                  // 平均桌台订单金额
		"MIN(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_amount ELSE NULL END) AS min_instant_order_amount", // 最小即时订单金额
		"MAX(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount > 0 THEN t.instant_order_amount ELSE NULL END) AS max_instant_order_amount",                                          // 最大即时订单金额
		"SUM(t.avg_instant_order_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS avg_instant_order_amount",                                                 // 平均即时订单金额
		"MIN(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.takeout_order_amount ELSE NULL END) AS min_takeout_order_amount",                     // 最小外送订单金额
		"MAX(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount > 0 THEN t.takeout_order_amount ELSE NULL END) AS max_takeout_order_amount",                                                              // 最大外送订单金额
		"SUM(t.avg_takeout_order_amount) / COUNT(CASE WHEN t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END) AS avg_takeout_order_amount",                                                                     // 平均外送订单金额
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
		"pm.id",
		"pm.sort",
		"pm.create_time",
		"sp.payment_method_uuid",
		"pm.payment_name AS payment_name",
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
		Order("pm.sort ASC").
		Order("pm.create_time DESC").
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

// CountTax 统计商品税类
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

// CountBuffetTax 统计自助餐税类
func (r *StatisticsRepo) CountBuffetTax(opts ...DBOption) []model.StatisticsTaxData {
	var result []model.StatisticsTaxData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsCustomerType{}).
		Select(
			"tax_rate",
			"SUM((product_price + tax_fee) * (product_num - refund_num)) AS total_product_amount",
			"SUM(tax_fee * (product_num - refund_num)) AS total_tax_fee",
		).Group("tax_rate").
		Find(&result)

	return result
}

// CountBuffetDelayTax 统计自助餐加钟税类
func (r *StatisticsRepo) CountBuffetDelayTax(opts ...DBOption) []model.StatisticsTaxData {
	var result []model.StatisticsTaxData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsDelay{}).
		Select(
			"tax_rate",
			"SUM((product_price + tax_fee) * (product_num - refund_num)) AS total_product_amount",
			"SUM(tax_fee * (product_num - refund_num)) AS total_tax_fee",
		).Group("tax_rate").
		Find(&result)

	return result
}

// CountCategory 统计分类
func (r *StatisticsRepo) CountCategory(categoryType int, language string, opts ...DBOption) (orderNum int64, result []model.StatisticsCategoryData) {
	db := r.db
	dbOrder := r.db
	for _, opt := range opts {
		db = opt(db)
		dbOrder = opt(dbOrder)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productBomTable := prefix + "product_bom as pb"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"

	// 统计订单数量
	dbOrder.Table(statisticsProductTable).Select("COUNT(DISTINCT sale_bill_uuid) AS order_num").Pluck("order_num", &orderNum)

	if categoryType != 2 {
		db.Table(statisticsProductTable).
			Select(
				"IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid) AS category_parent_uuid",
				"IF(pc.parent_uuid = 0, JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')), JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"'))) AS category_parent_name",
				"0 AS category_uuid",
				"'' AS category_name",
				"SUM(sp.product_num) AS sale_num",
				"SUM(sp.product_final_price * sp.product_num) AS sale_amount",
				"IF(pc.parent_uuid > 0, ppc.sort, pc.sort) AS sort",
			).
			Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
			Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
			Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
			Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
			Group("IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid)").
			Order("sort ASC").
			Order("category_parent_uuid DESC").
			Order("pp.uuid DESC").
			Find(&result)
	} else {
		db.Table(statisticsProductTable).
			Select(
				"pc.parent_uuid AS category_parent_uuid",
				"JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"')) AS category_parent_name",
				"pc.uuid AS category_uuid",
				"JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')) AS category_name",
				"SUM(sp.product_num) AS sale_num",
				"SUM(sp.product_final_price * sp.product_num) AS sale_amount",
			).
			Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
			Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
			Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
			Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
			Where("pc.parent_uuid > 0").
			Group("pc.uuid").
			Order("ppc.sort ASC").
			Order("ppc.uuid DESC").
			Order("pc.sort ASC").
			Order("pc.uuid DESC").
			Order("pp.uuid DESC").
			Find(&result)
	}

	return orderNum, result
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
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"

	db.Table(statisticsProductTable).
		Select(
			"CASE WHEN pp.name IS NOT NULL AND pp.name != '' THEN JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+language+"')) ELSE '' END AS product_name",
			"CASE WHEN pb.name IS NOT NULL AND pb.name != '' THEN JSON_UNQUOTE(JSON_EXTRACT(pb.name, '$."+language+"')) ELSE '' END AS flavor_name",
			"pb.price AS sale_price",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.product_final_price * sp.product_num) AS sale_amount",
			"IF(pc.parent_uuid = 0, pc.sort, ppc.sort) AS ppc_sort",
			"IF(pc.parent_uuid = 0, pc.create_time, ppc.create_time) AS ppc_create_time",
			"IF(pc.parent_uuid = 0, 0, pc.sort) AS pc_sort",
			"pp.product_type",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productBomTable + " ON sp.product_bom_uuid = pb.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid").
		Group("sp.product_bom_uuid").
		Order("ppc_sort ASC").
		Order("ppc_create_time DESC").
		Order("pc_sort ASC").
		Order("pc.create_time DESC").
		Order("pp.create_time DESC").
		Find(&result)

	return result
}

var (
	countAreaSelect = []string{
		"dr.name AS area_name",
		"SUM(ss.product_price + ss.product_tax + ss.service_fee + ss.service_tax + ss.payment_fee + ss.extend_price) AS area_sale_amount",
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
		Where("ss.desk_uuid > 0 and dr.uuid > 0 and dr.delete_time = 0").
		Group("dr.uuid").
		Order("dr.id ASC").
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
		Select(countAreaSelect, "dr.uuid AS area_id", "FROM_UNIXTIME(ss.complete_time, '%Y-%m-%d') AS day").
		Joins("LEFT JOIN " + deskTable + " ON ss.desk_uuid = d.uuid").
		Joins("LEFT JOIN " + deskRegionTable + " ON d.region_uuid = dr.uuid").
		Where("ss.desk_uuid > 0").
		Group("dr.uuid").
		Group("day").
		Order("day ASC").
		Order("dr.id ASC").
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

	query := db.Table(statisticsProductTable).
		Select(
			"sp.product_sale_price AS sale_price",
			"sp.product_package_uuid AS product_package_uuid",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.flavor_price * sp.product_num) AS sale_amount",
		).
		Where("sp.refund_time = 0").
		Group("sp.product_package_uuid")

	if rankType == constant.RankTypeSaleNum {
		query = query.Order("sale_num DESC")
	}

	if rankType == constant.RankTypeSaleAmount {
		query = query.Order("sale_amount DESC")
	}

	query = query.Limit(10)
	query.Find(&result)

	// 单独查询商品包名称
	result = r.QueryName(result, language)

	return result
}

// QueryName 查询商品包名称
func (r *StatisticsRepo) QueryName(result []model.StatisticsProductData, language string) []model.StatisticsProductData {
	productPackageUuids := make([]uint64, 0)
	for _, product := range result {
		if product.ProductPackageUuid.Valid {
			productPackageUuids = append(productPackageUuids, uint64(product.ProductPackageUuid.Int64))
		}
	}

	productPackageNameMap := r.QueryProductPackageName(productPackageUuids, language)
	for i, product := range result {
		result[i].ProductName = sql.NullString{
			String: productPackageNameMap[uint64(product.ProductPackageUuid.Int64)],
			Valid:  product.ProductPackageUuid.Valid,
		}
	}
	return result
}

// QueryProductPackageName 查询商品包名称
func (r *StatisticsRepo) QueryProductPackageName(uuids []uint64, language string) map[uint64]string {
	result := make(map[uint64]string)
	var productPackages []model.ProductPackage
	r.db.Model(&model.ProductPackage{}).Preload("MultiLanguageName").Where("uuid IN (?)", uuids).Find(&productPackages)
	for _, productPackage := range productPackages {
		result[productPackage.Uuid] = productPackage.MultiLanguageName.GetNameByLang(language)
	}
	return result
}

// Count7Days 统计销售天数
func (r *StatisticsRepo) Count7Days(opts ...DBOption) model.Statistics7DaysData {
	var result model.Statistics7DaysData

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.StatisticsSale{}).Select(
		"count(DISTINCT sale_order_uuid) as total_order_num",
		"FROM_UNIXTIME(complete_time, '%Y-%m-%d') day",
		"SUM(payment_amount) AS TotalReceivedAmount",
	).Find(&result)

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

// SaveCustomerType 保存客户类型
func (r *StatisticsRepo) SaveCustomerType(customerTypes []model.StatisticsCustomerType) error {
	return r.db.Create(&customerTypes).Error
}

// DeleteCustomerType 删除客户类型
func (r *StatisticsRepo) DeleteCustomerType(saleBillUuid uint64) error {
	return r.db.Where("sale_bill_uuid = ?", saleBillUuid).Delete(&model.StatisticsCustomerType{}).Error
}

// SaveDelay 保存加钟
func (r *StatisticsRepo) SaveDelay(delays []model.StatisticsDelay) error {
	return r.db.Create(&delays).Error
}

// DeleteDelay 删除加钟
func (r *StatisticsRepo) DeleteDelay(saleBillUuid uint64) error {
	return r.db.Where("sale_bill_uuid = ?", saleBillUuid).Delete(&model.StatisticsDelay{}).Error
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

// SaveMembers 保存会员
func (r *StatisticsRepo) SaveMembers(members []model.StatisticsMember) error {
	return r.db.Create(&members).Error
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
		"pm.id",
		"pm.sort",
		"pm.create_time",
		"smp.payment_method_uuid",
		"pm.payment_name AS payment_name",
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
		db2 = opt(db2)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"
	productPackageTable := prefix + "product_package as pp"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"
	deskTable := prefix + "desk as d"

	var total int64
	db = db.Table(statisticsProductTable).
		Select("COUNT(DISTINCT sp.product_package_uuid) AS total").
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid")

	if req.AreaUuid > 0 {
		db.Joins("LEFT JOIN " + deskTable + " ON sp.desk_uuid = d.uuid")
		db.Where("d.region_uuid = ?", req.AreaUuid)
	}
	if req.CategoryUuid > 0 {
		db.Where("pp.category_uuid = ? OR pp.category_uuid IN (?)", req.CategoryUuid, r.db.Table(productCategoryTable).Select("pc.uuid").Where("pc.parent_uuid = ?", req.CategoryUuid))
	}
	if req.ProductName != "" {
		db.Where("JSON_UNQUOTE(JSON_EXTRACT(pp.name, ?)) LIKE ?", "$."+req.Language, "%"+req.ProductName+"%")
	}
	db.Pluck("total", &total)

	listQuery := db2.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+req.Language+"')) AS product_name",
			"JSON_UNQUOTE(JSON_EXTRACT(pc.name, '$."+req.Language+"')) AS category_name",
			"JSON_UNQUOTE(JSON_EXTRACT(ppc.name, '$."+req.Language+"')) AS category_parent_name",
			"SUM(sp.product_num) AS sale_num",
			"SUM((sp.product_price + sp.tax_fee + sp.service_fee + service_tax) * sp.product_num) AS origin_sale_amount",
			"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num))) AS actual_sale_amount",
			"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, (sp.product_final_price - sp.tax_fee - sp.service_tax) * (sp.product_num - sp.refund_num))) AS business_amount",
			"SUM(IF(sp.free_num > 0, sp.free_num, sp.give_num)) AS give_num",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid")

	if req.AreaUuid > 0 {
		listQuery.Joins("LEFT JOIN " + deskTable + " ON sp.desk_uuid = d.uuid")
		listQuery.Where("d.region_uuid = ?", req.AreaUuid)
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

// CountCancelOrder 统计取消订单
func (r *StatisticsRepo) CountCancelOrder(opts ...DBOption) model.StatisticsCancelOrderData {
	var result model.StatisticsCancelOrderData
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db.Model(&model.SaleBill{}).
		Select("COUNT(uuid) AS total_cancel_order_num", "SUM(origin_amount) AS total_cancel_order_amount").
		Where("status = ?", constant.SaleOrderStatusCanceled).
		Where("production_time > 0").
		Where("bill_type IN (?)", []uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}).
		Find(&result)

	return result
}
