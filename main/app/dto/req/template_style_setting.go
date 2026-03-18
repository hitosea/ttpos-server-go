package req

// SaveTemplateStyleSettingReq 保存模板样式设置
type SaveTemplateStyleSettingReq struct {
	TemplateStyle int `json:"template_style" binding:"required,oneof=1 2 3"` // 模板样式：1/2/3
}
