package repository

import (
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
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
	Count7Days(opts ...DBOption) []struct {
		CompleteTime        int64   `gorm:"column:complete_time"`
		SaleOrderUuid       uint64  `gorm:"column:sale_order_uuid"`
		TotalReceivedAmount float64 `gorm:"column:total_received_amount"`
	} // 统计销售天数（返回原始数据，不进行日期分组）
	CountUnpaidOrder(opts ...DBOption) model.StatisticsUnpaidOrderData                                                                                                                                                    // 统计未结订单
	CountMemberNum(opts ...DBOption) int64                                                                                                                                                                                // 统计会员数量
	CountMemberNumDays(opts ...DBOption) []model.CountMemberNumDaysResp                                                                                                                                                   // 统计会员数量天数
	CountMember(opts ...DBOption) model.StatisticsMemberData                                                                                                                                                              // 统计会员
	CountMemberDays(opts ...DBOption) []model.StatisticsMemberDaysData                                                                                                                                                    // 统计会员天数
	CountMemberPayment(opts ...DBOption) []model.StatisticsPaymentData                                                                                                                                                    // 统计会员支付
	CountMemberPaymentDays(opts ...DBOption) []model.StatisticsPaymentDaysData                                                                                                                                            // 统计会员支付天数
	CountProductSale(req CountProductSaleRepoReq, opts ...DBOption) ([]model.StatisticsProductSaleData, int64)                                                                                                            // 统计商品销售
	CountFreePayment(opts ...DBOption) model.StatisticsFreePaymentData                                                                                                                                                    // 统计免单支付
	CountFreePaymentDays(opts ...DBOption) []model.StatisticsFreePaymentDaysData                                                                                                                                          // 统计免单支付天数
	CountCancelOrder(opts ...DBOption) model.StatisticsCancelOrderData                                                                                                                                                    // 统计取消订单
	CountBusinessTimePeriod(req CountBusinessTimePeriodReq, opts ...DBOption) (int64, []model.StatisticsBusinessTimePeriodData)                                                                                           // 统计营业时段
	CountBusinessSummary(req CountBusinessSummaryReq) (int64, []model.StatisticsBusinessSummaryData)                                                                                                                      // 统计综合运用数据
	CountBusinessPaymentMethod(req CountBusinessPaymentMethodReq) (int64, []model.StatisticsBusinessPaymentMethodData)                                                                                                    // 统计支付方式
	CountRefundSummary(req CountRefundSummaryReq) (int64, []model.StatisticsRefundSummaryData)                                                                                                                            // 统计退款金额汇总
	CountChannelSale(startTime, endTime int64, excludeDataManage bool) (map[string]*model.ChannelSaleRepoResult, error)                                                                                                   // 统计渠道营业数据
	CountUserAnalysis(startTime, endTime int64, language string, enableNationality bool, enableCashierOrder bool, enableTableOrder bool, excludeDataManage bool, opts ...DBOption) (*model.UserAnalysisRepoResult, error) // 统计用户分析数据
	RankProduct(rankType int, language string, timeStart int64, timeEnd int64, opts ...DBOption) []model.StatisticsProductData                                                                                            // 统计商品排行
	SaveSale(sales []model.StatisticsSale) error                                                                                                                                                                          // 保存销售
	SavePayment(payments []model.StatisticsPayment) error                                                                                                                                                                 // 保存支付
	SaveProduct(products []model.StatisticsProduct) error                                                                                                                                                                 // 保存商品
	SaveCustomerType(customerTypes []model.StatisticsCustomerType) error                                                                                                                                                  // 保存客户类型
	SaveDelay(delays []model.StatisticsDelay) error                                                                                                                                                                       // 保存加钟
	DeleteSale(saleBillUuid uint64) error                                                                                                                                                                                 // 删除销售
	DeletePayment(saleBillUuid uint64) error                                                                                                                                                                              // 删除支付
	DeleteProduct(saleBillUuid uint64) error                                                                                                                                                                              // 删除商品
	DeleteCustomerType(saleBillUuid uint64) error                                                                                                                                                                         // 删除客户类型
	DeleteDelay(saleBillUuid uint64) error                                                                                                                                                                                // 删除加钟
	SaveMember(member model.StatisticsMember) error                                                                                                                                                                       // 保存会员
	SaveMembers(members []model.StatisticsMember) error                                                                                                                                                                   // 保存会员
	SaveMemberPayment(payments []model.StatisticsMemberPayment) error                                                                                                                                                     // 保存会员支付
	DeleteMember(memberRechargeOrderUuid uint64) error                                                                                                                                                                    // 删除会员
	DeleteMemberPayment(memberRechargeOrderUuid uint64) error                                                                                                                                                             // 删除会员支付
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
		"order_source_uuid",
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
		"COALESCE(MAX(CASE WHEN desk_uuid > 0 THEN meal_num ELSE NULL END), 0) AS meal_num",
		"SUM(IF(is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS order_amount",
		"SUM(payment_amount - refund_amount - refund_payment_balance) AS avg_order_amount",
		"SUM(IF(desk_uuid > 0 AND is_takeout = 0 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS desk_order_amount",
		"SUM(IF(desk_uuid > 0 AND is_takeout = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_desk_order_amount",
		"SUM(IF(desk_uuid = 0 AND order_source_uuid = 0 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS instant_order_amount",
		"SUM(IF(desk_uuid = 0 AND order_source_uuid = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_instant_order_amount",
		"SUM(IF(desk_uuid = 0 AND order_source_uuid > 0 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS instant_order_takeaway_amount",
		"SUM(IF(desk_uuid = 0 AND order_source_uuid > 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_instant_order_takeaway_amount",
		"SUM(IF(is_takeout = 1 AND is_meger = 0, payment_amount - refund_amount - refund_payment_balance, 0)) AS takeout_order_amount",
		"SUM(IF(is_takeout = 1, payment_amount - refund_amount - refund_payment_balance, 0)) AS avg_takeout_order_amount",
		"complete_time",
	}
	// 统计销售
	countSaleSelect = []string{
		"SUM(t.sale_amount) AS total_sale_amount",                                          // 总销售额
		"SUM(t.received_amount) AS total_received_amount",                                  // 总实收金额
		"SUM(t.product_price) AS total_product_price",                                      // 总商品原价
		"SUM(t.product_origin_price) AS total_product_origin_price",                        // 总原商品金额
		"SUM(t.product_num) AS total_product_num",                                          // 总商品数量
		"SUM(t.discount_member) AS total_discount_member",                                  // 总会员折扣
		"SUM(t.business_amount) AS total_business_amount",                                  // 总营业收入
		"SUM(t.payment_fee) AS total_payment_fee",                                          // 总支付手续费
		"SUM(t.service_fee) AS total_service_fee",                                          // 总服务费
		"SUM(t.tax) AS total_tax",                                                          // 总税额
		"SUM(t.refund_amount) AS total_refund_amount",                                      // 总退款金额
		"SUM(t.discount) AS total_discount",                                                // 总优惠折扣
		"SUM(t.gift_amount) AS total_gift_amount",                                          // 总赠菜金额
		"SUM(t.gift_num) AS total_gift_num",                                                // 总赠菜数量
		"SUM(t.free_amount) AS total_free_amount",                                          // 总免单金额
		"SUM(t.free_num) AS total_free_num",                                                // 总免单数量
		"SUM(t.order_amount) AS total_order_amount",                                        // 总订单金额
		"SUM(IF(t.is_meger = 0, 1, 0)) AS total_order_num",                                 // 总订单数量
		"SUM(t.takeout_sale_amount) AS total_takeout_sale_amount",                          // 总外送销售
		"SUM(t.takeout_business_amount) AS total_takeout_business_amount",                  // 总外送营收
		"SUM(t.takeout_refund_amount) AS total_takeout_refund_amount",                      // 总外送退款金额
		"SUM(t.takeout_delivery_fee) AS total_takeout_delivery_fee",                        // 总外送配送费
		"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_meger = 0 THEN 1 END) AS total_desk_num", // 总桌台数量
		"SUM(t.desk_order_amount) AS total_desk_order_amount",                              // 总桌台订单金额
		"SUM(t.meal_num) AS total_meal_num",                                                // 总用餐人数
		"SUM(t.instant_order_amount) AS total_instant_order_amount",                        // 即时订单金额（店内）
		"SUM(t.instant_order_takeaway_amount) AS total_instant_order_takeaway_amount",      // 即时订单-外卖来源金额
		"COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_meger = 0 THEN 1 END) AS total_instant_order_num",                                                                                                                                    // 总即时订单数量（店内）
		"COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_meger = 0 THEN 1 END) AS total_instant_order_takeaway_num",                                                                                                                           // 总即时订单数量（外卖来源）
		"SUM(t.takeout_order_amount) AS total_takeout_order_amount",                                                                                                                                                                                                // 总外送订单金额
		"COUNT(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END) AS total_takeout_order_num",                                                                                                                                           // 总外送订单数量
		"MIN(CASE WHEN t.order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS min_order_amount",                                                                                                                         // 最小订单金额
		"MAX(CASE WHEN t.order_amount > 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS max_order_amount",                                                                                                                                               // 最大订单金额
		"ROUND(SUM(t.avg_order_amount) / SUM(IF(t.is_meger = 0, 1, 0)), 2) AS avg_order_amount",                                                                                                                                                                    // 平均订单金额
		"MIN(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS min_desk_order_amount",                                                                 // 最小桌台订单金额
		"MAX(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount > 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS max_desk_order_amount",                                                                                       // 最大桌台订单金额
		"ROUND(SUM(t.avg_desk_order_amount) / COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END), 2) AS avg_desk_order_amount",                                                                               // 平均桌台订单金额
		"MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_amount ELSE NULL END) AS min_instant_order_amount",                            // 最小即时订单金额（店内）
		"MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount > 0 THEN t.instant_order_amount ELSE NULL END) AS max_instant_order_amount",                                                                     // 最大即时订单金额（店内）
		"ROUND(SUM(t.avg_instant_order_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_meger = 0 AND t.instant_order_amount > 0 THEN 1 END), 2) AS avg_instant_order_amount",                                                        // 平均即时订单金额（店内）
		"MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.instant_order_takeaway_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_takeaway_amount ELSE NULL END) AS min_instant_order_takeaway_amount", // 最小即时订单金额（外卖来源）
		"MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.instant_order_takeaway_amount > 0 THEN t.instant_order_takeaway_amount ELSE NULL END) AS max_instant_order_takeaway_amount",                                          // 最大即时订单金额（外卖来源）
		"ROUND(SUM(t.avg_instant_order_takeaway_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_meger = 0 AND t.instant_order_takeaway_amount > 0 THEN 1 END), 2) AS avg_instant_order_takeaway_amount",                             // 平均即时订单金额（外卖来源）
		"MIN(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.takeout_order_amount ELSE NULL END) AS min_takeout_order_amount",                                                                            // 最小外送订单金额
		"MAX(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount > 0 THEN t.takeout_order_amount ELSE NULL END) AS max_takeout_order_amount",                                                                                                                     // 最大外送订单金额
		"ROUND(SUM(t.avg_takeout_order_amount) / COUNT(CASE WHEN t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END), 2) AS avg_takeout_order_amount",                                                                                                                  // 平均外送订单金额
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
		"pm.erpnext_payment AS erpnext_payment",
		"pm.erpnext_payment_id AS erpnext_payment_id",
		"pm.source AS source",
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
	// 获取语言，确保语言是支持的语言
	// GetLocaleType 会将无效的语言（包括空字符串）转换为默认值 LocaleZHTW
	locale := constant.LocaleList.GetLocaleType(language)
	language = string(locale)

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
				"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num))) AS sale_amount",
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
				"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num))) AS sale_amount",
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
	// 获取语言，确保语言是支持的语言
	// GetLocaleType 会将无效的语言（包括空字符串）转换为默认值 LocaleZHTW
	locale := constant.LocaleList.GetLocaleType(language)
	language = string(locale)

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
			"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num))) AS sale_amount",
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
func (r *StatisticsRepo) RankProduct(rankType int, language string, timeStart int64, timeEnd int64, opts ...DBOption) []model.StatisticsProductData {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	prefix := config.Database.TablePrefix
	statisticsProductTable := prefix + "statistics_product as sp"

	// 查询1：原有逻辑（ttpos_statistics_product 表）
	// 保持原有逻辑完全不变，应用所有 opts 条件
	var statisticsData []model.StatisticsProductData
	statisticsQuery := db.Table(statisticsProductTable).
		Select(
			"sp.product_sale_price AS sale_price",
			"sp.product_package_uuid AS product_package_uuid",
			"SUM(sp.product_num) AS sale_num",
			"SUM(sp.flavor_price * sp.product_num) AS sale_amount",
		).
		Where("sp.refund_time = 0").
		Group("sp.product_package_uuid")
	statisticsQuery.Find(&statisticsData)

	// 查询2：外卖订单商品（通过 StatisticsTakeoutRepo 查询）
	var takeoutData []model.StatisticsProductData
	takeoutRepo := NewStatisticsTakeoutRepo(r.db)
	takeoutData = takeoutRepo.RankTakeoutProduct(CountTakeoutReq{
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
	})

	// 在应用层合并两个表的数据
	// 使用 map 按 product_package_uuid 合并，保持原有逻辑不变
	mergedData := make(map[uint64]*model.StatisticsProductData)

	// 先处理统计表数据（保持原有逻辑，sale_price 使用统计表的值）
	for i := range statisticsData {
		if statisticsData[i].ProductPackageUuid.Valid {
			uuid := uint64(statisticsData[i].ProductPackageUuid.Int64)
			mergedData[uuid] = &statisticsData[i]
		}
	}

	// 再处理外卖订单数据，累加到统计表数据上
	// 使用 decimal 进行精确计算，避免浮点数精度问题
	for _, item := range takeoutData {
		if !item.ProductPackageUuid.Valid {
			continue
		}
		uuid := uint64(item.ProductPackageUuid.Int64)
		if existing, exists := mergedData[uuid]; exists {
			// 如果已存在，累加 sale_num 和 sale_amount
			// sale_price 保持使用统计表的值（原有逻辑）
			if item.SaleNum.Valid {
				if existing.SaleNum.Valid {
					// 使用 decimal 进行精确累加
					existingDecimal := decimal.NewFromFloat(existing.SaleNum.Float64)
					itemDecimal := decimal.NewFromFloat(item.SaleNum.Float64)
					sumDecimal := existingDecimal.Add(itemDecimal)
					existing.SaleNum.Float64 = sumDecimal.InexactFloat64()
				} else {
					existing.SaleNum = item.SaleNum
				}
			}
			if item.SaleAmount.Valid {
				if existing.SaleAmount.Valid {
					// 使用 decimal 进行精确累加
					existingDecimal := decimal.NewFromFloat(existing.SaleAmount.Float64)
					itemDecimal := decimal.NewFromFloat(item.SaleAmount.Float64)
					sumDecimal := existingDecimal.Add(itemDecimal)
					existing.SaleAmount.Float64 = sumDecimal.InexactFloat64()
				} else {
					existing.SaleAmount = item.SaleAmount
				}
			}
		} else {
			// 如果不存在，直接添加（只有外卖订单数据）
			mergedData[uuid] = &item
		}
	}

	// 转换为切片
	result := make([]model.StatisticsProductData, 0, len(mergedData))
	for _, item := range mergedData {
		result = append(result, *item)
	}

	// 排序
	if rankType == constant.RankTypeSaleNum {
		sort.Slice(result, func(i, j int) bool {
			var numI, numJ float64
			if result[i].SaleNum.Valid {
				numI = result[i].SaleNum.Float64
			}
			if result[j].SaleNum.Valid {
				numJ = result[j].SaleNum.Float64
			}
			return numI > numJ
		})
	} else if rankType == constant.RankTypeSaleAmount {
		sort.Slice(result, func(i, j int) bool {
			var amountI, amountJ float64
			if result[i].SaleAmount.Valid {
				amountI = result[i].SaleAmount.Float64
			}
			if result[j].SaleAmount.Valid {
				amountJ = result[j].SaleAmount.Float64
			}
			return amountI > amountJ
		})
	}

	// 限制前10条
	if len(result) > 10 {
		result = result[:10]
	}

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
// 返回原始数据（包含 complete_time 时间戳和 sale_order_uuid），不进行日期分组
// 日期分组应在应用层使用商家时区进行
func (r *StatisticsRepo) Count7Days(opts ...DBOption) []struct {
	CompleteTime        int64   `gorm:"column:complete_time"`
	SaleOrderUuid       uint64  `gorm:"column:sale_order_uuid"`
	TotalReceivedAmount float64 `gorm:"column:total_received_amount"`
} {
	var result []struct {
		CompleteTime        int64   `gorm:"column:complete_time"`
		SaleOrderUuid       uint64  `gorm:"column:sale_order_uuid"`
		TotalReceivedAmount float64 `gorm:"column:total_received_amount"`
	}

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	// 返回每个订单的完成时间和支付金额
	// 不在这里进行日期分组，避免使用数据库时区
	db.Model(&model.StatisticsSale{}).
		Select(
			"complete_time",
			"sale_order_uuid",
			"(payment_amount - refund_amount - payment_balance) AS total_received_amount",
		).
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
		"SUM(payment_fee - refund_fee) AS total_payment_fee",
		"SUM(refund_amount + refund_fee) AS total_refund_amount",
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
		"pm.erpnext_payment AS erpnext_payment",
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
	CategoryUuids []uint64
	ProductName   string
	OrderTypes    []uint
	OrderSource   int
	StartTime     int64 // 查询开始时间（用于外卖订单筛选）
	EndTime       int64 // 查询结束时间（用于外卖订单筛选）
}

func (r *StatisticsRepo) CountProductSale(req CountProductSaleRepoReq, opts ...DBOption) ([]model.StatisticsProductSaleData, int64) {
	// 获取语言，确保语言是支持的语言
	// GetLocaleType 会将无效的语言（包括空字符串）转换为默认值 LocaleZHTW
	locale := constant.LocaleList.GetLocaleType(req.Language)
	req.Language = string(locale)

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
	saleBillTable := prefix + "sale_bill as sb"
	takeoutOrderTable := prefix + "takeout_order as to_order"
	takeoutOrderItemTable := prefix + "takeout_order_item as to_item"

	// 构建商品分类筛选条件（用于店内和外卖查询）
	var allCategoryUuids []uint64
	if len(req.CategoryUuids) > 0 {
		for _, categoryUuid := range req.CategoryUuids {
			allCategoryUuids = append(allCategoryUuids, categoryUuid)
			var subCategoryUuids []uint64
			r.db.Table(productCategoryTable).
				Select("uuid").
				Where("parent_uuid = ?", categoryUuid).
				Pluck("uuid", &subCategoryUuids)
			allCategoryUuids = append(allCategoryUuids, subCategoryUuids...)
		}
	} else if req.CategoryUuid > 0 {
		allCategoryUuids = append(allCategoryUuids, req.CategoryUuid)
		var subCategoryUuids []uint64
		r.db.Table(productCategoryTable).
			Select("uuid").
			Where("parent_uuid = ?", req.CategoryUuid).
			Pluck("uuid", &subCategoryUuids)
		allCategoryUuids = append(allCategoryUuids, subCategoryUuids...)
	}

	// 构建店内商品销售查询（子查询）
	storeQuery := db2.Table(statisticsProductTable).
		Select(
			"JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+req.Language+"')) AS product_name",
			"JSON_UNQUOTE(JSON_EXTRACT(pc.name, '$."+req.Language+"')) AS category_name",
			"JSON_UNQUOTE(JSON_EXTRACT(ppc.name, '$."+req.Language+"')) AS category_parent_name",
			"sp.product_package_uuid AS product_package_uuid",
			"SUM(sp.product_num) AS sale_num",
			"SUM((sp.product_price + sp.tax_fee + sp.service_fee + sp.service_tax) * sp.product_num) AS origin_sale_amount",
			"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num))) AS actual_sale_amount",
			"SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, (sp.product_final_price - sp.tax_fee - sp.service_tax) * (sp.product_num - sp.refund_num))) AS business_amount",
			"SUM(IF(sp.free_num > 0, sp.free_num, sp.give_num)) AS give_num",
		).
		Joins("LEFT JOIN " + productPackageTable + " ON sp.product_package_uuid = pp.uuid").
		Joins("LEFT JOIN " + productCategoryTable + " ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN " + productParentCategoryTable + " ON pc.parent_uuid = ppc.uuid")

	// 订单类型筛选：需要关联 sale_bill 表
	needJoinSaleBill := len(req.OrderTypes) > 0 || (req.OrderSource > 0 && containsOrderType(req.OrderTypes, 1))
	if needJoinSaleBill {
		storeQuery.Joins("LEFT JOIN " + saleBillTable + " ON sp.sale_bill_uuid = sb.uuid")
	}

	if req.AreaUuid > 0 {
		storeQuery.Joins("LEFT JOIN " + deskTable + " ON sp.desk_uuid = d.uuid")
		storeQuery.Where("d.region_uuid = ?", req.AreaUuid)
	}

	// 商品分类筛选
	if len(allCategoryUuids) > 0 {
		storeQuery.Where("pp.category_uuid IN (?)", allCategoryUuids)
	}

	if req.ProductName != "" {
		storeQuery.Where("JSON_UNQUOTE(JSON_EXTRACT(pp.name, ?)) LIKE ?", "$."+req.Language, "%"+req.ProductName+"%")
	}

	// 订单类型筛选
	if len(req.OrderTypes) > 0 {
		var billTypes []uint
		for _, orderType := range req.OrderTypes {
			switch orderType {
			case 1: // 点餐订单
				billTypes = append(billTypes, constant.SaleBillTypeInstant)
			case 2: // 桌台订单
				billTypes = append(billTypes, constant.SaleBillTypeDesk)
			case 3: // 外送订单
				billTypes = append(billTypes, constant.SaleBillTypeTakeout)
			}
		}
		if len(billTypes) > 0 {
			storeQuery.Where("sb.bill_type IN (?)", billTypes)
		}
	}

	// 订单来源筛选（仅在订单类型包含点餐订单时生效）
	if req.OrderSource > 0 && containsOrderType(req.OrderTypes, 1) {
		if req.OrderSource == 1 {
			// 1=店内
			storeQuery.Where("sb.order_source_uuid = 0")
		} else if req.OrderSource == 2 {
			// 2=外卖
			storeQuery.Where("sb.order_source_uuid > 0")
		}
	}

	storeQuery.Group("sp.product_package_uuid")

	// 构建外卖商品销售查询（子查询）
	// 使用 validOrderStates 状态筛选：10=已接单配餐中, 20=待骑手接单, 30=骑手配送中, 40=已完成, 60=已取消
	// 仅统计 accepted_time > 0 的订单（有效状态和取消状态都需要接单后才能统计）
	// 实际销售额和营业收入需要减去已取消订单的金额（已接单并且已取消的订单，order_state = 60）
	validStatesStr := buildStateInCondition(validOrderStates)
	takeoutQuery := r.db.Table(takeoutOrderItemTable).
		Select(
			"COALESCE(JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$."+req.Language+"')), to_item.item_name) AS product_name",
			"COALESCE(JSON_UNQUOTE(JSON_EXTRACT(pc.name, '$."+req.Language+"')), IF(to_item.ttpos_category_name IS NULL OR to_item.ttpos_category_name = '', NULL, JSON_UNQUOTE(JSON_EXTRACT(to_item.ttpos_category_name, '$."+req.Language+"')))) AS category_name",
			"COALESCE(JSON_UNQUOTE(JSON_EXTRACT(ppc.name, '$."+req.Language+"')), IF(to_item.ttpos_parent_category_name IS NULL OR to_item.ttpos_parent_category_name = '', NULL, JSON_UNQUOTE(JSON_EXTRACT(to_item.ttpos_parent_category_name, '$."+req.Language+"')))) AS category_parent_name",
			"COALESCE(to_item.ttpos_product_package_uuid, 0) AS product_package_uuid",
			"SUM(to_item.quantity) AS sale_num",
			"SUM(to_item.price * to_item.quantity) AS origin_sale_amount",
			"SUM(IF(to_order.order_state = 60, 0, to_item.price * to_item.quantity)) AS actual_sale_amount",
			"SUM(IF(to_order.order_state = 60, 0, (to_item.price - to_item.tax) * to_item.quantity)) AS business_amount",
			"0 AS give_num",
		).
		Joins("INNER JOIN "+takeoutOrderTable+" ON to_item.takeout_order_uuid = to_order.uuid").
		Joins("LEFT JOIN "+productPackageTable+" ON to_item.ttpos_product_package_uuid = pp.uuid").
		Joins("LEFT JOIN "+productCategoryTable+" ON COALESCE(pp.category_uuid, to_item.ttpos_category_uuid) = pc.uuid").
		Joins("LEFT JOIN "+productParentCategoryTable+" ON pc.parent_uuid = ppc.uuid").
		Where("to_order.delete_time = ?", constant.NotDeleted).
		Where("to_item.delete_time = ?", constant.NotDeleted).
		Where("to_order.order_state IN " + validStatesStr)

	// 时间范围筛选（使用 accepted_time）
	if req.StartTime > 0 && req.EndTime > 0 {
		takeoutQuery.Where("to_order.accepted_time >= ? AND to_order.accepted_time <= ?", req.StartTime, req.EndTime)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	takeoutQuery.Where("to_order.accepted_time > 0")

	// 商品分类筛选
	if len(allCategoryUuids) > 0 {
		takeoutQuery.Where("(pp.category_uuid IN (?) OR to_item.ttpos_category_uuid IN (?))", allCategoryUuids, allCategoryUuids)
	}

	// 商品名称筛选
	if req.ProductName != "" {
		takeoutQuery.Where("(JSON_UNQUOTE(JSON_EXTRACT(pp.name, ?)) LIKE ? OR to_item.item_name LIKE ?)", "$."+req.Language, "%"+req.ProductName+"%", "%"+req.ProductName+"%")
	}

	takeoutQuery.Group("COALESCE(to_item.ttpos_product_package_uuid, 0)")

	// 临时结构，包含 product_package_uuid
	type productSaleDataWithUuid struct {
		model.StatisticsProductSaleData
		ProductPackageUuid uint64 `gorm:"column:product_package_uuid"`
	}

	// 执行店内商品销售查询
	var storeData []productSaleDataWithUuid
	storeQuery.Find(&storeData)

	// 执行外卖商品销售查询
	var takeoutData []productSaleDataWithUuid
	takeoutQuery.Find(&takeoutData)

	// 合并数据：按 product_package_uuid 分组
	productMap := make(map[uint64]*model.StatisticsProductSaleData)
	for i := range storeData {
		item := &storeData[i]
		key := item.ProductPackageUuid
		if existing, exists := productMap[key]; exists {
			// 合并数据
			existing.SaleNum.Float64 += item.SaleNum.Float64
			existing.OriginSaleAmount.Float64 += item.OriginSaleAmount.Float64
			existing.ActualSaleAmount.Float64 += item.ActualSaleAmount.Float64
			existing.BusinessAmount.Float64 += item.BusinessAmount.Float64
			existing.GiveNum.Float64 += item.GiveNum.Float64
		} else {
			// 创建新条目
			newItem := item.StatisticsProductSaleData
			productMap[key] = &newItem
		}
	}

	// 合并外卖数据
	for i := range takeoutData {
		item := &takeoutData[i]
		key := item.ProductPackageUuid
		if existing, exists := productMap[key]; exists {
			// 合并数据
			existing.SaleNum.Float64 += item.SaleNum.Float64
			existing.OriginSaleAmount.Float64 += item.OriginSaleAmount.Float64
			existing.ActualSaleAmount.Float64 += item.ActualSaleAmount.Float64
			existing.BusinessAmount.Float64 += item.BusinessAmount.Float64
			existing.GiveNum.Float64 += item.GiveNum.Float64
		} else {
			// 创建新条目
			newItem := item.StatisticsProductSaleData
			productMap[key] = &newItem
		}
	}

	// 转换为切片
	var combinedData []model.StatisticsProductSaleData
	for _, item := range productMap {
		combinedData = append(combinedData, *item)
	}

	// 排序
	direction := 1
	if req.RankDirection == 1 {
		direction = -1
	}
	sort.Slice(combinedData, func(i, j int) bool {
		if req.RankType == 2 {
			// 按原价销售额排序
			if direction > 0 {
				return combinedData[i].OriginSaleAmount.Float64 > combinedData[j].OriginSaleAmount.Float64
			}
			return combinedData[i].OriginSaleAmount.Float64 < combinedData[j].OriginSaleAmount.Float64
		}
		// 按销售数量排序
		if direction > 0 {
			return combinedData[i].SaleNum.Float64 > combinedData[j].SaleNum.Float64
		}
		return combinedData[i].SaleNum.Float64 < combinedData[j].SaleNum.Float64
	})

	// 计算总数
	total := int64(len(combinedData))

	// 分页
	start := (req.PageNo - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(combinedData) {
		start = len(combinedData)
	}
	if end > len(combinedData) {
		end = len(combinedData)
	}
	if start < end {
		result = combinedData[start:end]
	} else {
		result = []model.StatisticsProductSaleData{}
	}

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

// CountBusinessTimePeriodRepoReq 统计营业时段请求
type CountBusinessTimePeriodReq struct {
	StartTime, EndTime           int64 // 查询开始时间, 查询结束时间
	PeriodSeconds                int   // 时段: 1=15分钟, 2=30分钟, 3=1小时
	IsCreateTime                 bool  // 是否是创建时间
	PageNo, PageSize             int   // 页码, 每页大小
	IsDesk, IsInstant, IsTakeout bool  // 是否是桌台订单, 是否是点餐订单, 是否是配送订单
	IsDelivery                   bool  // 是否包含外卖订单
}

// CountBusinessTimePeriod 统计营业时段
func (r *StatisticsRepo) CountBusinessTimePeriod(req CountBusinessTimePeriodReq, opts ...DBOption) (int64, []model.StatisticsBusinessTimePeriodData) {
	// 确定时间字段
	timeField := "sb.finish_time"
	if req.IsCreateTime {
		timeField = "sb.create_time"
	}

	// 构建基础查询条件
	baseQuery := r.db.Table("ttpos_sale_bill AS sb").
		Joins("LEFT JOIN ttpos_sale_order AS so ON sb.uuid = so.sale_bill_uuid AND so.delete_time = ? AND so.status = ?", constant.NotDeleted, constant.SaleOrderStatusFinish).
		Joins("LEFT JOIN ttpos_return_order AS ro ON so.uuid = ro.related_order_uuid AND ro.delete_time = ?", constant.NotDeleted).
		Joins("LEFT JOIN ttpos_member_sale_order AS mso ON so.uuid = mso.sale_order_uuid AND mso.delete_time = ? AND mso.status = ?", constant.NotDeleted, 7).
		Where("sb.delete_time = ?", constant.NotDeleted).
		Where("sb.status = ?", constant.SaleBillStatusComplete).
		Where(fmt.Sprintf("%s >= ?", timeField), req.StartTime).
		Where(fmt.Sprintf("%s <= ?", timeField), req.EndTime)

	// 应用过滤选项
	for _, opt := range opts {
		baseQuery = opt(baseQuery)
	}

	if req.IsDesk || req.IsInstant || req.IsTakeout {
		var billTypeList []uint
		if req.IsDesk {
			billTypeList = append(billTypeList, constant.SaleBillTypeDesk)
		}
		if req.IsInstant {
			billTypeList = append(billTypeList, constant.SaleBillTypeInstant)
		}
		if req.IsTakeout {
			billTypeList = append(billTypeList, constant.SaleBillTypeTakeout)
		}
		baseQuery.Where("sb.bill_type IN (?)", billTypeList)
	}

	// 对于外送订单，需要确保 mso.status = 7（mso 必须存在且 status = 7）
	if req.IsTakeout {
		baseQuery.Where("(sb.bill_type != ? OR (sb.bill_type = ? AND mso.status IS NOT NULL AND mso.status = ?))", constant.SaleBillTypeTakeout, constant.SaleBillTypeTakeout, 7)
	}

	// 构建店内订单查询SQL
	// 使用子查询先对 ro 进行聚合，避免因为 ro 有多条记录导致 so 重复计算
	storeOrderQuery := fmt.Sprintf(`
		SELECT 
			FLOOR(%s / %d) * %d AS period_start_time,
			SUM(so_agg.origin_amount) + MAX(IF(sb.bill_type = 2, IFNULL(mso.delivery_fee_amount, 0), 0)) AS order_amount,
			SUM(so_agg.payment_amount) AS pay_amount,
			SUM(IFNULL(ro_summary.refund_amount, 0)) AS refund_amount,
			sb.uuid AS sale_bill_uuid,
			MAX(sb.meal_num) AS meal_num
		FROM ttpos_sale_bill AS sb
		LEFT JOIN (
			SELECT 
				so.sale_bill_uuid,
				so.uuid AS so_uuid,
				so.origin_amount,
				so.payment_amount
			FROM ttpos_sale_order AS so
			WHERE so.delete_time = 0 AND so.status = 1
		) AS so_agg ON sb.uuid = so_agg.sale_bill_uuid
		LEFT JOIN (
			SELECT 
				related_order_uuid,
				SUM(refund_amount) AS refund_amount
			FROM ttpos_return_order
			WHERE delete_time = 0
			GROUP BY related_order_uuid
		) AS ro_summary ON so_agg.so_uuid = ro_summary.related_order_uuid
		LEFT JOIN ttpos_member_sale_order AS mso ON so_agg.so_uuid = mso.sale_order_uuid AND mso.delete_time = 0 AND mso.status = 7
		WHERE sb.delete_time = 0
			AND sb.status = ?
			AND %s >= ?
			AND %s <= ?
	`, timeField, req.PeriodSeconds, req.PeriodSeconds, timeField, timeField)

	// 添加额外的查询条件
	storeOrderArgs := []any{constant.SaleBillStatusComplete, req.StartTime, req.EndTime}

	// 应用过滤选项（数据管理订单过滤）
	if len(opts) > 0 {
		// 直接添加过滤子查询条件
		storeOrderQuery += " AND sb.uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)"
	}

	if req.IsDesk || req.IsInstant || req.IsTakeout {
		var billTypeList []uint
		if req.IsDesk {
			billTypeList = append(billTypeList, constant.SaleBillTypeDesk)
		}
		if req.IsInstant {
			billTypeList = append(billTypeList, constant.SaleBillTypeInstant)
		}
		if req.IsTakeout {
			billTypeList = append(billTypeList, constant.SaleBillTypeTakeout)
		}
		storeOrderQuery += " AND sb.bill_type IN (?)"
		storeOrderArgs = append(storeOrderArgs, billTypeList)
	}

	// 对于外送订单，需要确保 mso.status = 7（mso 必须存在且 status = 7）
	if req.IsTakeout {
		storeOrderQuery += " AND (sb.bill_type != ? OR (sb.bill_type = ? AND mso.status IS NOT NULL AND mso.status = ?))"
		storeOrderArgs = append(storeOrderArgs, constant.SaleBillTypeTakeout, constant.SaleBillTypeTakeout, 7)
	}

	storeOrderQuery += `
		GROUP BY period_start_time, sb.uuid
	`

	// 构建外卖订单查询SQL（如果需要）
	var takeoutOrderQuery string
	var takeoutOrderArgs []any
	if req.IsDelivery {
		// 使用 statistics_takeout.go 中定义的变量，保持代码一致性
		validStatesStr := buildStateInCondition(validOrderStates)
		businessStatesStr := buildStateInCondition(businessOrderStates)

		takeoutOrderQuery = fmt.Sprintf(`
		SELECT 
			FLOOR(accepted_time / %d) * %d AS period_start_time,
			IF(order_state IN %s, eater_payment, 0) AS order_amount,
			IF(order_state = %d, 0, IF(order_state IN %s, eater_payment, 0)) AS pay_amount,
			0 AS refund_amount,
			uuid AS sale_bill_uuid,
			0 AS meal_num
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state IN %s
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		`, req.PeriodSeconds, req.PeriodSeconds, validStatesStr, canceledOrderState, businessStatesStr, validStatesStr)
		takeoutOrderArgs = []any{constant.NotDeleted, req.StartTime, req.EndTime}
	}

	// 构建合并查询
	var mainQuery string
	var args []any

	if req.IsDelivery && takeoutOrderQuery != "" {
		// 合并店内订单和外卖订单
		mainQuery = fmt.Sprintf(`
			SELECT 
				period_start_time,
				SUM(order_amount) AS order_amount,
				SUM(pay_amount) AS pay_amount,
				SUM(refund_amount) AS refund_amount,
				COUNT(DISTINCT sale_bill_uuid) AS order_num,
				SUM(meal_num) AS meal_num
			FROM (
				(%s)
				UNION ALL
				(%s)
			) AS combined_orders
			GROUP BY period_start_time
			ORDER BY period_start_time ASC
			LIMIT ? OFFSET ?
		`, storeOrderQuery, takeoutOrderQuery)
		args = append(storeOrderArgs, takeoutOrderArgs...)
		args = append(args, req.PageSize, (req.PageNo-1)*req.PageSize)
	} else {
		// 仅店内订单
		mainQuery = fmt.Sprintf(`
			SELECT 
				period_start_time,
				SUM(order_amount) AS order_amount,
				SUM(pay_amount) AS pay_amount,
				SUM(refund_amount) AS refund_amount,
				COUNT(DISTINCT sale_bill_uuid) AS order_num,
				SUM(meal_num) AS meal_num
			FROM (
				%s
			) AS subquery
			GROUP BY period_start_time
			ORDER BY period_start_time ASC
			LIMIT ? OFFSET ?
		`, storeOrderQuery)
		args = append(storeOrderArgs, req.PageSize, (req.PageNo-1)*req.PageSize)
	}

	// 执行查询
	var result []model.StatisticsBusinessTimePeriodData
	r.db.Raw(mainQuery, args...).Scan(&result)

	// 计算总时段数（需要考虑合并后的时段）
	var total int64
	if req.IsDelivery {
		// 计算合并后的总时段数
		// 注意：timeField 包含表别名（sb.finish_time 或 sb.create_time），需要提取字段名
		timeFieldName := "finish_time"
		if req.IsCreateTime {
			timeFieldName = "create_time"
		}
		countQuery := fmt.Sprintf(`
			SELECT COUNT(DISTINCT period_start_time) FROM (
				SELECT FLOOR(%s / %d) * %d AS period_start_time FROM ttpos_sale_bill WHERE delete_time = ? AND status = ? AND %s >= ? AND %s <= ?
		`, timeFieldName, req.PeriodSeconds, req.PeriodSeconds, timeFieldName, timeFieldName)
		countArgs := []any{constant.NotDeleted, constant.SaleBillStatusComplete, req.StartTime, req.EndTime}

		if len(opts) > 0 {
			countQuery += " AND uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)"
		}

		if req.IsDesk || req.IsInstant || req.IsTakeout {
			var billTypeList []uint
			if req.IsDesk {
				billTypeList = append(billTypeList, constant.SaleBillTypeDesk)
			}
			if req.IsInstant {
				billTypeList = append(billTypeList, constant.SaleBillTypeInstant)
			}
			if req.IsTakeout {
				billTypeList = append(billTypeList, constant.SaleBillTypeTakeout)
			}
			countQuery += " AND bill_type IN (?)"
			countArgs = append(countArgs, billTypeList)
		}

		// 使用 statistics_takeout.go 中定义的变量，保持代码一致性
		takeoutValidStatesStr := buildStateInCondition(validOrderStates)
		countQuery += fmt.Sprintf(`
			UNION
			SELECT FLOOR(accepted_time / ?) * ? AS period_start_time FROM ttpos_takeout_order WHERE delete_time = ? AND order_state IN %s AND accepted_time > 0 AND accepted_time >= ? AND accepted_time <= ?
			) AS all_periods
		`, takeoutValidStatesStr)
		countArgs = append(countArgs, req.PeriodSeconds, req.PeriodSeconds, constant.NotDeleted, req.StartTime, req.EndTime)
		r.db.Raw(countQuery, countArgs...).Scan(&total)
	} else {
		// 仅店内订单的总时段数
		countQuery := baseQuery.
			Select(fmt.Sprintf("COUNT(DISTINCT FLOOR(%s / %d))", timeField, req.PeriodSeconds))
		countQuery.Count(&total)
	}

	return total, result
}

// CountBusinessSummaryReq 统计综合运营请求
type CountBusinessSummaryReq struct {
	StartTime         int64  // 查询开始时间戳
	EndTime           int64  // 查询结束时间戳
	Cycle             int    // 周期: 0=按日、1=按月
	PageNo, PageSize  int    // 页码, 每页大小
	ExcludeDataManage bool   // 是否排除数据管理订单
	Timezone          string // 业务时区，如 "Asia/Shanghai"
}

// businessSummaryRawData 综合运营统计原始数据
type businessSummaryRawData struct {
	FinishTime         int64   // 完成时间戳
	SaleBillUuid       uint64  // 销售账单UUID
	OrderAmount        float64 // 订单金额
	PayAmount          float64 // 支付金额
	RefundAmount       float64 // 退款金额
	MealNum            uint    // 用餐人数
	DeskUuid           uint64  // 桌台UUID
	DeskOrderAmount    float64 // 桌台订单金额
	InstantOrderAmount float64 // 点餐订单金额
	TakeoutOrderAmount float64 // 外送订单金额
}

// CountBusinessSummary 统计综合运营
func (r *StatisticsRepo) CountBusinessSummary(req CountBusinessSummaryReq) (int64, []model.StatisticsBusinessSummaryData) {
	// 1. 查询原始数据（不分组）
	var rawData []businessSummaryRawData
	rawQuery := `
		SELECT 
			sb.finish_time,
			sb.uuid AS sale_bill_uuid,
			SUM(so_agg.origin_amount) + MAX(IF(sb.bill_type = 2, IFNULL(mso.delivery_fee_amount, 0), 0)) AS order_amount,
			SUM(so_agg.payment_amount) AS pay_amount,
			SUM(IFNULL(ro_summary.refund_amount, 0)) AS refund_amount,
			MAX(sb.meal_num) AS meal_num,
			MAX(sb.desk_uuid) AS desk_uuid,
			SUM(IF(sb.bill_type = 0, so_agg.origin_amount, 0)) as desk_order_amount,
			SUM(IF(sb.bill_type = 1, so_agg.origin_amount, 0)) as instant_order_amount,
			SUM(IF(sb.bill_type = 2, so_agg.origin_amount, 0)) + MAX(IF(sb.bill_type = 2, IFNULL(mso.delivery_fee_amount, 0), 0)) as takeout_order_amount
		FROM ttpos_sale_bill AS sb
		LEFT JOIN (
			SELECT 
				so.sale_bill_uuid,
				so.uuid AS so_uuid,
				so.origin_amount,
				so.payment_amount
			FROM ttpos_sale_order AS so
			WHERE so.delete_time = ? AND so.status = ?
		) AS so_agg ON sb.uuid = so_agg.sale_bill_uuid
		LEFT JOIN (
			SELECT 
				related_order_uuid,
				SUM(refund_amount) AS refund_amount
			FROM ttpos_return_order
			WHERE delete_time = ?
			GROUP BY related_order_uuid
		) AS ro_summary ON so_agg.so_uuid = ro_summary.related_order_uuid
		LEFT JOIN ttpos_member_sale_order AS mso ON so_agg.so_uuid = mso.sale_order_uuid AND mso.delete_time = ? AND mso.status = ?
		WHERE sb.delete_time = ?
			AND sb.status = ?
			AND sb.finish_time >= ?
			AND sb.finish_time <= ?
			AND (sb.bill_type != ? OR (sb.bill_type = ? AND so_agg.so_uuid IS NOT NULL AND mso.uuid IS NOT NULL))
	`

	// 如果排除数据管理订单，添加过滤条件
	if req.ExcludeDataManage {
		rawQuery += ` AND sb.uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)`
	}

	rawQuery += `
		GROUP BY sb.uuid, sb.finish_time
	`

	r.db.Raw(rawQuery,
		constant.NotDeleted, constant.SaleOrderStatusFinish,
		constant.NotDeleted,
		constant.NotDeleted, constant.MemberSaleOrderStatusCompleted,
		constant.NotDeleted, constant.SaleBillStatusComplete,
		req.StartTime, req.EndTime,
		constant.SaleBillTypeTakeout, constant.SaleBillTypeTakeout,
	).Scan(&rawData)

	// 1.1 查询外卖订单原始数据
	takeoutRepo := NewStatisticsTakeoutRepo(r.db)
	takeoutRawData := takeoutRepo.CountTakeoutBusinessSummary(CountTakeoutBusinessSummaryReq{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})

	// 将外卖订单数据转换为与店内订单相同的结构
	for _, takeoutItem := range takeoutRawData {
		rawData = append(rawData, businessSummaryRawData{
			FinishTime:         takeoutItem.AcceptedTime, // 使用 accepted_time
			SaleBillUuid:       takeoutItem.OrderUuid,
			OrderAmount:        takeoutItem.OrderAmount,
			PayAmount:          takeoutItem.PayAmount,
			RefundAmount:       0, // 外卖订单退款已在实付金额中体现
			MealNum:            0, // 外卖订单无用餐人数
			DeskUuid:           0, // 外卖订单无桌台
			DeskOrderAmount:    0, // 外卖订单不归入桌台订单
			InstantOrderAmount: 0, // 外卖订单不归入点餐订单
			TakeoutOrderAmount: 0, // 外卖订单不归入外送订单（单独统计）
		})
	}

	// 2. 在应用层按业务时区分组、统计
	timeUtil := utils.SetTimezone(req.Timezone)

	// 按日期分组统计
	// 使用 decimal 进行精确计算
	type groupData struct {
		OrderAmount        decimal.Decimal
		PayAmount          decimal.Decimal
		RefundAmount       decimal.Decimal
		OrderNum           int64
		MealNum            int64
		DeskNum            int64
		DeskOrderAmount    decimal.Decimal
		InstantOrderAmount decimal.Decimal
		TakeoutOrderAmount decimal.Decimal
	}
	groupedData := make(map[string]*groupData)
	for _, item := range rawData {
		// 将时间戳转换为业务时区的日期
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01-02")
		}

		// 初始化分组数据
		if groupedData[dateKey] == nil {
			groupedData[dateKey] = &groupData{}
		}

		// 使用 decimal 累加统计数据
		group := groupedData[dateKey]
		group.OrderAmount = group.OrderAmount.Add(decimal.NewFromFloat(item.OrderAmount))
		group.PayAmount = group.PayAmount.Add(decimal.NewFromFloat(item.PayAmount))
		group.RefundAmount = group.RefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
		group.OrderNum++
		group.MealNum += int64(item.MealNum)
		if item.DeskUuid > 0 {
			group.DeskNum++
		}
		group.DeskOrderAmount = group.DeskOrderAmount.Add(decimal.NewFromFloat(item.DeskOrderAmount))
		group.InstantOrderAmount = group.InstantOrderAmount.Add(decimal.NewFromFloat(item.InstantOrderAmount))
		group.TakeoutOrderAmount = group.TakeoutOrderAmount.Add(decimal.NewFromFloat(item.TakeoutOrderAmount))
	}

	// 转换为结果格式
	result := make([]model.StatisticsBusinessSummaryData, 0, len(groupedData))
	for dateKey, group := range groupedData {
		result = append(result, model.StatisticsBusinessSummaryData{
			Date:               sql.NullString{String: dateKey, Valid: true},
			OrderAmount:        sql.NullFloat64{Float64: group.OrderAmount.InexactFloat64(), Valid: true},
			PayAmount:          sql.NullFloat64{Float64: group.PayAmount.InexactFloat64(), Valid: true},
			RefundAmount:       sql.NullFloat64{Float64: group.RefundAmount.InexactFloat64(), Valid: true},
			OrderNum:           sql.NullInt64{Int64: group.OrderNum, Valid: true},
			MealNum:            sql.NullInt64{Int64: group.MealNum, Valid: true},
			DeskNum:            sql.NullInt64{Int64: group.DeskNum, Valid: true},
			DeskOrderAmount:    sql.NullFloat64{Float64: group.DeskOrderAmount.InexactFloat64(), Valid: true},
			InstantOrderAmount: sql.NullFloat64{Float64: group.InstantOrderAmount.InexactFloat64(), Valid: true},
			TakeoutOrderAmount: sql.NullFloat64{Float64: group.TakeoutOrderAmount.InexactFloat64(), Valid: true},
		})
	}

	// 3. 按日期排序
	slices.SortFunc(result, func(a, b model.StatisticsBusinessSummaryData) int {
		if a.Date.String < b.Date.String {
			return -1
		} else if a.Date.String > b.Date.String {
			return 1
		}
		return 0
	})

	// 4. 分页
	total := int64(len(result))
	start := (req.PageNo - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(result) {
		start = len(result)
	}
	if end > len(result) {
		end = len(result)
	}
	if start < end {
		result = result[start:end]
	} else {
		result = []model.StatisticsBusinessSummaryData{}
	}

	return total, result
}

// CountBusinessPaymentMethodReq 统计收款数据请求
type CountBusinessPaymentMethodReq struct {
	StartTime                    int64    // 查询开始时间戳
	EndTime                      int64    // 查询结束时间戳
	Cycle                        int      // 周期: 0=按日、1=按月
	PageNo, PageSize             int      // 页码, 每页大小
	IsDesk, IsInstant, IsTakeout bool     // 是否是桌台订单, 是否是点餐订单, 是否是外送订单
	IsDelivery                   bool     // 是否是外卖订单（Grab/LINE MAN等）
	PaymentMethodList            []uint64 // 支付方式UUID列表: 空=全部（优先使用）
	PaymentMethodNames           []string // 支付方式名称列表: 空=全部（PaymentMethodList为空时使用）
	ExcludeDataManage            bool     // 是否排除数据管理订单
	Timezone                     string   // 业务时区，如 "Asia/Shanghai"
}

// businessPaymentMethodRawData 支付方式统计原始数据
type businessPaymentMethodRawData struct {
	CreateTime              int64   // 创建时间戳
	PaymentMethodUuid       uint64  // 支付方式UUID
	PaymentMethodSort       int     // 支付方式排序
	PaymentMethodCreateTime int64   // 支付方式创建时间戳
	PaymentName             string  // 支付方式名称
	PaymentAmount           float64 // 支付金额（已扣除退款）
}

// CountBusinessPaymentMethod 统计收款数据
func (r *StatisticsRepo) CountBusinessPaymentMethod(req CountBusinessPaymentMethodReq) (int64, []model.StatisticsBusinessPaymentMethodData) {
	// 1. 查询原始数据（不分组）
	baseQuery := `
		SELECT 
			sb.finish_time AS create_time,
			po.payment_method_uuid,
			pm.payment_name,
			pm.sort AS payment_method_sort,
			pm.create_time AS payment_method_create_time,
			po.amount - IFNULL(roa.refund_amount, 0) AS payment_amount
		FROM ttpos_payment_order AS po
		LEFT JOIN ttpos_payment_method AS pm ON po.payment_method_uuid = pm.uuid
		LEFT JOIN ttpos_sale_order AS so ON po.related_uuid = so.uuid AND so.delete_time = 0
		LEFT JOIN ttpos_sale_bill AS sb ON so.sale_bill_uuid = sb.uuid AND sb.delete_time = 0
		LEFT JOIN ttpos_member_sale_order AS mso ON so.uuid = mso.sale_order_uuid AND mso.delete_time = 0
		LEFT JOIN (
			SELECT
				payment_order_uuid,
				SUM(amount) AS refund_amount
			FROM ttpos_return_order_amount
			WHERE delete_time = 0
				AND refund_status = 1
			GROUP BY payment_order_uuid
		) AS roa ON roa.payment_order_uuid = po.uuid
		WHERE po.delete_time = 0
			AND po.related_type = 0
			AND po.status = 1
			AND sb.finish_time >= ?
			AND sb.finish_time <= ?
			AND sb.status = ?
			AND so.status = ?
			AND (sb.bill_type != ? OR (sb.bill_type = ? AND mso.uuid IS NOT NULL AND mso.status = ?))
	`

	args := []any{
		req.StartTime,
		req.EndTime,
		constant.SaleBillStatusComplete,
		constant.SaleOrderStatusFinish,
		constant.SaleBillTypeTakeout,
		constant.SaleBillTypeTakeout,
		constant.MemberSaleOrderStatusCompleted,
	}

	// 如果排除数据管理订单，添加过滤条件
	if req.ExcludeDataManage {
		baseQuery += ` AND sb.uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)`
	}

	// 订单类型筛选
	if req.IsDesk || req.IsInstant || req.IsTakeout {
		billTypes := []uint{}
		if req.IsDesk {
			billTypes = append(billTypes, constant.SaleBillTypeDesk)
		}
		if req.IsInstant {
			billTypes = append(billTypes, constant.SaleBillTypeInstant)
		}
		if req.IsTakeout {
			billTypes = append(billTypes, constant.SaleBillTypeTakeout)
		}
		baseQuery += " AND sb.bill_type IN (?)"
		args = append(args, billTypes)
	}

	// 支付方式筛选：优先使用 PaymentMethodList（UUID），如果没有则使用 PaymentMethodNames（名称）
	if len(req.PaymentMethodList) > 0 {
		// 使用支付方式UUID筛选
		baseQuery += " AND po.payment_method_uuid IN (?)"
		args = append(args, req.PaymentMethodList)
	} else if len(req.PaymentMethodNames) > 0 {
		// 使用支付方式名称筛选（因为不同商家的同一支付方式UUID可能不同）
		baseQuery += " AND pm.payment_name IN (?)"
		args = append(args, req.PaymentMethodNames)
	}

	var rawData []businessPaymentMethodRawData
	r.db.Raw(baseQuery, args...).Scan(&rawData)

	// 1.1 如果需要查询外卖订单数据，合并外卖订单支付方式数据
	needQueryTakeout := false
	var grabPaymentMethodUuid, linemanPaymentMethodUuid uint64

	// 检查是否需要查询外卖订单数据
	if req.IsDelivery {
		// 如果 OrderDelivery=1，需要查询外卖订单
		needQueryTakeout = true
	} else if len(req.PaymentMethodList) > 0 {
		// 如果 paymentMethodList != ""，检查是否包含 Grab/LINE MAN 的支付方式UUID
		// 查询 Grab 和 LINE MAN 的支付方式UUID
		var grabPaymentMethod, linemanPaymentMethod model.PaymentMethod
		r.db.Model(&model.PaymentMethod{}).
			Where("payment_name = ? AND delete_time = ? AND source = ?", "Grab", constant.NotDeleted, constant.PaymentMethodSourceSystem).
			First(&grabPaymentMethod)
		r.db.Model(&model.PaymentMethod{}).
			Where("payment_name = ? AND delete_time = ? AND source = ?", "LINE MAN", constant.NotDeleted, constant.PaymentMethodSourceSystem).
			First(&linemanPaymentMethod)

		grabPaymentMethodUuid = grabPaymentMethod.Uuid
		linemanPaymentMethodUuid = linemanPaymentMethod.Uuid

		// 检查 PaymentMethodList 中是否包含 Grab 或 LINE MAN 的UUID
		for _, uuid := range req.PaymentMethodList {
			if uuid == grabPaymentMethodUuid || uuid == linemanPaymentMethodUuid {
				needQueryTakeout = true
				break
			}
		}
	} else if len(req.PaymentMethodNames) > 0 {
		// 如果使用 PaymentMethodNames，检查是否包含 Grab 或 LINE MAN
		for _, name := range req.PaymentMethodNames {
			if name == "Grab" || name == "LINE MAN" {
				needQueryTakeout = true
				break
			}
		}
	}

	// 如果需要查询外卖订单数据，查询并合并
	if needQueryTakeout {
		takeoutRepo := NewStatisticsTakeoutRepo(r.db)
		takeoutRawData := takeoutRepo.CountTakeoutPaymentMethodRawData(CountTakeoutReq{
			TimeStart: req.StartTime,
			TimeEnd:   req.EndTime,
		})

		// 将外卖订单数据转换为 businessPaymentMethodRawData 格式
		for _, takeoutItem := range takeoutRawData {
			// 如果 PaymentMethodUuid 为 0，说明没有匹配到支付方式，跳过
			if takeoutItem.PaymentMethodUuid == 0 {
				continue
			}

			// 如果指定了 PaymentMethodList，需要过滤（当 IsDelivery=false 时）
			if len(req.PaymentMethodList) > 0 && !req.IsDelivery {
				// 检查是否在 PaymentMethodList 中
				found := false
				for _, uuid := range req.PaymentMethodList {
					if uuid == takeoutItem.PaymentMethodUuid {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			} else if len(req.PaymentMethodNames) > 0 && !req.IsDelivery {
				// 如果使用 PaymentMethodNames，检查是否匹配（当 IsDelivery=false 时）
				found := false
				for _, name := range req.PaymentMethodNames {
					if name == takeoutItem.PaymentName {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// 转换为 businessPaymentMethodRawData 格式
			rawData = append(rawData, businessPaymentMethodRawData{
				CreateTime:              takeoutItem.AcceptedTime,
				PaymentMethodUuid:       takeoutItem.PaymentMethodUuid,
				PaymentMethodSort:       takeoutItem.PaymentMethodSort,
				PaymentMethodCreateTime: takeoutItem.PaymentMethodCreateTime,
				PaymentName:             takeoutItem.PaymentName,
				PaymentAmount:           takeoutItem.PaymentAmount,
			})
		}
	}

	// 2. 在应用层按业务时区分组、统计
	timeUtil := utils.SetTimezone(req.Timezone)

	// 按日期和支付方式分组统计
	// 使用 decimal 进行精确计算
	type groupKey struct {
		Date              string
		PaymentMethodUuid uint64
	}
	type paymentGroupData struct {
		PaymentName             string
		PaymentMethodSort       int
		PaymentMethodCreateTime int64
		PaymentNum              int64
		PaymentAmount           decimal.Decimal
	}
	groupedData := make(map[groupKey]*paymentGroupData)

	for _, item := range rawData {
		// 将时间戳转换为业务时区的日期
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.CreateTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.CreateTime, "2006-01-02")
		}

		key := groupKey{
			Date:              dateKey,
			PaymentMethodUuid: item.PaymentMethodUuid,
		}

		// 初始化分组数据
		if groupedData[key] == nil {
			groupedData[key] = &paymentGroupData{
				PaymentName:             item.PaymentName,
				PaymentMethodSort:       item.PaymentMethodSort,
				PaymentMethodCreateTime: item.PaymentMethodCreateTime,
			}
		}

		// 使用 decimal 累加统计数据
		group := groupedData[key]
		group.PaymentNum++
		group.PaymentAmount = group.PaymentAmount.Add(decimal.NewFromFloat(item.PaymentAmount))
	}

	// 转换为结果格式
	paymentResult := make([]model.StatisticsBusinessPaymentMethodData, 0, len(groupedData))
	for key, group := range groupedData {
		paymentResult = append(paymentResult, model.StatisticsBusinessPaymentMethodData{
			Date:                    sql.NullString{String: key.Date, Valid: true},
			PaymentMethodSort:       sql.NullInt64{Int64: int64(group.PaymentMethodSort), Valid: true},
			PaymentMethodCreateTime: sql.NullInt64{Int64: group.PaymentMethodCreateTime, Valid: true},
			PaymentName:             sql.NullString{String: group.PaymentName, Valid: true},
			PaymentNum:              sql.NullInt64{Int64: group.PaymentNum, Valid: true},
			PaymentAmount:           sql.NullFloat64{Float64: group.PaymentAmount.InexactFloat64(), Valid: true},
		})
	}

	// 3. 按日期和支付方式排序
	slices.SortFunc(paymentResult, func(a, b model.StatisticsBusinessPaymentMethodData) int {
		if a.Date.String < b.Date.String {
			return -1
		} else if a.Date.String > b.Date.String {
			return 1
		}
		// 日期相同，按支付方式Sort排序
		if a.PaymentMethodSort.Int64 < b.PaymentMethodSort.Int64 {
			return -1
		} else if a.PaymentMethodSort.Int64 > b.PaymentMethodSort.Int64 {
			return 1
		} else {
			if a.PaymentMethodCreateTime.Int64 < b.PaymentMethodCreateTime.Int64 {
				return 1
			} else if a.PaymentMethodCreateTime.Int64 > b.PaymentMethodCreateTime.Int64 {
				return -1
			}
			return 0
		}
	})

	// 4. 分页
	total := int64(len(paymentResult))
	start := (req.PageNo - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(paymentResult) {
		start = len(paymentResult)
	}
	if end > len(paymentResult) {
		end = len(paymentResult)
	}
	if start < end {
		paymentResult = paymentResult[start:end]
	} else {
		paymentResult = []model.StatisticsBusinessPaymentMethodData{}
	}

	return total, paymentResult
}

// containsOrderType 检查订单类型列表中是否包含指定类型
func containsOrderType(orderTypes []uint, orderType uint) bool {
	return slices.Contains(orderTypes, orderType)
}

// CountChannelSale 统计渠道营业数据
func (r *StatisticsRepo) CountChannelSale(startTime, endTime int64, excludeDataManage bool) (map[string]*model.ChannelSaleRepoResult, error) {
	result := make(map[string]*model.ChannelSaleRepoResult)
	db := r.db

	// 复用 CountSale 的子查询逻辑，并关联 sale_bill 表获取 dining_method 字段
	prefix := config.Database.TablePrefix
	saleBillTable := prefix + "sale_bill"
	statisticsSaleTable := prefix + "statistics_sale"

	// 构建子查询，添加 dining_method 字段，并为所有字段添加表别名以避免歧义
	subQuerySelect := []string{
		statisticsSaleTable + ".sale_bill_uuid",
		statisticsSaleTable + ".desk_uuid",
		statisticsSaleTable + ".order_source_uuid",
		statisticsSaleTable + ".is_meger",
		statisticsSaleTable + ".is_special",
		statisticsSaleTable + ".is_takeout",
		"SUM(" + statisticsSaleTable + ".product_price + " + statisticsSaleTable + ".product_tax + " + statisticsSaleTable + ".service_fee + " + statisticsSaleTable + ".service_tax + " + statisticsSaleTable + ".payment_fee + " + statisticsSaleTable + ".extend_price) AS sale_amount",
		"SUM(" + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".payment_balance) AS received_amount",
		"SUM(" + statisticsSaleTable + ".product_price) AS product_price",
		"SUM(" + statisticsSaleTable + ".product_origin_price) AS product_origin_price",
		"SUM(" + statisticsSaleTable + ".product_num) AS product_num",
		"SUM(" + statisticsSaleTable + ".discount_member) AS discount_member",
		"SUM(" + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance - " + statisticsSaleTable + ".product_tax - " + statisticsSaleTable + ".service_tax + " + statisticsSaleTable + ".refund_tax) AS business_amount",
		"SUM(" + statisticsSaleTable + ".payment_fee - " + statisticsSaleTable + ".refund_fee) AS payment_fee",
		"SUM(" + statisticsSaleTable + ".service_fee - " + statisticsSaleTable + ".refund_service_fee) AS service_fee",
		"SUM(" + statisticsSaleTable + ".product_tax + " + statisticsSaleTable + ".service_tax - " + statisticsSaleTable + ".refund_tax) AS tax",
		"SUM(" + statisticsSaleTable + ".refund_amount + " + statisticsSaleTable + ".refund_payment_balance) AS refund_amount",
		"SUM(" + statisticsSaleTable + ".discount - " + statisticsSaleTable + ".refund_discount) AS discount",
		"SUM(" + statisticsSaleTable + ".gift_amount) AS gift_amount",
		"SUM(" + statisticsSaleTable + ".gift_num) AS gift_num",
		"SUM(" + statisticsSaleTable + ".free_amount) AS free_amount",
		"SUM(" + statisticsSaleTable + ".free_num) AS free_num",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1, " + statisticsSaleTable + ".payment_amount, 0)) AS takeout_sale_amount",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".delivery_fee, 0)) AS takeout_business_amount",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1, " + statisticsSaleTable + ".refund_amount, 0)) AS takeout_refund_amount",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1, " + statisticsSaleTable + ".delivery_fee, 0)) AS takeout_delivery_fee",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid > 0, " + statisticsSaleTable + ".meal_num, 0)) AS meal_num",
		"SUM(IF(" + statisticsSaleTable + ".is_meger = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS order_amount",
		"SUM(" + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance) AS avg_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid > 0 AND " + statisticsSaleTable + ".is_takeout = 0 AND " + statisticsSaleTable + ".is_meger = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS desk_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid > 0 AND " + statisticsSaleTable + ".is_takeout = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS avg_desk_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid = 0 AND " + statisticsSaleTable + ".order_source_uuid = 0 AND " + statisticsSaleTable + ".is_meger = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS instant_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid = 0 AND " + statisticsSaleTable + ".order_source_uuid = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS avg_instant_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid = 0 AND " + statisticsSaleTable + ".order_source_uuid > 0 AND " + statisticsSaleTable + ".is_meger = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS instant_order_takeaway_amount",
		"SUM(IF(" + statisticsSaleTable + ".desk_uuid = 0 AND " + statisticsSaleTable + ".order_source_uuid > 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS avg_instant_order_takeaway_amount",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1 AND " + statisticsSaleTable + ".is_meger = 0, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS takeout_order_amount",
		"SUM(IF(" + statisticsSaleTable + ".is_takeout = 1, " + statisticsSaleTable + ".payment_amount - " + statisticsSaleTable + ".refund_amount - " + statisticsSaleTable + ".refund_payment_balance, 0)) AS avg_takeout_order_amount",
		statisticsSaleTable + ".complete_time",
		"sb.dining_method",
	}
	subQuery := db.Table(statisticsSaleTable).
		Select(subQuerySelect).
		Joins("LEFT JOIN "+saleBillTable+" AS sb ON "+statisticsSaleTable+".sale_bill_uuid = sb.uuid AND sb.delete_time = ?", constant.NotDeleted).
		Where(statisticsSaleTable+".complete_time >= ? AND "+statisticsSaleTable+".complete_time <= ?", startTime, endTime)

	// 应用数据管理过滤条件
	if excludeDataManage {
		subQuery = subQuery.Where(statisticsSaleTable + ".sale_bill_uuid NOT IN (SELECT data_uuid FROM " + prefix + "data_manage WHERE type = 0 AND delete_time = 0)")
	}

	subQuery = subQuery.Group(statisticsSaleTable + ".sale_bill_uuid")

	// 定义渠道分组条件
	channelSelects := map[string][]string{
		"summary": { // 合计：所有订单
			"SUM(IF(t.is_meger = 0, 1, 0)) AS total_order_num",
			"MIN(CASE WHEN t.order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.order_amount > 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_order_amount) / SUM(IF(t.is_meger = 0, 1, 0)), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
		"table": { // 桌台：desk_uuid > 0 && is_takeout = 0
			"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.desk_order_amount > 0 AND t.is_meger = 0 THEN t.desk_order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_desk_order_amount) / COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END), 2) AS avg_order_amount",
			"COUNT(CASE WHEN t.desk_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_desk_num",
			"SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0)) AS total_meal_num",
			"ROUND(SUM(t.desk_order_amount) / NULLIF(SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0)), 0), 2) AS order_amount_meal_avg", // 人均订单金额：桌台订单总金额 / 用餐人数，保留两位小数
		},
		"dine_in": { // 点餐-店内：desk_uuid = 0 && order_source_uuid = 0 && is_takeout = 0
			"COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.instant_order_amount > 0 THEN t.instant_order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_instant_order_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_meger = 0 AND t.instant_order_amount > 0 THEN 1 END), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
		"takeout_shop": { // 点餐-外卖：desk_uuid = 0 && order_source_uuid > 0 && is_takeout = 0
			"COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.instant_order_takeaway_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_takeaway_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.instant_order_takeaway_amount > 0 THEN t.instant_order_takeaway_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_instant_order_takeaway_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_meger = 0 AND t.instant_order_takeaway_amount > 0 THEN 1 END), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
		"takeout_delivery": { // 外送：is_takeout = 1
			"COUNT(CASE WHEN t.desk_uuid = 0 AND t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.takeout_order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.is_takeout = 1 AND t.takeout_order_amount > 0 THEN t.takeout_order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_takeout_order_amount) / COUNT(CASE WHEN t.is_takeout = 1 AND t.is_meger = 0 THEN 1 END), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
		"dine_in_store": { // 堂食：桌台订单 + 点餐订单（非打包）
			"COUNT(CASE WHEN ((t.desk_uuid > 0 AND t.is_takeout = 0) OR (t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND (t.dining_method = 0 OR t.dining_method IS NULL))) AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN ((t.desk_uuid > 0 AND t.is_takeout = 0) OR (t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND (t.dining_method = 0 OR t.dining_method IS NULL))) AND t.order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN ((t.desk_uuid > 0 AND t.is_takeout = 0) OR (t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND (t.dining_method = 0 OR t.dining_method IS NULL))) AND t.order_amount > 0 AND t.is_meger = 0 THEN t.order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(CASE WHEN ((t.desk_uuid > 0 AND t.is_takeout = 0) OR (t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND (t.dining_method = 0 OR t.dining_method IS NULL))) AND t.is_meger = 0 THEN t.avg_order_amount ELSE 0 END) / COUNT(CASE WHEN ((t.desk_uuid > 0 AND t.is_takeout = 0) OR (t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND (t.dining_method = 0 OR t.dining_method IS NULL))) AND t.is_meger = 0 THEN 1 END), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
		"takeaway": { // 外带：点餐订单（打包）
			"COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.dining_method = 1 AND t.is_meger = 0 THEN 1 END) AS total_order_num",
			"MIN(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.dining_method = 1 AND t.instant_order_amount >= 0 AND t.is_special = 0 AND t.is_meger = 0 THEN t.instant_order_amount ELSE NULL END) AS min_order_amount",
			"MAX(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.dining_method = 1 AND t.instant_order_amount > 0 THEN t.instant_order_amount ELSE NULL END) AS max_order_amount",
			"ROUND(SUM(t.avg_instant_order_amount) / COUNT(CASE WHEN t.desk_uuid = 0 AND t.order_source_uuid = 0 AND t.is_takeout = 0 AND t.dining_method = 1 AND t.is_meger = 0 AND t.instant_order_amount > 0 THEN 1 END), 2) AS avg_order_amount",
			"0 AS total_desk_num",
			"0 AS total_meal_num",
		},
	}

	// 为每个渠道执行查询
	for channelKey, selects := range channelSelects {
		var channelResult model.ChannelSaleRepoResult
		err := db.Table("(?) AS t", subQuery).
			Select(selects).
			Scan(&channelResult).Error
		if err != nil {
			return nil, err
		}
		result[channelKey] = &channelResult
	}

	// 查询外卖订单原始数据
	takeoutRepo := NewStatisticsTakeoutRepo(r.db)
	takeoutRawData := takeoutRepo.CountTakeoutChannelSale(CountTakeoutChannelSaleReq{
		StartTime: startTime,
		EndTime:   endTime,
	})

	// 合并外卖订单数据到 summary（合计）渠道
	if len(takeoutRawData) > 0 {
		// 确保 summary 渠道存在
		if result["summary"] == nil {
			result["summary"] = &model.ChannelSaleRepoResult{}
		}
		summary := result["summary"]

		// 收集外卖订单的实付金额（用于计算最小/最大/平均）
		var takeoutPayAmounts []float64
		var takeoutOrderNum int64
		var takeoutPayAmountSum decimal.Decimal

		for _, takeoutItem := range takeoutRawData {
			takeoutOrderNum++
			// 累加所有订单的实付金额（包括0，用于计算平均订单金额）
			takeoutPayAmountSum = takeoutPayAmountSum.Add(decimal.NewFromFloat(takeoutItem.PayAmount))
			// 只统计实付金额 > 0 的订单（用于最小/最大订单金额计算）
			if takeoutItem.PayAmount > 0 {
				takeoutPayAmounts = append(takeoutPayAmounts, takeoutItem.PayAmount)
			}
		}

		// 合并订单数
		summary.TotalOrderNum.Int64 += takeoutOrderNum

		// 查询店内订单的实付金额（用于合并计算最小/最大/平均）
		// 查询所有订单（包括实付金额为0的订单，用于计算平均订单金额）
		var storeOrderData []struct {
			OrderAmount float64
		}
		db.Table("(?) AS t", subQuery).
			Select("t.order_amount").
			Where("t.is_meger = 0").
			Scan(&storeOrderData)

		// 收集店内订单的实付金额
		var storePayAmounts []float64
		var storePayAmountSum decimal.Decimal
		for _, storeItem := range storeOrderData {
			// 累加所有订单的实付金额（包括0，用于计算平均订单金额）
			storePayAmountSum = storePayAmountSum.Add(decimal.NewFromFloat(storeItem.OrderAmount))
			// 只统计实付金额 > 0 且非特殊订单（用于最小/最大订单金额计算）
			if storeItem.OrderAmount > 0 {
				storePayAmounts = append(storePayAmounts, storeItem.OrderAmount)
			}
		}

		// 合并所有实付金额（用于最小/最大订单金额计算）
		allPayAmounts := append(storePayAmounts, takeoutPayAmounts...)
		// 总订单数：所有订单（包括实付金额为0的订单）
		totalOrderNum := int64(len(storeOrderData)) + takeoutOrderNum
		// 总实付金额：所有订单的实付金额总和（包括0）
		totalPayAmountSum := storePayAmountSum.Add(takeoutPayAmountSum)

		// 重新计算最小订单金额（只考虑实付金额 > 0 的订单）
		if len(allPayAmounts) > 0 {
			minAmount := allPayAmounts[0]
			maxAmount := allPayAmounts[0]
			for _, amount := range allPayAmounts {
				if amount > 0 && (minAmount <= 0 || amount < minAmount) {
					minAmount = amount
				}
				if amount > maxAmount {
					maxAmount = amount
				}
			}
			summary.MinOrderAmount.Float64 = minAmount
			summary.MinOrderAmount.Valid = true
			summary.MaxOrderAmount.Float64 = maxAmount
			summary.MaxOrderAmount.Valid = true
		}

		// 重新计算平均订单金额（基于所有订单，包括实付金额为0的订单）
		if totalOrderNum > 0 {
			avgAmount := totalPayAmountSum.Div(decimal.NewFromInt(totalOrderNum))
			summary.AvgOrderAmount.Float64 = avgAmount.Round(2).InexactFloat64()
			summary.AvgOrderAmount.Valid = true
		}
	}

	// 查询 Grab 和 LINE MAN 外卖订单数据
	grabRawData := takeoutRepo.CountTakeoutChannelSaleByPlatform(CountTakeoutChannelSaleReq{
		StartTime: startTime,
		EndTime:   endTime,
	}, "grab")
	linemanRawData := takeoutRepo.CountTakeoutChannelSaleByPlatform(CountTakeoutChannelSaleReq{
		StartTime: startTime,
		EndTime:   endTime,
	}, "lineman")

	// 计算 Grab 统计
	result["grab"] = calculateChannelSaleFromRawData(grabRawData)

	// 计算 LINE MAN 统计
	result["lineman"] = calculateChannelSaleFromRawData(linemanRawData)

	// 计算外卖统计（点餐标记外卖 + Grab + LINE MAN）
	// 查询点餐订单（标记外卖）的原始数据
	var takeoutShopOrderData []struct {
		OrderAmount float64
	}
	db.Table("(?) AS t", subQuery).
		Select("t.instant_order_takeaway_amount AS order_amount").
		Where("t.desk_uuid = 0 AND t.order_source_uuid > 0 AND t.is_takeout = 0 AND t.is_meger = 0 AND t.instant_order_takeaway_amount > 0").
		Scan(&takeoutShopOrderData)

	// 合并所有外卖订单数据（点餐标记外卖 + Grab + LINE MAN）
	var allTakeoutPayAmounts []float64
	var allTakeoutOrderNum int64
	var allTakeoutPayAmountSum decimal.Decimal

	// 添加点餐订单（标记外卖）
	for _, item := range takeoutShopOrderData {
		allTakeoutOrderNum++
		allTakeoutPayAmountSum = allTakeoutPayAmountSum.Add(decimal.NewFromFloat(item.OrderAmount))
		if item.OrderAmount > 0 {
			allTakeoutPayAmounts = append(allTakeoutPayAmounts, item.OrderAmount)
		}
	}

	// 添加 Grab 订单
	for _, item := range grabRawData {
		allTakeoutOrderNum++
		allTakeoutPayAmountSum = allTakeoutPayAmountSum.Add(decimal.NewFromFloat(item.PayAmount))
		if item.PayAmount > 0 {
			allTakeoutPayAmounts = append(allTakeoutPayAmounts, item.PayAmount)
		}
	}

	// 添加 LINE MAN 订单
	for _, item := range linemanRawData {
		allTakeoutOrderNum++
		allTakeoutPayAmountSum = allTakeoutPayAmountSum.Add(decimal.NewFromFloat(item.PayAmount))
		if item.PayAmount > 0 {
			allTakeoutPayAmounts = append(allTakeoutPayAmounts, item.PayAmount)
		}
	}

	// 计算外卖统计
	result["takeout"] = calculateChannelSaleFromRawDataAndAmounts(allTakeoutOrderNum, allTakeoutPayAmounts, allTakeoutPayAmountSum)

	return result, nil
}

// calculateChannelSaleFromRawData 从原始数据计算渠道统计
func calculateChannelSaleFromRawData(rawData []takeoutChannelSaleRawData) *model.ChannelSaleRepoResult {
	result := &model.ChannelSaleRepoResult{}
	if len(rawData) == 0 {
		return result
	}

	var payAmounts []float64
	var orderNum int64
	var payAmountSum decimal.Decimal

	for _, item := range rawData {
		orderNum++
		payAmountSum = payAmountSum.Add(decimal.NewFromFloat(item.PayAmount))
		if item.PayAmount > 0 {
			payAmounts = append(payAmounts, item.PayAmount)
		}
	}

	result.TotalOrderNum.Int64 = orderNum
	result.TotalOrderNum.Valid = true

	if len(payAmounts) > 0 {
		minAmount := payAmounts[0]
		maxAmount := payAmounts[0]
		for _, amount := range payAmounts {
			if amount > 0 && (minAmount <= 0 || amount < minAmount) {
				minAmount = amount
			}
			if amount > maxAmount {
				maxAmount = amount
			}
		}
		result.MinOrderAmount.Float64 = minAmount
		result.MinOrderAmount.Valid = true
		result.MaxOrderAmount.Float64 = maxAmount
		result.MaxOrderAmount.Valid = true
	}

	if orderNum > 0 {
		avgAmount := payAmountSum.Div(decimal.NewFromInt(orderNum))
		result.AvgOrderAmount.Float64 = avgAmount.Round(2).InexactFloat64()
		result.AvgOrderAmount.Valid = true
	}

	return result
}

// calculateChannelSaleFromRawDataAndAmounts 从订单数和金额数据计算渠道统计
func calculateChannelSaleFromRawDataAndAmounts(orderNum int64, payAmounts []float64, payAmountSum decimal.Decimal) *model.ChannelSaleRepoResult {
	result := &model.ChannelSaleRepoResult{}

	result.TotalOrderNum.Int64 = orderNum
	result.TotalOrderNum.Valid = true

	if len(payAmounts) > 0 {
		minAmount := payAmounts[0]
		maxAmount := payAmounts[0]
		for _, amount := range payAmounts {
			if amount > 0 && (minAmount <= 0 || amount < minAmount) {
				minAmount = amount
			}
			if amount > maxAmount {
				maxAmount = amount
			}
		}
		result.MinOrderAmount.Float64 = minAmount
		result.MinOrderAmount.Valid = true
		result.MaxOrderAmount.Float64 = maxAmount
		result.MaxOrderAmount.Valid = true
	}

	if orderNum > 0 {
		avgAmount := payAmountSum.Div(decimal.NewFromInt(orderNum))
		result.AvgOrderAmount.Float64 = avgAmount.Round(2).InexactFloat64()
		result.AvgOrderAmount.Valid = true
	}

	return result
}

// CountRefundSummaryReq 统计退款金额汇总请求
type CountRefundSummaryReq struct {
	StartTime         int64  // 查询开始时间戳
	EndTime           int64  // 查询结束时间戳
	Cycle             int    // 周期: 0=按日、1=按月
	PageNo, PageSize  int    // 页码, 每页大小
	ExcludeDataManage bool   // 是否排除数据管理订单
	Timezone          string // 业务时区，如 "Asia/Shanghai"
}

// refundSummaryRawData 退款金额汇总统计原始数据
type refundSummaryRawData struct {
	FinishTime          int64   // 完成时间戳
	SaleBillUuid        uint64  // 销售账单UUID
	RefundAmount        float64 // 退款金额
	PartialRefundAmount float64 // 部分退款金额
	FullRefundAmount    float64 // 整单退款金额
	PartialRefundNum    int64   // 部分退款笔数
	FullRefundNum       int64   // 整单退款笔数
	RefundNum           int64   // 退款笔数（每个退货单算一笔）
}

// CountRefundSummary 统计退款金额汇总
func (r *StatisticsRepo) CountRefundSummary(req CountRefundSummaryReq) (int64, []model.StatisticsRefundSummaryData) {
	// 1. 查询原始数据（不分组）
	var rawData []refundSummaryRawData
	rawQuery := `
		SELECT 
			sb.finish_time,
			sb.uuid AS sale_bill_uuid,
			SUM(ro.refund_amount) AS refund_amount,
			SUM(IF(ro.return_type = 2, ro.refund_amount, 0)) AS partial_refund_amount,
			SUM(IF(ro.return_type = 1, ro.refund_amount, 0)) AS full_refund_amount,
			COUNT(DISTINCT IF(ro.return_type = 2, ro.uuid, NULL)) AS partial_refund_num,
			COUNT(DISTINCT IF(ro.return_type = 1, ro.uuid, NULL)) AS full_refund_num,
			COUNT(DISTINCT ro.uuid) AS refund_num
		FROM ttpos_return_order AS ro
		LEFT JOIN ttpos_sale_order AS so ON ro.related_order_uuid = so.uuid AND so.delete_time = ?
		LEFT JOIN ttpos_sale_bill AS sb ON so.sale_bill_uuid = sb.uuid AND sb.delete_time = ?
		WHERE ro.delete_time = ?
			AND ro.related_order_type = ?
			AND sb.status = ?
			AND sb.finish_time >= ?
			AND sb.finish_time <= ?
	`

	// 如果排除数据管理订单，添加过滤条件
	if req.ExcludeDataManage {
		rawQuery += ` AND sb.uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)`
	}

	rawQuery += `
		GROUP BY sb.uuid, sb.finish_time
	`

	r.db.Raw(rawQuery,
		constant.NotDeleted,
		constant.NotDeleted,
		constant.NotDeleted,
		constant.ReturnOrderRelatedOrderTypeSaleOrder,
		constant.SaleBillStatusComplete,
		req.StartTime, req.EndTime,
	).Scan(&rawData)

	// 2. 查询订单总数（用于计算退款率）
	var orderCountData []struct {
		FinishTime int64
		OrderNum   int64
	}
	orderCountQuery := `
		SELECT 
			sb.finish_time,
			COUNT(DISTINCT sb.uuid) AS order_num
		FROM ttpos_sale_bill AS sb
		WHERE sb.delete_time = ?
			AND sb.status = ?
			AND sb.finish_time >= ?
			AND sb.finish_time <= ?
	`

	if req.ExcludeDataManage {
		orderCountQuery += ` AND sb.uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = 0 AND delete_time = 0)`
	}

	orderCountQuery += `
		GROUP BY FROM_UNIXTIME(sb.finish_time, IF(? = 1, '%Y-%m', '%Y-%m-%d'))
	`

	r.db.Raw(orderCountQuery,
		constant.NotDeleted,
		constant.SaleBillStatusComplete,
		req.StartTime, req.EndTime,
		req.Cycle,
	).Scan(&orderCountData)

	// 2.1. 查询外卖订单取消订单数据（order_state = 60，且 accepted_time > 0）
	type takeoutRefundRawData struct {
		AcceptedTime int64   // 接单时间戳
		RefundAmount float64 // 退款金额（eater_payment）
		RefundNum    int64   // 退款笔数（每个订单算一笔）
	}
	var takeoutRefundData []takeoutRefundRawData
	takeoutRefundQuery := `
		SELECT 
			accepted_time,
			eater_payment AS refund_amount,
			1 AS refund_num
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state = ?
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		`
	r.db.Raw(takeoutRefundQuery,
		constant.NotDeleted,
		canceledOrderState, // 60 = 已取消
		req.StartTime, req.EndTime,
	).Scan(&takeoutRefundData)

	// 2.2. 查询外卖订单总数（用于计算退款率）
	type takeoutOrderCountData struct {
		AcceptedTime int64
		OrderNum     int64
	}
	var takeoutOrderCount []takeoutOrderCountData
	// 使用 statistics_takeout.go 中定义的 validOrderStates（包含所有有效状态：10, 20, 30, 40, 60）
	validStatesStr := buildStateInCondition(validOrderStates)
	takeoutOrderCountQuery := fmt.Sprintf(`
		SELECT 
			MIN(accepted_time) AS accepted_time,
			COUNT(DISTINCT uuid) AS order_num
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state IN %s
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		GROUP BY FROM_UNIXTIME(accepted_time, IF(? = 1, '%%Y-%%m', '%%Y-%%m-%%d'))
	`, validStatesStr)
	r.db.Raw(takeoutOrderCountQuery,
		constant.NotDeleted,
		req.StartTime, req.EndTime,
		req.Cycle,
	).Scan(&takeoutOrderCount)

	// 3. 在应用层按业务时区分组、统计
	timeUtil := utils.SetTimezone(req.Timezone)

	// 构建订单数量映射（按日期）- 包含销售订单和外卖订单
	orderCountMap := make(map[string]int64)
	for _, item := range orderCountData {
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01-02")
		}
		orderCountMap[dateKey] += item.OrderNum
	}
	// 合并外卖订单总数
	for _, item := range takeoutOrderCount {
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.AcceptedTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.AcceptedTime, "2006-01-02")
		}
		orderCountMap[dateKey] += item.OrderNum
	}

	// 按日期分组统计
	// 使用 decimal 进行精确计算
	type groupData struct {
		RefundAmount        decimal.Decimal
		RefundNum           int64
		PartialRefundAmount decimal.Decimal
		PartialRefundNum    int64
		FullRefundAmount    decimal.Decimal
		FullRefundNum       int64
		OrderNum            int64
	}
	groupedData := make(map[string]*groupData)
	for _, item := range rawData {
		// 将时间戳转换为业务时区的日期
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.FinishTime, "2006-01-02")
		}

		// 初始化分组数据
		if groupedData[dateKey] == nil {
			groupedData[dateKey] = &groupData{}
		}

		// 使用 decimal 累加统计数据
		group := groupedData[dateKey]
		group.RefundAmount = group.RefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
		group.RefundNum += item.RefundNum
		// 累加部分退款和整单退款（已在 SQL 中按类型分别统计）
		group.PartialRefundAmount = group.PartialRefundAmount.Add(decimal.NewFromFloat(item.PartialRefundAmount))
		group.PartialRefundNum += item.PartialRefundNum
		group.FullRefundAmount = group.FullRefundAmount.Add(decimal.NewFromFloat(item.FullRefundAmount))
		group.FullRefundNum += item.FullRefundNum
	}

	// 合并外卖订单退款数据
	for _, item := range takeoutRefundData {
		// 将时间戳转换为业务时区的日期
		var dateKey string
		if req.Cycle == 1 {
			// 按月
			dateKey = timeUtil.FormatUnixTime(item.AcceptedTime, "2006-01")
		} else {
			// 按日
			dateKey = timeUtil.FormatUnixTime(item.AcceptedTime, "2006-01-02")
		}

		// 初始化分组数据
		if groupedData[dateKey] == nil {
			groupedData[dateKey] = &groupData{}
		}

		// 使用 decimal 累加统计数据
		group := groupedData[dateKey]
		// 退款金额：+ 外卖订单取消金额 "eaterPayment"
		group.RefundAmount = group.RefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
		// 退款笔数：+ 外卖订单取消笔数
		group.RefundNum += item.RefundNum
		// 部分退款金额：+外卖订单（无部分退款金额，所以不加）
		// 部分退款笔数：+外卖订单（无部分退款笔数，所以不加）
		// 整单退款金额：+外卖订单取消金额 "eaterPayment"
		group.FullRefundAmount = group.FullRefundAmount.Add(decimal.NewFromFloat(item.RefundAmount))
		// 整单退款笔数：+外卖订单取消笔数
		group.FullRefundNum += item.RefundNum
	}

	// 设置订单数量（从订单数量映射中获取）
	for dateKey, group := range groupedData {
		if orderNum, exists := orderCountMap[dateKey]; exists {
			group.OrderNum = orderNum
		}
	}

	// 转换为结果格式
	result := make([]model.StatisticsRefundSummaryData, 0, len(groupedData))
	for dateKey, group := range groupedData {
		// 计算退款率：退款订单数量 / 总订单数量 * 100
		var refundRate decimal.Decimal
		if group.OrderNum > 0 {
			refundRate = decimal.NewFromInt(group.RefundNum).Div(decimal.NewFromInt(group.OrderNum)).Mul(decimal.NewFromInt(100))
		}

		result = append(result, model.StatisticsRefundSummaryData{
			Date:                sql.NullString{String: dateKey, Valid: true},
			RefundAmount:        sql.NullFloat64{Float64: group.RefundAmount.InexactFloat64(), Valid: true},
			RefundNum:           sql.NullInt64{Int64: group.RefundNum, Valid: true},
			RefundRate:          sql.NullFloat64{Float64: refundRate.InexactFloat64(), Valid: true},
			PartialRefundAmount: sql.NullFloat64{Float64: group.PartialRefundAmount.InexactFloat64(), Valid: true},
			PartialRefundNum:    sql.NullInt64{Int64: group.PartialRefundNum, Valid: true},
			FullRefundAmount:    sql.NullFloat64{Float64: group.FullRefundAmount.InexactFloat64(), Valid: true},
			FullRefundNum:       sql.NullInt64{Int64: group.FullRefundNum, Valid: true},
			OrderNum:            sql.NullInt64{Int64: group.OrderNum, Valid: true},
		})
	}

	// 4. 按日期排序
	slices.SortFunc(result, func(a, b model.StatisticsRefundSummaryData) int {
		if a.Date.String < b.Date.String {
			return -1
		} else if a.Date.String > b.Date.String {
			return 1
		}
		return 0
	})

	// 5. 分页
	total := int64(len(result))
	start := (req.PageNo - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(result) {
		start = len(result)
	}
	if end > len(result) {
		end = len(result)
	}
	if start < end {
		result = result[start:end]
	} else {
		result = []model.StatisticsRefundSummaryData{}
	}

	return total, result
}
