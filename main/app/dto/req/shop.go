package req

import "ttpos-server-go/app/dto"

type AiTranslateRequest struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type CheckingName struct {
	Lang string `json:"lang" binding:"required"`
	Text string `json:"text" binding:"required"`
}

type CheckNameRequest struct {
	Uuid   uint64         `json:"uuid"`                          // 排除的UUID
	Source string         `json:"source" binding:"required"`     // 来源：unit-单位 product-商品 category-分类 sauce-加料 attribute-属性 attribute_group-属性组 flavor-规格
	Names  []CheckingName `json:"names" binding:"required,dive"` // 名称列表
}

type EditFreeOrGiftReasonItem struct {
	Uuid       uint64             `json:"uuid"`        // 如果有Uuid，则编辑，否则新增
	LocaleName dto.LocaleResponse `json:"locale_name"` // 名称列表
}

type EditFreeOrGiftReasonReq struct {
	List []EditFreeOrGiftReasonItem `json:"list" binding:"required,dive"` // 名称列表
}

type EditReturnFoodReasonItem struct {
	Uuid       uint64             `json:"uuid"`        // 如果有Uuid，则编辑，否则新增
	LocaleName dto.LocaleResponse `json:"locale_name"` // 名称列表
}

type EditReturnFoodReasonReq struct {
	List []EditReturnFoodReasonItem `json:"list" binding:"required,dive"` // 名称列表
}
