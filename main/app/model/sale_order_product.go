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
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

// SaleOrderProduct 销售订单产品 `ttpos_sale_order_product`
type SaleOrderProduct struct {
	// 基础字段
	BaseModel

	// 基本信息字段
	Name       string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'商品名称'" json:"name"`
	FlavorName string `gorm:"column:flavor_name;type:varchar(255);not null;default:'';comment:'规格名称'" json:"flavor_name"`
	Num        uint   `gorm:"column:num;type:int(11);not null;default:0;comment:'商品数量。不能减为0，当数量为1再减时，标记删除'" json:"num"`
	Remark     string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:'备注，顾客对商品的备注信息'" json:"remark"`
	IsBuffet   uint   `gorm:"column:is_buffet;type:tinyint(1);not null;default:0;comment:'是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0'" json:"is_buffet"`
	// 状态相关字段
	Status        uint `gorm:"column:status;type:tinyint(1);not null;default:0;comment:'状态, 0-未送厨 1-已送厨'" json:"status"`
	IsRequire     uint `gorm:"column:is_require;type:tinyint(1);not null;default:0;comment:'是否必点商品 0-否 1-是。用于在前端显示必点图标'" json:"is_require"`
	IsAcceptOrder uint `gorm:"column:is_accept_order;type:tinyint(1);not null;default:0;comment:'是否已接单, 0-否 1-是'" json:"is_accept_order"`

	// 价格相关字段
	FlavorPrice  float64 `gorm:"column:flavor_price;type:decimal(12,2);not null;default:0.00;comment:'规格原价（单商品）,仅某规格商品的原价'" json:"flavor_price"`
	SaucePrice   float64 `gorm:"column:sauce_price;type:decimal(12,2);not null;default:0.00;comment:'小料价（单商品）,所有小料的价格之和'" json:"sauce_price"`
	ProductPrice float64 `gorm:"column:product_price;type:decimal(12,2);not null;default:0.00;comment:'原始单价（单商品）,规格原价+小料价'" json:"product_price"`
	SalePrice    float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0.00;comment:'销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价'" json:"sale_price"`
	Price        float64 `gorm:"column:price;type:decimal(12,2);not null;default:0.00;comment:'最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率'" json:"price"`
	TotalPrice   float64 `gorm:"column:total_price;type:decimal(12,2);not null;default:0.00;comment:'应收金额(单商品)=最终单价+服务费+总税费'" json:"total_price"`

	// 折扣相关字段
	ChangePriceTime        int64   `gorm:"column:change_price_time;type:int(10);not null;default:0;comment:'改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并'" json:"change_price_time"`
	OpenMemberDiscount     uint    `gorm:"column:open_member_discount;type:tinyint(1);not null;default:0;comment:'是否开启会员折扣, 0-否 1-是'" json:"open_member_discount"`              // 快照设置相关，不受后台改变，结账时检查
	MemberDiscountRate     float64 `gorm:"column:member_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员折扣率(0-100%)'" json:"member_discount_rate"`            // 与sale_order的member_discount_rate一致
	MemberCardDiscountRate float64 `gorm:"column:member_card_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'会员卡折扣率(0-100%)'" json:"member_card_discount_rate"` // 与sale_order的member_card_discount_rate一致
	CustomDiscountRate     float64 `gorm:"column:custom_discount_rate;type:decimal(12,2);not null;default:0.00;comment:'自定义折扣率(0-100%)'" json:"custom_discount_rate"`           // 与sale_order的custom_discount_rate一致

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
	DeductStockType uint  `gorm:"column:deduct_stock_type;type:tinyint(1);not null;default:0;comment:'库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数'" json:"deduct_stock_type"`
	DeductStockTime int64 `gorm:"column:deduct_stock_time;type:int(10);not null;default:0;comment:'减库存的时间(时间戳），0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存'" json:"deduct_stock_time"`

	// 赠品相关字段
	GiftTime     int64  `gorm:"column:gift_time;type:int(10);not null;default:0;comment:'赠菜时间(时间戳),用于判断不同时间赠送的商品不合并'" json:"gift_time"`
	CancelTime   int64  `gorm:"column:cancel_time;type:int(10);not null;default:0;comment:'退菜时间(时间戳),用于判断不同时间退菜的商品不合并'" json:"cancel_time"`
	GiftReason   string `gorm:"column:gift_reason;type:varchar(255);not null;default:'';comment:'赠菜原因'" json:"gift_reason"`
	CancelReason string `gorm:"column:cancel_reason;type:varchar(255);not null;default:'';comment:'退菜原因'" json:"refund_reason"`

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
	Sign             string `gorm:"column:sign;type:varchar(255);not null;default:'';comment:'商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品'" json:"sign"`
	IsH5OrderProduct uint   `gorm:"column:is_h5_order_product;type:tinyint(1);not null;default:0;comment:'是否为扫码订单商品, 0-否 1-是'" json:"is_qrcode_order_product"`

	// 扫码订单相关
	H5OrderProductUuid uint64 `gorm:"column:h5_order_product_uuid;type:bigint(20) unsigned;default:0;comment:h5订单商品ID，用于关联h5订单商品，用于判断是否为h5订单商品;NOT NULL" json:"h5_order_product_uuid"`
	H5OrderUuid        uint64 `gorm:"column:h5_order_uuid;type:bigint(20) unsigned;default:0;comment:扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品;NOT NULL" json:"h5_order_uuid"`

	// 关联对象
	MultiLanguageName          *MultiLanguageName           `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	ImageFile                  *File                        `gorm:"foreignKey:image_file_uuid;references:uuid"`
	SaleOrderProductBoms       []*SaleOrderProductBom       `gorm:"foreignKey:sale_order_product_uuid;references:uuid"`
	SaleOrderProductAttributes []*SaleOrderProductAttribute `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ReturnOrderProducts        []*ReturnOrderProduct        `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
	ProductPackage             *ProductPackage              `gorm:"foreignKey:ProductPackageUuid;references:Uuid"`
	SaleBill                   *SaleBill                    `gorm:"foreignKey:SaleBillUuid;references:uuid"`
	CancelReasons              []*SaleOrderProductReason    `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`
}

// 是否为已送厨的商品
func (model *SaleOrderProduct) IsCookingProduct() bool {
	// 状态为已送厨且生产订单ID不为0
	return model.Status == constant.SaleOrderProductStatusCooking && model.ProductionOrderUuid != 0
}

// 是否为未送厨的商品
func (model *SaleOrderProduct) IsUnCookingProduct() bool {
	return !model.IsCookingProduct()
}

// 设置该商品为自助餐商品
func (model *SaleOrderProduct) SetIsBuffet() {
	model.IsBuffet = constant.SaleOrderProductIsBuffetYes
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
	model.Sign = model.GenerateProductSign() // 更新签名
	model.SetUpdate()                        // 标记该model需要更新
}

// 产品是否已经下架
func (model *SaleOrderProduct) CheckProduct() (int, string) {
	// 检查商品是否删除、下架、库存是否充足、价格变动
	for _, bom := range model.SaleOrderProductBoms {
		if bom.ProductBom.IsFlavorProduct() {
			// 商品已经沽清
			if bom.ProductBom.IsSoldOutStatus() {
				return constant.CodeOrderCheckProductStockZero, "商品已经沽清"
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
			// 下单商品数量超过库存数量
			if bom.ProductBom.IsStockShortage(model.Num) {
				return constant.CodeOrderCheckProductStockZero, "下单商品数量超过库存数量"
			}
			// 商品规格价格变动
			if bom.ProductBom.IsPriceChanged(model.FlavorPrice) {
				return constant.CodeOrderCheckProductPriceChanged, "商品规格价格变动"
			}
		}
		if bom.ProductBom.IsSauce() {
			// 商品已经沽清
			if bom.ProductBom.IsSoldOutStatus() {
				return constant.CodeOrderCheckProductStockZero, "小料已经售罄"
			}
			// 商品已经下架
			if bom.ProductBom.IsDown() {
				return constant.CodeOrderCheckProductFlavorDown, "小料已经下架"
			}
			// 商品已经删除
			if bom.ProductBom.IsDelete() {
				return constant.CodeOrderCheckProductFlavorDown, "小料已经删除"
			}
			// 下单商品数量超过库存数量
			// 每个订单商品一个小料只消耗一个小料库存
			if bom.ProductBom.IsStockShortage(1) {
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

// 小料的价格是否有变动
func (model *SaleOrderProduct) saucePriceChanged(saucePrice float64) bool {
	price := decimal.NewFromFloat(0)
	for _, bom := range model.SaleOrderProductBoms {
		if bom.ProductBom.IsSauce() {
			price.Add(decimal.NewFromFloat(bom.ProductBom.Price))
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
func (model *SaleOrderProduct) SetCancelInfo(reasons []*SaleOrderProductReason) {
	defer model.SetUpdate() // 标记该model需要更新
	model.CancelTime = time.Now().Unix()
	model.CancelReasons = append(model.CancelReasons, reasons...)
	model.Sign = model.GenerateProductSign() // 更新签名
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
func (model *SaleOrderProduct) SetNum(num uint) {
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

func (model *SaleOrderProduct) IsAcceptOrderBool() bool {
	return model.IsAcceptOrder == constant.OrderProductIsAcceptOrderAccepted
}

func (model *SaleOrderProduct) IsGiftBool() bool {
	return model.GiftTime > 0
}
func (model *SaleOrderProduct) IsCancelBool() bool {
	return model.CancelTime > 0
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
	salePrice := decimal.NewFromFloat(model.SalePrice).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return salePrice
}

// 获取最终价格（折后价）
func (model *SaleOrderProduct) GetPrice() float64 {
	// 最终价格*数量
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
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

// 是否取消商品
func (model *SaleOrderProduct) IsCancelProduct() bool {
	return model.CancelTime > 0
}

// 是否送到厨房
func (model *SaleOrderProduct) IsSendKitchen() bool {
	return model.Status == constant.OrderProductStatusSentKitchen
}

// 判断商品是否有打折
func (model *SaleOrderProduct) IsDiscount() bool {
	return model.Price != model.SalePrice // 折前价格不等于折后价格时，说明有折扣
}

// 判断是哪个业务状态
func (model *SaleOrderProduct) StatusValue() int {
	return int(model.Status)
}

// 获取该订单商品的材料组成及用量。
// 如一个珍珠奶茶加料珍珠，则计算成分珍珠、奶、茶等各个原材料等用量
func (model *SaleOrderProduct) GetMaterialBom() []*ProductionOrderMaterial {
	return nil // todo
}

func (model *SaleOrderProduct) GetAttributeName() dto.LocaleResponse {
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

	return getLocaleResponse(nameList, ";")
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
		if attributeResultNames.ZH != "" && index != len(nameList)-1 {
			attributeResultNames.ZH += div
			attributeResultNames.TH += div
			attributeResultNames.EN += div
			attributeResultNames.ZHTW += div
			attributeResultNames.JA += div
			attributeResultNames.KO += div
			attributeResultNames.MY += div
			attributeResultNames.TR += div
		}
	}
	return attributeResultNames
}

// 点餐时录入的原始数据
type DefaultSaleOrderProduct struct {
	Name                   string
	OpenMemberDiscount     uint
	TaxRate                float64
	DeductStockType        uint
	MultiLanguageNameUuid  uint64
	ImageFileUuid          uint64
	ProductPackageUuid     uint64
	SaleBillUuid           uint64
	SaleOrderUuid          uint64
	MemberDiscountRate     float64
	MemberCardDiscountRate float64
	CustomDiscountRate     float64
	Sauces                 []Sauce
	Flavor                 Flavor
	Attribute              []Attribute
}

func NewDefaultSaleOrderProduct(def DefaultSaleOrderProduct) *SaleOrderProduct {
	saleOrderProductUuid, _ := utils.GetID()
	saleOrderProductBoms := make([]*SaleOrderProductBom, 0)
	for _, bom := range def.Sauces {
		saleOrderProductBom := &SaleOrderProductBom{
			Name:                 bom.Name,
			Price:                bom.Price,
			IsFlavorBom:          0,
			SaleOrderProductUuid: saleOrderProductUuid,
			SaleOrderUuid:        def.SaleOrderUuid,
			ProductBomUuid:       bom.ProductBomUuid,
		}
		saleOrderProductBoms = append(saleOrderProductBoms, saleOrderProductBom)
	}
	saleOrderProductBoms = append(saleOrderProductBoms, &SaleOrderProductBom{
		Name:                 def.Flavor.Name,
		Price:                def.Flavor.Price,
		IsFlavorBom:          1,
		SaleOrderUuid:        def.SaleOrderUuid,
		SaleOrderProductUuid: saleOrderProductUuid,
		ProductBomUuid:       def.Flavor.ProductBomUuid,
	})

	saleOrderProductAttributes := []*SaleOrderProductAttribute{}
	for _, attribute := range def.Attribute {
		saleOrderProductAttribute := &SaleOrderProductAttribute{
			Name:                 attribute.Name,
			SaleOrderUuid:        def.SaleOrderUuid,
			SaleOrderProductUuid: saleOrderProductUuid,
			ProductAttributeUuid: attribute.ProductAttributeUuid,
		}
		saleOrderProductAttributes = append(saleOrderProductAttributes, saleOrderProductAttribute)
	}
	product := SaleOrderProduct{
		BaseModel: BaseModel{
			Uuid: saleOrderProductUuid,
		},
		Name:                       def.Name,
		FlavorName:                 def.Flavor.Name,
		Num:                        1,
		Status:                     constant.OrderProductStatusUnSending,
		IsAcceptOrder:              1,
		FlavorPrice:                def.Flavor.Price,
		OpenMemberDiscount:         def.OpenMemberDiscount,
		MemberDiscountRate:         def.MemberDiscountRate,
		MemberCardDiscountRate:     def.MemberCardDiscountRate,
		CustomDiscountRate:         def.CustomDiscountRate,
		TaxRate:                    def.TaxRate,
		DeductStockType:            def.DeductStockType,
		MultiLanguageNameUuid:      def.MultiLanguageNameUuid,
		ImageFileUuid:              def.ImageFileUuid,
		ProductPackageUuid:         def.ProductPackageUuid,
		SaleBillUuid:               def.SaleBillUuid,
		SaleOrderUuid:              def.SaleOrderUuid,
		SaleOrderProductBoms:       saleOrderProductBoms,
		SaleOrderProductAttributes: saleOrderProductAttributes,
	}
	return &product
}

// 判断商品有没有改价
func (model *SaleOrderProduct) IsCustomPriceBool() bool {
	return model.ChangePriceTime > 0
}

// GenerateProductSign 生成商品包签名. 相同的商品，商品签名相同,用于取消拆单时合并商品。
// 格式：物料,物料,物料-属性,属性,属性-备注内容-送厨批次-改价时间-赠菜时间-退菜时间
// 更新签名的场景：
// 1 改价销售订单商品价格后要重新生成签名
// 2 修改备注
// 3 送厨
// 4 赠菜
// 5 退菜
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
	// 物料ID列表和属性ID列表拼接。格式：物料,物料,物料-属性,属性,属性-备注内容-必点方案-送厨批次-改价时间-赠菜时间-退菜时间
	bomIdListStr := strings.Join(bomIdList, ",")
	attributeIdListStr := strings.Join(attributeIdList, ",")
	return fmt.Sprintf("%s-%s-%s-%d-%d-%d-%d-%d",
		bomIdListStr,
		attributeIdListStr,
		model.Remark,
		model.MustPlanUuid,
		model.ProductionOrderUuid,
		model.ChangePriceTime,
		model.GiftTime,
		model.CancelTime)
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
