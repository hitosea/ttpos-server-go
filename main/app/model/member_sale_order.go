package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// MemberSaleOrder 会员端销售订单表 `ttpos_member_sale_order`
type MemberSaleOrder struct {
	BaseModel
	MemberUuid        uint64 `gorm:"column:member_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员UUID'"`
	SaleBillUuid      uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:'销售账单UUID'"`
	SaleOrderUuid     uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:'销售订单UUID'"`
	Status            uint   `gorm:"column:status;type:int(10);not null;default:0;comment:'订单状态 0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消'"`
	SerialNumber      string `gorm:"column:serial_number;type:varchar(255);not null;default:'';comment:'订单流水号'"`
	OrderNo           string `gorm:"column:order_no;type:varchar(255);not null;default:'';comment:'订单号'"`
	CancelScene       string `gorm:"column:cancel_scene;type:varchar(50);not null;default:'';comment:'取消场景：merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单'"`
	IsAutoAccept      uint   `gorm:"column:is_auto_accept;type:int(10);not null;default:0;comment:'是否自动接单：0-否；1-是'"`
	Remark            string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'订单备注'"`
	CancelReason      string `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:'取消原因'"`
	IsVerifiedPhone   uint   `gorm:"column:is_verified_phone;type:int(10);not null;default:0;comment:'订单是否已经验证手机号,0-未验证 1-已验证,不再弹出验证手机号'"`
	PaymentMethodUuid uint64 `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;not null;default:0;comment:'支付方式UUID,订单已选择的支付方式'"`
	// 确认订单之后才有值的字段
	ProductNum          float64 `gorm:"column:product_num;type:decimal(12,2);not null;default:0;comment:'商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3'"`
	ProductAmount       float64 `gorm:"column:product_amount;type:decimal(12,2);not null;default:0;comment:'商品金额,折前价，已含税'"`
	OriginProductAmount float64 `gorm:"column:origin_product_amount;type:decimal(12,2);not null;default:0;comment:'商品原价,折前价，已含税'"`
	MemberDiscountFee   float64 `gorm:"column:member_discount_fee;type:decimal(12,2);not null;default:0;comment:'会员折扣'"`
	Amount              float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:'订单总金额.商品金额-会员折扣+配送费'"`
	RefundAmount        float64 `gorm:"column:refund_amount;type:decimal(12,2);not null;default:0;comment:'退款金额'"`
	// 配送费参数
	IsDistanceCalculated int     `gorm:"column:is_distance_calculated;type:int(10);not null;default:-1;comment:'是否已计算距离费，-1-未计算，1-已计算'"`
	DeliveryDistance     float64 `gorm:"column:delivery_distance;type:decimal(12,6);not null;default:0;comment:'配送距离，单位km'"`
	DeliveryFeeAmount    float64 `gorm:"column:delivery_fee_amount;type:decimal(12,6);not null;default:0;comment:'配送费'"`
	DeliveryFeeDistance  float64 `gorm:"column:delivery_fee_distance;type:decimal(12,6);not null;default:0;comment:'距离费送费'"`
	DeliveryFeeMinFee    float64 `gorm:"column:delivery_fee_min_fee;type:decimal(12,6);not null;default:0;comment:'起步配送费'"`
	DeliveryFeeBaseFee   float64 `gorm:"column:delivery_fee_base_fee;type:decimal(12,6);not null;default:0;comment:'基础服务费'"`
	DeliveryFeePerKm     float64 `gorm:"column:delivery_fee_per_km;type:decimal(12,6);not null;default:0;comment:'每公里配送费'"`
	RiderAcceptTimeout   int     `gorm:"column:rider_accept_timeout;type:int(10);not null;default:0;comment:'骑手接单超时时间,单位分钟'"`
	// 第三方订单信息
	RelatedOrderNo   string `gorm:"column:related_order_no;type:varchar(255);not null;default:'';comment:'关联订单号,skootar、grab等第三方平台上的订单号'"`
	RelatedOrderType string `gorm:"column:related_order_type;type:varchar(255);not null;default:'';comment:'关联订单类型,skootar、grab'"`
	// 骑手信息
	RiderName         string  `gorm:"column:rider_name;type:varchar(255);not null;default:'';comment:'骑手名称'"`
	RiderPhone        string  `gorm:"column:rider_phone;type:varchar(255);not null;default:'';comment:'骑手电话'"`
	Location          string  `gorm:"column:location;type:varchar(255);not null;default:'';comment:'骑手位置,格式:纬度,经度'"`
	RiderAvatar       string  `gorm:"column:rider_avatar;type:varchar(255);not null;default:'';comment:'骑手头像'"`
	RiderRating       float64 `gorm:"column:rider_rating;type:decimal(12,2);not null;default:0.00;comment:'骑手评分'"`
	RemainingDistance float64 `gorm:"column:remaining_distance;type:decimal(12,6);not null;default:0;comment:'剩余距离'"`
	// 收货人信息
	MemberAddressUuid    uint64 `gorm:"column:member_address_uuid;type:bigint(20) unsigned;not null;default:0;comment:'会员收货地址UUID'"`
	ContactLocation      string `gorm:"column:contact_location;type:varchar(100);not null;comment:位置坐标" json:"location"` // "纬度,经度"
	ContactAddress       string `gorm:"column:contact_address;type:varchar(255);not null;default:'';comment:'详细地址'"`
	ContactAddressDetail string `gorm:"column:contact_address_detail;type:varchar(255);not null;default:'';comment:'详细地址'"`
	ContactName          string `gorm:"column:contact_name;type:varchar(255);not null;default:'';comment:'联系人'"`
	ContactPhone         string `gorm:"column:contact_phone;type:varchar(255);not null;default:'';comment:'联系电话'"`
	ContactPhonePrefix   string `gorm:"column:contact_phone_prefix;type:varchar(255);not null;default:'';comment:'联系电话前缀'"`
	ContactGender        int    `gorm:"column:contact_gender;type:int(10);not null;default:0;comment:'联系人性别, 0-女士 1-先生'"`
	// 排序相关
	Sort int `gorm:"column:sort;type:int(10);not null;default:0;comment:'排序, 0-其他状态，1-骑手正在赶往商家，2-骑手配送中，降序排序'"`
	// 时间相关
	PayTime            int64  `gorm:"column:pay_time;type:int(10) unsigned;not null;default:0;comment:'支付完成时间（时间戳）'"`
	SubmitPayTime      int64  `gorm:"column:submit_pay_time;type:int(10) unsigned;not null;default:0;comment:'提交支付时间戳'"`
	AcceptTime         int64  `gorm:"column:accept_time;type:int(10) unsigned;not null;default:0;comment:'商家接单时间（时间戳）'"`
	CookTime           int64  `gorm:"column:cook_time;type:int(10) unsigned;not null;default:0;comment:'商家备餐完成时间（时间戳）'"`
	RiderAcceptTime    int64  `gorm:"column:rider_accept_time;type:int(10) unsigned;not null;default:0;comment:'骑手接单时间（时间戳）'"`
	RiderStartTime     int64  `gorm:"column:rider_start_time;type:int(10) unsigned;not null;default:0;comment:'骑手开始配送时间（时间戳）'"`
	FinishTime         int64  `gorm:"column:finish_time;type:int(10) unsigned;not null;default:0;comment:'骑手送达时间（时间戳）'"`
	ExpectedFinishTime string `gorm:"column:expected_finish_time;type:varchar(255);not null;default:'';comment:'预计送达时间'"`
	CancelTime         int64  `gorm:"column:cancel_time;type:int(10) unsigned;not null;default:0;comment:'取消时间（时间戳）'"`

	SaleBill      *SaleBill      `gorm:"foreignKey:MemberSaleOrderUuid;references:Uuid"`
	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"`
	Address       *MemberAddress `gorm:"foreignKey:MemberAddressUuid;references:Uuid"`
	Member        *Member        `gorm:"foreignKey:MemberUuid;references:Uuid"`
}

// 获取订单实际消费金额（订单金额-配送费）
func (model *MemberSaleOrder) GetActualConsumptionAmount() float64 {
	return decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(model.DeliveryFeeAmount)).Round(2).InexactFloat64()
}

// 获取脱敏手机号（只显示后四位）
func (model *MemberSaleOrder) GetContactPhoneMask() string {
	if model.ContactPhone == "" {
		return ""
	}

	phoneLen := len(model.ContactPhone)
	if phoneLen <= 4 {
		// 如果手机号长度小于等于4位，全部用*替换
		return strings.Repeat("*", phoneLen)
	}

	// 只显示后4位，前面的都用*替换
	maskCount := phoneLen - 4
	return strings.Repeat("*", maskCount) + model.ContactPhone[phoneLen-4:]
}

// 是否是自主取消
func (model *MemberSaleOrder) IsSelfCancel() bool {
	return model.CancelScene == constant.MemberSaleOrderSceneMemberCancel ||
		model.CancelScene == constant.MemberSaleOrderSceneMemberCancelUnpaid
}

// 是否是商家取消
func (model *MemberSaleOrder) IsMerchantCancel() bool {
	return model.CancelScene == constant.MemberSaleOrderSceneMerchantCancel
}

// 是否还可退款
func (model *MemberSaleOrder) IsCanRefund() bool {
	canRefundAmount := decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(model.DeliveryFeeAmount)).Round(2).InexactFloat64()
	return model.Status == constant.MemberSaleOrderStatusCompleted && model.RefundAmount < canRefundAmount
}

// 计算外送单会员折扣。外送单的会员折扣=Sum(商品原价(含税费)-商品折后价（含税费）)
func (model *MemberSaleOrder) CalculateMemberDiscount() float64 {
	saleOrder := model.SaleBill.SaleOrders[0]
	memberDiscountFee := decimal.NewFromFloat(0)
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		discountFee := decimal.NewFromFloat(saleOrderProduct.OriginTotalPrice).Sub(decimal.NewFromFloat(saleOrderProduct.TotalPrice)).Round(2)
		discount := discountFee.Mul(decimal.NewFromFloat(saleOrderProduct.Num)).Round(2)
		if discount.GreaterThan(decimal.NewFromFloat(0)) {
			memberDiscountFee = memberDiscountFee.Add(discount)
		}
	}
	return memberDiscountFee.Round(2).InexactFloat64()
}

// 更新配送费配置
func (model *MemberSaleOrder) UpdateDeliveryConfig(deliveryConfig DeliveryConfigResponse) {
	model.DeliveryFeeMinFee = deliveryConfig.BaseDeliveryFee
	model.DeliveryFeeBaseFee = deliveryConfig.BasicFee
	model.DeliveryFeePerKm = deliveryConfig.PricePerKm
	model.RiderAcceptTimeout = deliveryConfig.RiderAcceptanceTimeout
	// 重新计算配送费
	model.RecalculateDeliveryFee()
}

// 是否已计算距离费
func (model *MemberSaleOrder) GetIsDistanceCalculated() bool {
	return model.IsDistanceCalculated == constant.DistanceCalculated
}

// 设置配送距离
func (model *MemberSaleOrder) SetDeliveryDistance(distance float64) {
	model.DeliveryDistance = distance
	model.IsDistanceCalculated = constant.DistanceCalculated
	// 重新计算配送费
	model.RecalculateDeliveryFee()
}

// 取消已设置的配送距离
func (model *MemberSaleOrder) CancelDeliveryDistance() {
	model.DeliveryDistance = 0
	model.IsDistanceCalculated = constant.DistanceNotCalculated
	// 重新计算配送费
	model.RecalculateDeliveryFee()
}

// 重新计算配送费
func (model *MemberSaleOrder) RecalculateDeliveryFee() {
	model.DeliveryFeeDistance = model.CalculateDeliveryFeeDistance()
	model.DeliveryFeeAmount = model.CalculateDeliveryFee()
}

func (model *MemberSaleOrder) OriginAmountValue() float64 {
	return decimal.NewFromFloat(model.OriginProductAmount).Add(decimal.NewFromFloat(model.DeliveryFeeAmount)).Round(2).InexactFloat64()
}

// SetMemberAddress 设置会员地址
func (model *MemberSaleOrder) SetMemberAddress(address MemberAddress) {
	model.MemberAddressUuid = address.Uuid
	model.ContactLocation = address.Location
	model.ContactAddress = address.Address
	model.ContactAddressDetail = address.Street
	model.ContactName = address.Name
	model.ContactPhone = address.Phone
	model.ContactPhonePrefix = address.PhonePrefix
	model.Address = &address
	model.CancelDeliveryDistance()
}

// 获取剩余支付时间(单位秒)
func (model *MemberSaleOrder) GetRemainingPaymentTime() int64 {
	// 支付超时时间为24小时（86400秒）
	paymentTimeoutSeconds := config.Server.PaymentTimeout // 24小时

	// 如果订单已经支付或取消，返回0
	if model.Status != constant.MemberSaleOrderStatusPendingPayment {
		return 0
	}

	// 如果没有提交支付时间，从创建时间开始计算
	baseTime := model.SubmitPayTime
	if baseTime == 0 {
		baseTime = model.CreateTime
	}

	// 计算过期时间
	expireTime := baseTime + paymentTimeoutSeconds
	currentTime := time.Now().Unix()

	// 如果已经过期，返回0
	if currentTime >= expireTime {
		return 0
	}

	// 返回剩余时间（秒）
	return expireTime - currentTime
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
	model.PaymentMethod = nil
}

// 获取骑手经纬度信息。第一个是纬度，第二个是经度
func (model *MemberSaleOrder) GetLocation() (string, string) {
	if model.Location == "" {
		return "", ""
	}
	location := strings.Split(model.Location, ",")
	if len(location) != 2 {
		return "", ""
	}
	// 去掉前后空格并保留6位小数
	location[0] = strings.TrimSpace(location[0])
	location[1] = strings.TrimSpace(location[1])
	// 转换为float64并保留6位小数
	lat, err := strconv.ParseFloat(location[0], 64)
	if err == nil {
		location[0] = fmt.Sprintf("%.6f", lat)
	}
	lng, err := strconv.ParseFloat(location[1], 64)
	if err == nil {
		location[1] = fmt.Sprintf("%.6f", lng)
	}
	return location[0], location[1]
}

// 获取买家经纬度信息。第一个是纬度，第二个是经度
func (model *MemberSaleOrder) GetCustomerLocation() (string, string) {
	if model.ContactLocation == "" {
		return "", ""
	}
	location := strings.Split(model.ContactLocation, ",")
	if len(location) != 2 {
		return "", ""
	}
	// 去掉前后空格并保留6位小数
	location[0] = strings.TrimSpace(location[0])
	location[1] = strings.TrimSpace(location[1])
	// 转换为float64并保留6位小数
	lat, err := strconv.ParseFloat(location[0], 64)
	if err == nil {
		location[0] = fmt.Sprintf("%.6f", lat)
	}
	lng, err := strconv.ParseFloat(location[1], 64)
	if err == nil {
		location[1] = fmt.Sprintf("%.6f", lng)
	}
	return location[0], location[1]
}

// 设置订单为“已取消”状态
func (model *MemberSaleOrder) SetCancel(cancelReason string) {
	if model.Status < constant.MemberSaleOrderStatusPendingMerchantAccept {
		model.CancelScene = constant.MemberSaleOrderSceneMemberCancelUnpaid
	} else {
		model.CancelScene = constant.MemberSaleOrderSceneMemberCancel
	}
	model.Status = constant.MemberSaleOrderStatusCancelled
	model.CancelReason = cancelReason
	model.CancelTime = time.Now().Unix()
}

// 设置订单为“已取消”状态
func (model *MemberSaleOrder) SetCancelInCashier(cancelReason string) {
	model.Status = constant.MemberSaleOrderStatusCancelled
	model.CancelReason = cancelReason
	model.CancelTime = time.Now().Unix()
	model.CancelScene = constant.MemberSaleOrderSceneMerchantCancel
}

// 设置订单为"待支付"状态
func (model *MemberSaleOrder) SetPendingPayment(paymentMethodUuid uint64) error {
	statusMap := map[uint]bool{
		constant.MemberSaleOrderStatusSelecting:      true, // 选购中
		constant.MemberSaleOrderStatusPendingPayment: true, // 待支付
	}
	if !statusMap[model.Status] {
		return errors.New("订单状态不支持支付")
	}
	model.Status = constant.MemberSaleOrderStatusPendingPayment
	model.PaymentMethodUuid = paymentMethodUuid
	return nil
}

// 订单是否已经取消
func (model *MemberSaleOrder) IsCancel() bool {
	return model.Status == constant.MemberSaleOrderStatusCancelled
}

// 订单是否可以支付
func (model *MemberSaleOrder) IsCanPaid() bool {
	return model.Status == constant.MemberSaleOrderStatusPendingPayment
}

// 订单是否可取消。会员可以在商家没接单之前取消订单。商家接单之后，会员不能取消订单
func (model *MemberSaleOrder) IsCanCancel() bool {
	// 商家没接单之前可以取消
	return model.Status <= constant.MemberSaleOrderStatusPendingMerchantAccept
}

// 订单是否可取消。商家只有在备餐中可以取消
func (model *MemberSaleOrder) IsCanCancelInCashier() bool {
	// 商家在备餐完成之前可以取消
	return model.Status == constant.MemberSaleOrderStatusCooking
}

// IsRiderPickup 订单是否骑手接单
func (model *MemberSaleOrder) IsRiderPickup() bool {
	return model.Status > constant.MemberSaleOrderStatusPendingRiderPickup
}

// 订单是否已经验证手机号
func (model *MemberSaleOrder) IsVerifiedPhoneBool() bool {
	// 订单未配置收货地址前，默认都是未验证
	if model.Address == nil {
		return false
	}
	return model.Address.IsAuthPhone()
}

// 计算配送费
// 配送费计算规则：
// 1. 配送费=基础配送费 + 配送距离*每公里配送费
// 2. 如果配送费<起步配送费，则配送费=起步配送费
func (model *MemberSaleOrder) CalculateDeliveryFee() float64 {
	var deliveryFeeAmount float64
	distanceFee := decimal.NewFromFloat(model.CalculateDeliveryFeeDistance()).Round(2)      // 配送距离*每公里配送费
	deliveryFee := decimal.NewFromFloat(model.DeliveryFeeBaseFee).Add(distanceFee).Round(2) // 基础配送费 + 配送距离*每公里配送费
	if deliveryFee.LessThan(decimal.NewFromFloat(model.DeliveryFeeMinFee)) {                // 如果配送费<起步配送费
		deliveryFeeAmount = model.DeliveryFeeMinFee // 配送费=起步配送费
	} else {
		deliveryFeeAmount = deliveryFee.InexactFloat64() // 配送费=基础配送费 + 配送距离*每公里配送费
	}

	return deliveryFeeAmount
}

// 计算距离费
func (model *MemberSaleOrder) CalculateDeliveryFeeDistance() float64 {
	return decimal.NewFromFloat(model.DeliveryDistance).Mul(decimal.NewFromFloat(model.DeliveryFeePerKm)).Round(2).InexactFloat64() // 配送距离*每公里配送费
}

// 计算订单总金额. 订单总金额=商品金额+配送费-会员折扣
func (model *MemberSaleOrder) CalculateAmount() float64 {
	return decimal.NewFromFloat(model.OriginProductAmount).Add(decimal.NewFromFloat(model.CalculateDeliveryFee())).Sub(decimal.NewFromFloat(model.MemberDiscountFee)).Round(2).InexactFloat64()
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
	model.CancelReason = constant.MemberSaleOrderSceneReason[constant.MemberSaleOrderSceneMerchantReject]
}

// 备餐完成
func (model *MemberSaleOrder) CookFinish() {
	model.Status = constant.MemberSaleOrderStatusPendingRiderPickup // 商家备餐完成
	model.CookTime = time.Now().Unix()
}

// 骑手正在赶往商家
func (model *MemberSaleOrder) RiderAccept(riderName string, riderPhone string, location string, riderAvatar string, riderRating float64) {
	model.Status = constant.MemberSaleOrderStatusPendingRiderDelivery // 骑手正在赶往商家
	model.RiderAcceptTime = time.Now().Unix()
	model.RiderName = riderName
	model.RiderPhone = riderPhone
	model.Location = location
	model.RiderAvatar = riderAvatar
	model.RiderRating = riderRating
}

// 骑手配送中
func (model *MemberSaleOrder) RiderDelivery(riderName string, riderPhone string, location string) {
	model.Status = constant.MemberSaleOrderStatusDelivering // 骑手配送中
	model.RiderStartTime = time.Now().Unix()
	model.RiderName = riderName
	model.RiderPhone = riderPhone
	model.Location = location
}

// 骑手配送完成
func (model *MemberSaleOrder) RiderCompleted(riderName string, riderPhone string, location string) {
	model.Status = constant.MemberSaleOrderStatusCompleted // 骑手配送完成
	model.FinishTime = time.Now().Unix()
	model.RiderName = riderName
	model.RiderPhone = riderPhone
	model.Location = location
}

type CreateMemberSaleOrderParams struct {
	Uuid           uint64
	DeliveryConfig DeliveryConfigResponse
	SerialNo       string // 订单流水号
	OrderNo        string // 订单号
	SaleBillUuid   uint64 // 销售账单UUID
	SaleOrderUuid  uint64 // 销售订单UUID
	MemberUuid     uint64 // 会员UUID
}

func NewMemberSaleOrder(params CreateMemberSaleOrderParams) *MemberSaleOrder {
	var uuid uint64
	uuid = params.Uuid
	// 如果uuid没有值的话，重新生成uuid
	if uuid == 0 {
		uuid, _ = utils.GetID()
	}
	saleOrder := &MemberSaleOrder{
		BaseModel:          BaseModel{Uuid: uuid},
		Status:             constant.MemberSaleOrderStatusSelecting,
		OrderNo:            params.OrderNo,
		DeliveryFeeMinFee:  params.DeliveryConfig.BaseDeliveryFee,
		DeliveryFeeBaseFee: params.DeliveryConfig.BasicFee,
		DeliveryFeePerKm:   params.DeliveryConfig.PricePerKm,
		SerialNumber:       params.SerialNo,
		MemberUuid:         params.MemberUuid,
		SaleBillUuid:       params.SaleBillUuid,
		SaleOrderUuid:      params.SaleOrderUuid,
	}
	return saleOrder
}
