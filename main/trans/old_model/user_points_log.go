package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type UserPointsLog struct {
	LogID      uint    `gorm:"column:log_id;primaryKey;autoIncrement;comment:主键id"`
	UserID     uint    `gorm:"column:user_id;not null;default:0;comment:用户id"`
	CardID     int     `gorm:"column:card_id;not null;default:0;comment:会员卡id"`
	OrderID    uint    `gorm:"column:order_id;not null;default:0;comment:订单id"`
	Scene      int     `gorm:"column:scene;not null;default:0;comment:积分变动场景(10充值 20消费赠送 30管理员操作 40退款扣除)"`
	Value      float64 `gorm:"column:value;not null;default:0.00;comment:变动积分"`
	Describe   string  `gorm:"column:describe;not null;default:'';comment:描述/说明"`
	Remark     string  `gorm:"column:remark;not null;default:'';comment:管理员备注"`
	AppID      uint    `gorm:"column:app_id;not null;default:0;comment:小程序商城id"`
	CreateTime uint    `gorm:"column:create_time;not null;default:0;comment:创建时间"`
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
		memberPointLog := model.MemberPointLog{
			Uuid:        uint64(userPointsLog.LogID),
			MemberUuid:  uint64(userPointsLog.UserID),
			Scene:       int(userPointsLog.Scene),
			Value:       int(userPointsLog.Value),
			Description: userPointsLog.Describe,
			CreateTime:  int64(userPointsLog.CreateTime),
			UpdateTime:  int64(userPointsLog.CreateTime),
			DeleteTime:  0,
		}
		_, err := base.NewMemberPointLogRepo(s.targetDB).CreateMemberPointLog(memberPointLog)
		if err != nil {
			return err
		}
	}
	return nil
}
