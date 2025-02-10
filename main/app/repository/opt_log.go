package repository

import (
	"ttpos-server-go/pkg/database"
)

type IOptLogRepo interface {
	Save(source string, key string, shopUserId uint) error
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

func (r *OptLogRepo) Save(source string, key string, shopUserId uint) error {
	return nil
}
