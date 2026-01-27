package resp

// UserAnalysisItem 用户分析统计项
type UserAnalysisItem struct {
	Name       string  `json:"name"`        // 名称（国籍名称/来源名称/用餐方式名称）
	OrderCount int64   `json:"order_count"` // 订单数
	Percentage float64 `json:"percentage"`  // 占比（%，保留2位小数，从 decimal 转换）
}

// UserAnalysisResp 用户分析统计响应
type UserAnalysisResp struct {
	Nationality   []UserAnalysisItem `json:"nationality"`    // 国籍统计
	OrderSource   []UserAnalysisItem `json:"order_source"`   // 点餐方式来源统计
	DeskSource    []UserAnalysisItem `json:"desk_source"`    // 桌台方式来源统计
	DiningMethod  []UserAnalysisItem `json:"dining_method"`  // 用餐方式统计
	TakeoutMethod []UserAnalysisItem `json:"takeout_method"` // 外卖方式统计（Grab/LINE MAN）
}
