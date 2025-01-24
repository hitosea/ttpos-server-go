package repository

import (
	"ttpos-server-go/pkg/database"
)

type ShiftLogRepository struct {
	dbm *database.DBManager
}

func NewShiftLogRepository(dbm *database.DBManager) *ShiftLogRepository {
	return &ShiftLogRepository{dbm: dbm}
}
