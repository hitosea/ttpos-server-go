package resp

import "ttpos-server-go/app/dto"

// BatchRegenerateTaskResp 批量重新生成任务清单响应
type BatchRegenerateTaskResp struct {
	TaskFile string          `json:"task_file"`
	Summary  dto.TaskSummary `json:"summary"`
}

// BatchRegenerateProgressResp 批量重新生成进度响应
type BatchRegenerateProgressResp struct {
	OverallProgress   float64         `json:"overall_progress"` // 总体完成百分比（0-100）
	CompletedSteps    int             `json:"completed_steps"`
	FailedSteps       int             `json:"failed_steps"`
	PendingSteps      int             `json:"pending_steps"`
	EstimatedTimeLeft string          `json:"estimated_time_left"` // 预计剩余时间，格式：约 X 分钟
	CurrentStep       CurrentStep     `json:"current_step"`        // 当前正在执行的步骤
	CompanyProgress   CompanyProgress `json:"company_progress"`
	DateProgress      DateProgress    `json:"date_progress"`
	OrderProgress     OrderProgress   `json:"order_progress"`
}

// CurrentStep 当前正在执行的步骤
type CurrentStep struct {
	StepName      string `json:"step_name"`       // 步骤名称，如：regenerate-order-material
	StepType      string `json:"step_type"`       // 步骤类型：date（日期级别）或 order（订单级别）
	CompanyName   string `json:"company_name"`    // 公司名称
	Date          string `json:"date"`            // 日期（YYYY-MM-DD），仅订单级别步骤有值
	OrderNo       string `json:"order_no"`        // 订单号，仅订单级别步骤有值
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 订单UUID，仅订单级别步骤有值
}

// CompanyProgress 公司级别进度
type CompanyProgress struct {
	TotalCompanies     int      `json:"total_companies"`
	PendingCompanies   int      `json:"pending_companies"`
	CompletedCompanies []string `json:"completed_companies"`
	CurrentCompany     string   `json:"current_company"`
}

// DateProgress 日期级别进度
type DateProgress struct {
	TotalDates     int      `json:"total_dates"`
	PendingDates   int      `json:"pending_dates"`
	CompletedDates []string `json:"completed_dates"`
	CurrentDate    string   `json:"current_date"`
}

// OrderProgress 订单级别进度
type OrderProgress struct {
	TotalOrders     int      `json:"total_orders"`
	PendingOrders   int      `json:"pending_orders"`
	CompletedOrders []string `json:"completed_orders"`
	CurrentOrder    string   `json:"current_order"`
}
