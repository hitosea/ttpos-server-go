<?php

namespace app\admin\validate;

use app\common\enum\settings\SettingEnum;
use app\common\model\supplier\Supplier;
use app\common\validate\BaseValidate;


class DeliveryValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule =   [
        'channel|外送渠道' => ['require', 'checkChannel'],
        'basic_fee|外送基础服务费' => 'require',
        'base_delivery_fee|起步配送费' => 'require',
        'rider_acceptance_timeout|骑手未接单取消时间' => 'require|between:1,60', // 骑手未接单多少分钟后自动取消订单（1-60分钟）
        'distance_range|距离范围' => ['require', 'array', 'min:1', "checkDistanceRange"],

        'uuid' => 'require|checkIdExist',
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
    protected function checkDistanceRange($value)
    {
        $prevEnd = null;
        $hasUnlimited = false;

        foreach ($value as $index => $item) {
            // 检查无限范围元素后面是否还有元素
            if ($hasUnlimited) {
                return "距离范围设置了无限范围，后面不能再有其他范围";
            }

            // 如果不是第一个元素，检查 start 是否等于前一个的 end
            if ($prevEnd !== null && $item['start'] != $prevEnd) {
                return "距离范围的起始值必须等于前一个范围的结束值";
            }

            // 检查无限范围设置
            if ($item['is_unlimited']) {
                $hasUnlimited = true;
                // 无限范围可以不设置 end
                continue;
            }

            // 非无限范围必须设置 end
            if (!isset($item['end']) || $item['end'] === '') {
                return "距离范围不是无限范围，必须设置结束值";
            }

            // 检查 end 是否是数字
            if (!is_numeric($item['end'])) {
                return "距离范围的结束值必须是数字";
            }

            // 检查 end 是否大于等于0
            if ($item['end'] < 0) {
                return "距离范围的结束值必须大于等于0";
            }

            // 检查 end 是否大于 start
            if ($item['end'] <= $item['start']) {
                return "距离范围的结束值必须大于起始值";
            }

            $prevEnd = $item['end'];
        }

        return true;
    }

    protected function checkCompanyChannels($value)
    {
        foreach ($value as $index => $item) {
            if (!isset($item['channel'])) {
                return "外送渠道必填";
            }
            if (($err = $this->checkChannel($item['channel'])) !== true) {
                return $err;
            }
            if (!isset($item['config_type'])) {
                return "参数同步方式必填";
            }
            if (!in_array($item['config_type'], ['auto_sync', 'manual'])) {
                return "参数同步方式仅支持自动同步和手动设置";
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
            if (($err = $this->checkDistanceRange($item['distance_range'])) !== true) {
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
            'channel',
            'basic_fee',
            'base_delivery_fee',
            'rider_acceptance_timeout',
            'distance_range',
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
