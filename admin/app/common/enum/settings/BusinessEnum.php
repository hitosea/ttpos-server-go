<?php

namespace app\common\enum\settings;

use MyCLabs\Enum\Enum;

/**
 * 门店业务设置枚举类
 */
class BusinessEnum extends Enum
{
    /**
     * 自动抹零方式枚举数据
     */
    public static function zeroingMethodDefault()
    {
        return [
            [
                'key' => '0',
                'name' => __('实款实收'),
            ],
            [
                'key' => '1',
                'name' => __('抹分'),
            ],
            [
                'key' => '2',
                'name' => __('抹角'),
            ],
            [
                'key' => '3',
                'name' => __('四舍五入保留一位小数'),
            ],
            [
                'key' => '4',
                'name' => __('四舍五入到整数'),
            ],
        ];
    }

    /**
     * 结账自动抹零方式枚举数据
     */
    public static function checkoutZeroingMethodDefault()
    {
        return [
            [
                'key' => '0',
                'name' => __('实款实收'),
            ],
            [
                'key' => '1',
                'name' => __('抹分'),
            ],
            [
                'key' => '2',
                'name' => __('抹角'),
            ],
            [
                'key' => '5',
                'name' => __('抹元'),
            ],
        ];
    }

    /**
     * 赠菜计算方式枚举数据
     */
    public static function giftMethodDefault()
    {
        return [
            [
                'key' => '10',
                'name' => __('计入总销售额、优惠折扣'),
            ],
            [
                'key' => '20',
                'name' => __('不计入总销售额、优惠折扣'),
            ],
        ];
    }

    /**
     * 免单计算方式枚举数据
     */
    public static function freeMethodDefault()
    {
        return [
            [
                'key' => '10',
                'name' => __('计入总销售额、优惠折扣、服务费、税费'),
            ],
            [
                'key' => '20',
                'name' => __('不计入总销售额、优惠折扣、服务费、税费'),
            ],
        ];
    }
}
