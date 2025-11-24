package model

import (
	"encoding/json"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
)

// SaleBill 销售账单 `ttpos_sale_bill`
type SaleBill struct {
	BaseModel
	// 主键和标识字段
	OrderNo  string `gorm:"column:order_no;type:varchar(255);default:'';comment:销售账单编号" json:"order_no"`
	DutyNo   string `gorm:"column:duty_no;type:varchar(255);default:'';comment:当班编号,用于标记该账单属于哪个当班" json:"duty_no"` // 创建账单时，收银员通过duty_no来标记该账单属于哪个当班；结账完成时，通过duty_no来标记该账单属于哪个当班
	SerialNo string `gorm:"column:serial_no;type:varchar(255);default:'';comment:桌位编号 (点餐流水号)" json:"serial_no"`

	// 状态相关字段
	Status       uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-待付款、1-已完成、2-已取消" json:"status"`
	IsLock       uint `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`
	IsSplitOrder uint `gorm:"column:is_split_order;type:tinyint(1);default:0;comment:是否拆单, 0-否 1-是" json:"is_split_order"`

	// 订单类型字段
	BillType        uint   `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-桌台订单、1-点餐订单、2-会员端订单" json:"bill_type"`
	DiningMethod    uint   `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	OrderSourceUuid uint64 `gorm:"column:order_source_uuid;type:bigint(20);default:0;comment:订单来源UUID（0=店内，>0=外卖）" json:"order_source_uuid"`
	NationalityUuid uint64 `gorm:"column:nationality_uuid;type:bigint(20);default:0;comment:国籍UUID（0=未记录）" json:"nationality_uuid"`
	IsBuffet        uint   `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	BuffetDuration  uint   `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒），0为不限时. 原始值为自助餐的时长，加钟时会累加" json:"buffet_duration"`
	BuffetStartTime int64  `gorm:"column:buffet_start_time;type:int(10);default:0;comment:自助餐开始时间（秒）" json:"buffet_start_time"`
	DelayDuration   uint   `gorm:"column:delay_duration;type:int(10);default:0;comment:总延迟时长（秒）" json:"delay_duration"`
	DelayStartTime  int64  `gorm:"column:delay_start_time;type:int(10);default:0;comment:总延迟时长开始时间（秒）" json:"delay_start_time"`

	NonOrderingTime   uint `gorm:"default:0;comment:不可下单时间（分钟）"`
	ReminderOrderTime uint `gorm:"default:0;column:reminder_order_time;comment:提醒下单时间（分钟）"`

	// 订单基本信息
	MealNum     uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Remark      string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	OrderRemark string `gorm:"column:order_remark;type:text;default:'';comment:整单备注JSON" json:"order_remark"`
	Reason      string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 金额字段 - 主要金额
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
	OriginAmount          float64 `gorm:"column:origin_amount;type:decimal(12,2);default:0;comment:订单金额(折前价)。商品未含税时，订单金额(折前价)=商品金额+服务费+税费。商品已含税时，订单金额(折前价)=商品金额（含商品消费税）+服务费+税费（只有服务费税）" json:"origin_amount"`
	ProductAmount         float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额,关联销售订单的商品金额之和" json:"product_amount"`
	ProductOriginalAmount float64 `gorm:"column:product_original_amount;type:decimal(12,2);default:0;comment:原始商品金额。 商品原始金额=(订单.原始商品金额)之和。" json:"product_original_amount"`

	// 金额字段 - 支付相关
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,多次支付的支付手续费之和" json:"payment_commission_fee"`

	// 金额字段 - 费用相关
	ServiceFee float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费,关联销售订单的服务费之和" json:"service_fee"`
	TaxFee     float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费,关联销售订单的税费之和" json:"tax_fee"`

	// 金额字段 - 优惠相关
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);default:0;comment:折扣费用,关联销售订单的折扣费用之和" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用,关联销售订单的会员折扣费用之和" json:"member_discount_fee"`
	GiftAmount        float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,关联销售订单的赠菜金额之和" json:"gift_amount"`
	ActivityAmount    float64 `gorm:"column:activity_amount;type:decimal(20,8);default:0;comment:满减活动抵扣金额（所有sale_order的满减扣减金额总和）" json:"activity_amount"`
	FreeAmount        float64 `gorm:"column:free_amount;type:decimal(12,2);default:0;comment:免单金额,关联销售订单的免单金额之和" json:"free_amount"`

	// 时间相关字段
	LockTime       int64 `gorm:"column:lock_time;type:int(10);default:0;comment:锁单时间" json:"lock_time"`
	FinishTime     int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	HideBillTime   int64 `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`
	ProductionTime int64 `gorm:"column:production_time;type:int(10);default:0;comment:首次送厨时间（时间戳）" json:"production_time"`

	// 收银员名称
	CashierName string `gorm:"column:cashier_name;type:varchar(255);default:'';comment:收银员名称" json:"cashier_name"`

	// 关联ID字段
	ConsumerUuid        uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid         uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	DeskUuid            uint64 `gorm:"column:desk_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"desk_uuid"`
	BuffetPackage1Uuid  uint64 `gorm:"column:buffet_package1_uuid;type:bigint(20);default:0;comment:自助餐套餐1ID" json:"buffet_package1_uuid"`
	BuffetPackage2Uuid  uint64 `gorm:"column:buffet_package2_uuid;type:bigint(20);default:0;comment:自助餐套餐2ID" json:"buffet_package2_uuid"`
	DeviceUuid          uint64 `gorm:"column:device_uuid;type:bigint(20);default:0;comment:设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询" json:"device_uuid"`
	MemberSaleOrderUuid uint64 `gorm:"column:member_sale_order_uuid;type:bigint(20);default:0;comment:会员销售订单ID" json:"member_sale_order_uuid"`

	// 必点方案相关字段
	ShowMustPlan       uint `gorm:"column:show_must_plan;type:tinyint(1);default:1;comment:是否显示必点方案, 0-不显示 1-显示" json:"show_must_plan"`
	AutoAddMustProduct uint `gorm:"column:auto_add_must_product;type:tinyint(1);default:1;comment:是否自动加购必点商品, 0-不自动加购 1-自动加购" json:"auto_add_must_product"`

	// 2.4.0 版本新增字段，厨显端是否确认退菜整单
	IsKitchenConfirm uint `gorm:"column:is_kitchen_confirm;type:tinyint(1);default:0;comment:厨显端是否确认退菜整单, 0-否 1-是" json:"is_kitchen_confirm"`

	// 2.5.0 版本新增字段， 反结账次数
	ReverseSettleCount uint `gorm:"column:reverse_settle_count;type:int(11);default:0;comment:反结账次数" json:"reverse_settle_count"`

	// 分批类型颜色
	BatchTagUuid uint64 `gorm:"column:batch_tag_uuid;type:bigint(20);default:0;comment:分批类型UUID" json:"batch_tag_uuid"`

	// 关联模型
	SaleOrders      []*SaleOrder      `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	H5OrderProducts []*H5OrderProduct `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	SaleBillSetting *SaleBillSetting  `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	Cashier         Staff             `gorm:"foreignKey:CashierUuid;references:uuid"`
	Desk            *Desk             `gorm:"foreignKey:DeskUuid;references:uuid"`
	BuffetPackage1  *BuffetPackage    `gorm:"foreignKey:BuffetPackage1Uuid;references:uuid"`
	BuffetPackage2  *BuffetPackage    `gorm:"foreignKey:BuffetPackage2Uuid;references:uuid"`
	BatchTag        *BatchTag         `gorm:"foreignKey:BatchTagUuid;references:uuid"`
	OrderSource     *OrderSource      `gorm:"foreignKey:OrderSourceUuid;references:uuid"`
	Nationality     *Nationality      `gorm:"foreignKey:NationalityUuid;references:uuid"`
}

// 获取自助餐套餐名称
func (model *SaleBill) GetBuffetPackageName() MultiLanguageName {
	if model.BuffetPackage1 != nil && model.BuffetPackage1 == nil {
		return model.BuffetPackage1.MultiLanguageName
	}
	if model.BuffetPackage2 != nil && model.BuffetPackage1 == nil {
		return model.BuffetPackage2.MultiLanguageName
	}
	if model.BuffetPackage1 != nil && model.BuffetPackage2 != nil {
		return MultiLanguageName{
			EnName:   model.BuffetPackage1.MultiLanguageName.EnName + " + " + model.BuffetPackage2.MultiLanguageName.EnName,
			ZhName:   model.BuffetPackage1.MultiLanguageName.ZhName + " + " + model.BuffetPackage2.MultiLanguageName.ZhName,
			ThName:   model.BuffetPackage1.MultiLanguageName.ThName + "+ " + model.BuffetPackage2.MultiLanguageName.ThName,
			ZhTwName: model.BuffetPackage1.MultiLanguageName.ZhTwName + " + " + model.BuffetPackage2.MultiLanguageName.ZhTwName,
			JaName:   model.BuffetPackage1.MultiLanguageName.JaName + "+ " + model.BuffetPackage2.MultiLanguageName.JaName,
			KoName:   model.BuffetPackage1.MultiLanguageName.KoName + "+ " + model.BuffetPackage2.MultiLanguageName.KoName,
			MyName:   model.BuffetPackage1.MultiLanguageName.MyName + "+ " + model.BuffetPackage2.MultiLanguageName.MyName,
			TrName:   model.BuffetPackage1.MultiLanguageName.TrName + "+ " + model.BuffetPackage2.MultiLanguageName.TrName,
			SvName:   model.BuffetPackage1.MultiLanguageName.SvName + "+ " + model.BuffetPackage2.MultiLanguageName.SvName,
		}
	}
	if model.BuffetPackage1 == nil && model.BuffetPackage2 == nil {
		return MultiLanguageName{}
	}
	return MultiLanguageName{}
}

// 是否是“已取消”状态
func (model *SaleBill) IsCanceled() bool {
	return model.Status == constant.SaleBillStatusCanceled
}

// 获取整单备注信息
func (model *SaleBill) GetOrderRemark() (*resp.OrderRemarkInfo, error) {
	jsonStr := model.OrderRemark
	if jsonStr == "" {
		return nil, nil
	}
	data := resp.OrderRemarkInfo{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// 获取整单备注信息
func (model *SaleBill) GetOrderRemarkRes() (*resp.OrderRemarkRes, error) {
	orderRemark, err := model.GetOrderRemark()
	if err != nil {
		return nil, err
	}
	if orderRemark == nil {
		return nil, nil
	}
	result := orderRemark.GetOrderRemarkResponse()
	return &result, nil
}

// 获取整单备注信息, 最新的一条
func (model *SaleBill) GetLatestOrderRemarkRes() *resp.OrderRemarkResItem {
	orderRemark, err := model.GetOrderRemarkRes()
	if err != nil {
		return nil
	}
	if orderRemark == nil {
		return nil
	}
	if len(orderRemark.List) == 0 {
		return nil
	}
	for _, remark := range orderRemark.List {
		if remark.IsLatest {
			return &remark
		}
	}
	return nil
}

// 设置为“已取消”状态
func (model *SaleBill) SetCanceled() {
	model.Status = constant.SaleBillStatusCanceled
	// 未完整逻辑。部分逻辑还在service中处理。
}

// 获取预送厨的商品
func (model *SaleBill) GetSaleOrderProductPreCooking() []*SaleOrderProduct {
	preCookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsUnAcceptOrderBool() {
				continue
			}
			if saleOrderProduct.IsBatchBool() && saleOrderProduct.IsPreCooking() {
				preCookingSaleOrderProducts = append(preCookingSaleOrderProducts, saleOrderProduct)
			}
		}
	}
	return preCookingSaleOrderProducts
}

// 获取分批送厨的商品
func (model *SaleBill) GetSaleOrderProductBatchCooking() []*SaleOrderProduct {
	batchCookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsUnAcceptOrderBool() {
				continue
			}
			if saleOrderProduct.IsBatchBool() {
				batchCookingSaleOrderProducts = append(batchCookingSaleOrderProducts, saleOrderProduct)
			}
		}
	}
	return batchCookingSaleOrderProducts
}

// 获取分批送厨的商品,指定saleOrderUuids
func (model *SaleBill) GetSaleOrderProductBatchCookingBySaleOrderUuid(saleOrderProductUuids []uint64) []*SaleOrderProduct {
	if len(saleOrderProductUuids) == 0 {
		return make([]*SaleOrderProduct, 0)
	}
	batchCookingSaleOrderProducts := model.GetSaleOrderProductBatchCooking()
	uuidMap := make(map[uint64]bool)
	for _, saleOrderProductUuid := range saleOrderProductUuids {
		uuidMap[saleOrderProductUuid] = true
	}

	products := make([]*SaleOrderProduct, 0)
	for _, saleOrderProduct := range batchCookingSaleOrderProducts {
		if uuidMap[saleOrderProduct.Uuid] && saleOrderProduct.BatchTime == 0 {
			products = append(products, saleOrderProduct)
		}
	}
	return products
}

// 结束SaleBill和SaleOrder的生命周期。相当于用餐订单结账完成。
func (model *SaleBill) EndSaleBillAndSaleOrder(dutyNo string, cashierUuid uint64, cashierName string) {
	model.SetFinishSaleBill(dutyNo, cashierUuid, cashierName)

	for _, saleOrder := range model.SaleOrders {
		paymentInfoList := saleOrder.GetPaymentInfoList()
		totalPay := float64(0) // 总付款金额=各个付款单的实收金额之和
		for _, paymentOrder := range paymentInfoList {
			totalPay = decimal.NewFromFloat(totalPay).Add(decimal.NewFromFloat(paymentOrder.Amount)).Round(2).InexactFloat64()
		}
		saleOrder.Status = constant.SaleOrderStatusFinish
		// 修改订单为支付完成，并记录找零金额、最终付款金额等结算后才计算的字段
		final := FinalAmount{
			PaymentAmount:        totalPay,
			ChangeAmount:         0,
			ZeroCheckoutFee:      0,
			FinalPrice:           totalPay,
			PaymentCommissionFee: 0,
			GiftAmount:           0,
		}
		saleOrder.SetFinishStatus(final) // 设置销售订单状态为已结清
	}
	model.CalcAll()
}

// IsReverseSettle 是否是反结账
func (model *SaleBill) IsReverseSettle() bool {
	return model.ReverseSettleCount > 0
}

// 是否是点餐订单
func (model *SaleBill) IsInstantBill() bool {
	return model.BillType == constant.SaleBillTypeInstant
}

// 是否是桌台订单
func (model *SaleBill) IsDeskBill() bool {
	return model.BillType == constant.SaleBillTypeDesk
}

// 是否是外送订单
func (model *SaleBill) IsTakeoutBill() bool {
	return model.BillType == constant.SaleBillTypeTakeout || model.MemberSaleOrderUuid != 0
}

// 是否是订单来源为外卖的订单
func (model *SaleBill) IsOrderSourceTakeout() bool {
	return model.OrderSourceUuid > 0
}

// 获取订单来源为外卖的文本
func (model *SaleBill) GetOrderSourceTakeoutText() string {
	if model.IsOrderSourceTakeout() {
		return "外卖"
	}
	return ""
}

// 销售账单是否已经使用了通用优惠券
func (model *SaleBill) IsCommonCouponUsed() (bool, uint64) {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsDelete() {
			continue
		}
		for _, coupon := range saleOrder.Coupons {
			if coupon.IsDelete() {
				continue
			}
			if coupon.IsCommonCoupon() {
				return true, saleOrder.Uuid
			}
		}
	}
	return false, 0
}

// 获取销售账单的自助餐商品列表。key为自助餐商品包uuid，value为true
func (model *SaleBill) GetBuffetProductMap() map[uint64]bool {
	productPackageUuidMap := make(map[uint64]bool) // 自助餐商品包uuid
	if model.BuffetPackage1 != nil {
		for _, buffetProduct := range model.BuffetPackage1.BuffetProducts {
			if buffetProduct.IsDelete() {
				continue
			}
			productPackageUuidMap[buffetProduct.ProductPackageUuid] = true
		}
	}
	if model.BuffetPackage2 != nil {
		for _, buffetProduct := range model.BuffetPackage2.BuffetProducts {
			if buffetProduct.IsDelete() {
				continue
			}
			productPackageUuidMap[buffetProduct.ProductPackageUuid] = true
		}
	}
	return productPackageUuidMap
}

// 判断销售账单是否可反结账。
// 1. 账单已完成
// 2. 未交班。
// 3. 未退款 todo 在调用反结账接口时也检查
func (model *SaleBill) IsCellReverseSettle(staffUuid uint64, cashierLoginTime int64) bool {
	// 账单未完成，不能反结账
	if model.Status != constant.SaleBillStatusComplete {
		return false
	}
	// 账单已退款，不能反结账
	if model.GetTotalRefundAmount() > 0 {
		return false
	}

	order := model.SaleOrders[0]
	// 完成 && 等于当前用户 && 在班次时间内
	if order.Status == constant.SaleBillStatusComplete && staffUuid == model.CashierUuid && order.FinishTime > cashierLoginTime {
		return true
	}
	return false
}

func (model *SaleBill) IsLockStatus() bool {
	return model.IsLock == constant.SaleBillIsLockYes
}

// 判断销售账单是否打包
func (model *SaleBill) IsTakeout() bool {
	return model.DiningMethod == constant.SaleBillDiningMethodTakeout
}

// 获取销售账单的就餐方式
func (model *SaleBill) GetDiningMethod() uint {
	return constant.SaleBillDiningMethodMap[model.DiningMethod]
}

// 判断账单是否为已送厨状态
func (model *SaleBill) IsCookingStatus() bool {
	// 如果账单已经送出过一次，就记录第一次送厨的时间
	return model.ProductionTime > 0
}

// 判断账单是否为结束状态，包括已完成、已取消、已删除
func (model *SaleBill) IsEndStatus() bool {
	return model.Status == constant.SaleBillStatusComplete || model.Status == constant.SaleBillStatusCanceled || model.IsDelete()
}

// 判断自助餐是否限时
func (model *SaleBill) IsTimeLimited() bool {
	// 自助餐时长为0，表示不限时
	if model.BuffetDuration == 0 {
		return false
	}
	// 自助餐时长大于0，表示限时
	return model.BuffetDuration > 0
}

// 判断销售账单是否拆单
func (model *SaleBill) IsSplit() bool {
	return len(model.SaleOrders) > 1 || model.IsSplitOrder == constant.SaleBillIsSplitOrderYes
}

// 是否完成
func (model *SaleBill) IsFinish() bool {
	return model.Status == constant.SaleBillStatusComplete
}

// 是否存在已支付的销售订单
func (model *SaleBill) IsExistPaid() bool {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsPaid() {
			return true
		}
	}
	return false
}

// 判断销售账单是否部分支付
func (model *SaleBill) IsPartialPay() bool {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsPartialPay() {
			return true
		}
	}
	return false
}

// 是否显示必点方案
func (model *SaleBill) IsShowMustPlan() bool {
	return model.ShowMustPlan == constant.SaleBillShowMustPlanYes
}

func (model *SaleBill) IsAutoAddMustProduct() bool {
	return model.AutoAddMustProduct == constant.AutoAddMustProductYes
}

// 是否已取单的销售账单
func (model *SaleBill) IsShowSaleBill() bool {
	return model.HideBillTime == 0
}

// 判断是否为桌台销售账单
func (model *SaleBill) IsDeskSaleBill() bool {
	return model.DeskUuid != 0 // 桌台账单肯定是有桌台ID
}

// 判断是否为点餐销售账单
func (model *SaleBill) IsInstantSaleBill() bool {
	return !model.IsDeskSaleBill() // 点餐账单肯定是没有桌台ID
}

// 判断是否为自助餐销售账单
func (model *SaleBill) IsBuffetSaleBill() bool {
	if model.BuffetPackage1Uuid != 0 || model.BuffetPackage2Uuid != 0 {
		return true
	}
	return false
}

// 转台
func (model *SaleBill) ChangeDesk(deskUuid uint64, serialNo string) {
	model.DeskUuid = deskUuid
	model.SerialNo = serialNo
	// 设置旧桌台关闭
	model.Desk.SetCloseDesk()
}

// 判断销售账单已经是可以完成，即所有销售订单都已经结账
func (model *SaleBill) CanFinishSaleBill() bool {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsFreeSaleOrder() || saleOrder.IsDelete() {
			continue
		}
		// 只要有一个销售订单未结账，这个销售账单就不可以点完成
		if !saleOrder.IsSettled() {
			return false
		}
	}
	return true
}

// 判断自助餐是否超时
func (model *SaleBill) BuffetIsTimeOut() bool {
	return model.GetBuffetRemainingSeconds() == 0
}

// 添加自助餐加钟时间
func (model *SaleBill) AddBuffetDelayStartTimeAndDuration(delayTime int) {
	remainingDelayDuration := model.GetRemainingDelayDuration()
	if remainingDelayDuration == 0 {
		model.DelayStartTime = time.Now().Unix()
		model.DelayDuration = uint(delayTime)
	} else {
		model.DelayDuration = model.DelayDuration + uint(delayTime)
	}
}

// AddSaleOrderBuffetCustomerType 将自助餐添加到销售账单的对应销售订单中
func (model *SaleBill) AddSaleOrderBuffetCustomerType(saleOrderUuid uint64, customer *SaleOrderBuffetCustomerType) {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.Uuid == saleOrderUuid {
			// 将customer添加到销售订单的SaleOrderBuffetCustomerTypes列表中
			if saleOrder.SaleOrderBuffetCustomerTypes == nil {
				saleOrder.SaleOrderBuffetCustomerTypes = make([]*SaleOrderBuffetCustomerType, 0)
			}
			saleOrder.SaleOrderBuffetCustomerTypes = append(saleOrder.SaleOrderBuffetCustomerTypes, customer)
			break
		}
	}
}

// DeleteSaleOrderBuffetCustomerTypeAll 将自助餐从销售账单的对应销售订单中删除
func (model *SaleBill) DeleteSaleOrderBuffetCustomerTypeAll(saleOrderUuid uint64) {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.Uuid == saleOrderUuid {
			for _, customerType := range saleOrder.SaleOrderBuffetCustomerTypes {
				customerType.DeleteTime = time.Now().Unix()
			}
			break
		}
	}
}

// AddSaleOrderBuffetDelayProduct 将自助餐加钟产品添加到销售账单的对应销售订单中
func (model *SaleBill) AddSaleOrderBuffetDelayProduct(saleOrderUuid uint64, delayProduct SaleOrderBuffetDelayProduct) {
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.Uuid == saleOrderUuid {
			// 将delayProduct添加到销售订单的SaleOrderBuffetDelayProducts列表中
			if saleOrder.SaleOrderBuffetDelayProducts == nil {
				saleOrder.SaleOrderBuffetDelayProducts = make([]*SaleOrderBuffetDelayProduct, 0)
			}
			// 创建一个新的指针，指向delayProduct的副本
			delayProductPtr := &SaleOrderBuffetDelayProduct{
				BaseModel:       delayProduct.BaseModel,
				Name:            delayProduct.Name,
				Num:             delayProduct.Num,
				Sign:            delayProduct.Sign,
				Price:           delayProduct.Price,
				DelayTime:       delayProduct.DelayTime,
				SaleOrderUuid:   delayProduct.SaleOrderUuid,
				BuffetDelayUuid: delayProduct.BuffetDelayUuid,
			}
			saleOrder.SaleOrderBuffetDelayProducts = append(saleOrder.SaleOrderBuffetDelayProducts, delayProductPtr)
			break
		}
	}
}

type ValidateOrderStatusOption struct {
	NotCheckSubOrder bool // 是否不检查子订单,是否已经结账。场景：助手端的话，只要还有一个订单未结账，就可以操作
}

func WithIsAssistant() func(*ValidateOrderStatusOption) {
	return func(option *ValidateOrderStatusOption) {
		option.NotCheckSubOrder = true
	}
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleBill) ValidateOrderStatus(source string, operation string, saleOrderUuid uint64, options ...func(*ValidateOrderStatusOption)) error {
	option := &ValidateOrderStatusOption{}
	for _, opt := range options {
		opt(option)
	}

	// 解锁订单
	if operation == constant.OrderUnlock {
		return nil
	}

	// 当不是收银端的时候，拆单不可操作结账
	if source != constant.SourceCashier && operation == constant.OrderSettle && model.IsSplit() {
		return errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	// 当不是收银端的时候，拆单不可操作取消订单
	if !slices.Contains([]string{
		constant.SourceShop,
		constant.SourceCashier,
		constant.SourceAssistant,
	}, source) && operation == constant.OrderOrderCancel && model.IsSplit() {
		return errors.New("拆单不可操作")
	}

	// 锁定订单 - 白名单
	if model.IsLockStatus() && !slices.Contains([]string{constant.OrderSettle}, operation) {
		// 收银端可以操作折扣
		if source != constant.SourceCashier || !slices.Contains([]string{constant.OrderDiscount, constant.OrderCancelDiscount}, operation) {
			return errors.New("订单已被锁定，请解锁后重新操作")
		}
	}
	// 账单已退款，不能操作
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.IsDelete() {
		return errors.New("订单已删除")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}

	if option.NotCheckSubOrder {
		// 助手端的话，只要还有一个订单未结账，就可以操作
		return nil
	} else {
		if len(model.SaleOrders) > 0 {
			// 单个订单不能操作
			for _, so := range model.SaleOrders {
				if saleOrderUuid == 0 || saleOrderUuid == so.Uuid {
					if err := so.ValidateOrderStatus(); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
		}
	}
	return nil
}

// CopyOrderProductAndEdit 复制订单商品并编辑
func (model *SaleBill) CopyOrderProductAndEdit(saleOrderProductUuid uint64, updateProduct SaleOrderProduct, specialFields ...map[string]bool) {
	for i, order := range model.SaleOrders {
		for j, product := range order.SaleOrderProducts {
			if product.Uuid == saleOrderProductUuid {
				product := model.SaleOrders[i].SaleOrderProducts[j]
				// 克隆订单商品
				newSaleOrderProduct := product.CopyOrderProduct(order.Uuid)
				// 如果传入了 specialFields，使用第一个元素，否则传递 nil
				var fields map[string]bool
				if len(specialFields) > 0 {
					fields = specialFields[0]
				}
				newSaleOrderProduct.SetFields(updateProduct, fields)
				newSaleOrderProduct.SetUpdate()
				// 更新订单商品
				model.SaleOrders[i].SaleOrderProducts = append(model.SaleOrders[i].SaleOrderProducts, newSaleOrderProduct)
				break
			}
		}
	}
}

// NewH5Order 创建h5订单
func (model *SaleBill) NewH5Order(companySetting CompanySetting) (*H5Order, error) {
	h5OrderUuid, _ := utils.GetID()
	// 获取未下单的h5订单商品
	h5OrderProducts := model.GetUnOrderH5OrderProduct()
	if len(h5OrderProducts) == 0 {
		return nil, errors.WithMessage(errors.New("请先加购商品"))
	}
	h5OrderProductList := make([]*H5OrderProduct, 0)
	for _, h5OrderProduct := range h5OrderProducts {
		h5OrderProductList = append(h5OrderProductList, &H5OrderProduct{
			SaleOrderProductUuid: h5OrderProduct.Uuid, // 销售订单商品uuid
			H5OrderUuid:          h5OrderUuid,         // 扫码订单uuid
			SaleBillUuid:         model.Uuid,          // 销售账单uuid
		})
	}

	// 创建h5订单
	return &H5Order{
		BaseModel: BaseModel{
			Uuid: h5OrderUuid,
		},
		DeskUuid:        model.DeskUuid,                   // 桌台uuid
		SaleOrderUuid:   h5OrderProducts[0].SaleOrderUuid, // 销售订单uuid
		SaleBillUuid:    model.Uuid,                       // 销售账单uuid
		DeskNo:          model.Desk.DeskNo,                // 桌台编号
		Status:          constant.H5OrderStatusOrder,      // 状态，已下单
		H5OrderProducts: h5OrderProductList,               // 订单商品
		IsNeedAudit:     companySetting.IsOpenH5Order,     // 是否需要审核，关闭商家扫码点餐接单，则不需要审核
	}, nil
}
