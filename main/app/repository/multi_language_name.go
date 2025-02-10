// 定义多语言名称仓库接口
package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMultiLanguageNameRepo 定义多语言名称仓库接口
type IMultiLanguageNameRepo interface {
	GetMultiLanguageName(id uint) (model.MultiLanguageName, error)                    // 获取多语言名称
	CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint, error)  // 创建多语言名称
	UpdateMultiLanguageName(id uint, multiLanguageName model.MultiLanguageName) error // 更新多语言名称
	DeleteMultiLanguageName(id uint) error                                            // 删除多语言名称
}

// NewMultiLanguageNameRepoImpl 创建新的多语言名称仓库
func NewMultiLanguageNameRepoImpl(db *gorm.DB) IMultiLanguageNameRepo {
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
func (r *MultiLanguageNameRepoImpl) GetMultiLanguageName(id uint) (model.MultiLanguageName, error) {
	var multiLanguageName model.MultiLanguageName
	err := r.db.First(&multiLanguageName, id).Error // 从数据库中获取多语言名称
	return multiLanguageName, err
}

// CreateMultiLanguageName 创建多语言名称
func (r *MultiLanguageNameRepoImpl) CreateMultiLanguageName(multiLanguageName model.MultiLanguageName) (uint, error) {
	err := r.db.Create(&multiLanguageName).Error // 将多语言名称插入数据库
	return multiLanguageName.ID, err
}

// UpdateMultiLanguageName 更新多语言名称
func (r *MultiLanguageNameRepoImpl) UpdateMultiLanguageName(id uint, multiLanguageName model.MultiLanguageName) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("id = ?", id).Updates(&multiLanguageName).Error // 更新数据库中的多语言名称
}

// DeleteMultiLanguageName 删除多语言名称
func (r *MultiLanguageNameRepoImpl) DeleteMultiLanguageName(id uint) error {
	return r.db.Model(&model.MultiLanguageName{}).Where("id = ?", id).Update("delete_time", time.Now().Unix()).Error // 逻辑删除多语言名称
}
