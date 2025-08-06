package req

type AiTranslateRequest struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type CheckingName struct {
	Lang string `json:"lang" binding:"required"`
	Text string `json:"text" binding:"required"`
}

type CheckNameRequest struct {
	Source string         `json:"source" binding:"required"`     // 来源：unit-单位 product-商品 category-分类 sauce-加料 attribute-属性 attribute_group-属性组
	Names  []CheckingName `json:"names" binding:"required,dive"` // 名称列表
}
