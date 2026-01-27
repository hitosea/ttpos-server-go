package dto

// BatchRegenerateTask 批量重新生成任务清单
type BatchRegenerateTask struct {
	Companies []CompanyTask `json:"companies"`
	Summary   TaskSummary   `json:"summary"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// CompanyTask 公司任务
type CompanyTask struct {
	CompanyUuid uint64     `json:"company_uuid"`
	CompanyName string     `json:"company_name"`
	Dates       []DateTask `json:"dates"`
}

// DateTask 日期任务
type DateTask struct {
	Date     string      `json:"date"`
	DateStep StepTask    `json:"date_step"`
	Orders   []OrderTask `json:"orders"`
}

// OrderTask 订单任务
type OrderTask struct {
	SaleOrderUuid uint64     `json:"sale_order_uuid"`
	OrderNo       string     `json:"order_no"`
	OrderDate     string     `json:"order_date"`
	Steps         []StepTask `json:"steps"`
}

// StepTask 步骤任务
type StepTask struct {
	Step      int    `json:"step"`
	Name      string `json:"name"`
	Status    string `json:"status"`     // pending, running, completed, failed
	StartTime string `json:"start_time"` // RFC3339格式，可为空字符串
	EndTime   string `json:"end_time"`   // RFC3339格式，可为空字符串
	Error     string `json:"error"`      // 错误信息，可为空字符串
}

// TaskSummary 任务统计信息
type TaskSummary struct {
	TotalCompanies  int `json:"total_companies"`
	TotalDates      int `json:"total_dates"`
	TotalOrders     int `json:"total_orders"`
	TotalDateSteps  int `json:"total_date_steps"`
	TotalOrderSteps int `json:"total_order_steps"`
	CompletedSteps  int `json:"completed_steps"`
	FailedSteps     int `json:"failed_steps"`
	PendingSteps    int `json:"pending_steps"`
}

// StepStatus 步骤状态常量
const (
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusCompleted = "completed"
	StepStatusFailed    = "failed"
)

// IsValidStatus 检查状态值是否有效
func IsValidStatus(status string) bool {
	return status == StepStatusPending ||
		status == StepStatusRunning ||
		status == StepStatusCompleted ||
		status == StepStatusFailed
}
