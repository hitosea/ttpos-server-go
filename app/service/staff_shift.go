package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
)

type ShiftService struct {
	shiftLogRepo *repository.ShiftLogRepository
}

func NewShiftService(shiftLogRepo *repository.ShiftLogRepository) *ShiftService {
	return &ShiftService{
		shiftLogRepo: shiftLogRepo,
	}
}

func (s *ShiftService) CreateWorkingLog(staff model.Staff) model.StaffShiftLog {

	return model.StaffShiftLog{}
}
