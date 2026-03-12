package setting

import (
	"testing"
	"time"

	"ttpos-server-go/pkg/utils"
)

// mockNow 设置固定时间，返回清理函数。
// 固定时间使用 UTC，由 Timezone.Now() 转换到目标时区。
func mockNow(t *testing.T, utcTime time.Time) {
	t.Helper()
	utils.SetNowFunc(func() time.Time { return utcTime })
	t.Cleanup(utils.ResetNowFunc)
}

func TestStoreScanOrderSettingResp_IsStoreResting(t *testing.T) {
	const tz = "Asia/Shanghai"

	// 固定为 UTC 2025-06-15 06:30:00 → 上海时间 14:30
	fixedUTC := time.Date(2025, 6, 15, 6, 30, 0, 0, time.UTC)
	mockNow(t, fixedUTC)

	tests := []struct {
		name         string
		setting      StoreScanOrderSettingResp
		timezone     string
		openingHours string
		want         bool
	}{
		{
			name:         "未启用(IsEnabled=0)_应为休息中",
			setting:      StoreScanOrderSettingResp{IsEnabled: 0},
			timezone:     tz,
			openingHours: "00:00-23:59",
			want:         true,
		},
		{
			name:         "未启用(IsEnabled=2)_非标准值也视为休息中",
			setting:      StoreScanOrderSettingResp{IsEnabled: 2},
			timezone:     tz,
			openingHours: "00:00-23:59",
			want:         true,
		},
		{
			name:         "已启用_全天营业_应不休息",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     tz,
			openingHours: "00:00-23:59",
			want:         false,
		},
		{
			name:         "已启用_空营业时间视为全天营业_应不休息",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     tz,
			openingHours: "",
			want:         false,
		},
		{
			name:         "已启用_格式错误视为全天营业_应不休息",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     tz,
			openingHours: "invalid",
			want:         false,
		},
		{
			name:         "未启用_即使全天营业也是休息中",
			setting:      StoreScanOrderSettingResp{IsEnabled: 0},
			timezone:     tz,
			openingHours: "",
			want:         true,
		},
		{
			name:         "已启用_空时区_全天营业_应不休息",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     "",
			openingHours: "00:00-23:59",
			want:         false,
		},
		{
			name:         "已启用_非跨天_当前时间在范围内_应不休息",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     tz,
			openingHours: "09:00-22:00", // 上海 14:30 在 [09:00, 22:00) 内
			want:         false,
		},
		{
			name:         "已启用_非跨天_当前时间不在范围内_应休息中",
			setting:      StoreScanOrderSettingResp{IsEnabled: 1},
			timezone:     tz,
			openingHours: "15:00-22:00", // 上海 14:30 不在 [15:00, 22:00) 内
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.setting.IsStoreResting(tt.timezone, tt.openingHours)
			if got != tt.want {
				t.Errorf("IsStoreResting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStoreScanOrderSettingResp_IsStoreResting_CrossDay(t *testing.T) {
	const tz = "Asia/Shanghai"
	enabled := StoreScanOrderSettingResp{IsEnabled: 1}

	t.Run("跨天营业_当前时间在晚间段_应不休息", func(t *testing.T) {
		// 固定为 UTC 2025-06-15 12:00:00 → 上海 20:00
		mockNow(t, time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC))

		// 营业时间 18:00-02:00（跨天），上海 20:00 在 [18:00, 次日02:00) 内
		got := enabled.IsStoreResting(tz, "18:00-02:00")
		if got {
			t.Error("IsStoreResting() = true, want false")
		}
	})

	t.Run("跨天营业_当前时间在凌晨段_应不休息", func(t *testing.T) {
		// 固定为 UTC 2025-06-15 17:00:00 → 上海次日 01:00
		mockNow(t, time.Date(2025, 6, 15, 17, 0, 0, 0, time.UTC))

		// 营业时间 18:00-02:00（跨天），上海 01:00 在 [18:00, 次日02:00) 内
		got := enabled.IsStoreResting(tz, "18:00-02:00")
		if got {
			t.Error("IsStoreResting() = true, want false")
		}
	})

	t.Run("跨天营业_当前时间不在范围内_应休息中", func(t *testing.T) {
		// 固定为 UTC 2025-06-15 02:00:00 → 上海 10:00
		mockNow(t, time.Date(2025, 6, 15, 2, 0, 0, 0, time.UTC))

		// 营业时间 18:00-02:00（跨天），上海 10:00 不在 [18:00, 次日02:00) 内
		got := enabled.IsStoreResting(tz, "18:00-02:00")
		if !got {
			t.Error("IsStoreResting() = false, want true")
		}
	})

	t.Run("跨天营业_恰好在结束时间_应休息中", func(t *testing.T) {
		// 固定为 UTC 2025-06-14 18:00:00 → 上海 02:00
		mockNow(t, time.Date(2025, 6, 14, 18, 0, 0, 0, time.UTC))

		// 营业时间 18:00-02:00，上海 02:00 恰好等于 endMinutes，不在 [18:00, 02:00) 内
		got := enabled.IsStoreResting(tz, "18:00-02:00")
		if !got {
			t.Error("IsStoreResting() = false, want true")
		}
	})

	t.Run("不同时区_跨天营业_美东时间在范围内", func(t *testing.T) {
		// 固定为 UTC 2025-06-16 05:00:00 → 美东(UTC-4 夏令时) 01:00
		mockNow(t, time.Date(2025, 6, 16, 5, 0, 0, 0, time.UTC))

		// 营业时间 22:00-03:00（跨天），美东 01:00 在 [22:00, 次日03:00) 内
		got := enabled.IsStoreResting("America/New_York", "22:00-03:00")
		if got {
			t.Error("IsStoreResting() = true, want false")
		}
	})

	t.Run("不同时区_跨天营业_美东时间不在范围内", func(t *testing.T) {
		// 固定为 UTC 2025-06-16 14:00:00 → 美东(UTC-4 夏令时) 10:00
		mockNow(t, time.Date(2025, 6, 16, 14, 0, 0, 0, time.UTC))

		// 营业时间 22:00-03:00（跨天），美东 10:00 不在范围内
		got := enabled.IsStoreResting("America/New_York", "22:00-03:00")
		if !got {
			t.Error("IsStoreResting() = false, want true")
		}
	})
}
