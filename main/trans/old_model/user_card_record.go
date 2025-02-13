package old_model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type UserCardRecord struct {
	OrderID       int     `gorm:"primaryKey;autoIncrement;not null" json:"order_id" comment:"ID"`
	UserID        int     `gorm:"not null" json:"user_id" comment:"会员id"`
	CardID        int     `gorm:"not null" json:"card_id" comment:"会员卡id"`
	ExpireTime    int64   `gorm:"not null;default:0" json:"expire_time" comment:"有效期"`
	PayPrice      float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"pay_price" comment:"价格"`
	Discount      float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"discount" comment:"会员折扣"`
	PayType       int     `gorm:"type:tinyint(3) unsigned;not null;default:20" json:"pay_type" comment:"支付方式(10余额支付 20微信支付 30支付宝支付 40后台手动发卡)"`
	PayTime       int     `gorm:"not null;default:0" json:"pay_time" comment:"付款时间"`
	OpenPoints    int     `gorm:"type:tinyint(1);not null;default:0" json:"open_points" comment:"开卡赠送积分0否1是"`
	OpenPointsNum float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"open_points_num" comment:"开卡赠送积分数"`
	OpenCoupon    int     `gorm:"type:tinyint(1);not null;default:0" json:"open_coupon" comment:"开卡赠送优惠券0否1是"`
	OpenCoupons   string  `gorm:"type:longtext;not null" json:"open_coupons" comment:"每月赠送项目券"`
	OpenMoney     int     `gorm:"type:tinyint(1);not null;default:0" json:"open_money" comment:"开卡赠送余额0否1是"`
	OpenMoneyNum  float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"open_money_num" comment:"开卡赠送余额数"`
	PayStatus     int     `gorm:"type:tinyint(3) unsigned;not null;default:10" json:"pay_status" comment:"支付状态(10待支付 20已支付)"`
	OrderNo       string  `gorm:"type:varchar(20);not null;default:''" json:"order_no" comment:"订单号"`
	TransactionID string  `gorm:"type:varchar(30);not null;default:''" json:"transaction_id" comment:"微信支付交易号"`
	Balance       float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"balance" comment:"余额抵扣金额"`
	OnlineMoney   float64 `gorm:"type:decimal(12,2);not null;default:0.00" json:"online_money" comment:"在线支付金额"`
	TradeNo       string  `gorm:"type:varchar(30);not null;default:''" json:"trade_no" comment:"支付订单号"`
	IsDelete      int     `gorm:"type:tinyint(3) unsigned;not null;default:0" json:"is_delete" comment:"是否删除"`
	AppID         int     `gorm:"type:int(11) unsigned;not null;default:0" json:"app_id" comment:"小程序id"`
	CreateTime    int     `gorm:"type:int(11) unsigned;not null;default:0" json:"create_time" comment:"创建时间"`
	UpdateTime    int     `gorm:"type:int(11) unsigned;not null;default:0" json:"update_time" comment:"更新时间"`
}

type UserCardRecordRepository interface {
	GetUserCardRecordList() ([]*UserCardRecord, error)
	ConvertUserCardRecord() error
}

type UserCardRecordService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserCardRecordService) GetUserCardRecordList() ([]*UserCardRecord, error) {
	var userCardRecords []*UserCardRecord
	err := s.db.Find(&userCardRecords).Error
	return userCardRecords, err
}

func (s *UserCardRecordService) ConvertUserCardRecord() error {
	userCardRecords, err := s.GetUserCardRecordList()
	if err != nil {
		return err
	}
	for _, userCardRecord := range userCardRecords {
		fmt.Println(fmt.Sprintf("userCardRecord: %+v", userCardRecord))

		userService := UserService{
			db:       s.db,
			targetDB: s.targetDB,
		}
		user, err := userService.GetUserByID(uint(userCardRecord.UserID))
		if err != nil {
			return err
		}
		json, _ := json.Marshal(user)
		fmt.Println("NickName:", user.NickName, "===========", string(json))
		userCardService := UserCardService{
			db:       s.db,
			targetDB: s.targetDB,
		}
		userCard, err := userCardService.GetUserCardByID(userCardRecord.CardID)
		if err != nil {
			return err
		}
		memberCardLog := model.MemberCardLog{
			BaseModel: model.BaseModel{
				Uuid:       uint64(userCardRecord.OrderID),
				CreateTime: int64(userCardRecord.CreateTime),
				UpdateTime: int64(userCardRecord.UpdateTime),
				DeleteTime: int64(userCardRecord.IsDelete),
			},
			Price:              userCardRecord.PayPrice,
			Discount:           int(userCardRecord.Discount),
			Period:             int(userCardRecord.ExpireTime),
			MemberName:         user.NickName,
			MemberPhone:        user.Mobile,
			MemberNo:           strconv.Itoa(userCardRecord.UserID),
			MemberCardTypeName: userCard.CardName,
			MemberCardTypeUuid: uint64(userCardRecord.CardID),
			MemberUuid:         uint64(userCardRecord.UserID),
		}
		_, err = base.NewMemberCardLogRepo(s.targetDB).CreateMemberCardLog(memberCardLog)
		if err != nil {
			return err
		}
	}
	return nil
}
