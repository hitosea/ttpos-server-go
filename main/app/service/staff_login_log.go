package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IStaffLoginLogSrv interface {
	Add(ctx context.Context, staffUuid uint64, username string, ip string, result string) error
}

func NewStaffLoginLogSrv(dbm *database.DBManager) IStaffLoginLogSrv {
	return NewStaffLoginLogSrvImpl(dbm)
}

type staffLoginLogSrv struct {
	dbm *database.DBManager
}

func NewStaffLoginLogSrvImpl(dbm *database.DBManager) IStaffLoginLogSrv {
	return &staffLoginLogSrv{
		dbm: dbm,
	}
}

// Add 添加登录日志
func (s *staffLoginLogSrv) Add(ctx context.Context, staffUuid uint64, username string, ip string, result string) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	return db.Create(&model.StaffLoginLog{
		StaffUuid: staffUuid,
		Username:  username,
		Ip:        ip,
		Result:    result,
	}).Error

}
