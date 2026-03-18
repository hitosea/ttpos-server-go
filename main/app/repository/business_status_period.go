package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IBusinessStatusPeriodRepo interface {
	GetOpenPeriod() (*model.BusinessStatusPeriod, error)
	GetBusinessStatus() int
	Create(period *model.BusinessStatusPeriod) error
	CloseOpenPeriod(endTime int64) error
}

func NewBusinessStatusPeriodRepo(db *gorm.DB) IBusinessStatusPeriodRepo {
	return &businessStatusPeriodRepo{db: db}
}

type businessStatusPeriodRepo struct {
	db *gorm.DB
}

// GetOpenPeriod 获取当前未结束的测试营业时段
func (r *businessStatusPeriodRepo) GetOpenPeriod() (*model.BusinessStatusPeriod, error) {
	var period model.BusinessStatusPeriod
	err := r.db.Model(&model.BusinessStatusPeriod{}).
		Where("end_time = 0 AND delete_time = 0").
		Order("start_time DESC").
		Take(&period).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}
	return &period, nil
}

// GetBusinessStatus 获取当前营业状态，查询失败时默认返回正常营业
func (r *businessStatusPeriodRepo) GetBusinessStatus() int {
	openPeriod, _ := r.GetOpenPeriod()
	if openPeriod != nil {
		return constant.BusinessStatusTest
	}
	return constant.BusinessStatusNormal
}

// Create 创建时段记录
func (r *businessStatusPeriodRepo) Create(period *model.BusinessStatusPeriod) error {
	return r.db.Create(period).Error
}

// CloseOpenPeriod 关闭当前未结束的测试营业时段
func (r *businessStatusPeriodRepo) CloseOpenPeriod(endTime int64) error {
	return r.db.Model(&model.BusinessStatusPeriod{}).
		Where("end_time = 0 AND delete_time = 0").
		Update("end_time", endTime).Error
}
