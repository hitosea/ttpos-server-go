package req

// UserAnalysisReq 用户分析统计请求
type UserAnalysisReq struct {
	StartTime         int64  `form:"start_time" json:"start_time"`                   // 查询开始时间戳（Unix 秒），可选，未传时默认今天开始时间
	EndTime           int64  `form:"end_time" json:"end_time"`                       // 查询结束时间戳（Unix 秒），可选，未传时默认今天结束时间
	QueryStartDate    string `form:"query_start_date" json:"query_start_date"`       // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate      string `form:"query_end_date" json:"query_end_date"`           // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	ExcludeDataManage bool   `form:"exclude_data_manage" json:"exclude_data_manage"` // 是否排除数据管理订单
}
