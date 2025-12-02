package resp

import "ttpos-server-go/app/dto"

// FullReductionActivityResp 满减活动响应
type FullReductionActivityResp struct {
	Uuid              uint64                          `json:"uuid"`
	LocaleName        dto.LocaleResponse              `json:"locale_name"` // 多语言名称，使用 LocaleResponse（必须）
	StartDate         int64                           `json:"start_date"` // 时间戳，当天00:00:00
	EndDate           int64                           `json:"end_date"`  // 时间戳，当天23:59:59
	StartDateStr      string                          `json:"start_date_str"` // 格式2025-01-01
	EndDateStr        string                          `json:"end_date_str"`   // 格式2025-01-01
	StartTime         string                          `json:"start_time"`
	EndTime           string                          `json:"end_time"`
	IsAllDay          int                             `json:"is_all_day"`
	ReductionType     int                             `json:"reduction_type"`
	ReductionTypeName string                          `json:"reduction_type_name"` // 阶梯满减/循环满减
	IsDisabled        int                             `json:"is_disabled"`
	Status            string                          `json:"status"` // in_progress, not_start, end
	Rules             []FullReductionActivityRuleResp `json:"rules"`
}

// FullReductionActivityRuleResp 满减活动规则响应
type FullReductionActivityRuleResp struct {
	Uuid            uint64  `json:"uuid"`
	Threshold       float64 `json:"threshold"`
	ReductionAmount float64 `json:"reduction_amount"`
}

// FullReductionActivityListResp 满减活动列表响应
type FullReductionActivityListResp struct {
	List []*FullReductionActivityResp `json:"list"`
	Meta *PageMeta                    `json:"meta"`
}

// PageMeta 分页元数据
type PageMeta struct {
	PageNo   int   `json:"page_no"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
