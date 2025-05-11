package utils

import (
	"errors"
	"time"
	_ "time/tzdata"
)

type TimeUtil interface {
	Now() time.Time                                       // 获取当前时间
	NowUnix() int64                                       // 获取当前时间戳,10位，1739283862
	NowUnixMilli() int64                                  // 获取当前时间戳（毫秒）13位，1739283862825
	NowUnixMicro() int64                                  // 获取当前时间戳（微秒）,16位，1739283862825531
	TodayStartEnd() (time.Time, time.Time)                // 获取今天的开始时间和结束时间
	YesterdayStartEnd() (time.Time, time.Time)            // 获取昨天的开始时间和结束时间
	WeekStartEnd() (time.Time, time.Time)                 // 获取本周的开始时间和结束时间
	MonthStartEnd() (time.Time, time.Time)                // 获取本月的开始时间和结束时间
	TodayStartEndUnix() (int64, int64)                    // 获取今天的开始时间和结束时间戳
	YesterdayStartEndUnix() (int64, int64)                // 获取昨天的开始时间和结束时间戳
	WeekStartEndUnix() (int64, int64)                     // 获取本周的开始时间和结束时间戳
	MonthStartEndUnix() (int64, int64)                    // 获取本月的开始时间和结束时间戳
	FormatUnixTime(timestamp int64, layout string) string // 将时间戳转换为指定格式的时间字符串
	FormatUnixTimeDefault(timestamp int64) string         // 将时间戳转换为默认格式(2006-01-02 15:04:05)的时间字符串
	FormatUnixTimeWithSlash(timestamp int64) string       // 将时间戳转换为默认格式(2006/01/02 15:04:05)的时间字符串
	GetTimeRange(dayType DayType) (int64, int64, error)   // 订单列表：今天、昨天、本周搜索时间范围
	FormatTimeToUnix(timeStr string) (int64, error)       // 2025-04-30转为时间戳
}

type Timezone string

const (
	ZH_TIMEZONE Timezone = "Asia/Shanghai"   // 中国时区 UTC+8
	JP_TIMEZONE Timezone = "Asia/Tokyo"      // 日本时区 UTC+9
	TH_TIMEZONE Timezone = "Asia/Bangkok"    // 泰国时区 UTC+7
	TR_TIMEZONE Timezone = "Europe/Istanbul" // 土耳其时区 UTC+3
)

type DayType string

const (
	DayTypeToday     DayType = "today"
	DayTypeYesterday DayType = "yesterday"
	DayTypeThisWeek  DayType = "this_week"
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

func (t Timezone) MonthStartEnd() (time.Time, time.Time) {
	now := t.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())
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

func (t Timezone) MonthStartEndUnix() (int64, int64) {
	start, end := t.MonthStartEnd()
	return start.Unix(), end.Unix()
}

// FormatUnixTime 将Unix时间戳转换为指定格式的时间字符串
// 常用格式：
// - "2006-01-02 15:04:05" - 标准日期时间格式
// - "2006-01-02" - 仅日期格式
// - "15:04:05" - 仅时间格式
// 如果layout为空，则使用默认格式"2006-01-02 15:04:05"
func (t Timezone) FormatUnixTime(timestamp int64, layout string) string {
	loc, err := time.LoadLocation(string(t))
	if err != nil {
		loc = time.Local
	}
	// 如果layout为空，使用默认格式
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	// 将时间戳转换为time.Time
	tm := time.Unix(timestamp, 0).In(loc)
	// 使用指定格式返回时间字符串
	return tm.Format(layout)
}

// FormatUnixTimeDefault 将Unix时间戳转换为默认格式(2006-01-02 15:04:05)的时间字符串
// 这是一个便捷方法，等同于使用FormatUnixTime方法并传入空字符串作为layout
func (t Timezone) FormatUnixTimeDefault(timestamp int64) string {
	return t.FormatUnixTime(timestamp, "2006-01-02 15:04:05")
}

// FormatUnixTimeWithSlash 将Unix时间戳转换为默认格式(2006/01/02 15:04:05)的时间字符串
// 这是一个便捷方法，等同于使用FormatUnixTime方法并传入空字符串作为layout
func (t Timezone) FormatUnixTimeWithSlash(timestamp int64) string {
	return t.FormatUnixTime(timestamp, "2006/01/02 15:04:05")
}

// GetTimeRange 订单列表：今天、昨天、本周搜索时间范围
func (t Timezone) GetTimeRange(dayType DayType) (int64, int64, error) {
	// 加载时区
	loc, err := time.LoadLocation(string(t))
	if err != nil {
		return 0, 0, err
	}
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch dayType {
	case DayTypeToday:
		todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)
		return todayStart.Unix(), todayEnd.Unix(), nil
	case DayTypeYesterday:
		yesterdayStart := todayStart.Add(-24 * time.Hour)
		yesterdayEnd := todayStart.Add(-time.Second)
		return yesterdayStart.Unix(), yesterdayEnd.Unix(), nil
	case DayTypeThisWeek:
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		weekStart := todayStart.Add(-time.Duration(weekday-1) * 24 * time.Hour)
		weekEnd := weekStart.Add(7 * 24 * time.Hour).Add(-time.Second)
		return weekStart.Unix(), weekEnd.Unix(), nil
	default:
		return 0, 0, errors.New("typ error")
	}
}

// 2025-04-30转为时间戳
func (t Timezone) FormatTimeToUnix(timeStr string) (int64, error) {
	loc, err := time.LoadLocation(string(t))
	if err != nil {
		return 0, err
	}
	time, err := time.ParseInLocation("2006-01-02", timeStr, loc)
	if err != nil {
		return 0, err
	}
	return time.Unix(), nil
}
