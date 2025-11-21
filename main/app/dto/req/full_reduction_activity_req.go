package req

// FullReductionActivityCreateReq 创建满减活动请求
type FullReductionActivityCreateReq struct {
	Name          string                          `json:"name" binding:"required"` // JSON格式多语言名称，前端发送 JSON 字符串
	StartDate     int64                           `json:"start_date" binding:"required"`
	EndDate       int64                           `json:"end_date" binding:"required"`
	StartTime     string                          `json:"start_time"` // HH:mm格式
	EndTime       string                          `json:"end_time"`   // HH:mm格式
	IsAllDay      int                             `json:"is_all_day" binding:"required"` // 1=全天，0=特定时段
	ReductionType int                             `json:"reduction_type" binding:"required"` // 0=阶梯满减，1=循环满减
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// FullReductionActivityRuleCreateReq 创建满减活动规则请求
type FullReductionActivityRuleCreateReq struct {
	Threshold       float64 `json:"threshold" binding:"required,min=0.01,max=999999.99"`
	ReductionAmount float64 `json:"reduction_amount" binding:"required,min=0.01,max=999999.99"`
}

// FullReductionActivityUpdateReq 更新满减活动请求
type FullReductionActivityUpdateReq struct {
	Uuid          uint64                          `json:"uuid" binding:"required"`
	Name          string                          `json:"name" binding:"required"` // JSON格式多语言名称，前端发送 JSON 字符串
	StartDate     int64                           `json:"start_date" binding:"required"`
	EndDate       int64                           `json:"end_date" binding:"required"`
	StartTime     string                          `json:"start_time"`
	EndTime       string                          `json:"end_time"`
	IsAllDay      int                             `json:"is_all_day" binding:"required"`
	ReductionType int                             `json:"reduction_type" binding:"required"`
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// FullReductionActivityGetReq 获取满减活动请求
type FullReductionActivityGetReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

// FullReductionActivityListReq 获取满减活动列表请求
type FullReductionActivityListReq struct {
	PageNo   int    `json:"page_no" binding:"required"`
	PageSize int    `json:"page_size" binding:"required"`
	Status   string `json:"status"` // all, ongoing, not_started, ended
}

// FullReductionActivityDeleteReq 删除满减活动请求
type FullReductionActivityDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

// FullReductionActivityDisableReq 失效满减活动请求
type FullReductionActivityDisableReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}

