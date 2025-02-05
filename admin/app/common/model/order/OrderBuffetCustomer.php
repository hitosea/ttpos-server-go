<?php

namespace app\common\model\order;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\buffet\Buffet;
use app\common\model\product\ProductSku;
use app\common\enum\order\OrderStatusEnum;
use app\common\model\order\Order as OrderModel;

/**
 * 订单自助餐顾客类型优惠
 */
class OrderBuffetCustomer extends BaseModel
{
    protected $name = 'order_buffet_customer';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'buffet_name_text',
        'customer_type_name_text',
        'consumption_tax_pay_price',
        'total_consumption_tax_pay_price',
        'total_consumption_tax_order_price',
    ];

    /**
     * 构造自助餐名名称
     * @param $value
     * @param $data
     * @return array|string
     */
    public function getBuffetNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['buffet_name']);
    }

    /**
     * 构造自助餐类型名称
     * @param $value
     * @param $data
     * @return array|string
     */
    public function getCustomerTypeNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['customer_type_name']);
    }

    /**
     * 自助餐商品最终[单价]应付（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getConsumptionTaxPayPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_pay_price'])
            && isset($data['num'])
            && isset($data['product_consumption_tax'])
            && isset($data['product_service_fee'])
            && isset($data['product_service_consumption_tax'])
        ) {
            $unit_pay_price = helper::bcdiv($data['total_pay_price'], $data['num'] ?: 1);                                           // 商品应付单价
            $unit_product_consumption_tax = helper::bcdiv($data['product_consumption_tax'], $data['num']);                                      // 商品消费税单价
            $unit_product_service_fee = helper::bcdiv($data['product_service_fee'], $data['num']);                                              // 商品服务费单价
            $unit_product_service_consumption_tax = helper::bcdiv($data['product_service_consumption_tax'], $data['num']);                      // 商品服务费消费税单价
            $product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($unit_pay_price, $unit_product_consumption_tax) : $unit_pay_price;     // 含税单价
            $product_price = helper::bcadd($product_price, helper::bcadd($unit_product_service_fee, $unit_product_service_consumption_tax));    // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
            $consumption_tax_pay_price = $product_price;
            return floatval($consumption_tax_pay_price);
        }
        return 0;
    }

    /**
     * 自助餐商品最终[总价]应付（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getTotalConsumptionTaxPayPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_pay_price'])
            && isset($data['num'])
            && isset($data['product_consumption_tax'])
            && isset($data['product_service_fee'])
            && isset($data['product_service_consumption_tax'])
        ) {
            $unit_pay_price = helper::bcdiv($data['total_pay_price'], $data['num'] ?: 1);                                           // 商品应付单价
            $unit_product_consumption_tax = helper::bcdiv($data['product_consumption_tax'], $data['num']);                                      // 商品消费税单价
            $unit_product_service_fee = helper::bcdiv($data['product_service_fee'], $data['num']);                                              // 商品服务费单价
            $unit_product_service_consumption_tax = helper::bcdiv($data['product_service_consumption_tax'], $data['num']);                      // 商品服务费消费税单价
            $product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($unit_pay_price, $unit_product_consumption_tax) : $unit_pay_price;     // 含税单价
            $product_price = helper::bcadd($product_price, helper::bcadd($unit_product_service_fee, $unit_product_service_consumption_tax));    // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
            $consumption_tax_pay_price = helper::bcmul($product_price, $data['num']);
            return floatval($consumption_tax_pay_price);
        }
        return 0;
    }

    /**
     * 商品最终[总价]原价（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getTotalConsumptionTaxOrderPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_price'])
            && isset($data['product_original_consumption_tax'])
            && isset($data['product_original_service_fee'])
            && isset($data['product_original_service_consumption_tax'])
        ) {
            $total_product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($data['total_price'], $data['product_original_consumption_tax']) : $data['total_price'];                 // 商品含税价;
            $total_final_price = helper::bcadd($total_product_price, helper::bcadd($data['product_original_service_fee'], $data['product_original_service_consumption_tax']));          // 最终总价
            return floatval($total_final_price);
        }
        return 0;
    }

    /**
     * 删除订单自助餐
     * @param $order_buffet_customer_id
     * @return false|mixed
     */
    public function del($order_buffet_customer_id)
    {
        $model = $this->where('id', '=', $order_buffet_customer_id)->find();
        if (!$model) {
            $this->error = '记录不存在';
            return false;
        }
        $buffet_id = $model['buffet_id'];
        // 检查订单状态
        $order = OrderModel::detail([
            ['order_id', '=', $model['order_id']],
            ['order_status', '=', OrderStatusEnum::NORMAL]
        ]);
        if (!$order) {
            $this->rollback();
            $this->error = '当前订单不可修改';
            return false;
        }
        if ($order['is_lock'] == 1) {
            $this->rollback();
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }
        $this->startTrans();
        try {
            $model->delete();
            // 判断自助餐完全删除
            if (!(new self)->where('buffet_id', $buffet_id)->where('order_id', $order['order_id'])->find()) {
                // 删除OrderBuffet
                (new OrderBuffet)->where('buffet_id', $buffet_id)->where('order_id', $order['order_id'])->delete();
                // 就餐剩余时间变化
                $time_limit = (new OrderBuffet())->where('order_id', '=', $order['order_id'])->min('time_limit');
                //
                if (!(new OrderBuffet())->where('order_id', '=', $order['order_id'])->find()) {
                    $order->save(['buffet_expired_time' => 0]);
                } else if ($time_limit > 0) {
                    $time_limit = floatval($time_limit);
                    $buffet_expired_time = helper::bcadd(strtotime($order['create_time']), helper::bcmul($time_limit, 60));
                    // 加上加钟的时间
                    $order_delay_time = (new OrderDelay)->where('order_id', $order['order_id'])->where('expired_time', 0)->sum('delay_time');
                    $buffet_expired_time = helper::bcadd($buffet_expired_time, $order_delay_time * 60);
                    $order->save(['buffet_expired_time' => $buffet_expired_time]);
                }
                // 已下单商品价格复原
                $buffet_product_ids_arr = Buffet::getBuffetProductIds([$buffet_id]);            // 被删除自助餐商品
                $now_buffet_ids = (new OrderBuffet)->where('order_id', $order['order_id'])->column('buffet_id');
                $now_buffet_product_ids_arr = Buffet::getBuffetProductIds($now_buffet_ids);        // 删除后自助餐商品
                $orderProductList = (new OrderProduct)->where('order_id', $order['order_id'])->select();
                foreach ($orderProductList as $orderProduct) {
                    if (in_array($orderProduct->product_id, $buffet_product_ids_arr) && !in_array($orderProduct->product_id, $now_buffet_product_ids_arr)) {
                        $product_price = (new ProductSku)->where('product_sku_id', $orderProduct->product_sku_id)->value('product_price');
                        $product_price = helper::bcadd($product_price, $orderProduct->feed_price);
                        $updateArr = [
                            'product_price' => $product_price,
                            'total_price' => $totalPrice = $orderProduct->total_num * $product_price,
                            'total_pay_price' => $totalPrice,
                            'is_buffet_product' => 0,
                        ];
                        $orderProduct->save($updateArr);
                    }
                }
            }
            (new OrderModel())->reloadPrice($order['order_id']);
            $this->commit();
            return $order['order_id'];
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 转移自助餐到子单
     * @param $id 自助餐ID
     * @param $fromOrderId 原订单ID
     * @param $toOrderId 目标订单ID
     * @param $num 数量
     * @return OrderBuffetCustomer
     */
    public static function addToSubOrder($id, $fromOrderId, $toOrderId, $num)
    {
        // 获取自助餐顾客
        $fromOrderBuffetCustomer = self::where([
            'id' => $id,
            'sub_order_id' => $fromOrderId,
        ])->find();

        if ($fromOrderBuffetCustomer->num == $num) {
            // 转移全部数量
            $fromOrderBuffetCustomer->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderBuffetCustomer, $num);
            $fromOrderBuffetCustomer->force()->delete();
        } else {
            // 转移部分数量
            // 从原子单中扣除部分数量
            $fromTotalNum = helper::bcsub((string)$fromOrderBuffetCustomer->num, (string)$num);
            $fromTotalPrice = helper::bcmul((string)$fromOrderBuffetCustomer->price, $fromTotalNum);
            $fromOrderBuffetCustomer->num = (int)$fromTotalNum;
            $fromOrderBuffetCustomer->total_price = (float)$fromTotalPrice;
            $fromOrderBuffetCustomer->total_pay_price = (float)$fromTotalPrice;
            $fromOrderBuffetCustomer->save();
            // 添加到目标子单
            $fromOrderBuffetCustomer->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderBuffetCustomer, $num);
        }

        return $fromOrderBuffetCustomer;
    }

    /**
     * 添加或增加自助餐顾客数量
     * @param $orderBuffetCustomer
     * @param $num
     * @return bool
     */
    public static function addOrIncNum($orderBuffetCustomer, $num)
    {
        $existsOrderBuffetCustomer = self::where([
            'order_id' => $orderBuffetCustomer->order_id,
            'sub_order_id' => $orderBuffetCustomer->sub_order_id,
            'buffet_customer_id' => $orderBuffetCustomer->buffet_customer_id,
            'buffet_id' => $orderBuffetCustomer->buffet_id,
            'customer_type_id' => $orderBuffetCustomer->customer_type_id,
        ])->find();
        if ($existsOrderBuffetCustomer) {
            $totalNum = helper::bcadd((string)$existsOrderBuffetCustomer->num, (string)$num);
            $totalPrice = helper::bcmul((string)$existsOrderBuffetCustomer->price, $totalNum);
            $totalPayPrice = helper::bcmul((string)$existsOrderBuffetCustomer->price, $totalNum);
            $existsOrderBuffetCustomer->save([
                'num' => (int)$totalNum,
                'total_price' => (float)$totalPrice,
                'total_pay_price' => (float)$totalPayPrice,
            ]);
        } else {
            $totalPrice = (float)helper::bcmul((string)$orderBuffetCustomer->price, (string)$num);
            $totalPayPrice = (float)helper::bcmul((string)$orderBuffetCustomer->price, (string)$num);
            $newOrderBuffetCustomer = new self();
            $newOrderBuffetCustomer->create([
                'order_id' => $orderBuffetCustomer->order_id,
                'sub_order_id' => $orderBuffetCustomer->sub_order_id,
                'buffet_customer_id' => $orderBuffetCustomer->buffet_customer_id,
                'buffet_id' => $orderBuffetCustomer->buffet_id,
                'customer_type_id' => $orderBuffetCustomer->customer_type_id,
                'buffet_name' => $orderBuffetCustomer->buffet_name,
                'customer_type_name' => $orderBuffetCustomer->customer_type_name,
                'price' => $orderBuffetCustomer->price,
                'num' => (int)$num,
                'total_price' => (float)$totalPrice,
                'refund_money' => $orderBuffetCustomer->refund_money,
                'refund_consumption_tax' => $orderBuffetCustomer->refund_consumption_tax,
                'refund_num' => $orderBuffetCustomer->refund_num,
                'total_pay_price' => (float)$totalPayPrice,
                'buffet_name' => $orderBuffetCustomer->buffet_name,
                'tax_rate' => $orderBuffetCustomer->tax_rate,
                'consumption_tax' => $orderBuffetCustomer->consumption_tax,
                'tax_calc_type' => $orderBuffetCustomer->tax_calc_type,
                'product_original_consumption_tax' => $orderBuffetCustomer->product_original_consumption_tax,
                'product_original_service_consumption_tax' => $orderBuffetCustomer->product_original_service_consumption_tax,
                'product_original_service_fee' => $orderBuffetCustomer->product_original_service_fee,
                'product_consumption_tax' => $orderBuffetCustomer->product_consumption_tax,
                'product_service_consumption_tax' => $orderBuffetCustomer->product_service_consumption_tax,
                'product_service_fee' => $orderBuffetCustomer->product_service_fee,
                'product_service_rate' => $orderBuffetCustomer->product_service_rate,
                'app_id' => $orderBuffetCustomer->app_id,
                'shop_supplier_id' => $orderBuffetCustomer->shop_supplier_id,
            ]);
        }
    }
}
