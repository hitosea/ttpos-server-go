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
