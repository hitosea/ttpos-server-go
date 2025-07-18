package model

import (
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// MemberSaleOrder 会员端销售订单表 `ttpos_member_sale_order`
type MemberSaleOrder struct {
	BaseModel
	MemberUuid        uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员UUID'"`
	Status            uint    `gorm:"column:status;type:int(10);not null;default:0;comment:'订单状态 0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消'"`
	SerialNumber      string  `gorm:"column:serial_number;type:varchar(255);not null;default:'';comment:'订单流水号'"`
	OrderNo           string  `gorm:"column:order_no;type:varchar(255);not null;default:'';comment:'订单号'"`
	CancelScene       string  `gorm:"column:cancel_scene;type:varchar(50);not null;default:'';comment:'取消场景：merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单'"`
	IsAutoAccept      uint    `gorm:"column:is_auto_accept;type:int(10);not null;default:0;comment:'是否自动接单：0-否；1-是'"`
	DeliveryDistance  float64 `gorm:"column:delivery_distance;type:decimal(12,6);not null;default:0;comment:'配送距离，单位km'"`
	Remark            string  `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'订单备注'"`
	CancelReason      string  `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:'取消原因'"`
	IsVerifiedPhone   uint    `gorm:"column:is_verified_phone;type:int(10);not null;default:0;comment:'订单是否已经验证手机号,0-未验证 1-已验证,不再弹出验证手机号'"`
	PaymentMethodUuid uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;not null;default:0;comment:'支付方式UUID,订单已选择的支付方式'"`
	// 确认订单之后才有值的字段
	ProductNum        float64 `gorm:"column:product_num;type:decimal(12,2);not null;default:0;comment:'商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3'"`
	ProductAmount     float64 `gorm:"column:product_amount;type:decimal(12,2);not null;default:0;comment:'商品金额,折前价，已含税'"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);not null;default:0;comment:'会员折扣'"`
	Amount            float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:'订单总金额.商品金额-会员折扣+配送费'"`
	RefundAmount      float64 `gorm:"column:refund_amount;type:decimal(12,2);not null;default:0;comment:'退款金额'"`
	// 配送费参数
	DeliveryFeeAmount  float64 `gorm:"column:delivery_fee_amount;type:decimal(12,6);not null;default:0;comment:'配送费'"`
	DeliveryFeeMinFee  float64 `gorm:"column:delivery_fee_min_fee;type:decimal(12,6);not null;default:0;comment:'起步配送费'"`
	DeliveryFeeBaseFee float64 `gorm:"column:delivery_fee_base_fee;type:decimal(12,6);not null;default:0;comment:'基础配送费'"`
	DeliveryFeePerKm   float64 `gorm:"column:delivery_fee_per_km;type:decimal(12,6);not null;default:0;comment:'每公里配送费'"`
	// 第三方订单信息
	RelatedOrderNo   string `gorm:"column:related_order_no;type:varchar(255);not null;default:'';comment:'关联订单号,skootar、grab等第三方平台上的订单号'"`
	RelatedOrderType string `gorm:"column:related_order_type;type:varchar(255);not null;default:'';comment:'关联订单类型,skootar、grab'"`
	// 骑手信息
	RiderName         string  `gorm:"column:rider_name;type:varchar(255);not null;default:'';comment:'骑手名称'"`
	RiderPhone        string  `gorm:"column:rider_phone;type:varchar(255);not null;default:'';comment:'骑手电话'"`
	RiderLatitude     float64 `gorm:"column:rider_latitude;type:decimal(12,6);not null;default:0;comment:'骑手纬度'"`
	RiderLongitude    float64 `gorm:"column:rider_longitude;type:decimal(12,6);not null;default:0;comment:'骑手经度'"`
	RemainingDistance float64 `gorm:"column:remaining_distance;type:decimal(12,6);not null;default:0;comment:'剩余距离'"`
	// 时间相关
	PayTime            int64 `gorm:"column:pay_time;type:int(10) unsigned;not null;default:0;comment:'支付完成时间（时间戳）'"`
	AcceptTime         int64 `gorm:"column:accept_time;type:int(10) unsigned;not null;default:0;comment:'商家接单时间（时间戳）'"`
	CookTime           int64 `gorm:"column:cook_time;type:int(10) unsigned;not null;default:0;comment:'商家备餐完成时间（时间戳）'"`
	RiderAcceptTime    int64 `gorm:"column:rider_accept_time;type:int(10) unsigned;not null;default:0;comment:'骑手接单时间（时间戳）'"`
	RiderStartTime     int64 `gorm:"column:rider_start_time;type:int(10) unsigned;not null;default:0;comment:'骑手开始配送时间（时间戳）'"`
	FinishTime         int64 `gorm:"column:finish_time;type:int(10) unsigned;not null;default:0;comment:'骑手送达时间（时间戳）'"`
	ExpectedFinishTime int64 `gorm:"column:expected_finish_time;type:int(10) unsigned;not null;default:0;comment:'预计送达时间（时间戳）'"`
	CancelTime         int64 `gorm:"column:cancel_time;type:int(10) unsigned;not null;default:0;comment:'取消时间（时间戳）'"`

	SaleBill      *SaleBill               `gorm:"foreignKey:MemberSaleOrderUuid;references:Uuid"`
	Address       *MemberSaleOrderAddress `gorm:"foreignKey:MemberSaleOrderUuid;references:Uuid"`
	PaymentMethod *PaymentMethod          `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
}

func (model *MemberSaleOrder) OriginAmountValue() float64 {
	return decimal.NewFromFloat(model.ProductAmount).Add(decimal.NewFromFloat(model.DeliveryFeeAmount)).Round(2).InexactFloat64()
}

// 订单是否可以支付 0-未支付 1-已支付 2-已取消
func (model *MemberSaleOrder) GetPayStatus() uint {
	if model.Status == constant.MemberSaleOrderStatusCancelled {
		return 2
	}
	if model.Status <= constant.MemberSaleOrderStatusPendingPayment {
		return 0
	}
	return 1
}

func (model *MemberSaleOrder) SetNil() {
	model.SaleBill = nil
	model.Address = nil
	model.PaymentMethod = nil
}

// 设置订单为“待支付”状态
func (model *MemberSaleOrder) SetPendingPayment(paymentMethodUuid uint64) {
	model.Status = constant.MemberSaleOrderStatusPendingPayment
	model.PaymentMethodUuid = paymentMethodUuid
}

// 订单是否已经取消
func (model *MemberSaleOrder) IsCancel() bool {
	return model.Status == constant.MemberSaleOrderStatusCancelled
}

// 订单是否可以支付
func (model *MemberSaleOrder) IsCanPaid() bool {
	return model.Status == constant.MemberSaleOrderStatusPendingPayment
}

// 订单是否已经验证手机号
func (model *MemberSaleOrder) IsVerifiedPhoneBool() bool {
	// 订单未配置收货地址前，默认都是未验证
	if model.Address == nil {
		return false
	}
	return model.Address.MemberAddress.IsAuthPhone()
}

// 计算配送费
// 配送费计算规则：
// 1. 配送费=基础配送费 + 配送距离*每公里配送费
// 2. 如果配送费<起步配送费，则配送费=起步配送费
func (model *MemberSaleOrder) CalculateDeliveryFee() float64 {
	var deliveryFeeAmount float64
	distanceFee := decimal.NewFromFloat(model.DeliveryDistance).Mul(decimal.NewFromFloat(model.DeliveryFeePerKm)).Round(2) // 配送距离*每公里配送费
	deliveryFee := decimal.NewFromFloat(model.DeliveryFeeBaseFee).Add(distanceFee).Round(2)                                // 基础配送费 + 配送距离*每公里配送费
	if deliveryFee.LessThan(decimal.NewFromFloat(model.DeliveryFeeMinFee)) {                                               // 如果配送费<起步配送费
		deliveryFeeAmount = model.DeliveryFeeMinFee // 配送费=起步配送费
	} else {
		deliveryFeeAmount = deliveryFee.InexactFloat64() // 配送费=基础配送费 + 配送距离*每公里配送费
	}

	return deliveryFeeAmount
}

// 计算订单总金额. 订单总金额=商品金额+配送费-会员折扣
func (model *MemberSaleOrder) CalculateAmount() float64 {
	return decimal.NewFromFloat(model.ProductAmount).Add(decimal.NewFromFloat(model.CalculateDeliveryFee())).Sub(decimal.NewFromFloat(model.MemberDiscountFee)).Round(2).InexactFloat64()
}

// 接单
func (model *MemberSaleOrder) Accept() {
	model.Status = constant.MemberSaleOrderStatusCooking // 商家备餐中
	model.AcceptTime = time.Now().Unix()
}

// 拒单
func (model *MemberSaleOrder) Reject() {
	model.Status = constant.MemberSaleOrderStatusCancelled          // 已取消
	model.CancelScene = constant.MemberSaleOrderSceneMerchantReject // 商家拒单
	model.CancelTime = time.Now().Unix()
	model.CancelReason = "商家拒单"
}

type CreateMemberSaleOrderParams struct {
	DeliveryConfig DeliveryConfigResponse
	SerialNo       string // 订单流水号
	MemberUuid     uint64 // 会员UUID
}

func NewMemberSaleOrder(params CreateMemberSaleOrderParams) *MemberSaleOrder {
	uuid, _ := utils.GetID()
	saleOrder := &MemberSaleOrder{
		BaseModel:          BaseModel{Uuid: uuid},
		Status:             constant.MemberSaleOrderStatusSelecting,
		DeliveryFeeMinFee:  params.DeliveryConfig.BaseDeliveryFee,
		DeliveryFeeBaseFee: params.DeliveryConfig.BasicFee,
		DeliveryFeePerKm:   params.DeliveryConfig.PricePerKm,
		SerialNumber:       params.SerialNo,
		MemberUuid:         params.MemberUuid,
	}
	return saleOrder
}

// 会员销售订单地址 `ttpos_member_sale_order_address`
type MemberSaleOrderAddress struct {
	BaseModel
	MemberUuid          uint64  `gorm:"column:member_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员UUID'"`
	MemberAddressUuid   uint64  `gorm:"column:member_address_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员收货地址UUID'"`
	Longitude           float64 `gorm:"column:longitude;type:decimal(12,6);not null;default:0;comment:'经度'"`
	Latitude            float64 `gorm:"column:latitude;type:decimal(12,6);not null;default:0;comment:'纬度'"`
	Location            string  `gorm:"column:location;type:varchar(100);not null;comment:位置坐标" json:"location"` // "纬度,经度"
	Address             string  `gorm:"column:address;type:varchar(255);not null;default:'';comment:'地址'"`
	DetailAddress       string  `gorm:"column:detail_address;type:varchar(255);not null;default:'';comment:'详细地址'"`
	ContactName         string  `gorm:"column:contact_name;type:varchar(255);not null;default:'';comment:'联系人'"`
	ContactPhone        string  `gorm:"column:contact_phone;type:varchar(255);not null;default:'';comment:'联系电话'"`
	PhonePrefix         string  `gorm:"column:phone_prefix;type:varchar(255);not null;default:'';comment:'联系电话前缀'"`
	ContactGender       int     `gorm:"column:contact_gender;type:int(10);not null;default:0;comment:'联系人性别, 0-女士 1-先生'"`
	MemberSaleOrderUuid uint64  `gorm:"column:member_sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员销售订单UUID'"`

	Member        *Member        `gorm:"foreignKey:MemberUuid;references:Uuid"`
	MemberAddress *MemberAddress `gorm:"foreignKey:MemberAddressUuid;references:Uuid"`
}

func (model *MemberSaleOrderAddress) SetNil() {
}

// 获取位置坐标. 返回纬度,经度
func (model *MemberSaleOrderAddress) GetLocation() (string, string, error) {
	if model.Location == "" {
		return "", "", errors.New("位置坐标为空")
	}
	location := strings.Split(model.Location, ",")
	if len(location) != 2 {
		return "", "", errors.New("位置坐标格式错误")
	}
	return location[0], location[1], nil
}
