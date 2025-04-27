<?php

namespace app\common\model_old\lianlianpay;

use app\common\model_old\BaseModel;

/**
 * 支付订单
 */
class PaymentOrder extends BaseModel
{
    protected $name = 'll_payment_order';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 添加checkout订单记录
     */
    public function add($payResponse, $payRequest, $sysParams, $type = 'payment')
    {
        if (self::where('order_id', $payResponse->order_id)->find()) {
            return true;
        }
        $saveArr = [
            'sys_order_id' => $sysParams['sys_order_id'],
            'sys_order_no' => $sysParams['sys_order_no'],
            'merchant_id' => $payRequest->merchant_id,
            'merchant_order_id' => $payRequest->merchant_order_id,
            'order_id' => $payResponse->order_id,
            'order_status' => $payResponse->order_status,
            'order_type' => $type,
            'order_amount' => $payRequest->order_amount,
            'order_currency' => $payRequest->order_currency,
            'full_name' => $payRequest->customer->full_name,
            'merchant_user_id' => $payRequest->customer->merchant_user_id,
            'order_desc' => $payRequest->order_desc,
            'redirect_url' => $payRequest->notify_url,
            'notify_url' => $payRequest->redirect_url,
            'request_params' => json_encode(get_object_vars($payRequest)),
            'link_url' => $payResponse->link_url,
            'll_create_time' => $payResponse->create_time,
        ];
        return $this->save($saveArr);
    }

    /**
     * 添加checkout订单记录
     */
    public function addWechatPay($payResponse, $payRequest, $sysParams, $type = 'wechat_pay')
    {
        if (self::where('order_id', $payResponse->order_id)->find()) {
            return true;
        }
        $saveArr = [
            'sys_order_id' => $sysParams['sys_order_id'],
            'sys_order_no' => $sysParams['sys_order_no'],
            'merchant_id' => $payRequest->merchant_id,
            'merchant_order_id' => $payRequest->merchant_order_id,
            'order_id' => $payResponse->order_id,
            'order_status' => $payResponse->order_status,
            'order_type' => $type,
            'order_amount' => $payRequest->order_amount,
            'order_currency' => $payRequest->order_currency,
            'full_name' => $payRequest->customer->full_name,
            'merchant_user_id' => $payRequest->customer->merchant_user_id,
            'order_desc' => $payRequest->order_info,
            'redirect_url' => $payRequest->notify_url,
            'notify_url' => $payRequest->redirect_url,
            'request_params' => json_encode(get_object_vars($payRequest)),
            'link_url' => $payResponse->link_url,
            'll_create_time' => $payResponse->create_time,
        ];
        return $this->save($saveArr);
    }
}
