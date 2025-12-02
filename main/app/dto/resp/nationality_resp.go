package resp

import "ttpos-server-go/app/dto"

// NationalityListResp 国籍列表响应
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type NationalityListResp struct {
	List []NationalityItem `json:"list"` // 国籍列表
}

// NationalityItem 国籍项
type NationalityItem struct {
	Uuid       uint64             `json:"uuid"`        // 国籍UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 多语言名称
	Sort       int                `json:"sort"`        // 排序
	Status     int                `json:"status"`      // 状态：1-启用；0-禁用
}

// NationalityCreateResp 创建国籍响应
type NationalityCreateResp struct {
	Uuid uint64 `json:"uuid"` // 国籍UUID
}
