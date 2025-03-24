package req

import "ttpos-server-go/app/dto"

type BusinessDataPrinterReq struct {
	TimeType       int  `json:"time_type"`        // 时间类型 (1 按日, 2 昨天, 3 本周, 4 本月)
	StatisticsType int  `json:"statistics_type"`  // 统计类型 (0 全部, 1 按支付方式, 2 按商品分类, 3 按商品)
	QueryStartTime uint `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   uint `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int  `json:"category_type"`    // 分类类型 (1 按一级分类, 2 按二级分类)
}

// BusinessDataCountReq 营业数据统计请求
type BusinessDataCountReq struct {
	TimeType       int   `json:"time_type"`        // 时间类型 (1 按日, 2 昨天, 3 本周, 4 本月)
	QueryStartTime int64 `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64 `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int   `json:"category_type"`    // 分类类型 (1 按一级分类, 2 按二级分类)
}

// BusinessDataRankProductReq 营业数据排行请求
type BusinessDataRankProductReq struct {
	QueryStartTime int64 `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64 `json:"query_end_time"`   // 查询结束时间戳
	RankType       int   `json:"rank_type"`        // 排行类型 (1 按销量, 2 按销售额)
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
