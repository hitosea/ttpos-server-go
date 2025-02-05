<?php

namespace app\common\model\order;

use app\common\tasks\Task;
use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\tasks\RefundTask;
use app\common\model\store\PayType;
use app\common\exception\BaseException;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\service\payment\RefundService;
use app\common\service\order\OrderRefundService;
use app\common\model\order\UserRechargeOrderRefund;

/**
 * 订单退款目的地表
 */
class UserRechargeOrderRefundDestination extends BaseModel
{
    protected $name = 'user_recharge_order_refund_destination';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name',
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
            } else {
                $payType = PayType::where('value', $v)->withTrashed()->find();
                if ($payType) {
                    $text = $payType->name;
                } else {
                    $text = OrderPayTypeEnum::data($v)['name'] ?? '-';
                }
            }
        }
        return $text;
    }

    /**
     * 添加退款记录目的地
     * @param $value
     * @param $data
     * @return array|mixed|string
     */
    public static function createData($order, $orderRefundLog, $refund_type, $refund_method, $cashier_id, $bank_code = '', $account_no = '', $account_name = '')
    {
        $needRefundMoney = $orderRefundLog['refund_money'];
        $refundDestination = array_column($order->refundDestinations->toArray(), null, 'value');

        // 退款分配
        $lianlianPay = [];
        $cashRefundMoney = 0;
        $orderProductReturnList = [];
        foreach ($order['refundPayType'] as $pay) {
            $payTypeCellRefundMoney = $pay['price'];
            if (isset($refundDestination[(string)$pay['value']])) {
                $payTypeCellRefundMoney = helper::bcsub($payTypeCellRefundMoney, $refundDestination[$pay['value']]['refund_money']);
            }
            // 当前支付方式足额退 - 则跳过
            if ($payTypeCellRefundMoney <= 0) {
                continue;
            }

            // 记录
            $needRefundMoneys = helper::bcsub($needRefundMoney, $payTypeCellRefundMoney);
            $payTypeRefundMoney = $needRefundMoneys <= 0 ? $needRefundMoney : $payTypeCellRefundMoney;
            $orderProductReturnList[] = $saveData = [
                'order_id' => $order['id'],
                'refund_id' => $orderRefundLog['id'],
                'price' => $pay['price'],
                'value' => $pay['value'],
                'name' => $pay['name'],
                'remark' => $pay['remark'] ?? '',
                'source' => $pay['source'],
                'payment_order_id' => $pay['payment_order_id'] ?? 0,
                'refund_money' => $payTypeRefundMoney,
                'shop_supplier_id' => $order['shop_supplier_id'],
                'app_id' => $order['app_id'],
                'status' => in_array($pay['value'], array_keys(PayType::SOURCE_LIANLIAN_PAY_METHOD)) ? 0 : 1,
            ];
            $refundDestination = self::create($saveData);
            // 各种支付方式需要单独处理的业务
            if ($pay['value'] == OrderPayTypeEnum::CASH) {
                $cashRefundMoney += $payTypeRefundMoney;
            }
            // 余额支付
            else if ($pay['value'] == OrderPayTypeEnum::BALANCE) {
                // 可用退款余额
                $available_pay_price = $pay['price'];
                // 1.回退用户余额
                (new OrderRefundService)->balanceRefund($order, $available_pay_price);
                //
                $saveData = [
                    'order_id' => $order['id'],
                    'refund_type' => $refund_type,
                    'refund_method' => $refund_method,
                    'refund_money' => $available_pay_price,
                    'shop_supplier_id' => $order['shop_supplier_id'],
                    'app_id' => $order['app_id'],
                    'cashier_id' => $cashier_id,
                ];
                $orderRefundLog = UserRechargeOrderRefund::create($saveData);
            }
            // lianlianpay 支付
            else if (in_array($pay['value'], array_keys(PayType::SOURCE_LIANLIAN_PAY_METHOD))) {
                // QR PromptPay，银行卡信息必填
                if ($pay['value'] == OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY) {
                    if (!$bank_code) {
                        throw new BaseException(['msg' => '请选择银行', 'code' => OrderErrorEnum::ORDER_REFUND_PROMPT_PAY_MISSING_PARAM]);
                    }
                    if (!$account_no) {
                        throw new BaseException(['msg' => '请输入银行卡号', 'code' => OrderErrorEnum::ORDER_REFUND_PROMPT_PAY_MISSING_PARAM]);
                    }
                    if (!$account_name) {
                        throw new BaseException(['msg' => '请输入银行卡开户人名称', 'code' => OrderErrorEnum::ORDER_REFUND_PROMPT_PAY_MISSING_PARAM]);
                    }
                    $refundDestination->bank_code = $bank_code;
                    $refundDestination->account_no = $account_no;
                    $refundDestination->account_name = $account_name;
                }
                //
                $refundDestination->merchant_refund_order_no = $merchant_refund_order_no = (new RefundService)->generateMerchantRefundOrderNo();    // 生成退款单号
                $refundDestination->save();
                //
                $lianlianPay[] = [
                    'refundDestination' => $refundDestination,
                    'order_refund_destination_id' => $refundDestination->id,
                    'refund_money' => $payTypeRefundMoney,
                    'payment_order_id' => $pay['payment_order_id'],
                    'value' => $pay['value'],
                    'bank_code' => $bank_code,
                    'account_no' => $account_no,
                    'account_name' => $account_name,
                    'merchant_refund_order_no' => $merchant_refund_order_no
                ];
            }

            // 已分配完成 - 跳出
            $needRefundMoney = $needRefundMoneys;
            if ($needRefundMoney <= 0) {
                break;
            }
        }

        // 请求lianlian退款 -  如果大于1种使用异步方式请求
        if (!empty($lianlianPay)) {
            $refundService = (new RefundService);
            if (count($lianlianPay) > 1) {
                foreach ($lianlianPay as $pay_item) {
                    if ($pay_item['value'] == OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY) {
                        Task::deliver(new RefundTask($order['shop_supplier_id'], $pay_item['payment_order_id'], $pay_item['refund_money'], $pay_item['order_refund_destination_id'], $pay_item['merchant_refund_order_no'], $pay_item['bank_code'], $pay_item['account_no'], $pay_item['account_name']));
                    } else {
                        // 记录 银行卡信息
                        Task::deliver(new RefundTask($order['shop_supplier_id'], $pay_item['payment_order_id'], $pay_item['refund_money'], $pay_item['order_refund_destination_id'], $pay_item['merchant_refund_order_no']));
                    }
                }
            } else {
                if ($lianlianPay[0]['value'] == OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY) {
                    $result = $refundService->refund($lianlianPay[0]['payment_order_id'], $lianlianPay[0]['refund_money'], '', $lianlianPay[0]['merchant_refund_order_no'], $lianlianPay[0]['bank_code'], $lianlianPay[0]['account_no'], $lianlianPay[0]['account_name']);
                } else {
                    $result = $refundService->refund($lianlianPay[0]['payment_order_id'], $lianlianPay[0]['refund_money'], '', $lianlianPay[0]['merchant_refund_order_no']);
                }
                if (!$result) {
                    throw new BaseException(['msg' => $refundService->getError() ?: '请求支付服务失败']);
                }
                // 记录 refund_order_id
                self::where('id', $lianlianPay[0]['order_refund_destination_id'])->update(['refund_order_id' => $result['refund_order_id']]);
            }
        }

        //
        return compact('lianlianPay', 'cashRefundMoney', 'orderProductReturnList');
    }
}
