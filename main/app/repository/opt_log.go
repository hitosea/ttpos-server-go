package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type IOptLogRepo interface {
	Save(companyId uint, source string, key string, shopUserId uint) error
}

func NewOptLogRepo(dbm *database.DBManager) IOptLogRepo {
	return NewOptLogRepoImpl(dbm)
}

type OptLogRepo struct {
	dbm *database.DBManager
}

func NewOptLogRepoImpl(dbm *database.DBManager) *OptLogRepo {
	return &OptLogRepo{dbm: dbm}
}

func (r *OptLogRepo) Save(companyId uint, source string, key string, shopUserId uint) error {
	return r.dbm.GetDB(companyId).Debug().Create(&model.ShopOptLog{
		ShopUserId:  0,
		Title:       "",
		Url:         "",
		RequestType: "",
		Browser:     "",
		Agent:       "",
		Content:     "",
		Ip:          "",
		AppId:       0,
		CreateTime:  0,
	}).Error
}
