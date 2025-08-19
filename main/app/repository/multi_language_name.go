package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMultiLanguageNameRepo 定义多语言名称仓库接口
type IMultiLanguageNameRepo interface {
	IMultiLanguageNameQueryRepo
	CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint64, error)  // 创建多语言名称
	UpdateMultiLanguageName(id uint64, multiLanguageName model.MultiLanguageName) error // 更新多语言名称
	DeleteMultiLanguageName(id uint64) error                                            // 删除多语言名称
}

type IMultiLanguageNameQueryRepo interface {
	GetMultiLanguageName(opts ...DBOption) (*model.MultiLanguageName, error)
	GetMultiLanguageNameByUuid(uuid uint64) (model.MultiLanguageName, error) // 获取多语言名称
}

// NewMultiLanguageNameRepo 创建新的多语言名称仓库
func NewMultiLanguageNameRepo(db *gorm.DB) IMultiLanguageNameRepo {
	return NewMultiLanguageNameRepositoryImpl(db)
}

// MultiLanguageNameRepoImpl 多语言名称仓库实现
type MultiLanguageNameRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// NewMultiLanguageNameRepositoryImpl 创建新的多语言名称仓库实现
func NewMultiLanguageNameRepositoryImpl(db *gorm.DB) IMultiLanguageNameRepo {
	return &MultiLanguageNameRepoImpl{db: db}
}

// GetMultiLanguageName 获取多语言名称
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageName(opts ...DBOption) (*model.MultiLanguageName, error) {
	var multiLanguageName model.MultiLanguageName
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&multiLanguageName).Error
	return &multiLanguageName, errors.WithMessage(err)
}

// GetMultiLanguageNameByUuid 获取多语言名称
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageNameByUuid(uuid uint64) (model.MultiLanguageName, error) {
	multiLanguageName, err := r.GetMultiLanguageName(
		CommonRepo.WhereByUuid(uuid),
	)
	if err != nil {
		return model.MultiLanguageName{}, errors.WithMessage(err)
	}
	return *multiLanguageName, nil
}

// CreateMultiLanguageName 创建多语言名称
func (r *MultiLanguageNameRepoImpl) CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint64, error) {
	err := r.db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error // 将多语言名称插入数据库
	return multiLanguageName.Uuid, errors.WithMessage(err)
}

// UpdateMultiLanguageName 更新多语言名称
func (r *MultiLanguageNameRepoImpl) UpdateMultiLanguageName(uuid uint64, multiLanguageName model.MultiLanguageName) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"zh_name":    multiLanguageName.ZhName,
		"th_name":    multiLanguageName.ThName,
		"en_name":    multiLanguageName.EnName,
		"zh_tw_name": multiLanguageName.ZhTwName,
		"ja_name":    multiLanguageName.JaName,
		"ko_name":    multiLanguageName.KoName,
		"my_name":    multiLanguageName.MyName,
		"tr_name":    multiLanguageName.TrName,
		"sv_name":    multiLanguageName.SvName,
	}).Error // 更新数据库中的多语言名称
}

// DeleteMultiLanguageName 删除多语言名称
func (r *MultiLanguageNameRepoImpl) DeleteMultiLanguageName(id uint64) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", id).Update("delete_time", uint(time.Now().Unix())).Error // 逻辑删除多语言名称
}
