package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

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

func NewUserPointsLogService(db *gorm.DB, targetDB *gorm.DB) UserPointsLogRepository {
	return &UserPointsLogService{
		db:       db,
		targetDB: targetDB,
	}
}

type UserPointsLogService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserPointsLogService) GetUserPointsLogList() ([]*UserPointsLog, error) {
	var userPointsLogs []*UserPointsLog
	err := s.db.Find(&userPointsLogs).Error
	return userPointsLogs, errors.WithMessage(err)
}

func (s *UserPointsLogService) ConvertUserPointsLog() error {
	return s.convertUserPointsLog(0, 2000)
}

func (s *UserPointsLogService) convertUserPointsLog(offset, limit int) error {
	var userPointsLogs []*UserPointsLog
	err := s.db.Offset(offset).Limit(limit).Find(&userPointsLogs).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	//
	memberPointLogs := []model.MemberPointLog{}
	for _, userPointsLog := range userPointsLogs {
		memberPointLogs = append(memberPointLogs, model.MemberPointLog{
			BaseModel: model.BaseModel{
				Uuid:       uint64(userPointsLog.LogID),
				CreateTime: int64(userPointsLog.CreateTime),
				UpdateTime: int64(userPointsLog.CreateTime),
			},
			MemberUuid: uint64(userPointsLog.UserID),
			Scene:      int(userPointsLog.Scene),
			Value:      userPointsLog.Value,
			Describe:   userPointsLog.Describe,
		})

	}
	// 保存数据
	if len(memberPointLogs) > 0 {
		fmt.Println(fmt.Sprintf("userPointsLog - num: %d", len(memberPointLogs)))
		if err := s.targetDB.Create(&memberPointLogs).Error; err != nil {
			return errors.WithMessage(err)
		}
		// 递归
		offset += limit
		return s.convertUserPointsLog(offset, limit)
	}
	return nil
}
