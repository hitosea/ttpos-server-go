<?php

namespace app\common\model_old\payment;

use app\common\model_old\BaseModel;

/**
 * 支付订单
 */
class PaymentOrder extends BaseModel
{
    protected $name = 'payment_order';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    const PENDING = 0;
    const PAID = 1;
}
