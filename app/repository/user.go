package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

type UserRepository struct {
	dbm *database.DBManager
}

func NewUserRepository(db *database.DBManager) *UserRepository {
	return &UserRepository{dbm: db}
}

func (r *UserRepository) GetById(Id uint, withs ...Where) model.User {
	var user model.User
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).First(&user, Id)
	return user
}

func (r *UserRepository) WithSupplier() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Supplier")
	}
}

func (r *UserRepository) WithApp() Where {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("App")
	}
}

func (r *UserRepository) GetByUsername(username string, withs ...Where) model.User {
	var user model.User
	db := r.dbm.GetDB(constant.DefaultDB)
	r.handleWiths(db, withs).Where("user_name = ?", username).Debug().First(&user)
	return user
}

func (r *UserRepository) GetCurrentCashier(bindKey string) model.User {
	var user model.User
	r.dbm.GetDB(constant.DefaultDB).Where("bind_key = ? AND cashier_online = 1", bindKey).Debug().First(&user)
	return user
}

func (r *UserRepository) handleWiths(db *gorm.DB, withs []Where) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}
