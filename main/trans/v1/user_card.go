package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type UserCard struct {
	CardID        int     `gorm:"primaryKey;autoIncrement;not null" json:"card_id" comment:"ID"`
	CardName      string  `gorm:"type:varchar(100);not null;default:''" json:"card_name" comment:"会员卡名称"`
	CardStyle     string  `gorm:"type:varchar(255);not null;default:''" json:"card_style" comment:"样式"`
	Sort          int     `gorm:"type:int;not null;default:0" json:"sort" comment:"排序"`
	IsDiscount    int     `gorm:"type:tinyint(1);not null;default:0" json:"is_discount" comment:"会员权益0无折扣1会员折扣"`
	Discount      float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"discount" comment:"折扣"`
	OpenPoints    int     `gorm:"type:tinyint(1);not null;default:0" json:"open_points" comment:"开卡赠送积分0否1是"`
	OpenPointsNum float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"open_points_num" comment:"开卡赠送积分数"`
	OpenCoupon    int     `gorm:"type:tinyint(1);not null;default:0" json:"open_coupon" comment:"开卡赠送优惠券0否1是"`
	OpenCoupons   string  `gorm:"type:longtext;not null" json:"open_coupons" comment:"每月赠送项目券"`
	OpenMoney     int     `gorm:"type:tinyint(1);not null;default:0" json:"open_money" comment:"开卡赠送余额0否1是"`
	OpenMoneyNum  float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"open_money_num" comment:"开卡赠送余额数"`
	Expire        int     `gorm:"type:int;not null;default:0" json:"expire" comment:"有效期(月)0永久有效"`
	Money         float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"money" comment:"价格"`
	Status        int     `gorm:"type:tinyint(1);not null;default:0" json:"status" comment:"启用0关闭1"`
	Content       string  `gorm:"type:varchar(255);not null;default:''" json:"content" comment:"使用须知"`
	ReceiveNum    int     `gorm:"type:int;not null;default:0" json:"receive_num" comment:"领取人数"`
	IsDefault     int     `gorm:"type:tinyint(1);not null;default:0" json:"is_default" comment:"默认样式0自定义1"`
	DefaultStyle  string  `gorm:"type:varchar(255);default:''" json:"default_style" comment:"自定义背景"`
	IsDelete      int     `gorm:"type:tinyint(3) unsigned;not null;default:0" json:"is_delete" comment:"是否删除"`
	AppID         int     `gorm:"type:int(11) unsigned;not null;default:0" json:"app_id" comment:"小程序id"`
	CreateTime    int64   `gorm:"type:int(11) unsigned;not null;default:0" json:"create_time" comment:"创建时间"`
	UpdateTime    int64   `gorm:"type:int(11) unsigned;not null;default:0" json:"update_time" comment:"更新时间"`
}

type UserCardRepository interface {
	GetUserCardList() ([]*UserCard, error)
	GetUserCardByID(id int) (*UserCard, error)
	ConvertUserCard() error
}

type UserCardService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserCardService) GetUserCardList() ([]*UserCard, error) {
	var userCards []*UserCard
	err := s.db.Find(&userCards).Error
	return userCards, err
}

func (s *UserCardService) GetUserCardByID(id int) (*UserCard, error) {
	var userCard UserCard
	err := s.db.Where("card_id = ?", id).First(&userCard).Error
	return &userCard, err
}

func (s *UserCardService) ConvertUserCard() error {
	userCards, err := s.GetUserCardList()
	if err != nil {
		return err
	}
	for _, userCard := range userCards {
		fmt.Println(fmt.Sprintf("userCard: %+v", userCard))

		memberCardType := model.MemberCardType{
			BaseModel: model.BaseModel{
				Uuid:       uint64(userCard.CardID),
				CreateTime: int64(userCard.CreateTime),
				UpdateTime: int64(userCard.UpdateTime),
				DeleteTime: int64(userCard.IsDelete),
			},
			Name:         userCard.CardName,
			Expire:       userCard.Expire,
			Price:        userCard.Money,
			Discount:     int(userCard.Discount),
			Sort:         userCard.Sort,
			Status:       userCard.Status,
			OpenMoney:    userCard.OpenMoney,
			OpenMoneyNum: userCard.OpenMoneyNum,
			OpenPoint:    userCard.OpenPoints,
			OpenPointNum: userCard.OpenPointsNum,
			Describe:     userCard.Content,
		}
		_, err := base.NewMemberCardTypeRepo(s.targetDB).CreateMemberCardType(memberCardType)
		if err != nil {
			return err
		}
	}
	return nil
}
