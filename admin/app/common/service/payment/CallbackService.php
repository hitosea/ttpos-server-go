<?php

namespace app\common\service\payment;

use app\common\model\order\Order;
use app\common\service\BaseService;
use app\common\exception\BaseException;
use app\common\model\supplier\Supplier;
use app\common\model\payment\PaymentOrder;
use app\common\model\order\UserRechargeOrder;
use app\common\model\order\OrderRefundDestination;
use app\common\model\order\UserRechargeOrderRefundDestination;

/**
 * 回调处理
 */
class CallbackService extends BaseService
{

    /**
     * 处理LianLian回调
     * @param $params
     * @return bool|void
     * @throws BaseException
     * @throws \think\db\exception\DataNotFoundException
     * @throws \think\db\exception\DbException
     * @throws \think\db\exception\ModelNotFoundException
     */
    public function handleLianLianCallback($params)
    {

        /**
         * $params
          [
            'shop_supplier_id',
            'merchant_order_no',
            'merchant_user_id',
            'pay_type_desc',
            'pay_status',
            'payment_id',
            'order_amount',
            'order_currency',
            'pay_at',
          ]
         */

        //
        $shopSupplierId = $params['shop_supplier_id'] ?? 0;
        if (!$shopSupplierId) {
            trace('处理LianLian回调: shop_supplier_id不存在');
            throw new BaseException(['msg' => 'shop_supplier_id不存在']);
        }
        $supplier = (new Supplier([], 0))->setAppId(0)->field('shop_supplier_id, app_id')->find();
        if (empty($supplier) || empty($supplier['app_id'])) {
            trace('处理LianLian回调: supplier 不存在');
            throw new BaseException(['msg' => 'supplier 不存在']);
        }
        // 设置id
        request()->appId = $supplier['app_id'];
        request()->shopSupplierId = $supplier['shop_supplier_id'];
        //
        $paymentOrderModel = new PaymentOrder();
        $paymentOrder = $paymentOrderModel->where('merchant_order_no', $params['merchant_order_no'])->find();
        if (!$paymentOrder) {
            throw new BaseException(['msg' => '支付订单不存在']);
        }
        if ($paymentOrder->order_status == PaymentOrder::PAID) {
            return true;
        }
        $paymentOrderModel->startTrans();
        try {
            // 处理支付订单
            $updateArr = [
                'notify_result' => json_encode($params, JSON_UNESCAPED_UNICODE),
                'pay_time' => strtotime($params['pay_at']),
                'order_status' => PaymentOrder::PAID,
            ];
            $paymentOrder->save($updateArr);
            // 点餐订单
            if ($paymentOrder->order_id > 0) {
                /** @var Order $order */
                $orderModel = new Order();
                $order = $orderModel::where('order_id', $paymentOrder->order_id)->find();
                if (!$order) {
                    throw new BaseException(['msg' => '订单不存在']);
                }
                // 处理orderPayType
                $orderModel->addOnlinePayType($paymentOrder, $order->merge_parent_id);
                // 重算订单价格
                if ($order->merge_parent_id) {
                    $order->reloadMasterMergeOrder($order->merge_parent_id, $order->merge_id);
                } else {
                    $order->reloadPrice($order->order_id);
                }
            }
            // 会员充值订单
            else if ($paymentOrder->recharge_order_id > 0) {
                /** @var Order $order */
                $orderModel = new UserRechargeOrder();
                $order = $orderModel::where('id', $paymentOrder->recharge_order_id)->find();
                if (!$order) {
                    throw new BaseException(['msg' => '订单不存在']);
                }
                // 处理orderPayType
                (new Order)->addOnlinePayType($paymentOrder);
            }
            $paymentOrderModel->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $paymentOrderModel->rollback();
            return false;
        }
    }

    /**
     * 处理LianLian退款回调
     * @param $params
     * @return bool|void
     * @throws BaseException
     * @throws \think\db\exception\DataNotFoundException
     * @throws \think\db\exception\DbException
     * @throws \think\db\exception\ModelNotFoundException
     */
    public function handleLianLianRefundCallback($params)
    {
        /**
         * $params
        [
            'shop_supplier_id' => '1724054084',
            'refund_status' => 'RS',
            'refund_order_id' => '122024112744112499',
            'payment_order_id' => '1',
        ]
         */

        //
        $shopSupplierId = $params['shop_supplier_id'] ?? 0;
        if (!$shopSupplierId) {
            trace('处理LianLian回调: shop_supplier_id不存在');
            throw new BaseException(['msg' => 'shop_supplier_id不存在']);
        }
        $supplier = (new Supplier([], 0))->setAppId(0)->field('shop_supplier_id, app_id')->find();
        if (empty($supplier) || empty($supplier['app_id'])) {
            trace('处理LianLian回调: supplier 不存在');
            throw new BaseException(['msg' => 'supplier 不存在']);
        }

        // 设置id
        request()->appId = $supplier['app_id'];
        request()->shopSupplierId = $supplier['shop_supplier_id'];

        //
        $paymentOrderType = 1;
        if (isset($params['payment_order_id'])) {
            $paymentOrder = PaymentOrder::where('id', $params['payment_order_id'])->find();
            if (!$paymentOrder) {
                throw new BaseException(['msg' => '支付记录不存在']);
            }
            if ($paymentOrder->recharge_order_id > 0) {
                $paymentOrderType = 2;
            }
        }

        // 订单
        if ($paymentOrderType == 1) {
            $orderRefundDestination = OrderRefundDestination::where('refund_order_id', $params['refund_order_id'])->find();
        }
        // 充值订单
        else {
            $orderRefundDestination = UserRechargeOrderRefundDestination::where('refund_order_id', $params['refund_order_id'])->find();
        }

        // 处理状态
        if (!$orderRefundDestination) {
            trace("处理LianLian回调: 退款记录不存在{$params['refund_order_id']}不存在");
            throw new BaseException(['msg' => '退款记录不存在']);
        }
        if ($orderRefundDestination->status == 1 || $orderRefundDestination->status == -1) {
            return true;
        }
        $status = $params['refund_status'] == 'RS' ? 1 : -1;
        $orderRefundDestination->save(['status' => $status]);

        //
        return true;
    }
}
