<?php

namespace app\common\model_old\order;

use app\common\model_old\BaseModel;
use app\common\model_old\order\OrderProduct;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model_old\order\UserRechargeOrderRefundDestination;

/**
 * 用户充值订单操作日志
 */
class UserRechargeOrderOperationLog extends BaseModel
{
    protected $name = 'user_recharge_order_operation_log';
    protected $pk = 'id';

    // 操作行为
    const ACTION_GENERATE_ORDER = 'GENERATE_ORDER';          // 生成订单
    const ACTION_CHANGE_AMOUNT = 'CHANGE_AMOUNT';           // 变更充值金额
    const ACTION_ORDER_CANCEL = 'ORDER_CANCEL';             // 取消
    const ACTION_RECHARGE = 'RECHARGE';                     // 充值
    const ACTION_REVERSE_SETTLE = 'REVERSE_SETTLE';         // 反结账
    const ACTION_REFUND = 'REFUND';                         // 退款

    /**
     * 获取行为文本
     */
    public static function getActionText($action, $data = [])
    {
        $texts = [
            self::ACTION_GENERATE_ORDER => __('生成订单'),
            self::ACTION_CHANGE_AMOUNT => __('变更充值金额'),
            self::ACTION_ORDER_CANCEL => __('取消'),
            self::ACTION_RECHARGE => __('充值'),
            self::ACTION_REVERSE_SETTLE => __('反结账'),
            self::ACTION_REFUND => __('部分退款'),
        ];
        //
        if ($action == self::ACTION_REFUND && ($data['refund_type'] ?? 1) == 1) {
            return __('整单退款');
        }
        //
        return isset($texts[$action]) ? $texts[$action] : __('未知操作');
    }

    /**
     * todo 获取行为描述
     */
    public static function getActionDescription($item)
    {
        $desc = '';
        $data = $item['data'];
        switch ($item['action']) {
                // 变更充值金额
            case self::ACTION_CHANGE_AMOUNT:
                $desc =  ($data['old_recharge_money'] ?? 0 ). ' ' . __('变更为') . ' ' . ($data['recharge_money'] ?? 0);
                break;
                // 充值
            case self::ACTION_RECHARGE:
                $payTypeList = [];
                foreach (($data['pay_type'] ?? []) as $payType) {
                    $payName = $payType['name'];
                    if ($payType['value'] == OrderPayTypeEnum::FREE_PAY) {
                        $payName = OrderPayTypeEnum::data($payType['value'])['name'] ?? $payType['name'];
                    }
                    if ($payType['value'] == OrderPayTypeEnum::CASH) {
                        $payType['price'] = $payType['price'] - ($data['change_due'] ?? 0);
                    }
                    $payTypeList[] = $payName . ': ' . __('¥') . $payType['price'];
                }
                $desc = __('订单金额') . ' ' . __('¥') . ($data['recharge_money'] ?? 0 ). '，'
                    . __('实付金额') . ' ' . __('¥') . floatval($data['pay_price'] ?? 0);
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
                    $payTypeList[] = $payName . ': ' . __('¥') . $payType['price'];
                }
                $desc = implode('、', $payTypeList);
                break;
                // 退款
            case self::ACTION_REFUND:
                // 部分退
                if ($data['refund_type'] == 2) {
                    $productDesc = [];
                    $desc =  implode('、', $productDesc);
                }
                break;
        }
        return $desc;
    }

    /**
     * todo 获取来源描述
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
        //
        self::create([
            'app_id' => $appId,
            'order_id' => $orderId,
            'source' => $source,
            'shop_user_id' => $userInfoId,
            'shop_supplier_id' => $shopSupplierId,
            'action' => $action,
            'remark' => $remark,
            'data' => json_encode($data, JSON_UNESCAPED_UNICODE),
        ]);
        // 异常日志
        OrderAbnormalLog::createLog(OrderAbnormalLog::SOURCE_RECHARGE, $orderId, $action, $data, $remark);
    }

    /**
     * todo 生成订单操作日志
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
            if ($value['action'] == self::ACTION_REFUND) {
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
                    if ($payType['source'] == 2) {
                        $paymentFail = UserRechargeOrderRefundDestination::where('order_id', $payType['order_id'])->where('refund_id', $payType['refund_id'])->where('value', $payType['value'])->find();
                        if ($paymentFail) {
                            $payment_status = $paymentFail->status == -1 || empty($paymentFail->refund_order_id) ? 0 : 1;
                            $refund_destination_id = $paymentFail->id;
                            $bank_code = $paymentFail->bank_code;
                            $account_no = $paymentFail->account_no;
                            $account_name = $paymentFail->account_name;
                        }
                    }
                    $payType['bank_code'] = $bank_code;
                    $payType['account_no'] = $account_no;
                    $payType['account_name'] = $account_name;
                    $payType['refund_destination_id'] = $refund_destination_id;
                    // payment_status 1-正常 0-不正常
                    $payType['payment_status'] = $payment_status;
                    $payType['unit'] = __('¥');
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
            ];
        }
        return $data;
    }
}
