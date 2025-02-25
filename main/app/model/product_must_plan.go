package model

import (
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
)

// 商品必选商品计划表 `ttpos_product_must_plan`
type ProductMustPlan struct {
	BaseModel
	Name         string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'方案名称'"`
	UseChannel   string `gorm:"column:use_channel;type:varchar(255);not null;default:'';comment:'使用渠道 10-点餐方式 20-桌台方式'"`
	MustType     uint   `gorm:"column:must_type;default:0;comment:'必点类型 0-每笔订单必点1份 1-每人必点1份 '"`
	MustRule     uint   `gorm:"column:must_rule;default:1;comment:'必点规则 0-固定商品 1-可选商品'"`
	Status       uint   `gorm:"column:status;default:1;comment:'状态,1-开启 0-关闭'"`
	AutoCart     uint   `gorm:"column:auto_cart;default:1;comment:'自动加入购物车 1-是 0-否'"`
	AutoChange   uint   `gorm:"column:auto_change;default:1;comment:'顾客可修改必点数量 1-是 0-否'"`
	AutoCheck    uint   `gorm:"column:auto_check;default:1;comment:'下单时检查必点商品 1-是 0-否'"`
	AutoCheckout uint   `gorm:"column:auto_checkout;default:1;comment:'结账时检查必点商品 1-是 0-否'"`

	ProductMustPlanItems []ProductMustPlanItem `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
}

func (model *ProductMustPlan) GetMustType() int {
	switch model.MustType {
	// 每单必点
	case constant.ProductMustPlanMustTypeEachOrder:
		return constant.ProductMustPlanMustTypeEachOrder
	//	每人必点
	case constant.ProductMustPlanMustTypeEachPerson:
		return constant.ProductMustPlanMustTypeEachPerson
	//	默认为每单必点
	default:
		return constant.ProductMustPlanMustTypeEachOrder
	}
}

func (model *ProductMustPlan) GetMustRule() int {
	switch model.MustRule {
	// 固定商品
	case constant.ProductMustPlanMustRuleAll:
		return constant.ProductMustPlanMustRuleAll
	//	可选商品
	case constant.ProductMustPlanMustRuleAny:
		return constant.ProductMustPlanMustRuleAny
	//	默认为固定商品
	default:
		return constant.ProductMustPlanMustRuleAll
	}
}

func (model *ProductMustPlan) IsAutoCart() bool {
	return model.AutoCart == constant.ProductMustPlanAutoCartOn
}

func (model *ProductMustPlan) IsCustomerCanChange() bool {
	return model.AutoChange == constant.ProductMustPlanCustomerCanChangeOn
}

func (model *ProductMustPlan) IsDeskMustPlan() bool {
	return strings.Contains(model.UseChannel, constant.ProductMustPlanUseChannelDesk)
}

// 商品必选商品计划区域表 `ttpos_product_must_plan_region`
type ProductMustPlanRegion struct {
	BaseModel
	ProductMustPlanUuid uint64 `gorm:"column:product_must_plan_uuid;type:bigint unsigned;not null;default:0;comment:'商品必选商品计划ID'"`
	DeskRegionUuid      uint64 `gorm:"column:desk_region_uuid;type:bigint unsigned;not null;default:0;comment:'桌台区域ID'"`

	ProductMustPlan ProductMustPlan `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
}

// 商品必点计划商品表 `ttpos_product_must_plan_item`
type ProductMustPlanItem struct {
	BaseModel
	ProductMustPlanUuid uint64 `gorm:"column:product_must_plan_uuid;type:bigint unsigned;not null;default:0;comment:'商品必选商品计划ID'"`
	ProductPackageUuid  uint64 `gorm:"column:product_package_uuid;type:bigint unsigned;not null;default:0;comment:'商品包ID'"`

	ProductMustPlan ProductMustPlan `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
	ProductPackage  ProductPackage  `gorm:"foreignKey:ProductPackageUuid;references:Uuid"`
}

func (model *ProductMustPlanItem) GetProductInfo() *resp.InstantMustPlanProduct {
	if model.IsDelete() {
		return nil
	}

	flavorList := make([]resp.ProductFlavor, 0)                        // 商品的规格列表
	sauceList := make([]resp.ProductSauce, 0)                          // 商品的小料列表
	productAttributeGroupList := make([]resp.ProductAttributeGroup, 0) // 商品的属性组列表

	// 组建商品的规格列表和商品的小料列表
	for _, productBom := range model.ProductPackage.ProductBoms {
		if productBom.IsDelete() {
			continue
		}
		if productBom.IsFlavorProduct() {
			flavor := resp.ProductFlavor{
				Uuid:       productBom.Uuid,
				LocaleName: productBom.ProductFlavor.MultiLanguageName.GetNames(),
				Price:      productBom.Price,
				StockNum:   int(productBom.StockNum),
			}
			flavorList = append(flavorList, flavor)
		}
		if !productBom.IsFlavorProduct() {
			sauce := resp.ProductSauce{
				Uuid:              productBom.Uuid,
				LocaleName:        productBom.ProductSauce.MultiLanguageName.GetNames(),
				Price:             productBom.Price,
				IsDefaultSelected: productBom.IsDefaultSelectBool(),
				StockNum:          int(productBom.StockNum),
			}
			sauceList = append(sauceList, sauce)
		}
	}

	// 组建商品的属性组列表
	for _, group := range model.ProductPackage.ProductPackageAttributeGroups {
		if group.IsDelete() {
			continue
		}
		productAttributeList := make([]resp.Attribute, 0)
		for _, attribute := range group.ProductPackageAttributes {
			if attribute.IsDelete() {
				continue
			}
			productAttribute := resp.Attribute{
				Uuid:              attribute.Uuid,
				LocaleName:        attribute.Attribute.MultiLanguageName.GetNames(),
				IsDefaultSelected: attribute.IsDefaultSelectedBool(),
			}
			productAttributeList = append(productAttributeList, productAttribute)
		}
		productAttributeGroup := resp.ProductAttributeGroup{
			LocaleName: group.ProductAttributeGroup.MultiLanguageName.GetNames(),
			Attributes: resp.Attributes{List: productAttributeList},
			IsMust:     group.IsMustBool(),
			MaxSelect:  group.MaxSelection,
		}
		productAttributeGroupList = append(productAttributeGroupList, productAttributeGroup)
	}

	productPackage := resp.InstantMustPlanProduct{
		Uuid:       model.ProductPackage.Uuid,
		LocaleName: model.ProductPackage.MultiLanguageName.GetNames(),
		Image:      model.ProductPackage.ImageFile.GetUrl(),
		Unit:       model.ProductPackage.ProductUnit.MultiLanguageName.GetNames(),
		LimitNum:   model.ProductPackage.LimitNum,
		Price:      model.ProductPackage.GetMinPrice(),
		Flavors:    resp.Flavors{List: flavorList},
		Sauces: resp.ProductSauces{
			List:      sauceList,
			IsMust:    model.ProductPackage.GetSauceRequired(),
			MaxSelect: int(model.ProductPackage.SauceMaxSelection),
		},
		AttributeGroups: resp.AttributeGroups{List: productAttributeGroupList},
	}

	return &productPackage
}
