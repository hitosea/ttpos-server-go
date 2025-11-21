package resp

import "ttpos-server-go/app/dto"

// OrderSourceListResp 外卖来源列表响应
//
// 任务: story-order-source-nationality Phase 2.4
// 需求: R1.1-R1.6
//
// @version v2.10.0
type OrderSourceListResp struct {
	List []OrderSourceItem `json:"list"` // 外卖来源列表
}

// OrderSourceItem 外卖来源项
type OrderSourceItem struct {
	Uuid       uint64             `json:"uuid"`        // 外卖来源UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 多语言名称
	Sort       int                `json:"sort"`        // 排序
	Status     int                `json:"status"`      // 状态：1-启用；0-禁用
}

// OrderSourceCreateResp 创建外卖来源响应
type OrderSourceCreateResp struct {
	Uuid uint64 `json:"uuid"` // 外卖来源UUID
}
