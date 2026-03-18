<?php

namespace app\common\enum\settings;

use MyCLabs\Enum\Enum;

/**
 * 时区枚举类
 */
class TimeZoneEnum extends Enum
{
    /**
     * 获取订单类型值
     */
    public static function data($type = 1)
    {
        $data = [
            [
                'name' => '（UTC+09:00）' . __('东京'),
                'value' => 'Asia/Tokyo',
            ],
            [
                'name' => '（UTC+08:00）' . __('北京，重庆，香港特别行政区，乌鲁木齐'),
                'value' => 'Asia/Shanghai',
            ],
            [
                'name' => '（UTC+07:00）' . __('曼谷，河内，雅加达'),
                'value' => 'Asia/Bangkok',
            ],
            [
                'name' => '（UTC+07:00）' . __('胡志明市，河内'),
                'value' => 'Asia/Ho_Chi_Minh',
            ],
            [
                'name' => '（UTC+06:30）' . __('内比都，仰光'),
                'value' => 'Asia/Yangon',
            ],
            [
                'name' => '（UTC+3:00）' . __('安卡拉'),
                'value' => 'Europe/Istanbul',
            ],
            [
                'name' => '（UTC+2:00）' . __('斯德哥尔摩'),
                'value' => 'Europe/Stockholm',
            ]
        ];

        if ($type == 2) {
            foreach ($data as $key => $val) {
                $data[$key]['key'] = $val['value'];
                unset($data[$key]['value']);
                $data[$key]['value'] = $val['name'];
            }
        }

        return $data;
    }
}
