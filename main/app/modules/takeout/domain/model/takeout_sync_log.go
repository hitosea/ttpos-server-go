package model

// TakeoutSyncLog 外卖订单同步日志表（多平台）
type TakeoutSyncLog struct {
	BaseModel
	Platform string `gorm:"column:platform" json:"platform"`

	// 同步信息
	PlatformOrderId string `gorm:"column:platform_order_id" json:"platform_order_id"`
	SyncType        string `gorm:"column:sync_type" json:"sync_type"`
	SyncStatus      int    `gorm:"column:sync_status" json:"sync_status"`
	RetryCount      int    `gorm:"column:retry_count" json:"retry_count"`

	// 错误信息
	ErrorCode    string `gorm:"column:error_code" json:"error_code"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	// 请求/响应数据
	RequestData  string `gorm:"column:request_data;type:mediumtext" json:"request_data"`
	ResponseData string `gorm:"column:response_data;type:mediumtext" json:"response_data"`
}

func (*TakeoutSyncLog) TableName() string {
	return "ttpos_takeout_sync_logs"
}
