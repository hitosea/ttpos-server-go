<?php

namespace help;

use DateTime;

// 获取某月的时间列表 				DateHelp::getMonthDays($date);
// 获取本周的时间列表 				DateHelp::getWeeksDays();
// 获取某天的信息-包含星期几 		 DateHelp::getDayInfo($date);
// 返回字符串的毫秒数时间戳			 DateHelp::getMillisecond();
// 公历转佛历			           DateHelp::changeBuddhistCalendar($date);

class DateHelp
{

    //获取某月的时间列表
    public static function getMonthDays($time = '')
    {
        $time = $time != '' ? (strstr($time, "-")  ? strtotime($time) : $time) : time();
        $week = date('d', $time);
        $weekarray = array("日", "一", "二", "三", "四", "五", "六");
        $dates = [];
        for ($i = 1; $i <= date('t', $time); $i++) {
            $date = date('Y-m-d', strtotime('+' . ($i - $week) . ' days', $time));
            $dates[] = [
                'date' => $date,
                'day' => date('d', strtotime('+' . ($i - $week) . ' days', $time)),
                'week' => "星期" . $weekarray[date("w", strtotime($date))],
            ];
        }
        return $dates;
    }

    //获取本周的时间列表
    public static function getWeeksDays()
    {
        $time = time();
        $weekarray = array("日", "一", "二", "三", "四", "五", "六");
        $dates = [];
        for ($i = 1; $i <= 7; $i++) {
            $date = date('Y-m-d', strtotime('+' . ($i - 1) . ' days', $time));
            $dates[] = [
                'date' => $date,
                'day' => date('d', strtotime('+' . ($i - 1) . ' days', $time)),
                'week' => "星期" . $weekarray[date("w", strtotime($date))],
            ];
        }
        return $dates;
    }

    //获取某天的信息-包含星期几
    public static function getDayInfo($time = '')
    {
        $time = $time != '' ? (strstr($time, "-")  ? strtotime($time) : $time) : time();
        $weekarray = array("日", "一", "二", "三", "四", "五", "六");
        $weekarray2 = array("7", "1", "2", "3", "4", "5", "6");
        $dates = [
            'date' => $date = date('Y-m-d', strtotime('+' . 0 . ' days', $time)),
            'day' => date('d', strtotime('+' . 0 . ' days', $time)),
            'wk' => $weekarray2[date("w", strtotime($date))],
            'week' => "星期" . $weekarray[date("w", strtotime($date))],
        ];
        return $dates;
    }

    /* 
	* 
	* 返回字符串的毫秒数时间戳 
	*/
    public static function getMillisecond()
    {
        list($msec, $sec) = explode(' ', microtime());
        return (float)sprintf('%.0f', (floatval($msec) + floatval($sec)) * 1000);
    }

    /* 
	* 
	* 公历转佛历
	*/
    public static function changeBuddhistCalendar($datetime)
    {
        // 将公历日期转换为时间戳
        $timestamp = is_int($datetime) ? $datetime : strtotime($datetime);
        // 公历转佛历的偏移量，以泰国为例，泰国使用佛历543年为起始（公元543年对应佛历1年）
        $thaiOffset = 543;
        // 计算佛历年份
        $thaiYear = intval(date('Y', $timestamp)) + $thaiOffset;
        // 获取月份和日期
        if (strstr($datetime,'/')) {
            $result = $thaiYear . '/' . date('m/d H:i:s', $timestamp);
        } else {
            $result = $thaiYear . '-' . date('m-d H:i:s', $timestamp);
        }
        // 
        return is_int($datetime) ? strtotime($result) : $result;
    }

    /**
     * 格式化时间
     *
     * @param int $value
     * @return string
     */
    public static function formatTimeHis($value)
    {
        try {
            return date('Y-m-d H:i:s', $value);
        } catch (\Throwable $th) {
            return $value;
        }
    }

    /**
     * get remaining_days
     */
    public static function getLicenseRemainingDays(array $license): int
    {
        $c_time = new DateTime('@' . $license['c_time']);
        $c_time->modify('+' . $license['day'] . ' days');
        $licenses_exp_time = $c_time->getTimestamp();
        $remaining_days = ceil(($licenses_exp_time - time()) / 60 / 60 / 24);
        return $license['day'] == 0 ? -1 : $remaining_days;
    }
}
