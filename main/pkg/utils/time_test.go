package utils

import (
	"testing"
	"time"
)

// TestFormatDateTimeToUnix 测试 FormatDateTimeToUnix 方法
func TestFormatDateTimeToUnix(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		timeStr   string
		wantErr   bool
		checkFunc func(t *testing.T, got int64, timezone string, timeStr string)
	}{
		{
			name:     "Asia/Shanghai YYYY-MM-DD 格式",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 验证时间戳不为0
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				// 验证可以转换回日期
				timeUtil := SetTimezone(timezone)
				dateStr := timeUtil.FormatUnixTime(got, "2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Shanghai YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				// 验证可以转换回日期时间
				timeUtil := SetTimezone(timezone)
				dateTimeStr := timeUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "Asia/Tokyo YYYY-MM-DD 格式",
			timezone: "Asia/Tokyo",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateStr := timeUtil.FormatUnixTime(got, "2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Tokyo YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Tokyo",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateTimeStr := timeUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "Asia/Bangkok YYYY-MM-DD 格式",
			timezone: "Asia/Bangkok",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateStr := timeUtil.FormatUnixTime(got, "2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Bangkok YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Bangkok",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateTimeStr := timeUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "时区差异验证 - 同一时刻不同时区",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 获取上海时区的时间戳
				shanghaiTimestamp := got

				// 转换为泰国时区，应该显示为 11:00:00（UTC+7 vs UTC+8，差1小时）
				bangkokUtil := SetTimezone("Asia/Bangkok")
				bangkokTimeStr := bangkokUtil.FormatUnixTime(shanghaiTimestamp, "2006-01-02 15:04:05")
				expectedBangkokTime := "2025-12-30 11:00:00"
				if bangkokTimeStr != expectedBangkokTime {
					t.Errorf("时区转换不正确: 上海 %s 应该对应泰国 %s，实际 %s", timeStr, expectedBangkokTime, bangkokTimeStr)
				}
			},
		},
		{
			name:     "错误格式 - 长度不正确",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 11:42",
			wantErr:  true,
		},
		{
			name:     "错误格式 - 无效日期",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-13-30 11:42:00",
			wantErr:  true,
		},
		{
			name:     "错误格式 - 无效时间",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 25:00:00",
			wantErr:  true,
		},
		{
			name:     "空字符串",
			timezone: "Asia/Shanghai",
			timeStr:  "",
			wantErr:  true,
		},
		{
			name:     "无效时区",
			timezone: "Invalid/Timezone",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUtil := SetTimezone(tt.timezone)
			got, err := timeUtil.FormatDateTimeToUnix(tt.timeStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDateTimeToUnix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got, tt.timezone, tt.timeStr)
			}
		})
	}
}

// TestFormatDateTimeToTime 测试 FormatDateTimeToTime 方法
func TestFormatDateTimeToTime(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		timeStr   string
		wantErr   bool
		checkFunc func(t *testing.T, got time.Time, timezone string, timeStr string)
	}{
		{
			name:     "Asia/Shanghai YYYY-MM-DD 格式",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got time.Time, timezone string, timeStr string) {
				// 验证时间不为零值
				if got.IsZero() {
					t.Error("时间不应该为零值")
				}
				// 验证时区正确
				loc, _ := time.LoadLocation(timezone)
				if got.Location().String() != loc.String() {
					t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), got.Location().String())
				}
				// 验证日期正确
				dateStr := got.Format("2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Shanghai YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got time.Time, timezone string, timeStr string) {
				if got.IsZero() {
					t.Error("时间不应该为零值")
				}
				loc, _ := time.LoadLocation(timezone)
				if got.Location().String() != loc.String() {
					t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), got.Location().String())
				}
				dateTimeStr := got.Format("2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "Asia/Tokyo YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Tokyo",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got time.Time, timezone string, timeStr string) {
				if got.IsZero() {
					t.Error("时间不应该为零值")
				}
				loc, _ := time.LoadLocation(timezone)
				if got.Location().String() != loc.String() {
					t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), got.Location().String())
				}
				dateTimeStr := got.Format("2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "Asia/Bangkok YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Bangkok",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got time.Time, timezone string, timeStr string) {
				if got.IsZero() {
					t.Error("时间不应该为零值")
				}
				loc, _ := time.LoadLocation(timezone)
				if got.Location().String() != loc.String() {
					t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), got.Location().String())
				}
				dateTimeStr := got.Format("2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "时区差异验证 - 同一时刻不同时区",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got time.Time, timezone string, timeStr string) {
				// 转换为 UTC 时间戳
				utcTimestamp := got.Unix()

				// 使用泰国时区转换回时间
				bangkokUtil := SetTimezone("Asia/Bangkok")
				bangkokTimeStr := bangkokUtil.FormatUnixTime(utcTimestamp, "2006-01-02 15:04:05")
				expectedBangkokTime := "2025-12-30 11:00:00"
				if bangkokTimeStr != expectedBangkokTime {
					t.Errorf("时区转换不正确: 上海 %s 应该对应泰国 %s，实际 %s", timeStr, expectedBangkokTime, bangkokTimeStr)
				}
			},
		},
		{
			name:     "错误格式 - 长度不正确",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-12-30 11:42",
			wantErr:  true,
		},
		{
			name:     "错误格式 - 无效日期",
			timezone: "Asia/Shanghai",
			timeStr:  "2025-13-30 11:42:00",
			wantErr:  true,
		},
		{
			name:     "空字符串",
			timezone: "Asia/Shanghai",
			timeStr:  "",
			wantErr:  true,
		},
		{
			name:     "无效时区",
			timezone: "Invalid/Timezone",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUtil := SetTimezone(tt.timezone)
			got, err := timeUtil.FormatDateTimeToTime(tt.timeStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDateTimeToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got, tt.timezone, tt.timeStr)
			}
		})
	}
}

// TestFormatDateTimeToUnix_EdgeCases 测试边界情况
func TestFormatDateTimeToUnix_EdgeCases(t *testing.T) {
	t.Run("带空格的日期时间字符串", func(t *testing.T) {
		timeUtil := SetTimezone("Asia/Shanghai")
		// 测试前后有空格的情况
		timeStr := "  2025-12-30 11:42:00  "
		got, err := timeUtil.FormatDateTimeToUnix(timeStr)
		if err != nil {
			t.Errorf("应该能处理带空格的字符串，但返回错误: %v", err)
		}
		if got == 0 {
			t.Error("时间戳不应该为0")
		}
	})

	t.Run("跨年日期", func(t *testing.T) {
		timeUtil := SetTimezone("Asia/Shanghai")
		timeStr := "2024-12-31 23:59:59"
		got, err := timeUtil.FormatDateTimeToUnix(timeStr)
		if err != nil {
			t.Errorf("应该能处理跨年日期，但返回错误: %v", err)
		}
		if got == 0 {
			t.Error("时间戳不应该为0")
		}
	})

	t.Run("跨月日期", func(t *testing.T) {
		timeUtil := SetTimezone("Asia/Shanghai")
		timeStr := "2025-01-31 23:59:59"
		got, err := timeUtil.FormatDateTimeToUnix(timeStr)
		if err != nil {
			t.Errorf("应该能处理跨月日期，但返回错误: %v", err)
		}
		if got == 0 {
			t.Error("时间戳不应该为0")
		}
	})
}

// BenchmarkFormatDateTimeToUnix 性能测试
func BenchmarkFormatDateTimeToUnix(b *testing.B) {
	timeUtil := SetTimezone("Asia/Shanghai")
	testCases := []string{
		"2025-12-30",
		"2025-12-30 11:42:00",
		"2025-12-30 23:59:59",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = timeUtil.FormatDateTimeToUnix(tc)
			}
		})
	}
}

// BenchmarkFormatDateTimeToTime 性能测试
func BenchmarkFormatDateTimeToTime(b *testing.B) {
	timeUtil := SetTimezone("Asia/Shanghai")
	testCases := []string{
		"2025-12-30",
		"2025-12-30 11:42:00",
		"2025-12-30 23:59:59",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = timeUtil.FormatDateTimeToTime(tc)
			}
		})
	}
}

// TestTimezoneConstants 验证所有时区常量可以被 time.LoadLocation 正确加载
func TestTimezoneConstants(t *testing.T) {
	tests := []struct {
		name     string
		timezone Timezone
		wantName string
	}{
		{"中国时区", ZH_TIMEZONE, "Asia/Shanghai"},
		{"日本时区", JP_TIMEZONE, "Asia/Tokyo"},
		{"泰国时区", TH_TIMEZONE, "Asia/Bangkok"},
		{"越南时区", VN_TIMEZONE, "Asia/Ho_Chi_Minh"},
		{"缅甸时区", MM_TIMEZONE, "Asia/Yangon"},
		{"土耳其时区", TR_TIMEZONE, "Europe/Istanbul"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.timezone) != tt.wantName {
				t.Errorf("常量值不匹配: 期望 %s, 实际 %s", tt.wantName, string(tt.timezone))
			}
			loc, err := time.LoadLocation(string(tt.timezone))
			if err != nil {
				t.Fatalf("无法加载时区 %s: %v", tt.timezone, err)
			}
			if loc.String() != tt.wantName {
				t.Errorf("时区名称不匹配: 期望 %s, 实际 %s", tt.wantName, loc.String())
			}
		})
	}
}

// TestVietnamTimezone 验证越南时区（UTC+07:00）时间计算
func TestVietnamTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		timeStr   string
		wantErr   bool
		checkFunc func(t *testing.T, got int64, timezone string, timeStr string)
	}{
		{
			name:     "Asia/Ho_Chi_Minh YYYY-MM-DD 格式",
			timezone: "Asia/Ho_Chi_Minh",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateStr := timeUtil.FormatUnixTime(got, "2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Ho_Chi_Minh YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Ho_Chi_Minh",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateTimeStr := timeUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "越南与泰国同为UTC+7 - 同一时刻应产生相同时间戳",
			timezone: "Asia/Ho_Chi_Minh",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 泰国同一时间应该产生相同时间戳（都是 UTC+7）
				bangkokUtil := SetTimezone("Asia/Bangkok")
				bangkokTimestamp, err := bangkokUtil.FormatDateTimeToUnix(timeStr)
				if err != nil {
					t.Fatalf("泰国时区解析失败: %v", err)
				}
				if got != bangkokTimestamp {
					t.Errorf("越南和泰国同为UTC+7，同一时间应产生相同时间戳: 越南=%d, 泰国=%d", got, bangkokTimestamp)
				}
			},
		},
		{
			name:     "越南与上海时区差异验证 - 差1小时",
			timezone: "Asia/Ho_Chi_Minh",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 上海 UTC+8 比越南 UTC+7 快1小时
				shanghaiUtil := SetTimezone("Asia/Shanghai")
				shanghaiTimeStr := shanghaiUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				expectedShanghaiTime := "2025-12-30 13:00:00"
				if shanghaiTimeStr != expectedShanghaiTime {
					t.Errorf("时区转换不正确: 越南 %s 应对应上海 %s，实际 %s", timeStr, expectedShanghaiTime, shanghaiTimeStr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUtil := SetTimezone(tt.timezone)
			got, err := timeUtil.FormatDateTimeToUnix(tt.timeStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDateTimeToUnix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got, tt.timezone, tt.timeStr)
			}
		})
	}
}

// TestMyanmarTimezone 验证缅甸时区（UTC+06:30）时间计算，特别是半小时偏移
func TestMyanmarTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		timeStr   string
		wantErr   bool
		checkFunc func(t *testing.T, got int64, timezone string, timeStr string)
	}{
		{
			name:     "Asia/Yangon YYYY-MM-DD 格式",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateStr := timeUtil.FormatUnixTime(got, "2006-01-02")
				if dateStr != timeStr {
					t.Errorf("日期不匹配: 期望 %s, 实际 %s", timeStr, dateStr)
				}
			},
		},
		{
			name:     "Asia/Yangon YYYY-MM-DD HH:mm:ss 格式",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30 11:42:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				if got == 0 {
					t.Error("时间戳不应该为0")
				}
				timeUtil := SetTimezone(timezone)
				dateTimeStr := timeUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				if dateTimeStr != timeStr {
					t.Errorf("日期时间不匹配: 期望 %s, 实际 %s", timeStr, dateTimeStr)
				}
			},
		},
		{
			name:     "缅甸与上海时区差异验证 - 差1.5小时（半小时偏移）",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30 10:30:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 上海 UTC+8 比缅甸 UTC+6:30 快1.5小时
				shanghaiUtil := SetTimezone("Asia/Shanghai")
				shanghaiTimeStr := shanghaiUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				expectedShanghaiTime := "2025-12-30 12:00:00"
				if shanghaiTimeStr != expectedShanghaiTime {
					t.Errorf("时区转换不正确: 缅甸 %s 应对应上海 %s，实际 %s", timeStr, expectedShanghaiTime, shanghaiTimeStr)
				}
			},
		},
		{
			name:     "缅甸与泰国时区差异验证 - 差30分钟",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 泰国 UTC+7 比缅甸 UTC+6:30 快30分钟
				bangkokUtil := SetTimezone("Asia/Bangkok")
				bangkokTimeStr := bangkokUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				expectedBangkokTime := "2025-12-30 12:30:00"
				if bangkokTimeStr != expectedBangkokTime {
					t.Errorf("时区转换不正确: 缅甸 %s 应对应泰国 %s，实际 %s", timeStr, expectedBangkokTime, bangkokTimeStr)
				}
			},
		},
		{
			name:     "缅甸与越南时区差异验证 - 差30分钟",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30 12:00:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 越南 UTC+7 比缅甸 UTC+6:30 快30分钟
				vietnamUtil := SetTimezone("Asia/Ho_Chi_Minh")
				vietnamTimeStr := vietnamUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				expectedVietnamTime := "2025-12-30 12:30:00"
				if vietnamTimeStr != expectedVietnamTime {
					t.Errorf("时区转换不正确: 缅甸 %s 应对应越南 %s，实际 %s", timeStr, expectedVietnamTime, vietnamTimeStr)
				}
			},
		},
		{
			name:     "缅甸半小时偏移 - 跨日场景",
			timezone: "Asia/Yangon",
			timeStr:  "2025-12-30 23:30:00",
			wantErr:  false,
			checkFunc: func(t *testing.T, got int64, timezone string, timeStr string) {
				// 上海 UTC+8 比缅甸快1.5小时，23:30 + 1.5h = 次日01:00
				shanghaiUtil := SetTimezone("Asia/Shanghai")
				shanghaiTimeStr := shanghaiUtil.FormatUnixTime(got, "2006-01-02 15:04:05")
				expectedShanghaiTime := "2025-12-31 01:00:00"
				if shanghaiTimeStr != expectedShanghaiTime {
					t.Errorf("跨日时区转换不正确: 缅甸 %s 应对应上海 %s，实际 %s", timeStr, expectedShanghaiTime, shanghaiTimeStr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeUtil := SetTimezone(tt.timezone)
			got, err := timeUtil.FormatDateTimeToUnix(tt.timeStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDateTimeToUnix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got, tt.timezone, tt.timeStr)
			}
		})
	}
}

// TestMyanmarTimezoneTimeUtil 验证缅甸时区的 TimeUtil 接口方法
func TestMyanmarTimezoneTimeUtil(t *testing.T) {
	timeUtil := SetTimezone(string(MM_TIMEZONE))

	t.Run("Now 返回正确时区", func(t *testing.T) {
		now := timeUtil.Now()
		loc, _ := time.LoadLocation("Asia/Yangon")
		if now.Location().String() != loc.String() {
			t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), now.Location().String())
		}
	})

	t.Run("TodayStartEnd 日期一致", func(t *testing.T) {
		start, end := timeUtil.TodayStartEnd()
		if start.Format("2006-01-02") != end.Format("2006-01-02") {
			t.Errorf("今日开始和结束日期不一致: start=%s, end=%s", start.Format("2006-01-02"), end.Format("2006-01-02"))
		}
		if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
			t.Errorf("今日开始时间应为 00:00:00, 实际 %s", start.Format("15:04:05"))
		}
		if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
			t.Errorf("今日结束时间应为 23:59:59, 实际 %s", end.Format("15:04:05"))
		}
	})

	t.Run("TodayStartEndUnix 时间戳差值验证", func(t *testing.T) {
		startUnix, endUnix := timeUtil.TodayStartEndUnix()
		// 一天的时间戳差值应为 86399 秒 (23:59:59 - 00:00:00)
		diff := endUnix - startUnix
		if diff != 86399 {
			t.Errorf("今日开始和结束时间戳差值应为 86399, 实际 %d", diff)
		}
		// 验证开始时间在缅甸时区确实是 00:00:00
		startTimeStr := timeUtil.FormatUnixTime(startUnix, "15:04:05")
		if startTimeStr != "00:00:00" {
			t.Errorf("今日开始时间应为 00:00:00, 实际 %s", startTimeStr)
		}
		endTimeStr := timeUtil.FormatUnixTime(endUnix, "15:04:05")
		if endTimeStr != "23:59:59" {
			t.Errorf("今日结束时间应为 23:59:59, 实际 %s", endTimeStr)
		}
	})

	t.Run("WeekStartEndUnix 时间戳验证", func(t *testing.T) {
		startUnix, endUnix := timeUtil.WeekStartEndUnix()
		// 一周的时间戳差值应为 7*86400 - 1 = 604799 秒
		diff := endUnix - startUnix
		if diff != 604799 {
			t.Errorf("本周开始和结束时间戳差值应为 604799, 实际 %d", diff)
		}
	})

	t.Run("MonthStartEndUnix 时间戳验证", func(t *testing.T) {
		startUnix, endUnix := timeUtil.MonthStartEndUnix()
		if startUnix >= endUnix {
			t.Errorf("月开始时间应早于结束时间: start=%d, end=%d", startUnix, endUnix)
		}
		// 验证开始时间在缅甸时区是当月1号 00:00:00
		startTimeStr := timeUtil.FormatUnixTime(startUnix, "02 15:04:05")
		if startTimeStr != "01 00:00:00" {
			t.Errorf("月开始时间应为 01 00:00:00, 实际 %s", startTimeStr)
		}
	})

	t.Run("缅甸与上海 TodayStartEndUnix 偏移验证", func(t *testing.T) {
		// 缅甸 00:00:00 对应上海 01:30:00，所以缅甸的一天开始比上海晚 1.5 小时
		mmStart, _ := timeUtil.TodayStartEndUnix()
		shUtil := SetTimezone(string(ZH_TIMEZONE))
		shStart, _ := shUtil.TodayStartEndUnix()

		// 缅甸的一天开始时间戳应比上海晚 5400 秒 (1.5小时)
		diff := mmStart - shStart
		if diff != 5400 {
			t.Errorf("缅甸与上海日开始时间差应为 5400 秒 (1.5h), 实际 %d", diff)
		}
	})

	t.Run("UTC 偏移验证 +06:30", func(t *testing.T) {
		// 构造一个已知时间点验证 UTC 偏移
		loc, _ := time.LoadLocation("Asia/Yangon")
		testTime := time.Date(2025, 6, 15, 12, 0, 0, 0, loc)
		_, offset := testTime.Zone()
		expectedOffset := 6*3600 + 30*60 // 6小时30分钟 = 23400秒
		if offset != expectedOffset {
			t.Errorf("UTC 偏移不正确: 期望 %d 秒 (UTC+06:30), 实际 %d 秒", expectedOffset, offset)
		}
	})
}

// TestVietnamTimezoneTimeUtil 验证越南时区的 TimeUtil 接口方法
func TestVietnamTimezoneTimeUtil(t *testing.T) {
	timeUtil := SetTimezone(string(VN_TIMEZONE))

	t.Run("Now 返回正确时区", func(t *testing.T) {
		now := timeUtil.Now()
		loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
		if now.Location().String() != loc.String() {
			t.Errorf("时区不匹配: 期望 %s, 实际 %s", loc.String(), now.Location().String())
		}
	})

	t.Run("TodayStartEndUnix 时间戳差值验证", func(t *testing.T) {
		startUnix, endUnix := timeUtil.TodayStartEndUnix()
		diff := endUnix - startUnix
		if diff != 86399 {
			t.Errorf("今日开始和结束时间戳差值应为 86399, 实际 %d", diff)
		}
	})

	t.Run("越南与上海 TodayStartEndUnix 偏移验证", func(t *testing.T) {
		vnStart, _ := timeUtil.TodayStartEndUnix()
		shUtil := SetTimezone(string(ZH_TIMEZONE))
		shStart, _ := shUtil.TodayStartEndUnix()

		// 越南的一天开始时间戳应比上海晚 3600 秒 (1小时)
		diff := vnStart - shStart
		if diff != 3600 {
			t.Errorf("越南与上海日开始时间差应为 3600 秒 (1h), 实际 %d", diff)
		}
	})

	t.Run("越南与泰国 TodayStartEndUnix 相同", func(t *testing.T) {
		vnStart, vnEnd := timeUtil.TodayStartEndUnix()
		thUtil := SetTimezone(string(TH_TIMEZONE))
		thStart, thEnd := thUtil.TodayStartEndUnix()

		// 越南和泰国同为 UTC+7，日开始和结束时间戳应完全相同
		if vnStart != thStart {
			t.Errorf("越南和泰国日开始时间戳应相同: 越南=%d, 泰国=%d", vnStart, thStart)
		}
		if vnEnd != thEnd {
			t.Errorf("越南和泰国日结束时间戳应相同: 越南=%d, 泰国=%d", vnEnd, thEnd)
		}
	})
}
