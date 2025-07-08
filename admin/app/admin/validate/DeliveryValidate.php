<?php

namespace app\admin\validate;

use app\common\enum\settings\SettingEnum;
use app\common\model\supplier\Supplier;
use app\common\validate\BaseValidate;


class DeliveryValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'uuid' => 'require|integer|checkIdExist',
        'channels' => ['require', 'array', 'min:1', 'checkCompanyChannels']
    ];



    // 合并后的验证方法
    protected function checkChannel($value)
    {
        if (!in_array($value, SettingEnum::DELIVERY_CHANNELS)) {
            return "外送渠道错误";
        }
        return true;
    }

    // 合并后的验证方法
    protected function checkDistanceRange($value, $rule, $data = [])
    {
        $prevEnd = null;
        $hasUnlimited = false;

        foreach ($value as $index => $item) {

            if (!isset($item['price_per_km']) || !is_float($item['price_per_km']) && !is_int($item['price_per_km'])) {
                return "距离范围单价错误";
            }

            // 检查无限范围元素后面是否还有元素
            if ($hasUnlimited) {
                return "距离范围设置了无限范围，后面不能再有其他范围";
            }

            // 检查无限范围设置
            if ($item['is_unlimited']) {
                $hasUnlimited = true;
                // 无限范围可以不设置 end
                continue;
            }

            // 非无限范围必须设置 end
            if (!isset($item['end']) || !is_float($item['end']) && !is_int($item['end']) || !($item['end'] > 0)) {
                return "距离范围不是最大范围，必须设置结束距离，且必须大于0";
            }

            // 检查 end 是否大于 start
            if ($item['end'] < $prevEnd) {
                return "距离范围的结束距离必须大于上一个结束距离";
            }

            $prevEnd = $item['end'];
        }

        return true;
    }

    protected function checkCompanyChannels($value, $rule, $data = [])
    {
        foreach ($value as $index => $item) {
            if (!isset($item['channel'])) {
                return "外送渠道必填";
            }
            if (($err = $this->checkChannel($item['channel'])) !== true) {
                return $err;
            }
            if (!empty(($data['uuid'] ?? 0))) {
                if (!isset($item['config_type'])) {
                    return "参数同步方式必填";
                }
                if (!in_array($item['config_type'], ['auto_sync', 'manual'])) {
                    return "参数同步方式仅支持自动同步和手动设置";
                }
            }

            if (!isset($item['basic_fee'])) {
                return "外送基础服务费必填";
            }

            if (!isset($item['base_delivery_fee'])) {
                return "起步配送费必填";
            }

            if (!isset($item['rider_acceptance_timeout'])) {
                return "骑手未接单取消时间必填";
            }
            if ($item['rider_acceptance_timeout'] < 1 || $item['rider_acceptance_timeout'] > 60) {
                return "骑手未接单取消时间范围1-60分钟";
            }
            if (!isset($item['distance_range']) || !is_array($item['distance_range']) || count($item['distance_range']) == 0) {
                return "距离范围参数错误";
            }
            if (($err = $this->checkDistanceRange($item['distance_range'], $rule)) !== true) {
                return $err;
            }
        }
        return true;
    }


    /**
     * 验证id是否存在
     */
    protected function checkIdExist($value, $rule, $data = [])
    {
        if (!Supplier::where('uuid', $value)->find()) {
            return false;
        } else {
            return true;
        }
    }


    protected $message  = [];

    protected $scene = [
        'edit' => [
            'channels',
        ],
        "uuid" => [
            'uuid'
        ],
        "add_company" => [
            'uuid',
            'channels',
        ]
    ];
}
