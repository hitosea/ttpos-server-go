package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IReturnFoodReasonRepo 退菜原因仓库接口
type IReturnFoodReasonRepo interface {
	GetReturnFoodReasonList() ([]model.ReturnFoodReason, error)                    // 获取退菜原因列表
	UpdateReturnFoodReason(id uint, returnFoodReason model.ReturnFoodReason) error // 更新退菜原因
	CreateReturnFoodReason(returnFoodReason model.ReturnFoodReason) (uint, error)  // 创建退菜原因
	DeleteReturnFoodReason(id uint) error                                          // 删除退菜原因
}

// NewReturnFoodReasonRepo 创建新的退菜原因仓库
func NewReturnFoodReasonRepo(db *gorm.DB) IReturnFoodReasonRepo {
	return NewReturnFoodReasonRepoImpl(db)
}

// NewReturnFoodReasonRepoImpl 创建新的退菜原因仓库实现
func NewReturnFoodReasonRepoImpl(db *gorm.DB) *ReturnFoodReasonRepoImpl {
	return &ReturnFoodReasonRepoImpl{db: db}
}

type ReturnFoodReasonRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetReturnFoodReasonList 获取退菜原因列表，排除逻辑删除的退菜原因
func (r *ReturnFoodReasonRepoImpl) GetReturnFoodReasonList() ([]model.ReturnFoodReason, error) {
	var returnFoodReasons []model.ReturnFoodReason
	err := r.db.Model(&model.ReturnFoodReason{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&returnFoodReasons).Error
	return returnFoodReasons, err
}

// UpdateReturnFoodReason 更新退菜原因
func (r *ReturnFoodReasonRepoImpl) UpdateReturnFoodReason(id uint, returnFoodReason model.ReturnFoodReason) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ReturnFoodReason{}).Where("id = ?", id).Updates(returnFoodReason).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return err
	}

	if err := tx.Model(&returnFoodReason.MultiLanguageName).Where("id = ?", returnFoodReason.MultiLanguageNameUuid).Updates(returnFoodReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return err
	}

	return tx.Commit().Error // 提交事务
}

// CreateReturnFoodReason 创建退菜原因
func (r *ReturnFoodReasonRepoImpl) CreateReturnFoodReason(returnFoodReason model.ReturnFoodReason) (uint, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&returnFoodReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建退菜原因
	if err := tx.Create(&returnFoodReason).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return returnFoodReason.Uuid, tx.Commit().Error // 提交事务
}

// DeleteReturnFoodReason 软删除退菜原因
func (r *ReturnFoodReasonRepoImpl) DeleteReturnFoodReason(id uint) error {
	return r.db.Model(&model.ReturnFoodReason{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
