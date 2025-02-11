package utils

import "time"

type TimeUtil interface {
	Now() time.Time                            // 获取当前时间
	NowUnix() int64                            // 获取当前时间戳,10位，1739283862
	NowUnixMilli() int64                       // 获取当前时间戳（毫秒）13位，1739283862825
	NowUnixMicro() int64                       // 获取当前时间戳（微秒）,16位，1739283862825531
	TodayStartEnd() (time.Time, time.Time)     // 获取今天的开始时间和结束时间
	YesterdayStartEnd() (time.Time, time.Time) // 获取昨天的开始时间和结束时间
	WeekStartEnd() (time.Time, time.Time)      // 获取本周的开始时间和结束时间
	TodayStartEndUnix() (int64, int64)         // 获取今天的开始时间和结束时间戳
	YesterdayStartEndUnix() (int64, int64)     // 获取昨天的开始时间和结束时间戳
	WeekStartEndUnix() (int64, int64)          // 获取本周的开始时间和结束时间戳
}

type Timezone string

const (
	ZH_TIMEZONE Timezone = "Asia/Shanghai"   // 中国时区 UTC+8
	JP_TIMEZONE Timezone = "Asia/Tokyo"      // 日本时区 UTC+9
	TH_TIMEZONE Timezone = "Asia/Bangkok"    // 泰国时区 UTC+7
	TR_TIMEZONE Timezone = "Europe/Istanbul" // 土耳其时区 UTC+3
)

func SetTimezone(timezone string) TimeUtil {
	return Timezone(timezone)
}

func (t Timezone) Now() time.Time {
	loc, err := time.LoadLocation(string(t))
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func (t Timezone) NowUnix() int64 {
	return t.Now().Unix()
}

func (t Timezone) NowUnixMilli() int64 {
	return t.Now().UnixMilli()
}

func (t Timezone) NowUnixMicro() int64 {
	return t.Now().UnixMicro()
}

func (t Timezone) TodayStartEnd() (time.Time, time.Time) {
	now := t.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	return start, end
}

func (t Timezone) YesterdayStartEnd() (time.Time, time.Time) {
	now := t.Now().AddDate(0, 0, -1)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	return start, end
}

func (t Timezone) WeekStartEnd() (time.Time, time.Time) {
	now := t.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day()+(7-weekday), 23, 59, 59, 999999999, now.Location())
	return start, end
}

func (t Timezone) TodayStartEndUnix() (int64, int64) {
	start, end := t.TodayStartEnd()
	return start.Unix(), end.Unix()
}

func (t Timezone) YesterdayStartEndUnix() (int64, int64) {
	start, end := t.YesterdayStartEnd()
	return start.Unix(), end.Unix()
}

func (t Timezone) WeekStartEndUnix() (int64, int64) {
	start, end := t.WeekStartEnd()
	return start.Unix(), end.Unix()
}
