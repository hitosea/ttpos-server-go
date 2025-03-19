<?php

namespace app\common\enum\user\pointsLog;

use MyCLabs\Enum\Enum;

/**
 * 积分变动场景枚举类
 */
class PointsLogSceneEnum extends Enum
{
    // 用户充值
    const RECHARGE = 10;

    // 消费赠送
    const CONSUME = 20;

    // 管理员操作
    const ADMIN = 30;

    // 退款扣除
    const REFUND = 40;

    // 订单反结账
    const REVERSE = 60;

    // 充值赠送
    const RECHARGE_GIVE = 70;

    // 充值反结账
    const RECHARGE_REVERSE = 80;

    // 扣减
    const DEDUCT = 90;

    /**
     * 获取订单类型值
     */
    public static function data()
    {
        return [
            self::CONSUME => [
                'name' => __('订单赠送'),
                'value' => self::CONSUME,
                'describe' => '订单赠送：%s',
            ],
            self::ADMIN => [
                'name' => __('管理员操作'),
                'value' => self::ADMIN,
                'describe' => '后台管理员 [%s] 操作',
            ],
            self::REFUND => [
                'name' => __('退款扣除'),
                'value' => self::REFUND,
                'describe' => '退款扣除：%s',
            ],
            self::REVERSE => [
                'name' => __('订单反结账'),
                'value' => self::REVERSE,
                'describe' => '订单反结账：%s',
            ],
            self::RECHARGE_GIVE => [
                'name' => __('充值赠送'),
                'value' => self::RECHARGE_GIVE,
                'describe' => '充值赠送：%s',
            ],
            self::RECHARGE_REVERSE => [
                'name' => __('充值反结账'),
                'value' => self::RECHARGE_REVERSE,
                'describe' => '充值反结账：%s',
            ],
            self::DEDUCT => [
                'name' => __('扣减'),
                'value' => self::DEDUCT,
                'describe' => '后台管理员扣减',
            ],
        ];
    }
}
