package v1

import (
	"fmt"
	"strconv"
	"strings"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type User struct {
	UserID         uint    `gorm:"primaryKey;autoIncrement;not null;comment:'用户id'"`
	OpenID         string  `gorm:"type:varchar(255);not null;default:'';comment:'微信openid(唯一标示)'"`
	MpOpenID       string  `gorm:"type:varchar(255);default:'';comment:'微信公众号openid'"`
	AppOpenID      string  `gorm:"type:varchar(255);default:'';comment:'openappid'"`
	UnionID        string  `gorm:"type:varchar(255);default:'';comment:'微信开放平台id'"`
	AlipayID       string  `gorm:"type:varchar(50);default:'0';comment:'支付宝用户id'"`
	RegSource      string  `gorm:"type:varchar(50);default:'';comment:'注册来源'"`
	NickName       string  `gorm:"column:nickName;type:varchar(255);not null;default:'';comment:'微信昵称'"`
	Mobile         string  `gorm:"type:varchar(20);default:'';comment:'手机号'"`
	Password       string  `gorm:"type:varchar(120);not null;default:'';comment:'密码'"`
	AvatarUrl      string  `gorm:"type:varchar(255);not null;default:'';comment:'微信头像'"`
	Gender         uint8   `gorm:"type:tinyint(3) unsigned;not null;default:2;comment:'性别0=女1=男2=未知'"`
	Country        string  `gorm:"type:varchar(50);not null;default:'';comment:'国家'"`
	Province       string  `gorm:"type:varchar(50);not null;default:'';comment:'省份'"`
	City           string  `gorm:"type:varchar(50);not null;default:'';comment:'城市'"`
	AddressID      uint    `gorm:"type:int(11) unsigned;not null;default:0;comment:'默认收货地址'"`
	Balance        float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'用户可用余额'"`
	GiftBalance    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'赠送余额'"`
	Points         float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'用户可用积分'"`
	PayMoney       float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'用户总支付的金额'"`
	ExpendMoney    float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'实际消费的金额(不含退款)'"`
	GradeID        uint    `gorm:"type:int(11) unsigned;not null;default:1;comment:'会员等级id'"`
	CardID         uint    `gorm:"type:int(11) unsigned;not null;default:0;comment:'会员卡id'"`
	RefereeID      int     `gorm:"type:int(11);not null;default:0;comment:'推荐人id'"`
	TotalPoints    float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'累计积分'"`
	TotalInvite    int     `gorm:"type:int(11);default:0;comment:'总邀请人数'"`
	UserType       uint8   `gorm:"type:tinyint(2);not null;default:1;comment:'供应商状态1普通用户2供应商'"`
	GiftMoney      int     `gorm:"type:int(11);default:0;comment:'虚拟币，刷礼物'"`
	GiftSupplierID string  `gorm:"type:varchar(255);not null;default:''" comment:"新人礼包门店id"`
	Birthday       int     `gorm:"type:int(11);not null;default:0;comment:'生日'"`
	ReceiveTime    int     `gorm:"type:int(11);not null;default:0;comment:'生日礼物领取时间'"`
	SendTime       int     `gorm:"type:int(11);not null;default:0;comment:'发送时间'"`
	FreezeMoney    float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'已冻结佣金'"`
	CashMoney      float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:'累积提现佣金'"`
	RealName       string  `gorm:"type:varchar(30);default:''; comment:'姓名'"`
	IsDelete       uint8   `gorm:"type:tinyint(3) unsigned;not null;default:0;comment:'是否删除'"`
	AppID          uint    `gorm:"type:int(11) unsigned;not null;default:0;comment:'小程序id'"`
	CreateTime     int64   `gorm:"type:int(11) unsigned;not null;default:0;comment:'创建时间'"`
	UpdateTime     int64   `gorm:"type:int(11) unsigned;not null;default:0;comment:'更新时间'"`
}

type UserRepository interface {
	GetUserList() ([]*User, error)
	GetUserByID(id uint) (*User, error)
	GetUserCardRecordByUserIdAndCardId(userId uint, cardId uint) (*UserCardRecord, error)
	GetUserAccumulatedConsumptionAmount(appId uint) (map[uint]*UserAccumulatedConsumptionAmount, error)
	ConvertUser() error
}

func NewUserService(db *gorm.DB, targetDB *gorm.DB) UserRepository {
	return &UserService{
		db:       db,
		targetDB: targetDB,
	}
}

type UserService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserService) GetUserCardRecordByUserIdAndCardId(userId uint, cardId uint) (*UserCardRecord, error) {
	var userCardRecord UserCardRecord
	err := s.db.Where("user_id = ? AND card_id = ?", userId, cardId).First(&userCardRecord).Error
	return &userCardRecord, errors.WithMessage(err)
}

func (s *UserService) GetUserList() ([]*User, error) {
	var users []*User
	err := s.db.Find(&users).Error
	return users, errors.WithMessage(err)
}

func (s *UserService) GetUserByID(id uint) (*User, error) {
	var user User
	err := s.db.Where("user_id = ?", id).First(&user).Error
	return &user, errors.WithMessage(err)
}

type UserAccumulatedConsumptionAmount struct {
	UserID            uint    `gorm:"column:user_id;not null;default:0;comment:'用户id'"`
	ConsumptionNum    int     `gorm:"column:consumption_num;not null;default:0;comment:'消费次数'"`
	ConsumptionAmount float64 `gorm:"column:consumption_amount;not null;default:0.00;comment:'消费金额'"`
	Nickname          string  `gorm:"column:nickname;not null;default:'';comment:'昵称'"`
}

// 查询用户的累计消费金额
func (s *UserService) GetUserAccumulatedConsumptionAmount(appId uint) (map[uint]*UserAccumulatedConsumptionAmount, error) {
	var consumptionAmounts []*UserAccumulatedConsumptionAmount
	query := `
		SELECT
			a.user_id,
			COUNT(a.order_id) AS consumption_num,
			SUM(a.pay_price) - SUM(a.refund_money) AS consumption_amount,
			b.nickname AS nickname 
		FROM
			jjjfood_order a
			LEFT JOIN jjjfood_user b ON a.user_id = b.user_id
		WHERE
			(a.app_id = ?
			AND a.order_status = '30'
			AND a.pay_status = '20'
			AND a.delete_time = '0'
			AND a.user_id > '0'
			AND b.is_delete = '0')
			AND a.delete_time = '0'
		GROUP BY
			a.user_id
		ORDER BY
			consumption_num DESC
	`
	err := s.db.Raw(query, appId).Scan(&consumptionAmounts).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	dataMap := make(map[uint]*UserAccumulatedConsumptionAmount)
	for _, consumptionAmount := range consumptionAmounts {
		dataMap[consumptionAmount.UserID] = consumptionAmount
	}
	return dataMap, nil
}

func (s *UserService) ConvertUser() error {
	users, err := s.GetUserList()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 获取用户累计消费金额
	consumptionAmounts := make(map[uint]*UserAccumulatedConsumptionAmount)
	if len(users) > 0 {
		appId := users[0].AppID
		consumptionAmounts, err = s.GetUserAccumulatedConsumptionAmount(appId)
		if err != nil {
			return errors.WithMessage(err)
		}
	}

	for _, user := range users {
		fmt.Println(fmt.Sprintf("user: %+v", user))

		var userCardRecord *UserCardRecord
		userCardRecord, err = s.GetUserCardRecordByUserIdAndCardId(user.UserID, user.CardID)
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				userCardRecord = nil
			} else {
				return errors.WithMessage(err)
			}
		}
		fmt.Println(fmt.Sprintf("userCardRecord: %+v", userCardRecord))

		var memberCard *model.MemberCard
		if userCardRecord != nil {
			memberCardUuid, _ := utils.GetID()
			memberCard = &model.MemberCard{
				BaseModel: model.BaseModel{
					Uuid:       memberCardUuid,
					CreateTime: int64(user.CreateTime),
					UpdateTime: int64(user.UpdateTime),
					DeleteTime: int64(user.IsDelete),
				},
				CardTypeUuid: uint64(user.CardID),
				MemberUuid:   uint64(user.UserID),
				ExpireTime:   int64(userCardRecord.ExpireTime),
				Discount:     userCardRecord.GetDiscount(),
			}
		}

		var consumptionAmount *UserAccumulatedConsumptionAmount
		consumptionAmount = consumptionAmounts[user.UserID]
		amount := float64(0)
		consumptionNum := 0
		if consumptionAmount != nil {
			amount = consumptionAmount.ConsumptionAmount
			consumptionNum = consumptionAmount.ConsumptionNum
		}
		member := model.Member{
			BaseModel: model.BaseModel{
				ID:         user.UserID,
				Uuid:       uint64(user.UserID),
				CreateTime: int64(user.CreateTime),
				UpdateTime: int64(user.UpdateTime),
				DeleteTime: int64(user.IsDelete),
			},
			MemberNo: strconv.FormatUint(uint64(user.UserID), 10),
			Nickname: user.NickName,
			Gender:   int(user.Gender),
			Phone:    user.Mobile,
			// Password:                     user.Password,
			Birthday:                     int64(user.Birthday),
			Point:                        user.Points,
			AccumulatedConsumptionAmount: amount,
			ConsumptionCount:             consumptionNum,
			Balance:                      user.Balance,
			GiftBalance:                  user.GiftBalance,
			MemberLevelUuid:              uint64(user.GradeID),
			MemberCardUuid:               uint64(user.CardID),
			MemberCard:                   memberCard,
		}
		err = repository.NewMemberRepo(s.targetDB).CreateMemberAndMemberCard(member)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
