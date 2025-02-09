package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 客户呼叫记录
type CustomerCallRepoInterface interface {
	GetCustomerCallList() ([]model.CustomerCall, error)                // 获取客户呼叫记录列表
	UpdateCustomerCall(id uint, customerCall model.CustomerCall) error // 更新客户呼叫记录
	CreateCustomerCall(customerCall model.CustomerCall) (uint, error)  // 创建客户呼叫记录
	DeleteCustomerCall(id uint) error                                  // 软删除客户呼叫记录
}

func NewCustomerCallRepo(db *gorm.DB) CustomerCallRepoInterface {
	return NewCustomerCallRepoImpl(db)
}

// 创建新的客户呼叫记录仓库实现
func NewCustomerCallRepoImpl(db *gorm.DB) *CustomerCallRepoImpl {
	return &CustomerCallRepoImpl{db: db}
}

type CustomerCallRepoImpl struct {
	db *gorm.DB
}

// 获取客户呼叫记录列表，排除逻辑删除的客户呼叫记录
func (r *CustomerCallRepoImpl) GetCustomerCallList() ([]model.CustomerCall, error) {
	var customerCalls []model.CustomerCall
	err := r.db.Model(&model.CustomerCall{}).Where("delete_time = ?", 0).Find(&customerCalls).Error
	return customerCalls, err
}

// 更新客户呼叫记录
func (r *CustomerCallRepoImpl) UpdateCustomerCall(id uint, customerCall model.CustomerCall) error {
	return r.db.Model(&model.CustomerCall{}).Where("id = ?", id).Updates(customerCall).Error
}

// 创建客户呼叫记录
func (r *CustomerCallRepoImpl) CreateCustomerCall(customerCall model.CustomerCall) (uint, error) {
	return customerCall.Uuid, r.db.Create(&customerCall).Error
}

// 软删除客户呼叫记录
func (r *CustomerCallRepoImpl) DeleteCustomerCall(id uint) error {
	return r.db.Model(&model.CustomerCall{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}
