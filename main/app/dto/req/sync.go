package req

import "ttpos-server-go/app/dto"

// SyncReq 同步请求
type SyncReq struct {
	TaskUuid uint64 `json:"task_uuid" form:"task_uuid"` // 任务UUID，如果传递则重新执行该任务
}

// SyncTaskListReq 同步任务列表请求
type SyncTaskListReq struct {
	dto.PageReq
	Status *uint8 `json:"status" form:"status"` // 同步状态: 0-进行中, 1-已完成, 2-失败
}

// SyncTaskDetailReq 同步任务详情请求
type SyncTaskDetailReq struct {
	TaskUuid uint64 `json:"task_uuid" form:"task_uuid" binding:"required"` // 任务UUID
}
