package req

// ChannelSalesReq 渠道营业统计请求
type ChannelSalesReq struct {
	StartTime int64 `form:"start_time" json:"start_time"` // 查询开始时间戳（Unix 秒）
	EndTime   int64 `form:"end_time" json:"end_time"`     // 查询结束时间戳（Unix 秒）
}
