package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type ICashBoxLogRepo interface {
	Create(cashBox model.CashBoxLog) (model.CashBoxLog, error)
}

func NewCashBoxLogRepo(db *gorm.DB) ICashBoxLogRepo {
	return NewCashBoxLogRepoImpl(db)
}

type cashBoxLogRepo struct {
	db *gorm.DB
}

func NewCashBoxLogRepoImpl(db *gorm.DB) ICashBoxLogRepo {
	return &cashBoxLogRepo{db: db}
}

func (r *cashBoxLogRepo) Create(cashBoxLog model.CashBoxLog) (model.CashBoxLog, error) {
	err := r.db.Model(&model.CashBoxLog{}).Create(&cashBoxLog).Error
	return cashBoxLog, errors.WithMessage(err)
}
