package old_model

import (
	"fmt"

	"gorm.io/gorm"
)

type UserPointsLog struct {
	ID          uint    `gorm:"primaryKey;autoIncrement;comment:主键id"`
	MemberUUID  string  `gorm:"not null;default:'';comment:会员uuid"`
	Scene       uint8   `gorm:"not null;default:10;comment:变动场景(10用户充值 20用户消费 30管理员操作 40订单退款)"`
	Value       float64 `gorm:"not null;default:0.00;comment:变动值"`
	Description string  `gorm:"size:500;not null;default:'';comment:描述/说明"`
	CreateTime  int64   `gorm:"not null;default:0;comment:创建时间"`
}

type UserPointsLogRepository interface {
	GetUserPointsLogList() ([]*UserPointsLog, error)
	ConvertUserPointsLog() error
}

type UserPointsLogService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserPointsLogService) GetUserPointsLogList() ([]*UserPointsLog, error) {
	var userPointsLogs []*UserPointsLog
	err := s.db.Find(&userPointsLogs).Error
	return userPointsLogs, err
}

func (s *UserPointsLogService) ConvertUserPointsLog() error {
	userPointsLogs, err := s.GetUserPointsLogList()
	if err != nil {
		return err
	}
	for _, userPointsLog := range userPointsLogs {
		fmt.Println(fmt.Sprintf("userPointsLog: %+v", userPointsLog))

	}
	return nil
}
