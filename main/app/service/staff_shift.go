package service

import (
	"fmt"
	"math/rand"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
)

type ShiftService struct {
	shiftLogRepo *repository.ShiftLogRepository
	cache        cache.Cache
}

func NewShiftService(shiftLogRepo *repository.ShiftLogRepository, cache cache.Cache) *ShiftService {
	return &ShiftService{
		shiftLogRepo: shiftLogRepo,
		cache:        cache,
	}
}

// CreateWorkingLog 创建当班记录
func (s *ShiftService) CreateWorkingLog(staff model.Staff) model.StaffShiftLog {
	previousShiftCash, _ := s.shiftLogRepo.GetPreviousShiftCash(staff.CompanyId)
	startTime := staff.CashierLoginTime
	if startTime == 0 {
		startTime = int(time.Now().Unix())
	}
	shiftLog, _ := s.shiftLogRepo.Create(staff.CompanyId, model.StaffShiftLog{
		ShiftUserId:       int(staff.ID),
		ShiftNo:           s.generateNumber(),
		PreviousShiftCash: previousShiftCash,
		CurrentCashTotal:  previousShiftCash,
		CashLeft:          previousShiftCash,
		ShiftStartTime:    startTime,
		ShiftEndTime:      0,
	})
	return shiftLog
}

func (s *ShiftService) generateNumber() string {
	// 日期部分：年月日
	datePart := time.Now().Format("20060102")
	// 固定部分
	fixedPart := "01"
	// 随机部分：8位数字
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	no := datePart + fixedPart + randomPart
	cacheKey := "__USERSHIFTLOG_GENERATENUMBER__" + no
	if _, ok := s.cache.Get(cacheKey); ok {
		return s.generateNumber()
	}
	s.cache.Set(cacheKey, no, 86400*time.Second)
	// 组合订单号
	return no
}
