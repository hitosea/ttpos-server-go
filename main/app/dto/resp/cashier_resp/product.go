package cashier_resp

import "ttpos-server-go/app/dto"

// Product 商品
type Product struct {
	Uuid       uint                      `json:"uuid"`       // 商品UUID
	Name       dto.LocaleResponse        `json:"name"`       // 商品名称
	Image      string                    `json:"image"`      // 商品图片
	Unit       dto.LocaleResponse        `json:"unit"`       // 商品单位
	Price      float64                   `json:"price"`      // 商品价格
	Flavors    ProductFlavorList         `json:"flavors"`    // 商品规格
	Sauces     ProductSauceList          `json:"sauces"`     // 商品小料
	Attributes ProductAttributeGroupList `json:"attributes"` // 商品属性组
}

// ProductFlavor 商品规格
type ProductFlavor struct {
	Uuid  uint               `json:"uuid"`  // 商品规格UUID
	Name  dto.LocaleResponse `json:"name"`  // 商品规格名称
	Price float64            `json:"price"` // 商品规格价格
}

// ProductSauce 商品小料
type ProductSauce struct {
	Uuid  uint               `json:"uuid"`  // 商品小料UUID
	Name  dto.LocaleResponse `json:"name"`  // 商品小料名称
	Price float64            `json:"price"` // 商品小料价格
}

// ProductAttributeGroup 商品属性组
type ProductAttributeGroup struct {
	Uuid  uint                      `json:"uuid"`  // 商品属性组UUID
	Name  dto.LocaleResponse        `json:"name"`  // 商品属性组名称
	Value ProductAttributeValueList `json:"value"` // 商品属性值
}

// ProductAttributeValue 商品属性值
type ProductAttributeValue struct {
	Uuid uint               `json:"uuid"` // 商品属性UUID
	Name dto.LocaleResponse `json:"name"` // 商品属性名称
}

// ProductFlavorList 商品规格列表
type ProductFlavorList struct {
	List []ProductFlavor `json:"list"`
}

// ProductSauceList 商品小料列表
type ProductSauceList struct {
	List []ProductSauce `json:"list"`
}

// ProductAttributeGroupList 商品属性组列表
type ProductAttributeGroupList struct {
	List []ProductAttributeGroup `json:"list"`
}

// ProductAttributeValueList 商品属性值列表
type ProductAttributeValueList struct {
	List []ProductAttributeValue `json:"list"`
}

// ProductListWithPaginationResp 商品列表响应
type ProductListWithPaginationResp struct {
	List []Product        `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}
