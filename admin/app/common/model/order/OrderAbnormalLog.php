<?php

namespace app\common\model\order;

use app\common\model\BaseModel;

/**
 * 订单异常日志
 */
class OrderAbnormalLog extends BaseModel
{
    protected $name = 'order_abnormal_log';
    protected $pk = 'id';

    // 来源
    const SOURCE_ORDER = 'order'; // 订单
    const SOURCE_RECHARGE = 'recharge'; // 充值

    /**
     * 生成订单异常日志
     */
    public static function createLog($source, $orderId, $action, $data, $remark = '')
    {
        $request = request();
        $appId = $request->appId;
        $userInfo = $request->userInfo;
        $userInfoId = $userInfo['user']['shop_user_id'] ?? $userInfo['shop_user_id'] ?? 0;
        $dutyNo = $userInfo['user']['duty_no'] ?? ''; // 当班编号
        $shopSupplierId = $request->shopSupplierId;
        //
        $parentId = $data['parent_id'] ?? 0;
        $newOrderId = $parentId > 0 ? $parentId : $orderId;
        $subOrderId = $parentId > 0 ? $orderId : 0;
        //
        $insertData = [
            'source' => $source,
            'order_id' => $newOrderId,
            'sub_order_id' => $subOrderId,
            'shop_user_id' => $userInfoId,
            'duty_no' => $dutyNo,
            'action' => $action,
            'sub_action' => '',
            'remark' => $remark,
            'key' => '',
        ];
        // 唯一键值处理
        $key = md5(json_encode($data, JSON_UNESCAPED_UNICODE));
        // 处理自定义子行为
        $subAction = '';
        switch ($action) {
            case OrderOperationLog::ACTION_REFUND_PRODUCT:
                // 退菜 对一个商品反复操作，记录为1次
                $productId = $data['product_id'] ?? 0; // 商品id
                $productAttr = $data['product_attr'] ?? []; // 商品规格、属性、加料
                $reason = $data['reason'] ?? ''; // 退菜原因
                $customReason = $data['custom_reason'] ?? ''; // 自定义退菜原因
                $remark = $data['remark'] ?? ''; // 商品备注
                $key = md5(json_encode(['order_id' => $newOrderId, 'product_id' => $productId, 'product_attr' => $productAttr, 'reason' => $reason, 'custom_reason' => $customReason, 'remark' => $remark]));
                $info = self::where('order_id', $newOrderId)->where('action', $action)->where('key', $key)->find();
                if ($info) {
                    $info->save(['key' => $key]);
                } else {
                    self::create(array_merge($insertData, ['key' => $key]));
                }
                break;
            case OrderOperationLog::ACTION_CANCEL_REFUND_PRODUCT:
                // 取消退菜 重置所有所选赠菜操作
                $productId = $data['product_id'] ?? 0; // 商品id
                $productAttr = $data['product_attr'] ?? []; // 商品规格、属性、加料
                $reason = $data['reason'] ?? ''; // 退菜原因
                $customReason = $data['custom_reason'] ?? ''; // 自定义退菜原因
                $remark = $data['remark'] ?? ''; // 商品备注
                $key = md5(json_encode(['order_id' => $newOrderId, 'product_id' => $productId, 'product_attr' => $productAttr, 'reason' => $reason, 'custom_reason' => $customReason, 'remark' => $remark]));
                self::where('order_id', $newOrderId)->where('action', OrderOperationLog::ACTION_REFUND_PRODUCT)->where('key', $key)->delete();
                break;
            case OrderOperationLog::ACTION_PRODUCT_FREE:
                // 赠菜 对一个商品反复操作，记录为1次
                $orderProductId = $data['order_product_id'] ?? 0;
                $key = md5(json_encode(['order_id' => $newOrderId, 'order_product_id' => $orderProductId]));
                $info = self::where('order_id', $newOrderId)->where('action', $action)->where('key', $key)->find();
                if ($info) {
                    $info->save(['key' => $key]);
                } else {
                    self::create(array_merge($insertData, ['key' => $key]));
                }
                break;
            case OrderOperationLog::ACTION_CANCEL_PRODUCT_FREE:
                // 取消赠菜 重置所有所选赠菜操作
                $orderProductId = $data['order_product_id'] ?? 0;
                $key = md5(json_encode(['order_id' => $newOrderId, 'order_product_id' => $orderProductId]));
                self::where('order_id', $newOrderId)->where('action', OrderOperationLog::ACTION_PRODUCT_FREE)->where('key', $key)->delete();
                break;
            case OrderOperationLog::ACTION_CHANGE_PRICE:
                // 单品改价 对一个商品反复操作，记录为1次
                $orderProductId = $data['order_product_id'] ?? 0;
                $key = md5(json_encode(['order_id' => $newOrderId, 'order_product_id' => $orderProductId]));
                $info = self::where('order_id', $newOrderId)->where('action', $action)->where('key', $key)->find();
                if ($info) {
                    $info->save(['key' => $key]);
                } else {
                    self::create(array_merge($insertData, ['key' => $key]));
                }
                break;
            case OrderOperationLog::ACTION_DISCOUNT:
                // 优惠折扣 拆单优惠折扣在主单中重复只算一次，记录分开记，查询时再去重
                if (isset($data['discount_type'])) {
                    $subAction = $data['discount_type'];
                }
                // 查询该订单是否已经有优惠折扣
                if (in_array($subAction, [1, 2])) {
                    self::where('order_id', $newOrderId)
                        ->where('sub_order_id', $subOrderId)
                        ->where('action', $action)
                        ->delete();
                } elseif ($subAction == 3) {
                    self::where('order_id', $newOrderId)
                        ->where('sub_order_id', $subOrderId)
                        ->where('action', $action)
                        ->whereIn('sub_action', [1, 3])
                        ->delete();
                }
                $key = md5(json_encode(['order_id' => $newOrderId, 'sub_order_id' => $subOrderId, 'sub_action' => $subAction]));
                $info = self::where('order_id', $newOrderId)->where('action', $action)->where('key', $key)->find();
                if ($info) {
                    $info->save(['key' => $key]);
                } else {
                    self::create(array_merge($insertData, ['sub_action' => $subAction, 'key' => $key]));
                }
                break;
            case OrderOperationLog::ACTION_CANCEL_DISCOUNT:
                // 撤销优惠折扣 重置所有优惠折扣操作
                self::where('order_id', $newOrderId)->where('action', OrderOperationLog::ACTION_DISCOUNT)->delete();
                break;
            case OrderOperationLog::ACTION_CHECKOUT_DISCOUNT:
                // 结账抹零 对一个订单反复操作，记录为1次
                $key = md5(json_encode(['order_id' => $newOrderId]));
                $info = self::where('order_id', $newOrderId)->where('action', $action)->where('key', $key)->find();
                if ($info) {
                    $info->save(['key' => $key]);
                } else {
                    self::create(array_merge($insertData, ['key' => $key]));
                }
                break;
            case OrderOperationLog::ACTION_REVERSE_SETTLE:
                // 反结账 重置该订单免单操作
                $key = md5(json_encode(['order_id' => $newOrderId, 'sub_order_id' => $subOrderId]));
                self::where('order_id', $newOrderId)->where('sub_order_id', $subOrderId)->where('action', OrderOperationLog::ACTION_DISCOUNT)->where('sub_action', 4)->delete();
                self::create(array_merge($insertData, ['key' => $key]));
                break;
            default:
                if (in_array($action, [OrderOperationLog::ACTION_REFUND, OrderOperationLog::ACTION_PRODUCT_MOVE, UserRechargeOrderOperationLog::ACTION_REFUND, UserRechargeOrderOperationLog::ACTION_REVERSE_SETTLE])) {
                    self::create(array_merge($insertData, ['key' => $key]));
                }
                break;
        }
    }

    /**
     * 重置订单的优惠折扣操作
     */
    public static function resetDiscount($mainOrderId, $subOrderId, $subAction)
    {
        self::where('order_id', $mainOrderId)->where('sub_order_id', $subOrderId)->where('action', OrderOperationLog::ACTION_DISCOUNT)->whereIn('sub_action', $subAction)->delete();
    }
}
