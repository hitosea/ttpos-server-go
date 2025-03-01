package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type ILoginLogRepo interface {
	Save(companyId uint64, username, ip, result string) error
}

func NewLoginLogRepo(db *gorm.DB) ILoginLogRepo {
	return NewLoginLogRepoImpl(db)
}

type LoginLogRepo struct {
	db *gorm.DB
}

func NewLoginLogRepoImpl(db *gorm.DB) *LoginLogRepo {
	return &LoginLogRepo{db: db}
}

func (r *LoginLogRepo) Save(companyId uint64, username, ip, result string) error {
	return r.db.Model(&model.StaffLoginLog{}).Create(&model.StaffLoginLog{
		StaffUuid: 0, // todo
		Username:  "",
		Ip:        "",
		Result:    "",
	}).Error
}
