<?php

namespace app\common\enum\http;

use MyCLabs\Enum\Enum;

/**
 * 订单错误类型枚举类,用于后期扩展，比如虚拟物品
 */
class StatusCode extends Enum
{
    // 请求成功
    const SUCCESS = 1;

    // 请求失败
    const ERROR = 0;

    // FORCE_ERROR 前缀
    // 强制错误:  -10000 到 -19999 代表前端弹窗确认
    const FORCE_ERROR_SPLIT_ORDER = -10001;         // 存在拆单，不能操作

    // 当前登录用户被删除或者其他，需要回到登录页
    const USER_ERROR = -1;

    // TOKEN异常
    const TOKEN_ERROR = -2;

    // 桌台已关闭，请重新开台，返回桌台页
    const CLOSE_ERROR = -3;

    // 访问失效
    const VISIT_ERROR = -4;

    // 桌台用餐已关闭, 对应业务被关闭，需要自行处理
    const TABLE_ERROR = -5;

    // 桌台已结账
    const CHECKOUT_ERROR = -6;

    // 未开启功能
    const OPEN_ERROR = -20;

    // 设备上限
    const DEVICE_ERROR = -21;

    // 店铺到期
    const EXPIRE_ERROR = -102;

    // 设备已解绑，请重新绑定
    const UNBIND_ERROR = -201;

    // token异常，不需要重新登录，定格当前页面
    const TOKEN_ERROR_NOT_LOGIN = -202;

    // 产品错误：PRODUCT_ERROR 前缀
    // 产品错误:  -20001 到 -29999 代表有关产品提示
    const PRODUCT_ERROR_NOT_EXIST = -20001;         // 商品被删除或者已下架
    const PRODUCT_ERROR_NOT_EXIST_SKU = -20002;     // 商品规格被删除或者已下架
    const PRODUCT_ERROR_NOT_EXIST_FEED = -20003;    // 商品加料被删除或者已下架
}
