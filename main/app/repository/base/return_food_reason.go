package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// IReturnFoodReasonRepo 退菜原因仓库接口
type IReturnFoodReasonRepo interface {
	GetReturnFoodReasonList() ([]model.ReturnFoodReason, error)                        // 获取退菜原因列表
	UpdateReturnFoodReason(uuid uint64, returnFoodReason model.ReturnFoodReason) error // 更新退菜原因
	CreateReturnFoodReason(returnFoodReason model.ReturnFoodReason) (uint64, error)    // 创建退菜原因
	DeleteReturnFoodReason(uuid uint64) error                                          // 删除退菜原因
	DeleteReturnFoodReasons(uuids []uint64) error                                      // 批量删除退菜原因
	ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error)                       // 根据uuid数组验证退菜原因是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
	GetReturnFoodReasonByUuid(uuid uint64) (*model.ReturnFoodReason, error)            // 根据uuid获取退菜原因
	GetReturnFoodReasons(opts ...repository.DBOption) ([]*model.ReturnFoodReason, error)
	GetReturnFoodReasonListByUuids(uuids []uint64) ([]*model.ReturnFoodReason, error)
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

func (r *ReturnFoodReasonRepoImpl) GetReturnFoodReasonListByUuids(uuids []uint64) ([]*model.ReturnFoodReason, error) {
	reasons, err := r.GetReturnFoodReasons(
		repository.CommonRepo.WhereInUuids(uuids),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.SortWithCreateTime("desc"),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return reasons, nil
}

func (r *ReturnFoodReasonRepoImpl) GetReturnFoodReasons(opts ...repository.DBOption) ([]*model.ReturnFoodReason, error) {
	returnFoodReasons := make([]*model.ReturnFoodReason, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&returnFoodReasons)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return returnFoodReasons, nil
}

// GetReturnFoodReasonList 获取退菜原因列表，排除逻辑删除的退菜原因
func (r *ReturnFoodReasonRepoImpl) GetReturnFoodReasonList() ([]model.ReturnFoodReason, error) {
	var returnFoodReasons []model.ReturnFoodReason
	err := r.db.Model(&model.ReturnFoodReason{}).
		Preload("MultiLanguageName").
		Where("delete_time = ?", 0).
		Order("id ASC").
		Find(&returnFoodReasons).Error
	return returnFoodReasons, errors.WithMessage(err)
}

// UpdateReturnFoodReason 更新退菜原因
func (r *ReturnFoodReasonRepoImpl) UpdateReturnFoodReason(id uint64, returnFoodReason model.ReturnFoodReason) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ReturnFoodReason{}).Where("uuid = ?", id).Updates(returnFoodReason).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&returnFoodReason.MultiLanguageName).Where("uuid = ?", returnFoodReason.MultiLanguageNameUuid).Updates(returnFoodReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateReturnFoodReason 创建退菜原因
func (r *ReturnFoodReasonRepoImpl) CreateReturnFoodReason(returnFoodReason model.ReturnFoodReason) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&returnFoodReason.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建退菜原因
	if err := tx.Create(&returnFoodReason).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	return returnFoodReason.Uuid, tx.Commit().Error // 提交事务
}

// DeleteReturnFoodReason 软删除退菜原因
func (r *ReturnFoodReasonRepoImpl) DeleteReturnFoodReason(uuid uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		reason := model.ReturnFoodReason{}
		tx.Model(&model.ReturnFoodReason{}).Where("uuid = ?", uuid).First(&reason)
		if err := tx.Model(&model.ReturnFoodReason{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
			return errors.WithMessage(err)
		}
		if reason.MultiLanguageNameUuid != 0 {
			if err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", reason.MultiLanguageNameUuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	})
}

// ExistsByUuids 根据uuid数组验证退菜原因是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
func (r *ReturnFoodReasonRepoImpl) ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error) {
	if len(uuids) == 0 {
		return [][2]uint64{}, []uint64{}, nil
	}

	var reasons []model.ReturnFoodReason
	err := r.db.Where("uuid IN ? AND delete_time = 0", uuids).Find(&reasons).Error
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	// 创建存在的UUID集合
	existMap := make(map[uint64]model.ReturnFoodReason)
	for _, reason := range reasons {
		existMap[reason.Uuid] = reason
	}

	// 找出不存在的UUID
	notFound := make([]uint64, 0)
	for _, uuid := range uuids {
		if _, ok := existMap[uuid]; !ok {
			notFound = append(notFound, uuid)
		}
	}

	// 创建结果数组，每个元素是[uuid, 多语言名称UUID]
	result := make([][2]uint64, len(reasons))
	for i, reason := range reasons {
		result[i] = [2]uint64{reason.Uuid, reason.MultiLanguageNameUuid}
	}

	return result, notFound, nil
}

func (r *ReturnFoodReasonRepoImpl) DeleteReturnFoodReasons(uuids []uint64) error {
	return r.db.Model(&model.ReturnFoodReason{}).Where("uuid IN (?)", uuids).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *ReturnFoodReasonRepoImpl) GetReturnFoodReasonByUuid(uuid uint64) (*model.ReturnFoodReason, error) {
	var reason model.ReturnFoodReason
	err := r.db.Where("uuid = ?", uuid).Scopes(repository.NotDeleted).First(&reason).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &reason, nil
}
