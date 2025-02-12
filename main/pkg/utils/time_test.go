package utils

import "testing"

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
