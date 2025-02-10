package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ILoginLogRepo interface {
	Save(companyId uint64, username, ip, result string) error
}

func NewLoginLogRepo(dbm *database.DBManager) ILoginLogRepo {
	return NewLoginLogRepoImpl(dbm)
}

type LoginLogRepo struct {
	dbm *database.DBManager
}

func NewLoginLogRepoImpl(dbm *database.DBManager) *LoginLogRepo {
	return &LoginLogRepo{dbm: dbm}
}

func (r *LoginLogRepo) Save(companyId uint64, username, ip, result string) error {
	return r.dbm.GetDB(companyId).Debug().Create(&model.LoginLog{
		Uuid:       0, // todo 生成uuid
		StaffUuid:  0,
		Username:   "",
		Ip:         "",
		Result:     "",
		CreateTime: 0,
	}).Error
}
