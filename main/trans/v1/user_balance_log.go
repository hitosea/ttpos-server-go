package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type UserBalanceLog struct {
	LogID           uint    `gorm:"primaryKey;autoIncrement;comment:主键id"`
	OrderID         int     `gorm:"default:0;comment:订单ID"`
	RechargeOrderID int     `gorm:"default:0;comment:会员充值订单ID"`
	UserID          uint    `gorm:"not null;default:0;comment:用户id"`
	CardID          int     `gorm:"not null;default:0;comment:会员卡id"`
	Scene           uint8   `gorm:"not null;default:10;comment:余额变动场景(10用户充值 20用户消费 30管理员操作 40订单退款)"`
	Money           float64 `gorm:"not null;default:0.00;comment:变动金额"`
	GiftMoney       float64 `gorm:"default:0.00;comment:赠送余额"`
	BeforeMoney     float64 `gorm:"default:0.00;comment:变更前金额"`
	AfterMoney      float64 `gorm:"default:0.00;comment:变更后金额"`
	Describe        string  `gorm:"size:500;not null;default:'';comment:描述/说明"`
	Remark          string  `gorm:"size:500;not null;default:'';comment:管理员备注"`
	Version         string  `gorm:"size:50;default:'1.0.7';comment:数据来源版本"`
	AppID           uint64  `gorm:"not null;default:0;comment:小程序商城id"`
	CreateTime      int64   `gorm:"not null;default:0;comment:创建时间"`
}

type UserBalanceLogRepository interface {
	GetUserBalanceLogList() ([]*UserBalanceLog, error)
	ConvertUserBalanceLog() error
}

func NewUserBalanceLogService(db *gorm.DB, targetDB *gorm.DB) UserBalanceLogRepository {
	return &UserBalanceLogService{
		db:       db,
		targetDB: targetDB,
	}
}

type UserBalanceLogService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserBalanceLogService) GetUserBalanceLogList() ([]*UserBalanceLog, error) {
	var userBalanceLogs []*UserBalanceLog
	err := s.db.Find(&userBalanceLogs).Error
	return userBalanceLogs, err
}

func (s *UserBalanceLogService) ConvertUserBalanceLog() error {
	userBalanceLogs, err := s.GetUserBalanceLogList()
	if err != nil {
		return err
	}
	for _, userBalanceLog := range userBalanceLogs {
		fmt.Println(fmt.Sprintf("userBalanceLog: %+v", userBalanceLog))
		log := model.MemberBalanceLog{
			BaseModel: model.BaseModel{
				Uuid:       uint64(userBalanceLog.LogID),
				CreateTime: int64(userBalanceLog.CreateTime),
				UpdateTime: int64(userBalanceLog.CreateTime),
			},
			MemberUuid: uint64(userBalanceLog.UserID),
			Scene:      int(userBalanceLog.Scene),
			Money:      userBalanceLog.Money,
			GiftMoney:  userBalanceLog.GiftMoney,
			Describe:   userBalanceLog.Describe,
		}
		_, err := base.NewMemberBalanceLogRepo(s.targetDB).CreateMemberBalanceLog(log)
		if err != nil {
			return err
		}
	}
	return nil
}
