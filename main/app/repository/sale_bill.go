package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleBillRepo interface {
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)
	GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error)
	GetSaleBillByDeviceUuid(deviceSn uint64) (*model.SaleBill, error)
	UpdateSaleBill(saleBill *model.SaleBill) error
	UpdateSaleBillShowMustPlan(saleBillUuid uint64) error
}

type saleBillRepo struct {
	db *gorm.DB
}

func NewSaleBillRepo(db *gorm.DB) ISaleBillRepo {
	return &saleBillRepo{db: db}
}

func (r *saleBillRepo) GetSaleBill(opts ...DBOption) (model.SaleBill, error) {
	var saleBill model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&saleBill)
	if result.Error != nil {
		return saleBill, result.Error
	}

	return saleBill, nil
}

func (r *saleBillRepo) GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(CommonRepo.WhereByUuid(uuid))
	if err != nil {
		return nil, err
	}
	return &saleBill, nil
}

// 通过deviceSn查询点餐页面未挂单的账单
func (r *saleBillRepo) GetSaleBillByDeviceUuid(deviceUuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("device_uuid = ? AND hide_bill_time = 0", deviceUuid)
		})
	if err != nil {
		return nil, err
	}
	return &saleBill, nil
}

func (r *saleBillRepo) UpdateSaleBill(saleBill *model.SaleBill) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBill.Uuid).Updates(saleBill).Error
}

func (r *saleBillRepo) UpdateSaleBillShowMustPlan(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Update("show_must_plan", constant.SaleBillShowMustPlanNo).Error
}
