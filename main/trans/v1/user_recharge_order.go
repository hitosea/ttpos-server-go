package v1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"

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

	PayTypes      []UserRechargeOrderPayType      `gorm:"foreignKey:OrderID;references:ID"`
	OperationLogs []UserRechargeOrderOperationLog `gorm:"foreignKey:OrderID;references:ID"`
	OrderRefunds  []UserRechargeOrderRefund       `gorm:"foreignKey:OrderID;references:ID"`
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

type UserRechargeOrderOperationLog struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;not null"`
	OrderID        int    `gorm:"type:int;default:0;comment:'订单ID'"`
	Source         string `gorm:"type:varchar(150);not null;default:'';comment:'来源 cashier-收银 assistant-助手 shop-商家后台'"`
	ShopUserID     int    `gorm:"type:int;default:0;comment:'操作用户id'"`
	Action         string `gorm:"type:varchar(150);not null;default:'';comment:'行为'"`
	Data           string `gorm:"type:text;default:'';comment:'数据'"`
	Remark         string `gorm:"type:varchar(255);not null;default:'';comment:'备注'"`
	ShopSupplierID int    `gorm:"type:int;default:0;comment:'门店id'"`
	AppID          int    `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime     int    `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime     int    `gorm:"type:int;not null;default:0;comment:'更新时间'"`

	ShopUser *ShopUser `gorm:"foreignKey:ShopUserID;references:ShopUserID"`
}

type PayTypeStruct struct {
	OrderID        int     `json:"order_id"`
	RefundID       int     `json:"refund_id"`
	Price          float64 `json:"price"`
	Value          int     `json:"value"`
	Name           string  `json:"name"`
	Remark         string  `json:"remark"`
	Source         int     `json:"source"`
	PaymentOrderID int     `json:"payment_order_id"`
	RefundMoney    string  `json:"refund_money"`
	ShopSupplierID int     `json:"shop_supplier_id"`
	AppID          int     `json:"app_id"`
	Status         int     `json:"status"`
}

func (model *PayTypeStruct) GetRefundMoney() float64 {
	refundMoney, err := strconv.ParseFloat(model.RefundMoney, 64)
	if err != nil {
		return 0
	}
	return refundMoney
}

type UserRechargeOrderData struct {
	PayType      []PayTypeStruct `json:"pay_type"`
	RefundType   int             `json:"refund_type"`
	RefundMethod int             `json:"refund_method"`
	RefundMoney  float64         `json:"refund_money"`
}

func (model *UserRechargeOrderOperationLog) GetData() (string, error) {
	if model.Action == "REFUND" {
		return model.GetDataRefund()
	}
	return model.Data, nil
}

// 转换退款的日志数据
func (model *UserRechargeOrderOperationLog) GetDataRefund() (string, error) {
	if model.Data == "" {
		return "", nil
	}
	if model.Data == "[]" {
		return "", nil
	}
	if len(model.Data) < 5 {
		return "", nil
	}

	var data UserRechargeOrderData
	fmt.Println(fmt.Sprintf("model.Data: %s", model.Data))
	err := json.Unmarshal([]byte(model.Data), &data)
	if err != nil {
		return "", errors.WithMessage(err)
	}

	refundPayTypes := make([]RefundPayType, 0)
	for _, payType := range data.PayType {
		refundPayTypes = append(refundPayTypes, RefundPayType{
			Name:         payType.Name,
			Code:         payType.Value,
			Amount:       payType.GetRefundMoney(),
			RefundStatus: 0,
		})
	}
	refundData := RefundData{
		RefundType:     data.RefundType,
		RefundMoney:    data.RefundMoney,
		RefundPayTypes: refundPayTypes,
	}
	stringData := utils.ToJson(refundData)
	return stringData, nil
}

type RefundPayType struct {
	Name              string  `json:"name"`
	Code              int     `json:"code"`
	Amount            float64 `json:"amount"`
	RefundStatus      int     `json:"refund_status"`
	ReturnAmountUUID  int64   `json:"return_amount_uuid"`
	ReturnOrderUUID   int64   `json:"return_order_uuid"`
	PaymentOrderUUID  int     `json:"payment_order_uuid"`
	PaymentMethodUUID int     `json:"payment_method_uuid"`
}

type RefundData struct {
	RefundType     int             `json:"refund_type"`
	RefundMoney    float64         `json:"refund_money"`
	RefundPayTypes []RefundPayType `json:"refund_pay_types"`
}

type UserRechargeOrderRefund struct {
	ID             uint    `gorm:"primaryKey;autoIncrement;not null"`
	OrderID        int     `gorm:"type:int;default:0;comment:'订单ID'"`
	CashierID      int     `gorm:"type:int;default:0;comment:'收银员ID'"`
	RefundType     int     `gorm:"type:int;default:0;comment:'退款类型 1-整单 2-部分'"`
	RefundMethod   int     `gorm:"type:int;default:0;comment:'退款方式 1-系统退款 2-线下退款'"`
	RefundMoney    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'支付金额'"`
	ShopSupplierID int     `gorm:"type:int;default:0;comment:'店铺id'"`
	AppID          int     `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime     int     `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime     int     `gorm:"type:int;not null;default:0;comment:'更新时间'"`

	RefundDestinations []UserRechargeOrderRefundDestination `gorm:"foreignKey:RefundID;references:ID"`
}
type UserRechargeOrderRefundDestination struct {
	ID                    uint    `gorm:"primaryKey;autoIncrement;not null"`
	OrderID               int     `gorm:"type:int;default:0;comment:'订单ID'"`
	RefundID              int     `gorm:"type:int;default:0;comment:'订单退款ID'"`
	PaymentOrderID        int     `gorm:"type:int;default:0;comment:'支付订单ID'"`
	RefundOrderID         string  `gorm:"type:varchar(100);default:'';comment:'退款订单ID'"`
	MerchantRefundOrderNo string  `gorm:"type:varchar(100);default:'';comment:'商户退款单号'"`
	Status                int     `gorm:"type:int;default:1;comment:'退款状态 -1-失败 0-处理中 1-完成'"`
	Value                 int     `gorm:"type:int;default:0;comment:'支付方式'"`
	Price                 float64 `gorm:"type:decimal(12,2);default:0.00;comment:'支付金额'"`
	RefundMoney           float64 `gorm:"type:decimal(12,2);default:0.00;comment:'退款金额'"`
	BankCode              string  `gorm:"type:varchar(100);default:'';comment:'银行代码'"`
	AccountNo             string  `gorm:"type:varchar(100);default:'';comment:'银行卡号'"`
	AccountName           string  `gorm:"type:varchar(100);default:'';comment:'银行开户人名'"`
	Remark                string  `gorm:"type:varchar(255);not null;default:'';comment:'备注'"`
	ShopSupplierID        int     `gorm:"type:int;default:0;comment:'门店id'"`
	AppID                 int     `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime            int     `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime            int     `gorm:"type:int;not null;default:0;comment:'更新时间'"`

	PayType *PayType `gorm:"foreignKey:Value;references:Value"`
}

// 0-退款中 1-退款成功 2-退款失败
// -1-失败 0-处理中 1-完成
func (model *UserRechargeOrderRefundDestination) GetStatus() int {
	if model.Status == -1 {
		return 2
	}
	return model.Status
}

type UserRechargeOrderRepository interface {
	GetUserRechargeOrderList() ([]*UserRechargeOrder, error)
	GeUserRechargeOrderPayType(orderId int, value int) (*UserRechargeOrderPayType, error)
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
	err := s.db.Preload("PayTypes.PayType").Preload("OperationLogs.ShopUser").Preload("OrderRefunds.RefundDestinations.PayType").Find(&userRechargeOrders).Error
	return userRechargeOrders, err
}

func (s *UserRechargeOrderService) GeUserRechargeOrderPayType(orderId int, value int) (*UserRechargeOrderPayType, error) {
	var userRechargeOrderPayTypes *UserRechargeOrderPayType
	err := s.db.Preload("PayType").Where("order_id = ? AND value = ?", orderId, value).First(&userRechargeOrderPayTypes).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return userRechargeOrderPayTypes, nil
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

		operationLogs := make([]model.MemberRechargeOrderOperationLog, 0)
		for _, operationLog := range order.OperationLogs {
			data, err := operationLog.GetData()
			if err != nil {
				return errors.WithMessage(err)
			}
			operationLogs = append(operationLogs, model.MemberRechargeOrderOperationLog{
				BaseModel: model.BaseModel{
					Uuid:       uint64(operationLog.ID),
					CreateTime: int64(operationLog.CreateTime),
					UpdateTime: int64(operationLog.UpdateTime),
				},
				OperatorName:      operationLog.ShopUser.RealName,
				OperatorEmail:     operationLog.ShopUser.UserName,
				Client:            operationLog.Source,
				Message:           operationLog.Remark,
				Action:            operationLog.Action,
				Data:              data,
				RechargeOrderUuid: uint64(order.ID),
			})
		}

		returnOrders := make([]model.ReturnOrder, 0)
		for _, orderRefund := range order.OrderRefunds {
			returnOrderAmounts := make([]model.ReturnOrderAmount, 0)
			for _, refundDestination := range orderRefund.RefundDestinations {
				payType, err := s.GeUserRechargeOrderPayType(orderRefund.OrderID, refundDestination.Value)
				if err != nil {
					return errors.WithMessage(err)
				}
				returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
					BaseModel: model.BaseModel{
						Uuid:       uint64(refundDestination.ID),
						CreateTime: int64(refundDestination.CreateTime),
						UpdateTime: int64(refundDestination.UpdateTime),
					},
					Amount:                refundDestination.RefundMoney,
					RefundStatus:          refundDestination.GetStatus(),
					ReturnOrderUuid:       uint64(refundDestination.RefundID),
					PaymentMethodUuid:     uint64(refundDestination.PayType.ID),
					PaymentOrderUuid:      uint64(payType.ID),
					MerchantRefundOrderNo: "",
					LlReturnOrderid:       "",
				})
			}

			returnOrders = append(returnOrders, model.ReturnOrder{
				BaseModel: model.BaseModel{
					Uuid:       uint64(orderRefund.ID),
					CreateTime: int64(orderRefund.CreateTime),
					UpdateTime: int64(orderRefund.UpdateTime),
				},
				RelatedOrderType:    1,
				RelatedOrderUuid:    uint64(orderRefund.OrderID),
				RelatedOrderNo:      order.OrderNo,
				IsReverseSettlement: 0,
				ReturnType:          uint(orderRefund.RefundType),
				RefundAmount:        orderRefund.RefundMoney,
				Unit:                "",
				RefundTaxAmount:     0,
				RefundReason:        "",
				BankCode:            "",
				AccountNo:           "",
				AccountName:         "",
				ReturnOrderAmounts:  returnOrderAmounts,
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
			OrderNo:                    order.OrderNo,
			DutyNo:                     order.DutyNo,
			Status:                     order.GetOrderStatus(),
			Amount:                     order.OrderPrice,
			RefundMoney:                order.RefundMoney,
			ChargeDue:                  order.ChangeDue,
			RechargeAmount:             order.RechargeMoney,
			RefundAmount:               order.RefundMoney,
			GiftAmount:                 order.GiftMoney,
			GiftPoint:                  order.GiftPoint,
			MemberUuid:                 uint64(order.UserID),
			StaffUuid:                  statffUuid,
			PaymentTime:                int64(order.PayTime),
			PaymentOrders:              paymentOrders,
			RechargeOrderOperationLogs: operationLogs,
			ReturnOrders:               returnOrders,
		}
		err = repository.NewMemberRechargeOrderRepo(s.targetDB).CreateOldRecord(userRechargeOrder)
		if err != nil {
			return err
		}
	}
	return nil
}
