package req

import "ttpos-server-go/app/dto"

// ProductListReq 商品列表查询
type ProductListReq struct {
	dto.PageReq // 分页参数
	// 以下字段不参与json序列化,内部方法使用
	RecommendProductPackageUuids []uint64 `json:"-"` // 推荐商品uuid列表
	IsMember                     bool     `json:"-"` // 是否是会员端查询商品列表
}

func (p *ProductListReq) ToPageReq() dto.PageReq {
	return dto.PageReq{
		PageNo:   p.PageNo,
		PageSize: p.PageSize,
	}
}

// ProductRecommendListReq 商品推荐列表查询
type ProductRecommendListReq struct {
}

// ProductSearchReq 商品搜索查询
type ProductSearchReq struct {
	Keyword  string `form:"keyword" json:"keyword" binding:"required"` // 搜索关键词
	IsMember bool   `json:"-"`                                         // 是否是会员端查询商品列表
}

// ProductUnitListReq 商品单位列表查询
type ProductUnitListReq struct {
	dto.PageReq // 分页参数
}

type ProductSauceListReq struct {
	dto.PageReq // 分页参数
}

type ProductSauceReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品加料UUID
}

type ProductUnitReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品单位UUID
}

type ProductUnitAddReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"` // 关联商品包UUID列表
}

type ProductUnitEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"` // 商品单位UUID
	LocaleName          dto.LocaleResponse `json:"locale_name"`             // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"`   // 关联商品包UUID列表
}

type ProductUnitSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品单位UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductUnitSortReq struct {
	List []ProductUnitSortItem `json:"list" binding:"required,dive"` // 商品单位排序列表
}

type ProductSauceAddReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品加料名称
	Price               float64            `json:"price"`                 // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"` // 关联商品包UUID列表
}

type ProductSauceEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"` // 商品加料UUID
	LocaleName          dto.LocaleResponse `json:"locale_name"`             // 商品加料名称
	Price               float64            `json:"price"`                   // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"`   // 关联商品包UUID列表
}

type ProductSauceSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品加料UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductSauceSortReq struct {
	List []ProductSauceSortItem `json:"list" binding:"required,dive"` // 商品加料排序列表
}

// ProductShopCategoryListReq 商品分类列表查询
type ProductShopCategoryListReq struct {
	Keyword    *string `form:"keyword"`     // 搜索关键词, 可选
	ParentUuid *uint64 `form:"parent_uuid"` // 父级分类UUID, 可选
	IsSpecial  *bool   `form:"is_special"`  // 是否特色分类, false-否 true-是, 可选
}

// ProductShopCategorySortItem 商品分类排序项
type ProductShopCategorySortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品分类UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

// ProductShopCategorySortReq 商品分类排序请求
type ProductShopCategorySortReq struct {
	List []ProductShopCategorySortItem `json:"list" binding:"required,dive"` // 商品分类排序列表
}

// ProductShopCategoryAddReq 商品分类添加请求
type ProductShopCategoryAddReq struct {
	IsSpecial  bool               `json:"is_special"`                     // 是否特殊分类, false-否 true-是
	ParentUuid uint64             `json:"parent_uuid"`                    // 父级分类UUID, 一级分类为0, 二级分类为一级分类的uuid
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品分类名称, 多语言
	Status     int                `json:"status"`                         // 商品分类状态 0-关闭 1-开启
}

// ProductShopCategoryEditReq 商品分类编辑请求
type ProductShopCategoryEditReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`        // 商品分类UUID
	ParentUuid uint64             `json:"parent_uuid"`                    // 父级分类UUID, 一级分类为0, 二级分类为一级分类的uuid
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品分类名称, 多语言
	Status     int                `json:"status"`                         // 商品分类状态 0-关闭 1-开启
}

// ProductShopCategoryReq 商品分类请求
type ProductShopCategoryReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品分类UUID
}

type ProductAttributeGroupListReq struct {
	dto.PageReq // 分页参数
}

type ProductAttributeGroupReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品属性分组UUID
}

type ProductAttributeGroupAddReq struct {
	LocaleName        dto.LocaleResponse                            `json:"locale_name" binding:"required"`             // 商品属性分组名称, 多语言
	ProductAttributes []ProductAttributeGroupAddProductAttributeReq `json:"product_attributes" binding:"required,dive"` // 商品属性
}

type ProductAttributeGroupAddProductAttributeReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品属性名称, 多语言
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductAttributeGroupEditReq struct {
	Uuid              uint64                                         `json:"uuid" binding:"required"`                    // 商品属性分组UUID
	LocaleName        dto.LocaleResponse                             `json:"locale_name" binding:"required"`             // 商品属性分组名称, 多语言
	ProductAttributes []ProductAttributeGroupEditProductAttributeReq `json:"product_attributes" binding:"required,dive"` // 商品属性
}

type ProductAttributeGroupEditProductAttributeReq struct {
	Uuid                uint64             `json:"uuid"`                           // 商品属性UUID, 可选，如果有，是编辑，没有是添加
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品属性名称, 多语言
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductAttributeGroupSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品属性分组UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductAttributeGroupSortReq struct {
	List []ProductAttributeGroupSortItem `json:"list" binding:"required,dive"` // 商品属性分组排序列表
}

type ProductAttributeSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品属性UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductAttributeSortReq struct {
	ProductAttributeGroupUuid uint64                     `json:"product_attribute_group_uuid" binding:"required"` // 商品属性分组UUID
	List                      []ProductAttributeSortItem `json:"list" binding:"required,dive"`                    // 商品属性排序列表
}
