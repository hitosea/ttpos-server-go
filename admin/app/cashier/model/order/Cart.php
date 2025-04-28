<?php

namespace app\cashier\model\order;

use think\facade\Cache;
use app\common\library\helper;
use app\common\model\order\Order;
use app\common\enum\http\StatusCode;
use app\common\model\product\Product;
use app\common\model\order\OrderDelay;
use app\common\model\order\OrderPayType;
use app\common\model\order\OrderProduct;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderSourceEnum;
use app\common\enum\order\OrderStatusEnum;
use app\cashier\model\user\User as UserModel;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\Cart as CartModel;
use app\cashier\model\order\Order as OrderModel;
use app\common\model\order\Order as CommonOrderModel;
use app\cashier\model\product\Product as ProductModel;
use app\common\model\settings\Setting as SettingModel;
use app\cashier\model\user\CardRecord as CardRecordModel;
use app\cashier\model\product\ProductSku as ProductSkuModel;

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

    /**
     * 获取当前购物车列表 (含商品信息)
     */
    public function getList($data, $eat_type = 20, $table_id = 0)
    {
        // 获取购物车商品列表
        $model = $this;
        if ($table_id) {
            $model = $model->where('table_id', '=', $table_id);
        }
        if ($eat_type == 10) {
            $model = $model->where(function ($query) use ($data) {
                $query->where('c.cashier_id', '=', 0)->whereOr('c.cashier_id', '=', $data['cashier_id']);
            });
        } else {
            $model = $model->where('c.cashier_id', '=', $data['cashier_id']);
        }
        $list = $model->alias('c')
            ->with(['product'])
            ->where('c.shop_supplier_id', '=', $data['shop_supplier_id'])
            ->where('is_stay', '=', 0)
            ->where('eat_type', '=', $eat_type)
            ->field('c.*')
            ->select();
        return $list;
    }

    //查询桌号订单信息
    public static function getTableCartInfo($user, $table_id)
    {
        return (new static())->with('product')
            ->where('cashier_id', '=', $user['cashier_id'])
            ->where('table_id', '=', $table_id)
            ->select();
    }

    /**
     * 删除商品
     */
    public function delProduct($cart_id)
    {
        return $this->where('cart_id', '=', $cart_id)->delete();
    }

    /**
     * 是否已生成订单（已送厨）
     *
     * @param string $cart_no
     * @return int
     */
    public function checkOrderByCardNo($cart_no)
    {
        $order_id = $this->where('cart_no', '=', $cart_no)
            ->where('order_id', '>', 0)
            ->value('order_id');
        return $order_id ?: 0;
    }

    /**
     * 整单取消
     *
     * @param string $cart_no
     * @return bool
     */
    public function delStay($cart_no = '')
    {
        $query = $this->where('eat_type', '=', 20);
        if ($cart_no) {
            $query = $query->where('cart_no', '=', $cart_no);
        } else {
            $query = $query->where('is_stay', '=', 0);
        }
        // 是否有订单（已送厨）
        if ($order_id = $this->checkOrderByCardNo($cart_no)) {
            // 把送厨的订单删除
            (new OrderProduct)->where('order_id', '=', $order_id)->delete();
            // 把订单取消
            $detail = CommonOrderModel::detail($order_id);
            $detail?->cancels();
        }
        return $query->delete();
    }

    /**
     * 挂单数量
     */
    public function stayNum($user)
    {
        return $this->alias('c')
            ->where('c.shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('c.cashier_id', '=', $user['cashier_id'])
            ->where('is_stay', '=', 1)
            ->where('eat_type', '=', 20)
            ->group('cart_no')
            ->count();
    }

    /**
     * 购物车列表 (含商品信息)
     */
    public function getCartList($user, $table_id = 0, $eat_type = 20)
    {
        // 获取购物车商品列表
        $model = $this;
        if ($table_id) {
            $model = $model->where('table_id', '=', $table_id);
        }
        if ($eat_type == 10) {
            $model = $model->where(function ($query) use ($user) {
                $query->where('cashier_id', '=', 0)->whereOr('cashier_id', '=', $user['cashier_id']);
            });
        } else {
            $model = $model->where('cashier_id', '=', $user['cashier_id']);
        }
        $list = $model->with(['product', 'sku', 'image.file'])
            ->field("*,(bag_price*product_num) as total_bag_price,(product_price*product_num) as line_money")
            ->where('eat_type', '=', $eat_type)
            ->where('shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('is_stay', '=', 0)
            ->select();
        if ($list) {
            foreach ($list as $item) {
                $item['total_price'] = $item['price'] * $item['product_num'];
                $item['is_points_gift'] = $item['product']['is_points_gift'];
            }
        }
        return $list;
    }

    /**
     * 获取挂单列表
     */
    public function getStayList($user)
    {
        $list = $this->alias('c')
            ->where('c.shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('c.cashier_id', '=', $user['cashier_id'])
            ->where('is_stay', '=', 1)
            ->where('eat_type', '=', 20)
            ->group('cart_no')
            ->order('stay_time asc')
            ->field('c.*')
            ->select();
        foreach ($list as &$item) {
            $item['stay_time'] = date('Y-m-d H:i:s', $item['stay_time']);
            $item['total_price'] = $item['price'] * $item['product_num'];
            $item['product'] = $this->alias('c')
                ->with(['product', 'image.file'])
                ->where('c.shop_supplier_id', '=', $user['shop_supplier_id'])
                ->where('c.cashier_id', '=', $user['cashier_id'])
                ->where('is_stay', '=', 1)
                ->where('cart_no', '=', $item['cart_no'])
                ->field('c.*')
                ->select();
        }
        return $list;
    }

    /**
     * 挂单
     */
    public function stayCart($user, $order_id = 0)
    {
        // 获取当前购物车商品列表
        $model = $this;
        $cartIds = $model->alias('c')
            ->where('c.shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('c.cashier_id', '=', $user['cashier_id'])
            ->where('is_stay', '=', 0)
            ->where('eat_type', '=', 20)
            ->column('cart_id');
        if (count($cartIds) == 0) {
            $this->error = "暂无商品挂单";
            return false;
        }
        $data['is_stay'] = 1;
        $cartInfo = $model->where('cart_id', 'in', $cartIds)->find();
        if (!$cartInfo['cart_no']) {
            $data['cart_no'] = date('YmdHis');
            $data['stay_time'] = time();
        }
        // 送厨挂单
        if ($order_id > 0) {
            $data['order_id'] = $order_id;
        }
        return $this->where('cart_id', 'in', $cartIds)->update($data);
    }

    /**
     * 取单
     */
    public function pickCart($cart_no, $user)
    {
        $count = $this->where('shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('cashier_id', '=', $user['cashier_id'])
            ->where('eat_type', '=', 20)
            ->where('is_stay', '=', 0)
            ->count();
        if ($count > 0) {
            $this->error = "购物车内存在商品,请先结账或者挂单后再取单";
            return false;
        }
        return $this->where('cart_no', 'in', $cart_no)->update(['is_stay' => 0]);
    }

    /**
     * 删单
     */
    public function delCart($cart_no)
    {
        return $this->where('cart_no', '=', $cart_no)->delete();
    }

    /**
     * 加入购物车
     */
    public function add($data, $user)
    {
        //判断商品是否下架
        $product = $this->productState($data['product_id']);
        if (!$product) {
            $this->error = '商品已下架';
            $this->errorData = ['product_id' => $data['product_id']];
            $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST;
            return false;
        }
        $stockStatus = $this->productStockState($data['product_id'], $data['product_sku_id']);
        if (!$stockStatus) {
            $this->error = '商品库存不足';
            return false;
        }
        //判断是否存在
        $cart_id = $this->isExist($data, $user);
        if ($cart_id) {
            return $this->where('cart_id', '=', $cart_id)->inc('product_num', $data['product_num'])->update();
        } else {
            $data['describe'] = trim($data['describe'], ';');
            $data['app_id'] = self::$app_id;
            $data['shop_supplier_id'] = $user['shop_supplier_id'];
            $data['cashier_id'] = $user['cashier_id'];
            return $this->save($data);
        }
    }

    /**
     * 判断购物车商品是否存在
     */
    public function isExist($data, $user)
    {
        $model = $this;
        if (isset($data['table_id']) && $data['table_id']) {
            $model = $model->where('table_id', '=', $data['table_id']);
        }
        if ($data['eat_type'] == 10) {
            $model = $model->where(function ($query) use ($user) {
                $query->where('cashier_id', '=', 0)->whereOr('cashier_id', '=', $user['cashier_id']);
            });
        } else {
            $model = $model->where('cashier_id', '=', $user['cashier_id']);
        }
        $cart_id = $model->where('is_stay', '=', 0)
            ->where('product_id', '=', $data['product_id'])
            ->where('shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('product_sku_id', '=', $data['product_sku_id'])
            ->where('feed', '=', $data['feed'])
            ->where('attr', '=', $data['attr'])
            ->where('eat_type', '=', $data['eat_type'])
            ->value('cart_id');
        return $cart_id;
    }

    /**
     *清空桌号购物车
     */
    public function deleteTableAll($user, $table_id)
    {
        $model = $this;
        $model = $model->where(function ($query) use ($user) {
            $query->where('cashier_id', '=', 0)->whereOr('cashier_id', '=', $user['cashier_id']);
        });
        return $model->where('shop_supplier_id', '=', $user['shop_supplier_id'])
            ->where('eat_type', '=', 10)
            ->where('table_id', '=', $table_id)
            ->delete();
    }

    //判断商品是否下架
    public function productState($product_id)
    {
        return (new ProductModel)->where('product_id', '=', $product_id)
            ->where('product_status', '=', 10)
            ->where('is_delete', '=', 0)
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

    /**
     * 获取购物车 + 订单统计数据-送厨
     * @param array $param [shop_supplier_id]
     * @param int $table_id
     * @param int $order_id
     * @param string $product_source
     * @return array
     */
    public static function getHallCartOrderDetail($param, $table_id, $order_id = 0, $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        return (new self())->getOrderCartDetail($param, $table_id, $order_id, '', $product_source);
    }

    /**
     * 获取购物车 + 订单统计数据
     * @param array $param [shop_supplier_id]
     * @param int $table_id
     * @param int $order_id
     * @param string $product_source
     * @return array
     */
    public function getOrderCartDetail($param, $table_id = 0, $order_id = 0, $device_id = '', $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        $order = null;
        // 收银接单页【进入桌台】会从header取值
        $batch_no = request()->header('batchNo') !== 'undefined' ? request()->header('batchNo') : '';
        //
        if ($device_id && !$order_id && !$table_id) {
            $order_id = OrderModel::where('device_id', $device_id)
                ->where('is_stay', 0)
                ->where('is_settled', 0)
                ->where('table_id', 0)
                ->where('order_status', OrderStatusEnum::NORMAL)
                ->where('pay_status', OrderPayStatusEnum::PENDING)
                ->where('order_source', OrderSourceEnum::CASHIER)
                ->order('order_id', 'desc')
                ->value('order_id') ?: 0;
        }
        // 购物车商品列表（未送厨的订单商品）
        $cartList = [];
        //
        if ($order_id > 0 || $table_id > 0) {
            $order_field = [
                'order_id',
                'app_id',
                'shop_supplier_id',
                'is_buffet',
                'order_status',
                'pay_status',
                'user_id',
                'discount_ratio',
                'order_source',
                'order_no',
                'meal_num',
                'total_product_price',
                'service_money',
                'discount_money',
                'consumption_tax_money',
                'user_discount_money',
                'pay_price',
                'order_price',
                'refund_money',
                'table_no',
                'table_id',
                'original_price',
                'order_type',
                'delivery_status',
                'buffet_expired_time',
                'pay_time',
                'is_lock',
                'create_time',
                'discount_change_price',
                'delivery_type',
                'small_discount_type',
                'small_diff_money',
                'small_auto',
                'checkout_discount_type',
                'checkout_diff_money',
                'merge_id',
                'is_change_price',
                'merge_parent_id',
                'pay_fee_money',
                'total_price',
                'total_product_service_fee',
                'total_product_consumption_tax',
                'total_product_service_consumption_tax',
                'setting_service_money',
                'is_must_notice',
                'parent_id',
                'order_name',
            ];
            // 订单商品列表数据
            $order = OrderModel::detail([
                $order_id > 0 ? ['order_id', '=', $order_id] : ['table_id', '=', $table_id],
                ['order_status', '=', OrderStatusEnum::NORMAL],
                ['is_merge', '=', 0],
            ], [
                'product' => function ($q) use ($product_source, $batch_no) {
                    $q->with([
                        'product' => function ($q) {
                            $q->field('product_id, category_id, is_enable_grade, is_alone_grade, alone_grade_equity, alone_grade_type, is_points_gift, sales_initial, sales_actual');
                        },
                        'orderProductFree'
                    ])->where(function ($q) use ($product_source, $batch_no) {
                        $q->where('add_source', $product_source)
                            ->whereOr('is_send_kitchen', 1)
                            ->when(!empty($batch_no), function ($q) use ($batch_no) {
                                $q->whereOr('batch_no', $batch_no);
                            });
                    })->hidden(['product_name']);
                },
                'buffet',
                'buffetCustomerType',
                'delay',
                'supplier' => function ($q) {
                    $q->field('shop_supplier_id, settle_type, service_type, service_money');
                },
                'payType',
                'refundType',
                'parentOrder' => function ($q) {
                    $q->field(['order_id']);
                },
                'orderFree',
            ], $order_field);
            if (!$order) {
                $order = null;
            } else {
                // 购物车商品列表（未送厨的订单商品）
                if (!empty($batch_no)) {
                    $cartList = array_filter($order->product?->toArray(), function ($orderProduct) use ($product_source, $batch_no) {
                        return ($orderProduct['is_send_kitchen'] == 0 && $orderProduct['add_source'] == $product_source) || ($orderProduct['batch_no'] == $batch_no && $orderProduct['is_send_kitchen'] == 0);
                    });
                } else {
                    $cartList = array_filter($order->product?->toArray(), function ($orderProduct) use ($product_source) {
                        return $orderProduct['is_send_kitchen'] == 0 && $orderProduct['add_source'] == $product_source;
                    });
                }
            }
        }
        //
        $consumptionTaxSetting = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $param['shop_supplier_id']);
        // 购物车(未送厨商品)列表统计
        $cart_total_num = 0;                        // 购物商品数量
        $cart_total_product_price = 0;              // 购物车商品原价
        $cart_product_pay_price = 0;                // 购物车商品应付价钱
        $cart_user_discount_money = 0;              // 会员折扣
        $cart_consumption_tax = 0;                  // 购物车总消费税
        $cart_product_discount_money = 0;           // 购物车优惠折扣金额
        $cart_product_consumption_tax = 0;          // 购物车商品消费税
        $cart_product_service_fee = 0;              // 购物车商品服务费
        $cart_product_service_consumption_tax = 0;  // 购物车商品服务费消费税
        $cart_o_product_consumption_tax = 0;          // 购物车商品消费税(原价)
        $cart_o_product_service_fee = 0;              // 购物车商品服务费(原价)
        $cart_o_product_service_consumption_tax = 0;  // 购物车商品服务费消费税(原价)
        // 是否优惠折扣比例
        foreach ($cartList as $product) {
            $cart_total_num += $product['total_num'];                   // 购物商品数量
            $cart_total_product_price += $product['total_product_price'];     // 购物车商品原价
            $cart_product_pay_price += $product['total_price'];         // 购物车商品实付价钱
            $cart_user_discount_money += $product['grade_total_money']; // 会员折扣
            if ($consumptionTaxSetting['is_open']) {
                $cart_consumption_tax += $product['consumption_tax'];
            }
            $cart_product_consumption_tax += $product['product_consumption_tax'];
            $cart_product_service_fee += $product['product_service_fee'];
            $cart_product_service_consumption_tax += $product['product_service_consumption_tax'];
            $cart_o_product_consumption_tax += $product['product_original_consumption_tax'];          // 购物车商品消费税(原价)
            $cart_o_product_service_fee += $product['product_original_service_fee'];              // 购物车商品服务费(原价)
            $cart_o_product_service_consumption_tax += $product['product_original_service_consumption_tax'];  // 购物车商品服务费消费税(原价)
            // 购物车优惠折扣
            $cart_product_discount_money += $product['product_discount_money'];
        }

        // 购物车优惠折扣
        $cart_discount_money = 0;
        if ($order) {
            if ($cart_product_discount_money > 0) {
                $cart_discount_money = $cart_product_discount_money;
            }
        }
        $cart_pay_price = $cart_product_pay_price + $cart_product_service_fee;  // 购物车产生的pay_price

        // 订单信息
        $buffet_discount_money = 0;                 // 自助餐套餐的优惠折扣
        $total_send_product_price = 0;                    //已送厨总商品价格
        $total_send_consumption_tax = 0;                  //已送厨总消费税
        $total_send_product_discount_money = 0;           //已送厨优惠折扣金额
        $total_send_product_consumption_tax = 0;          //已送厨商品消费税
        $total_send_product_service_fee = 0;              //已送厨商品服务费
        $total_send_product_service_consumption_tax = 0;  //已送厨商品服务费消费税
        $total_o_send_product_price = 0;                    //已送厨总商品价格（原价）
        $total_o_send_product_consumption_tax = 0;          //已送厨商品消费税(原价)
        $total_o_send_product_service_fee = 0;              //已送厨商品服务费(原价)
        $total_o_send_product_service_consumption_tax = 0;  //已送厨商品服务费消费税(原价)

        $total_buffet_price = 0;                        //自助餐总商品价格
        $total_buffet_consumption_tax = 0;              //自助餐总消费税
        $total_buffet_discount_money = 0;               //自助餐优惠折扣金额
        $total_buffet_product_consumption_tax = 0;      //自助餐商品消费税
        $total_buffet_service_fee = 0;                  //自助餐商品服务费
        $total_buffet_service_consumption_tax = 0;      //自助餐商品服务费消费税
        $total_o_buffet_price = 0;                        //自助餐总商品价格(原价)
        $total_o_buffet_product_consumption_tax = 0;      //自助餐商品消费税(原价)
        $total_o_buffet_service_fee = 0;                  //自助餐商品服务费(原价)
        $total_o_buffet_service_consumption_tax = 0;      //自助餐商品服务费消费税(原价)
        if ($order) {
            $sendNum = 0;   // 已送厨的
            $order_send_consumption_tax_money = 0; // 已送厨房的消费税
            $order_discount_money = 0; // 订单商品优惠折扣
            $setting_service_money = $order['setting_service_money']; // 固定服务费
            if (isset($order['product'])) {
                $orderProductList = array_filter($order['product']->toArray(), function ($orderProduct) use ($product_source) {
                    return $orderProduct['is_send_kitchen'] == 1 || $orderProduct['add_source'] == $product_source;
                });
                foreach ($orderProductList as $item) {
                    if ($item['is_send_kitchen'] == 1 && $item['is_return'] == 0) {
                        // 商品数量
                        $sendNum += $item['total_num'];
                        // 已送厨的消费税
                        if ($consumptionTaxSetting['is_open']) {
                            $order_send_consumption_tax_money += $item['consumption_tax'];
                        }
                        // 订单优惠折扣
                        $order_discount_money += $item['product_discount_money'];
                        // 已送厨的商品价格
                        $total_send_product_price += $item['total_price'];
                        $total_send_consumption_tax += $item['consumption_tax'];
                        $total_send_product_discount_money += $item['product_consumption_tax'];
                        $total_send_product_consumption_tax += $item['product_consumption_tax'];
                        $total_send_product_service_fee += $item['product_service_fee'];
                        $total_send_product_service_consumption_tax += $item['product_service_consumption_tax'];

                        $total_o_send_product_price += $item['total_product_price'];
                        $total_o_send_product_consumption_tax += $item['product_original_consumption_tax'];          //已送厨商品消费税(原价)
                        $total_o_send_product_service_fee += $item['product_original_service_fee'];              //已送厨商品服务费(原价)
                        $total_o_send_product_service_consumption_tax += $item['product_original_service_consumption_tax'];  //已送厨商品服务费消费税(原价)

                    }
                }
            }
            // 自助餐的消费税
            $order_buffet_consumption_tax_money = 0; // 已送厨房的自助餐消费税
            $buffet_total_price = 0;  // 自助餐原价格
            if ($order['is_buffet']) {
                foreach ($order['buffetCustomerType'] as $item) {
                    $buffet_total_price += $item['total_price'];
                    if ($consumptionTaxSetting['is_open']) {
                        $buffet_consumption_tax = $item['consumption_tax'];
                        $order_buffet_consumption_tax_money += $buffet_consumption_tax;
                    }
                    $discount_money = helper::bcsub($item['total_price'], $item['total_pay_price'], 2);  // 折扣后差价
                    $buffet_discount_money = helper::bcadd($buffet_discount_money, $discount_money);    // 所有优惠金额累加

                    $total_buffet_price += $item['total_pay_price'];
                    $total_buffet_consumption_tax += $item['consumption_tax'];
                    $total_buffet_discount_money += $item['product_consumption_tax'];
                    $total_buffet_product_consumption_tax += $item['product_consumption_tax'];
                    $total_buffet_service_fee += $item['product_service_fee'];
                    $total_buffet_service_consumption_tax += $item['product_service_consumption_tax'];

                    $total_o_buffet_price += $item['total_price'];                        //自助餐总商品价格(原价)
                    $total_o_buffet_product_consumption_tax += $item['product_original_consumption_tax'];      //自助餐商品消费税(原价)
                    $total_o_buffet_service_fee += $item['product_original_service_fee'];                  //自助餐商品服务费(原价)
                    $total_o_buffet_service_consumption_tax += $item['product_original_service_consumption_tax'];      //自助餐商品服务费消费税(原价)
                }
            }
            // 加钟数量
            $delayNum = Order::getDelayNum($order);
            // 自助餐数量
            $buffetCustomerNum = Order::getBuffetCustomerNum($order);
            // 最终消费税 （送厨+未送厨）
            $order_consumption_tax_money = $cart_consumption_tax + $order_send_consumption_tax_money + $order_buffet_consumption_tax_money;  // 总消费税 = 购物车商品消费税 + 订单已送厨 商品消费税 + 自助餐商品消费税
            //
            $order_total_num = $sendNum + $delayNum + $buffetCustomerNum;
            $order_total_price = $order['total_product_price'];
            $order_total_product_pay_price = $order['total_price'];
            $order_service_money = $order['service_money'];
            $order_user_discount_money = $order['user_discount_money'];
            $order_original_price = $order['original_price'];
            $order_discount_change_price = $order['discount_change_price'];

            // 不包含消费税的pay_price
            if ($consumptionTaxSetting['calc_type'] == 1) { // 1-已含税 2-未含税
                // 含税不需要关联消费税到pay_price
                $order_pay_price =  $order['pay_price'];
            } else {
                $order_pay_price =  $order['pay_price'] - $order['consumption_tax_money'];
            }
            // 支付手续费
            $order_pay_fee_money = $order['pay_fee_money'];
        } else {
            $setting_service_money = 0;
            $order_total_num = 0;
            $order_total_price = 0;
            $order_total_product_pay_price = 0;
            $order_service_money = 0;
            $order_discount_money = 0;
            $order_consumption_tax_money = 0;
            $order_user_discount_money = 0;
            $order_original_price = 0;
            $order_discount_change_price = 0;
            $order_pay_price = 0;
            $order_pay_fee_money = 0;
        }
        // 订单 + 购物车 统计
        $total_num = $order_total_num + $cart_total_num;                                                    // 商品总数量
        $order_discount_money = helper::bcadd($order_discount_money, $cart_discount_money);                 // 优惠折扣
        $order_discount_money = helper::bcadd($order_discount_money, $buffet_discount_money);
        // 小计
        $total_price = helper::bcadd($order_total_price, $cart_total_product_price);                        // 小计【商品原金额】（不包含商品服务费、商品服务费消费税）

        $pay_subtotal = helper::bcadd($order_total_product_pay_price, $cart_product_pay_price);             // 折后小计【商品折后金额】
        $service_money = helper::bcadd($order_service_money, $cart_product_service_fee);                    // 服务费
        $special_discount = $order_discount_money;                                                          // 優惠折扣
        $total_consumption_tax_money = $order_consumption_tax_money;                                        // 消费税
        $total_user_discount_money = helper::bcadd($order_user_discount_money, $cart_user_discount_money);  // 會員折扣
        $no_tax_total_pay_price = helper::bcadd($order_pay_price, $cart_pay_price);                         // 合计应收(还不包含消费税)

        // 存在订单改价
        if ($order_discount_change_price > 0 || $order_discount_change_price == -1) {
            $order_discount_change_price = $order_discount_change_price == -1 ? 0 : $order_discount_change_price;
            // 改价后的优惠折扣
            $total_pay_price = helper::bcadd($order_discount_change_price, $order_pay_fee_money);  // 这里是为了还原pay_price, 必须加上手续费

            if ($consumptionTaxSetting['calc_type'] == 1) { // 1-已含税 2-未含税
                // 订单商品应付，不含税时但要加商品服务费的消费税
                $cart_total = $cart_product_pay_price + $cart_product_service_fee + $cart_product_service_consumption_tax;
                $send_total = $total_send_product_price + $total_send_product_service_fee + $total_send_product_service_consumption_tax;
                $buffet_total = $total_buffet_price + $total_buffet_service_fee + $total_buffet_service_consumption_tax;
            } else {
                $cart_total = $cart_product_pay_price + $cart_product_consumption_tax + $cart_product_service_fee + $cart_product_service_consumption_tax;
                $send_total = $total_send_product_price + $total_send_product_consumption_tax + $total_send_product_service_fee + $total_send_product_service_consumption_tax;
                $buffet_total = $total_buffet_price + $total_buffet_product_consumption_tax + $total_buffet_service_fee + $total_buffet_service_consumption_tax;
            }
            $final_total = $cart_total + $send_total + $buffet_total + $setting_service_money;
            $special_discount = helper::bcsub($final_total, $total_pay_price);  // 最终原价 - 最终应付价
            $special_discount = helper::bcadd($special_discount, $order_discount_money);  // 加上之前的优惠

        } else {
            // 1-已含税 2-未含税
            if ($consumptionTaxSetting['calc_type'] == 1) {
                // 含税不需要关联消费税到pay_price，但要加商品服务费的消费税
                $total_pay_price = helper::bcadd($no_tax_total_pay_price, $cart_product_service_consumption_tax);
            } else {
                $total_pay_price = helper::bcadd($no_tax_total_pay_price, $total_consumption_tax_money);
            }
        }
        // 抹零
        $small_discount_money = 0;
        if ($order) {
            // 还原溢价抹零
            $total_pay_price = helper::bcsub($total_pay_price, $order_pay_fee_money); // 还原为无手续费pay_pryce
            $total_pay_price = helper::bcadd($total_pay_price, $order['small_diff_money']);
            if ($order['small_discount_type'] == 1) { //抹分
                $after_total_pay_price = floor($total_pay_price * 10) / 10;
                $small_discount_money = floatval(helper::bcsub($total_pay_price, $after_total_pay_price));
                $total_pay_price = $after_total_pay_price;
            } elseif ($order['small_discount_type'] == 2) { //抹角
                $after_total_pay_price = (int)$total_pay_price;
                $small_discount_money = floatval(helper::bcsub($total_pay_price, $after_total_pay_price));
                $total_pay_price = $after_total_pay_price;
            } elseif ($order['small_discount_type'] == 3) { //四舍五入到角
                $after_total_pay_price = round($total_pay_price, 1);
                $small_discount_money = floatval(helper::bcsub($total_pay_price, $after_total_pay_price));
                $total_pay_price = $after_total_pay_price;
            } elseif ($order['small_discount_type'] == 4) { //四舍五入到元
                $after_total_pay_price = round($total_pay_price);
                $small_discount_money = floatval(helper::bcsub($total_pay_price, $after_total_pay_price));
                $total_pay_price = $after_total_pay_price;
            }
        }
        $special_discount += $small_discount_money;

        // 合并订单
        $pay_order_id = 0;
        $mergeTableList = [];
        $merge_pay_price = helper::bcadd($total_pay_price, $order_pay_fee_money);
        $merge_order_price = $order_original_price;
        $merge_final_price = 0;
        $merge_pay_type_fee = 0;
        $checkout_diff_money = 0; // 结账抹零金额
        if ($order) {
            if ($order['merge_id']) {
                /**
                 * 合单
                 */
                $list = $order->mergeTableList();
                foreach ($list as $o) {
                    $o_pay_price = $o->getBackPayPrice();
                    $o_order_price = $o->getBackOrderPrice();
                    $merge_pay_price = helper::bcadd($merge_pay_price, $o_pay_price);
                    $merge_order_price = helper::bcadd($merge_order_price, $o_order_price);
                    $mergeTableList[] = [
                        'order_id' => $o->order_id,
                        'table_no' => $o->table_no,
                        'pay_price' => $o_pay_price,
                    ];
                }
                if ($order['merge_parent_id']) {
                    $merge_pay_type_fee = OrderPayType::where('order_id', $order['merge_parent_id'])->sum('fee_money');
                    // 合计应收回显不计手续费
                    $merge_final_price = helper::bcadd($merge_pay_price, $merge_pay_type_fee);
                }
                $pay_order_id = $order['merge_parent_id'];
            } else {
                /**
                 * 独单
                 */
                // 合计应收回显不计手续费
                $merge_pay_type_fee = $order_pay_fee_money;
                $merge_final_price = $merge_pay_price;
                $merge_pay_price = helper::bcsub($merge_pay_price, $merge_pay_type_fee);
                $pay_order_id = $order['order_id'];
            }

            // 最终应收应该减去结账抹零金额 v1.1.0
            if (($order['checkout_discount_type'] ?? 0) > 0) {
                $checkout_diff_money = $order['checkout_diff_money'];
                $merge_final_price = helper::bcsub($merge_final_price, $checkout_diff_money);
            }

            //
            Cache::set('order_pay_price_' . $order['order_id'] . '_' . $order['app_id'], floatval($merge_pay_price), 60 * 60 * 24 * 1);

            // 如果子单 table_id 从父级取
            if ($order['parent_id'] > 0) {
                $table_id = Order::where('order_id', $order['parent_id'])->value('table_id');
            }
        }

        // 缓存起来给其他地方使用
        if (($table_id = ($order['table_id'] ?? 0) ?: $table_id) && ($app_id = ($order['app_id'] ?? 0))) {
            $tableCacheKey = $table_id . '_table_price' . $app_id;
            if (($order['parent_id'] ?? 0) > 0) {
                $tableSumOrderCacheKey = 'table_price_sub_sum_' . $order['parent_id'] . $app_id;
                Cache::hset($tableSumOrderCacheKey, $order['order_id'], floatval($merge_pay_price));
                Cache::set($tableCacheKey, array_sum(array_values(Cache::hgetall($tableSumOrderCacheKey))) );
                Cache::set($tableCacheKey . '_sub_sum', $tableSumOrderCacheKey );
            } else if (($order['table_id'] ?? 0) > 0) {
                if ($tableSumOrderCacheKey = Cache::get($tableCacheKey . '_sub_sum')) {
                    Cache::delete($tableSumOrderCacheKey);
                    Cache::delete($tableCacheKey . '_sub_sum');
                }
                Cache::set(($order['table_id'] ?? 0) . '_table_price' . $app_id, floatval($merge_pay_price));
            }
        }

        // 子单列表
        $subOrderList = [];
        $orderCompleted = false;
        $orderSendKitchen = false;
        $tmpOrder = OrderModel::detail($order_id);
        if ($tmpOrder) {
            $parentId = $tmpOrder['parent_id'] == 0 ? $tmpOrder['order_id'] : $tmpOrder['parent_id'];
            $orderCompleted = OrderModel::checkOrderComplete($parentId);
            $orderSendKitchen = OrderModel::checkOrderSendKitchen($parentId);
            if (!$orderCompleted) {
                $subOrderList = OrderModel::getSubOrderList($parentId);
            }
        }

        $buffetNames = [];
        if ($order) {
            $mainOrderId = $order->parent_id > 0 ? $order->parent_id : $order->order_id;
            $mainOrder = OrderModel::detail($mainOrderId, ['buffet']);
            if ($mainOrder && $mainOrder['buffet']) {
                foreach ($mainOrder['buffet'] as $buffet) {
                    if (!in_array($buffet['name_text'], $buffetNames)) {
                        $buffetNames[] = $buffet['name_text'];
                    }
                }
            }
        }

        //
        return [
            'orderInfo' => $order ?? [],
            'perOrderInfo' => [
                'pay_order_id' => $pay_order_id,
                'pay_price' => floatval($merge_pay_price),    // 【合计应收：】
                'final_price' => $merge_final_price,
                'order_price' => floatval($merge_order_price),
                'pay_type_fee' => floatval($merge_pay_type_fee),
                'checkout_diff_money' => floatval($checkout_diff_money), // 结账抹零金额
                'checkout_discount_type' => $order['checkout_discount_type'] ?? 0, // 结账抹零类型
                'buffet_names' => $buffetNames,
            ],
            'mergeTableList' => $mergeTableList,
            'userInfo' => $order ? UserModel::cardDetail($order['user_id']) : [],
            'sumInfo' => [
                'total_num' => $total_num,                                                    // 商品总数
                'subtotal' => floatval($total_price),                                         // 小计【商品金额】
                'pay_subtotal' => floatval($pay_subtotal),                                    // 商品折后小计
                'service_money' => floatval($service_money),                                  // 服务费
                'special_discount' => floatval($special_discount),                            // 優惠折扣
                'total_consumption_tax_money' => floatval($total_consumption_tax_money),      // 消费税
                'total_user_discount_money' => floatval($total_user_discount_money),          // 會員折扣
                'total_pay_price' => floatval($merge_pay_price),                              // 【合计应收：】
                'total_original_price' => floatval($order_original_price),                    // 订单原价应收
            ],
            'schemeProductList' => $order ? $order->getSchemeProductList() : (new Order)->getSchemeProductList(),
            'subOrderList' => $subOrderList,
            'orderCompleted' => $orderCompleted,
            'orderSendKitchen' => $orderSendKitchen,
        ];
    }

    /**
     * 预计算 购物车 + 订单 价格
     * @param $mobile
     * @param $table_id
     * @param $order_id
     * @param $product_source
     * @return array|false
     */
    public function preOrderCartPrice($mobile = '', $user_id = 0, $table_id = 0, $order_id = 0, $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        if (empty($mobile) && empty($user_id)) {
            $this->error = "用户不存在";
            return false;
        }
        // 用户信息
        $user = (new UserModel)
            ->field("user_id, IF(password != '', 1, 0) AS have_password, nickName, mobile, grade_id, card_id, (balance + gift_balance) as balance, points, is_delete")
            ->with(['grade', 'card', 'cardRecord'])
            ->when($user_id, function ($q) use ($user_id) {
                $q->where('user_id', $user_id);
            })
            ->when($mobile && empty($user_id), function ($q) use ($mobile) {
                $q->where('mobile', $mobile);
            })
            ->where(['is_delete' => 0])
            ->find();
        if (!$user) {
            $this->error = "用户不存在";
            return false;
        }

        // 桌台
        if ($table_id > 0) {
            // 计算订单会员优惠后价格
            $query = (new Order())->with([
                'product' => function ($q) use ($product_source) {
                    $q->where(function ($q) use ($product_source) {
                        $q->where('is_send_kitchen', 1);
                        $q->whereOr('add_source', $product_source);
                    });
                },
            ])
            ->where('order_status', '=', 10)
            ->where('is_delete', '=', 0)
            ->order('order_id', 'desc');
            if ($order_id > 0) {
                $query->where('order_id', '=', $order_id);
            } else {
                $query->where('table_id', '=', $table_id);
            }
            $order = $query->find();
        } else if ($order_id > 0) {
            $order = (new Order())->with([
                'product' => function ($q) use ($product_source) {
                    $q->where(function ($q) use ($product_source) {
                        $q->where('is_send_kitchen', 1);
                        $q->whereOr('add_source', $product_source);
                    });
                },
            ])
            ->where('order_id', '=', $order_id)
            ->where('order_status', '=', 10)
            ->where('is_delete', '=', 0)
            ->order('order_id', 'desc')
            ->find();
        } else {
            $order = null;
        }

        if ($order) {
            // 合单
            if ($order->merge_id) {
                //
                $order_total_pay_price = 0;
                //
                $mergeOrderList = Order::where('merge_id', $order->merge_id)->select();
                foreach ($mergeOrderList as $subOrder) {
                    $re = self::prePriceByOrder($subOrder, $user);
                    $order_total_pay_price = helper::bcadd($order_total_pay_price, $re['order_total_pay_price']);
                }
                $order_arr = [
                    'order_total_pay_price' => $order_total_pay_price,
                ];
            } else {
                $order_arr = self::prePriceByOrder($order, $user);
            }
        } else {
            $order_arr = [
                'order_total_pay_price' => 0,
            ];
        }
        //
        $total_pay_price = $order_arr['order_total_pay_price'];    // 应收
        $total_pay_price = round($total_pay_price, 2);

        return compact('user', 'total_pay_price');
    }

    // 预算订单会员优惠后价格
    public static function prePriceByOrder($order, $user)
    {
        $pay_money = 0;
        $order_price = 0;
        $user_discount_money = 0;
        $consumption_tax = 0;   // 总消费税（商品消费税+服务费消费税）
        $total_product_consumption_tax = 0;   // 总商品消费税
        $total_product_service_consumption_tax = 0;   // 总商品服务费消费税
        $total_service_fee = 0; // 总服务费
        // 配置
        $setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        $consumptionTaxSetting = $setting[SettingEnum::TAX_RATE]['values'];     // 消费税设置
        $serviceFee = $setting[SettingEnum::SERVICE_CHARGE]['values'];          // 服务费设置
        $service_charge_rate = $serviceFee['service_charge_rate'] ?: 0; // 商品服务费率
        // 是否优惠折扣比例
        // 10 - 百分比打折：如果订单原价为100，打8折，即表示消费者需要支付商品价格的80%，计算方式为：100 × 80% = 80
        // 20 - 直接减免：如果订单原价为100，20% OFF，表示从原价中减去20%的价格。计算方式为：100 - (100 × 20%) = 80
        $discount_ratio = 1;
        $discount_method = $order['discount_method'] ?? 10; // 折扣计算方式 10-按百分比 20-直接减免
        if ($order['discount_ratio'] > 0) {
            if ($discount_method == 20) {
                $discount_ratio = (100 - $order['discount_ratio']) / 100;
            } else {
                $discount_ratio = $order['discount_ratio'] / 100;
            }
        }

        if ($order && isset($order['product'])) {
            //
            foreach ($order['product'] as $product) {
                if ($product['is_return'] == 1) {
                    continue;
                }
                // 会员折扣的总额差
                $grade_total_money = 0;
                $unit_pay_price = $product['product_price'];
                if ($product['product']['is_enable_grade'] && $product['total_price'] > 0) {
                    if ($user) {
                        $discount = (new CardRecordModel)->getDiscount($user['user_id']);
                    } else {
                        $discount = 0;
                    }
                    $alone_grade_type = 10;
                    // 商品单独设置了会员折扣
                    if ($user) {
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
                            $discountRatio = helper::bcdiv($user['grade']['equity'], 100);
                        }
                    } else {
                        $discountRatio = 1;
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
                            $grade_product_price = $product['product']['alone_grade_equity'][$user['grade_id']];
                            $discount && $grade_product_price = round($grade_product_price * $discount, 2);
                        } else {
                            // 商品会员折扣后单价
                            $grade_product_price = helper::bcmul($product['product_price'], $discountRatio, 3);
                        }
                        $grade_product_price = round($grade_product_price, 2);  // 有折扣后就四舍五入确定一次两位数价格
                        $gradeTotalPrice = $grade_product_price * $product['total_num'];
                        // 原商品总价 - 折扣后
                        $grade_total_money = helper::number2(helper::bcsub($product['product_price'] * $product['total_num'], $gradeTotalPrice, 3));
                        // 商品应付单价
                        $unit_pay_price = $grade_product_price;
                    }
                }

                // 主表order数据累加
                $unit_pay_price = round($unit_pay_price * $discount_ratio, 2);  // 商品折扣后单价(四舍五入两位小数)
                $order_product_total_price = $unit_pay_price * $product['total_num'];
                $pay_money += $order_product_total_price;  // 实付金额
                $order_price += $product['product_price'] * $product['total_num'];  // 商品原价
                $user_discount_money += $grade_total_money; // 商品优惠金额
                $product_rate = $product['tax_rate'];
                // 消费税
                if ($consumptionTaxSetting['is_open']) {
                    $unit_consumption_tax = Product::getConsumptionTax($product_rate, $unit_pay_price, $product['tax_calc_type']);
                    $total_unit_consumption_tax = helper::bcmul($unit_consumption_tax, $product['total_num']);
                    //
                    $total_product_consumption_tax += floatval($total_unit_consumption_tax);  // 商品消费税累计
                    $consumption_tax += floatval($total_unit_consumption_tax);  // 总消费税累计
                }
                // 服务费
                if ($serviceFee['is_open'] && $serviceFee['charge_type'] == 2) {
                    // 折扣后的商品服务费
                    $unit_product_service_price = ProductModel::getProductServiceFee($unit_pay_price, $service_charge_rate, $consumptionTaxSetting['calc_type'], $product_rate);  // 商品单价服务费
                    $product_service_fee = helper::bcmul($unit_product_service_price, $product['total_num']); // 应付商品服务费
                    //
                    $total_service_fee += floatval($product_service_fee);   // 服务费累计
                    // VAT 开启才计算消费税
                    if ($consumptionTaxSetting['is_open'] && $serviceFee['is_open_tax']) {
                        // 折扣后商品服务费消费税
                        $unit_product_service_consumption_tax = ProductModel::getProductServiceConsumptionTax($unit_product_service_price, $product_rate);  // 商品单价服务费的消费税
                        $product_service_consumption_tax = helper::bcmul($unit_product_service_consumption_tax, $product['total_num']);  // 商品total服务费的消费税
                        //
                        $total_product_service_consumption_tax += floatval($product_service_consumption_tax);  // 商品服务费消费税累计
                        $consumption_tax += floatval($product_service_consumption_tax); // 消费税累计
                    }
                }
            }
        }
        // 自助餐
        $buffetPrice = 0;
        $delayPrice = 0;
        $order_buffet_consumption_tax_money = 0;                    // 自助餐的消费税
        $order_buffet_product_service_money = 0;                    // 自助餐的商品服务费
        $order_buffet_product_service_consumption_tax_money = 0;    // 自助餐的商品服务费的消费税
        if ($order && $order['is_buffet'] == 1) {
            // 自助餐顾客费用
            $orderId = $order['order_id'] ?? 0;
            $orderField = $order['parent_id'] > 0 ? 'sub_order_id' : 'order_id';
            $buffetPrice = Order::getBuffetCustomerPrice($orderId, $orderField);
            $buffetPrice = $buffetPrice * $discount_ratio; //  如果有优惠折扣，再次计算 v1.1.0
            // 自助餐消费税
            $order_buffet_consumption_tax_money = Order::getBuffetCustomerTotalConsumptionTax($orderId, $orderField);
            // 自助餐商品服务费的消费税
            $order_buffet_product_service_consumption_tax_money = Order::getBuffetCustomerTotalProductServiceConsumptionTax($orderId, $orderField);
            // 自助餐的服务费
            $order_buffet_product_service_money = Order::getBuffetCustomerTotalProductServiceFee($orderId, $orderField);
            // 加钟费用
            $delayPrice = (new OrderDelay())->where($orderField, '=', $orderId)->sum('total_price');
        }
        //
        $total_price = helper::bcadd($pay_money, helper::bcadd($buffetPrice, $delayPrice));
        // 服务费
        $setting_service_money = ($serviceFee['is_open'] && $serviceFee['charge_type'] == 1) ? $serviceFee['service_charge'] : 0;
        $total_service_fee =  $total_service_fee + $order_buffet_product_service_money + $setting_service_money;
        $total_price = helper::bcadd($total_price, $total_service_fee);
        //
        $consumption_tax = helper::bcadd($consumption_tax, $order_buffet_consumption_tax_money);

        $total_price = round($total_price, 2); // 订单商品总价（不是商品原价总价、是商品折扣后(如果有)的总价）
        //
        if ($consumptionTaxSetting['calc_type'] == 1) {
            // 含税不需要关联消费税到pay_price
            $order_total_pay_price = $total_price + $total_product_service_consumption_tax + $order_buffet_product_service_consumption_tax_money; // 应付金额 = 商品应付 + 消费税(商品服务费消费税)
        } else {
            $order_total_pay_price = $total_price + floatval($consumption_tax); // 应付金额 = 商品折扣总价（会员折扣） + 固定服务费用 + 商品服务费 + 自助餐 + 加钟费 + 消费税(商品消费税和商品服务费消费税)
        }

        return [
            'order_total_product_price' => $order_price,                             // 订单商品原价
            'order_total_pay_price' => $order_total_pay_price,                       // 使用会员后应付
            'order_total_user_discount_money' => $order_price - $total_price,        // 订单会员折扣优惠金额
            'order_total_consumption_tax' => $consumption_tax,                       // 消费税
        ];
    }
}
