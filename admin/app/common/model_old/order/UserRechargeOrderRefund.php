<?php

namespace app\common\model_old\order;

use app\common\model_old\BaseModel;

/**
 * 订单退款类型
 */
class UserRechargeOrderRefund extends BaseModel
{
    protected $name = 'user_recharge_order_refund';
    protected $pk = 'id';

    // 1-系统退款 2-线下退款 3-部分退款
    const SYS_REFUND = 1;
    const OFFLINE_REFUND = 2;
    const PARTIAL_REFUND = 3;
}
