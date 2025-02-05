<?php

namespace app\common\enum\order;

use MyCLabs\Enum\Enum;

/**
 * 订单错误类型枚举类,用于后期扩展，比如虚拟物品
 */
class OrderErrorEnum extends Enum
{
    // 商城订单
    const STOCK_ERROR = 102;                        // 库存不足
    const OUT_LIMIT_NUM = -11;                      // 商品超限购
    const OUT_LIMIT_TIME = -12;                     // 超时
    const OUT_ADD_MATERIAL_NUM = -13;               // 添加材料库存不足
    const OUT_SEND_MATERIAL_NUM = -14;              // 送厨材料库存不足
    const MISS_MUST_PRODUCT = -15;                  // 缺少必点商品提示
    const DIFF_PRICE_PRODUCT = -16;                 // 商品价格变动
    const OUT_PRODUCT_STOCK = -17;                  // 库存数量不足

    const RELOAD_PRICE = -20;                       // 订单价格变动
    const RESET_DISCOUNT_NOTICE = -21;              // 使用会员重置改价抹零提醒
    const ORDER_PAID = -22;                         // 订单已支付
    const ORDER_OVERAGE = -23;                      // 订单超额

    const TABLET_SEND_PRODUCT_NOT_FOUND = -30;      // 平板送厨商品不存在
    const TABLET_SEND_TIME_LIMIT = -31;             // 平板送厨时间未冷却
    const TABLET_SEND_NUM_LIMIT = -32;              // 平板送厨数量限制
    const TABLET_ORDER_PRICE_CHANGE = -33;          // 平板送厨数量限制
    const TABLET_ALREADY_SET = -34;                 // 平板送厨数量限制

    const BUFFET_TIME_OUT = -40;                    // 自助餐用餐时间超时
    const BUFFET_SEND_TIME_OUT = -41;               // 自助餐点餐(下单)时间超时

    const SCAN_EMPTY_CART = -50;                    // 扫码端下单商品为空

    const ORDER_REFUND_PROMPT_PAY_MISSING_PARAM = -901;           // 订单退款 QR PromptPay 提示银行卡信息
}
