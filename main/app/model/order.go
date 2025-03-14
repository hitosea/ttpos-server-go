package model

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

// SaleBill 销售账单 `ttpos_sale_bill`
type SaleBill struct {
	BaseModel
	// 主键和标识字段
	OrderNo  string `gorm:"column:order_no;type:varchar(255);default:'';comment:销售账单编号" json:"order_no"`
	DutyNo   string `gorm:"column:duty_no;type:varchar(255);default:'';comment:当班编号,用于标记该账单属于哪个当班" json:"duty_no"`
	SerialNo string `gorm:"column:serial_no;type:varchar(255);default:'';comment:桌位编号 (点餐流水号)" json:"serial_no"`

	// 状态相关字段
	Status uint `gorm:"column:status;type:tinyint(1);default:0;comment:订单状态, 0-待付款、1-已完成、2-已取消" json:"status"`
	IsLock uint `gorm:"column:is_lock;type:tinyint(1);default:0;comment:是否锁单, 0-否 1-是" json:"is_lock"`

	// 订单类型字段
	BillType        uint  `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-桌台订单、1-点餐订单" json:"bill_type"`
	DiningMethod    uint  `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	IsBuffet        uint  `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	BuffetDuration  uint  `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒），0为不限时. 原始值为自助餐的时长，加钟时会累加" json:"buffet_duration"`
	BuffetStartTime int64 `gorm:"column:buffet_start_time;type:int(10);default:0;comment:自助餐开始时间（秒）" json:"buffet_start_time"`
	DelayDuration   uint  `gorm:"column:delay_duration;type:int(10);default:0;comment:总延迟时长（秒）" json:"delay_duration"`
	DelayStartTime  int64 `gorm:"column:delay_start_time;type:int(10);default:0;comment:总延迟时长开始时间（秒）" json:"delay_start_time"`

	// 订单基本信息
	MealNum uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Remark  string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	Reason  string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 金额字段 - 主要金额
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
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
	FreeAmount        float64 `gorm:"column:free_amount;type:decimal(12,2);default:0;comment:免单金额,关联销售订单的免单金额之和" json:"free_amount"`

	// 时间相关字段
	FinishTime     int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	HideBillTime   int64 `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`
	ProductionTime int64 `gorm:"column:production_time;type:int(10);default:0;comment:首次送厨时间（时间戳）" json:"production_time"`

	// 收银员名称
	CashierName string `gorm:"column:cashier_name;type:varchar(255);default:'';comment:收银员名称" json:"cashier_name"`

	// 关联ID字段
	ConsumerUuid       uint64 `gorm:"column:consumer_uuid;type:bigint(20);default:0;comment:消费者ID" json:"consumer_uuid"`
	CashierUuid        uint64 `gorm:"column:cashier_uuid;type:bigint(20);default:0;comment:收银员ID" json:"cashier_uuid"`
	DeskUuid           uint64 `gorm:"column:desk_uuid;type:bigint(20);default:0;comment:餐桌ID" json:"desk_uuid"`
	BuffetPackage1Uuid uint64 `gorm:"column:buffet_package1_uuid;type:bigint(20);default:0;comment:自助餐套餐1ID" json:"buffet_package1_uuid"`
	BuffetPackage2Uuid uint64 `gorm:"column:buffet_package2_uuid;type:bigint(20);default:0;comment:自助餐套餐2ID" json:"buffet_package2_uuid"`
	DeviceUuid         uint64 `gorm:"column:device_uuid;type:bigint(20);default:0;comment:设备ID，用于标识这个账单是由哪个设备创建的。点餐账单通过设备uuid查询" json:"device_uuid"`

	// 必点方案相关字段
	ShowMustPlan       uint `gorm:"column:show_must_plan;type:tinyint(1);default:1;comment:是否显示必点方案, 0-不显示 1-显示" json:"show_must_plan"`
	AutoAddMustProduct uint `gorm:"column:auto_add_must_product;type:tinyint(1);default:1;comment:是否自动加购必点商品, 0-不自动加购 1-自动加购" json:"auto_add_must_product"`

	// 关联模型
	SaleOrders      []*SaleOrder      `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	H5OrderProducts []*H5OrderProduct `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	SaleBillSetting *SaleBillSetting  `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	Cashier         Staff             `gorm:"foreignKey:CashierUuid;references:uuid"`
	Desk            *Desk             `gorm:"foreignKey:DeskUuid;references:uuid"`
	BuffetPackage1  *BuffetPackage    `gorm:"foreignKey:BuffetPackage1Uuid;references:uuid"`
	BuffetPackage2  *BuffetPackage    `gorm:"foreignKey:BuffetPackage2Uuid;references:uuid"`
}

func NewDeskSaleBill(saleBillUuid uint64, orderNo string, buffetUuids []uint64, mealNum uint, remark string, deskUuid uint64, serialNo string) *SaleBill {
	isBuffet := len(buffetUuids) > 0

	if saleBillUuid == 0 {
		saleBillUuid, _ = utils.GetID()
	}

	saleBill := &SaleBill{
		BaseModel:    BaseModel{Uuid: saleBillUuid},
		OrderNo:      orderNo,
		BillType:     constant.OrderSourceMapToBillType[constant.OrderSourceDesk],
		DiningMethod: constant.SaleBillDiningMethodDineIn,
		IsBuffet:     utils.BoolToUint(isBuffet),
		MealNum:      mealNum, // 非自助餐订单，就餐人数等于开台时填写的人数。 自助餐订单，就餐人数等于各个顾客类型数量的累加，如老人2人、小孩3人，则就餐人数为5人。不会因为销售账单是两个自助餐套餐而导致人数变为10人
		Remark:       remark,
		DeskUuid:     deskUuid,
		SerialNo:     serialNo,
	}

	// 设置自助餐套餐
	if isBuffet {
		saleBill.SetBuffetPackage(buffetUuids)
		saleBill.BuffetStartTime = time.Now().Unix()
	}

	return saleBill
}
func (model *SaleBill) SetNil() {
	model.SaleOrders = nil
	model.H5OrderProducts = nil
	model.SaleBillSetting = nil
	model.Cashier = Staff{}
	model.Desk = nil
	model.BuffetPackage1 = nil
	model.BuffetPackage2 = nil
}

func (model *SaleBill) GetBuffetProductList() resp.BuffetProductList {
	buffetProductList := resp.BuffetProductList{}
	buffetProductList.List = make([]resp.BuffetProduct, 0)
	// 去重的列表
	list := make([]resp.BuffetProduct, 0)
	if model.BuffetPackage1 != nil {
		buffetProductList.List = append(buffetProductList.List, model.BuffetPackage1.GetBuffetProductList().List...)
	}
	if model.BuffetPackage2 != nil {

		buffetProductList.List = append(buffetProductList.List, model.BuffetPackage2.GetBuffetProductList().List...)
	}
	// 去重
	buffetProductMap := make(map[uint64]bool)
	for _, buffetProduct := range buffetProductList.List {
		if !buffetProductMap[buffetProduct.Uuid] {
			buffetProductMap[buffetProduct.Uuid] = true
			list = append(list, buffetProduct)
		}
	}
	buffetProductList.List = list
	return buffetProductList
}

// 判断销售账单是否可反结账。
// 1. 账单已完成
// 2. 未交班。 todo
// 3. 未退款 todo 在调用反结账接口时也检查
func (model *SaleBill) IsCellReverseSettle() bool {
	// 账单未完成，不能反结账
	if model.Status != constant.SaleBillStatusComplete {
		return false
	}
	// 账单已退款，不能反结账
	if model.GetTotalRefundAmount() > 0 {
		return false
	}
	return true
}

// 设置打包销售账单。并更新订单的税率
func (model *SaleBill) SetTakeoutSaleBill(diningMethod uint) {
	// 如果没有改变，则不更新
	if model.DiningMethod == diningMethod {
		return
	}
	// 默认堂食
	method := constant.SaleBillDiningMethodDineIn
	// 严谨判断。拒绝非法的值
	if diningMethod == constant.SaleBillDiningMethodTakeout {
		method = constant.SaleBillDiningMethodTakeout
	}
	// 严谨判断。拒绝非法的值
	if diningMethod == constant.SaleBillDiningMethodDineIn {
		method = constant.SaleBillDiningMethodDineIn
	}
	model.DiningMethod = uint(method)
	model.SetUpdate() // 标记要更新model

	// 由于就餐方式改变，导致税率改变，要重新计算账单金额
	// 从ProductPackage中获取税率
	for _, saleOrder := range model.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			if saleOrderProduct.ProductPackage == nil {
				// 这种情况不应该出现，因为商品包是必填的.panic是为了提示没有预加载这个表
				panic("saleOrderProduct.ProductPackage is nil")
			}
			taxRate := saleOrderProduct.ProductPackage.TaxRate(model.DiningMethod)
			saleOrderProduct.SetTaxRate(taxRate)
		}
	}
}

// 设置反结账
func (model *SaleBill) SetReverseSettle() {
	// 销售账单状态变为待付款
	// 销售订单状态变为未结账状态
	// 销售订单的所有付款单都退款，并生成退款单
	model.Status = constant.SaleBillStatusPending
	for _, saleOrder := range model.SaleOrders {
		saleOrder.Status = constant.SaleOrderStatusPending
		for _, paymentOrder := range saleOrder.PaymentOrders {
			paymentOrder.Status = constant.PaymentOrderStatusRefund
			fmt.Println("paymentOrder.Status", paymentOrder.Status)
		}
	}
}

// 获取支付方式名称列表. 获取所有子单的付款单用到的支付方式，不重复
func (model *SaleBill) GetPaymentMethodNameList() []string {
	payMethods := make(map[string]bool)
	for _, saleOrder := range model.SaleOrders {
		for _, paymentOrder := range saleOrder.PaymentOrders {
			payMethods[paymentOrder.PaymentMethodName] = true
		}
	}
	payMethodList := make([]string, 0)
	for payMethod := range payMethods {
		payMethodList = append(payMethodList, payMethod)
	}
	return payMethodList
}

func (model *SaleBill) IsLockStatus() bool {
	return model.IsLock == constant.SaleBillIsLockYes
}

func (model *SaleBill) IsTakeout() bool {
	return model.DiningMethod == constant.SaleBillDiningMethodTakeout
}

// 设置账单为已送厨状态。如果状态已经是送厨，则不修改
func (model *SaleBill) SetCookingStatus() {
	if model.IsCookingStatus() {
		return
	}
	model.SetUpdate()
	model.ProductionTime = time.Now().Unix()
}

func (model *SaleBill) calcPaymentCommissionFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.PaymentCommissionFee))
	}
	return amount.InexactFloat64()
}

func (model *SaleBill) calcPaymentAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.PaymentAmount))
	}
	return amount.InexactFloat64()
}

func (model *SaleBill) calcProductOriginalAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.ProductOriginalAmount))
	}
	return amount.InexactFloat64()
}

// 判断账单是否为已送厨状态
func (model *SaleBill) IsCookingStatus() bool {
	// 如果账单已经送出过一次，就记录第一次送厨的时间
	return model.ProductionTime > 0
}

// 判断账单是否为结束状态，包括已完成、已取消
func (model *SaleBill) IsEndStatus() bool {
	return model.Status == constant.SaleBillStatusComplete || model.Status == constant.SaleBillStatusCanceled
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
		// 只要有一个未完成支付，这个销售账单就不可以点完成
		if saleOrder.Status != constant.SaleOrderStatusFinish {
			return false
		}
	}
	return true
}

// 设置销售账单完成
func (model *SaleBill) SetFinishSaleBill() {
	model.Status = constant.SaleBillStatusComplete
	model.FinishTime = time.Now().Unix()
}

// 设置自助餐套餐
func (model *SaleBill) SetBuffetPackage(buffetPackageUuids []uint64) {
	if len(buffetPackageUuids) == 1 {
		model.BuffetPackage1Uuid = buffetPackageUuids[0]
		model.BuffetPackage2Uuid = 0
	}
	if len(buffetPackageUuids) == 2 {
		model.BuffetPackage1Uuid = buffetPackageUuids[0]
		model.BuffetPackage2Uuid = buffetPackageUuids[1]
	}
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
	return len(model.SaleOrders) > 1
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

// 获取未送厨的销售订单商品
func (model *SaleBill) GetSaleOrderProductUnCooking() []*SaleOrderProduct {
	unCookingSaleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for i, _ := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if !orderProduct.IsAcceptOrderBool() && orderProduct.IsDelete() {
				continue
			}
			if orderProduct.Status == constant.SaleOrderProductStatusNormal {
				unCookingSaleOrderProducts = append(unCookingSaleOrderProducts, saleOrder.SaleOrderProducts[i])
			}
		}
	}
	return unCookingSaleOrderProducts
}

// 获取已送厨和未送厨的销售订单商品
func (model *SaleBill) GetSaleOrderProductAll() []*SaleOrderProduct {
	saleOrderProducts := make([]*SaleOrderProduct, 0)
	for _, saleOrder := range model.SaleOrders {
		for i, _ := range saleOrder.SaleOrderProducts {
			orderProduct := saleOrder.SaleOrderProducts[i]
			if !orderProduct.IsAcceptOrderBool() && orderProduct.IsDelete() {
				continue
			}
			saleOrderProducts = append(saleOrderProducts, saleOrder.SaleOrderProducts[i])
		}
	}
	return saleOrderProducts
}

// 计算销售账单的总金额。总金额=销售订单的应收金额之和
func (model *SaleBill) calcAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.Amount))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的商品金额。商品金额=销售订单的商品金额之和
func (model *SaleBill) calcProductAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.ProductAmount))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的服务费。服务费=销售订单的服务费之和
func (model *SaleBill) calcServiceFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.ServiceFee))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的税费。税费=销售订单的税费之和
func (model *SaleBill) calcTaxFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.TaxFee))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的折扣费用。折扣费用=销售订单的折扣费用之和
func (model *SaleBill) calcDiscountFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.CustomDiscountFee))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的会员折扣费用。会员折扣费用=销售订单的会员折扣费用之和
func (model *SaleBill) calcMemberDiscountFee() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		amount = amount.Add(decimal.NewFromFloat(saleOrder.MemberDiscountFee))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的赠菜金额。赠菜金额=销售订单的赠菜金额之和
func (model *SaleBill) calcGiftAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		// 赠菜金额=销售订单的赠菜金额之和 累加
		amount = amount.Add(decimal.NewFromFloat(saleOrder.GiftAmount))
	}
	return amount.InexactFloat64()
}

// 计算销售订单的免单金额。免单金额=销售订单的免单金额之和
func (model *SaleBill) calcFreeAmount() float64 {
	amount := decimal.NewFromFloat(0)
	for _, saleOrder := range model.SaleOrders {
		if saleOrder.IsFreeSaleOrder() {
			amount = amount.Add(decimal.NewFromFloat(saleOrder.Amount))
		}
	}
	return amount.InexactFloat64()
}

// 重新计算销售账单的金额
func (model *SaleBill) CalcAll() {
	setting := model.SaleBillSetting
	for i, _ := range model.SaleOrders {
		saleOrder := model.SaleOrders[i]
		for j, _ := range saleOrder.SaleOrderProducts {
			saleOrderProduct := saleOrder.SaleOrderProducts[j]
			// 如果订单商品已删除或已取消，则不计算
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() {
				continue
			}
			_ = saleOrderProduct.BeforeCalc()
			_ = saleOrderProduct.CalcSaleOrderProduct(*setting)
		}
		// 计算自助餐顾客价格之和
		for _, buffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
			if buffetCustomer.IsDelete() {
				continue
			}
			_ = buffetCustomer.CalcSaleOrderBuffetCustomerType(*setting)
		}
		saleOrder.CalcSaleOrder(*setting)
	}
	model.CalcSaleBill()
}

type SaleBillCalc struct {
	Amount            float64 `json:"amount"`              // 订单总金额=销售订单的应收金额之和
	ProductAmount     float64 `json:"product_amount"`      // 商品金额=销售订单的商品金额之和
	ServiceFee        float64 `json:"service_fee"`         // 服务费=销售订单的服务费之和
	TaxFee            float64 `json:"tax_fee"`             // 税费=销售订单的税费之和
	DiscountFee       float64 `json:"discount_fee"`        // 折扣费用=销售订单的折扣费用之和
	MemberDiscountFee float64 `json:"member_discount_fee"` // 会员折扣费用=销售订单的会员折扣费用之和
	GiftAmount        float64 `json:"gift_amount"`         // 赠菜金额=销售订单的赠菜金额之和
	FreeAmount        float64 `json:"free_amount"`         // 免单金额=销售订单的免单金额之和

	ProductOriginalAmount float64 `json:"product_original_amount"` //
	PaymentAmount         float64
	PaymentCommissionFee  float64
}

// 重新计算销售订单的金额
func (model *SaleBill) CalcSaleBill() *SaleBillCalc {
	calc := SaleBillCalc{}
	calc.Amount = model.calcAmount()
	model.Amount = calc.Amount
	calc.ProductAmount = model.calcProductAmount()
	model.ProductAmount = calc.ProductAmount
	calc.ServiceFee = model.calcServiceFee()
	model.ServiceFee = calc.ServiceFee
	calc.TaxFee = model.calcTaxFee()
	model.TaxFee = calc.TaxFee
	calc.DiscountFee = model.calcDiscountFee()
	model.CustomDiscountFee = calc.DiscountFee
	calc.MemberDiscountFee = model.calcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.GiftAmount = model.calcGiftAmount()
	model.GiftAmount = calc.GiftAmount
	calc.FreeAmount = model.calcFreeAmount()
	model.FreeAmount = calc.FreeAmount
	calc.ProductOriginalAmount = model.calcProductOriginalAmount()
	model.ProductOriginalAmount = calc.ProductOriginalAmount
	calc.PaymentAmount = model.calcPaymentAmount()
	model.PaymentAmount = calc.PaymentAmount
	calc.PaymentCommissionFee = model.calcPaymentCommissionFee()
	model.PaymentCommissionFee = calc.PaymentCommissionFee
	return &calc
}

func (model *SaleBill) IsShowMustPlan() bool {
	return model.ShowMustPlan == constant.SaleBillShowMustPlanYes
}

func (model *SaleBill) IsAutoAddMustProduct() bool {
	return model.AutoAddMustProduct == constant.AutoAddMustProductYes
}

// 获取销售账单的取单列表item卡片的信息
func (model *SaleBill) GetHideSaleBillCardInfo() {

}

// 是否已取单的销售账单
func (model *SaleBill) IsShowSaleBill() bool {
	return model.HideBillTime == 0
}

// 返回新的销售账单
func (model *SaleBill) GetSaleOrder(saleOrderUuid uint64) *SaleOrder {
	for i, saleOrder := range model.SaleOrders {
		if saleOrderUuid == saleOrder.Uuid {
			if len(model.SaleOrders) > 1 {
				saleOrder.Index = i + 1
			}
			return saleOrder
		}
	}
	return nil
}

// 返回第一个销售订单
func (model *SaleBill) GetFirstSaleOrder() *SaleOrder {
	return model.SaleOrders[0]
}

// 获取销售账单的销售订单和销售订单商品
func (model *SaleBill) GetSaleOrderAndProduct(saleOrderUuid uint64, saleOrderProductUuid uint64) (*SaleOrder, *SaleOrderProduct) {
	// 获取销售订单
	saleOrder := model.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		return nil, nil
	}
	// 获取销售订单商品
	saleOrderProduct, _, err := saleOrder.GetSaleOrderProduct(saleOrderProductUuid)
	if err != nil {
		return saleOrder, nil
	}
	return saleOrder, saleOrderProduct
}

// 获取自助餐名称
func (model *SaleBill) GetBuffetName() (name dto.LocaleResponse) {
	name1 := dto.LocaleResponse{}
	name2 := dto.LocaleResponse{}
	if model.BuffetPackage1 != nil {
		name1 = model.BuffetPackage1.MultiLanguageName.GetNames()
	}
	if model.BuffetPackage2 != nil {
		name2 = model.BuffetPackage2.MultiLanguageName.GetNames()
	}
	if model.BuffetPackage1 != nil && model.BuffetPackage2 != nil {
		name = dto.LocaleResponse{
			ZH:   fmt.Sprintf("%s+%s", name1.ZH, name2.ZH),
			TH:   fmt.Sprintf("%s+%s", name1.TH, name2.TH),
			EN:   fmt.Sprintf("%s+%s", name1.EN, name2.EN),
			ZHTW: fmt.Sprintf("%s+%s", name1.ZHTW, name2.ZHTW),
			JA:   fmt.Sprintf("%s+%s", name1.JA, name2.JA),
			KO:   fmt.Sprintf("%s+%s", name1.KO, name2.KO),
			MY:   fmt.Sprintf("%s+%s", name1.MY, name2.MY),
			TR:   fmt.Sprintf("%s+%s", name1.TR, name2.TR),
		}
		return
	}
	// 只有一个自助餐时都是只填在BuffetPackage1
	if model.BuffetPackage1 != nil {
		name = dto.LocaleResponse{
			ZH:   fmt.Sprintf("%s", name1.ZH),
			TH:   fmt.Sprintf("%s", name1.TH),
			EN:   fmt.Sprintf("%s", name1.EN),
			ZHTW: fmt.Sprintf("%s", name1.ZHTW),
			JA:   fmt.Sprintf("%s", name1.JA),
			KO:   fmt.Sprintf("%s", name1.KO),
			MY:   fmt.Sprintf("%s", name1.MY),
			TR:   fmt.Sprintf("%s", name1.TR),
		}
		return
	}
	return name
}

// 判断是否为自助餐销售账单
func (model *SaleBill) IsDeskSaleBill() bool {
	return model.DeskUuid != 0 // 桌台账单肯定是有桌台ID
}

// 判断是否为自助餐销售账单
func (model *SaleBill) IsBuffetSaleBill() bool {
	if model.BuffetPackage1Uuid != 0 || model.BuffetPackage2Uuid != 0 {
		return true
	}
	return false
}

// 获取自助餐结束时间
func (model *SaleBill) GetBuffetEndTime() int64 {
	endTime := model.BuffetStartTime + int64(model.BuffetDuration)
	return endTime
}

// 判断自助餐是否超时
func (model *SaleBill) BuffetIsTimeOut() bool {
	return model.GetBuffetRemainingSeconds() == 0
}

// 自助餐还剩余多少秒
func (model *SaleBill) GetBuffetRemainingSeconds() int64 {
	if model.BuffetDuration == 0 {
		return -1
	}
	remainingTime := model.GetBuffetEndTime() - time.Now().Unix()
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

// 获取加钟剩余时长
func (model *SaleBill) GetRemainingDelayDuration() int64 {
	useDuration := time.Now().Unix() - model.DelayStartTime
	if useDuration <= 0 {
		useDuration = 0
	}
	duration := int64(model.DelayDuration) - useDuration
	if duration < 0 {
		return 0
	}
	return duration
}

// 获取总的剩余时长
func (model *SaleBill) GetTotalRemainingSeconds() int64 {
	return model.GetRemainingDelayDuration() + model.GetBuffetRemainingSeconds()
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleBill) ValidateOrderStatus(operation string, saleOrderUuid ...uint64) error {
	if operation != constant.OrderSettle && operation != constant.OrderUnlock && model.IsLockStatus() {
		return errors.New("订单已被锁定，请解锁后重新操作")
	}
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.IsDelete() {
		return errors.New("订单已删除")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	if len(model.SaleOrders) > 0 {
		// todo: 要判断来源 除了收银端一样 拆单没有取消权限
		// if operation == constant.OrderOrderCancel && len(model.SaleOrders) > 1 {
		// 	return errors.New("拆单不可操作")
		// }
		// 单个订单不能操作
		for _, so := range model.SaleOrders {
			if len(saleOrderUuid) == 0 || slices.Contains(saleOrderUuid, so.Uuid) {
				if err := so.ValidateOrderStatus(); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
	}
	return nil
}

// 设置显示销售账单(取单)
func (model *SaleBill) SetShowSaleBill(deviceUuid uint64) {
	model.HideBillTime = 0
	model.DeviceUuid = deviceUuid
}

// 设置隐藏销售账单(挂单)
func (model *SaleBill) SetHideSaleBill() {
	model.HideBillTime = time.Now().Unix()
}

// 获取总的退款金额
func (model *SaleBill) GetTotalRefundAmount() float64 {
	refundAmount := 0.0
	for _, saleOrder := range model.SaleOrders {
		for _, refundOrder := range saleOrder.ReturnOrders {
			refundAmount += refundOrder.RefundAmount
		}
	}
	return refundAmount
}

// GetAmount 获取账单的付款金额。订单金额=amount-退款金额
func (model *SaleBill) GetPaymentAmount() float64 {
	// 退款金额
	refundAmount := model.GetTotalRefundAmount()
	//订单金额=amount-退款金额
	return decimal.NewFromFloat(model.Amount).Sub(decimal.NewFromFloat(refundAmount)).InexactFloat64()
}

// 获取所有自助餐名称
func (model *SaleBill) GetBuffetNames(language string) string {
	buffets := make([]string, 0)
	for _, order := range model.SaleOrders {
		for _, buffet := range order.SaleOrderBuffetCustomerTypes {
			name := buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language)
			if !slices.Contains(buffets, name) {
				buffets = append(buffets, name)
			}
		}
	}
	return strings.Join(buffets, "+")
}

// SetProductFields 动态更新商品多个字段值
func (model *SaleBill) SetProductFields(saleOrderProductUuid uint64, updateProduct SaleOrderProduct, specialFields ...map[string]bool) {
	for i, order := range model.SaleOrders {
		for j, product := range order.SaleOrderProducts {
			if product.Uuid == saleOrderProductUuid {
				product := model.SaleOrders[i].SaleOrderProducts[j]
				var fields map[string]bool
				if len(specialFields) > 0 {
					fields = specialFields[0]
				}
				product.SetFields(updateProduct, fields)
				product.SetUpdate()
				break
			}
		}
	}
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

// SetAllDiscountCancel 设置整单折扣取消
func (model *SaleBill) SetAllDiscountCancel() bool {
	isChange := false
	for _, saleOrder := range model.SaleOrders {
		isChange = saleOrder.SetAllDiscountCancel() || isChange
	}
	return isChange
}

type Sauce struct {
	Name           string
	Price          float64
	ProductBomUuid uint64
}
type Flavor struct {
	Name           string
	Price          float64
	ProductBomUuid uint64
}
type Attribute struct {
	Name                 string
	ProductAttributeUuid uint64
}

// SaleOrderProductAttribute 销售订单产品属性 `ttpos_sale_order_product_attribute`
type SaleOrderProductAttribute struct {
	BaseModel
	Name                 string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品属性名称,不随后台更新'"`
	SaleOrderUuid        uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductAttributeUuid uint64 `gorm:"column:product_attribute_uuid;not null;default:0;comment:'商品属性ID'"`

	ProductAttribute ProductAttribute `gorm:"foreignKey:ProductAttributeUuid;references:uuid"`
}

func (model *SaleOrderProductAttribute) SetNil() {
	model.ProductAttribute = ProductAttribute{}
}

func (model *SaleOrderProductAttribute) CopyAttribute(saleOrderUuid uint64, saleOrderProductUuid uint64) *SaleOrderProductAttribute {
	newAttribute := &SaleOrderProductAttribute{}
	copier.Copy(newAttribute, model)
	// 设置销售订单信息
	newAttribute.SaleOrderUuid = saleOrderUuid
	newAttribute.SaleOrderProductUuid = saleOrderProductUuid
	// 设置uuid
	uuid, _ := utils.GetID()
	newAttribute.BaseModel = BaseModel{
		Uuid: uuid,
	}
	newAttribute.SetNil()
	return newAttribute
}

// SaleOrderProductBom 销售订单产品原料 `ttpos_sale_order_product_bom`
type SaleOrderProductBom struct {
	BaseModel
	Name                 string           `gorm:"column:name;type:varchar(255);not null;default:'';comment:'规格或小料规格名称,不随后台更新'"`
	Price                float64          `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:'单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动'"`
	IsFlavorBom          uint             `gorm:"column:is_flavor_bom;type:tinyint(1);not null;default:0;comment:'是否为规格商品BOM, 0-否,加料商品 1-是,规格商品'"`
	SaleOrderUuid        uint64           `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64           `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductBomUuid       uint64           `gorm:"column:product_bom_uuid;not null;default:0;comment:'商品BOM ID'"`
	ProductBom           ProductBom       `gorm:"foreignKey:product_bom_uuid;references:uuid"`
	SaleOrderProduct     SaleOrderProduct `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
}

func (model *SaleOrderProductBom) SetNil() {
	model.ProductBom = ProductBom{}
	model.SaleOrderProduct = SaleOrderProduct{}
}

func (model *SaleOrderProductBom) CopyBom(saleOrderUuid uint64, saleOrderProductUuid uint64) *SaleOrderProductBom {
	newBom := &SaleOrderProductBom{}
	copier.Copy(newBom, model)
	// 设置销售订单信息
	newBom.SaleOrderUuid = saleOrderUuid
	newBom.SaleOrderProductUuid = saleOrderProductUuid
	// 设置uuid
	uuid, _ := utils.GetID()
	newBom.BaseModel = BaseModel{
		Uuid: uuid,
	}
	// 将对象置空，否则这些对象会产生新的insert sql语句
	newBom.SetNil()
	return newBom
}

func (model *SaleOrderProductBom) IsFlavor() bool {
	return model.IsFlavorBom == 1
}

// SaleBillSetting 销售账单设置 ttpos_sale_bill_setting
type SaleBillSetting struct {
	// 基础字段
	BaseModel
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`

	// 费用计算设置
	ServiceFeeType  uint    `gorm:"column:service_fee_type;type:tinyint(1);default:0;comment:服务费类型, 0-免服务费 1-按固定金额 2-按比例-不收取税费 3-按比例-收取税费" json:"service_fee_type"`
	ServiceFeeValue float64 `gorm:"column:service_fee_value;type:decimal(12,2);default:0;comment:服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例" json:"service_fee_value"`
	TaxFeeType      uint    `gorm:"column:tax_fee_type;type:tinyint(1);default:0;comment:税费类型, 0-关闭消费税 1-商品未含税 2-商品已含税" json:"tax_fee_type"`

	// 优惠和抹零设置
	DiscountType     uint `gorm:"column:discount_type;type:tinyint(1);default:0;comment:打折类型, 0-百分比打折% 1-百分比直接减免% off" json:"discount_type"`
	ZeroRule         uint `gorm:"column:zero_rule;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero_rule"`
	ZeroCheckoutRule uint `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元" json:"zero_checkout_rule"`

	// 统计设置
	IsStatGift uint `gorm:"column:is_stat_gift;type:tinyint(1);default:0;comment:是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣" json:"is_stat_gift"`
	IsStatFree uint `gorm:"column:is_stat_free;type:tinyint(1);default:0;comment:是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费" json:"is_stat_free"`
}

// 获取销售账单商品税费类型
func (model *SaleBillSetting) GetTaxFeeType() int {
	switch model.TaxFeeType {
	// 关闭税费、不收取税费
	case constant.TaxFeeTypeNone:
		return constant.TaxFeeTypeNone
	//	商品已含税
	case constant.TaxFeeTypeTax:
		return constant.TaxFeeTypeTax
	// 商品未含税
	case constant.TaxFeeTypeNoTax:
		return constant.TaxFeeTypeNoTax
	// 默认为关闭税费
	default:
		return constant.TaxFeeTypeNone
	}
}

// 获取销售账单商品税费类型
func (model *SaleBillSetting) GetServiceFeeType() int {
	switch model.ServiceFeeType {
	// 不收取服务费
	case constant.SaleBillSettingServiceFeeTypeNone:
		return constant.SaleBillSettingServiceFeeTypeNone
	//	固定服务费
	case constant.SaleBillSettingServiceFeeTypeFixed:
		return constant.SaleBillSettingServiceFeeTypeFixed
	// 按比例-不收取税费
	case constant.SaleBillSettingServiceFeeTypePercent:
		return constant.SaleBillSettingServiceFeeTypePercent
	// 按比例-收取税费
	case constant.SaleBillSettingServiceFeeTypePercentTax:
		return constant.SaleBillSettingServiceFeeTypePercentTax
	// 默认为不收取服务费
	default:
		return constant.TaxFeeTypeNone
	}
}

// 获取该销售账单的服务费比例。
// 当服务费关闭时，服务费比例为0，即不收取销售订单商品的服务费
// 当服务费按固定金额收取时，服务费比例为0，即不收取销售订单商品的服务费，只在销售订单加上固定金额的服务费
// 当服务费按比例收取时，服务费比例为ServiceFeeValue字段记录的值
func (model *SaleBillSetting) GetServiceFeeRate() float64 {
	switch model.ServiceFeeType {
	// 不收取服务费时，服务费比率为0
	case constant.SaleBillSettingServiceFeeTypeNone:
		return 0
	// 收固定服务费时，服务费比率为0
	case constant.SaleBillSettingServiceFeeTypeFixed:
		return 0
	//	按比例收时，服务费比率为ServiceFeeValue
	case constant.SaleBillSettingServiceFeeTypePercent:
		return model.ServiceFeeValue
	//	按比例收时，服务费比率为ServiceFeeValue
	case constant.SaleBillSettingServiceFeeTypePercentTax:
		return model.ServiceFeeValue
	//	服务费比率为0
	default:
		return 0

	}
}

// SaleOrderBuffetDelayProduct 销售订单加钟价格商品表 `ttpos_sale_order_buffet_delay_product`
type SaleOrderBuffetDelayProduct struct {
	BaseModel

	// 数值字段
	Name      string  `gorm:"default:'';column:name;comment:'自助餐加钟商品名称，下单时固定不受后台改变'"`
	Num       uint    `gorm:"default:0;column:num;comment:'数量,默认是等于桌台人数。但拆单移动后等于移动的数量；调整自助餐人数后，同一个加钟商品（不管此时被拆分为多少个单）的数量等于桌台人数'"`
	Sign      string  `gorm:"default:'';column:sign;comment:'加钟商品签名。生成uuid,用于标识不同拆单中的加钟商品是不是同一次加购的。在同一个子单中相同签名的加钟商品要合并'"`
	Price     float64 `gorm:"column:price;type:decimal(12,2);default:0;comment:'价格（单价）,下单时固定不受后台改变，结账时再检查是否改变'"`
	DelayTime int64   `gorm:"column:delay_time;type:decimal(12,2);default:0;comment:'加钟时间'"`

	// 关联ID字段
	SaleOrderUuid   uint64 `gorm:"default:0;column:sale_order_uuid;comment:'销售订单ID'"`
	BuffetDelayUuid uint64 `gorm:"default:0;column:buffet_delay_uuid;comment:'自助餐加钟价格ID'"`
}

func (model *SaleOrderBuffetDelayProduct) SetNil() {

}

func (model *SaleOrderBuffetDelayProduct) GetSign() string {
	return model.Sign // 创建加钟商品的时候就生产的uuid，保证绝对唯一。不管该商品移动到哪个子单，只要该uuid一样就是同一个加钟商品
}

func (model *SaleOrderBuffetDelayProduct) CopyBuffetDelayProduct(saleOrderUuid uint64) *SaleOrderBuffetDelayProduct {
	delayProduct := SaleOrderBuffetDelayProduct{}
	err := copier.Copy(&delayProduct, model)
	if err != nil {
		return nil
	}
	// 重置base_model
	delayProduct.BaseModel = BaseModel{}
	delayProduct.SetNil()
	// 指定目标销售订单。如果不移动到别的销售订单可以修改销售订单uuid
	if saleOrderUuid != 0 {
		delayProduct.SaleOrderUuid = saleOrderUuid
	}
	return &delayProduct
}

// 获取商品的价格。单价*数量
func (model *SaleOrderBuffetDelayProduct) GetAmount() float64 {
	amount := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromInt(int64(model.Num))).Round(2).InexactFloat64()
	return amount
}

// 销售订单优惠策略表
type SaleOrderDiscountStrategy struct {
	BaseModel
	// 基本信息
	Type      uint    `gorm:"column:type;type:tinyint(2);not null;default:0;comment:优惠策略类型,0-整单折扣、1-会员折扣" json:"type"`
	Name      string  `gorm:"column:name;type:varchar(50);not null;default:'';comment:优惠策略名称" json:"name"`
	Value     float64 `gorm:"column:value;type:decimal(12,2);not null;default:0.00;comment:优惠策略值" json:"value"`
	JsonField string  `gorm:"column:json_field;type:text;default:null;comment:JSON字段" json:"json_field"`
	// 关联ID字段
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
}

// SaleOrderProductReason 销售订单产品各种原因 `ttpos_sale_order_product_reason`
type SaleOrderProductReason struct {
	// 基础字段
	BaseModel
	// 关联ID字段
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	SaleOrderProductUuid  uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单商品ID.如果说退菜和赠菜，则sale_order_product_uuid不为0；如果是整单免单，则sale_order_product_uuid为0	" json:"sale_order_product_uuid"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;not null;default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	// 三选一。
	ReturnFoodReasonUuid uint64 `gorm:"column:return_food_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:退菜原因ID" json:"return_food_reason_uuid"`
	FreeReasonUuid       uint64 `gorm:"column:free_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:免单原因ID" json:"free_reason_uuid"`
	GiftReasonUuid       uint64 `gorm:"column:gift_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:赠菜原因ID" json:"gift_reason_uuid"`

	// 关联对象
	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

func (model *SaleOrderProductReason) SetNil() {
	model.MultiLanguageName = nil
}
