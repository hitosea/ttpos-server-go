package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// FullReductionActivityCreateReq 创建满减活动请求
type FullReductionActivityCreateReq struct {
	LocaleName    dto.LocaleResponse                   `json:"locale_name" binding:"required"` // 多语言名称
	StartDate     int64                                `json:"start_date" binding:"required"`
	EndDate       int64                                `json:"end_date" binding:"required"`
	StartTime     string                               `json:"start_time"`     // HH:mm格式
	EndTime       string                               `json:"end_time"`       // HH:mm格式
	IsAllDay      int                                  `json:"is_all_day"`     // 1=全天，0=特定时段
	ReductionType int                                  `json:"reduction_type"` // 0=阶梯满减，1=循环满减
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// Validate 验证创建请求
func (req *FullReductionActivityCreateReq) Validate() error {
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}
	return nil
}

// FullReductionActivityRuleCreateReq 创建满减活动规则请求
type FullReductionActivityRuleCreateReq struct {
	Threshold       float64 `json:"threshold" binding:"required,min=0.01,max=999999.99"`
	ReductionAmount float64 `json:"reduction_amount" binding:"required,min=0.01,max=999999.99"`
}

// FullReductionActivityUpdateReq 更新满减活动请求
type FullReductionActivityUpdateReq struct {
	Uuid          uint64                               `json:"uuid" binding:"required"`
	LocaleName    dto.LocaleResponse                   `json:"locale_name" binding:"required"` // 多语言名称
	StartDate     int64                                `json:"start_date" binding:"required"`
	EndDate       int64                                `json:"end_date" binding:"required"`
	StartTime     string                               `json:"start_time"`
	EndTime       string                               `json:"end_time"`
	IsAllDay      int                                  `json:"is_all_day"`
	ReductionType int                                  `json:"reduction_type"`
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// Validate 验证更新请求
func (req *FullReductionActivityUpdateReq) Validate() error {
	if req.Uuid == 0 {
		return errors.New("UUID不能为空")
	}
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}
	return nil
}

// FullReductionActivityGetReq 获取满减活动请求
type FullReductionActivityGetReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"`
}

// FullReductionActivityListReq 获取满减活动列表请求
type FullReductionActivityListReq struct {
	dto.PageReq        // 分页参数
	Status      string `form:"status" json:"status"` // all, in_progress, not_start, end
}

// FullReductionActivityDeleteReq 删除满减活动请求
type FullReductionActivityDeleteReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"`
}

// FullReductionActivityDisableReq 失效满减活动请求
type FullReductionActivityDisableReq struct {
	Uuid uint64 `json:"uuid" binding:"required"`
}
