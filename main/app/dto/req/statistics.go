package req

import (
	"time"
	"ttpos-server-go/app/dto"
)

type BusinessDataPrinterReq struct {
	TimeType       int  `json:"time_type"`        // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月)
	StatisticsType int  `json:"statistics_type"`  // 统计类型 (0 全部, 1 按支付方式, 2 按商品分类, 3 按商品)
	QueryStartTime uint `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   uint `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int  `json:"category_type"`    // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
}

// GetParam 获取参数
func (r *BusinessDataPrinterReq) GetParam() BusinessDataPrinterReq {
	var (
		queryStartTime int64
		queryEndTime   int64
	)
	// 处理时间范围
	if r.TimeType > 0 && r.TimeType < 5 {
		now := time.Now()
		var startTime, endTime time.Time
		switch r.TimeType {
		case 1: // 今天
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endTime = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		case 2: // 昨天
			startTime = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
			endTime = time.Date(now.Year(), now.Month(), now.Day()-1, 23, 59, 59, 0, now.Location())
		case 3: // 本周
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startTime = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
			endTime = startTime.AddDate(0, 0, 7).Add(-time.Second)
		case 4: // 本月
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endTime = startTime.AddDate(0, 1, 0).Add(-time.Second)
		}
		queryStartTime = startTime.Unix()
		queryEndTime = endTime.Unix()
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
	}
}

// BusinessDataCountReq 营业数据统计请求
type BusinessDataCountReq struct {
	TimeType       int    `form:"time_type"`        // 时间类型 (-1 未选择, 1 今天, 2 昨天, 3 本周, 4 本月)
	QueryStartTime int64  `form:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time"`   // 查询结束时间戳
	CategoryType   int    `form:"category_type"`    // 分类类型 (-1 未选择, 1 按一级分类, 2 按二级分类)
	DutyNo         string `form:"duty_no"`          // 班次编号
}

// BusinessDataRankProductReq 营业数据排行请求
type BusinessDataRankProductReq struct {
	QueryStartTime int64 `form:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64 `form:"query_end_time"`   // 查询结束时间戳
	RankType       int   `form:"rank_type"`        // 排行类型 (1 按销量, 2 按销售额)
}

// BusinessDataCountProductSalesReq 营业数据商品销售统计列表请求
type BusinessDataCountProductSalesReq struct {
	dto.PageReq           // 分页参数
	ProductName    string `form:"product_name"`     // 商品名称
	QueryStartTime int64  `form:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time"`   // 查询结束时间戳
	AreaUuid       uint64 `form:"area_uuid"`        // 区域UUID, -1=全都
	CategoryUuid   uint64 `form:"category_uuid"`    // 分类UUID, -1=全都
	SortType       int    `form:"sort_type"`        // 排序类型 0=默认、 1=按销售数量、 2=按原销售额
	SortDirection  int    `form:"sort_direction"`   // 排序方向 0=默认、 1=升序、 2=降序
}
