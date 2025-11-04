package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"
)

type BusinessDataPrinterReq struct {
	TimeType       int  `json:"time_type"`        // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月, 5 营业时间)
	StatisticsType int  `json:"statistics_type"`  // 统计类型 (0 全部, 1 按支付方式, 2 按商品分类, 3 按商品)
	QueryStartTime uint `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   uint `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int  `json:"category_type"`    // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
	NotQueryFree   bool `json:"not_query_free"`   // 是否不查询免费使用场景
	SortType       int  `json:"sort_type"`        // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection  int  `json:"sort_direction"`   // 排序方向 0=默认、 1=升序、 2=降序
}

// GetParam 获取参数
func (r *BusinessDataPrinterReq) GetParam(timezone string, openingHours string) BusinessDataPrinterReq {
	var (
		queryStartTime int64
		queryEndTime   int64
	)
	// 处理时间范围
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
	if r.QueryStartTime > 0 && r.QueryEndTime > 0 {
		queryStartTime = int64(r.QueryStartTime)
		queryEndTime = int64(r.QueryEndTime)
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
	TimeType       int    `form:"time_type" json:"time_type"`               // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月, 5 营业时间)
	QueryStartTime int64  `form:"query_start_time" json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time" json:"query_end_time"`     // 查询结束时间戳
	CategoryType   int    `form:"category_type" json:"category_type"`       // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
	DutyNo         string `form:"duty_no" json:"duty_no"`                   // 班次编号
	NotQueryFree   bool   `form:"not_query_free" json:"not_query_free"`     // 是否不查询免费使用场景
	StaffUuid      uint64 `form:"staff_uuid" json:"staff_uuid"`             // 操作员UUID
	SortType       int    `form:"sort_type" json:"sort_type"`               // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection  int    `form:"sort_direction" json:"sort_direction"`     // 排序方向 0=默认、 1=升序、 2=降序
}

// GetParam 获取参数
func (r *BusinessDataCountReq) GetParam(timezone string, openingHours string) BusinessDataCountReq {
	var (
		queryStartTime int64
		queryEndTime   int64
	)
	// 处理时间范围
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
	if r.QueryStartTime > 0 && r.QueryEndTime > 0 {
		queryStartTime = int64(r.QueryStartTime)
		queryEndTime = int64(r.QueryEndTime)
	}
	return BusinessDataCountReq{
		TimeType:       r.TimeType,
		QueryStartTime: queryStartTime,
		QueryEndTime:   queryEndTime,
		CategoryType:   r.CategoryType,
		DutyNo:         r.DutyNo,
		NotQueryFree:   r.NotQueryFree,
		SortType:       r.SortType,
		SortDirection:  r.SortDirection,
	}
}

// BusinessDataRankProductReq 营业数据排行请求
type BusinessDataRankProductReq struct {
	QueryStartTime int64 `form:"query_start_time" json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64 `form:"query_end_time" json:"query_end_time"`     // 查询结束时间戳
	RankType       int   `form:"rank_type" json:"rank_type"`               // 排行类型 (1 按销量, 2 按销售额)
}

// BusinessDataCountProductSalesReq 营业数据商品销售统计列表请求
type BusinessDataCountProductSalesReq struct {
	dto.PageReq           // 分页参数
	ProductName    string `form:"product_name" json:"product_name"`         // 商品名称
	QueryStartTime int64  `form:"query_start_time" json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time" json:"query_end_time"`     // 查询结束时间戳
	AreaUuid       uint64 `form:"area_uuid" json:"area_uuid"`               // 区域UUID, -1=全都
	CategoryUuid   uint64 `form:"category_uuid" json:"category_uuid"`       // 分类UUID, -1=全都
	SortType       int    `form:"sort_type" json:"sort_type"`               // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection  int    `form:"sort_direction" json:"sort_direction"`     // 排序方向 0=默认、 1=升序、 2=降序
}

// KitchenEfficiencyAnalysisReq 统计后厨效率分析请求
type KitchenEfficiencyAnalysisReq struct {
	StartTime     int64    `form:"start_time" json:"start_time"`         // 查询开始时间戳
	EndTime       int64    `form:"end_time" json:"end_time"`             // 查询结束时间戳
	Keyword       string   `form:"keyword" json:"keyword"`               // 关键词, 仅商品名称模糊搜索
	CategoryUuids []uint64 `form:"category_uuids" json:"category_uuids"` // 分类UUID列表
}

// KitchenEfficiencyAnalysisAvgReq 统计后厨效率分析平均时长请求
type KitchenEfficiencyAnalysisAvgReq struct {
	StartTime int64 `form:"start_time" json:"start_time"` // 查询开始时间戳
	EndTime   int64 `form:"end_time" json:"end_time"`     // 查询结束时间戳
}
