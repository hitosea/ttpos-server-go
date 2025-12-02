package req

import (
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
)

// FullReductionActivityCreateReq 创建满减活动请求
type FullReductionActivityCreateReq struct {
	LocaleName    dto.LocaleResponse                   `json:"locale_name" binding:"required"` // 多语言名称
	StartDate     int64                                `json:"start_date"`                     // 时间戳，当天00:00:00
	EndDate       int64                                `json:"end_date"`                       // 时间戳，当天23:59:59
	StartDateStr  string                               `json:"start_date_str"`                 // 格式2025-01-01
	EndDateStr    string                               `json:"end_date_str" `                  // 格式2025-01-01
	StartTime     string                               `json:"start_time"`                     // HH:mm格式
	EndTime       string                               `json:"end_time"`                       // HH:mm格式
	IsAllDay      int                                  `json:"is_all_day"`                     // 1=全天，0=特定时段
	ReductionType int                                  `json:"reduction_type"`                 // 0=阶梯满减，1=循环满减
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// initDatesFromStrings 从日期字符串初始化日期字段
// timezone: 商家所在时区，如 "Asia/Shanghai"，如果为空则使用默认时区
// startDate: 指向开始日期字段的指针
// endDate: 指向结束日期字段的指针
// startDateStr: 开始日期字符串，格式为 YYYY-MM-DD
// endDateStr: 结束日期字符串，格式为 YYYY-MM-DD
func initDatesFromStrings(timezone string, startDate *int64, endDate *int64, startDateStr, endDateStr string) error {
	// 使用默认时区或传入的时区
	tz := timezone
	if tz == "" {
		tz = string(utils.ZH_TIMEZONE)
	}
	timeUtil := utils.Timezone(tz)

	// 如果 StartDate 为 0，则从 StartDateStr 解析
	if *startDate == 0 && startDateStr != "" {
		// 解析日期字符串，获取当天的 00:00:00 时间戳
		startTimestamp, err := timeUtil.FormatTimeToUnix(startDateStr)
		if err != nil {
			return errors.WithMessage(errors.New("开始日期格式错误，应为 YYYY-MM-DD 格式"), err.Error())
		}
		*startDate = startTimestamp
	}

	// 如果 EndDate 为 0，则从 EndDateStr 解析
	if *endDate == 0 && endDateStr != "" {
		// 解析日期字符串，获取当天的 00:00:00 时间戳
		endTimestamp, err := timeUtil.FormatTimeToUnix(endDateStr)
		if err != nil {
			return errors.WithMessage(errors.New("结束日期格式错误，应为 YYYY-MM-DD 格式"), err.Error())
		}
		// 将结束日期设置为当天的 23:59:59
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.Local
		}
		endTime := time.Unix(endTimestamp, 0).In(loc)
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, loc)
		*endDate = endTime.Unix()
	}

	return nil
}

// Validate 验证创建请求
// timezone: 商家所在时区，如 "Asia/Shanghai"，如果为空则使用默认时区
func (req *FullReductionActivityCreateReq) Validate(timezone string) error {
	if req.LocaleName.IsNull() {
		return errors.New("多语言名称不能为空")
	}
	// 至少需要中文或英文名称
	if req.LocaleName.ZH == "" && req.LocaleName.EN == "" {
		return errors.New("至少需要提供中文或英文名称")
	}

	// 初始化日期字段
	if err := initDatesFromStrings(timezone, &req.StartDate, &req.EndDate, req.StartDateStr, req.EndDateStr); err != nil {
		return err
	}
	// 验证日期字段是否已初始化
	if req.StartDate == 0 {
		return errors.New("开始日期不能为空，请提供 start_date 或 start_date_str")
	}
	if req.EndDate == 0 {
		return errors.New("结束日期不能为空，请提供 end_date 或 end_date_str")
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
	StartDate     int64                                `json:"start_date"`                     // 时间戳，当天00:00:00
	EndDate       int64                                `json:"end_date"`                       // 时间戳，当天23:59:59
	StartDateStr  string                               `json:"start_date_str"`                 // 格式2025-01-01
	EndDateStr    string                               `json:"end_date_str"`                   // 格式2025-01-01
	StartTime     string                               `json:"start_time"`
	EndTime       string                               `json:"end_time"`
	IsAllDay      int                                  `json:"is_all_day"`
	ReductionType int                                  `json:"reduction_type"`
	Rules         []FullReductionActivityRuleCreateReq `json:"rules" binding:"required,min=1"`
}

// Validate 验证更新请求
// timezone: 商家所在时区，如 "Asia/Shanghai"，如果为空则使用默认时区
func (req *FullReductionActivityUpdateReq) Validate(timezone string) error {
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

	// 初始化日期字段
	if err := initDatesFromStrings(timezone, &req.StartDate, &req.EndDate, req.StartDateStr, req.EndDateStr); err != nil {
		return err
	}

	// 验证日期字段是否已初始化
	if req.StartDate == 0 {
		return errors.New("开始日期不能为空，请提供 start_date 或 start_date_str")
	}
	if req.EndDate == 0 {
		return errors.New("结束日期不能为空，请提供 end_date 或 end_date_str")
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
