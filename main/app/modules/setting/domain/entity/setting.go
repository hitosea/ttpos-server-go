package entity

import (
	"time"
)

// Setting 聚合根 - 设置项
type Setting struct {
	Key         string    // 设置项标示
	Description string    // 设置项描述
	Values      string    // 设置内容（json格式）
	CreateTime  time.Time // 创建时间
	UpdateTime  time.Time // 更新时间
}

// SettingID 设置项ID值对象
type SettingID struct {
	Key string
}

// NewSettingID 创建设置项ID
func NewSettingID(key string) SettingID {
	return SettingID{Key: key}
}

// String 返回字符串表示
func (s SettingID) String() string {
	return s.Key
}

// IsEmpty 检查是否为空
func (s SettingID) IsEmpty() bool {
	return s.Key == ""
}

// NewSetting 创建新的设置项
func NewSetting(key, description, values string) *Setting {
	now := time.Now()
	return &Setting{
		Key:         key,
		Description: description,
		Values:      values,
		CreateTime:  now,
		UpdateTime:  now,
	}
}

// UpdateValues 更新设置值
func (s *Setting) UpdateValues(values string) {
	s.Values = values
	s.UpdateTime = time.Now()
}

// GetKey 获取设置项键
func (s *Setting) GetKey() string {
	return s.Key
}

// GetValues 获取设置值
func (s *Setting) GetValues() string {
	return s.Values
}
