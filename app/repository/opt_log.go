package repository

import (
	"jjjshop-server-go/app/model"
	"jjjshop-server-go/pkg/database"
)

type ShopOptLogRepository struct {
	dbm *database.DBManager
}

func NewShopOptLogRepository(dbm *database.DBManager) *ShopOptLogRepository {
	return &ShopOptLogRepository{dbm: dbm}
}

func (r *ShopOptLogRepository) Save(appId uint, source string, key string, shopUserId uint) error {
	return r.dbm.GetDB(appId).Debug().Create(&model.ShopOptLog{
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
