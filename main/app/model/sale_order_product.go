package model

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

// SaleOrderProduct 销售订单产品 `ttpos_sale_order_product`
type SaleOrderProduct struct {
	// 基础字段
	BaseModel

	// 基本信息字段
	Name       string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品名称'" json:"name"`
	FlavorName string  `gorm:"column:flavor_name;type:varchar(255);not null;default:'';comment:'规格名称'" json:"flavor_name"`
	Num        float64 `gorm:"column:num;type:int(11);not null;default:0;comment:'商品数量。不能减为0，当数量为1再减时，标记删除'" json:"num"`
	UnitNum    float64 `gorm:"column:unit_num;type:decimal(12,4);not null;default:0.00;comment:'单位数量，用于套餐子商品'" json:"unit_num"`
	NumType    uint    `gorm:"column:num_type;type:tinyint(1);not null;default:0;comment:'数量类型, 0-整数 1-小数'" json:"num_type"`
	Remark     string  `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'备注，顾客对商品的备注信息'" json:"remark"`
	IsBuffet   uint    `gorm:"column:is_buffet;type:tinyint(1);not null;default:0;comment:'是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0'" json:"is_buffet"`
	DeviceId   string  `gorm:"column:device_id;type:varchar(255);not null;default:'';comment:'设备ID,用于标识订单来源设备.来源h5时，device_id为h5的device_id'" json:"device_id"`
	// 状态相关字段
	Status        uint `gorm:"column:status;type:tinyint(1);not null;default:0;comment:'状态, 0-未送厨 1-已送厨'" json:"status"`
	IsRequire     uint `gorm:"column:is_require;type:tinyint(1);not null;default:0;comment:'是否必点商品 0-否 1-是。用于在前端显示必点图标'" json:"is_require"`
	IsAcceptOrder uint `gorm:"column:is_accept_order;type:tinyint(1);not null;default:0;comment:'是否已接单, 0-否 1-是'" json:"is_accept_order"`

	// 价格相关字段
	FlavorPrice      float64 `gorm:"column:flavor_price;type:decimal(12,2);not null;default:0.00;comment:'规格原价（单商品）,仅某规格商品的原价'" json:"flavor_price"`
	SaucePrice       float64 `gorm:"column:sauce_price;type:decimal(12,2);not null;default:0.00;comment:'小料价（单商品）,所有小料的价格之和'" json:"sauce_price"`
	ProductPrice     float64 `gorm:"column:product_price;type:decimal(12,2);not null;default:0.00;comment:'原始单价（单商品）,规格原价+小料价'" json:"product_price"`
	SalePrice        float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0.00;comment:'销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价'" json:"sale_price"`
	SalePriceNoTax   float64 `gorm:"column:sale_price_no_tax;type:decimal(12,2);not null;default:0.00;comment:'销售价,未含税价格（折前）'" json:"sale_price_no_tax"`
	Price            float64 `gorm:"column:price;type:decimal(12,2);not null;default:0.00;comment:'最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率'" json:"price"`
	TotalPrice       float64 `gorm:"column:total_price;type:decimal(12,2);not null;default:0.00;comment:'应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费'" json:"total_price"`
	OriginTotalPrice float64 `gorm:"column:origin_total_price;type:decimal(12,2);not null;default:0.00;comment:'应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费'" json:"origin_total_price"`

	// 折扣相关字段
	ChangePriceTime         int64   `gorm:"column:change_price_time;type:int(10);not null;default:0;comment:'改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并'" json:"change_price_time"`
	OpenMemberDiscount      uint    `gorm:"column:open_member_discount;type:tinyint(1);not null;default:0;comment:'是否开启会员折扣, 0-否 1-是'" json:"open_member_discount"` // 快照设置相关，不受后台改变，结账时检查
	OpenOverallDiscount     uint    `gorm:"column:open_overall_discount;type:tinyint(1);not null;default:0;comment:'是否开启 Overall 折扣, 0-否 1-是'" json:"open_overall_discount"`
	MemberDiscountRate      float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员折扣率(0-100%)'" json:"member_discount_rate"`               // 与sale_order的member_discount_rate一致
	MemberCardDiscountRate  float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员卡折扣率(0-100%)'" json:"member_card_discount_rate"`    // 与sale_order的member_card_discount_rate一致
	MemberOrderDiscountRate float64 `gorm:"column:member_order_discount_rate;type:decimal(12,2);not null;default:1.00;comment:'会员订单折扣率(1-300%)'" json:"member_order_discount_rate"` // 用于上浮会员端上的商品价格
	CustomDiscountRate      float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣率(0-100%)'" json:"custom_discount_rate"`              // 与sale_order的custom_discount_rate一致

	// 折扣金额字段
	DiscountFee       float64 `gorm:"column:discount_fee;type:decimal(12,2);not null;default:0.00;comment:'打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额'" json:"discount_fee"`
	MemberDiscountFee float64 `gorm:"column:member_discount_fee;type:decimal(12,2);default:0;comment:会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）;NOT NULL" json:"member_discount_fee"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣金额（单商品）=销售价-最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)'" json:"custom_discount_fee"`

	// 税费和服务费字段
	TaxRate       float64 `gorm:"column:tax_rate;type:decimal(12,2);not null;default:0;comment:'税率,单位%.加购时记录税率,结账时再重新核算'" json:"tax_rate"`
	ServiceTaxFee float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率'" json:"service_tax_fee"`
	TaxFee        float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0.00;comment:'商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率'" json:"tax_fee"`
	ServiceFee    float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0.00;comment:'服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例'" json:"service_fee"`

	// 库存相关字段
	DeductStockType uint  `gorm:"column:deduct_stock_type;type:tinyint(1);not null;default:0;comment:'库存计算方式,0-付款减库存 1-下单减库存。加购商品时记录，不受后台影响，用于减少查询次数'" json:"deduct_stock_type"`
	DeductStockTime int64 `gorm:"column:deduct_stock_time;type:int(10);not null;default:0;comment:'减库存的时间(时间戳），0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存'" json:"deduct_stock_time"`

	// 赠品相关字段
	GiftTime     int64  `gorm:"column:gift_time;type:int(10);not null;default:0;comment:'赠菜时间(时间戳),用于判断不同时间赠送的商品不合并'" json:"gift_time"`
	WrapTime     int64  `gorm:"column:wrap_time;type:int(10);not null;default:0;comment:'打包时间(时间戳),用于判断不同时间打包的商品不合并'" json:"wrap_time"`
	CancelTime   int64  `gorm:"column:cancel_time;type:int(10);not null;default:0;comment:'退菜时间(时间戳)'" json:"cancel_time"`
	GiftReason   string `gorm:"column:gift_reason;type:varchar(255);not null;default:'';comment:'赠菜原因'" json:"gift_reason"`
	CancelReason string `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:'退菜原因'" json:"cancel_reason"`

	// 关联ID字段
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;not null;default:0;comment:'多语言名称UUID'" json:"multi_language_name_uuid"`
	ImageFileUuid         uint64 `gorm:"column:image_file_uuid;not null;default:0;comment:'商品图片ID'" json:"image_file_uuid"`
	ProductionOrderUuid   uint64 `gorm:"column:production_order_uuid;type:bigint(20);not null;default:0;comment:'生产订单ID'" json:"production_order_uuid"`
	ProductPackageUuid    uint64 `gorm:"column:product_package_uuid;type:bigint(20);not null;default:0;comment:'商品包ID'" json:"product_package_uuid"`
	SaleBillUuid          uint64 `gorm:"column:sale_bill_uuid;type:bigint(20);not null;default:0;comment:'销售账单ID'" json:"sale_bill_uuid"`
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20);not null;default:0;comment:'销售订单ID'" json:"sale_order_uuid"`
	MustPlanUuid          uint64 `gorm:"column:must_plan_uuid;type:bigint(20);not null;default:0;comment:'必点方案ID,产品要求用这种方式标注各个必点'" json:"must_plan_uuid"`
	DeskUuid              uint64 `gorm:"column:desk_uuid;type:bigint(20);not null;default:0;comment:'桌台ID, 默认为0是本台，大于0为合并过来的桌台'" json:"desk_uuid"`

	// 其他字段
	Sign string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:'商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品'" json:"sign"`

	// 扫码订单相关
	H5OrderProductUuid uint64 `gorm:"column:h5_order_product_uuid;type:bigint(20) unsigned;default:0;comment:h5订单商品ID，用于关联h5订单商品，用于判断是否为h5订单商品;NOT NULL" json:"h5_order_product_uuid"`
	H5OrderUuid        uint64 `gorm:"column:h5_order_uuid;type:bigint(20) unsigned;default:0;comment:扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品;NOT NULL" json:"h5_order_uuid"`

	// 套餐相关
	PackageUuid             uint64 `gorm:"column:package_uuid;type:bigint(20);not null;default:0;comment:'套餐uuid'" json:"package_uuid"`                                            // 只有套餐子商品才会有这个字段
	PackageGroupUuid        uint64 `gorm:"column:package_group_uuid;type:bigint(20);not null;default:0;comment:'套餐分组uuid';index:idx_package_group_uuid" json:"package_group_uuid"` // 只有套餐子商品才会有这个字段
	ProductType             uint8  `gorm:"column:product_type;type:tinyint(1);not null;default:0;comment:'商品类型, 0-商品 1-套餐 2-套餐子商品'" json:"product_type"`
	PackageSubProductParams string `gorm:"column:package_sub_product_params;type:text;comment:'套餐子商品参数'" json:"package_sub_product_params"`

	// 送厨时间
	SendKitchenTime int64 `gorm:"column:send_kitchen_time;type:int(10);not null;default:0;comment:'送厨时间'" json:"send_kitchen_time"`

	// 关联对象
	MultiLanguageName          *MultiLanguageName           `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	ImageFile                  *File                        `gorm:"foreignKey:image_file_uuid;references:uuid"`
	SaleOrderProductBoms       []*SaleOrderProductBom       `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
	SaleOrderProductAttributes []*SaleOrderProductAttribute `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ReturnOrderProducts        []*ReturnOrderProduct        `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ProductPackage             *ProductPackage              `gorm:"foreignKey:ProductPackageUuid;references:Uuid"`
	SaleBill                   *SaleBill                    `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	CancelReasons              []*SaleOrderProductReason    `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ProductionOrderProduct     *ProductionOrderProduct      `gorm:"foreignKey:SaleOrderProductUuid;references:uuid"`
	H5Order                    *H5Order                     `gorm:"foreignKey:H5OrderUuid;references:uuid"`
	ProductMustPlan            *ProductMustPlan             `gorm:"foreignKey:MustPlanUuid;references:uuid"`

	// 内部字段
	operation        string `gorm:"-"` // 操作类型。add: 加购，sub: 减购
	unOrderH5Product bool   `gorm:"-"` // 是否为未下单的h5订单商品。 特别标记该商品为正在下单的h5订单商品
}

// 更换规格，重新计算商品的价格
func (model *SaleOrderProduct) ChangeFlavor(flavor *SaleOrderProductBom) {
	model.FlavorPrice = flavor.Price
	model.FlavorName = flavor.Name
	model.SetUpdate() // 标记该记录要更新
}

// 删除所有规格、加料和属性
func (model *SaleOrderProduct) DeleteAllSaleOrderProductBomsAndAttributes() {
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		saleOrderProductBom.SetDelete()
		saleOrderProductBom.SetUpdate() // 标记该记录要更新
	}
	for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
		saleOrderProductAttribute.SetDelete()
		saleOrderProductAttribute.SetUpdate() // 标记该记录要更新
	}
	model.SetUpdate() // 标记该记录要更新
}

// 获取商品的简要，如 牛排*1（标准，黑椒汁）
func (model *SaleOrderProduct) GetProductNameAttributes(language string) string {
	name := model.MultiLanguageName.GetNameByLang(language)
	flarvorSaleOrderProductBom := model.GetFlarvorSaleOrderProductBom()
	flavorName := flarvorSaleOrderProductBom.ProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(language)
	attributes := make([]string, 0)
	for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
		attributes = append(attributes, saleOrderProductAttribute.ProductAttribute.MultiLanguageName.GetNameByLang(language))
	}
	num := decimal.NewFromFloat(model.Num).Round(3).InexactFloat64() // 去掉末尾的0
	message := fmt.Sprintf("%s*%v（%s）", name, num, flavorName)
	if len(attributes) > 0 {
		message = fmt.Sprintf("%s*%v（%s，%s）", name, num, flavorName, strings.Join(attributes, ","))
	}
	return message
}

func (model *SaleOrderProduct) GetFlarvorSaleOrderProductBom() *SaleOrderProductBom {
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if saleOrderProductBom.IsFlavor() {
			return saleOrderProductBom
		}
	}
	return &SaleOrderProductBom{}
}

// 设置商品为赠菜
func (model *SaleOrderProduct) SetGiftProduct(giftReason string) {
	model.GiftTime = time.Now().Unix()
	model.SetUpdate()
	model.GiftReason = giftReason
	if model.IsPackageProduct() {
		model.Sign = model.GeneratePackageSign()
	} else {
		model.Sign = model.GenerateProductSign()
	}
}

// 是否为套餐子商品
func (model *SaleOrderProduct) IsPackageSubProduct() bool {
	return model.ProductType == constant.ProductTypePackageSubProduct
}

// 是否为套餐商品
func (model *SaleOrderProduct) IsPackageProduct() bool {
	return model.ProductType == constant.ProductTypePackage
}

// 获取商品的就餐类型。打包或堂食
func (model *SaleOrderProduct) GetDiningMethod() uint {
	if model.IsWrapProduct() {
		return constant.SaleBillDiningMethodTakeout
	}
	return constant.SaleBillDiningMethodDineIn
}

// 获取会员端商品价格上浮比例
func (model *SaleOrderProduct) GetMemberOrderDiscountRate() float64 {
	// 如果会员端商品价格上浮比例小于等于0，则返回1,表示未设置上浮比例
	if model.MemberOrderDiscountRate <= 0 {
		return 1
	}
	return model.MemberOrderDiscountRate
}

// GetOpenOverallDiscount 获取是否开启 Overall 折扣
func (model *SaleOrderProduct) GetOpenOverallDiscount() bool {
	return model.OpenOverallDiscount == 1
}

// GetCustomDiscountRate 获取自定义折扣率
func (model *SaleOrderProduct) GetCustomDiscountRate() float64 {
	if !model.GetOpenOverallDiscount() {
		// 不开启整单打折
		return constant.NoDiscount
	}
	return model.CustomDiscountRate
}

// 获取商品包的规格uuid
func (model *SaleOrderProduct) GetFlavorBomUuid() uint64 {
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavor() {
			return bom.ProductBomUuid
		}
	}
	return 0
}

// 获取销售订单商品的会员卡折扣率
func (model *SaleOrderProduct) GetMemberCardDiscountRate() float64 {
	// 如果商品不参与会员打折，则返回1,表示不打折
	if model.OpenMemberDiscount == constant.ProductMemberDiscountOff {
		return constant.NoDiscount
	}
	// 现在不能折扣为0
	if model.MemberCardDiscountRate <= 0 {
		return constant.NoDiscount
	}
	return model.MemberCardDiscountRate
}

// 获取销售订单商品的会员折扣率
func (model *SaleOrderProduct) GetMemberDiscountRate() float64 {
	// 如果商品不参与会员打折，则返回1,表示不打折
	if model.OpenMemberDiscount == constant.ProductMemberDiscountOff {
		return constant.NoDiscount
	}
	// 现在不能折扣为0
	if model.MemberDiscountRate <= 0 {
		return constant.NoDiscount
	}
	return model.MemberDiscountRate
}

// 获取商品在会员端显示的价格（折前）单个商品的价格
// 会员端显示的价格= 商品规格价*会员端折扣率 + 小料A*会员端折扣率 + 小料B*会员端折扣率 + ...
// （商品规格价=商品规格原价*会员端折扣率）=》 转换为商品未含税价格。  商品未含税价格*税率=税费 。 （商品未含税价格+税费）
// 如果规格价已含税，计算
// func (model *SaleOrderProduct) GetPriceInMemberClient() float64 {
// 	saucePrices := make([]float64, 0) // 该销售订单商品的各个小料原始价格
// 	flavorPrice := model.FlavorPrice  // 该销售订单商品的规格原始价格
// 	for _, bom := range model.SaleOrderProductBoms {
// 		if bom.IsDelete() {
// 			continue
// 		}
// 		if bom.ProductBom.IsSauce() {
// 			saucePrices = append(saucePrices, bom.ProductBom.Price)
// 		}
// 	}
// 	// 上浮后的价格
// 	flavorPrice = decimal.NewFromFloat(flavorPrice).Mul(decimal.NewFromFloat(model.GetMemberOrderDiscountRate())).Round(2).InexactFloat64()
// 	// 上浮后的小料价格
// 	for i, saucePrice := range saucePrices {
// 		saucePrices[i] = decimal.NewFromFloat(saucePrice).Mul(decimal.NewFromFloat(model.GetMemberOrderDiscountRate())).Round(2).InexactFloat64()
// 	}
// 	// 上浮后的小料价格总和
// 	var saucePriceTotal float64
// 	for _, saucePrice := range saucePrices {
// 		saucePriceTotal += saucePrice
// 	}
// 	return decimal.NewFromFloat(flavorPrice).Add(decimal.NewFromFloat(saucePriceTotal)).Round(2).InexactFloat64()
// }

// // 获取商品在会员端显示的价格（折前）
// func (model *SaleOrderProduct) GetTotalPriceInMemberClient() float64 {
// 	return decimal.NewFromFloat(model.GetPriceInMemberClient()).Mul(model.GetNumDecimal()).Round(2).InexactFloat64()
// }

// 获取销售订单商品的总税费。包含服务费税费
func (model *SaleOrderProduct) GetTotalTaxFee() float64 {
	return model.GetTaxFee() + model.GetServiceTaxFee()
}

// 获取销售订单商品的税费。税费=销售订单商品的税费*销售订单商品的数量
func (model *SaleOrderProduct) GetTaxFee() float64 {
	return decimal.NewFromFloat(model.TaxFee).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
}

// 获取销售订单商品的原始税费(折前价)。税费=销售订单商品的税费*销售订单商品的数量
func (model *SaleOrderProduct) GetOriginTaxFee(taxFeeType int) float64 {
	taxFee := model.calcOriginTaxFee(taxFeeType) // 税费（折前）
	return decimal.NewFromFloat(taxFee).Mul(model.GetNumDecimal()).InexactFloat64()
}

func (model *SaleOrderProduct) GetNumDecimal() decimal.Decimal {
	return decimal.NewFromFloat(model.Num)
}

// 获取销售订单商品的服务费税费。服务费税费=销售订单商品的服务费税费*销售订单商品的数量
func (model *SaleOrderProduct) GetServiceTaxFee() float64 {
	return decimal.NewFromFloat(model.ServiceTaxFee).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
}

// 获取销售订单商品的服务费税费。服务费税费=销售订单商品的服务费税费*销售订单商品的数量
func (model *SaleOrderProduct) GetOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	serviceTaxFee := model.calcServiceTaxFee(model.calcSalePrice(), serviceFeeRate, taxFeeType, serviceFeeType) // 服务费（折前）
	return decimal.NewFromFloat(serviceTaxFee).Mul(model.GetNumDecimal()).InexactFloat64()
}

// 获取销售订单商品的服务费。服务费=销售订单商品的服务费*销售订单商品的数量
func (model *SaleOrderProduct) GetServiceFee() float64 {
	return decimal.NewFromFloat(model.ServiceFee).Mul(model.GetNumDecimal()).InexactFloat64()
}

// 获取销售订单商品的原始服务费(折前价)。服务费=销售订单商品的服务费*销售订单商品的数量
func (model *SaleOrderProduct) GetOriginServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
	serviceFee := model.calcServiceFee(model.SalePrice, serviceFeeRate, taxFeeType, WithOriginPrice()) // 服务费（折前）
	return decimal.NewFromFloat(serviceFee).Mul(model.GetNumDecimal()).InexactFloat64()
}

func (model *SaleOrderProduct) GetMemberDiscountFee() float64 {
	return decimal.NewFromFloat(model.MemberDiscountFee).Mul(model.GetNumDecimal()).InexactFloat64()
}

func (model *SaleOrderProduct) GetCustomDiscountFee() float64 {
	return decimal.NewFromFloat(model.CustomDiscountFee).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
}

// 获取销售订单商品的未含税价格。
func (model *SaleOrderProduct) GetUnitPriceNoneTax() float64 {
	return model.SalePriceNoTax
}

func (model *SaleOrderProduct) SetAddOperation() {
	model.operation = "add"
}

func (model *SaleOrderProduct) SetSubOperation() {
	model.operation = "sub"
}

func (model *SaleOrderProduct) IsAddOperation() bool {
	// 如果operation为空，则默认是加购
	if model.operation == "" {
		return true
	}
	return model.operation == "add"
}

func (model *SaleOrderProduct) IsSubOperation() bool {
	return model.operation == "sub"
}

// 判断购物车中的商品是否可以编辑。只有套餐商品和非单规格的商品可以编辑
func (model *SaleOrderProduct) IsCanEdit() bool {
	// 已送出的商品不可编辑
	if model.IsSendKitchen() {
		return false
	}

	// 套餐商品可以编辑
	if model.IsPackageProduct() {
		return true
	}
	// 不是单规格的商品可以编辑
	if !model.IsSingleFlavorPackageProduct() {
		return true
	}

	// 默认不可编辑
	return false
}

// 判断套餐商品是否是单规格商品
func (model *SaleOrderProduct) IsSingleFlavorPackageProduct() bool {
	return model.ProductPackage.IsSingleFlavor()
}

// 设置打包时间
func (model *SaleOrderProduct) SetWrap() {
	defer model.SetUpdate() // 标记要更新model
	model.WrapTime = time.Now().Unix()
	if model.IsPackageProduct() {
		model.Sign = model.GeneratePackageSign()
	} else {
		model.Sign = model.GenerateProductSign()
	}
}

// 设置取消打包
func (model *SaleOrderProduct) SetUnwrap() {
	defer model.SetUpdate() // 标记要更新model
	model.WrapTime = 0      // 取消打包，打包时间置为0。注：暂时不更新商品的签名，历史遗留问题取消赠菜也没有更新签名，如果更新签名的话可能会与未打包的商品签名一致需要合并商品。
}

// GetAcceptTime 获取接单时间
func (model *SaleOrderProduct) GetAcceptTime() int64 {
	if model.H5Order != nil {
		return model.H5Order.HandleTime
	}
	return 0
}

// IsAcceptOrderProduct 商品是否已接单
func (model *SaleOrderProduct) IsAcceptOrderProduct() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

// 将已下单的h5订单商品变为已接单单的h5订单商品
func (model *SaleOrderProduct) SetAcceptOrderProduct() {
	defer model.SetUpdate() // 标记要更新model
	model.IsAcceptOrder = constant.OrderProductIsAcceptOrderAccepted
	model.Sign = model.GenerateProductSign() // 更新签名
}

// 将未下单的h5订单商品变为已下单的h5订单商品
func (model *SaleOrderProduct) SetH5OrderProduct(h5OrderUuid uint64) {
	model.unOrderH5Product = true // 标记为未下单的h5订单商品
	model.H5OrderUuid = h5OrderUuid
	model.Sign = model.GenerateProductSign() // 更新签名
	// model.H5OrderProductUuid = h5OrderProductUuid
}

// SetTaxRate 设置税率. 这个方法的定位是只能用这个方法设置商品的税率，不能用其他方式设置商品的税率
func (model *SaleOrderProduct) SetTaxRate(taxRate float64) {
	defer model.SetUpdate() // 标记要更新model
	model.TaxRate = taxRate
}

// SetFlavorPrice 设置规格价格
func (model *SaleOrderProduct) SetFlavorPrice(flavorPrice float64) {
	defer model.SetUpdate() // 标记要更新model
	model.FlavorPrice = flavorPrice
}

func (model *SaleOrderProduct) GetFlavorPrice() float64 {
	memberCardDiscountRate := model.GetMemberOrderDiscountRate()
	if memberCardDiscountRate != 1 {
		return decimal.NewFromFloat(model.FlavorPrice).Mul(decimal.NewFromFloat(memberCardDiscountRate)).Round(2).InexactFloat64()
	}
	return model.FlavorPrice
}

// GetCanReturnNum 获取销售订单商品的可退货数量. 可退货数量=订单商品数量-已退货数量
func (model *SaleOrderProduct) GetCanReturnNum() float64 {
	amount := decimal.NewFromFloat(0)
	for _, returnOrderProduct := range model.ReturnOrderProducts {
		amount = amount.Add(decimal.NewFromFloat(returnOrderProduct.Num))
	}
	num := decimal.NewFromFloat(model.Num).Sub(amount).InexactFloat64()
	// 如果可退货数量小于0，则返回0
	// 这个判断很有必要，否则会出现可退货数量为负数的情况但uint是无符号的，结果会得到一个很大的数。如2-14=18446744073709551604
	if num < 0 {
		return 0
	}
	return num
}

// GetReturnNum 获取销售订单商品的已退货数量. 已退货数量=订单商品数量-可退货数量
func (model *SaleOrderProduct) GetReturnNum() float64 {
	return decimal.NewFromFloat(model.Num).Sub(decimal.NewFromFloat(model.GetCanReturnNum())).InexactFloat64()
}

// GetReturnPrice 获取销售订单商品的已退金额。已退金额=订单商品金额*已退货数量
func (model *SaleOrderProduct) GetReturnPrice() float64 {
	if model.GetReturnNum() == 0 {
		return 0
	}
	returnPrice := decimal.NewFromFloat(model.TotalPrice).Mul(decimal.NewFromFloat(model.GetReturnNum())).Truncate(3).Round(2).InexactFloat64()
	return returnPrice
}

// GetCanReturnPrice 获取销售订单商品的可退货金额. 可退货金额=订单商品金额-已退货金额
func (model *SaleOrderProduct) GetCanReturnPrice() float64 {
	return decimal.NewFromFloat(model.TotalPrice).Mul(decimal.NewFromFloat(model.GetCanReturnNum())).Truncate(3).Round(2).InexactFloat64()
}

func (model *SaleOrderProduct) IsCurrentDeskProduct() bool {
	// 默认0是本台的商品。不为0的商品是从其他桌台并台过来的商品
	return model.DeskUuid == 0
}

// 是否为已送厨的商品。 状态为已送厨且生产订单ID不为0、且为已接单
func (model *SaleOrderProduct) IsCookingProduct() bool {
	// 状态为已送厨且生产订单ID不为0
	return model.Status == constant.SaleOrderProductStatusCooking && model.ProductionOrderUuid != 0 && model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

// 是否为未送厨的商品。 状态为未送厨且生产订单ID为0、且为已接单
func (model *SaleOrderProduct) IsUnCookingProduct() bool {
	return model.Status == constant.SaleOrderProductStatusNormal && model.ProductionOrderUuid == 0 && model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

// 设置该商品为自助餐商品
func (model *SaleOrderProduct) SetIsBuffet() {
	model.IsBuffet = constant.SaleOrderProductIsBuffetYes
}

// 设置该商品为非自助餐商品
func (model *SaleOrderProduct) SetNotBuffet() {
	model.IsBuffet = constant.SaleOrderProductIsBuffetNo
}

// 使用反射动态更新字段
func (model *SaleOrderProduct) SetFields(updateProduct SaleOrderProduct, specialFields map[string]bool) bool {
	// 使用反射动态更新字段
	productVal := reflect.ValueOf(model).Elem() // 因为product是指针
	// 更新字段
	updateVal := reflect.ValueOf(updateProduct)
	// 获取结构体类型信息
	typ := updateVal.Type()
	// 遍历所有字段
	for i := 0; i < updateVal.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		// 跳过BaseModel和关联对象字段
		if fieldName == "BaseModel" || field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Slice {
			continue
		}
		updateFieldVal := updateVal.Field(i)
		productFieldVal := productVal.FieldByName(fieldName)
		// 检查字段是否可设置
		if !productFieldVal.IsValid() || !productFieldVal.CanSet() {
			continue
		}
		// 特殊处理某些字段
		if specialFields[fieldName] {
			productFieldVal.Set(updateFieldVal)
			continue
		}
		// 对于普通字段，只有当它们不是零值时才设置
		isZero := false
		switch updateFieldVal.Kind() {
		case reflect.String:
			isZero = updateFieldVal.String() == ""
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isZero = updateFieldVal.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			isZero = updateFieldVal.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isZero = updateFieldVal.Float() == 0
		case reflect.Bool:
			isZero = !updateFieldVal.Bool()
		}
		if !isZero {
			productFieldVal.Set(updateFieldVal)
		}
	}
	return true
}

// 设置商品状态为送厨状态
func (model *SaleOrderProduct) SetCooking(productionOrderUuid uint64) {
	model.Status = constant.SaleOrderProductStatusCooking
	model.ProductionOrderUuid = productionOrderUuid
	model.Sign = model.GenerateProductSign()  // 更新签名
	model.SendKitchenTime = time.Now().Unix() // 送厨时间
	model.SetUpdate()                         // 标记该model需要更新
}

// 结账检查
func (model *SaleOrderProduct) CheckOutProduct() (int, string) {
	// 检查商品是否删除、下架、库存是否充足、价格变动
	for _, bom := range model.SaleOrderProductBoms {
		if bom.ProductBom.IsFlavor() {
			// 只检查结账减库存的商品
			if model.DeductStockType == constant.ProductPackageDeductStockTypePay {
				// 商品已经沽清
				if bom.ProductBom.IsSoldOutStatus() {
					return constant.CodeOrderCheckProductStockZero, "商品已经沽清"
				}
				// 下单商品数量超过库存数量. 只检查结账减库存的商品
				if bom.ProductBom.IsStockShortage(model.Num) {
					return constant.CodeOrderCheckProductStockZero, "下单商品数量超过库存数量"
				}
			}
			// 商品规格价格变动
			if bom.ProductBom.IsPriceChanged(model.FlavorPrice) {
				return constant.CodeOrderCheckProductPriceChanged, "商品规格价格变动"
			}
		}
		if bom.ProductBom.IsSauce() {
			// 只检查结账减库存的商品
			if model.DeductStockType == constant.ProductPackageDeductStockTypePay {
				// 商品已经沽清
				if bom.ProductBom.IsSoldOutStatus() {
					return constant.CodeOrderCheckProductStockZero, "小料已经售罄"
				}
				// 下单商品数量超过库存数量
				// 每个订单商品一个小料只消耗一个小料库存
				if bom.ProductBom.IsStockShortage(1) {
					return constant.CodeOrderCheckProductStockZero, "下单小料数量超过库存数量"
				}
			}
		}
	}

	// 商品是否享受会员折扣发生变动
	if model.memberDiscountChanged() {
		// 如果商品已经有会员折扣时才返回价格变动
		if model.MemberCardDiscountRate < 1 || model.MemberDiscountRate < 1 {
			return constant.CodeOrderCheckProductPriceChanged, "商品会员折扣发生变动"
		}
	}

	// 商品是否享受整单折扣发生变动
	if model.overallDiscountChanged() {
		// 如果商品已经有整单折扣时才返回价格变动
		if model.CustomDiscountRate < 1 {
			return constant.CodeOrderCheckProductPriceChanged, "商品整单折扣发生变动"
		}
	}

	// 小料的价格有变动
	if model.saucePriceChanged(model.SaucePrice) {
		return constant.CodeOrderCheckProductPriceChanged, "小料价格变动"
	}

	return constant.CodeSuccess, "商品检查通过"
}

// 送厨检查
func (model *SaleOrderProduct) CheckCookingProduct(lang string) (int, string) {
	// 如果是已送厨的商品，则不检查
	if model.IsCookingProduct() {
		return constant.CodeSuccess, "商品检查通过"
	}
	// 检查商品是否删除、下架、库存是否充足、价格变动
	for _, bom := range model.SaleOrderProductBoms {
		if bom.ProductBom.IsFlavor() {
			// 商品已经沽清
			if bom.ProductBom.IsSoldOutStatus() {
				return constant.CodeOrderCheckProductStockZero, "商品已经沽清"
			}
			// 下单商品数量超过库存数量
			if bom.ProductBom.IsStockShortageWithMaterial(model.Num) {
				productInfoString := model.GetNameAndFlavorName()
				return constant.CodeOrderCheckProductStockZero, fmt.Sprintf("%s 库存不足", productInfoString.GetLocale(lang))
			}
			// 商品已经下架
			if bom.ProductBom.IsProductPackageDown() {
				return constant.CodeOrderCheckProductDown, "商品已经下架"
			}
			// 商品规格已经下架
			if bom.ProductBom.IsDown() {
				return constant.CodeOrderCheckProductFlavorDown, "商品规格已经下架"
			}
			// 商品已经删除
			if bom.ProductBom.IsDelete() {
				return constant.CodeOrderCheckProductFlavorDown, "商品已经删除"
			}
			// 商品规格价格变动
			if bom.ProductBom.IsPriceChanged(model.FlavorPrice) {
				return constant.CodeOrderCheckProductPriceChanged, "商品规格价格变动"
			}
		}
		if bom.ProductBom.IsSauce() {
			// 小料已经沽清
			if bom.ProductBom.IsSoldOutStatus() {
				return constant.CodeOrderCheckProductStockZero, "小料已经售罄"
			}
			// 小料已经下架
			if bom.ProductBom.IsDown() {
				sauceName := bom.ProductBom.ProductSauce.MultiLanguageName.GetNameByLang(lang)
				tipsPrefix := i18n.Translate(lang, "加料")
				tipsPostfix := i18n.Translate(lang, "已下架，请重新选择其他加料") // "已下架，请重新选择其他加料"
				return constant.CodeOrderCheckProductSauceDown, fmt.Sprintf("%s %s %s", tipsPrefix, sauceName, tipsPostfix)
			}
			// 小料已经删除
			if bom.ProductBom.IsDelete() {
				sauceName := bom.ProductBom.ProductSauce.MultiLanguageName.GetNameByLang(lang)
				tipsPrefix := i18n.Translate(lang, "加料")
				tipsPostfix := i18n.Translate(lang, "已下架，请重新选择其他加料") // "已下架，请重新选择其他加料"
				return constant.CodeOrderCheckProductSauceDown, fmt.Sprintf("%s %s %s", tipsPrefix, sauceName, tipsPostfix)
			}
			// 下单商品数量超过库存数量
			// 每个订单商品一个小料只消耗一个小料库存
			if bom.ProductBom.IsStockShortageWithMaterial(1) {
				return constant.CodeOrderCheckProductStockZero, "下单小料数量超过库存数量"
			}
		}
	}

	// 小料的价格有变动
	if model.saucePriceChanged(model.SaucePrice) {
		return constant.CodeOrderCheckProductPriceChanged, "小料价格变动"
	}

	return constant.CodeSuccess, "商品检查通过"
}

// 商品是否享受会员折扣发生变动
func (model *SaleOrderProduct) memberDiscountChanged() bool {
	latestOpenMemberDiscount := model.ProductPackage.OpenDiscount
	return latestOpenMemberDiscount != model.OpenMemberDiscount
}

// 商品是否享受整单折扣发生变动
func (model *SaleOrderProduct) overallDiscountChanged() bool {
	latestOpenOverallDiscount := model.ProductPackage.OpenOverallDiscount
	return latestOpenOverallDiscount != model.OpenOverallDiscount
}

// 小料的价格是否有变动
func (model *SaleOrderProduct) saucePriceChanged(saucePrice float64) bool {
	price := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if bom.ProductBom.IsSauce() {
			price = price.Add(decimal.NewFromFloat(bom.ProductBom.Price))
		}
	}
	return price.InexactFloat64() != saucePrice
}

// 将model的关联对象置空
// 设置为0
func (model *SaleOrderProduct) SetNil() {
	model.MultiLanguageName = nil
	model.ImageFile = nil
	model.SaleOrderProductBoms = nil
	model.SaleOrderProductAttributes = nil
	model.ReturnOrderProducts = nil
	model.ProductPackage = nil
	model.SaleBill = nil
}

// 复制销售订单商品
func (model *SaleOrderProduct) CopyOrderProduct(saleOrderUuid uint64) *SaleOrderProduct {
	product := SaleOrderProduct{}
	err := copier.Copy(&product, model)
	if err != nil {
		return nil
	}
	// 重置base_model
	productUuid, _ := utils.GetID()
	product.BaseModel = BaseModel{
		Uuid: productUuid,
	}
	// 指定目标销售订单。如果不移动到别的销售订单可以修改销售订单uuid
	if saleOrderUuid != 0 {
		product.SaleOrderUuid = saleOrderUuid
	}
	// 复制SaleOrderProductBoms
	product.SaleOrderProductBoms = make([]*SaleOrderProductBom, 0)
	for _, bom := range model.SaleOrderProductBoms {
		newBom := bom.CopyBom(saleOrderUuid, productUuid)
		product.SaleOrderProductBoms = append(product.SaleOrderProductBoms, newBom)
	}
	// 复制SaleOrderProductAttributes
	product.SaleOrderProductAttributes = make([]*SaleOrderProductAttribute, 0)
	for _, attribute := range model.SaleOrderProductAttributes {
		newAttribute := attribute.CopyAttribute(model.SaleOrderUuid, productUuid)
		product.SaleOrderProductAttributes = append(product.SaleOrderProductAttributes, newAttribute)
	}
	return &product
}

// NewSaleOrderProductReasonList 新建一个销售订单商品的退菜原因列表
func (model *SaleOrderProduct) NewSaleOrderProductReasonList(reasons []*ReturnFoodReason) []*SaleOrderProductReason {
	list := make([]*SaleOrderProductReason, 0)
	for _, reason := range reasons {
		reasonUuid, _ := utils.GetID()
		list = append(list, &SaleOrderProductReason{
			BaseModel: BaseModel{
				Uuid: reasonUuid,
			},
			SaleOrderUuid:         model.SaleOrderUuid,
			SaleOrderProductUuid:  model.Uuid,
			MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
			ReturnFoodReasonUuid:  reason.Uuid,
		})
	}
	return list
}

// SetCancelInfo 设置订单商品的退菜信息，标记该商品为退菜商品
func (model *SaleOrderProduct) SetCancelInfo(reason string, reasons []*SaleOrderProductReason) {
	defer model.SetUpdate() // 标记该model需要更新
	model.CancelTime = time.Now().Unix()
	model.CancelReason = reason
	for index, _ := range reasons {
		reasons[index].SaleOrderProductUuid = model.Uuid // 设置退菜原因的销售订单商品
	}
	model.CancelReasons = append(model.CancelReasons, reasons...)

	model.ProductionOrderUuid = 0 // 取消生产订单关联
	// 更新签名
	if model.IsPackageSubProduct() {
		model.Sign = model.GeneratePackageSign()
	} else {
		model.Sign = model.GenerateProductSign()
	}
}

// 获取退菜原因
func (model *SaleOrderProduct) GetCancelReason() dto.LocaleResponse {
	zhNames := make([]string, 0)
	thNames := make([]string, 0)
	enNames := make([]string, 0)
	zhtwNames := make([]string, 0)
	jaNames := make([]string, 0)
	koNames := make([]string, 0)
	myNames := make([]string, 0)
	trNames := make([]string, 0)
	svNames := make([]string, 0)
	// 遍历选择的退菜原因
	for _, reason := range model.CancelReasons {
		if !reason.IsReturnFoodReason() {
			continue
		}
		if reason.IsDelete() {
			continue
		}
		zhNames = append(zhNames, reason.MultiLanguageName.ZhName)
		thNames = append(thNames, reason.MultiLanguageName.ThName)
		enNames = append(enNames, reason.MultiLanguageName.EnName)
		zhtwNames = append(zhtwNames, reason.MultiLanguageName.ZhTwName)
		jaNames = append(jaNames, reason.MultiLanguageName.JaName)
		koNames = append(koNames, reason.MultiLanguageName.KoName)
		myNames = append(myNames, reason.MultiLanguageName.MyName)
		trNames = append(trNames, reason.MultiLanguageName.TrName)
		svNames = append(svNames, reason.MultiLanguageName.SvName)
	}
	// 添加自定义的退菜原因
	if model.CancelReason != "" {
		zhNames = append(zhNames, model.CancelReason)
		thNames = append(thNames, model.CancelReason)
		enNames = append(enNames, model.CancelReason)
		zhtwNames = append(zhtwNames, model.CancelReason)
		jaNames = append(jaNames, model.CancelReason)
		koNames = append(koNames, model.CancelReason)
		myNames = append(myNames, model.CancelReason)
		trNames = append(trNames, model.CancelReason)
		svNames = append(svNames, model.CancelReason)
	}
	reasonDto := dto.LocaleResponse{
		ZH:   strings.Join(zhNames, "、"),
		TH:   strings.Join(thNames, "、"),
		EN:   strings.Join(enNames, "、"),
		ZHTW: strings.Join(zhtwNames, "、"),
		JA:   strings.Join(jaNames, "、"),
		KO:   strings.Join(koNames, "、"),
		MY:   strings.Join(myNames, "、"),
		TR:   strings.Join(trNames, "、"),
		SV:   strings.Join(svNames, "、"),
	}
	return reasonDto
}

// 获取赠品原因
func (model *SaleOrderProduct) GetGiftReason() dto.LocaleResponse {
	zhNames := make([]string, 0)
	thNames := make([]string, 0)
	enNames := make([]string, 0)
	zhtwNames := make([]string, 0)
	jaNames := make([]string, 0)
	koNames := make([]string, 0)
	myNames := make([]string, 0)
	trNames := make([]string, 0)
	svNames := make([]string, 0)
	// 遍历选择的赠品原因
	for _, reason := range model.CancelReasons {
		if !reason.IsGiftReason() {
			continue
		}
		zhNames = append(zhNames, reason.MultiLanguageName.ZhName)
		thNames = append(thNames, reason.MultiLanguageName.ThName)
		enNames = append(enNames, reason.MultiLanguageName.EnName)
		zhtwNames = append(zhtwNames, reason.MultiLanguageName.ZhTwName)
		jaNames = append(jaNames, reason.MultiLanguageName.JaName)
		koNames = append(koNames, reason.MultiLanguageName.KoName)
		myNames = append(myNames, reason.MultiLanguageName.MyName)
		trNames = append(trNames, reason.MultiLanguageName.TrName)
		svNames = append(svNames, reason.MultiLanguageName.SvName)
	}
	// 添加自定义的退菜原因
	if model.GiftReason != "" {
		zhNames = append(zhNames, model.GiftReason)
		thNames = append(thNames, model.GiftReason)
		enNames = append(enNames, model.GiftReason)
		zhtwNames = append(zhtwNames, model.GiftReason)
		jaNames = append(jaNames, model.GiftReason)
		koNames = append(koNames, model.GiftReason)
		myNames = append(myNames, model.GiftReason)
		trNames = append(trNames, model.GiftReason)
		svNames = append(svNames, model.GiftReason)
	}
	reasonDto := dto.LocaleResponse{
		ZH:   strings.Join(zhNames, "、"),
		TH:   strings.Join(thNames, "、"),
		EN:   strings.Join(enNames, "、"),
		ZHTW: strings.Join(zhtwNames, "、"),
		JA:   strings.Join(jaNames, "、"),
		KO:   strings.Join(koNames, "、"),
		MY:   strings.Join(myNames, "、"),
		TR:   strings.Join(trNames, "、"),
		SV:   strings.Join(svNames, "、"),
	}
	return reasonDto
}

// SetCancelInfo 设置订单商品的退菜信息，标记该商品为退菜商品
//func (model *SaleOrderProduct) SetCancelInfo(reasons []*ReturnFoodReason) {
//	defer model.SetUpdate() // 标记该model需要更新
//	model.CancelTime = time.Now().Unix()
//	list := make([]*SaleOrderProductReason, 0)
//	for _, reason := range reasons {
//		reasonUuid, _ := utils.GetID()
//		list = append(list, &SaleOrderProductReason{
//			BaseModel: BaseModel{
//				Uuid: reasonUuid,
//			},
//			SaleOrderUuid:         model.SaleOrderUuid,
//			SaleOrderProductUuid:  model.Uuid,
//			ReturnFoodReasonUuid:  reason.Uuid,
//			MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
//		})
//	}
//	model.CancelReasons = list
//}

// 从商品的原因列表中，只取出退菜原因
func (model *SaleOrderProduct) GetCancelReasons() []*SaleOrderProductReason {
	list := make([]*SaleOrderProductReason, 0)
	for _, reason := range model.CancelReasons {
		// 严谨性检查。避免有不是该商品的退菜原因
		if reason.SaleOrderProductUuid != model.Uuid || reason.SaleOrderUuid != model.SaleOrderUuid {
			continue
		}
		// 三选一。只取出退菜原因，忽略赠菜原因和免单原因
		if reason.ReturnFoodReasonUuid != 0 {
			list = append(list, reason)
		}
	}
	return list
}

// 设置销售订单商品的数量
func (model *SaleOrderProduct) SetNum(num float64) {
	defer model.SetUpdate() // 标记该model需要更新
	model.Num = num
}

// 设置销售订单商品的折扣信息
func (model *SaleOrderProduct) SetDiscountInfo(memberDiscountRate, memberCardDiscountRate, customDiscountRate float64) {
	model.SetMemberDiscountInfo(memberDiscountRate, memberCardDiscountRate)
	model.CustomDiscountRate = customDiscountRate
}

func (model *SaleOrderProduct) SetMemberDiscountInfo(memberDiscountRate, memberCardDiscountRate float64) {
	model.MemberDiscountRate = memberDiscountRate
	model.MemberCardDiscountRate = memberCardDiscountRate
}

// 设置销售订单商品的必点信息，标记该商品是必点商品，标记该商品是某个必点方案的必点商品
func (model *SaleOrderProduct) SetMustPlanInfo(mustPlanUuid uint64) {
	model.IsRequire = constant.IsMustProductYes
	model.MustPlanUuid = mustPlanUuid
}

// 标记该订单商品相关的资源为删除
func (model *SaleOrderProduct) DeleteProduct() {
	defer model.SetUpdate()
	deleteTime := time.Now().Unix()
	model.DeleteTime = deleteTime
	for index, _ := range model.SaleOrderProductBoms {
		saleOrderProductBom := model.SaleOrderProductBoms[index]
		saleOrderProductBom.DeleteTime = deleteTime
	}
	for index, _ := range model.SaleOrderProductAttributes {
		saleOrderProductAttribute := model.SaleOrderProductAttributes[index]
		saleOrderProductAttribute.DeleteTime = deleteTime
	}
}

// 订单商品是否已接单
func (model *SaleOrderProduct) IsAcceptOrderBool() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

// 订单商品是否未接单
func (model *SaleOrderProduct) IsUnAcceptOrderBool() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderUnAccept
}

// 是否是H5下单商品。 未接单且有h5订单uuid
func (model *SaleOrderProduct) IsH5OrderProductBool() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderUnAccept && model.H5OrderUuid != 0
}

// 是否是H5未下单商品。 未接单且无h5订单uuid
func (model *SaleOrderProduct) IsUnOrderH5OrderProduct() bool {
	if model.unOrderH5Product {
		// 如果是特别标记的未下单的h5订单商品，返回true。此类商品此时正在下单的处理程序中
		return true
	}
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderUnAccept && model.H5OrderUuid == 0
}

// 是否时h5购物车的商品，h5未下单的商品。未接单且无h5订单uuid
func (model *SaleOrderProduct) IsH5CartProduct() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderUnAccept && model.H5OrderUuid == 0
}

func (model *SaleOrderProduct) ChangeProductPrice(price float64) {
	model.ChangePriceTime = time.Now().Unix()
	model.SalePrice = price
	// 重新签名商品
	model.Sign = model.GenerateProductSign()
}

// 获取商品销售价(折前价)
func (model *SaleOrderProduct) GetSalePrice() float64 {
	// 销售价*数量
	salePrice := decimal.NewFromFloat(model.SalePrice).Mul(model.GetNumDecimal()).Round(2).InexactFloat64()
	return salePrice
}

// 获取商品销售价(折前价) 单个商品
func (model *SaleOrderProduct) GetSalePriceUnit() float64 {
	// 销售价
	salePrice := decimal.NewFromFloat(model.SalePrice).Truncate(2).InexactFloat64()
	return salePrice
}

// 获取商品的最终售价（单价）。最终售价为商品的最终单价，默认等于Price，如果免单，则等于0
func (model *SaleOrderProduct) GetFinalSalePrice() float64 {
	if model.IsGiftProduct() {
		return 0
	}
	return model.Price
}

// 获取商品的最终售价（*数量）。商品的最终售价*数量
func (model *SaleOrderProduct) GetProductFinalSalePrice() float64 {
	return decimal.NewFromFloat(model.GetFinalSalePrice()).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
}

// 获取商品总金额（折前价）（商品原价）(包括税费)
func (model *SaleOrderProduct) GetOriginTotalPriceWithTax() float64 {
	// 金额*数量
	price := decimal.NewFromFloat(model.ProductPrice).Add(decimal.NewFromFloat(model.TaxFee)).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
	return price
}

// 获取商品总金额（折前价）（商品原价）
func (model *SaleOrderProduct) GetTotalProductPrice() float64 {
	// 金额*数量
	price := decimal.NewFromFloat(model.ProductPrice).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
	return price
}

// 获取商品总金额（折后价）
func (model *SaleOrderProduct) GetTotalPrice() float64 {
	// 金额*数量
	price := decimal.NewFromFloat(model.TotalPrice).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
	return price
}

// 获取商品总小料价格（折后价）
func (model *SaleOrderProduct) GetTotalSaucePrice() float64 {
	// 金额*数量
	price := decimal.NewFromFloat(model.SaucePrice).Mul(model.GetNumDecimal()).Round(2).InexactFloat64()
	return price
}

// 获取商品总金额（折前价）
func (model *SaleOrderProduct) GetTotalPriceOrigin() float64 {
	originTotalPrice := model.OriginTotalPrice
	// 如果折前价为0，则使用折后价. 兼容未加OriginTotalPrice字段前的订单数据
	if originTotalPrice == 0 {
		originTotalPrice = model.GetTotalPrice()
	}
	// 金额*数量
	price := decimal.NewFromFloat(originTotalPrice).Mul(model.GetNumDecimal()).Truncate(3).Round(2).InexactFloat64()
	return price
}

// 获取最终价格（折后价）
func (model *SaleOrderProduct) GetPrice() float64 {
	// 最终价格*数量
	price := decimal.NewFromFloat(model.Price).Mul(model.GetNumDecimal()).Round(2).InexactFloat64()
	return price
}

// 是否必须商品
func (model *SaleOrderProduct) IsMustProduct() bool {
	return model.IsRequire == 1
}

// 是否是赠品
func (model *SaleOrderProduct) IsGiftProduct() bool {
	return model.GiftTime > 0
}

// 专门处理 IsWrap 逻辑的方法
func (sop *SaleOrderProduct) CalculateIsWrap(saleBill *SaleBill) bool {
	// 如果是打包订单且不是会员订单，则强制打包
	if saleBill.IsTakeout() && saleBill.MemberSaleOrderUuid == 0 {
		return true
	}
	return sop.IsWrapProduct()
}

// 是否是包装商品
func (model *SaleOrderProduct) IsWrapProduct() bool {
	return model.WrapTime > 0
}

// 是否取消商品
func (model *SaleOrderProduct) IsCancelProduct() bool {
	return model.CancelTime > 0
}

// 是否送到厨房
func (model *SaleOrderProduct) IsSendKitchen() bool {
	return model.Status == constant.OrderProductStatusSentKitchen
}

// 判断是哪个业务状态
func (model *SaleOrderProduct) StatusValue() int {
	return int(model.Status)
}

// 获取该订单商品的材料组成及用量。
// 如一个珍珠奶茶加料珍珠，则计算成分珍珠、奶、茶等各个原材料等用量
func (model *SaleOrderProduct) GetMaterialBom() []*ProductionOrderMaterial {
	return nil // TODO 植焕
}

// 获取商品的名称。格式：`商品名 (规格名)`
func (model *SaleOrderProduct) GetNameAndFlavorName() dto.LocaleResponse {
	var flavorName dto.LocaleResponse
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if saleOrderProductBom.IsFlavor() {
			flavorName = saleOrderProductBom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
		}
	}
	productPackageName := model.MultiLanguageName.GetNames()

	flavorNameZH := fmt.Sprintf(" (%s)", flavorName.ZH)
	flavorNameTH := fmt.Sprintf(" (%s)", flavorName.TH)
	flavorNameEN := fmt.Sprintf(" (%s)", flavorName.EN)
	flavorNameZHTW := fmt.Sprintf(" (%s)", flavorName.ZHTW)
	flavorNameJA := fmt.Sprintf(" (%s)", flavorName.JA)
	flavorNameKO := fmt.Sprintf(" (%s)", flavorName.KO)
	flavorNameMY := fmt.Sprintf(" (%s)", flavorName.MY)
	flavorNameTR := fmt.Sprintf(" (%s)", flavorName.TR)
	flavorNameSV := fmt.Sprintf(" (%s)", flavorName.SV)
	if model.IsPackageProduct() {
		flavorNameZH = ""
		flavorNameTH = ""
		flavorNameEN = ""
		flavorNameZHTW = ""
		flavorNameJA = ""
		flavorNameKO = ""
		flavorNameMY = ""
		flavorNameTR = ""
		flavorNameSV = ""
	}
	return dto.LocaleResponse{
		ZH:   fmt.Sprintf("%s%s", productPackageName.ZH, flavorNameZH),
		TH:   fmt.Sprintf("%s%s", productPackageName.TH, flavorNameTH),
		EN:   fmt.Sprintf("%s%s", productPackageName.EN, flavorNameEN),
		ZHTW: fmt.Sprintf("%s%s", productPackageName.ZHTW, flavorNameZHTW),
		JA:   fmt.Sprintf("%s%s", productPackageName.JA, flavorNameJA),
		KO:   fmt.Sprintf("%s%s", productPackageName.KO, flavorNameKO),
		MY:   fmt.Sprintf("%s%s", productPackageName.MY, flavorNameMY),
		TR:   fmt.Sprintf("%s%s", productPackageName.TR, flavorNameTR),
		SV:   fmt.Sprintf("%s%s", productPackageName.SV, flavorNameSV),
	}
}

// 获取商品的名称。格式：`商品名 (规格名)`
func (model *SaleOrderProduct) GetNameAndFlavorNameFrom(ProductBom *ProductBom, productName *MultiLanguageName) dto.LocaleResponse {
	flavorName := ProductBom.ProductFlavor.MultiLanguageName.GetNames()
	productPackageName := productName.GetNames()

	flavorNameZH := fmt.Sprintf(" (%s)", flavorName.ZH)
	flavorNameTH := fmt.Sprintf(" (%s)", flavorName.TH)
	flavorNameEN := fmt.Sprintf(" (%s)", flavorName.EN)
	flavorNameZHTW := fmt.Sprintf(" (%s)", flavorName.ZHTW)
	flavorNameJA := fmt.Sprintf(" (%s)", flavorName.JA)
	flavorNameKO := fmt.Sprintf(" (%s)", flavorName.KO)
	flavorNameMY := fmt.Sprintf(" (%s)", flavorName.MY)
	flavorNameTR := fmt.Sprintf(" (%s)", flavorName.TR)
	flavorNameSV := fmt.Sprintf(" (%s)", flavorName.SV)
	if model.IsPackageProduct() {
		flavorNameZH = ""
		flavorNameTH = ""
		flavorNameEN = ""
		flavorNameZHTW = ""
		flavorNameJA = ""
		flavorNameKO = ""
		flavorNameMY = ""
		flavorNameTR = ""
		flavorNameSV = ""
	}
	return dto.LocaleResponse{
		ZH:   fmt.Sprintf("%s%s", productPackageName.ZH, flavorNameZH),
		TH:   fmt.Sprintf("%s%s", productPackageName.TH, flavorNameTH),
		EN:   fmt.Sprintf("%s%s", productPackageName.EN, flavorNameEN),
		ZHTW: fmt.Sprintf("%s%s", productPackageName.ZHTW, flavorNameZHTW),
		JA:   fmt.Sprintf("%s%s", productPackageName.JA, flavorNameJA),
		KO:   fmt.Sprintf("%s%s", productPackageName.KO, flavorNameKO),
		MY:   fmt.Sprintf("%s%s", productPackageName.MY, flavorNameMY),
		TR:   fmt.Sprintf("%s%s", productPackageName.TR, flavorNameTR),
		SV:   fmt.Sprintf("%s%s", productPackageName.SV, flavorNameSV),
	}
}

func (model *SaleOrderProduct) GetAttributeName() dto.LocaleResponse {
	return getLocaleResponse(model.GetAttributeNameList(), ";")
}

func (model *SaleOrderProduct) GetAttributeNameList() []dto.LocaleResponse {
	var flavorName dto.LocaleResponse
	var sauceNames []dto.LocaleResponse
	var attributeNames []dto.LocaleResponse

	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if saleOrderProductBom.IsFlavor() {
			flavorName = saleOrderProductBom.ProductBom.ProductFlavor.MultiLanguageName.GetNames()
		} else {
			sauceName := saleOrderProductBom.ProductBom.ProductSauce.MultiLanguageName.GetNames()
			sauceNames = append(sauceNames, sauceName)
		}
	}

	// 获取商品属性
	for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
		attributeName := saleOrderProductAttribute.ProductAttribute.MultiLanguageName.GetNames()
		attributeNames = append(attributeNames, attributeName)
	}
	// 根据规格生成字符串。`(规格；属性；小料)`
	nameList := make([]dto.LocaleResponse, 0)
	nameList = append(nameList, flavorName)
	if len(attributeNames) > 0 {
		nameList = append(nameList, attributeNames...)
	}
	if len(sauceNames) > 0 {
		nameList = append(nameList, sauceNames...)
	}

	return nameList
}

func (model *SaleOrderProduct) GetSauceNamesList() []dto.LocaleResponse {
	var sauceNames []dto.LocaleResponse
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if !saleOrderProductBom.IsFlavor() {
			sauceName := saleOrderProductBom.ProductBom.ProductSauce.MultiLanguageName.GetNames()
			sauceNames = append(sauceNames, sauceName)
		}
	}
	nameList := make([]dto.LocaleResponse, 0)
	if len(sauceNames) > 0 {
		nameList = append(nameList, sauceNames...)
	}
	return nameList
}

// 将多个LocaleResponse合并成一个
func getLocaleResponse(nameList []dto.LocaleResponse, div string) dto.LocaleResponse {
	if len(nameList) == 0 {
		return dto.LocaleResponse{}
	}
	attributeResultNames := dto.LocaleResponse{}
	for index, name := range nameList {
		attributeResultNames.ZH += name.ZH
		attributeResultNames.TH += name.TH
		attributeResultNames.EN += name.EN
		attributeResultNames.ZHTW += name.ZHTW
		attributeResultNames.JA += name.JA
		attributeResultNames.KO += name.KO
		attributeResultNames.MY += name.MY
		attributeResultNames.TR += name.TR
		attributeResultNames.SV += name.SV
		if attributeResultNames.ZH != "" && index != len(nameList)-1 {
			attributeResultNames.ZH += div
			attributeResultNames.TH += div
			attributeResultNames.EN += div
			attributeResultNames.ZHTW += div
			attributeResultNames.JA += div
			attributeResultNames.KO += div
			attributeResultNames.MY += div
			attributeResultNames.TR += div
			attributeResultNames.SV += div
		}
	}
	return attributeResultNames
}

func (model *SaleOrderProduct) GetAttributeNamesByLang(lang string, showSku ...bool) string {
	var flavorName string
	var sauceNames []string
	var attributeNames []string
	isShowSku := true
	if len(showSku) > 0 {
		isShowSku = showSku[0]
	}
	for _, saleOrderProductBom := range model.SaleOrderProductBoms {
		if saleOrderProductBom.IsFlavor() {
			flavorName = saleOrderProductBom.ProductBom.ProductFlavor.MultiLanguageName.GetNameByLang(lang)
		} else {
			sauceName := saleOrderProductBom.ProductBom.ProductSauce.MultiLanguageName.GetNameByLang(lang)
			sauceNames = append(sauceNames, sauceName)
		}
	}
	// 获取商品属性
	for _, saleOrderProductAttribute := range model.SaleOrderProductAttributes {
		attributeName := saleOrderProductAttribute.ProductAttribute.MultiLanguageName.GetNameByLang(lang)
		attributeNames = append(attributeNames, attributeName)
	}
	// 根据规格生成字符串。`(规格；属性；小料)`
	nameList := make([]string, 0)
	// 是否显示sku
	if isShowSku {
		nameList = append(nameList, flavorName)
		if len(attributeNames) > 0 {
			nameList = append(nameList, attributeNames...)
		}
	}
	// 小料
	if len(sauceNames) > 0 {
		nameList = append(nameList, sauceNames...)
	}
	return strings.Join(nameList, ";")
}

// 点餐时录入的原始数据
type DefaultSaleOrderProduct struct {
	DeviceId                string // 设备ID,用于标识订单来源设备.来源h5时，device_id为h5的device_id
	Name                    string
	OpenMemberDiscount      uint
	TaxRate                 float64
	DeductStockType         uint
	MultiLanguageNameUuid   uint64
	ImageFileUuid           uint64
	ProductPackageUuid      uint64
	SaleBillUuid            uint64
	SaleOrderUuid           uint64
	MemberDiscountRate      float64
	MemberCardDiscountRate  float64
	CustomDiscountRate      float64
	Sauces                  []Sauce
	Flavor                  Flavor
	Attribute               []Attribute
	IsAcceptOrder           uint    // 是否接单
	Num                     float64 // 数量
	NumType                 uint    // 数量类型
	Remark                  string  // 备注
	PackageUuid             uint64  // 套餐uuid
	PackageGroupUuid        uint64  // 套餐分组uuid
	ProductType             uint8   // 商品类型
	PackageSubProductParams string  // 套餐子商品参数
}

func NewDefaultSaleOrderProduct(def DefaultSaleOrderProduct, productPackage *ProductPackage, operation string) *SaleOrderProduct {
	saleOrderProductUuid, _ := utils.GetID()
	saleOrderProductBoms := make([]*SaleOrderProductBom, 0)
	for _, bom := range def.Sauces {
		saleOrderProductBom := NewSaleOrderProductSauce(saleOrderProductUuid, def.SaleOrderUuid, bom)
		saleOrderProductBoms = append(saleOrderProductBoms, saleOrderProductBom)
	}
	saleOrderProductBoms = append(saleOrderProductBoms, NewSaleOrderProductFlavor(saleOrderProductUuid, def.SaleOrderUuid, def.Flavor))

	saleOrderProductAttributes := []*SaleOrderProductAttribute{}
	for _, attribute := range def.Attribute {
		saleOrderProductAttribute := NewSaleOrderProductAttribute(saleOrderProductUuid, def.SaleOrderUuid, attribute)
		saleOrderProductAttributes = append(saleOrderProductAttributes, saleOrderProductAttribute)
	}
	product := SaleOrderProduct{
		BaseModel: BaseModel{
			Uuid: saleOrderProductUuid,
		},
		DeviceId:                   def.DeviceId,
		Name:                       def.Name,
		Remark:                     def.Remark,
		FlavorName:                 def.Flavor.Name,
		Num:                        def.Num,
		NumType:                    def.NumType,
		Status:                     constant.OrderProductStatusUnSending,
		IsAcceptOrder:              def.IsAcceptOrder,
		FlavorPrice:                def.Flavor.Price,
		OpenMemberDiscount:         def.OpenMemberDiscount,
		MemberDiscountRate:         def.MemberDiscountRate,
		MemberCardDiscountRate:     def.MemberCardDiscountRate,
		CustomDiscountRate:         def.CustomDiscountRate,
		OpenOverallDiscount:        productPackage.OpenOverallDiscount,
		DeductStockType:            def.DeductStockType,
		MultiLanguageNameUuid:      def.MultiLanguageNameUuid,
		ImageFileUuid:              def.ImageFileUuid,
		ProductPackageUuid:         def.ProductPackageUuid,
		SaleBillUuid:               def.SaleBillUuid,
		SaleOrderUuid:              def.SaleOrderUuid,
		SaleOrderProductBoms:       saleOrderProductBoms,
		SaleOrderProductAttributes: saleOrderProductAttributes,
		PackageUuid:                def.PackageUuid,
		PackageGroupUuid:           def.PackageGroupUuid,
		ProductType:                def.ProductType,
		PackageSubProductParams:    def.PackageSubProductParams,
	}
	// 套餐子商品，设置单位数量
	if def.ProductType == constant.ProductTypePackageSubProduct {
		product.UnitNum = def.Num
	}
	product.SetTaxRate(def.TaxRate)
	// 设置商品包. 加购并送厨时用到，用于计算限购
	{
		product.ProductPackage = productPackage
		product.MultiLanguageName = &productPackage.MultiLanguageName
	}

	if operation == "sub" {
		product.SetSubOperation()
	}
	return &product
}

// 构建“销售订单商品”的规格
func NewSaleOrderProductFlavor(saleOrderProductUuid uint64, saleOrderUuid uint64, flavor Flavor) *SaleOrderProductBom {
	return &SaleOrderProductBom{
		Name:                 flavor.Name,
		Price:                flavor.Price,
		IsFlavorBom:          1,
		SaleOrderUuid:        saleOrderUuid,
		SaleOrderProductUuid: saleOrderProductUuid,
		ProductBomUuid:       flavor.ProductBomUuid,
	}
}

// 构建“销售订单商品”的加料
func NewSaleOrderProductSauce(saleOrderProductUuid uint64, saleOrderUuid uint64, sauce Sauce) *SaleOrderProductBom {
	return &SaleOrderProductBom{
		Name:                 sauce.Name,
		Price:                sauce.Price,
		IsFlavorBom:          0,
		SaleOrderProductUuid: saleOrderProductUuid,
		SaleOrderUuid:        saleOrderUuid,
		ProductBomUuid:       sauce.ProductBomUuid,
	}
}

// 构建“销售订单商品”的属性
func NewSaleOrderProductAttribute(saleOrderProductUuid uint64, saleOrderUuid uint64, attribute Attribute) *SaleOrderProductAttribute {
	return &SaleOrderProductAttribute{
		Name:                        attribute.Name,
		SaleOrderUuid:               saleOrderUuid,
		SaleOrderProductUuid:        saleOrderProductUuid,
		ProductAttributeUuid:        attribute.ProductAttributeUuid,
		ProductPackageAttributeUuid: attribute.ProductPackageAttributeUuid,
	}
}

// 判断商品有没有改价
func (model *SaleOrderProduct) IsCustomPriceBool() bool {
	return model.ChangePriceTime > 0
}

func (model *SaleOrderProduct) ProductKey() string {
	flavorUuid := uint64(0)
	sauceUuidList := make([]uint64, 0)
	attributeIdList := make([]uint64, 0)

	// 物料ID列表
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavor() {
			flavorUuid = bom.ProductBomUuid
		} else if bom.IsSauce() {
			sauceUuidList = append(sauceUuidList, bom.ProductBomUuid)
		}
	}
	// 属性ID列表
	for _, attribute := range model.SaleOrderProductAttributes {
		attributeIdList = append(attributeIdList, attribute.ProductAttributeUuid)
	}

	// 物料ID列表和属性ID列表排序
	sort.Slice(sauceUuidList, func(i, j int) bool {
		return sauceUuidList[i] < sauceUuidList[j]
	})
	sort.Slice(attributeIdList, func(i, j int) bool {
		return attributeIdList[i] < attributeIdList[j]
	})

	sauceUuidStrList := make([]string, 0)
	for _, sauceUuid := range sauceUuidList {
		sauceUuidStrList = append(sauceUuidStrList, fmt.Sprintf("%d", sauceUuid))
	}
	attributeIdStrList := make([]string, 0)
	for _, attributeId := range attributeIdList {
		attributeIdStrList = append(attributeIdStrList, fmt.Sprintf("%d", attributeId))
	}

	// 按照“规格id-属性id,属性id-加料id,加料id”的格式拼接
	return fmt.Sprintf("%d-%s-%s", flavorUuid, strings.Join(attributeIdStrList, ","), strings.Join(sauceUuidStrList, ","))
}

// GenerateProductSign 生成商品包签名. 相同的商品，商品签名相同,用于取消拆单时合并商品。
// 格式：物料,物料,物料-属性,属性,属性-备注内容-必点方案uuid-送厨批次uuid-改价时间-赠菜时间-打包时间-退菜原因-H5OrderUuid-是否接单-套餐uuid
// 更新签名的场景：
// 1 改价销售订单商品价格后要重新生成签名
// 2 修改备注
// 3 送厨
// 4 赠菜
// 5 打包
// 6 退菜
// 7 h5下单
// 8 接单
func (model *SaleOrderProduct) GenerateProductSign() string {
	bomIdList := make([]string, 0)
	attributeIdList := make([]string, 0)

	// 物料ID列表
	for _, bom := range model.SaleOrderProductBoms {
		bomIdList = append(bomIdList, strconv.FormatUint(bom.ProductBomUuid, 10))
	}
	// 属性ID列表
	for _, attributeGroup := range model.SaleOrderProductAttributes {
		attributeIdList = append(attributeIdList, strconv.FormatUint(attributeGroup.ProductAttributeUuid, 10))
	}
	// 物料ID列表和属性ID列表排序
	sort.Slice(bomIdList, func(i, j int) bool {
		return bomIdList[i] < bomIdList[j]
	})
	sort.Slice(attributeIdList, func(i, j int) bool {
		return attributeIdList[i] < attributeIdList[j]
	})
	// 物料ID列表和属性ID列表拼接。格式：物料,物料,物料-属性,属性,属性-备注内容-必点方案uuid-送厨批次uuid-改价时间-赠菜时间-打包时间-退菜原因-H5OrderUuid-是否接单
	bomIdListStr := strings.Join(bomIdList, ",")
	attributeIdListStr := strings.Join(attributeIdList, ",")

	// 构建商品的退菜原因
	type Reason struct {
		Uuids []uint64 `json:"uuids"`
		Text  string   `json:"text"`
	}
	reason := Reason{Uuids: make([]uint64, 0), Text: model.CancelReason}
	for _, item := range model.CancelReasons {
		reason.Uuids = append(reason.Uuids, item.ReturnFoodReasonUuid)
	}
	reasonStr := utils.ToJson(reason)

	return fmt.Sprintf("%s-%s-%s-%d-%d-%d-%d-%d-%s-%d-%d-%d",
		bomIdListStr,
		attributeIdListStr,
		model.Remark,
		model.MustPlanUuid,
		model.ProductionOrderUuid,
		model.ChangePriceTime,
		model.GiftTime,
		model.WrapTime,
		reasonStr,
		model.H5OrderUuid,
		model.IsAcceptOrder, // 是否接单. 为了让未下单的h5商品和未送厨的商品不被合并在一起
		model.PackageUuid,
	)
}

// GeneratePackageSign 生成商品套餐签名. 相同的商品套餐，商品套餐签名相同,用于取消拆单时合并商品。
// 格式：套餐uuid-[子商品规格uuid,属性,属性;子商品规格uuid,属性,属性;]-备注内容-送厨批次uuid-改价时间-赠菜时间-打包时间-退菜原因-H5OrderUuid-是否接单
// 更新签名的场景：
// 1 改价销售订单商品价格后要重新生成签名
// 2 修改备注
// 3 送厨
// 4 赠菜
// 5 打包
// 6 退菜
// 7 h5下单
// 8 接单
func (model *SaleOrderProduct) GeneratePackageSign() string {
	packageUuid := model.ProductPackageUuid

	// 构建商品的退菜原因
	type Reason struct {
		Uuids []uint64 `json:"uuids"`
		Text  string   `json:"text"`
	}
	reason := Reason{Uuids: make([]uint64, 0), Text: model.CancelReason}
	for _, item := range model.CancelReasons {
		reason.Uuids = append(reason.Uuids, item.ReturnFoodReasonUuid)
	}
	reasonStr := utils.ToJson(reason)

	return fmt.Sprintf("%d-%s-%s-%d-%d-%d-%d-%d-%s-%d-%d",
		packageUuid,
		model.PackageSubProductParams,
		model.Remark,
		model.MustPlanUuid,
		model.ProductionOrderUuid,
		model.ChangePriceTime,
		model.GiftTime,
		model.WrapTime,
		reasonStr,
		model.H5OrderUuid,
		model.IsAcceptOrder, // 是否接单. 为了让未下单的h5商品和未送厨的商品不被合并在一起
	)
}

// GetAttributeNames 获取属性名称字符串
func (model *SaleOrderProduct) GetAttributeNames() string {
	attributeNames := []string{}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom == 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	for _, attribute := range model.SaleOrderProductAttributes {
		attributeNames = append(attributeNames, attribute.Name)
	}
	for _, bom := range model.SaleOrderProductBoms {
		if bom.IsFlavorBom != 1 {
			attributeNames = append(attributeNames, bom.Name)
		}
	}
	return strings.Join(attributeNames, "; ")
}

// 获取销售订单商品的属性uuid列表
func (model *SaleOrderProduct) GetAttributeUuidList() []uint64 {
	attributeUuidList := make([]uint64, 0)
	for _, attribute := range model.SaleOrderProductAttributes {
		attributeUuidList = append(attributeUuidList, attribute.ProductPackageAttributeUuid)
	}
	sort.Slice(attributeUuidList, func(i, j int) bool {
		return attributeUuidList[i] < attributeUuidList[j]
	})
	return attributeUuidList
}

// 获取销售订单商品的加料uuid列表
func (model *SaleOrderProduct) GetBomUuidList() []uint64 {
	bomUuidList := make([]uint64, 0)
	for _, bom := range model.SaleOrderProductBoms {
		if !bom.IsSauce() {
			continue
		}
		bomUuidList = append(bomUuidList, bom.ProductBomUuid)
	}
	sort.Slice(bomUuidList, func(i, j int) bool {
		return bomUuidList[i] < bomUuidList[j]
	})
	return bomUuidList
}

// 生成规格属性加料签名。格式为：商品规格uuid:属性uuid1,属性uuid2,属性uuid3:加料uuid1,加料uuid2,加料uuid3。属性和加料的uuid要按从小到大排序
func (model *SaleOrderProduct) GenerateProductPackageSign() string {
	flavorUuid := model.GetFlavorBomUuid()
	attributeUuidList := model.GetAttributeUuidList()
	bomUuidList := model.GetBomUuidList()

	// 转换成字符串
	attributeUuidStringList := make([]string, 0)
	bomUuidStringList := make([]string, 0)
	for _, uuid := range attributeUuidList {
		attributeUuidStringList = append(attributeUuidStringList, strconv.FormatUint(uuid, 10))
	}
	for _, uuid := range bomUuidList {
		bomUuidStringList = append(bomUuidStringList, strconv.FormatUint(uuid, 10))
	}

	return utils.GenerateProductPackageSign(flavorUuid, attributeUuidStringList, bomUuidStringList)
}

// 获取商品选购详情
func (model *SaleOrderProduct) GetProductPackageDetail() resp.ProductPackageDetail {
	return resp.ProductPackageDetail{
		FlavorUuid:     model.GetFlavorBomUuid(),
		AttributesUuid: model.GetAttributeUuidList(),
		SaucesUuid:     model.GetBomUuidList(),
		Num:            model.Num, // 数量
	}
}

// 获取套餐的选购详情
func (model *SaleOrderProduct) GetPackageDetail() []resp.PackageSelectedInfo {
	packageSelectedInfoList := make([]resp.PackageSelectedInfo, 0)

	type SubProduct struct {
		FlavorUuid              uint64   `json:"flavor_uuid"`                // 商品规格uuid
		AttributeUuid           []uint64 `json:"attribute_uuid"`             // 属性uuid列表
		ProductPackageGroupUuid uint64   `json:"product_package_group_uuid"` // 套餐分组uuid
	}

	subProductList := make([]SubProduct, 0)
	utils.FromJson(model.PackageSubProductParams, &subProductList)
	for _, subProduct := range subProductList {
		packageSelectedInfoList = append(packageSelectedInfoList, resp.PackageSelectedInfo{
			ProductPackageGroupUuid: subProduct.ProductPackageGroupUuid,
			FlavorUuid:              subProduct.FlavorUuid,
			AttributeUuidList:       subProduct.AttributeUuid,
		})
	}
	return packageSelectedInfoList
}
