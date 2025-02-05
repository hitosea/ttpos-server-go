<?php

namespace app\common\model\order;

use app\common\model\BaseModel;

/**
 * 订单退款类型
 */
class OrderRefund extends BaseModel
{
    // 1-系统退款 2-线下退款 3-部分退款
    const SYS_REFUND = 1;
    const OFFLINE_REFUND = 2;
    const PARTIAL_REFUND = 3;
}
