package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type ShopOptLogRepository struct {
	dbm *database.DBManager
}

func NewShopOptLogRepository(dbm *database.DBManager) *ShopOptLogRepository {
	return &ShopOptLogRepository{dbm: dbm}
}

func (r *ShopOptLogRepository) Save(companyId uint, source string, key string, shopUserId uint) error {
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
