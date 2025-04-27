<?php

namespace app\common\model_old\order;

use help\QueueHelp;
use think\facade\Cache;
use app\common\model_old\BaseModel;
use app\common\model_old\settings\Setting;
use app\common\model_old\order\OrderProduct;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model_old\order\OrderAbnormalLog;
use app\common\service\payment\RefundService;

/**
 * 订单操作日志
 */
class OrderOperationLog extends BaseModel
{
    protected $name = 'order_operation_log';
    protected $pk = 'id';

    // 操作行为
    const ACTION_OPEN_TABLE = 'OPEN_TABLE';             // 开台
    const ACTION_SEND_KITCHEN = 'SEND_KITCHEN';         // 送厨
    const ACTION_REFUND_PRODUCT = 'REFUND_PRODUCT';     // 退菜
    const ACTION_CANCEL_REFUND_PRODUCT = 'CANCEL_REFUND_PRODUCT';     // 取消退菜
    const ACTION_CHANGE_TABLE = 'CHANGE_TABLE';         // 转台
    const ACTION_CHANGE_PRICE = 'CHANGE_PRICE';         // 改价
    const ACTION_UPDATE_MEAL_NUM = 'UPDATE_MEAL_NUM';   // 修改桌台就餐人数
    const ACTION_STAY_ORDER = 'STAY_ORDER';             // 挂单
    const ACTION_PICK_ORDER = 'PICK_ORDER';             // 取单
    const ACTION_PRODUCT_FREE = 'PRODUCT_FREE';         // 赠菜
    const ACTION_CANCEL_PRODUCT_FREE = 'CANCEL_PRODUCT_FREE';         // 取消赠菜
    const ACTION_PRODUCT_MOVE = 'PRODUCT_MOVE';         // 转菜
    const ACTION_DISCOUNT = 'DISCOUNT';                 // 优惠折扣
    const ACTION_CANCEL_DISCOUNT = 'CANCEL_DISCOUNT';   // 撤销优惠折扣
    const ACTION_SETTLE = 'SETTLE';                     // 结账
    const ACTION_REVERSE_SETTLE = 'REVERSE_SETTLE';     // 反结账
    const ACTION_REFUND = 'REFUND';                     // 退款
    const ACTION_ORDER_TAKING = 'ORDER_TAKING';         // 接单
    const ACTION_ORDER_REJECT = 'ORDER_REJECT';         // 拒单
    const ACTION_MERGE_TABLE = 'MERGE_TABLE';           // 并台
    const ACTION_ORDER_CANCEL = 'ORDER_CANCEL';         // 整单取消
    const ACTION_CHECKOUT_DISCOUNT = 'CHECKOUT_DISCOUNT';                 // 结账手动抹零
    const ACTION_SPLIT_ORDER = 'SPLIT_ORDER';           // 拆单
    const ACTION_CANCEL_SPLIT_ORDER = 'CANCEL_SPLIT_ORDER';           // 撤销拆单

    /**
     * 获取行为文本
     */
    public static function getActionText($action, $data = [])
    {
        $texts = [
            self::ACTION_OPEN_TABLE => __('开台'),
            self::ACTION_SEND_KITCHEN => __('送厨'),
            self::ACTION_REFUND_PRODUCT => __('退菜'),
            self::ACTION_CANCEL_REFUND_PRODUCT => __('取消退菜'),
            self::ACTION_CHANGE_TABLE => __('转台'),
            self::ACTION_CHANGE_PRICE => __('改价'),
            self::ACTION_UPDATE_MEAL_NUM => __('人数'),
            self::ACTION_STAY_ORDER => __('挂单'),
            self::ACTION_PICK_ORDER => __('取单'),
            self::ACTION_PRODUCT_FREE => __('赠菜'),
            self::ACTION_CANCEL_PRODUCT_FREE => __('取消赠菜'),
            self::ACTION_PRODUCT_MOVE => __('转菜'),
            self::ACTION_DISCOUNT => __('优惠折扣'),
            self::ACTION_CANCEL_DISCOUNT => __('撤销优惠折扣'),
            self::ACTION_SETTLE => __('结账'),
            self::ACTION_REVERSE_SETTLE => __('反结账'),
            self::ACTION_REFUND => __('部分退款'),
            self::ACTION_ORDER_TAKING => __('接单'),
            self::ACTION_ORDER_REJECT => __('拒单'),
            self::ACTION_MERGE_TABLE => __('并台'),
            self::ACTION_ORDER_CANCEL => __('整单取消'),
            self::ACTION_CHECKOUT_DISCOUNT => __('结账手动抹零'),
            self::ACTION_SPLIT_ORDER => __('拆单'),
            self::ACTION_CANCEL_SPLIT_ORDER => __('撤销拆单'),
        ];
        // 拆单信息拼接
        $splitStr = '';
        if (($data['parent_id'] ?? 0) > 0) {
            $splitStr = '（' . __('拆单') . ($data['order_name'] ?? '') . '）';
        }
        //
        if ($action == self::ACTION_REFUND && ($data['refund_type'] ?? 1) == 1) {
            $content = __('整单退款');
            return $splitStr ? $splitStr . $content : $content;
        }
        //
        if ($action == self::ACTION_ORDER_TAKING && ($data['is_auto_order'] ?? false) == true) {
            return __('系统自动接单');
        }
        //
        $content = isset($texts[$action]) ? $texts[$action] : __('未知操作');
        return $splitStr ? $splitStr . $content : $content;
    }

    /**
     * 获取行为描述
     */
    public static function getActionDescription($item)
    {
        $shopSupplierId = $item['shop_supplier_id'] ?? 0;
        $appId = $item['app_id'] ?? 0;
        $unit = self::getCurrencyUnit($shopSupplierId, $appId);
        $desc = '';
        $data = $item['data'];
        switch ($item['action']) {
                // 开台
            case self::ACTION_OPEN_TABLE:
                if ($data) {
                    $desc = $data['table_no'] . ', ' . __('人数') . ': ' . $data['meal_num'];
                }
                break;
                // 并台
            case self::ACTION_MERGE_TABLE:
                if ($data) {
                    $desc = implode('、', $data);
                }
                break;
                // 送厨
            case self::ACTION_SEND_KITCHEN:
                $descs = [];
                foreach ($data ?: [] as $key => $val) {
                    $descs[] = extractLanguage($val['product_name']) . ' (' . (new OrderProduct)->getProductAttrAttr($val['product_attr'], false) . ') ' . '*' . $val['total_num'];
                }
                $desc = implode('、', $descs);
                break;
                // 退菜
            case self::ACTION_REFUND_PRODUCT:
                $reason = (new OrderProductReturn)->getReasons($data['custom_reason'] ?? '', $data['reason'] ?? '') ?: '';
                $desc = extractLanguage($data['product_name'] ?? '')
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'] ?? '', false) . ') ' . ('*' . $data['num'])
                    . ' (' . __('原因') . ': ' . $reason . ')';
                break;
            case self::ACTION_CANCEL_REFUND_PRODUCT:
                $desc = extractLanguage($data['product_name'] ?? '')
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'] ?? '', false) . ') ' . ('*' . $data['num']);
                break;
                // 转台
            case self::ACTION_CHANGE_TABLE:
                $desc = $data['old']['table_no'] . '->' . $data['new']['table_no'];
                break;
                // 改价
            case self::ACTION_CHANGE_PRICE:
                $desc = extractLanguage($data['product_name'])
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'], false) . ') ' . ('*' . ($data['total_num'] ?? 1))
                    . ' (' . $unit . $data['price'] . ')';
                break;
                // 修改桌台就餐人数
            case self::ACTION_UPDATE_MEAL_NUM:
                $desc = $data['new_meal_num'];
                break;
                // 赠菜
            case self::ACTION_PRODUCT_FREE:
                $desc = extractLanguage($data['product_name'])
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'], false) . ') ' . ('*' . $data['total_num'])
                    . ' (' . $unit . floatval($data['total_price']) . ')';
                break;
            case self::ACTION_CANCEL_PRODUCT_FREE:
                $desc = extractLanguage($data['product_name'])
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'], false) . ') ' . ('*' . $data['total_num'])
                    . ' (' . $unit . floatval($data['total_price']) . ')';
                break;
                // 转菜
            case self::ACTION_PRODUCT_MOVE:
                $desc = extractLanguage($data['product_name'])
                    . ' (' . (new OrderProduct)->getProductAttrAttr($data['product_attr'], false) . ') '
                    . '(' . __('转至') . $data['to_table_no'] . ')';
                break;
                // 优惠折扣
            case self::ACTION_DISCOUNT:
                if ($data['discount_type'] == 1) {
                    $desc = ($data['total_original_price'] ?: $data['old_price']) . __('改价') . $data['new_price'] . ' (' . $unit . floatval($data['special_discount']) . ')';
                }
                if ($data['discount_type'] == 2) {
                    $desc = __('折扣') . '-' . $data['rounding_rate'] . '% (' . $unit . floatval($data['special_discount'] ?? 0) . ')';
                }
                if ($data['discount_type'] == 3) {
                    $discountTypes = [
                        1 => __('抹分'),
                        2 => __('抹角'),
                        3 => __('四舍五入保留一位小数'),
                        4 => __('四舍五入到整数'),
                    ];
                    $desc = __('抹零') . '-' . ($discountTypes[$data['rounding_type']] ?? '') . ' (' . $unit . floatval($data['special_discount']) . ')';
                }
                if ($data['discount_type'] == 4) {
                    $desc = __('免单') . ' (' . $unit . floatval($data['price'] ?? 0) . ')';
                }
                break;
                // 结账
            case self::ACTION_SETTLE:
                $payTypeList = [];
                foreach ($data['pay_type'] as $payType) {
                    //
                    $payName = $payType['name'];
                    if ($payType['value'] == OrderPayTypeEnum::FREE_PAY) {
                        $payName = OrderPayTypeEnum::data($payType['value'])['name'] ?? $payType['name'];
                    }
                    if ($payType['value'] == OrderPayTypeEnum::CASH) {
                        $payType['price'] = $payType['price'] - $data['change_due'];
                    }
                    $payTypeList[] = $payName . ': ' . $unit . $payType['price'];
                }
                // 兼容数据
                $orderPrice = isset($data['order_price']) ? $data['order_price'] : $data['pay_price'];
                $desc = __('订单金额') . ' ' . $unit . floatval(($data['is_free'] ?? 0) > 0 ? $data['discount_money'] : $orderPrice) . '，'
                    . __('实付金额') . ' ' . $unit . floatval($data['pay_price']);
                if (!empty($payTypeList)) {
                    $desc .= ' (' . implode('、', $payTypeList) . ')';
                }
                break;
                // 反结账
            case self::ACTION_REVERSE_SETTLE:
                $payTypeList = [];
                foreach ($data['pay_type'] as $payType) {
                    $payName = $payType['name'];
                    if ($payType['value'] == OrderPayTypeEnum::FREE_PAY) {
                        $payName = OrderPayTypeEnum::data($payType['value'])['name'] ?? $payType['name'];
                    }
                    if ($payType['value'] == OrderPayTypeEnum::CASH) {
                        $payType['price'] = $payType['price'] - $data['change_due'];
                    }
                    $payTypeList[] = $payName . ': ' . $unit . $payType['price'];
                }
                $desc = implode('、', $payTypeList);
                break;
                // 退款
            case self::ACTION_REFUND:
                // 部分退
                if ($data['refund_type'] == 2) {
                    $productDesc = [];
                    foreach ($data['refund_product'] as $product) {
                        $productDesc[] = extractLanguage($product['product_name'])
                            . ' (' . (new OrderProduct)->getProductAttrAttr($product['product_attr'], false) . ') ' . ('*' . $product['refund_num']);
                    }
                    foreach ($data['refund_buffet'] as $buffet) {
                        $productDesc[] = extractLanguage($buffet['name'])
                            . (($buffet['customer_type_name'] ?? '')  ? ' (' . $buffet['customer_type_name'] . ') ' : '')
                            . ('*' . $buffet['refund_num']);
                    }
                    foreach ($data['refund_delay'] as $delay) {
                        $productDesc[] = extractLanguage($delay['name']) . ('*' . $delay['refund_num']);
                    }
                    $desc =  implode('、', $productDesc);
                }
                break;
                // 结账手动抹零
            case self::ACTION_CHECKOUT_DISCOUNT:
                $operation = $data['operation'] ?? '';
                $remark = $data['remark'] ?? '';
                $discountTypes = [
                    0 => __('实款实收'),
                    1 => __('抹分'),
                    2 => __('抹角'),
                    5 => __('抹元'),
                ];
                if ($operation == 'cancel') {
                    $desc = __('自动取消');
                    if (!empty($remark)) {
                        $desc .= ' (' . __('原因') . '：' . __($remark) . ')';
                    }
                } else {
                    $roundingTypeKey = $data['rounding_type'] ?? null;
                    $roundingType = $roundingTypeKey !== null ? ($discountTypes[$roundingTypeKey] ?? '') : '';
                    $specialDiscount = isset($data['special_discount']) ? ' (' . $unit . floatval($data['special_discount']) . ')' : '';
                    $desc = $roundingType . $specialDiscount;
                }
                break;
                // 拆单
            case self::ACTION_SPLIT_ORDER:
                $orderDetails = [];
                foreach ($data['split_order'] ?? [] as $item) {
                    $orderDetails[] = $item['order_name'] . '（' . __('订单金额') . '：' . $unit . $item['pay_price'] . '）';
                }
                $desc = implode(', ', $orderDetails);
                break;
        }
        return $desc;
    }

    /**
     * 获取当前货币单位
     */
    public static function getCurrencyUnit($shopSupplierId, $appId)
    {
        $cacheKey = "currency_unit_{$shopSupplierId}_{$appId}";
        $currency = Cache::get($cacheKey);
        if (!$currency) {
            $currency = Setting::getSupplierItem(SettingEnum::CURRENCY, $shopSupplierId, $appId);
            $currency = $currency['unit'] ?? '¥';
            Cache::tag('unit')->set($cacheKey, $currency, 60 * 60 * 24 * 1); // 1天
        }
        return $currency;
    }

    /**
     * 获取来源描述
     */
    public static function getSourceText($source)
    {
        $texts = [
            'cashier' => __('收银端'),
            'assistant' => __('点餐助手'),
            'shop' => __('商家后台'),
            'tablet' => __('平板端'),
            'scan' => __('扫码点餐'),
            'admin' => '-',
        ];
        return isset($texts[$source]) ? $texts[$source] : __('未知来源');
    }


    /**
     * 生成订单操作日志
     */
    public static function createLog($orderId, $action, $data, $remark = '')
    {
        $request = request();
        $appId = $request->appId;
        $userInfo = $request->userInfo;
        $userInfoId = $userInfo['assistant_user']['shop_user_id'] ?? $userInfo['user']['shop_user_id'] ?? $userInfo['shop_user_id'] ?? 0;
        $shopSupplierId = $request->shopSupplierId;
        $source = app('http')->getName();
        $remark = !empty($remark) ? $remark : '';
        // 兼容拆单
        $parentId = $data['parent_id'] ?? 0;
        $newOrderId = $parentId > 0 ? $parentId : $orderId;
        $subOrderId = $parentId > 0 ? $orderId : 0;
        //
        self::create([
            'app_id' => $appId,
            'order_id' => $newOrderId,
            'sub_order_id' => $subOrderId,
            'source' => $source,
            'shop_user_id' => $userInfoId,
            'shop_supplier_id' => $shopSupplierId,
            'action' => $action,
            'remark' => $remark,
            'data' => json_encode($data, JSON_UNESCAPED_UNICODE),
        ]);
        // 异常日志
        OrderAbnormalLog::createLog(OrderAbnormalLog::SOURCE_ORDER, $orderId, $action, $data, $remark);
    }

    /**
     *  获取订单操作日志
     */
    public static function getLogList($orderId, $batchNo = '')
    {
        $list = self::alias('a')
            ->field('a.*, user.real_name, user.user_name')
            ->leftJoin('shop_user user', 'a.shop_user_id = user.shop_user_id')
            ->where('a.order_id', $orderId)
            ->when($batchNo, function ($query) use ($batchNo) {
                $query->where('a.remark', $batchNo);
            })
            ->order('a.id', 'desc')
            ->select();
        $data = [];
        foreach ($list as $key => $value) {
            if ($value['data']) {
                $list[$key]['data'] = json_decode($value['data'], true);
            }
            //
            $desc = self::getActionDescription($value);
            // 支付方式
            $payTypeList = [];
            $refundType = 0;
            if ($value['action'] == self::ACTION_REFUND) {
                $refundType = $list[$key]['data']['refund_type'] ?? 0;
                foreach ($value['data']['pay_type'] as $payType) {
                    $payType['price'] = $payType['refund_money'];
                    $payName = $payType['name'] ?? '';
                    if ($payType['value'] == OrderPayTypeEnum::FREE_PAY) {
                        $payName = OrderPayTypeEnum::data($payType['value'])['name'] ?? $payName;
                    }
                    $payment_status = 1;
                    $refund_destination_id = 0;
                    $bank_code = '';
                    $account_no = '';
                    $account_name = '';
                    //
                    if ($payType['source'] == 2) {
                        $paymentFail = OrderRefundDestination::where('order_id', $payType['order_id'])->where('refund_id', $payType['refund_id'])->where('value', $payType['value'])->find();
                        if ($paymentFail) {
                            $payment_status = $paymentFail->status == -1 || (empty($paymentFail->refund_order_id) && $paymentFail->refund_order_id !== 0 && $paymentFail->refund_order_id !== '0') ? 0 : 1;
                            $refund_destination_id = $paymentFail->id;
                            $bank_code = $paymentFail->bank_code;
                            $account_no = $paymentFail->account_no;
                            $account_name = $paymentFail->account_name;
                        }
                    }
                    //
                    $payType['bank_code'] = $bank_code;
                    $payType['account_no'] = $account_no;
                    $payType['account_name'] = $account_name;
                    $payType['refund_destination_id'] = $refund_destination_id;
                    // payment_status 1-正常 0-不正常
                    $payType['payment_status'] = $payment_status;
                    $payType['unit'] = self::getCurrencyUnit($value['shop_supplier_id'], $value['app_id']); // 单位
                    $payTypeList[] = $payType;
                }
            }
            //
            $data[] = [
                'user_name' => ($value['source'] == 'admin') ? '-' : (($value['source'] == 'scan') ? __('用户') : ($value['real_name'] ?: '')),
                'user_email' => ($value['source'] == 'scan') ? '' : ($value['user_name'] ?: ''),
                'source' => self::getSourceText($value['source']),
                'create_time' => $value['create_time'],
                'description' => self::getActionText($value['action'], $value['data']) . ($desc || $payTypeList ? (': ' . $desc) : ''),
                'pay_type' => $payTypeList,
                'refund_type' => $refundType,
            ];
        }
        return $data;
    }

    /**
     * 操作重新退款
     * @param $payment_order_id
     * @param $refund_money
     * @param $refund_destination_id
     * @return bool
     */
    public function refundAgain($payment_order_id, $refund_money, $refund_destination_id, $bank_code = '', $account_no = '', $account_name = '')
    {
        if (!$payment_order_id || !$refund_money) {
            $this->error = '参数错误';
            return false;
        }
        //
        $orderRefundDestination = OrderRefundDestination::where('id', $refund_destination_id)->find();
        if (!$orderRefundDestination) {
            $this->error = '退款记录不存在';
            return false;
        }

        // 禁止并发操作
        $queue = new QueueHelp('CASHIER_REFUNDAGAIN_' . $orderRefundDestination->app_id . '-' . $orderRefundDestination->id);
        $queue->while();

        // 原退款信息
        $old_bank_code = $orderRefundDestination->bank_code;
        $old_account_no = $orderRefundDestination->account_no;
        $old_account_name = $orderRefundDestination->account_name;

        //
        $refundService = (new RefundService);
        //
        if ($orderRefundDestination->refund_order_id === 0) {
            $queue->release();
            $this->error = '该订单正在进行退款，无法重复操作';
            return false;
        }
        // 已发起过退款 - 验证退款状态
        if ($orderRefundDestination->refund_order_id) {
            $refund = $refundService->refund($payment_order_id, $refund_money, $orderRefundDestination->refund_order_id, $orderRefundDestination->merchant_refund_order_no, $old_bank_code, $old_account_no, $old_account_name);
            if (!$refund) {
                $this->error = $refundService->getError() ?: '请求支付服务失败';
                $queue->release();
                return false;
            }
            // 判断退款状态
            if (isset($refund['refund_status'])) {
                if ($refund['refund_status'] == 'RP') {
                    $queue->release();
                    $this->error = '该订单正在进行退款，无法重复操作';
                    return false;
                }
                if ($refund['refund_status'] == 'RS') {
                    $orderRefundDestination->save(['status' => 1]);
                    $queue->release();
                    $this->error = '该订单已成功退款，无法重复退款';
                    return false;
                }
            }
            // 更新银行信息 - 重新发起退款
            if (($bank_code && $bank_code != $old_bank_code) || ($account_no && $account_no != $old_account_no) || ($account_name && $account_name != $old_account_name)) {
                $merchant_refund_order_no = (new RefundService)->generateMerchantRefundOrderNo();    // 生成退款单号
                $orderRefundDestination->bank_code = $bank_code ?: $old_bank_code;
                $orderRefundDestination->account_no = $account_no ?: $old_account_no;
                $orderRefundDestination->account_name = $account_name ?: $old_account_name;
                $orderRefundDestination->status = 0;
                $orderRefundDestination->merchant_refund_order_no = $merchant_refund_order_no;
                //
                $refund = $refundService->refund($payment_order_id, $refund_money, '', $merchant_refund_order_no, $orderRefundDestination->bank_code, $orderRefundDestination->account_no, $orderRefundDestination->account_name);
                if (!$refund) {
                    $this->error = $refundService->getError() ?: '请求支付服务失败';
                    $queue->release();
                    return false;
                }
                //
                $orderRefundDestination->refund_order_id = $refund['refund_order_id'];
                $orderRefundDestination->save();
            }

            
        }
        // 更新银行信息 - 发起退款
        else {
            if (($bank_code && $bank_code != $old_bank_code) || ($account_no && $account_no != $old_account_no) || ($account_name && $account_name != $old_account_name)) {
                $orderRefundDestination->bank_code = $bank_code ?: $old_bank_code;
                $orderRefundDestination->account_no = $account_no ?: $old_account_no;
                $orderRefundDestination->account_name = $account_name ?: $old_account_name;
            }
            //
            $refund = $refundService->refund($payment_order_id, $refund_money, '', $orderRefundDestination->merchant_refund_order_no, $orderRefundDestination->bank_code, $orderRefundDestination->account_no, $orderRefundDestination->account_name);
            if (!$refund) {
                $this->error = $refundService->getError() ?: '请求支付服务失败';
                $queue->release();
                return false;
            }
            // 记录 refund_order_id 因为是新发的请求，不能存状态，因为刚发起肯定是处理中,记录refund_order_id就行
            $orderRefundDestination->refund_order_id = $refund['refund_order_id'];
            $orderRefundDestination->status = 0;
            $orderRefundDestination->save();
        }

        //
        $queue->release();
        return true;
    }
}
