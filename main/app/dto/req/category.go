package req

type CreateCategoryRequest struct {
	ParentID   uint                `json:"parent_id"`
	CategoryID uint                `json:"category_id"`
	Name       CategoryTranslation `json:"name"`
	Sort       int                 `json:"sort"`
}

type CategoryTranslation struct {
	TH   string `json:"th"`
	ZH   string `json:"zh"`
	ZHTW string `json:"zhtw"`
	EN   string `json:"en"`
	JA   string `json:"ja"`
	KO   string `json:"ko"`
	MY   string `json:"my"`
	TR   string `json:"tr"`
}
