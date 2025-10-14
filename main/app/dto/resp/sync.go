package resp

import "ttpos-server-go/app/dto"

// SyncResp 同步响应
type SyncResp struct {
	TaskUuid uint64 `json:"task_uuid"` // 任务UUID
	Message  string `json:"message"`   // 提示消息
}

// SyncTaskItemResp 同步任务明细响应
type SyncTaskItemResp struct {
	Uuid         uint64 `json:"uuid"`          // UUID
	TaskType     string `json:"task_type"`     // 任务类型
	TaskName     string `json:"task_name"`     // 任务名称
	Status       uint8  `json:"status"`        // 任务状态: 0-待执行, 1-执行中, 2-已完成, 3-失败
	ErrorMessage string `json:"error_message"` // 错误消息
	StartTime    int64  `json:"start_time"`    // 开始时间
	EndTime      int64  `json:"end_time"`      // 结束时间
	Duration     int64  `json:"duration"`      // 执行时长（秒）
}

// SyncTaskDetailResp 同步任务详情响应
type SyncTaskDetailResp struct {
	Uuid         uint64             `json:"uuid"`          // UUID
	Status       uint8              `json:"status"`        // 同步状态: 0-进行中, 1-已完成, 2-失败
	TotalCount   uint32             `json:"total_count"`   // 总任务数
	SuccessCount uint32             `json:"success_count"` // 成功任务数
	FailCount    uint32             `json:"fail_count"`    // 失败任务数
	StartTime    int64              `json:"start_time"`    // 开始时间
	EndTime      int64              `json:"end_time"`      // 结束时间
	Duration     int64              `json:"duration"`      // 执行时长（秒）
	CreateTime   int64              `json:"create_time"`   // 创建时间
	Items        []SyncTaskItemResp `json:"items"`         // 任务明细
}

// SyncTaskListResp 同步任务列表响应
type SyncTaskListResp struct {
	Uuid         uint64 `json:"uuid"`          // UUID
	Status       uint8  `json:"status"`        // 同步状态: 0-进行中, 1-已完成, 2-失败
	TotalCount   uint32 `json:"total_count"`   // 总任务数
	SuccessCount uint32 `json:"success_count"` // 成功任务数
	FailCount    uint32 `json:"fail_count"`    // 失败任务数
	StartTime    int64  `json:"start_time"`    // 开始时间
	EndTime      int64  `json:"end_time"`      // 结束时间
	Duration     int64  `json:"duration"`      // 执行时长（秒）
	CreateTime   int64  `json:"create_time"`   // 创建时间
}

// SyncTaskListPaginationResp 同步任务列表分页响应
type SyncTaskListPaginationResp struct {
	List []SyncTaskListResp `json:"list"`
	Meta dto.PageResponse   `json:"meta"`
}
