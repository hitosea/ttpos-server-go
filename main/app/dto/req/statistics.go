package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"
)

type BusinessDataPrinterReq struct {
	TimeType       int    `json:"time_type"`        // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月, 5 营业时间)
	StatisticsType int    `json:"statistics_type"`  // 统计类型 (0 全部, 1 按支付方式, 2 按商品分类, 3 按商品)
	QueryStartTime uint   `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   uint   `json:"query_end_time"`   // 查询结束时间戳
	QueryStartDate string `json:"query_start_date"` // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate   string `json:"query_end_date"`   // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	CategoryType   int    `json:"category_type"`    // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
	NotQueryFree   bool   `json:"not_query_free"`   // 是否不查询免费使用场景
	SortType       int    `json:"sort_type"`        // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection  int    `json:"sort_direction"`   // 排序方向 0=默认、 1=升序、 2=降序
}

// GetParam 获取参数
// 参数优先级：1) time_type, 2) query_start_time + query_end_time, 3) query_start_date + query_end_date
func (r *BusinessDataPrinterReq) GetParam(timezone string, openingHours string) BusinessDataPrinterReq {
	var (
		queryStartTime int64
		queryEndTime   int64
	)

	// 优先处理时间类型（保持原有优先级）
	if r.TimeType > 0 && r.TimeType < 6 {
		switch r.TimeType {
		case 1: // 今天
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).TodayStartEndUnix()
		case 2: // 昨天
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).YesterdayStartEndUnix()
		case 3: // 本周
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).WeekStartEndUnix()
		case 4: // 本月
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).MonthStartEndUnix()
		case 5: // 营业时间
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).OpeningHoursStartEndUnix(openingHours)
		}
	}

	// 其次处理时间戳参数（如果设置了，会覆盖 time_type）
	if r.QueryStartTime > 0 && r.QueryEndTime > 0 {
		queryStartTime = int64(r.QueryStartTime)
		queryEndTime = int64(r.QueryEndTime)
	}

	// 最后处理日期时间字符串参数（如果设置了，会覆盖前面的）
	if r.QueryStartDate != "" && r.QueryEndDate != "" {
		timeUtil := utils.SetTimezone(timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(r.QueryStartDate)
		if err == nil {
			queryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(r.QueryEndDate)
		if err == nil {
			queryEndTime = endTime
		}
	}

	return BusinessDataPrinterReq{
		TimeType:       r.TimeType,
		StatisticsType: r.StatisticsType,
		QueryStartTime: uint(queryStartTime),
		QueryEndTime:   uint(queryEndTime),
		CategoryType:   r.CategoryType,
		NotQueryFree:   r.NotQueryFree,
		SortType:       r.SortType,
		SortDirection:  r.SortDirection,
	}
}

// BusinessDataCountReq 营业数据统计请求
type BusinessDataCountReq struct {
	TimeType          int    `form:"time_type" json:"time_type"`                     // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月, 5 营业时间)
	QueryStartTime    int64  `form:"query_start_time" json:"query_start_time"`       // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time" json:"query_end_time"`           // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	CategoryType      int    `form:"category_type" json:"category_type"`             // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
	DutyNo            string `form:"duty_no" json:"duty_no"`                         // 班次编号
	NotQueryFree      bool   `form:"not_query_free" json:"not_query_free"`           // 是否不查询免费使用场景
	StaffUuid         uint64 `form:"staff_uuid" json:"staff_uuid"`                   // 操作员UUID
	SortType          int    `form:"sort_type" json:"sort_type"`                     // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection     int    `form:"sort_direction" json:"sort_direction"`           // 排序方向 0=默认、 1=升序、 2=降序
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}

// GetParam 获取参数
// 参数优先级：1) time_type, 2) query_start_time + query_end_time, 3) query_start_date + query_end_date
func (r *BusinessDataCountReq) GetParam(timezone string, openingHours string) BusinessDataCountReq {
	var (
		queryStartTime int64
		queryEndTime   int64
	)

	// 优先处理时间类型（保持原有优先级）
	if r.TimeType > 0 && r.TimeType < 6 {
		switch r.TimeType {
		case 1: // 今天
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).TodayStartEndUnix()
		case 2: // 昨天
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).YesterdayStartEndUnix()
		case 3: // 本周
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).WeekStartEndUnix()
		case 4: // 本月
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).MonthStartEndUnix()
		case 5: // 营业时间
			queryStartTime, queryEndTime = utils.SetTimezone(timezone).OpeningHoursStartEndUnix(openingHours)
		}
	}

	// 其次处理时间戳参数（如果设置了，会覆盖 time_type）
	if r.QueryStartTime > 0 && r.QueryEndTime > 0 {
		queryStartTime = int64(r.QueryStartTime)
		queryEndTime = int64(r.QueryEndTime)
	}

	// 最后处理日期时间字符串参数（如果设置了，会覆盖前面的）
	if r.QueryStartDate != "" && r.QueryEndDate != "" {
		timeUtil := utils.SetTimezone(timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(r.QueryStartDate)
		if err == nil {
			queryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(r.QueryEndDate)
		if err == nil {
			queryEndTime = endTime
		}
	}

	return BusinessDataCountReq{
		TimeType:          r.TimeType,
		QueryStartTime:    queryStartTime,
		QueryEndTime:      queryEndTime,
		CategoryType:      r.CategoryType,
		DutyNo:            r.DutyNo,
		NotQueryFree:      r.NotQueryFree,
		SortType:          r.SortType,
		SortDirection:     r.SortDirection,
		ExcludeDataManage: r.ExcludeDataManage,
	}
}

// BusinessDataRankProductReq 营业数据排行请求
type BusinessDataRankProductReq struct {
	QueryStartTime    int64  `form:"query_start_time" json:"query_start_time"`       // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time" json:"query_end_time"`           // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	RankType          int    `form:"rank_type" json:"rank_type"`                     // 排行类型 (1 按销量, 2 按销售额)
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}

// BusinessDataCountProductSalesReq 营业数据商品销售统计列表请求
type BusinessDataCountProductSalesReq struct {
	dto.PageReq              // 分页参数
	ProductName       string `form:"product_name" json:"product_name"`               // 商品名称
	QueryStartTime    int64  `form:"query_start_time" json:"query_start_time"`       // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time" json:"query_end_time"`           // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	TimeType          int    `form:"time_type" json:"time_type"`                     // 时间类型: 1=今天, 2=昨天, 3=本周, 4=本月, 5=近7天, 6=上月, 7=今年
	AreaUuid          uint64 `form:"area_uuid" json:"area_uuid"`                     // 区域UUID, -1=全部
	CategoryUuid      uint64 `form:"category_uuid" json:"category_uuid"`             // 分类UUID, -1=全部 (向后兼容)
	CategoryUuids     string `form:"category_uuids" json:"category_uuids"`           // 分类UUID列表, 格式: "uuid1,uuid2,,,," 空字符串=全部
	OrderType         string `form:"order_type" json:"order_type"`                   // 订单类型: ""=全部, "1"=点餐, "2"=桌台, "3"=外送, 可多选如"1,2,3"
	OrderSource       int    `form:"order_source" json:"order_source"`               // 订单来源: -1=全部, 1=店内, 2=外卖
	SortType          int    `form:"sort_type" json:"sort_type"`                     // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection     int    `form:"sort_direction" json:"sort_direction"`           // 排序方向 0=默认、 1=升序、 2=降序
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}

// KitchenEfficiencyAnalysisReq 统计后厨效率分析请求
type KitchenEfficiencyAnalysisReq struct {
	dto.PageReq             // 分页参数
	StartTime      int64    `form:"start_time" json:"start_time"`             // 查询开始时间戳
	EndTime        int64    `form:"end_time" json:"end_time"`                 // 查询结束时间戳
	QueryStartDate string   `form:"query_start_date" json:"query_start_date"` // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate   string   `form:"query_end_date" json:"query_end_date"`     // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	Keyword        string   `form:"keyword" json:"keyword"`                   // 关键词, 仅商品名称模糊搜索
	CategoryUuids  []uint64 `form:"category_uuids" json:"category_uuids"`     // 分类UUID列表
}

// KitchenEfficiencyAnalysisAvgReq 统计后厨效率分析平均时长请求
type KitchenEfficiencyAnalysisAvgReq struct {
	StartTime      int64  `form:"start_time" json:"start_time"`             // 查询开始时间戳
	EndTime        int64  `form:"end_time" json:"end_time"`                 // 查询结束时间戳
	QueryStartDate string `form:"query_start_date" json:"query_start_date"` // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate   string `form:"query_end_date" json:"query_end_date"`     // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
}

// KitchenProductionDetailReq 统计后厨菜品出品明细请求
type KitchenProductionDetailReq struct {
	dto.PageReq             // 分页参数
	StartTime      int64    `form:"start_time" json:"start_time"`             // 查询开始时间戳
	EndTime        int64    `form:"end_time" json:"end_time"`                 // 查询结束时间戳
	QueryStartDate string   `form:"query_start_date" json:"query_start_date"` // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate   string   `form:"query_end_date" json:"query_end_date"`     // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	Keyword        string   `form:"keyword" json:"keyword"`                   // 关键词, 仅商品名称、内部编码模糊搜索
	CategoryUuids  []uint64 `form:"category_uuids" json:"category_uuids"`     // 分类UUID列表
}

// BusinessTimePeriodReq 统计营业时段请求
type BusinessTimePeriodReq struct {
	dto.PageReq              // 分页参数
	QueryStartTime    int64  `form:"query_start_time" json:"query_start_time"`       // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time" json:"query_end_time"`           // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	TimePeriod        int    `form:"time_period" json:"time_period"`                 // 时段， 0=默认、 1=15分钟、 2=半小时、 3=小时
	OrderInstant      int    `form:"order_instant" json:"order_instant"`             // 点餐订单， 0=否、 1=是
	OrderDesk         int    `form:"order_desk" json:"order_desk"`                   // 桌台订单， 0=否、 1=是
	OrderTakeout      int    `form:"order_takeout" json:"order_takeout"`             // 外送订单， 0=否、 1=是
	StatisticsType    int    `form:"statistics_type" json:"statistics_type"`         // 统计类型， 0=开台时间、 1=结账时间
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}

// StatisticsSummaryReq 统计综合运营请求
type StatisticsSummaryReq struct {
	dto.PageReq              // 分页参数
	QueryStartTime    int64  `form:"query_start_time" json:"query_start_time"`       // 查询开始时间戳
	QueryEndTime      int64  `form:"query_end_time" json:"query_end_time"`           // 查询结束时间戳
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	Cycle             int    `form:"cycle" json:"cycle"`                             // 周期: 0=按日、1=按月
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}

// StatisticsPaymentMethodReq 统计收款数据请求
type StatisticsPaymentMethodReq struct {
	dto.PageReq                 // 分页参数
	QueryStartTime     int64    `form:"query_start_time" json:"query_start_time"`         // 查询开始时间戳
	QueryEndTime       int64    `form:"query_end_time" json:"query_end_time"`             // 查询结束时间戳
	QueryStartDate     string   `form:"query_start_date" json:"query_start_date"`         // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate       string   `form:"query_end_date" json:"query_end_date"`             // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	Cycle              int      `form:"cycle" json:"cycle"`                               // 周期: 0=按日、1=按月
	OrderInstant       int      `form:"order_instant" json:"order_instant"`               // 点餐订单， 0=否、 1=是
	OrderDesk          int      `form:"order_desk" json:"order_desk"`                     // 桌台订单， 0=否、 1=是
	OrderTakeout       int      `form:"order_takeout" json:"order_takeout"`               // 外送订单， 0=否、 1=是
	PaymentMethodList  string   `form:"payment_method_list" json:"payment_method_list"`   // 支付方式UUID列表: 空=全部, 多个用"uuid1,uuid2,uuid3,,,"分割（优先使用）
	PaymentMethodNames []string `form:"payment_method_names" json:"payment_method_names"` // 支付方式名称列表: 空=全部（PaymentMethodList为空时使用）
	ExcludeDataManage  bool     `form:"exclude_data_manage" json:"exclude_data_manage"`   // 是否排除数据管理订单
}

// StatisticsCompanySummaryReq 门店汇总统计请求
type StatisticsCompanySummaryReq struct {
	dto.PageReq                 // 分页参数
	IndicatorType      string   `form:"indicator_type" json:"indicator_type"`             // 数据指标类型：business(营业数据汇总)、payment_method(支付方式汇总)、refund(退款金额汇总)
	CompanyUuids       []uint64 `form:"company_uuids" json:"company_uuids"`               // 门店UUID列表（多选），为空时默认为当前门店
	QueryStartDate     string   `form:"query_start_date" json:"query_start_date"`         // 开始日期（格式：YYYY-MM-DD HH:mm:ss），为空时默认为今日开始日期：YYYY-MM-DD 00:00:00
	QueryEndDate       string   `form:"query_end_date" json:"query_end_date"`             // 结束日期（格式：YYYY-MM-DD HH:mm:ss），为空时默认为今日结束日期：YYYY-MM-DD 23:59:59
	Cycle              int      `form:"cycle" json:"cycle"`                               // 周期: 0=按日、1=按月
	Report             int      `form:"report" json:"report"`                             // 报表类型: 0=明细表、1=汇总表
	PaymentMethodNames []string `form:"payment_method_names" json:"payment_method_names"` // 支付方式名称列表（仅支付方式汇总时使用，可选）
}
