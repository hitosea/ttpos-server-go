<?php

namespace app\common\enum\user\balanceLog;

use MyCLabs\Enum\Enum;

/**
 * 余额变动场景枚举类
 */
class BalanceLogSceneEnum extends Enum
{
    // 用户充值
    const RECHARGE = 10;

    // 用户消费
    const CONSUME = 20;

    // 管理员操作
    const ADMIN = 30;

    // 订单退款
    const REFUND = 40;

    // 余额提现
    const CASH = 50;

    // 订单反结账
    const REVERSE = 60;

    // 充值反结账
    const RECHARGE_REVERSE = 70;

    // 充值退款
    const RECHARGE_REFUND = 80;

    // 扣减
    const DEDUCT = 90;

    /**
     * 获取订单类型值
     */
    public static function data()
    {
        return [
            self::RECHARGE => [
                'name' => __('充值'),
                'value' => self::RECHARGE,
                'describe' => '用户充值：%s',
            ],
            self::CONSUME => [
                'name' => __('消费'),
                'value' => self::CONSUME,
                'describe' => '用户消费：%s',
            ],
            self::ADMIN => [
                'name' => __('管理员操作'),
                'value' => self::ADMIN,
                'describe' => '后台管理员 [%s] 操作',
            ],
            self::REFUND => [
                'name' => __('退款'),
                'value' => self::REFUND,
                'describe' => '订单退款：%s',
            ],
            self::REVERSE => [
                'name' => __('订单反结账'),
                'value' => self::REVERSE,
                'describe' => '订单反结账：%s',
            ],
            self::RECHARGE_REVERSE => [
                'name' => __('充值反结账'),
                'value' => self::RECHARGE_REVERSE,
                'describe' => '充值反结账：%s',
            ],
            self::RECHARGE_REFUND => [
                'name' => __('充值退款'),
                'value' => self::RECHARGE_REFUND,
                'describe' => '退款：%s',
            ],
            self::DEDUCT => [
                'name' => __('扣减'),
                'value' => self::DEDUCT,
                'describe' => '后台管理员扣减',
            ],
        ];
    }
}
