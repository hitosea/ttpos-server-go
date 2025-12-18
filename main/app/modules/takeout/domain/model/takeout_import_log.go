package model

import (
	"time"
)

// TakeoutImportLog 外卖导入日志实体
// 记录所有导入操作的历史记录，支持 UI 展示和追溯
type TakeoutImportLog struct {
	// 基础字段
	ID   uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID uint64 `gorm:"uniqueIndex:uk_uuid;not null;default:0" json:"uuid"`

	// 导入信息字段
	Platform        string `gorm:"type:varchar(50);not null;default:'';index:idx_platform" json:"platform"`     // 外卖平台(grab/lineman等)
	ImportType      int8   `gorm:"type:tinyint(3);not null;default:0;index:idx_import_type" json:"import_type"` // 导入类型(1-TTPOS推送到平台 2-平台推送到TTPOS)
	ImportDirection string `gorm:"type:varchar(200);not null;default:''" json:"import_direction"`               // 导入方向描述

	// 状态和进度字段
	Status   int8 `gorm:"type:tinyint(3);not null;default:0;index:idx_status" json:"status"` // 导入状态(0-进行中 1-成功 2-失败)
	Progress int  `gorm:"type:int(10);not null;default:0" json:"progress"`                   // 进度百分比(0-100)

	// 统计信息字段
	SuccessCount int    `gorm:"type:int(10);not null;default:0" json:"success_count"` // 成功数量
	FailureCount int    `gorm:"type:int(10);not null;default:0" json:"failure_count"` // 失败数量
	TotalCount   int    `gorm:"type:int(10);not null;default:0" json:"total_count"`   // 总数量
	ErrorMessage string `gorm:"type:text" json:"error_message"`                       // 错误信息

	// 时间字段
	StartTime  int64 `gorm:"type:int(10);not null;default:0" json:"start_time"`                       // 开始时间
	EndTime    int64 `gorm:"type:int(10);not null;default:0" json:"end_time"`                         // 结束时间
	Duration   int64 `gorm:"type:int(10);not null;default:0" json:"duration"`                         // 耗时(秒)
	CreateTime int64 `gorm:"type:int(10);not null;default:0;index:idx_create_time" json:"createtime"` // 创建时间
	UpdateTime int64 `gorm:"type:int(10);not null;default:0" json:"updatetime"`                       // 更新时间
	DeleteTime int64 `gorm:"type:int(10);not null;default:0;index:idx_delete_time" json:"deletetime"` // 删除时间
}

// TableName 表名
func (TakeoutImportLog) TableName() string {
	return "ttpos_takeout_import_log"
}

// ImportType 导入类型常量
const (
	ImportTypeTTPOSToPlatform = 1 // TTPOS推送到平台
	ImportTypePlatformToTTPOS = 2 // 平台推送到TTPOS
)

// ImportStatus 导入状态常量
const (
	ImportLogStatusInProgress = 0 // 进行中
	ImportLogStatusSuccess    = 1 // 导入成功
	ImportLogStatusFailed     = 2 // 导入失败
)

// IsInProgress 是否进行中
func (l *TakeoutImportLog) IsInProgress() bool {
	return l.Status == ImportLogStatusInProgress
}

// IsSuccess 是否成功
func (l *TakeoutImportLog) IsSuccess() bool {
	return l.Status == ImportLogStatusSuccess
}

// IsFailed 是否失败
func (l *TakeoutImportLog) IsFailed() bool {
	return l.Status == ImportLogStatusFailed
}

// MarkSuccess 标记为成功
func (l *TakeoutImportLog) MarkSuccess() {
	l.Status = ImportLogStatusSuccess
	l.Progress = 100
	l.EndTime = time.Now().Unix()
	l.Duration = l.EndTime - l.StartTime
	l.UpdateTime = time.Now().Unix()
}

// MarkFailed 标记为失败
func (l *TakeoutImportLog) MarkFailed(errorMsg string) {
	l.Status = ImportLogStatusFailed
	l.ErrorMessage = errorMsg
	l.EndTime = time.Now().Unix()
	l.Duration = l.EndTime - l.StartTime
	l.UpdateTime = time.Now().Unix()
}

// UpdateProgress 更新进度
func (l *TakeoutImportLog) UpdateProgress(current, total int) {
	if total > 0 {
		l.Progress = (current * 100) / total
	}
	l.UpdateTime = time.Now().Unix()
}

// IncrementSuccess 增加成功计数
func (l *TakeoutImportLog) IncrementSuccess() {
	l.SuccessCount++
	l.UpdateTime = time.Now().Unix()
}

// IncrementFailure 增加失败计数
func (l *TakeoutImportLog) IncrementFailure() {
	l.FailureCount++
	l.UpdateTime = time.Now().Unix()
}

// SetTotalCount 设置总数量
func (l *TakeoutImportLog) SetTotalCount(count int) {
	l.TotalCount = count
	l.UpdateTime = time.Now().Unix()
}
