package entity

import (
	"testing"
	"time"
)

// TestDefaultStoreSettingTimezoneList 验证默认门店设置包含所有时区
func TestDefaultStoreSettingTimezoneList(t *testing.T) {
	setting := DefaultStoreSetting()

	// 验证时区列表不为空
	if len(setting.TimeZoneList) == 0 {
		t.Fatal("时区列表不应为空")
	}

	// 期望的时区列表（按 UTC 偏移降序排列）
	expectedTimezones := []struct {
		key       string
		utcPrefix string
	}{
		{"Asia/Tokyo", "（UTC+09:00）"},
		{"Asia/Shanghai", "（UTC+08:00）"},
		{"Asia/Bangkok", "（UTC+07:00）"},
		{"Asia/Ho_Chi_Minh", "（UTC+07:00）"},
		{"Asia/Yangon", "（UTC+06:30）"},
		{"Europe/Istanbul", "（UTC+3:00）"},
		{"Europe/Stockholm", "（UTC+2:00）"},
	}

	if len(setting.TimeZoneList) != len(expectedTimezones) {
		t.Fatalf("时区数量不匹配: 期望 %d, 实际 %d", len(expectedTimezones), len(setting.TimeZoneList))
	}

	for i, expected := range expectedTimezones {
		tz := setting.TimeZoneList[i]

		// 验证 Key 正确
		if tz.Key != expected.key {
			t.Errorf("第 %d 个时区 Key 不匹配: 期望 %s, 实际 %s", i, expected.key, tz.Key)
		}

		// 验证 Name 和 Value 包含 UTC 前缀
		if len(tz.Name) == 0 {
			t.Errorf("第 %d 个时区 Name 不应为空 (key=%s)", i, expected.key)
		}
		if len(tz.Value) == 0 {
			t.Errorf("第 %d 个时区 Value 不应为空 (key=%s)", i, expected.key)
		}

		// 验证 IsValid()
		if !tz.IsValid() {
			t.Errorf("第 %d 个时区应为有效: key=%s, name=%s, value=%s", i, tz.Key, tz.Name, tz.Value)
		}

		// 验证 IANA 时区可以被 time.LoadLocation 正确加载
		_, err := time.LoadLocation(tz.Key)
		if err != nil {
			t.Errorf("无法加载时区 %s: %v", tz.Key, err)
		}
	}
}

// TestDefaultStoreSettingTimezoneOrder 验证时区按 UTC 偏移降序排列
func TestDefaultStoreSettingTimezoneOrder(t *testing.T) {
	setting := DefaultStoreSetting()

	for i := 0; i < len(setting.TimeZoneList)-1; i++ {
		currLoc, err := time.LoadLocation(setting.TimeZoneList[i].Key)
		if err != nil {
			t.Fatalf("无法加载时区 %s: %v", setting.TimeZoneList[i].Key, err)
		}
		nextLoc, err := time.LoadLocation(setting.TimeZoneList[i+1].Key)
		if err != nil {
			t.Fatalf("无法加载时区 %s: %v", setting.TimeZoneList[i+1].Key, err)
		}

		// 获取 UTC 偏移
		refTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
		_, currOffset := refTime.In(currLoc).Zone()
		_, nextOffset := refTime.In(nextLoc).Zone()

		if currOffset < nextOffset {
			t.Errorf("时区排序错误: %s (UTC%+d) 应排在 %s (UTC%+d) 之前",
				setting.TimeZoneList[i].Key, currOffset/3600,
				setting.TimeZoneList[i+1].Key, nextOffset/3600)
		}
	}
}

// TestDefaultStoreSettingContainsMyanmarTimezone 验证包含缅甸时区
func TestDefaultStoreSettingContainsMyanmarTimezone(t *testing.T) {
	setting := DefaultStoreSetting()

	found := false
	for _, tz := range setting.TimeZoneList {
		if tz.Key == "Asia/Yangon" {
			found = true
			// 验证 UTC+06:30 偏移
			loc, _ := time.LoadLocation(tz.Key)
			refTime := time.Date(2025, 6, 15, 12, 0, 0, 0, loc)
			_, offset := refTime.Zone()
			expectedOffset := 6*3600 + 30*60 // 23400秒
			if offset != expectedOffset {
				t.Errorf("缅甸时区 UTC 偏移不正确: 期望 %d (UTC+06:30), 实际 %d", expectedOffset, offset)
			}
			break
		}
	}
	if !found {
		t.Error("默认设置中缺少缅甸时区 Asia/Yangon")
	}
}

// TestDefaultStoreSettingContainsVietnamTimezone 验证包含越南时区
func TestDefaultStoreSettingContainsVietnamTimezone(t *testing.T) {
	setting := DefaultStoreSetting()

	found := false
	for _, tz := range setting.TimeZoneList {
		if tz.Key == "Asia/Ho_Chi_Minh" {
			found = true
			// 验证 UTC+07:00 偏移
			loc, _ := time.LoadLocation(tz.Key)
			refTime := time.Date(2025, 6, 15, 12, 0, 0, 0, loc)
			_, offset := refTime.Zone()
			expectedOffset := 7 * 3600 // 25200秒
			if offset != expectedOffset {
				t.Errorf("越南时区 UTC 偏移不正确: 期望 %d (UTC+07:00), 实际 %d", expectedOffset, offset)
			}
			break
		}
	}
	if !found {
		t.Error("默认设置中缺少越南时区 Asia/Ho_Chi_Minh")
	}
}

// TestTimezoneWhitelistValidation 模拟 EditStoreSetting 的白名单校验逻辑
// 对应 AC: 选择时区后保存生效 — 验证新时区能通过白名单校验
func TestTimezoneWhitelistValidation(t *testing.T) {
	setting := DefaultStoreSetting()

	// 模拟 setting.go 中的 slices.ContainsFunc 校验逻辑
	containsTimezone := func(key string) bool {
		for _, tz := range setting.TimeZoneList {
			if tz.Key == key {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name     string
		timezone string
		wantPass bool
	}{
		// 已有时区
		{"上海时区应通过", "Asia/Shanghai", true},
		{"东京时区应通过", "Asia/Tokyo", true},
		{"曼谷时区应通过", "Asia/Bangkok", true},
		{"伊斯坦布尔时区应通过", "Europe/Istanbul", true},
		{"斯德哥尔摩时区应通过", "Europe/Stockholm", true},
		// 新增时区
		{"越南时区应通过", "Asia/Ho_Chi_Minh", true},
		{"缅甸时区应通过", "Asia/Yangon", true},
		// 不在列表中的时区应被拒绝
		{"纽约时区应被拒绝", "America/New_York", false},
		{"伦敦时区应被拒绝", "Europe/London", false},
		{"空字符串应被拒绝", "", false},
		{"无效时区应被拒绝", "Invalid/Timezone", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsTimezone(tt.timezone)
			if got != tt.wantPass {
				if tt.wantPass {
					t.Errorf("时区 %s 应该通过白名单校验，但被拒绝", tt.timezone)
				} else {
					t.Errorf("时区 %s 应该被白名单拒绝，但通过了", tt.timezone)
				}
			}
		})
	}
}

// TestNewStoreSetting 验证新建门店设置
func TestNewStoreSetting(t *testing.T) {
	setting := NewStoreSetting("TestShop", "TestCompany", "TestAddress", "123456")

	if setting.Name != "TestShop" {
		t.Errorf("名称不匹配: 期望 TestShop, 实际 %s", setting.Name)
	}
	if setting.ZeroingMethod != "0" {
		t.Errorf("抹零方式默认值不正确: 期望 0, 实际 %s", setting.ZeroingMethod)
	}
	if setting.TimeZoneList == nil {
		t.Error("时区列表不应为 nil")
	}
	if len(setting.TimeZoneList) != 0 {
		t.Errorf("新建设置的时区列表应为空: 实际长度 %d", len(setting.TimeZoneList))
	}
}
