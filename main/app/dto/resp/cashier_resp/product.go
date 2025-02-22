package cashier_resp

import "ttpos-server-go/app/dto"

// Product 商品
type Product struct {
	Uuid            uint64                    `json:"uuid"`        // 商品UUID
	LocaleName      dto.LocaleResponse        `json:"locale_name"` // 商品名称
	Image           string                    `json:"image"`       // 商品图片
	Unit            dto.LocaleResponse        `json:"unit"`        // 商品单位
	Price           float64                   `json:"price"`       // 商品价格
	LimitNum        uint                      `json:"limit_num"`   // 商品限购数量
	Flavors         ProductFlavorList         `json:"flavors"`     // 商品规格
	Sauces          ProductSauceList          `json:"sauces"`      // 商品小料
	AttributeGroups ProductAttributeGroupList `json:"attributes"`  // 商品属性组
}

// ProductFlavor 商品规格
type ProductFlavor struct {
	Uuid              uint64             `json:"uuid"`                // 商品规格UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品规格名称
	Price             float64            `json:"price"`               // 商品规格价格
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
}

// ProductSauce 商品小料
type ProductSauce struct {
	Uuid              uint64             `json:"uuid"`                // 商品小料UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品小料名称
	Price             float64            `json:"price"`               // 商品小料价格
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
}

// ProductAttributeGroup 商品属性组
type ProductAttributeGroup struct {
	Uuid       uint64                    `json:"uuid"`        // 商品属性组UUID
	LocaleName dto.LocaleResponse        `json:"locale_name"` // 商品属性组名称
	Value      ProductAttributeValueList `json:"value"`       // 商品属性值
	IsMust     bool                      `json:"is_must"`     // 是否必选
	MaxSelect  uint                      `json:"max_select"`  // 最大可选数量
}

// ProductAttributeValue 商品属性值
type ProductAttributeValue struct {
	Uuid              uint64             `json:"uuid"`                // 商品属性UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品属性名称
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
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

// ProductCategory 商品类别
type ProductCategory struct {
	Uuid       uint64                  `json:"uuid"`        // 商品类别UUID
	LocaleName dto.LocaleResponse      `json:"locale_name"` // 商品类别名称
	ParentUuid uint64                  `json:"parent_uuid"` // 父级类别UUID
	IsSpecial  bool                    `json:"is_special"`  // 是否特殊类别
	Children   ProductCategoryListResp `json:"children"`    // 子级类别
}

// ProductCategoryListResp 商品类别列表响应
type ProductCategoryListResp struct {
	List []ProductCategory `json:"list"`
}
