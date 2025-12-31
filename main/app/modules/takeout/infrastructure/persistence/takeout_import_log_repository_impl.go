package persistence

import (
	"context"
	"errors"
	"time"

	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/repository"

	"gorm.io/gorm"
)

// TakeoutImportLogRepositoryImpl 外卖导入日志仓储实现
type TakeoutImportLogRepositoryImpl struct {
	db *gorm.DB
}

// NewTakeoutImportLogRepository 创建外卖导入日志仓储实例
func NewTakeoutImportLogRepository(db *gorm.DB) repository.TakeoutImportLogRepository {
	return &TakeoutImportLogRepositoryImpl{
		db: db,
	}
}

// Create 创建导入日志
func (r *TakeoutImportLogRepositoryImpl) Create(ctx context.Context, log *model.TakeoutImportLog) error {
	if log == nil {
		return errors.New("log cannot be nil")
	}

	now := time.Now().Unix()
	if log.CreateTime == 0 {
		log.CreateTime = now
	}
	if log.UpdateTime == 0 {
		log.UpdateTime = now
	}

	return r.db.WithContext(ctx).Create(log).Error
}

// Update 更新导入日志
func (r *TakeoutImportLogRepositoryImpl) Update(ctx context.Context, log *model.TakeoutImportLog) error {
	if log == nil {
		return errors.New("log cannot be nil")
	}

	log.UpdateTime = time.Now().Unix()

	return r.db.WithContext(ctx).
		Model(&model.TakeoutImportLog{}).
		Where("uuid = ? AND delete_time = 0", log.Uuid).
		Updates(log).Error
}

// FindByUUID 根据 UUID 查询导入日志
func (r *TakeoutImportLogRepositoryImpl) FindByUUID(ctx context.Context, uuid uint64) (*model.TakeoutImportLog, error) {
	var log model.TakeoutImportLog
	err := r.db.WithContext(ctx).
		Where("uuid = ? AND delete_time = 0", uuid).
		First(&log).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

// FindByID 根据 ID 查询导入日志
func (r *TakeoutImportLogRepositoryImpl) FindByID(ctx context.Context, id uint64) (*model.TakeoutImportLog, error) {
	var log model.TakeoutImportLog
	err := r.db.WithContext(ctx).
		Where("id = ? AND delete_time = 0", id).
		First(&log).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

// FindLatestByPlatform 查询指定平台最新的一条导入日志
func (r *TakeoutImportLogRepositoryImpl) FindLatestByPlatform(ctx context.Context, platform string) (*model.TakeoutImportLog, error) {
	var log model.TakeoutImportLog
	err := r.db.WithContext(ctx).
		Where("platform = ? AND delete_time = 0", platform).
		Order("create_time DESC").
		First(&log).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

// FindInProgressByPlatform 查询指定平台进行中的导入日志
func (r *TakeoutImportLogRepositoryImpl) FindInProgressByPlatform(ctx context.Context, platform string) (*model.TakeoutImportLog, error) {
	var log model.TakeoutImportLog
	err := r.db.WithContext(ctx).
		Where("platform = ? AND status = ? AND delete_time = 0", platform, model.ImportLogStatusInProgress).
		Order("create_time DESC").
		First(&log).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

// List 查询导入日志列表
func (r *TakeoutImportLogRepositoryImpl) List(ctx context.Context, platform string, importType int8, status int8, page, pageSize int) ([]*model.TakeoutImportLog, int64, error) {
	var logs []*model.TakeoutImportLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.TakeoutImportLog{}).Where("delete_time = 0")

	// 平台筛选
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	// 导入类型筛选
	if importType > 0 {
		query = query.Where("import_type = ?", importType)
	}

	// 状态筛选
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.
		Order("create_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Delete 软删除导入日志
func (r *TakeoutImportLogRepositoryImpl) Delete(ctx context.Context, uuid uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.TakeoutImportLog{}).
		Where("uuid = ? AND delete_time = 0", uuid).
		Update("delete_time", time.Now().Unix()).Error
}
