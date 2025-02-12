<?php

namespace app\scan\model\order;

use app\common\library\helper;
use app\common\model\user\User as UserModel;
use app\common\enum\order\OrderStatusEnum;
use app\common\enum\settings\SettingEnum;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\order\Cart as CartModel;
use app\common\model\product\ProductSku as ProductSkuModel;
use app\scan\model\order\Order as OrderModel;
use app\scan\model\product\Product as ProductModel;
use app\cashier\model\user\CardRecord as CardRecordModel;

/**
 * 收银购物车模型
 */
class Cart extends CartModel
{
    protected $table = 'jjjfood_cashier_cart';
    /**
     * 隐藏字段
     * @var array
     */
    protected $hidden = [
        'app_id',
        'update_time'
    ];

    // 重新计算购物车价格信息
    public function reloadPrice($shop_supplier_id, $table_id, $order_id = 0)
    {
        if ($order_id > 0) {
            // 购物车商品列表
            $cartList = (new static())->with('product')
                ->where('order_id', '=', $order_id)
                ->select();
            // 是否存在订单
            $order = OrderModel::detail($order_id);
        } else {
            // 购物车商品列表
            $cartList = (new static())->with('product')
                ->where('table_id', '=', $table_id)
                ->select();
            // 是否存在订单
            $order = OrderModel::detail([
                ['table_id', '=', $table_id],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
        }
        $user_id = $order ? $order['user_id'] : 0;

        $pay_money = 0;
        $order_price = 0;
        $user_discount_money = 0;
        $user = UserModel::detail($user_id);

        foreach ($cartList as $product) {
            // 标记参与会员折扣
            $is_user_grade = false;
            // 会员等级抵扣的金额
            $grade_ratio = 0;
            // 会员折扣的商品单价
            $grade_product_price = 0;
            // 会员折扣的总额差
            $grade_total_money = 0;
            if ($product->product['is_enable_grade'] && $product['price'] > 0) {

                if ($user) {
                    $discount = (new CardRecordModel)->getDiscount($user['user_id']);
                } else {
                    $discount = 0;
                }
                $alone_grade_type = 10;
                // 商品单独设置了会员折扣
                if ($product['product']['is_alone_grade'] && isset($product['product']['alone_grade_equity'][$user['grade_id']])) {
                    if ($product['product']['alone_grade_type'] == 10) {
                        // 折扣比例
                        $discountRatio = helper::bcdiv($product['product']['alone_grade_equity'][$user['grade_id']], 100);
                    } else {
                        $alone_grade_type = 20;
                        $discountRatio = helper::bcdiv($product['product']['alone_grade_equity'][$user['grade_id']], $product['product_price'], 2);
                    }
                } else {
                    // 折扣比例
                    if ($user) {
                        $discountRatio = helper::bcdiv($user['grade']['equity'], 100);
                    } else {
                        $discountRatio = 1;
                    }
                }
                // 计算最终折扣
                if ($discount && $discountRatio) {
                    // 会员等级 * 会员卡
                    $discountRatio = round($discountRatio * $discount, 3);
                } elseif ($discount) {
                    // 会员卡
                    $discountRatio = $discount;
                }
                if ($discountRatio <= 1) {
                    if ($alone_grade_type == 20) {
                        // 固定金额
                        $grade_product_price = $product['alone_grade_equity'][$user['grade_id']];
                        $discount && $grade_product_price = round($grade_product_price * $discount, 2);
                    } else {
                        // 商品会员折扣后单价
                        $grade_product_price = helper::number2(helper::bcmul($product['price'], $discountRatio), true);
                    }
                    $gradeTotalPrice = $grade_product_price * $product['product_num'];
                    // 原商品总价 - 折扣后
                    $grade_total_money = helper::number2(helper::bcsub($product['price'] * $product['product_num'], $gradeTotalPrice));
                    $product['total_price'] = $gradeTotalPrice;
                }
            }

            $updateArr = [
                'user_id' => $user_id,
                'product_price' => $grade_product_price,   // 购物车价格（可能是原价或折扣的）
            ];
            $product->save($updateArr);

            // 累加
            $pay_money += $product['total_price'];  // 实付金额
            $order_price += $product['price'] * $product['product_num'];  // 商品原价
            $user_discount_money += $grade_total_money; // 商品优惠金额
        }

        $total_price = $pay_money; // 订单商品总价（不是商品原价总价、是商品折扣后(如果有)的总价）
        // 消费税计算
        $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $shop_supplier_id);
        $consume_fee = 0;
        if ($consumeFee['is_open']) {
            $consume_rate = $consumeFee['tax_rate'];
            $consume_fee = floatval(helper::bcmul($total_price, $consume_rate));
        }

        return [
            'cart_product_price' => $order_price,                                   // 购物车商品原价
            'cart_product_pay_price' => $pay_money,                                 // 购物车商品实付价钱
            'cart_pay_price' => floatval(helper::bcadd($pay_money, $consume_fee)),  // 购物车实付价钱
            'cart_consumption_tax_money' => $consume_fee,                           // 消费税
            'cart_user_discount_money' => $user_discount_money,                     // 会员折扣
        ];
    }

    //判断商品是否下架
    public function productState($product_id)
    {
        return (new ProductModel)->where('product_id', '=', $product_id)
            ->where('product_status', '=', 10)
            ->count();
    }

    //判断商品库存
    public function productStockState($product_id, $product_sku_id)
    {
        return (new ProductSkuModel)->where('product_id', '=', $product_id)
            ->where('product_sku_id', '=', $product_sku_id)
            ->where('stock_num', '>', 0)
            ->count();
    }
}
