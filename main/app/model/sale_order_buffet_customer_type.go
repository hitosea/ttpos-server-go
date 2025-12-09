package model

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

// SaleOrderBuffetCustomerType 销售订单自助餐顾客类型 `ttpos_sale_order_buffet_customer_type`
type SaleOrderBuffetCustomerType struct {
	// 主键字段
	BaseModel
	Name string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'名称'"`
	// 价格信息
	Num                uint    `gorm:"column:num;type:int(11);default:0;comment:人数" json:"num"`
	SalePrice          float64 `gorm:"column:sale_price;type:decimal(12,2);not null;default:0;comment:原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变" json:"sale_price"`
	SalePriceNoTax     float64 `gorm:"column:sale_price_no_tax;type:decimal(12,2);not null;default:0.00;comment:'销售价,未含税价格（折前）'" json:"sale_price_no_tax"`
	CustomDiscountRate float64 `gorm:"column:custom_discount_rate;type:decimal(12,4);not null;default:1;comment:自定义折扣率(0-100%)" json:"custom_discount_rate"`
	TaxRate            float64 `gorm:"column:tax_rate;type:decimal(10,2);not null;default:0;comment:税率,单位%.加购时记录税率,结账时再重新核算" json:"tax_rate"`

	OpenOverallDiscount uint `gorm:"column:open_overall_discount;type:tinyint(1);not null;default:0;comment:'是否开启整单折扣, 0-否 1-是'" json:"open_overall_discount"`

	// 价格计算相关
	Price             float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:最终单价（折后价），只进行自定义打折，不进行会员打折" json:"price"`
	CustomDiscountFee float64 `gorm:"column:custom_discount_fee;type:decimal(12,2);not null;default:0;comment:自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*(1-自定义折扣率)" json:"custom_discount_fee"`
	ServiceTaxFee     float64 `gorm:"column:service_tax_fee;type:decimal(12,2);not null;default:0;comment:服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率" json:"service_tax_fee"`
	TaxFee            float64 `gorm:"column:tax_fee;type:decimal(12,2);not null;default:0;comment:自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率" json:"tax_fee"`
	ServiceFee        float64 `gorm:"column:service_fee;type:decimal(12,2);not null;default:0;comment:服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例" json:"service_fee"`
	TotalPrice        float64 `gorm:"column:total_price;type:decimal(12,2);not null;default:0;comment:应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费" json:"total_price"`
	OriginTotalPrice  float64 `gorm:"column:origin_total_price;type:decimal(12,2);not null;default:0;comment:原始应收金额(单人)。商品已含税时，应收金额(单人)=（原始单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=原始单价+服务费+总税费" json:"origin_total_price"`

	// 关联ID字段
	SaleOrderUuid               uint64 `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	SaleBillUuid                uint64 `gorm:"column:sale_bill_uuid;comment:销售账单ID" json:"sale_bill_uuid"`
	BuffetPackageUuid           uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
	BuffetPackageName           string `gorm:"column:buffet_package_name;type:text" json:"buffet_package_name"` // 新增快照字段（JSON）
	BuffetCustomerTypePriceUuid uint64 `gorm:"column:buffet_customer_type_price_uuid;comment:顾客类型定价ID" json:"buffet_customer_type_price_uuid"`

	// 关联字段
	BuffetPackage           BuffetPackage           `gorm:"foreignKey:BuffetPackageUuid;references:uuid"`
	BuffetCustomerTypePrice BuffetCustomerTypePrice `gorm:"foreignKey:BuffetCustomerTypePriceUuid;references:uuid"` // 用于关联后台设置的顾客类型定价。在结账时，判断价格是否改变
	ReturnOrderProducts     []*ReturnOrderProduct   `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid"`        // 退货单商品表，用于获取该顾客商品的退货数量
}

// 获取整单折扣率
func (model *SaleOrderBuffetCustomerType) GetCustomDiscountRate() float64 {
	if model.OpenOverallDiscount == 0 {
		return constant.NoDiscount
	}
	return model.CustomDiscountRate
}

// 商品未含税价格（折后）
func (model *SaleOrderBuffetCustomerType) GetFinalSalePriceNoneTax() float64 {
	// 商品是否含税
	isTax := true                                // 默认商品含税
	if model.SalePriceNoTax == model.SalePrice { // 折前价格未含税价格和折前价格相同时，说明商品未含税
		isTax = false
	}
	if isTax {
		// 商品含税，则返回折后价格-税费
		return decimal.NewFromFloat(model.Price).Sub(decimal.NewFromFloat(model.TaxFee)).Truncate(3).Round(2).InexactFloat64()
	} else {
		// 商品未含税，则返回折后价格
		return model.Price
	}
}

// 设置为空。为了更新数据库数据时，不更新关联对象
func (model *SaleOrderBuffetCustomerType) SetNil() {
	model.BuffetPackage = BuffetPackage{}
	model.BuffetCustomerTypePrice = BuffetCustomerTypePrice{}
}

// GetReturnPrice 获取销售订单商品的已退金额。订单商品金额 - 可退货金额
func (model *SaleOrderBuffetCustomerType) GetReturnPrice() float64 {
	if model.GetReturnNum() == 0 {
		return 0
	}
	returnPrice := decimal.NewFromFloat(model.TotalPrice).Mul(decimal.NewFromUint64(uint64(model.GetReturnNum()))).Truncate(3).Round(2).InexactFloat64()
	return returnPrice
}

// 获取销售订单自助餐顾客类型的总税费。包含服务费税费
func (model *SaleOrderBuffetCustomerType) GetTotalTaxFee() float64 {
	return model.GetTaxFee() + model.GetServiceTaxFee()
}

// 获取销售订单自助餐顾客类型的税费。税费=销售订单自助餐顾客类型的税费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetTaxFee() float64 {
	return decimal.NewFromFloat(model.TaxFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
}

// 获取销售订单自助餐顾客类型的原始税费。税费=销售订单自助餐顾客类型的税费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetOriginTaxFee(taxFeeType int) float64 {
	taxFee := model.calcTaxFee(model.SalePrice, model.TaxRate, taxFeeType) // 税费（折前）
	totalTaxFee := decimal.NewFromFloat(taxFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
	return totalTaxFee
}

// 获取销售订单自助餐顾客类型的服务费税费。服务费税费=销售订单自助餐顾客类型的服务费税费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetServiceTaxFee() float64 {
	return decimal.NewFromFloat(model.ServiceTaxFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
}

// 获取销售订单自助餐顾客类型的服务费税费。服务费税费=销售订单自助餐顾客类型的服务费税费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetOriginServiceTaxFee(serviceFeeRate float64, taxFeeType int, serviceFeeType int) float64 {
	serviceTaxFee := model.calcServiceTaxFee(model.SalePrice, serviceFeeRate, taxFeeType, serviceFeeType) // 服务费（折前）
	return decimal.NewFromFloat(serviceTaxFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
}

// 获取销售订单自助餐顾客类型的服务费。服务费=销售订单自助餐顾客类型的服务费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetServiceFee() float64 {
	return decimal.NewFromFloat(model.ServiceFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
}

// 获取销售订单自助餐顾客类型的原始服务费。服务费=销售订单自助餐顾客类型的服务费*销售订单自助餐顾客类型的数量
func (model *SaleOrderBuffetCustomerType) GetOriginServiceFee(serviceFeeRate float64, taxFeeType int) float64 {
	serviceFee := model.calcServiceFee(model.SalePrice, serviceFeeRate, taxFeeType) // 服务费（折前）
	return decimal.NewFromFloat(serviceFee).Mul(decimal.NewFromUint64(uint64(model.Num))).InexactFloat64()
}

// GetCanReturnNum 获取销售订单自助餐顾客类型的可退货数量. 可退货数量=订单自助餐顾客类型数量-已退货数量
func (model *SaleOrderBuffetCustomerType) GetCanReturnNum() uint {
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

// GetReturnNum 获取销售订单自助餐顾客类型的已退货数量. 已退货数量=订单自助餐顾客类型数量-可退货数量
func (model *SaleOrderBuffetCustomerType) GetReturnNum() uint {
	return model.Num - model.GetCanReturnNum()
}

// GetCanReturnPrice 获取销售订单自助餐顾客类型的可退货金额. 可退货金额=订单自助餐顾客类型金额-已退货金额
func (model *SaleOrderBuffetCustomerType) GetCanReturnPrice() float64 {
	return decimal.NewFromFloat(model.TotalPrice).Mul(decimal.NewFromUint64(uint64(model.GetCanReturnNum()))).Round(2).InexactFloat64()
}

func (model *SaleOrderBuffetCustomerType) GetSign() string {
	return fmt.Sprintf("%d-%d", model.BuffetPackageUuid, model.BuffetCustomerTypePriceUuid)
}

func (model *SaleOrderBuffetCustomerType) CopyBuffetCustomer(saleOrderUuid uint64) *SaleOrderBuffetCustomerType {
	customerType := SaleOrderBuffetCustomerType{}
	err := copier.Copy(&customerType, model)
	if err != nil {
		return nil
	}
	// 重置base_model
	customerType.BaseModel = BaseModel{}
	customerType.SetNil()
	// 指定目标销售订单。如果不移动到别的销售订单可以修改销售订单uuid
	if saleOrderUuid != 0 {
		customerType.SaleOrderUuid = saleOrderUuid
	}
	return &customerType
}

// 获取顾客原价. = 原价*人数
func (model *SaleOrderBuffetCustomerType) GetOriginPrice() float64 {
	price := decimal.NewFromFloat(model.SalePrice).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// 获取顾客折后价 = 折扣价*人数
func (model *SaleOrderBuffetCustomerType) GetDiscountPrice() float64 {
	price := decimal.NewFromFloat(model.Price).Mul(decimal.NewFromFloat(float64(model.Num))).Round(2).InexactFloat64()
	return price
}

// 获取顾客折后价(含VAT) = 应收金额(单人、折后)*人数
// 获取顾客折前价(含VAT) = 应收金额(单人、折前)*人数
func (model *SaleOrderBuffetCustomerType) GetDiscountPriceWithVAT(options ...func(option *CalcOption)) float64 {
	option := &CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}
	totalPrice := model.TotalPrice
	if option.IsOriginPrice {
		// 折前价. = 顾客原价+服务费+税费
		totalPrice = model.OriginTotalPrice
		// 兼容没有OriginTotalPrice字段时的情况

	}
	price := decimal.NewFromFloat(totalPrice).Mul(decimal.NewFromFloat(float64(model.Num))).Truncate(3).Round(2).InexactFloat64()
	return price
}

// // 获取顾客折后价(含VAT) = 应收金额(单人、折后)*人数
// func (model *SaleOrderBuffetCustomerType) GetDiscountPriceWithVAT() float64 {
// 	price := decimal.NewFromFloat(model.TotalPrice).Mul(decimal.NewFromFloat(float64(model.Num))).Truncate(3).Round(2).InexactFloat64()
// 	return price
// }

// GetLocaleBuffetPackageName 获取自助餐套餐名称（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
func (model *SaleOrderBuffetCustomerType) GetLocaleBuffetPackageName() dto.LocaleResponse {
	// 优先使用快照字段
	snapshotName := model.BuffetPackageName

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
	if !model.BuffetPackage.MultiLanguageName.IsNullName() {
		return model.BuffetPackage.MultiLanguageName.GetNames()
	}

	// 兜底：如果关联表也没有数据，返回空的多语言响应
	return dto.LocaleResponse{}
}

// SetBuffetPackageNameSnapshot 设置自助餐套餐名称快照（JSON）
// 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
// Requirement: story-main-buffet-customer-type-package-name-snapshot-fix (JSON 方案)
func (model *SaleOrderBuffetCustomerType) SetBuffetPackageNameSnapshot(multiLangName MultiLanguageName) error {
	// 如果多语言名称为空，设置为空字符串
	if multiLangName.IsNullName() {
		model.BuffetPackageName = ""
		return nil
	}

	// 构建 LocaleResponse
	localeResp := multiLangName.GetNames()

	// 序列化为 JSON
	jsonData, err := json.Marshal(localeResp)
	if err != nil {
		return err
	}

	model.BuffetPackageName = string(jsonData)
	return nil
}
