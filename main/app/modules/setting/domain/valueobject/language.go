package valueobject

// LanguageItem 语言项值对象
type LanguageItem struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// IsValid 验证语言项是否有效
func (l LanguageItem) IsValid() bool {
	return l.Name != "" && l.Key != "" && l.Value != ""
}

// NewLanguageItem 创建新的语言项
func NewLanguageItem(name, key, value string) LanguageItem {
	return LanguageItem{
		Name:  name,
		Key:   key,
		Value: value,
	}
}
