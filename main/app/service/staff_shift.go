package service

import (
	"fmt"
	"math/rand"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

type IStaffShiftSrv interface {
	CreateWorkingLog(staff model.Staff) model.StaffShiftLog
}

func NewStaffShiftSrv(cache cache.Cache, dbm *database.DBManager) IStaffShiftSrv {
	return NewShiftSrvImpl(cache, dbm)
}

type StaffShiftSrv struct {
	dbm            *database.DBManager
	cache          cache.Cache
	cacheKeyPrefix string
}

func NewShiftSrvImpl(cache cache.Cache, dbm *database.DBManager) *StaffShiftSrv {
	return &StaffShiftSrv{
		dbm:            dbm,
		cache:          cache,
		cacheKeyPrefix: "__USERSHIFTLOG_GENERATENUMBER__",
	}
}

// CreateWorkingLog 创建当班记录
func (s *StaffShiftSrv) CreateWorkingLog(staff model.Staff) model.StaffShiftLog {
	shiftLogRepo := repository.NewShiftLogRepo(s.dbm.GetDB(staff.CompanyUuid))
	previousShiftCash, _ := shiftLogRepo.GetPreviousShiftCash(staff.CompanyUuid)
	startTime := staff.CashierLoginTime
	if startTime == 0 {
		startTime = int(time.Now().Unix())
	}
	shiftLog, _ := shiftLogRepo.Create(staff.CompanyUuid, model.StaffShiftLog{
		StaffUuid:         staff.Uuid,
		ShiftNo:           s.generateNumber(),
		PreviousShiftCash: previousShiftCash,
		CurrentCashTotal:  previousShiftCash,
		CashLeft:          previousShiftCash,
		ShiftStartTime:    startTime,
		ShiftEndTime:      0,
	})
	return shiftLog
}

func (s *StaffShiftSrv) generateNumber() string {
	// 日期部分：年月日
	datePart := time.Now().Format("20060102")
	// 固定部分
	fixedPart := "01"
	// 随机部分：8位数字
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	no := datePart + fixedPart + randomPart
	cacheKey := s.cacheKeyPrefix + no
	if _, ok := s.cache.Get(cacheKey); ok {
		return s.generateNumber()
	}
	s.cache.Set(cacheKey, no, 86400*time.Second)
	// 组合订单号
	return no
}
