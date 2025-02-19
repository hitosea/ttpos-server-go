package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ISaleBillRepo interface {
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)
	GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error)
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
