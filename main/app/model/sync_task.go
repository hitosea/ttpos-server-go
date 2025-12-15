package model

// SyncTask 同步任务表 ttpos_sync_task
type SyncTask struct {
	BaseModel
	Status        uint8  `gorm:"column:status;type:int(10);default:0;comment:同步状态: 0-进行中, 1-已完成, 2-失败;NOT NULL" json:"status"`
	TotalCount    uint32 `gorm:"column:total_count;type:int(10);default:0;comment:总任务数;NOT NULL" json:"total_count"`
	SuccessCount  uint32 `gorm:"column:success_count;type:int(10);default:0;comment:成功任务数;NOT NULL" json:"success_count"`
	FailCount     uint32 `gorm:"column:fail_count;type:int(10);default:0;comment:失败任务数;NOT NULL" json:"fail_count"`
	Panic         string `gorm:"column:panic;type:text;comment:panic错误信息" json:"panic"`
	StartTime     int64  `gorm:"column:start_time;type:int(10);default:0;comment:开始时间;NOT NULL" json:"start_time"`
	EndTime       int64  `gorm:"column:end_time;type:int(10);default:0;comment:结束时间;NOT NULL" json:"end_time"`
	RequestParams string `gorm:"column:request_params;type:text;comment:请求参数(JSON格式)" json:"request_params"`

	// 关联
	Items []SyncTaskItem `gorm:"foreignKey:SyncTaskUuid;references:Uuid" json:"items,omitempty"`
}

// SyncTaskItem 同步任务明细表 ttpos_sync_task_item
type SyncTaskItem struct {
	BaseModel
	SyncTaskUuid uint64 `gorm:"column:sync_task_uuid;type:bigint;default:0;comment:同步任务UUID;NOT NULL" json:"sync_task_uuid"`
	TaskType     string `gorm:"column:task_type;type:varchar(50);default:'';comment:任务类型;NOT NULL" json:"task_type"`
	TaskName     string `gorm:"column:task_name;type:varchar(100);default:'';comment:任务名称;NOT NULL" json:"task_name"`
	Status       uint8  `gorm:"column:status;type:int(10);default:0;comment:任务状态: 0-待执行, 1-执行中, 2-已完成, 3-失败;NOT NULL" json:"status"`
	ErrorMessage string `gorm:"column:error_message;type:text;comment:错误消息" json:"error_message"`
	StartTime    int64  `gorm:"column:start_time;type:int(10);default:0;comment:开始时间;NOT NULL" json:"start_time"`
	EndTime      int64  `gorm:"column:end_time;type:int(10);default:0;comment:结束时间;NOT NULL" json:"end_time"`
}
