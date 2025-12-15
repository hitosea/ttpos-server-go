package persistence

import (
	"errors"
	"fmt"
	"time"
	"ttpos-server-go/app/model"
	settingRepo "ttpos-server-go/app/modules/setting/domain/repository"
	companyRepo "ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// SettingRepositoryImpl 设置存储实现
type SettingRepositoryImpl struct {
	db    *database.DBManager
	cache cache.Cache
}

// NewSettingRepository 创建设置存储实例
func NewSettingRepository(db *database.DBManager, cache cache.Cache) settingRepo.ISettingRepository {
	return &SettingRepositoryImpl{
		db:    db,
		cache: cache,
	}
}

// GetRepository 获取数据库连接
func (r *SettingRepositoryImpl) GetRepository(db *gorm.DB) *gorm.DB {
	return db
}

// FindByKey 根据键查找设置
func (r *SettingRepositoryImpl) FindByKey(ctx context.Context, key string) (*model.Setting, error) {
	db := ctx.GetDB()

	var setting model.Setting
	err := db.Where("`key` = ? AND delete_time = 0", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

// FindAll 获取所有设置
func (r *SettingRepositoryImpl) FindAll(ctx context.Context) ([]*model.Setting, error) {
	db := ctx.GetDB()

	var settings []*model.Setting
	err := db.Where("delete_time = 0").Find(&settings).Error
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// Save 保存设置
func (r *SettingRepositoryImpl) Save(ctx context.Context, setting *model.Setting) error {
	db := ctx.GetDB()

	return db.Create(setting).Error
}

// Update 更新设置
func (r *SettingRepositoryImpl) Update(ctx context.Context, setting *model.Setting) error {
	db := ctx.GetDB()

	return db.Where("`key` = ?", setting.Key).Updates(setting).Error
}

// Delete 删除设置
func (r *SettingRepositoryImpl) Delete(ctx context.Context, key string) error {
	db := ctx.GetDB()

	return db.Where("key = ?", key).Update("delete_time", time.Now().Unix()).Error
}

// SaveBatch 批量保存设置
func (r *SettingRepositoryImpl) SaveBatch(ctx context.Context, settings []*model.Setting) error {
	db := ctx.GetDB()

	return db.CreateInBatches(settings, len(settings)).Error
}

// UpdateBatch 批量更新设置
func (r *SettingRepositoryImpl) UpdateBatch(ctx context.Context, settings []*model.Setting) error {
	db := ctx.GetDB()

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, setting := range settings {
		if err := tx.Where("`key` = ?", setting.Key).Updates(setting).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// FindByKeys 根据多个键查找设置
func (r *SettingRepositoryImpl) FindByKeys(ctx context.Context, keys []string) ([]*model.Setting, error) {
	db := ctx.GetDB()

	var settings []*model.Setting
	err := db.Where("key IN (?) AND delete_time = 0", keys).Find(&settings).Error
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// Exists 检查设置是否存在
func (r *SettingRepositoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	db := ctx.GetDB()

	var count int64
	err := db.Model(&model.Setting{}).Where("key = ? AND delete_time = 0", key).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFromCache 从缓存获取
func (r *SettingRepositoryImpl) GetFromCache(ctx context.Context, key string) (string, bool) {
	cacheKey := r.buildCacheKey(ctx, key)
	if value, exists := r.cache.Get(cacheKey); exists {
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
	}
	return "", false
}

// SetToCache 设置缓存
func (r *SettingRepositoryImpl) SetToCache(ctx context.Context, key, value string, ttl int) error {
	cacheKey := r.buildCacheKey(ctx, key)
	r.cache.Set(cacheKey, value, 0) // 使用默认TTL
	return nil
}

// DeleteFromCache 删除缓存
func (r *SettingRepositoryImpl) DeleteFromCache(ctx context.Context, key string) error {
	cacheKey := r.buildCacheKey(ctx, key)
	r.cache.Del(cacheKey)
	return nil
}

// ClearCache 清除缓存
func (r *SettingRepositoryImpl) ClearCache(ctx context.Context) error {
	companyUuid := ctx.Value("company_uuid").(uint64)
	cacheKey := fmt.Sprintf("setting:company_id:%d", companyUuid)
	r.cache.Del(cacheKey)
	return nil
}

// buildCacheKey 构建缓存键
func (r *SettingRepositoryImpl) buildCacheKey(ctx context.Context, key string) string {
	companyUuid := ctx.GetCompanyUuid()
	return fmt.Sprintf("setting:company_id:%d:key:%s", companyUuid, key)
}

// GetCompanySetting 获取公司设置
func (r *SettingRepositoryImpl) GetCompanySetting(ctx context.Context) (*model.CompanySetting, error) {
	db := ctx.GetDB()
	if db == nil {
		return nil, fmt.Errorf("db not found")
	}
	companySettingRepo := companyRepo.NewCompanySettingRepo(db)
	companySetting := companySettingRepo.Get()
	return &companySetting, nil
}

// GetPaymentMethodList 获取支付方式列表（与旧服务保持一致）
func (r *SettingRepositoryImpl) GetPaymentMethodList(ctx context.Context) ([]model.PaymentMethod, error) {
	db := ctx.GetDB()
	if db == nil {
		return nil, fmt.Errorf("db not found")
	}

	var paymentMethods []model.PaymentMethod
	// 按照 sort 升序、create_time 降序排序
	err := db.Where("delete_time = 0").
		Order("sort ASC").
		Order("create_time DESC").
		Find(&paymentMethods).Error
	if err != nil {
		return nil, err
	}

	return paymentMethods, nil
}
