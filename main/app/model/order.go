package model

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
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
	BillType       uint `gorm:"column:bill_type;type:tinyint(1);default:0;comment:账单类型, 0-桌台订单、1-点餐订单" json:"bill_type"`
	DiningMethod   uint `gorm:"column:dining_method;type:tinyint(1);default:0;comment:用餐方式,0-堂食 1-打包" json:"dining_method"`
	IsBuffet       uint `gorm:"column:is_buffet;type:tinyint(1);default:0;comment:是否自助餐, 0-否 1-是" json:"is_buffet"`
	BuffetDuration uint `gorm:"column:buffet_duration;type:int(10);default:0;comment:自助餐可用时长（秒），0为不限时. 原始值为自助餐的时长，加钟时会累加" json:"buffet_duration"`

	// 订单基本信息
	MealNum uint   `gorm:"column:meal_num;type:int(10);default:0;comment:就餐人数" json:"meal_num"`
	Remark  string `gorm:"column:remark;type:varchar(255);default:'';comment:备注(开台备注)" json:"remark"`
	Reason  string `gorm:"column:reason;type:varchar(255);default:'';comment:原因" json:"reason"`

	// 金额字段 - 主要金额
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:订单总金额,关联销售订单的总金额之和" json:"amount"`
	ProductAmount float64 `gorm:"column:product_amount;type:decimal(12,2);default:0;comment:商品金额,关联销售订单的商品金额之和" json:"product_amount"`

	// 金额字段 - 支付相关
	PaymentAmount        float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0;comment:支付金额,支付金额-订单总金额=支付手续费" json:"payment_amount"`
	PaymentCommissionFee float64 `gorm:"column:payment_commission_fee;type:decimal(12,2);default:0;comment:支付手续费,多次支付的支付手续费之和" json:"payment_commission_fee"`

	// 金额字段 - 费用相关
	ServiceFee float64 `gorm:"column:service_fee;type:decimal(12,2);default:0;comment:服务费,关联销售订单的服务费之和" json:"service_fee"`
	TaxFee     float64 `gorm:"column:tax_fee;type:decimal(12,2);default:0;comment:税费,关联销售订单的税费之和" json:"tax_fee"`

	// 金额字段 - 优惠相关
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);default:0;comment:折扣费用,关联销售订单的折扣费用之和" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣费用,关联销售订单的会员折扣费用之和" json:"member_discount_fee"`
	GiftAmount        float64 `gorm:"column:gift_amount;type:decimal(12,2);default:0;comment:赠菜金额,关联销售订单的赠菜金额之和" json:"gift_amount"`
	FreeAmount        float64 `gorm:"column:free_amount;type:decimal(12,2);default:0;comment:免单金额,关联销售订单的免单金额之和" json:"free_amount"`

	// 时间相关字段
	FinishTime   int64 `gorm:"column:finish_time;type:int(10);default:0;comment:完成时间（时间戳）" json:"finish_time"`
	HideBillTime int64 `gorm:"column:hide_bill_time;type:int(10);default:0;comment:隐藏账单时间（时间戳）" json:"hide_bill_time"`

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
	BuffetPackage1  BuffetPackage     `gorm:"foreignKey:BuffetPackage1Uuid;references:uuid"`
	BuffetPackage2  BuffetPackage     `gorm:"foreignKey:BuffetPackage2Uuid;references:uuid"`
}

func (model *SaleBill) SetNil() {
	model.SaleOrders = nil
	model.H5OrderProducts = nil
	model.SaleBillSetting = nil
	model.Cashier = Staff{}
	model.Desk = nil
	model.BuffetPackage1 = BuffetPackage{}
	model.BuffetPackage2 = BuffetPackage{}
}

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

type SaleBillCalc struct {
	Amount            float64 `json:"amount"`              // 订单总金额=销售订单的应收金额之和
	ProductAmount     float64 `json:"product_amount"`      // 商品金额=销售订单的商品金额之和
	ServiceFee        float64 `json:"service_fee"`         // 服务费=销售订单的服务费之和
	TaxFee            float64 `json:"tax_fee"`             // 税费=销售订单的税费之和
	DiscountFee       float64 `json:"discount_fee"`        // 折扣费用=销售订单的折扣费用之和
	MemberDiscountFee float64 `json:"member_discount_fee"` // 会员折扣费用=销售订单的会员折扣费用之和
	GiftAmount        float64 `json:"gift_amount"`         // 赠菜金额=销售订单的赠菜金额之和
	FreeAmount        float64 `json:"free_amount"`         // 免单金额=销售订单的免单金额之和
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
		for _, product := range saleOrder.SaleOrderProducts {
			if product.IsDelete() || product.IsCancelProduct() {
				continue
			}
			if product.IsGiftProduct() {
				price := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromUint64(uint64(product.Num)))
				amount = amount.Add(price)
			}
		}
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
			beforeCalc := saleOrderProduct.BeforeCalc()
			afterCalc := saleOrderProduct.CalcSaleOrderProduct(*setting)
			fmt.Println(fmt.Sprintf("beforeCalc %+v", beforeCalc))
			fmt.Println(fmt.Sprintf("afterCalc %+v", afterCalc))
		}
		saleOrder.CalcSaleOrder(*setting)
	}
	model.CalcSaleBill()
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
	model.DiscountFee = calc.DiscountFee
	calc.MemberDiscountFee = model.calcMemberDiscountFee()
	model.MemberDiscountFee = calc.MemberDiscountFee
	calc.GiftAmount = model.calcGiftAmount()
	model.GiftAmount = calc.GiftAmount
	calc.FreeAmount = model.calcFreeAmount()
	model.FreeAmount = calc.FreeAmount
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
	for _, saleOrder := range model.SaleOrders {
		if saleOrderUuid == saleOrder.Uuid {
			return saleOrder
		}
	}
	return nil
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
	name1 := model.BuffetPackage1.MultiLanguageName.GetNames()
	name2 := model.BuffetPackage2.MultiLanguageName.GetNames()
	if model.BuffetPackage1.Uuid != 0 && model.BuffetPackage2.Uuid != 0 {
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
	if model.BuffetPackage1.Uuid != 0 {
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
	return model.IsBuffet == 1 // 桌台账单肯定是有桌台ID
}

// 获取自助餐结束时间
func (model *SaleBill) BuffetEndTime() int64 {
	endTime := model.BaseModel.CreateTime + int64(model.BuffetDuration)
	return endTime
}

// 自助餐还剩余多少秒。可以为负数，表示自助餐已经结束了多少秒
func (model *SaleBill) BuffetRemainingSeconds() int64 {
	if model.BuffetDuration == 0 {
		return -1
	}
	remainingTime := model.BuffetEndTime() - time.Now().Unix()
	if remainingTime < 0 {
		return 0
	}
	return remainingTime
}

// ValidateOrderStatus 判断订单是否可操作
func (model *SaleBill) ValidateOrderStatus(operation string, saleOrderUuid ...uint64) error {
	if operation != constant.OrderSettle && model.IsLock == 1 {
		return errors.New("订单已被锁定，请解锁后重新操作")
	}
	if model.Status == constant.SaleBillStatusCanceled {
		return errors.New("订单已取消")
	}
	if model.Status == constant.SaleBillStatusComplete {
		return errors.New("订单已结账")
	}
	if len(model.SaleOrders) > 0 {
		// todo: 要判断来源 拆单没有取消权限
		// if operation == constant.OrderOrderCancel && len(model.SaleOrders) > 1 {
		// 	return errors.New("拆单不可操作")
		// }
		// 单个订单不能操作
		for _, so := range model.SaleOrders {
			if len(saleOrderUuid) == 0 || slices.Contains(saleOrderUuid, so.Uuid) {
				if err := so.ValidateOrderStatus(); err != nil {
					return err
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

// SaleOrderBuffetCustomerType 销售订单自助餐顾客类型
type SaleOrderBuffetCustomerType struct {
	// 主键字段
	BaseModel

	Name string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'名称'"`
	// 关联ID字段
	SaleOrderUuid               uint64 `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	BuffetPackageUuid           uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
	BuffetCustomerTypePriceUuid uint64 `gorm:"column:buffet_customer_type_price_uuid;comment:顾客类型定价ID" json:"buffet_customer_type_price_uuid"`

	// 数值字段
	Num                uint    `gorm:"column:num;type:int(11);default:0;comment:人数" json:"num"`
	CustomerPrice      float64 `gorm:"column:customer_price;type:decimal(12,2);not null;default:0;comment:原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变" json:"customer_price"`
	Price              float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:价格（折后价），只进行自定义打折，不进行会员打折" json:"price"`
	CustomDiscountRate float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:1;comment:自定义折扣率(0-100%)" json:"custom_discount_rate"`
	CustomDiscountFee  float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0;comment:自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*(1-自定义折扣率)" json:"custom_discount_fee"`
	TaxRate            float64 `gorm:"column:tax_rate;type:decimal(10,2);not null;default:0;comment:税率,单位%.加购时记录税率,结账时再重新核算" json:"tax_rate"`
	ServiceTaxFee      float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0;comment:服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率" json:"service_tax_fee"`
	TaxFee             float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0;comment:自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率" json:"tax_fee"`
	ServiceFee         float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0;comment:服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例" json:"service_fee"`
	Amount             float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:应收金额(单人)。自助餐顾客类型已含税时，应收金额(单人)=(自助餐顾客类型原价-自助餐顾客类型税费)+服务费+自助餐顾客类型税费；自助餐顾客类型未含税时，应收金额(单人)=自助餐顾客类型原价+服务费+自助餐顾客类型税费" json:"amount"`

	// 关联字段
	BuffetPackage           BuffetPackage           `gorm:"foreignKey:BuffetPackageUuid;references:uuid"`
	BuffetCustomerTypePrice BuffetCustomerTypePrice `gorm:"foreignKey:BuffetCustomerTypePriceUuid;references:uuid"` // 用于关联后台设置的顾客类型定价。在结账时，判断价格是否改变
}

// 获取顾客原价
func (model *SaleOrderBuffetCustomerType) GetOriginPrice() float64 {
	price := decimal.NewFromFloat(model.CustomerPrice).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// 获取顾客折后价
func (model *SaleOrderBuffetCustomerType) GetDiscountPrice() float64 {
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// SaleOrderBuffetDelayProduct 销售订单加钟价格商品表 `ttpos_sale_order_buffet_delay_product`
type SaleOrderBuffetDelayProduct struct {
	BaseModel

	// 数值字段
	Name string `gorm:"default:'';column:name;comment:'自助餐加钟商品名称，下单时固定不受后台改变'"`
	// 废弃，直接使用桌台人数即可
	//Num   uint    `gorm:"default:0;column:num;comment:'数量'"`
	Price float64 `gorm:"column:price;type:decimal(12,2);default:0;comment:'价格（单价）,下单时固定不受后台改变，结账时再检查是否改变'"`

	// 关联ID字段
	SaleOrderUuid   uint64 `gorm:"default:0;column:sale_order_uuid;comment:'销售订单ID'"`
	BuffetDelayUuid uint64 `gorm:"default:0;column:buffet_delay_uuid;comment:'自助餐加钟价格ID'"`
}

// 获取商品的价格。单价*数量
func (model *SaleOrderBuffetDelayProduct) GetPrice(num uint) float64 {
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromInt(int64(num))).Round(2).InexactFloat64()
	return price
}

// SaleBillOperationRecord 桌台账单操作记录
type SaleBillOperationRecord struct {
	BaseModel
	// 基本信息
	Data   string `gorm:"column:data;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"data"`
	Source string `gorm:"column:source;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"source"`
	Action string `gorm:"column:action;comment:操作行为" json:"action"`
	Remark string `gorm:"column:remark;comment:备注" json:"remark"`
	// 关联ID字段
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作员ID" json:"operator_uuid"`
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
	SaleOrderProductUuid  uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单商品ID" json:"sale_order_product_uuid"`
	ReturnFoodReasonUuid  uint64 `gorm:"column:return_food_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:退菜原因ID" json:"return_food_reason_uuid"`
	FreeReasonUuid        uint64 `gorm:"column:free_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:免单原因ID" json:"free_reason_uuid"`
	GiftReasonUuid        uint64 `gorm:"column:gift_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:赠菜原因ID" json:"gift_reason_uuid"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;not null;default:0;comment:退菜原因-多语言名称ID" json:"multi_language_name_uuid"`

	// 关联对象
	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
