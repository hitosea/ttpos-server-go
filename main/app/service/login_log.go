package service

import (
	"ttpos-server-go/pkg/database"
)

// ILoginLogSrv 多语言名称服务接口
type ILoginLogSrv interface {
	Save() error
}

// NewLoginLogSrv 登录日志服务
func NewLoginLogSrv(dbm *database.DBManager) ILoginLogSrv {
	return NewLoginLogServiceImpl(dbm)
}

type loginLogSrv struct {
	dbm *database.DBManager
}

func NewLoginLogServiceImpl(dbm *database.DBManager) ILoginLogSrv {
	return &loginLogSrv{
		dbm: dbm,
	}
}

func (s *loginLogSrv) Save() error {
	//return s.shopLoginLogRepo.Save(1, "", "", "")

	return nil
}
