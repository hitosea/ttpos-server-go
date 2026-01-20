package repository

import (
	"database/sql"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	takeoutmodel "ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	// 有效状态：10=已接单配餐中, 20=待骑手接单, 30=骑手配送中, 40=已完成, 60=已取消，接单时间>0
	validOrderStates = []int{
		valueobject.TakeoutOrderStateAccepted,        // 10
		valueobject.TakeoutOrderStateRiderPending,    // 20
		valueobject.TakeoutOrderStateRiderProcessing, // 30
		valueobject.TakeoutOrderStateCompleted,       // 40
		valueobject.TakeoutOrderStateCanceled,        // 60
	}
	// 营业收入状态：10=已接单配餐中, 20=待骑手接单, 30=骑手配送中, 40=已完成（不包括60已取消），接单时间>0
	businessOrderStates = []int{
		valueobject.TakeoutOrderStateAccepted,        // 10
		valueobject.TakeoutOrderStateRiderPending,    // 20
		valueobject.TakeoutOrderStateRiderProcessing, // 30
		valueobject.TakeoutOrderStateCompleted,       // 40
	}
	// 拒单状态
	rejectedOrderState = valueobject.TakeoutOrderStateRejected // 50
	// 取消状态
	canceledOrderState = valueobject.TakeoutOrderStateCanceled // 60，接单时间>0
)

// buildStateInCondition 构建状态 IN 条件字符串
func buildStateInCondition(states []int) string {
	if len(states) == 0 {
		return ""
	}
	condition := "("
	for i, state := range states {
		if i > 0 {
			condition += ","
		}
		condition += fmt.Sprintf("%d", state)
	}
	condition += ")"
	return condition
}

// CountTakeoutReq 统计外卖订单请求参数
type CountTakeoutReq struct {
	TimeStart         int64  // 时间开始（时间戳）
	TimeEnd           int64  // 时间结束（时间戳）
	StaffShiftLogUuid uint64 // 员工班次日志UUID（0表示不筛选）
	Platform          string // 平台筛选（空字符串表示不筛选）
}

// IStatisticsTakeoutRepo 外卖统计仓储接口
type IStatisticsTakeoutRepo interface {
	CountTakeoutSale(req CountTakeoutReq) model.StatisticsTakeoutSaleData                                          // 统计外卖订单销售额
	CountTakeoutPayment(req CountTakeoutReq) []model.StatisticsTakeoutPaymentData                                  // 统计外卖订单支付方式
	CountTakeoutReceivedAmount(req CountTakeoutReq) []model.StatisticsTakeoutReceivedAmountData                    // 统计外卖订单实收金额
	RankTakeoutProduct(req CountTakeoutReq) []model.StatisticsProductData                                          // 统计外卖订单商品排行
	CountTakeoutBusinessTimePeriod(req CountTakeoutBusinessTimePeriodReq) []model.StatisticsBusinessTimePeriodData // 统计外卖订单营业时段
	CountTakeoutBusinessSummary(req CountTakeoutBusinessSummaryReq) []takeoutBusinessSummaryRawData                // 统计外卖订单综合运营（返回原始数据）
	CountTakeoutChannelSale(req CountTakeoutChannelSaleReq) []TakeoutChannelSaleRawData                            // 统计外卖订单渠道营业（返回原始数据）
	CountTakeoutChannelSaleByPlatform(req CountTakeoutChannelSaleReq, platform string) []TakeoutChannelSaleRawData // 统计外卖订单渠道营业（按平台）
	CountTakeoutPaymentMethodRawData(req CountTakeoutReq) []TakeoutPaymentMethodRawData                            // 查询外卖订单支付方式原始数据（用于合并统计）
	CountTakeoutCategory(req CountTakeoutReq, categoryType int, language string) []model.StatisticsCategoryData    // 统计外卖订单商品分类
	CountTakeoutProduct(req CountTakeoutReq, language string) []model.StatisticsProductData                        // 统计外卖订单商品
	CountTakeoutRefundAmount(req CountTakeoutReq) float64                                                          // 统计外卖订单退款金额
}

// CountTakeoutBusinessTimePeriodReq 统计外卖订单营业时段请求
type CountTakeoutBusinessTimePeriodReq struct {
	StartTime     int64 // 查询开始时间戳
	EndTime       int64 // 查询结束时间戳
	PeriodSeconds int   // 时段秒数
}

// CountTakeoutBusinessSummaryReq 统计外卖订单综合运营请求
type CountTakeoutBusinessSummaryReq struct {
	StartTime int64 // 查询开始时间戳
	EndTime   int64 // 查询结束时间戳
}

// takeoutBusinessSummaryRawData 外卖订单综合运营统计原始数据
type takeoutBusinessSummaryRawData struct {
	AcceptedTime int64   // 接单时间戳
	OrderUuid    uint64  // 订单UUID
	OrderAmount  float64 // 订单金额（subtotal）
	PayAmount    float64 // 实付金额（状态为60时=0，其他状态=eater_payment）
}

// CountTakeoutChannelSaleReq 统计外卖订单渠道营业请求
type CountTakeoutChannelSaleReq struct {
	StartTime int64 // 查询开始时间戳
	EndTime   int64 // 查询结束时间戳
}

// TakeoutChannelSaleRawData 外卖订单渠道营业统计原始数据
type TakeoutChannelSaleRawData struct {
	AcceptedTime int64   // 接单时间戳
	OrderUuid    uint64  // 订单UUID
	OrderAmount  float64 // 订单金额（subtotal）
	PayAmount    float64 // 实付金额（状态为60时=0，其他状态=eater_payment）
	OrderNum     int64   // 订单数
	RefundNum    int64   // 退款笔数
}

// TakeoutPaymentMethodRawData 外卖订单支付方式统计原始数据
type TakeoutPaymentMethodRawData struct {
	AcceptedTime            int64   // 接单时间戳
	PaymentMethodUuid       uint64  // 支付方式UUID
	PaymentMethodSort       int     // 支付方式排序
	PaymentMethodCreateTime int64   // 支付方式创建时间戳
	PaymentName             string  // 支付方式名称
	PaymentAmount           float64 // 支付金额（营业收入状态的 eater_payment，取消状态为0）
}

// NewStatisticsTakeoutRepo 创建外卖统计仓储实例
func NewStatisticsTakeoutRepo(db *gorm.DB) IStatisticsTakeoutRepo {
	return NewStatisticsTakeoutRepoImpl(db)
}

// StatisticsTakeoutRepo 外卖统计仓储实现
type StatisticsTakeoutRepo struct {
	db *gorm.DB
}

// NewStatisticsTakeoutRepoImpl 创建外卖统计仓储实例
func NewStatisticsTakeoutRepoImpl(db *gorm.DB) *StatisticsTakeoutRepo {
	return &StatisticsTakeoutRepo{db: db}
}

// CountTakeoutSale 统计外卖订单销售额
// 使用 IF 判断 order_state 进行统计
// 支持通过 req 传入筛选条件：
//   - TimeStart/TimeEnd: 按时间范围筛选
//   - StaffShiftLogUuid: 按员工班次日志UUID筛选（>0时生效）
//   - Platform: 按平台筛选（非空时生效）
//
// 统计规则：
//   - 仅统计有效状态的订单（10,20,30,40,60），且 accepted_time > 0（接单后才能统计）
//   - 取消状态（60）的订单需要 accepted_time > 0 才会被统计
//
// 字段说明：
//   - TotalSaleAmount: 总销售额（使用顾客实付 eater_payment，与堂食统计统一）
//   - MinOrderAmount/MaxOrderAmount: 最小/最大订单金额（使用顾客实付 eater_payment）
//   - TotalOrderAmount: 总订单金额（使用顾客实付 eater_payment，用于计算平均值）
//   - TotalProductAmount: 原商品金额（使用小计金额 subtotal，兼容前端）
func (r *StatisticsTakeoutRepo) CountTakeoutSale(req CountTakeoutReq) model.StatisticsTakeoutSaleData {
	var result model.StatisticsTakeoutSaleData
	baseQuery := r.db.Model(&takeoutmodel.TakeoutOrder{})

	// 只统计未删除的订单
	baseQuery = baseQuery.Where("ttpos_takeout_order.delete_time = ?", constant.NotDeleted)

	// 按接单时间范围筛选
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time >= ? AND ttpos_takeout_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("ttpos_takeout_order.platform = ?", req.Platform)
	}

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 使用子查询避免重复统计：先对订单进行聚合，再关联商品表统计商品数量
	// 订单级别的字段在子查询中已经去重（每个订单一行），商品数量通过关联查询统计
	// 使用 COALESCE 处理 NULL 值，直接返回普通类型
	selectFields := []string{
		// 1. 总销售额：当 order_state 为有效状态时统计（使用顾客实付）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.eater_payment, 0)), 0) AS total_sale_amount", validStatesStr),
		// 2. 总实付金额：当 order_state 为营业收入状态时统计
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.eater_payment, 0)), 0) AS total_pay_amount", businessStatesStr),
		// 3. 总订单数：当 order_state 为有效状态时统计
		fmt.Sprintf("COALESCE(COUNT(DISTINCT IF(t.order_state IN %s, t.uuid, NULL)), 0) AS total_order_num", validStatesStr),
		// 4. 总退款金额：当 order_state = 60 时统计（已取消订单的顾客实付）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state = %d, t.eater_payment, 0)), 0) AS total_refund_amount", canceledOrderState),
		// 5. 最小订单金额：当 order_state 为有效状态时统计（使用顾客实付）
		fmt.Sprintf("COALESCE(MIN(IF(t.order_state IN %s, t.eater_payment, NULL)), 0) AS min_order_amount", businessStatesStr),
		// 6. 最大订单金额：当 order_state 为有效状态时统计（使用顾客实付）
		fmt.Sprintf("COALESCE(MAX(IF(t.order_state IN %s, t.eater_payment, NULL)), 0) AS max_order_amount", businessStatesStr),
		// 7. 总优惠折扣：当 order_state 为有效状态时统计 platform_discount + merchant_discount + basket_promo
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.platform_discount + t.merchant_discount + t.basket_promo, 0)), 0) AS total_discount", validStatesStr),
		// 8. 总税费：当 order_state 为营业收入状态时统计 tax
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.tax, 0)), 0) AS total_tax", businessStatesStr),
		// 9. 总营业收入：当 order_state 为营业收入状态时统计（实付金额 - 税费）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.eater_payment - t.tax, 0)), 0) AS total_business_amount", businessStatesStr),
		// 10. 取消订单数：当 order_state = 60 时统计（已取消订单）
		fmt.Sprintf("COALESCE(COUNT(DISTINCT IF(t.order_state = %d, t.uuid, NULL)), 0) AS cancel_order_num", canceledOrderState),
		// 11. 取消订单金额：当 order_state = 60 时统计（已取消订单的顾客实付）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state = %d, t.eater_payment, 0)), 0) AS cancel_order_amount", canceledOrderState),
		// 12. 总商品数量：当 order_state 为有效状态时统计，关联商品表的 quantity 字段
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, IFNULL(t.total_quantity, 0), 0)), 0) AS total_product_num", validStatesStr),
		// 13. 总订单金额：当 order_state 为有效状态时统计（使用顾客实付，用于计算平均值）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.eater_payment, 0)), 0) AS total_order_amount", businessStatesStr),
		// 14. 原商品金额：当 order_state 为有效状态时统计（使用小计金额，兼容前端）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.subtotal, 0)), 0) AS total_product_origin_price", validStatesStr),
	}

	// 构建子查询：先对订单进行聚合（避免重复），同时关联商品表统计每个订单的商品数量
	subQuery := baseQuery.Select([]string{
		"ttpos_takeout_order.uuid",
		"ttpos_takeout_order.order_state",
		"ttpos_takeout_order.subtotal",
		"ttpos_takeout_order.eater_payment",
		"ttpos_takeout_order.platform_discount",
		"ttpos_takeout_order.merchant_discount",
		"ttpos_takeout_order.basket_promo",
		"ttpos_takeout_order.tax",
		// 使用子查询统计每个订单的商品数量（避免 JOIN 导致重复）
		fmt.Sprintf("(SELECT COALESCE(SUM(quantity), 0) FROM ttpos_takeout_order_item WHERE takeout_order_uuid = ttpos_takeout_order.uuid AND delete_time = %d) AS total_quantity", constant.NotDeleted),
	})

	// 使用子查询进行最终统计
	r.db.Table("(?) AS t", subQuery).
		Select(selectFields).
		Scan(&result)

	return result
}

// CountTakeoutPayment 统计外卖订单支付方式
// 根据 platform 字段映射支付方式：
//   - platform = "grab" → payment_name = "Grab"
//   - platform = "lineman" → payment_name = "LINE MAN"
//
// 统计规则：
//   - 仅统计 validOrderStates 状态的订单（10, 20, 30, 40, 60），且 accepted_time > 0（接单后才能统计）
//   - 取消状态（60）的订单需要 accepted_time > 0 才会被统计
//
// TotalPaymentAmount = eater_payment[10,20,30,40]（营业收入状态，优化后直接统计，避免先求和再相减）
// TotalRefundAmount = eater_payment[60]（已取消订单，且 accepted_time > 0）
func (r *StatisticsTakeoutRepo) CountTakeoutPayment(req CountTakeoutReq) []model.StatisticsTakeoutPaymentData {
	var result []model.StatisticsTakeoutPaymentData

	baseQuery := r.db.Model(&takeoutmodel.TakeoutOrder{})

	// 只统计未删除的订单
	baseQuery = baseQuery.Where("ttpos_takeout_order.delete_time = ?", constant.NotDeleted)

	// 仅统计有效状态的订单
	validStatesStr := buildStateInCondition(validOrderStates)
	baseQuery = baseQuery.Where(fmt.Sprintf("ttpos_takeout_order.order_state IN %s", validStatesStr))

	// 按接单时间范围筛选
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time >= ? AND ttpos_takeout_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("ttpos_takeout_order.platform = ?", req.Platform)
	}

	// 构建支付方式名称映射：根据 platform 字段映射到 payment_name
	// platform = "grab" → payment_name = "Grab"
	// platform = "lineman" → payment_name = "LINE MAN"
	paymentNameMapping := fmt.Sprintf("CASE WHEN ttpos_takeout_order.platform = '%s' THEN '%s' WHEN ttpos_takeout_order.platform = '%s' THEN '%s' ELSE NULL END AS mapped_payment_name",
		valueobject.TakeoutPlatformGrab, valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformGrab],
		valueobject.TakeoutPlatformLineman, valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformLineman])

	// 使用子查询先映射 payment_name，然后关联 payment_method 表
	subQuery := baseQuery.Select([]string{
		"ttpos_takeout_order.uuid",
		"ttpos_takeout_order.order_state",
		"ttpos_takeout_order.eater_payment",
		paymentNameMapping,
	})

	// 构建统计字段（使用子查询别名 t）
	// 使用 businessOrderStates（10,20,30,40）优化总支付金额计算，避免先求和再相减
	businessStatesStr := buildStateInCondition(businessOrderStates)
	selectFields := []string{
		"pm.id",
		"pm.sort",
		"pm.create_time",
		"pm.payment_name",
		"pm.code AS payment_code",
		"pm.erpnext_payment",
		"pm.erpnext_payment_id",
		"pm.source",
		// 总订单数：有效状态的订单数
		fmt.Sprintf("COALESCE(COUNT(DISTINCT IF(t.order_state IN %s, t.uuid, NULL)), 0) AS total_order_num", validStatesStr),
		// 总支付金额：eater_payment[10,20,30,40]（优化：直接统计营业收入状态，避免 eater_payment[10,20,30,40,60] - eater_payment[60]）
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state IN %s, t.eater_payment, 0)), 0) AS total_payment_amount", businessStatesStr),
		// 总退款金额：eater_payment[60]
		fmt.Sprintf("COALESCE(SUM(IF(t.order_state = %d, t.eater_payment, 0)), 0) AS total_refund_amount", canceledOrderState),
	}

	// 构建排序字段：Grab 在 LINE MAN 前
	// 使用 CASE WHEN 指定排序顺序：Grab = 1, LINE MAN = 2
	orderByPaymentName := fmt.Sprintf("CASE WHEN pm.payment_name = '%s' THEN 1 WHEN pm.payment_name = '%s' THEN 2 ELSE 99 END ASC",
		valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformGrab],
		valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformLineman])

	// 关联 payment_method 表，根据 payment_name 匹配
	// 使用子查询结果关联 payment_method 表
	// 添加 source = 0（系统默认）和 code 的过滤条件
	r.db.Table("(?) AS t", subQuery).
		Select(selectFields).
		Joins("LEFT JOIN ttpos_payment_method AS pm ON t.mapped_payment_name = pm.payment_name AND pm.delete_time = ? AND pm.source = ? AND pm.code IN (?, ?)", constant.NotDeleted, constant.PaymentMethodSourceSystem, constant.PaymentMethodCodeGrab, constant.PaymentMethodCodeLineMan).
		Where("t.mapped_payment_name IS NOT NULL"). // 只统计有映射的支付方式
		Where("pm.uuid IS NOT NULL").               // 只统计关联到支付方式的记录
		Group("pm.uuid").
		Order(orderByPaymentName). // Grab 在 LINE MAN 前
		Scan(&result)

	return result
}

// CountTakeoutReceivedAmount 统计外卖订单实收金额
// 查询 validOrderStates 状态下的订单，且 accepted_time > 0（接单后才能统计）
// 使用 IF 判断：如果状态是营业收入状态(10,20,30,40)，则取 eater_payment；如果状态是取消状态(60)，则取 0
// 返回每个订单的接单时间、订单UUID和实收金额
func (r *StatisticsTakeoutRepo) CountTakeoutReceivedAmount(req CountTakeoutReq) []model.StatisticsTakeoutReceivedAmountData {
	var result []model.StatisticsTakeoutReceivedAmountData

	baseQuery := r.db.Model(&takeoutmodel.TakeoutOrder{})

	// 只统计未删除的订单
	baseQuery = baseQuery.Where("ttpos_takeout_order.delete_time = ?", constant.NotDeleted)

	// 仅统计有效状态的订单
	validStatesStr := buildStateInCondition(validOrderStates)
	baseQuery = baseQuery.Where(fmt.Sprintf("ttpos_takeout_order.order_state IN %s", validStatesStr))

	// 按接单时间范围筛选
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time >= ? AND ttpos_takeout_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("ttpos_takeout_order.platform = ?", req.Platform)
	}

	// 构建状态条件字符串
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 计算每个订单的实收金额：
	// 使用 IF 判断：如果状态是营业收入状态(10,20,30,40)，则取 eater_payment；如果状态是取消状态(60)，则取 0
	selectFields := []string{
		"ttpos_takeout_order.accepted_time",
		"ttpos_takeout_order.uuid AS takeout_order_uuid",
		fmt.Sprintf("IF(ttpos_takeout_order.order_state IN %s, ttpos_takeout_order.eater_payment, 0) AS total_received_amount", businessStatesStr),
	}

	baseQuery.Select(selectFields).
		Order("ttpos_takeout_order.accepted_time ASC").
		Scan(&result)

	return result
}

// RankTakeoutProduct 统计外卖订单商品排行
// 统计有效状态的订单（10,20,30,40,60），且 accepted_time > 0（接单后才能统计）
// 按 product_package_uuid 分组统计销售数量和金额
// 销量：所有 validOrderStates 都统计（包括取消订单60）
// 销售金额：只统计 businessOrderStates（不包括取消订单60）
func (r *StatisticsTakeoutRepo) RankTakeoutProduct(req CountTakeoutReq) []model.StatisticsProductData {
	var result []model.StatisticsProductData

	prefix := config.Database.TablePrefix
	takeoutOrderTable := prefix + "takeout_order"
	takeoutOrderItemTable := prefix + "takeout_order_item"

	// 统计有效状态的订单（10,20,30,40,60），包括取消订单（60）
	validStatesStr := buildStateInCondition(validOrderStates)
	// 营业收入状态（10,20,30,40），不包括取消订单（60），用于计算销售金额
	businessStatesStr := buildStateInCondition(businessOrderStates)

	query := r.db.Table(takeoutOrderItemTable+" AS toi").
		Select(
			"toi.price AS sale_price",
			"toi.ttpos_product_package_uuid AS product_package_uuid",
			// 销量：所有 validOrderStates 都统计（包括取消订单60）
			"SUM(CAST(toi.quantity AS DECIMAL(14,2))) AS sale_num",
			// 销售金额：只统计 businessOrderStates（不包括取消订单60）
			fmt.Sprintf("SUM(IF(to_order.order_state IN %s, toi.price * toi.quantity, 0)) AS sale_amount", businessStatesStr),
		).
		Joins(fmt.Sprintf("INNER JOIN %s AS to_order ON toi.takeout_order_uuid = to_order.uuid", takeoutOrderTable)).
		Where("toi.delete_time = ?", constant.NotDeleted).
		Where("to_order.delete_time = ?", constant.NotDeleted).
		Where(fmt.Sprintf("to_order.order_state IN %s", validStatesStr)).
		Where("toi.ttpos_product_package_uuid > 0"). // 只统计已映射的商品
		Group("toi.ttpos_product_package_uuid")

	// 应用时间条件（使用 accepted_time）
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		query = query.Where("to_order.accepted_time >= ? AND to_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	query = query.Where("to_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		query = query.Where("to_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		query = query.Where("to_order.platform = ?", req.Platform)
	}

	query.Find(&result)

	return result
}

// CountTakeoutBusinessTimePeriod 统计外卖订单营业时段
// 使用 accepted_time（接单时间）进行时段分组
// 统计规则：
//   - 仅统计有效状态的订单（10,20,30,40,60），且 accepted_time > 0（接单后才能统计）
//   - 取消状态（60）的订单需要 accepted_time > 0 才会被统计
//
// 订单金额：所有有效状态的 subtotal
// 实付金额：状态为60（已取消）时=0，其他有效状态=eater_payment
// 订单数量：按 uuid 去重统计
// 用餐人数：始终为0（外卖订单无用餐人数）
func (r *StatisticsTakeoutRepo) CountTakeoutBusinessTimePeriod(req CountTakeoutBusinessTimePeriodReq) []model.StatisticsBusinessTimePeriodData {
	var result []model.StatisticsBusinessTimePeriodData

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 构建查询SQL
	// 使用 accepted_time 进行时段分组
	mainQuery := fmt.Sprintf(`
		SELECT 
			period_start_time,
			SUM(order_amount) AS order_amount,
			SUM(pay_amount) AS pay_amount,
			0 AS refund_amount,
			COUNT(DISTINCT order_uuid) AS order_num,
			0 AS meal_num
		FROM (
			SELECT 
				FLOOR(accepted_time / %d) * %d AS period_start_time,
				IF(order_state IN %s, subtotal, 0) AS order_amount,
				IF(order_state = %d, 0, IF(order_state IN %s, eater_payment, 0)) AS pay_amount,
				uuid AS order_uuid
			FROM ttpos_takeout_order
			WHERE delete_time = ?
				AND order_state IN %s
				AND accepted_time > 0
				AND accepted_time >= ?
				AND accepted_time <= ?
		) AS subquery
		GROUP BY period_start_time
		ORDER BY period_start_time ASC
	`, req.PeriodSeconds, req.PeriodSeconds, validStatesStr, canceledOrderState, businessStatesStr, validStatesStr)

	// 执行查询
	r.db.Raw(mainQuery, constant.NotDeleted, req.StartTime, req.EndTime).Scan(&result)

	return result
}

// CountTakeoutBusinessSummary 统计外卖订单综合运营（返回原始数据，不分组）
// 使用 accepted_time（接单时间）作为时间字段
// 统计规则：
//   - 仅统计有效状态的订单（10,20,30,40,60），且 accepted_time > 0（接单后才能统计）
//   - 取消状态（60）的订单需要 accepted_time > 0 才会被统计
//
// 订单金额：所有有效状态的 subtotal
// 实付金额：状态为60（已取消）时=0，其他有效状态=eater_payment
func (r *StatisticsTakeoutRepo) CountTakeoutBusinessSummary(req CountTakeoutBusinessSummaryReq) []takeoutBusinessSummaryRawData {
	var result []takeoutBusinessSummaryRawData

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 构建查询SQL
	mainQuery := fmt.Sprintf(`
		SELECT 
			accepted_time,
			uuid AS order_uuid,
			IF(order_state IN %s, eater_payment, 0) AS order_amount,
			IF(order_state = %d, 0, IF(order_state IN %s, eater_payment, 0)) AS pay_amount
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state IN %s
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		ORDER BY accepted_time ASC
	`, validStatesStr, canceledOrderState, businessStatesStr, validStatesStr)

	// 执行查询
	r.db.Raw(mainQuery, constant.NotDeleted, req.StartTime, req.EndTime).Scan(&result)

	return result
}

// CountTakeoutChannelSale 统计外卖订单渠道营业（返回原始数据，不分组）
// 使用 accepted_time（接单时间）作为时间字段
// 订单金额：所有有效状态的 subtotal
// 实付金额：状态为60（已取消）时=0，其他有效状态=eater_payment
func (r *StatisticsTakeoutRepo) CountTakeoutChannelSale(req CountTakeoutChannelSaleReq) []TakeoutChannelSaleRawData {
	var result []TakeoutChannelSaleRawData

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 构建查询SQL
	mainQuery := fmt.Sprintf(`
		SELECT 
			accepted_time,
			uuid AS order_uuid,
			IF(order_state IN %s, subtotal, 0) AS order_amount,
			IF(order_state = %d, 0, IF(order_state IN %s, eater_payment, 0)) AS pay_amount
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state IN %s
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		ORDER BY accepted_time ASC
	`, validStatesStr, canceledOrderState, businessStatesStr, validStatesStr)

	// 执行查询
	r.db.Raw(mainQuery, constant.NotDeleted, req.StartTime, req.EndTime).Scan(&result)

	return result
}

// CountTakeoutChannelSaleByPlatform 统计外卖订单渠道营业（按平台，返回原始数据，不分组）
// 使用 accepted_time（接单时间）作为时间字段
// 订单金额：所有有效状态的 subtotal
// 实付金额：状态为60（已取消）时=0，其他有效状态=eater_payment
func (r *StatisticsTakeoutRepo) CountTakeoutChannelSaleByPlatform(req CountTakeoutChannelSaleReq, platform string) []TakeoutChannelSaleRawData {
	var result []TakeoutChannelSaleRawData

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 构建查询SQL
	mainQuery := fmt.Sprintf(`
		SELECT 
			accepted_time,
			uuid AS order_uuid,
			IF(order_state IN %s, subtotal, 0) AS order_amount,
			IF(order_state = %d, 0, IF(order_state IN %s, eater_payment, 0)) AS pay_amount
		FROM ttpos_takeout_order
		WHERE delete_time = ?
			AND order_state IN %s
			AND platform = ?
			AND accepted_time > 0
			AND accepted_time >= ?
			AND accepted_time <= ?
		ORDER BY accepted_time ASC
	`, validStatesStr, canceledOrderState, businessStatesStr, validStatesStr)

	// 执行查询
	r.db.Raw(mainQuery, constant.NotDeleted, platform, req.StartTime, req.EndTime).Scan(&result)

	return result
}

// CountTakeoutPaymentMethodRawData 查询外卖订单支付方式原始数据（用于合并统计）
// 查询 validOrderStates 状态下的订单，且 accepted_time > 0（接单后才能统计）
// 使用 IF 判断：如果状态是营业收入状态(10,20,30,40)，则取 eater_payment；如果状态是取消状态(60)，则取 0
// 返回每个订单的接单时间、支付方式UUID、支付方式排序、支付方式创建时间、支付方式名称和支付金额
func (r *StatisticsTakeoutRepo) CountTakeoutPaymentMethodRawData(req CountTakeoutReq) []TakeoutPaymentMethodRawData {
	var result []TakeoutPaymentMethodRawData

	baseQuery := r.db.Model(&takeoutmodel.TakeoutOrder{})

	// 只统计未删除的订单
	baseQuery = baseQuery.Where("ttpos_takeout_order.delete_time = ?", constant.NotDeleted)

	// 仅统计有效状态的订单
	validStatesStr := buildStateInCondition(validOrderStates)
	baseQuery = baseQuery.Where(fmt.Sprintf("ttpos_takeout_order.order_state IN %s", validStatesStr))

	// 按接单时间范围筛选
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time >= ? AND ttpos_takeout_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("ttpos_takeout_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("ttpos_takeout_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("ttpos_takeout_order.platform = ?", req.Platform)
	}

	// 构建状态条件字符串
	businessStatesStr := buildStateInCondition(businessOrderStates)

	// 构建支付方式名称映射：根据 platform 字段映射到 payment_name
	// platform = "grab" → payment_name = "Grab"
	// platform = "lineman" → payment_name = "LINE MAN"
	paymentNameMapping := fmt.Sprintf("CASE WHEN ttpos_takeout_order.platform = '%s' THEN '%s' WHEN ttpos_takeout_order.platform = '%s' THEN '%s' ELSE NULL END AS mapped_payment_name",
		valueobject.TakeoutPlatformGrab, valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformGrab],
		valueobject.TakeoutPlatformLineman, valueobject.TakeoutPlatformNames[valueobject.TakeoutPlatformLineman])

	// 使用子查询先映射 payment_name，然后关联 payment_method 表
	subQuery := baseQuery.Select([]string{
		"ttpos_takeout_order.accepted_time",
		"ttpos_takeout_order.order_state",
		"ttpos_takeout_order.eater_payment",
		paymentNameMapping,
	})

	// 构建查询字段（使用子查询别名 t）
	selectFields := []string{
		"t.accepted_time",
		"pm.uuid AS payment_method_uuid",
		"pm.sort AS payment_method_sort",
		"pm.create_time AS payment_method_create_time",
		"pm.payment_name",
		// 计算每个订单的支付金额：如果状态是营业收入状态(10,20,30,40)，则取 eater_payment；如果状态是取消状态(60)，则取 0
		fmt.Sprintf("IF(t.order_state IN %s, t.eater_payment, 0) AS payment_amount", businessStatesStr),
	}

	// 关联 payment_method 表，根据 payment_name 匹配
	// 使用子查询结果关联 payment_method 表
	// 添加 source = 0（系统默认）和 code 的过滤条件
	r.db.Table("(?) AS t", subQuery).
		Select(selectFields).
		Joins("LEFT JOIN ttpos_payment_method AS pm ON t.mapped_payment_name = pm.payment_name AND pm.delete_time = ? AND pm.source = ? AND pm.code IN (?, ?)", constant.NotDeleted, constant.PaymentMethodSourceSystem, constant.PaymentMethodCodeGrab, constant.PaymentMethodCodeLineMan).
		Where("t.mapped_payment_name IS NOT NULL"). // 只统计有映射的支付方式
		Where("pm.uuid IS NOT NULL").               // 只统计关联到支付方式的记录
		Order("t.accepted_time ASC").
		Scan(&result)

	return result
}

// CountTakeoutCategory 统计外卖订单商品分类
// 统计 validOrderStates 状态下的订单商品，按分类分组，且 accepted_time > 0（接单后才能统计）
// 销售量：所有 validOrderStates 都统计（包括取消订单，但需 accepted_time > 0）
// 销售额：只统计 businessOrderStates（不包括取消订单），使用商品实收*售卖数量
func (r *StatisticsTakeoutRepo) CountTakeoutCategory(req CountTakeoutReq, categoryType int, language string) []model.StatisticsCategoryData {
	var result []model.StatisticsCategoryData

	// 获取语言，确保语言是支持的语言
	locale := constant.LocaleList.GetLocaleType(language)
	language = string(locale)

	prefix := config.Database.TablePrefix
	takeoutOrderTable := prefix + "takeout_order"
	takeoutOrderItemTable := prefix + "takeout_order_item"
	productPackageTable := prefix + "product_package as pp"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	baseQuery := r.db.Table(takeoutOrderItemTable+" AS toi").
		Joins(fmt.Sprintf("INNER JOIN %s AS to_order ON toi.takeout_order_uuid = to_order.uuid", takeoutOrderTable)).
		Joins("LEFT JOIN "+productPackageTable+" ON toi.ttpos_product_package_uuid = pp.uuid").
		Joins("LEFT JOIN "+productCategoryTable+" ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN "+productParentCategoryTable+" ON pc.parent_uuid = ppc.uuid").
		Where("toi.delete_time = ?", constant.NotDeleted).
		Where("to_order.delete_time = ?", constant.NotDeleted).
		Where(fmt.Sprintf("to_order.order_state IN %s", validStatesStr)).
		Where("toi.ttpos_product_package_uuid > 0"). // 只统计已映射的商品
		Where("pp.category_uuid > 0")                // 只统计有分类的商品

	// 应用时间条件（使用 accepted_time）
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("to_order.accepted_time >= ? AND to_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("to_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("to_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("to_order.platform = ?", req.Platform)
	}

	if categoryType != 2 {
		// 一级分类统计
		baseQuery.Select(
			"IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid) AS category_parent_uuid",
			"IF(pc.parent_uuid = 0, JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')), JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"'))) AS category_parent_name",
			"0 AS category_uuid",
			"'' AS category_name",
			// 销售量：所有 validOrderStates 都统计（包括取消订单）
			"SUM(CAST(toi.quantity AS DECIMAL(14,2))) AS sale_num",
			// 销售额：只统计 businessOrderStates（不包括取消订单），使用商品实收*售卖数量
			fmt.Sprintf("SUM(IF(to_order.order_state IN %s, toi.price * toi.quantity, 0)) AS sale_amount", businessStatesStr),
			"IF(pc.parent_uuid > 0, ppc.sort, pc.sort) AS sort",
		).
			Group("IF(pc.parent_uuid = 0, pp.category_uuid, pc.parent_uuid)").
			Order("sort ASC").
			Order("category_parent_uuid DESC").
			Find(&result)
	} else {
		// 二级分类统计
		baseQuery.Select(
			"pc.parent_uuid AS category_parent_uuid",
			"JSON_UNQUOTE(JSON_EXTRACT(ppc.NAME, '$."+language+"')) AS category_parent_name",
			"pc.uuid AS category_uuid",
			"JSON_UNQUOTE(JSON_EXTRACT(pc.NAME, '$."+language+"')) AS category_name",
			// 销售量：所有 validOrderStates 都统计（包括取消订单）
			"SUM(CAST(toi.quantity AS DECIMAL(14,2))) AS sale_num",
			// 销售额：只统计 businessOrderStates（不包括取消订单），使用商品实收*售卖数量
			fmt.Sprintf("SUM(IF(to_order.order_state IN %s, toi.price * toi.quantity, 0)) AS sale_amount", businessStatesStr),
		).
			Where("pc.parent_uuid > 0").
			Group("pc.uuid").
			Order("ppc.sort ASC").
			Order("ppc.uuid DESC").
			Order("pc.sort ASC").
			Order("pc.uuid DESC").
			Find(&result)
	}

	return result
}

// CountTakeoutProduct 统计外卖订单商品
// 统计 validOrderStates 状态下的订单商品，按 product_bom_uuid 分组，且 accepted_time > 0（接单后才能统计）
// 商品名称：使用店内名称（从 ttpos_item_name JSON 字段提取）
// 销量：所有 validOrderStates 都统计（包括取消订单，但需 accepted_time > 0）
// 单价：从 product_bom 表获取（普通商品通过 takeout_order_item_modifier 关联，套餐商品直接关联 takeout_order_item）
// 合计：只统计 businessOrderStates（不包括取消订单），使用商品实收=单价*数量
// 分组：按 product_bom_uuid 分组，和 CountProduct 一致（普通商品使用 pb_flavor.uuid，套餐商品使用 pb_package.uuid）
func (r *StatisticsTakeoutRepo) CountTakeoutProduct(req CountTakeoutReq, language string) []model.StatisticsProductData {
	var result []model.StatisticsProductData

	// 获取语言，确保语言是支持的语言
	locale := constant.LocaleList.GetLocaleType(language)
	language = string(locale)

	prefix := config.Database.TablePrefix
	takeoutOrderTable := prefix + "takeout_order"
	takeoutOrderItemTable := prefix + "takeout_order_item"
	takeoutOrderItemModifierTable := prefix + "takeout_order_item_modifier as toim"
	productPackageTable := prefix + "product_package as pp"
	productCategoryTable := prefix + "product_category as pc"
	productParentCategoryTable := prefix + "product_category as ppc"
	// 普通商品的 product_bom 关联（通过 modifier）
	productBomFlavorTable := prefix + "product_bom as pb_flavor"
	// 套餐商品的 product_bom 关联（直接关联 item）
	productBomPackageTable := prefix + "product_bom as pb_package"

	// 构建状态条件字符串
	validStatesStr := buildStateInCondition(validOrderStates)
	businessStatesStr := buildStateInCondition(businessOrderStates)

	baseQuery := r.db.Table(takeoutOrderItemTable+" AS toi").
		Joins(fmt.Sprintf("INNER JOIN %s AS to_order ON toi.takeout_order_uuid = to_order.uuid", takeoutOrderTable)).
		Joins("LEFT JOIN "+productPackageTable+" ON toi.ttpos_product_package_uuid = pp.uuid").
		Joins("LEFT JOIN "+productCategoryTable+" ON pp.category_uuid = pc.uuid").
		Joins("LEFT JOIN "+productParentCategoryTable+" ON pc.parent_uuid = ppc.uuid").
		// 关联修饰符表，获取规格名称（只关联 flavor 类型的修饰符）
		Joins(fmt.Sprintf("LEFT JOIN %s ON toim.takeout_order_item_uuid = toi.uuid AND toim.ttpos_modifier_type = 'flavor' AND toim.delete_time = %d", takeoutOrderItemModifierTable, constant.NotDeleted)).
		// 关联 product_bom 表（普通商品：通过 modifier 关联）
		Joins(fmt.Sprintf("LEFT JOIN %s ON pb_flavor.uuid = toim.ttpos_modifier_uuid AND toim.ttpos_modifier_type = 'flavor'", productBomFlavorTable)).
		// 关联 product_bom 表（套餐商品：直接关联 item，条件为 product_flavor_uuid = 0 AND product_sauce_uuid = 0）
		Joins(fmt.Sprintf("LEFT JOIN %s ON pb_package.product_package_uuid = toi.ttpos_product_package_uuid AND pb_package.product_flavor_uuid = 0 AND pb_package.product_sauce_uuid = 0", productBomPackageTable)).
		Where("toi.delete_time = ?", constant.NotDeleted).
		Where("to_order.delete_time = ?", constant.NotDeleted).
		Where(fmt.Sprintf("to_order.order_state IN %s", validStatesStr)).
		Where("toi.ttpos_product_package_uuid > 0") // 只统计已映射的商品

	// 应用时间条件（使用 accepted_time）
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		baseQuery = baseQuery.Where("to_order.accepted_time >= ? AND to_order.accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（有效状态和取消状态都需要接单后才能统计）
	baseQuery = baseQuery.Where("to_order.accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		baseQuery = baseQuery.Where("to_order.staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		baseQuery = baseQuery.Where("to_order.platform = ?", req.Platform)
	}

	// 商品名称：优先使用外卖订单商品的 ttpos_item_name（JSON格式），否则使用 product_package 的 name
	// 规格名称：从 ttpos_takeout_order_item_modifier 表的 ttpos_modifier_name 获取（只取 flavor 类型）
	// ttpos_modifier_name 是多语言 JSON 字符串，需要提取对应语言的值
	// 如果是套餐商品（ttpos_product_type = 1），则没有规格名称
	// 单价：从 product_bom 表获取（普通商品使用 pb_flavor.price，套餐商品使用 pb_package.price）
	// 销量：所有 validOrderStates 都统计（包括取消订单）
	// 合计：只统计 businessOrderStates（不包括取消订单），使用商品实收=单价*数量
	err := baseQuery.Select(
		"toi.ttpos_product_package_uuid AS product_package_uuid",
		// product_bom_uuid：普通商品使用 pb_flavor.uuid，套餐商品使用 pb_package.uuid
		// 使用 MAX 聚合函数，因为已经按相同的表达式分组
		"MAX(IF(toi.ttpos_product_type = 1, pb_package.uuid, pb_flavor.uuid)) AS product_bom_uuid",
		fmt.Sprintf("COALESCE(MAX(JSON_UNQUOTE(JSON_EXTRACT(toi.ttpos_item_name, '$.%s'))), JSON_UNQUOTE(JSON_EXTRACT(pp.name, '$.%s')), '') AS product_name", language, language),
		// 规格名称：如果是套餐商品（ttpos_product_type = 1），则为空；否则从修饰符表获取（JSON格式需要提取）
		// 注意：在 GROUP BY 中需要使用原始表达式，不能使用聚合函数
		fmt.Sprintf("IF(MAX(toi.ttpos_product_type) = 1, '', COALESCE(MAX(JSON_UNQUOTE(JSON_EXTRACT(toim.ttpos_modifier_name, '$.%s'))), '')) AS flavor_name", language),
		// 单价：从 product_bom 表获取
		// 普通商品使用 pb_flavor.price，套餐商品使用 pb_package.price
		fmt.Sprintf("COALESCE(MAX(IF(toi.ttpos_product_type = 1, pb_package.price, pb_flavor.price)), 0) AS sale_price"),
		// 销量：所有 validOrderStates 都统计（包括取消订单）
		"SUM(CAST(toi.quantity AS DECIMAL(14,2))) AS sale_num",
		// 合计：只统计 businessOrderStates（不包括取消订单），使用商品实收=单价*数量
		fmt.Sprintf("SUM(IF(to_order.order_state IN %s, toi.price * toi.quantity, 0)) AS sale_amount", businessStatesStr),
		"MAX(toi.ttpos_product_type) AS product_type",
		// 排序字段：和 CountProduct 一致
		"IF(MAX(pc.parent_uuid) = 0, MAX(pc.sort), MAX(ppc.sort)) AS ppc_sort",
		"IF(MAX(pc.parent_uuid) = 0, MAX(pc.create_time), MAX(ppc.create_time)) AS ppc_create_time",
		"IF(MAX(pc.parent_uuid) = 0, 0, MAX(pc.sort)) AS pc_sort",
		"MAX(pc.create_time) AS pc_create_time",
		"MAX(pp.create_time) AS pp_create_time",
	).
		// 按 product_bom_uuid 分组，和 CountProduct 一致
		// 普通商品使用 pb_flavor.uuid，套餐商品使用 pb_package.uuid
		Group("IF(toi.ttpos_product_type = 1, pb_package.uuid, pb_flavor.uuid)").
		// 排序：和 CountProduct 一致
		Order("ppc_sort ASC").
		Order("ppc_create_time DESC").
		Order("pc_sort ASC").
		Order("pc.create_time DESC").
		Order("pp.create_time DESC").
		Find(&result).Error
	if err != nil {
		// 记录错误日志
		logger.Logger.Error("查询外卖订单商品失败", zap.Error(err))
	}

	return result
}

// CountTakeoutRefundAmount 统计外卖订单退款金额
// 统计 canceledOrderState（60）状态下的订单，且 accepted_time > 0（接单后才能统计），退款金额为 eater_payment
func (r *StatisticsTakeoutRepo) CountTakeoutRefundAmount(req CountTakeoutReq) float64 {
	var amount sql.NullFloat64

	prefix := config.Database.TablePrefix
	takeoutOrderTable := prefix + "takeout_order"

	query := r.db.Table(takeoutOrderTable).
		Select("COALESCE(SUM(eater_payment), 0) AS amount").
		Where("delete_time = ?", constant.NotDeleted).
		Where("order_state = ?", canceledOrderState) // 只统计取消状态的订单

	// 应用时间条件（使用 accepted_time）
	if req.TimeStart > 0 && req.TimeEnd > 0 {
		query = query.Where("accepted_time >= ? AND accepted_time <= ?", req.TimeStart, req.TimeEnd)
	}

	// 仅统计接单时间>0的订单（取消状态需要接单后才能统计）
	query = query.Where("accepted_time > 0")

	// 按员工班次日志UUID筛选
	if req.StaffShiftLogUuid > 0 {
		query = query.Where("staff_shift_log_uuid = ?", req.StaffShiftLogUuid)
	}

	// 按平台筛选
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}

	if err := query.Scan(&amount).Error; err != nil {
		// 记录日志，但不中断统计流程
		logger.Logger.Warn("查询外卖订单退款金额失败",
			zap.Error(err),
			zap.Int64("timeStart", req.TimeStart),
			zap.Int64("timeEnd", req.TimeEnd),
		)
		return 0 // 返回默认值 0
	}

	return amount.Float64
}
