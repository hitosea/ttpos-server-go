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
