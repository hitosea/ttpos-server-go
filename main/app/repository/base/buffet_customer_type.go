package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetCustomerTypeRepo 自助餐客户类型
type IBuffetCustomerTypeRepo interface {
	GetBuffetCustomerTypeList() ([]model.BuffetCustomerType, error)
	UpdateBuffetCustomerType(uuid uint, buffetCustomerType model.BuffetCustomerType) error
	CreateBuffetCustomerType(buffetCustomerType model.BuffetCustomerType) (uint, error)
	DeleteBuffetCustomerType(uuid uint) error
}

func NewBuffetCustomerTypeRepo(db *gorm.DB) IBuffetCustomerTypeRepo {
	return NewBuffetCustomerTypeRepoImpl(db)
}

// NewProductFlavorRepoImpl 创建新的商品规格仓库实现
func NewBuffetCustomerTypeRepoImpl(db *gorm.DB) *BuffetCustomerTypeRepoImpl {
	return &BuffetCustomerTypeRepoImpl{db: db}
}

type BuffetCustomerTypeRepoImpl struct {
	db *gorm.DB
}

// GetProductFlavorList 获取商品规格列表，排除逻辑删除的规格
func (r *BuffetCustomerTypeRepoImpl) GetBuffetCustomerTypeList() ([]model.BuffetCustomerType, error) {
	var buffetCustomerTypes []model.BuffetCustomerType
	err := r.db.Model(&model.BuffetCustomerType{}).Where("delete_time = ?", 0).Find(&buffetCustomerTypes).Error
	return buffetCustomerTypes, err
}

// UpdateBuffetCustomerType 更新自助餐客户类型
func (r *BuffetCustomerTypeRepoImpl) UpdateBuffetCustomerType(uuid uint, buffetCustomerType model.BuffetCustomerType) error {
	if err := r.db.Model(&model.BuffetCustomerType{}).Where("uuid = ?", uuid).Updates(buffetCustomerType).Error; err != nil {
		return err
	}
	return nil
}

// CreateBuffetCustomerType 创建自助餐客户类型
func (r *BuffetCustomerTypeRepoImpl) CreateBuffetCustomerType(buffetCustomerType model.BuffetCustomerType) (uint, error) {

	// 创建自助餐客户类型
	if err := r.db.Create(&buffetCustomerType).Error; err != nil {
		return 0, err
	}
	return buffetCustomerType.UUID, nil
}

// DeleteProductFlavor 软删除商品规格
func (r *BuffetCustomerTypeRepoImpl) DeleteBuffetCustomerType(uuid uint) error {
	return r.db.Model(&model.BuffetCustomerType{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
