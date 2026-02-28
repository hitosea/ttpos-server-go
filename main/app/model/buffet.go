package model

import "ttpos-server-go/app/dto/resp"

// BuffetPackage 自助餐套餐信息表 ttpos_buffet_package
type BuffetPackage struct {
	BaseModel
	Name                  string  `gorm:"default:'';column:name;comment:'自助餐名称'"`
	MultiLanguageNameUuid uint64  `gorm:"default:0;column:multi_language_name_uuid;comment:多语言名称ID"`
	Sort                  uint    `gorm:"default:0;column:sort;comment:排序顺序"`
	TaxUuid               uint64  `gorm:"default:0;column:tax_uuid;comment:税率ID"`
	IsLimitTime           uint    `gorm:"default:0;column:is_limit_time;comment:是否限时, 0-否、1-是"`
	LimitTime             uint    `gorm:"default:0;column:limit_time;comment:限时时间"`
	CanCombined           uint    `gorm:"default:0;comment:是否可组合, 0-否、1-是"`
	NonOrderingTime       uint    `gorm:"default:0;comment:不可下单时间（分钟）"`
	ReminderOrderTime     uint    `gorm:"default:0;column:reminder_order_time;comment:提醒下单时间（分钟）"`
	ActualSaleNum         float64 `gorm:"default:0;column:actual_sale_num;comment:实际销量"`
	OpenOverallDiscount   uint    `gorm:"default:1;column:open_overall_discount;comment:是否开启整单折扣, 0-否、1-是"`

	MultiLanguageName        MultiLanguageName         `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
	BuffetCustomerTypePrices []BuffetCustomerTypePrice `gorm:"foreignKey:buffet_package_uuid;references:uuid"`
	BuffetProducts           []BuffetProduct           `gorm:"foreignKey:buffet_package_uuid;references:uuid"`
	Tax                      *Tax                      `gorm:"foreignKey:tax_uuid;references:uuid"`
}

func (model *BuffetPackage) SetNil() {
	model.MultiLanguageName = MultiLanguageName{}
	model.BuffetCustomerTypePrices = nil
	model.BuffetProducts = nil
	model.Tax = nil
}

func (model *BuffetPackage) GetBuffetProductList() resp.BuffetProductList {
	buffetProductList := resp.BuffetProductList{}
	buffetProductList.List = make([]resp.BuffetProduct, 0)
	for _, buffetProduct := range model.BuffetProducts {
		buffetProductList.List = append(buffetProductList.List, resp.BuffetProduct{
			Uuid:            buffetProduct.ProductPackageUuid,
			Limit:           buffetProduct.Limit,
			Name:            buffetProduct.ProductPackage.Name,
			IsShowCashier:   buffetProduct.IsShowCashier,
			IsShowTablet:    buffetProduct.IsShowTablet,
			IsShowKitchen:   buffetProduct.IsShowKitchen,
			IsShowAssistant: buffetProduct.IsShowAssistant,
		})
	}
	return buffetProductList
}

// 获取自助餐税率。
func (model *BuffetPackage) GeTaxRate() float64 {
	if model.Tax == nil {
		// 当税率不存在时，判断系统是否开启收税，如果开启，则去查询商家第一个创建的税，用这个税作为该自助餐的税率

		return 0
	}
	return model.Tax.TaxRate
}

// GetMinPrice 获取最小价格. 如果只有一个价格, 则返回该价格, 否则返回最小价格
func (b *BuffetPackage) GetMinPrice() float64 {
	if len(b.BuffetCustomerTypePrices) == 1 {
		return b.BuffetCustomerTypePrices[0].Price
	}
	minPrice := float64(0)
	for index, buffetCustomerTypePrice := range b.BuffetCustomerTypePrices {
		if index == 0 {
			minPrice = buffetCustomerTypePrice.Price
			continue
		}
		if buffetCustomerTypePrice.Price < minPrice {
			minPrice = buffetCustomerTypePrice.Price
		}
	}
	return minPrice
}

// BuffetCustomerType 自助餐客户类型信息表 ttpos_buffet_customer_type
type BuffetCustomerType struct {
	BaseModel
	Name string `gorm:"default:'';column:name;comment:'自助餐客户类型名称'"`
}

// BuffetCustomerTypePrice 自助餐客户类型价格信息表 ttpos_buffet_customer_type_price
type BuffetCustomerTypePrice struct {
	BaseModel
	BuffetPackageUuid uint64  `gorm:"default:0;column:buffet_package_uuid;comment:自助餐套餐信息表ID"`
	CustomerTypeUuid  uint64  `gorm:"default:0;column:customer_type_uuid;comment:自助餐客户类型信息表ID"`
	Price             float64 `gorm:"column:price;default:0;comment:'价格'"`

	BuffetCustomerType BuffetCustomerType `gorm:"foreignKey:customer_type_uuid;references:uuid"`
}

func (model *BuffetCustomerTypePrice) SetNil() {
	model.BuffetCustomerType = BuffetCustomerType{}
}

// BuffetProduct 自助餐产品信息表 ttpos_buffet_product
type BuffetProduct struct {
	BaseModel
	ProductPackageUuid uint64 `gorm:"default:0;column:product_package_uuid;comment:产品包ID"`
	BuffetPackageUuid  uint64 `gorm:"default:0;column:buffet_package_uuid;comment:自助餐套餐信息表ID"`
	IsShowCashier      uint   `gorm:"default:0;column:is_show_cashier;comment:是否在收银台显示, 0-否、1-是"`
	IsShowTablet       uint   `gorm:"default:0;column:is_show_tablet;comment:是否在平板显示, 0-否、1-是"`
	IsShowKitchen      uint   `gorm:"default:0;column:is_show_kitchen;comment:是否在厨房显示, 0-否、1-是"`
	IsShowAssistant    uint   `gorm:"default:0;column:is_show_assistant;comment:是否在助手显示, 0-否、1-是"`
	Limit              uint   `gorm:"default:0;column:limit;comment:限购数量"`

	ProductPackage *ProductPackage `gorm:"foreignKey:product_package_uuid;references:uuid"`
}

func (model *BuffetProduct) SetNil() {
	model.ProductPackage = nil
}

// BuffetDelay 自助餐加钟价格表 `ttpos_buffet_delay`
type BuffetDelay struct {
	BaseModel
	Name      string  `gorm:"default:'';column:name;comment:'自助餐加钟价格名称'"`
	DelayTime int64   `gorm:"default:0;column:delay_time;comment:'加钟时间(分钟)'"`
	Price     float64 `gorm:"default:0;column:price;comment:'价格'"`
	Status    uint    `gorm:"default:0;column:status;comment:'状态 0-禁用 1-启用'"`
}
