package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type LoginLogRepository struct {
	dbm *database.DBManager
}

func NewLoginLogRepository(dbm *database.DBManager) *LoginLogRepository {
	return &LoginLogRepository{dbm: dbm}
}

func (r *LoginLogRepository) Save(companyId uint, username, ip, result string) error {
	return r.dbm.GetDB(companyId).Debug().Create(&model.LoginLog{
		Username: username,
		Ip:       ip,
		Result:   result,
		AppId:    companyId,
	}).Error
}
