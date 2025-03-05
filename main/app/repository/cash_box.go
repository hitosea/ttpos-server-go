package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type ICashBoxRepo interface {
	Get() model.CashBox
	Create(cashBox model.CashBox) (model.CashBox, error)

	Update(uuid uint64, vars map[string]any) error
}

func NewCashBoxRepo(db *gorm.DB) ICashBoxRepo {
	return NewCashBoxRepoImpl(db)
}

type cashBoxRepo struct {
	db *gorm.DB
}

func NewCashBoxRepoImpl(db *gorm.DB) ICashBoxRepo {
	return &cashBoxRepo{db: db}
}

func (r *cashBoxRepo) Get() model.CashBox {
	var cashBox model.CashBox
	r.db.Model(&model.CashBox{}).First(&cashBox)
	return cashBox
}

func (r *cashBoxRepo) Create(cashBox model.CashBox) (model.CashBox, error) {
	err := r.db.Create(&cashBox).Error
	return cashBox, err
}

func (r *cashBoxRepo) Update(uuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.CashBox{}).Where("uuid = ?", uuid).Updates(vars).Error
	return errors.WithMessage(err)
}
