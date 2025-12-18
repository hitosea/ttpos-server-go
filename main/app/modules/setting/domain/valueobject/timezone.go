package valueobject

// TimeZoneItem 时区项值对象
type TimeZoneItem struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// IsValid 验证时区项是否有效
func (t TimeZoneItem) IsValid() bool {
	return t.Name != "" && t.Key != "" && t.Value != ""
}

// NewTimeZoneItem 创建新的时区项
func NewTimeZoneItem(name, key, value string) TimeZoneItem {
	return TimeZoneItem{
		Name:  name,
		Key:   key,
		Value: value,
	}
}
