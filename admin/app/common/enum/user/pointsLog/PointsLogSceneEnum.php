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

    // 收银机、点餐助手发卡赠送
    const CashierOrAssistant = 100;

    // 积分抵扣
    const POINTS_DEDUCT = 110;

    // 抵扣反结账
    const POINTS_REVERSE = 120;

    // 营销活动赠送
    const MARKETING_ACTIVITY = 130;

    // 后台发卡赠送
    const ADMIN_CARD_GIVE = 140;

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
            self::CashierOrAssistant => [
                'name' => __('添加会员发卡'),
                'value' => self::CashierOrAssistant,
                'describe' => '%s管理员添加会员发卡赠送操作 [%s]',
            ],
            self::POINTS_DEDUCT => [
                'name' => __('积分抵扣'),
                'value' => self::POINTS_DEDUCT,
                'describe' => '积分抵扣：%s',
            ],
            self::POINTS_REVERSE => [
                'name' => __('抵扣反结账'),
                'value' => self::POINTS_REVERSE,
                'describe' => '抵扣反结账：%s',
            ],
            self::MARKETING_ACTIVITY => [
                'name' => __('营销活动'),
                'value' => self::MARKETING_ACTIVITY,
                'describe' => __('邀请消费有礼'),
            ],
            self::ADMIN_CARD_GIVE => [
                'name' => __('添加会员发卡'),
                'value' => self::ADMIN_CARD_GIVE,
                'describe' => '后台管理员 [%s] 操作',
            ],
        ];
    }
}
