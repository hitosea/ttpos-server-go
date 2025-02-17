package service

import (
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
)

type ICashBoxSrv interface {
	UpdateBalance(companyUuid uint64)
}

func NewCashBoxSrv(dbm *database.DBManager) ICashBoxSrv {
	return NewCashBoxSrvImpl(dbm)
}

type CashBoxSrv struct {
	dbm      *database.DBManager
	uuidLock lock.Lock
}

func NewCashBoxSrvImpl(dbm *database.DBManager) *CashBoxSrv {
	return &CashBoxSrv{
		dbm:      dbm,
		uuidLock: lock.NewSystemLock(),
	}
}

// UpdateBalance 更新钱箱余额
func (s *CashBoxSrv) UpdateBalance(companyUuid uint64) {
	s.uuidLock.LockUuid(companyUuid)
	defer s.uuidLock.UnlockUuid(companyUuid)

}
