package v1

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Start of Selection
type UserRechargeOrder struct {
	ID             uint    `gorm:"primaryKey;autoIncrement;not null"`
	OrderNo        string  `gorm:"type:varchar(255);default:'';unique;comment:'订单号'"`
	OrderType      int     `gorm:"type:int;default:10;comment:'订单类型 10-充值订单'"`
	OrderStatus    int     `gorm:"type:int;default:0;comment:'订单状态 -1 -已取消 0-进行中 1-已完成'"`
	DutyNo         string  `gorm:"type:varchar(64);default:'';comment:'当班编号'"`
	PayTime        int     `gorm:"type:int;default:0;comment:'完成结账时间'"`
	DeviceID       string  `gorm:"type:varchar(200);default:'';comment:'设备ID'"`
	CashierID      string  `gorm:"type:varchar(200);default:'';comment:'收银员ID'"`
	UserID         int     `gorm:"type:int;default:0;comment:'会员ID'"`
	RechargeMoney  float64 `gorm:"type:decimal(12,2);default:0.00;comment:'充值金额'"`
	RefundMoney    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'退款金额'"`
	GiftMoney      float64 `gorm:"type:decimal(12,2);default:0.00;comment:'赠送金额'"`
	GiftPoint      float64 `gorm:"type:decimal(12,2);default:0.00"`
	OrderPrice     float64 `gorm:"type:decimal(12,2);default:0.00;comment:'订单金额'"`
	PayFee         float64 `gorm:"type:decimal(12,2);default:0.00;comment:'支付手续费'"`
	PayPrice       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'订单应付'"`
	ChangeDue      float64 `gorm:"type:decimal(12,2);default:0.00;comment:'找零'"`
	CancelRemark   string  `gorm:"type:varchar(255);default:'';comment:'取消备注'"`
	ShopSupplierID int     `gorm:"type:int;default:0;index;comment:'店铺id'"`
	AppID          int     `gorm:"type:int;default:0;index;comment:'应用id'"`
	CreateTime     int     `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime     int     `gorm:"type:int;not null;default:0;comment:'更新时间'"`

	PayTypes []UserRechargeOrderPayType `gorm:"foreignKey:OrderID;references:ID"`
}

func (model *UserRechargeOrder) GetOrderStatus() int {
	// 订单状态 -1 -已取消 0-进行中 1-已完成
	if model.OrderStatus == -1 {
		// 0-pending待支付 1-paid已支付 2-canceled已取消
		return 2
	}
	return model.OrderStatus
}

func (model *UserRechargeOrder) GetCashierID() (uint64, error) {
	if model.CashierID == "" {
		return 0, nil
	}
	uuid, err := strconv.Atoi(model.CashierID)
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return uint64(uuid), nil
}

// UserRechargeOrderPayType 订单支付方式表 `jjjfood_user_recharge_order_pay_type`
type UserRechargeOrderPayType struct {
	ID             uint    `gorm:"primaryKey;autoIncrement;not null"`
	OrderID        int     `gorm:"type:int;default:0;comment:'订单ID'"`
	PayStatus      int     `gorm:"type:int;default:0;comment:'支付状态 0-未支付 1-已支付'"`
	PaymentOrderID int     `gorm:"type:int;default:0;comment:'支付订单id'"`
	Value          int     `gorm:"type:int;default:0;comment:'支付方式'"`
	Price          float64 `gorm:"type:decimal(12,2);default:0.00;comment:'支付金额'"`
	Fee            int     `gorm:"type:int;default:0;comment:'支付费率0-100'"`
	FeeMoney       float64 `gorm:"type:decimal(12,2);default:0.00;comment:'单支付手续费'"`
	DisabledCancel int     `gorm:"type:int;default:0;comment:'禁止撤销 0-否 1-是'"`
	ShopSupplierID int     `gorm:"type:int;default:0;index;comment:'店铺id'"`
	AppID          int     `gorm:"type:int;default:0;index;comment:'应用id'"`
	CreateTime     int     `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime     int     `gorm:"type:int;not null;default:0;comment:'更新时间'"`

	PayType *PayType `gorm:"foreignKey:Value;references:Value"`
}

func (model *UserRechargeOrderPayType) GetFee() float64 {
	rate := decimal.NewFromInt(int64(model.Fee)).Div(decimal.NewFromInt(100)).Round(2).InexactFloat64()
	return rate
}

type UserRechargeOrderRepository interface {
	GetUserRechargeOrderList() ([]*UserRechargeOrder, error)
	ConvertUserRechargeOrder() error
}

func NewUserRechargeOrderService(db *gorm.DB, targetDB *gorm.DB) UserRechargeOrderRepository {
	return &UserRechargeOrderService{
		db:       db,
		targetDB: targetDB,
	}
}

type UserRechargeOrderService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserRechargeOrderService) GetUserRechargeOrderList() ([]*UserRechargeOrder, error) {
	var userRechargeOrders []*UserRechargeOrder
	err := s.db.Preload("PayTypes.PayType").Find(&userRechargeOrders).Error
	return userRechargeOrders, err
}

func (s *UserRechargeOrderService) ConvertUserRechargeOrder() error {
	userRechargeOrders, err := s.GetUserRechargeOrderList()
	if err != nil {
		return err
	}
	for _, order := range userRechargeOrders {
		fmt.Println(fmt.Sprintf("userRechargeOrder: %+v", order))

		paymentOrders := make([]model.PaymentOrder, 0)
		for _, payType := range order.PayTypes {
			paymentOrders = append(paymentOrders, model.PaymentOrder{
				BaseModel: model.BaseModel{
					Uuid:       uint64(payType.ID),
					CreateTime: int64(payType.CreateTime),
					UpdateTime: int64(payType.UpdateTime),
				},
				PaymentMethodName:    payType.PayType.Name,
				PaymentMethodUuid:    uint64(payType.PayType.ID),
				PaymentFeePercent:    payType.GetFee(),
				RelatedType:          1,
				RelatedUuid:          uint64(order.ID),
				CurrencyUnit:         "",
				PaymentAmount:        payType.Price,
				PaymentCommissionFee: payType.FeeMoney,
				Amount:               payType.Price + payType.FeeMoney,
				TransactionNumber:    "",
				Status:               payType.PayStatus,
				StatusReason:         "",
				BalanceAmount:        0,
				GiftBalanceAmount:    0,
			})
		}
		statffUuid, err := order.GetCashierID()
		if err != nil {
			return errors.WithMessage(err)
		}
		userRechargeOrder := model.MemberRechargeOrder{
			BaseModel: model.BaseModel{
				Uuid:       uint64(order.ID),
				CreateTime: int64(order.CreateTime),
				UpdateTime: int64(order.UpdateTime),
			},
			OrderNo:        order.OrderNo,
			DutyNo:         order.DutyNo,
			Status:         order.GetOrderStatus(),
			Amount:         order.OrderPrice,
			RefundMoney:    order.RefundMoney,
			ChargeDue:      order.ChangeDue,
			RechargeAmount: order.RechargeMoney,
			RefundAmount:   order.RefundMoney,
			GiftAmount:     order.GiftMoney,
			GiftPoint:      order.GiftPoint,
			MemberUuid:     uint64(order.UserID),
			StaffUuid:      statffUuid,
			PaymentTime:    int64(order.PayTime),
			PaymentOrders:  paymentOrders,
		}
		err = repository.NewMemberRechargeOrderRepo(s.targetDB).CreateOldRecord(userRechargeOrder)
		if err != nil {
			return err
		}
	}
	return nil
}
