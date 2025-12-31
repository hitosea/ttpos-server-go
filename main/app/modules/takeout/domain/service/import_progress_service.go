package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/repository"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// ImportProgressService 导入进度管理服务
type ImportProgressService struct {
	importLogRepo repository.TakeoutImportLogRepository
}

// NewImportProgressService 创建导入进度管理服务实例
func NewImportProgressService(db *gorm.DB) *ImportProgressService {
	return &ImportProgressService{
		importLogRepo: persistence.NewTakeoutImportLogRepository(db),
	}
}

// ImportProgressInfo 导入进度信息
type ImportProgressInfo struct {
	UUID            uint64 `json:"uuid"`
	Platform        string `json:"platform"`
	ImportType      int8   `json:"import_type"`
	ImportDirection string `json:"import_direction"`
	Status          int8   `json:"status"`
	Progress        int    `json:"progress"`
	SuccessCount    int    `json:"success_count"`
	FailureCount    int    `json:"failure_count"`
	TotalCount      int    `json:"total_count"`
	ErrorMessage    string `json:"error_message"`
	StartTime       int64  `json:"start_time"`
	EndTime         int64  `json:"end_time"`
	Duration        int64  `json:"duration"`
	EstimatedTime   int64  `json:"estimated_time"` // 预估剩余时间（秒）
}

// CheckImportStatus 检查导入状态，判断是否可以开始新的导入
// 返回：是否可以导入、正在进行的导入日志
func (s *ImportProgressService) CheckImportStatus(ctx context.Context, platform string) (bool, *model.TakeoutImportLog, error) {
	// 查询是否有进行中的导入任务
	inProgressLog, err := s.importLogRepo.FindInProgressByPlatform(ctx, platform)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check import status: %w", err)
	}

	// 如果有进行中的任务，检查是否超时
	if inProgressLog != nil {
		// 超时时间：30分钟
		timeout := int64(30 * 60)
		if time.Now().Unix()-inProgressLog.StartTime > timeout {
			// 超时，标记为失败
			inProgressLog.MarkFailed("导入超时")
			if updateErr := s.importLogRepo.Update(ctx, inProgressLog); updateErr != nil {
				return false, nil, fmt.Errorf("failed to mark timeout import as failed: %w", updateErr)
			}
			return true, nil, nil
		}
		return false, inProgressLog, nil
	}

	return true, nil, nil
}

// StartImport 开始导入，创建导入日志
func (s *ImportProgressService) StartImport(ctx context.Context, platform string, importType int8, importDirection string) (*model.TakeoutImportLog, error) {
	// 检查是否可以开始导入
	canImport, _, err := s.CheckImportStatus(ctx, platform)
	if err != nil {
		return nil, err
	}

	if !canImport {
		return nil, errors.New("已有导入任务正在进行中")
	}

	// 创建导入日志
	now := time.Now().Unix()
	log := &model.TakeoutImportLog{
		BaseModel: model.BaseModel{
			Uuid:       utils.MustGetID(),
			CreateTime: now,
			UpdateTime: now,
		},
		Platform:        platform,
		ImportType:      importType,
		ImportDirection: importDirection,
		Status:          model.ImportStatusInProgress,
		Progress:        0,
		SuccessCount:    0,
		FailureCount:    0,
		TotalCount:      0,
		StartTime:       now,
	}

	if err := s.importLogRepo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create import log: %w", err)
	}

	return log, nil
}

// UpdateProgress 更新导入进度
func (s *ImportProgressService) UpdateProgress(ctx context.Context, uuid uint64, current, total int) error {
	log, err := s.importLogRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to find import log: %w", err)
	}

	if log == nil {
		return errors.New("import log not found")
	}

	log.UpdateProgress(current, total)
	log.SetTotalCount(total)

	return s.importLogRepo.Update(ctx, log)
}

// IncrementSuccess 增加成功计数
func (s *ImportProgressService) IncrementSuccess(ctx context.Context, uuid uint64) error {
	log, err := s.importLogRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to find import log: %w", err)
	}

	if log == nil {
		return errors.New("import log not found")
	}

	log.IncrementSuccess()

	return s.importLogRepo.Update(ctx, log)
}

// IncrementFailure 增加失败计数
func (s *ImportProgressService) IncrementFailure(ctx context.Context, uuid uint64) error {
	log, err := s.importLogRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to find import log: %w", err)
	}

	if log == nil {
		return errors.New("import log not found")
	}

	log.IncrementFailure()

	return s.importLogRepo.Update(ctx, log)
}

// CompleteImport 完成导入
func (s *ImportProgressService) CompleteImport(ctx context.Context, uuid uint64, isSuccess bool, errorMsg string) error {
	log, err := s.importLogRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to find import log: %w", err)
	}

	if log == nil {
		return errors.New("import log not found")
	}

	if isSuccess {
		log.MarkSuccess()
	} else {
		log.MarkFailed(errorMsg)
	}

	return s.importLogRepo.Update(ctx, log)
}

// GetProgressInfo 获取导入进度信息
func (s *ImportProgressService) GetProgressInfo(ctx context.Context, uuid uint64) (*ImportProgressInfo, error) {
	log, err := s.importLogRepo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to find import log: %w", err)
	}

	if log == nil {
		return nil, errors.New("import log not found")
	}

	info := &ImportProgressInfo{
		UUID:            log.Uuid,
		Platform:        log.Platform,
		ImportType:      log.ImportType,
		ImportDirection: log.ImportDirection,
		Status:          log.Status,
		Progress:        log.Progress,
		SuccessCount:    log.SuccessCount,
		FailureCount:    log.FailureCount,
		TotalCount:      log.TotalCount,
		ErrorMessage:    log.ErrorMessage,
		StartTime:       log.StartTime,
		EndTime:         log.EndTime,
		Duration:        log.Duration,
	}

	// 计算预估剩余时间
	if log.IsInProgress() && log.Progress > 0 {
		elapsed := time.Now().Unix() - log.StartTime
		estimatedTotal := (elapsed * 100) / int64(log.Progress)
		info.EstimatedTime = estimatedTotal - elapsed
		if info.EstimatedTime < 0 {
			info.EstimatedTime = 0
		}
	}

	return info, nil
}

// GetLatestProgressByPlatform 获取指定平台最新的导入进度
func (s *ImportProgressService) GetLatestProgressByPlatform(ctx context.Context, platform string) (*ImportProgressInfo, error) {
	log, err := s.importLogRepo.FindLatestByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to find latest import log: %w", err)
	}

	if log == nil {
		return nil, nil
	}

	return s.GetProgressInfo(ctx, log.Uuid)
}

// ListImportLogs 查询导入日志列表
func (s *ImportProgressService) ListImportLogs(ctx context.Context, platform string, importType int8, status int8, page, pageSize int) ([]*model.TakeoutImportLog, int64, error) {
	return s.importLogRepo.List(ctx, platform, importType, status, page, pageSize)
}
