package utils

import (
	"testing"
)

func TestTimezone_Now(t *testing.T) {
	//tz := Timezone(ZH_TIMEZONE)
	//tz := Timezone("Asia/Shanghai")
	tz := SetTimezone("Asia/Shanghai")
	t.Log(tz.Now())
	t.Log(tz.NowUnix())
	t.Log(tz.NowUnixMilli())
	t.Log(tz.NowUnixMicro())
	t.Log(tz.TodayStartEnd())
	t.Log(tz.YesterdayStartEnd())
	t.Log(tz.WeekStartEnd())
	t.Log(tz.TodayStartEndUnix())
	t.Log(tz.YesterdayStartEndUnix())
	t.Log(tz.WeekStartEndUnix())
	jp := Timezone(JP_TIMEZONE)
	t.Log(jp.Now())
	t.Log(jp.NowUnix())
	t.Log(jp.NowUnixMilli())
	t.Log(jp.NowUnixMicro())
	t.Log(jp.TodayStartEnd())
	t.Log(jp.YesterdayStartEnd())
	t.Log(jp.WeekStartEnd())
}

func TestOpeningHoursStartEndUnix(t *testing.T) {
	// 测试中国时区
	tz := SetTimezone("Asia/Shanghai")

	// 打印当前时间戳，用于调试
	t.Logf("当前时间戳: %d", tz.NowUnix())
	t.Logf("当前时间: %s", tz.FormatUnixTimeDefault(tz.NowUnix()))

	// 测试用例1: 未设置营业时间
	startUnix, endUnix := tz.OpeningHoursStartEndUnix("")
	t.Logf("未设置营业时间:")
	t.Logf("- 开始时间: %s", tz.FormatUnixTimeDefault(startUnix))
	t.Logf("- 结束时间: %s", tz.FormatUnixTimeDefault(endUnix))

	// 测试用例2: 非跨天营业时间 "10:00-22:00"
	startUnix, endUnix = tz.OpeningHoursStartEndUnix("10:00-22:00")
	t.Logf("\n非跨天营业时间(10:00-22:00):")
	t.Logf("- 开始时间: %s", tz.FormatUnixTimeDefault(startUnix))
	t.Logf("- 结束时间: %s", tz.FormatUnixTimeDefault(endUnix))

	// 测试用例3: 跨天营业时间 "18:00-02:00"
	startUnix, endUnix = tz.OpeningHoursStartEndUnix("18:00-02:00")
	t.Logf("\n跨天营业时间(18:00-02:00):")
	t.Logf("- 开始时间: %s", tz.FormatUnixTimeDefault(startUnix))
	t.Logf("- 结束时间: %s", tz.FormatUnixTimeDefault(endUnix))

	// 测试用例4: 含分钟的营业时间 "19:30-23:45"
	startUnix, endUnix = tz.OpeningHoursStartEndUnix("19:30-23:45")
	t.Logf("\n含分钟的营业时间(19:30-23:45):")
	t.Logf("- 开始时间: %s", tz.FormatUnixTimeDefault(startUnix))
	t.Logf("- 结束时间: %s", tz.FormatUnixTimeDefault(endUnix))
}
