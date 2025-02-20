package repository

import (
	"gorm.io/gorm"
)

type IOptLogRepo interface {
	Save(source string, key string, shopUserId uint) error
}

func NewOptLogRepo(db *gorm.DB) IOptLogRepo {
	return NewOptLogRepoImpl(db)
}

type OptLogRepo struct {
	db *gorm.DB
}

func NewOptLogRepoImpl(db *gorm.DB) *OptLogRepo {
	return &OptLogRepo{db: db}
}

func (r *OptLogRepo) Save(source string, key string, shopUserId uint) error {
	return nil
}
