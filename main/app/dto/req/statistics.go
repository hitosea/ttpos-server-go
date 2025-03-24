package req

type BusinessDataPrinterReq struct {
	TimeType       int  `json:"time_type"`        // 时间类型 (1 按日, 2 昨天, 3 本周, 4 本月)
	StatisticsType int  `json:"statistics_type"`  // 统计类型 (0 全部, 1 按支付方式, 2 按商品分类, 3 按商品)
	QueryStartTime uint `json:"query_start_time"` // 查询开始时间戳
	QueryEndTime   uint `json:"query_end_time"`   // 查询结束时间戳
	CategoryType   int  `json:"category_type"`    // 分类类型 (1 按一级分类, 2 按二级分类)
}
