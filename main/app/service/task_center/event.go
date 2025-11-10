package task_center

type EventType string

const (
	EventTypeTaskCreate EventType = "task_create" // 新增任务
	EventTypeTaskUpdate EventType = "task_update" // 修改任务
	EventTypeTaskDelete EventType = "task_delete" // 删除任务
)
