package repository

import (
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/pkg/database"
)

type LoginLogRepository struct {
	dbm *database.DBManager
}

func NewLoginLogRepository(dbm *database.DBManager) *LoginLogRepository {
	return &LoginLogRepository{dbm: dbm}
}

func (r *LoginLogRepository) Save(appId uint, username, ip, result string) error {
	return r.dbm.GetDB(appId).Debug().Create(&model.LoginLog{
		Username: username,
		Ip:       ip,
		Result:   result,
		AppId:    appId,
	}).Error
}
