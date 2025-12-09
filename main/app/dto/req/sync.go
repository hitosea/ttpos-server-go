package req

import "ttpos-server-go/app/dto"

// SyncReq 同步请求
type SyncReq struct {
	TaskUuid      uint64 `json:"task_uuid" form:"task_uuid"`             // 任务UUID，如果传递则重新执行该任务
	IsSyncExecute bool   `json:"is_sync_execute" form:"is_sync_execute"` // 是否同步执行
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

// GetHeadquartersDataListReq 获取总部可同步数据列表请求
type GetHeadquartersDataListReq struct {
	DataTypes []string `json:"data_types"` // 可选，指定查询的数据类型，不传则查询所有
}

// GranularSyncReq 颗粒化同步请求（简化：只需要提交需要同步的分组）
type GranularSyncReq struct {
	ProductDataChecked  bool `json:"product_data_checked"`  // 商品数据组是否勾选
	ActivityDataChecked bool `json:"activity_data_checked"` // 活动数据组是否勾选
	PaymentDataChecked  bool `json:"payment_data_checked"`  // 支付数据组是否勾选
}
