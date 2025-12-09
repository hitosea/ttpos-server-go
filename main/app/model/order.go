package model

import (
	"encoding/json"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

func NewDeskSaleBill(saleBillUuid uint64, orderNo string, buffetUuids []uint64, mealNum uint, remark string, deskUuid uint64, serialNo string, dutyNo string, cashierUuid uint64, cashierName string) *SaleBill {
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
	// 设置收银员信息
	saleBill.SetCashier(dutyNo, cashierUuid, cashierName)

	// 设置自助餐套餐
	if isBuffet {
		saleBill.SetBuffetPackage(buffetUuids)
		saleBill.BuffetStartTime = time.Now().Unix()
	}

	return saleBill
}

type SaleBillCalc struct {
	Amount            float64 `json:"amount"`              // 订单总金额=销售订单的应收金额之和
	OriginAmount      float64 `json:"origin_amount"`       // 订单金额(折前价)。销售订单的原始应收金额之和
	ProductAmount     float64 `json:"product_amount"`      // 商品金额=销售订单的商品金额之和
	ServiceFee        float64 `json:"service_fee"`         // 服务费=销售订单的服务费之和
	TaxFee            float64 `json:"tax_fee"`             // 税费=销售订单的税费之和
	DiscountFee       float64 `json:"discount_fee"`        // 折扣费用=销售订单的折扣费用之和
	MemberDiscountFee float64 `json:"member_discount_fee"` // 会员折扣费用=销售订单的会员折扣费用之和
	GiftAmount        float64 `json:"gift_amount"`         // 赠菜金额=销售订单的赠菜金额之和
	ActivityAmount    float64 `json:"activity_amount"`     // 满减活动抵扣金额=销售订单的活动抵扣金额之和
	FreeAmount        float64 `json:"free_amount"`         // 免单金额=销售订单的免单金额之和

	ProductOriginalAmount float64 `json:"product_original_amount"` //
	PaymentAmount         float64
	PaymentCommissionFee  float64
}

// SaleOrderProductAttribute 销售订单产品属性 `ttpos_sale_order_product_attribute`
type SaleOrderProductAttribute struct {
	BaseModel
	Name                        string `gorm:"column:name;type:text;not null;default:'';comment:'商品属性名称快照（JSON），不随后台更新'"`
	SaleOrderUuid               uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid        uint64 `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductAttributeUuid        uint64 `gorm:"column:product_attribute_uuid;not null;default:0;comment:'商品属性ID'"`
	ProductPackageAttributeUuid uint64 `gorm:"column:product_package_attribute_uuid;not null;default:0;comment:'商品包属性ID'"`

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

// GetLocaleName 获取属性名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProductAttribute) GetLocaleName() dto.LocaleResponse {
	// 优先使用快照字段
	snapshotName := model.Name

	// 如果快照字段不为空，尝试反序列化为多语言数据
	if snapshotName != "" {
		var snapshotLocale dto.LocaleResponse
		if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
			// 反序列化成功，检查是否有主语言数据
			if !snapshotLocale.IsNull() {
				// 如果快照数据完整，直接返回
				return snapshotLocale
			}
		}
		// 如果反序列化失败或数据不完整，继续后续降级逻辑
	}

	// 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
	if !model.ProductAttribute.MultiLanguageName.IsNullName() {
		return model.ProductAttribute.MultiLanguageName.GetNames()
	}

	// 兜底：如果关联表也没有数据，返回空的多语言响应
	return dto.LocaleResponse{}
}

// SetNameSnapshot 设置属性名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProductAttribute) SetNameSnapshot(multiLangName MultiLanguageName) error {
	// 如果多语言名称为空，设置为空字符串
	if multiLangName.IsNullName() {
		model.Name = ""
		return nil
	}

	// 构建 LocaleResponse
	localeResp := multiLangName.GetNames()

	// 序列化为 JSON
	jsonData, err := json.Marshal(localeResp)
	if err != nil {
		return err
	}

	model.Name = string(jsonData)
	return nil
}

// SaleOrderProductBom 销售订单产品原料 `ttpos_sale_order_product_bom`
type SaleOrderProductBom struct {
	BaseModel
	Name        string  `gorm:"column:name;type:text;not null;default:'';comment:'规格或小料名称快照（JSON），不随后台更新'"`
	Price       float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:'单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动'"`
	IsFlavorBom uint    `gorm:"column:is_flavor_bom;type:tinyint(1);not null;default:0;comment:'是否为规格商品BOM, 0-否,加料商品 1-是,规格商品'"`

	// 关联ID字段
	SaleOrderUuid        uint64 `gorm:"column:sale_order_uuid;not null;default:0;comment:'销售订单ID'"`
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;not null;default:0;comment:'销售订单商品ID'"`
	ProductBomUuid       uint64 `gorm:"column:product_bom_uuid;not null;default:0;comment:'商品BOM ID'"`

	// 关联对象
	ProductBom       ProductBom       `gorm:"foreignKey:product_bom_uuid;references:uuid"`
	SaleOrderProduct SaleOrderProduct `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
}

func (model *SaleOrderProductBom) SetPrice(price float64) {
	defer model.SetUpdate() // 设置更新标记
	model.Price = price
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
	return model.IsFlavorBom == constant.ProductBomTypeFlavor
}

func (model *SaleOrderProductBom) IsSauce() bool {
	return model.IsFlavorBom == constant.ProductBomTypeSauce
}

// GetLocaleName 获取规格或小料名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-product-attribute-snapshot-fix
func (model *SaleOrderProductBom) GetLocaleName() dto.LocaleResponse {
	// 优先使用快照字段
	snapshotName := model.Name

	// 如果快照字段不为空，尝试反序列化为多语言数据
	if snapshotName != "" {
		var snapshotLocale dto.LocaleResponse
		if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
			// 反序列化成功，检查是否有主语言数据
			if !snapshotLocale.IsNull() {
				// 如果快照数据完整，直接返回
				return snapshotLocale
			}
		}
		// 如果反序列化失败或数据不完整，继续后续降级逻辑
	}

	// 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
	if model.IsFlavor() {
		// 规格
		if !model.ProductBom.ProductFlavor.MultiLanguageName.IsNullName() {
			return model.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
		}
	} else {
		// 小料
		if !model.ProductBom.ProductSauce.MultiLanguageName.IsNullName() {
			return model.ProductBom.ProductSauce.MultiLanguageName.GetNames()
		}
	}

	// 兜底：如果关联表也没有数据，返回空的多语言响应
	return dto.LocaleResponse{}
}

// SetNameSnapshot 设置规格或小料名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-product-attribute-snapshot-fix (JSON 方案)
func (model *SaleOrderProductBom) SetNameSnapshot(multiLangName MultiLanguageName) error {
	// 如果多语言名称为空，设置为空字符串
	if multiLangName.IsNullName() {
		model.Name = ""
		return nil
	}

	// 构建 LocaleResponse
	localeResp := multiLangName.GetNames()

	// 序列化为 JSON
	jsonData, err := json.Marshal(localeResp)
	if err != nil {
		return err
	}

	model.Name = string(jsonData)
	return nil
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
	ServiceApply    uint    `gorm:"column:service_apply;type:tinyint(1);default:0;comment:是否收取服务费，0-不收取 1-收取。根据后台的服务费应用范围决定" json:"service_apply"`
	ServiceFeeBase  uint    `gorm:"column:service_fee_base;type:tinyint(1);default:0;comment:服务费计算基准, 0-商品惠后价 1-商品价格合计" json:"service_fee_base"`

	// 优惠和抹零设置
	DiscountType     uint `gorm:"column:discount_type;type:tinyint(1);default:0;comment:打折类型, 0-百分比打折% 1-百分比直接减免% off" json:"discount_type"`
	ZeroRule         uint `gorm:"column:zero_rule;type:tinyint(1);default:0;comment:优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数" json:"zero_rule"`
	ZeroCheckoutRule uint `gorm:"column:zero_checkout_rule;type:tinyint(1);default:0;comment:结账抹零, 0-实款实收 1-抹分 2-抹角 5-抹元" json:"zero_checkout_rule"`

	// 统计设置
	IsStatGift uint `gorm:"column:is_stat_gift;type:tinyint(1);default:0;comment:是否统计赠菜金额, 0-不计入总销售额、优惠折扣 1-计入总销售额、优惠折扣" json:"is_stat_gift"`
	IsStatFree uint `gorm:"column:is_stat_free;type:tinyint(1);default:0;comment:是否统计免单金额, 0-不计入总销售额、优惠折扣、服务费、税费 1-计入总销售额、优惠折扣、服务费、税费" json:"is_stat_free"`

	// 积分抵扣设置
	OpenPointsExchange uint    `gorm:"column:open_points_exchange;type:tinyint(1);default:0;comment:是否开启积分抵扣, 0-不开启 1-开启" json:"open_points_exchange"`
	PointsExchangeRate float64 `gorm:"column:points_exchange_rate;type:decimal(12,2);default:0;comment:积分汇率，每积分抵扣的金额，可输入大于0的两位小数" json:"points_exchange_rate"`
	AutoPointsExchange uint    `gorm:"column:auto_points_exchange;type:tinyint(1);default:0;comment:积分抵扣类型,0-手动抵扣 1-自动抵扣" json:"auto_points_exchange"`

	// 分批送厨模式
	BatchCookingMode string `gorm:"column:batch_cooking_mode;type:varchar(10);default:'post';comment:分批送厨模式: pre-前置 / post-后置，默认 post" json:"batch_cooking_mode"`
}

// 积分是否开启自动抵扣
func (model *SaleBillSetting) IsAutoPointsExchange() bool {
	return model.OpenPointsExchange == 1 && model.AutoPointsExchange == 1
}

// 是否开启积分抵扣
func (model *SaleBillSetting) IsOpenPointsExchange() bool {
	return model.OpenPointsExchange == 1
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
	// 如果后台设置该账单不在服务费应用范围内，则不收取服务费
	if model.ServiceApply == 0 {
		return constant.SaleBillSettingServiceFeeTypeNone
	}
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
	// 如果后台设置该账单不在服务费应用范围内，则不收取服务费
	if model.ServiceApply == 0 {
		return 0
	}
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

	// 关联对象
	ReturnOrderProducts []*ReturnOrderProduct `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"` // 退货单商品表，用于获取该加钟商品的退货数量
}

func (model *SaleOrderBuffetDelayProduct) SetNil() {
	model.ReturnOrderProducts = nil
}

// GetCanReturnNum 获取销售订单自助餐顾客类型的可退货数量. 可退货数量=订单自助餐顾客类型数量-已退货数量
func (model *SaleOrderBuffetDelayProduct) GetCanReturnNum() uint {
	amount := decimal.NewFromFloat(0)
	for _, returnOrderProduct := range model.ReturnOrderProducts {
		amount = amount.Add(decimal.NewFromFloat(returnOrderProduct.Num))
	}
	num := float64(model.Num) - amount.InexactFloat64()
	// 如果可退货数量小于0，则返回0
	// 这个判断很有必要，否则会出现可退货数量为负数的情况但uint是无符号的，结果会得到一个很大的数。如2-14=18446744073709551604
	if num < 0 {
		return 0
	}
	return uint(num)
}

// GetReturnNum 获取销售订单自助餐加钟的已退货数量. 已退货数量=订单自助餐加钟数量-可退货数量
func (model *SaleOrderBuffetDelayProduct) GetReturnNum() uint {
	return model.Num - model.GetCanReturnNum()
}

// GetCanReturnPrice 获取销售订单自助餐顾客类型的可退货金额. 可退货金额=订单自助餐顾客类型金额-已退货金额
func (model *SaleOrderBuffetDelayProduct) GetCanReturnPrice() float64 {
	return decimal.NewFromFloat(model.Price).Mul(decimal.NewFromUint64(uint64(model.GetCanReturnNum()))).InexactFloat64()
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

// GetReturnPrice 获取销售订单商品的已退金额。订单商品金额 - 可退货金额
func (model *SaleOrderBuffetDelayProduct) GetReturnPrice() float64 {
	if model.GetReturnNum() == 0 {
		return 0
	}
	returnPrice := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromUint64(uint64(model.GetReturnNum()))).Truncate(3).Round(2).InexactFloat64()
	return returnPrice
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
	// 四选一。
	ReturnFoodReasonUuid uint64 `gorm:"column:return_food_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:退菜原因ID" json:"return_food_reason_uuid"`
	FreeReasonUuid       uint64 `gorm:"column:free_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:免单原因ID" json:"free_reason_uuid"`
	GiftReasonUuid       uint64 `gorm:"column:gift_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:赠菜原因ID" json:"gift_reason_uuid"`
	OrderItemRemarkUuid  uint64 `gorm:"column:order_item_remark_uuid;type:bigint(20) unsigned;not null;default:0;comment:备注预设UUID" json:"order_item_remark_uuid"`

	// 快照字段（JSON 方案）
	// Requirement: story-main-reason-snapshot-fix
	Name string `gorm:"column:name;type:text;default:'';comment:原因名称快照（JSON），不随后台更新" json:"name"`

	// 关联对象
	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// 是否是退菜原因
func (model *SaleOrderProductReason) IsReturnFoodReason() bool {
	return model.ReturnFoodReasonUuid != 0
}

// 是否是免单原因
func (model *SaleOrderProductReason) IsFreeReason() bool {
	return model.FreeReasonUuid != 0
}

// 是否是赠菜原因
func (model *SaleOrderProductReason) IsGiftReason() bool {
	return model.GiftReasonUuid != 0
}

// 是否是备注预设
func (model *SaleOrderProductReason) IsOrderItemRemark() bool {
	return model.OrderItemRemarkUuid != 0
}

func (model *SaleOrderProductReason) SetNil() {
	model.MultiLanguageName = nil
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
	ErpCode        string
}
type Attribute struct {
	Name                        string
	ProductAttributeUuid        uint64
	ProductPackageAttributeUuid uint64
}
