<?php

namespace app\common\model_old\order;

use app\common\model_old\BaseModel;
use app\common\model_old\store\PayType;
use app\common\enum\order\OrderPayTypeEnum;

/**
 * 订单支付类型
 */
class OrderPayType extends BaseModel
{
    protected $name = 'order_pay_type';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name',
        'source_text'
    ];

    /**
     * 支付方式
     * @param $value
     * @return array
     */
    public function getPayTypeAttr($value, $data)
    {
        $result = [
            'text' => $this->getNameAttr($data['value'], $data),
            'value' => $data['value']
        ];
        return $result;
    }

    /**
     * 支付方式名称
     * @param $value
     * @param $data
     * @return array|mixed|string
     */
    public function getNameAttr($value, $data)
    {
        $v = $data['value'] ?? '-';
        if ($v == OrderPayTypeEnum::FREE_PAY) {
            $text = OrderPayTypeEnum::data($v)['name'] ?? '-';
        } else {
            if (isset($data['name'])) {
                $text = $data['name'];
                $this->remark = $data['remark'] ?? $text;
            } else {
                $payType = PayType::where('value', $v)->withTrashed()->find();
                if ($payType) {
                    $text = $payType->name;
                    $this->remark = $payType->remark ?: $text;
                } else {
                    $text = OrderPayTypeEnum::data($v)['name'] ?? '-';
                    $this->remark = $text;
                }
            }
        }
        return $text;
    }

    /**
     * 备注
     */
    public function getRemarkAttr($value, $data)
    {
        $remark = $value ?? $data['remark'] ?? '-';
        return __($remark);
    }

    /**
     * 渠道获取
     */
    public function getSourceTextAttr($value, $data)
    {
        if (isset($data['source'])) {
            return __(PayType::SOURCE[$data['source']] ?? '-');
        }
        $this->source = $source = PayType::where('value', $data['value'])->withTrashed()->cache(2)->value('source');
        return __(PayType::SOURCE[$source] ?? '-');
    }

    /**
     * 手续费
     * @param $value
     * @param $data
     * @return array|mixed|string
     */
    public function getFeeMoneyAttr($value)
    {
        return floatval($value);
    }

    /**
     * 支付金额
     * @param $value
     * @param $data
     * @return array|mixed|string
     */
    public function getPriceAttr($value)
    {
        return floatval($value);
    }

    /**
     * 生成唯一支付hash
     * @param $order_id
     * @param $value
     * @param $payment_order_id
     * @return false|string
     */
    public static function generateUniqueHash($order_id, $value, $payment_order_id = 0) {
        $combinedString = $order_id . '-' . $payment_order_id . '-' . $value;
        return hash('sha256', $combinedString);
    }
}
