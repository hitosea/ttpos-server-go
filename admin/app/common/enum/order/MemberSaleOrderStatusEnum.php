<?php

namespace app\common\enum\order;

use MyCLabs\Enum\Enum;

/**
 * 订单错误类型枚举类,用于后期扩展，比如虚拟物品
 */
class MemberSaleOrderStatusEnum extends Enum
{
    // '订单状态 0-选购中 1-待支付 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消',
    const FINISHED = 7;
}
