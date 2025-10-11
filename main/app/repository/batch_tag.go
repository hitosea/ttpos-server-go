package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IBatchTagRepo interface {
	GetBatchTags(opts ...DBOption) ([]*model.BatchTag, error) // 获取分批类型列表
	GetBatchTag(opts ...DBOption) (*model.BatchTag, error)    // 获取分批类型详情
	GetBatchTagList() ([]*model.BatchTag, error)              // 获取分批类型列表
	GetBatchTagCount() (int64, error)                         // 获取分批类型数量
	GetBatchTagInfo(uuid uint64) (*model.BatchTag, error)     // 获取分批类型详情
	CreateBatchTag(batchTag model.BatchTag) error             // 创建分批类型
	GetMaxSort() (int, error)                                 // 获取当前最大的排序值
	CheckColorExists(color string, uuid uint64) bool          // 检查颜色是否已被使用
	UpdateBatchTag(batchTag model.BatchTag) error             // 更新分批类型
	DeleteBatchTag(uuid uint64) error                         // 删除分批类型
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
	opts := []DBOption{NotDeleted}
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

func (r *BatchTagRepoImpl) GetBatchTagInfo(uuid uint64) (*model.BatchTag, error) {
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

// 创建分批类型
func (r *BatchTagRepoImpl) CreateBatchTag(batchTag model.BatchTag) error {
	batchTag.SetNil()
	return r.db.Model(&model.BatchTag{}).Create(&batchTag).Error
}

// 获取当前最大的排序值
func (r *BatchTagRepoImpl) GetMaxSort() (int, error) {
	var maxSort int
	err := r.db.Model(&model.BatchTag{}).Select("MAX(sort)").Scan(&maxSort).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return maxSort, nil
}

// 检查颜色是否已被使用。uuid不为0时，排除该uuid
func (r *BatchTagRepoImpl) CheckColorExists(color string, uuid uint64) bool {
	var exists uint64
	if uuid != 0 {
		r.db.Model(&model.BatchTag{}).Where("color = ? AND uuid != ?", color, uuid).Select("uuid").Scan(&exists)
	} else {
		r.db.Model(&model.BatchTag{}).Where("color = ?", color).Select("uuid").Scan(&exists)
	}
	return exists > 0
}

// 更新分批类型
func (r *BatchTagRepoImpl) UpdateBatchTag(batchTag model.BatchTag) error {
	return r.db.Model(&model.BatchTag{}).Where("uuid = ?", batchTag.Uuid).Updates(map[string]any{
		"color": batchTag.Color,
		"name":  batchTag.Name,
	}).Error
}

// 删除分批类型
func (r *BatchTagRepoImpl) DeleteBatchTag(uuid uint64) error {
	return r.db.Model(&model.BatchTag{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}
