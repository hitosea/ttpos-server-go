package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetCustomerTypePriceRepo 自助餐客户类型价格
type IBuffetCustomerTypePriceRepo interface {
	GetBuffetCustomerTypePriceList() ([]model.BuffetCustomerTypePrice, error)
	UpdateBuffetCustomerTypePrice(uuid uint, buffetCustomerTypePrice model.BuffetCustomerTypePrice) error
	CreateBuffetCustomerTypePrice(buffetCustomerTypePrice model.BuffetCustomerTypePrice) (uint64, error)
	DeleteBuffetCustomerTypePrice(uuid uint) error
}

func NewBuffetCustomerTypePriceRepo(db *gorm.DB) IBuffetCustomerTypePriceRepo {
	return NewBuffetCustomerTypePriceRepoImpl(db)
}

// NewBuffetCustomerTypePriceRepoImpl 创建新的自助餐客户类型价格仓库实现
func NewBuffetCustomerTypePriceRepoImpl(db *gorm.DB) *BuffetCustomerTypePriceRepoImpl {
	return &BuffetCustomerTypePriceRepoImpl{db: db}
}

type BuffetCustomerTypePriceRepoImpl struct {
	db *gorm.DB
}

// GetBuffetCustomerTypePriceList 获取自助餐客户类型列表，排除逻辑删除的自助餐客户类型
func (r *BuffetCustomerTypePriceRepoImpl) GetBuffetCustomerTypePriceList() ([]model.BuffetCustomerTypePrice, error) {
	var buffetCustomerTypePrices []model.BuffetCustomerTypePrice
	err := r.db.Model(&model.BuffetCustomerTypePrice{}).Where("delete_time = ?", 0).Find(&buffetCustomerTypePrices).Error
	return buffetCustomerTypePrices, err
}

// UpdateBuffetCustomerTypePrice 更新自助餐客户类型
func (r *BuffetCustomerTypePriceRepoImpl) UpdateBuffetCustomerTypePrice(uuid uint, buffetCustomerTypePrice model.BuffetCustomerTypePrice) error {
	if err := r.db.Model(&model.BuffetCustomerTypePrice{}).Where("uuid = ?", uuid).Updates(buffetCustomerTypePrice).Error; err != nil {
		return err
	}
	return nil
}

// CreateBuffetCustomerTypePrice 创建自助餐客户类型
func (r *BuffetCustomerTypePriceRepoImpl) CreateBuffetCustomerTypePrice(buffetCustomerTypePrice model.BuffetCustomerTypePrice) (uint64, error) {

	// 创建自助餐客户类型
	if err := r.db.Create(&buffetCustomerTypePrice).Error; err != nil {
		return 0, err
	}
	return buffetCustomerTypePrice.Uuid, nil
}

// DeleteBuffetCustomerTypePrice 软删除自助餐客户类型
func (r *BuffetCustomerTypePriceRepoImpl) DeleteBuffetCustomerTypePrice(uuid uint) error {
	return r.db.Model(&model.BuffetCustomerTypePrice{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}
