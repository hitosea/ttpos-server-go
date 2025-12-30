package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	cacheRepo "ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// GetBatchTagInfoOption GetBatchTagInfo 方法的选项
type GetBatchTagInfoOption struct {
	// SkipCache 是否跳过缓存检查，直接执行查询
	// true: 跳过所有缓存（L1/L2），直接从数据库查询（仍会写入缓存）
	// false: 正常流程，按配置决定是否使用缓存（默认）
	SkipCache bool
}

// WithSkipCacheForBatchTag 设置是否跳过缓存选项
func WithSkipCacheForBatchTag() func(*GetBatchTagInfoOption) {
	return func(opt *GetBatchTagInfoOption) {
		opt.SkipCache = true
	}
}

type IBatchTagRepo interface {
	GetBatchTags(opts ...DBOption) ([]*model.BatchTag, error)                                                       // 获取分批类型列表
	GetBatchTag(opts ...DBOption) (*model.BatchTag, error)                                                          // 获取分批类型详情
	GetBatchTagList() ([]*model.BatchTag, error)                                                                    // 获取分批类型列表
	GetBatchTagCount() (int64, error)                                                                               // 获取分批类型数量
	GetBatchTagInfo(companyUuid uint64, uuid uint64, opts ...func(*GetBatchTagInfoOption)) (*model.BatchTag, error) // 获取分批类型详情
	CreateBatchTag(batchTag model.BatchTag) error                                                                   // 创建分批类型
	GetMaxSort() (int, error)                                                                                       // 获取当前最大的排序值
	CheckColorExists(color string, uuid uint64) bool                                                                // 检查颜色是否已被使用
	UpdateBatchTag(batchTag model.BatchTag) error                                                                   // 更新分批类型
	DeleteBatchTag(uuid uint64) error                                                                               // 删除分批类型
}

func NewBatchTagRepo(db *gorm.DB) IBatchTagRepo {
	return NewBatchTagRepoImpl(db)
}

func NewBatchTagRepoImpl(db *gorm.DB) *BatchTagRepoImpl {
	return &BatchTagRepoImpl{db: db}
}

type BatchTagRepoImpl struct {
	db *gorm.DB
}

func (r *BatchTagRepoImpl) GetBatchTags(opts ...DBOption) ([]*model.BatchTag, error) {
	var batchTags []*model.BatchTag
	db := r.db.Model(&model.BatchTag{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&batchTags).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return batchTags, nil
}

func (r *BatchTagRepoImpl) GetBatchTagList() ([]*model.BatchTag, error) {
	opts := []DBOption{
		NotDeleted,
		// 根据sort 排序
		CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	}
	list, err := r.GetBatchTags(opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return list, nil
}

func (r *BatchTagRepoImpl) GetBatchTagCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.BatchTag{}).Where("delete_time = ?", 0).Count(&count).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return count, nil
}

func (r *BatchTagRepoImpl) GetBatchTag(opts ...DBOption) (*model.BatchTag, error) {
	var batchTag model.BatchTag
	db := r.db.Model(&model.BatchTag{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&batchTag).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &batchTag, nil
}

func (r *BatchTagRepoImpl) GetBatchTagInfo(companyUuid uint64, uuid uint64, opts ...func(*GetBatchTagInfoOption)) (*model.BatchTag, error) {
	// 解析选项
	option := &GetBatchTagInfoOption{
		SkipCache: false, // 默认不跳过缓存
	}
	for _, opt := range opts {
		opt(option)
	}

	// 检查是否启用对象存储缓存
	var batchTag *model.BatchTag
	var err error

	if adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 使用对象存储模块缓存查询
		batchTag, err = r.getBatchTagInfoWithCache(companyUuid, uuid, option.SkipCache)
	} else {
		// 直接查询数据库
		batchTag, err = r.queryBatchTagInfo(uuid)
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return batchTag, nil
}

// queryBatchTagInfo 查询分批标签信息（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *BatchTagRepoImpl) queryBatchTagInfo(uuid uint64) (*model.BatchTag, error) {
	batchTag, err := r.GetBatchTag(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return batchTag, nil
}

// getBatchTagInfoWithCache 使用对象存储模块缓存查询分批标签信息（包含预加载的关联数据）
func (r *BatchTagRepoImpl) getBatchTagInfoWithCache(companyUuid uint64, uuid uint64, skipCache bool) (*model.BatchTag, error) {
	if uuid == 0 {
		return nil, errors.New("getBatchTagInfoWithCache uuid cannot be 0")
	}
	if companyUuid == 0 {
		return nil, errors.New("getBatchTagInfoWithCache companyUuid cannot be 0")
	}

	// 构建缓存 key（使用 companyUuid 而不是 context）
	key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeBatchTag, uuid)

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.BatchTag](cache.Global, 5*time.Minute)

	// 使用缓存查询，根据 skipCache 选项决定是否跳过缓存
	var result *model.BatchTag
	var err error
	if skipCache {
		// 跳过缓存，直接查询数据库
		result, err = cacheLayer.GET(key, func() (*model.BatchTag, error) {
			return r.queryBatchTagInfo(uuid)
		}, cacheRepo.WithSkipCache())
	} else {
		// 正常流程，检查缓存
		result, err = cacheLayer.GET(key, func() (*model.BatchTag, error) {
			// 缓存未命中时，从数据库查询（包含所有预加载）
			return r.queryBatchTagInfo(uuid)
		})
	}

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryBatchTagInfo(uuid)
	}

	return result, nil
}

// 创建分批类型
func (r *BatchTagRepoImpl) CreateBatchTag(batchTag model.BatchTag) error {
	batchTag.SetNil()
	return r.db.Model(&model.BatchTag{}).Create(&batchTag).Error
}

// 获取当前最大的排序值
func (r *BatchTagRepoImpl) GetMaxSort() (int, error) {
	var maxSort int
	err := r.db.Model(&model.BatchTag{}).Select("ifnull(max(sort), 0)").Scan(&maxSort).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return maxSort, nil
}

// 检查颜色是否已被使用。uuid不为0时，排除该uuid
func (r *BatchTagRepoImpl) CheckColorExists(color string, uuid uint64) bool {
	var exists uint64
	if uuid != 0 {
		r.db.Model(&model.BatchTag{}).Where("color = ? AND uuid != ? AND delete_time = 0", color, uuid).Select("uuid").Scan(&exists)
	} else {
		r.db.Model(&model.BatchTag{}).Where("color = ? AND delete_time = 0", color).Select("uuid").Scan(&exists)
	}
	return exists > 0
}

// 更新分批类型
func (r *BatchTagRepoImpl) UpdateBatchTag(batchTag model.BatchTag) error {
	return r.db.Model(&model.BatchTag{}).Where("uuid = ?", batchTag.Uuid).Updates(map[string]any{
		"color":        batchTag.Color,
		"name":         batchTag.Name,
		"abbreviation": batchTag.Abbreviation,
	}).Error
}

// 删除分批类型
func (r *BatchTagRepoImpl) DeleteBatchTag(uuid uint64) error {
	return r.db.Model(&model.BatchTag{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}
