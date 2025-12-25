<?php

namespace app\common\model;

use help\DateHelp;
use help\QueueHelp;
use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\store\PayType;
use think\model\concern\SoftDelete;
use app\common\enum\http\StatusCode;
use app\common\model\order\OrderFree;
use app\common\model\product\Product;
use app\common\model\buffet\BuffetTax;
use app\common\model\order\OrderBuffet;
use app\common\model\order\OrderPayType;
use app\common\model\order\OrderProduct;
use app\common\model\product\ProductSku;
use app\common\model\product\ProductTax;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\settings\SettingEnum;
use app\common\model\product\ProductFeed;
use app\common\enum\order\OrderSourceEnum;
use app\common\enum\order\OrderStatusEnum;
use app\common\model\buffet\BuffetProduct;
use app\common\service\order\OrderService;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\product\ProductSoldOut;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\enum\settings\DeliveryTypeEnum;
use app\common\enum\product\DeductStockTypeEnum;
use app\common\model\order\Order;
use app\common\model\product\ProductSkuMaterial;
use app\common\model\product\Product as ProductModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\product\ProductSku as ProductSkuModel;

/**
 * 订单基础模型
 */
class BaseModelOrder extends BaseModel
{
    use SoftDelete;
    protected $name = 'sale_bill';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    const CASHIER_PRODUCT_SOURCE = 1;   // 收银
    const TABLET_PRODUCT_SOURCE = 2;    // 平板
    const SCAN_PRODUCT_SOURCE = 3;      // 扫码

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'order_id',
        'state_text',
        'order_source_text',
        'order_type_text',
        'deliver_text',
        'elapsed_time',
        'pay_time_text',
        'buffet_remaining_time',
        'actual_price',
        'actual_receive_price',
        'lock_elapsed_time',
    ];

    /**
     * 兼容字段
     */
    public function getOrderIdAttr()
    {
        return $this->uuid ?? 0;
    }

    /**
     * 订单商品列表
     */
    public function product()
    {
        if ($this->parent_id == 0) {
            return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'order_id', 'order_id')->hidden(['content']);
        } else {
            return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'sub_order_id', 'order_id')->hidden(['content']);
        }
    }

    /**
     * 订单自助餐列表
     */
    public function buffet()
    {
        if ($this->parent_id == 0) {
            return $this->hasMany('app\\common\\model\\order\\OrderBuffet', 'order_id', 'order_id');
        } else {
            return $this->hasMany('app\\common\\model\\order\\OrderBuffet', 'sub_order_id', 'order_id');
        }
    }

    /**
     * 订单自助餐优惠列表
     */
    public function buffetDiscount()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderBuffetDiscount', 'order_id', 'order_id');
    }

    /**
     * 订单自助餐顾客类型价格列表
     */
    public function buffetCustomer()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderBuffetCustomer', 'order_id', 'order_id');
    }

    /**
     * 订单子单关联(拆单)
     */
    public function subOrder()
    {
        return $this->hasMany('app\\common\\model\\order\\Order', 'parent_id', 'order_id');
    }

    /**
     * 订单自助餐顾客类型价格列表
     */
    public function buffetCustomerType()
    {
        if ($this->parent_id == 0) {
            return $this->hasMany('app\\common\\model\\order\\OrderBuffetCustomer', 'order_id', 'order_id');
        } else {
            return $this->hasMany('app\\common\\model\\order\\OrderBuffetCustomer', 'sub_order_id', 'order_id');
        }
    }

    /**
     * 订单自助餐加钟列表
     */
    public function delay()
    {
        if ($this->parent_id == 0) {
            return $this->hasMany('app\\common\\model\\order\\OrderDelay', 'order_id', 'order_id');
        } else {
            return $this->hasMany('app\\common\\model\\order\\OrderDelay', 'sub_order_id', 'order_id');
        }
    }

    /**
     * 收银员
     */
    public function cashier()
    {
        return $this->hasOne('app\\common\\model\\shop\\User', 'shop_user_id', 'cashier_id')->hidden(['update_time', 'password', 'user_name']);
    }

    /**
     * 关联订单收货地址表
     */
    public function address()
    {
        return $this->hasOne('app\\common\\model\\order\\OrderAddress');
    }

    /**
     * 关联自提订单联系方式
     */
    public function extract()
    {
        return $this->hasOne('app\\common\\model\\order\\OrderExtract');
    }

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\user\\User', 'user_id', 'user_id');
    }

    /**
     * 关联用户基础信息
     */
    public function userBaseInfo()
    {
        return $this->belongsTo('app\\common\\model\\user\\User', 'user_id', 'user_id')->field(['user_id', 'mobile', 'nickname as nickName', 'real_name']);
    }

    /**
     * 关联供应商表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id');
    }

    /**
     * 关联桌台
     */
    public function tables()
    {
        return $this->hasOne('app\\common\\model\\store\\Table', 'table_id', 'table_id');
    }

    /**
     * 关联配送信息
     */
    public function deliver()
    {
        return $this->belongsTo('app\\common\\model\\order\\OrderDeliver', 'order_id', 'order_id')->order('deliver_id desc');
    }

    /**
     * 已送厨
     */
    public function sendKitchenProduct()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'order_id', 'order_id')->where('is_send_kitchen', 1)->hidden(['content']);
    }

    /**
     * 已送厨
     */
    public function sendAndBatchKitchenProduct()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'order_id', 'order_id')->where(function ($q) {
            $q->where('is_send_kitchen', 1)->whereOr('batch_time', '>', 0);
        })->hidden(['content']);
    }

    /**
     * 未送厨
     */
    public function unSendKitchenProduct()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'order_id', 'order_id')->where('is_send_kitchen', 0)->hidden(['content']);
    }

    /**
     * 付款扣库存的产品
     */
    public function payDeductStockProduct()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderProduct', 'order_id', 'order_id')->where('deduct_stock_type', 20)->hidden(['content']);
    }

    /**
     * 发票信息
     */
    public function invoiceInfo()
    {
        return $this->hasOne('app\\common\\model\\order\\OrderInvoiceInfo', 'order_id', 'order_id')->order('id', 'desc');
    }

    /**
     * 支付方式
     */
    public function payType()
    {
        $prefix = env('DB_PREFIX');
        return $this->hasMany('app\\common\\model\\order\\OrderPayType', 'order_id', 'order_id')
            ->field("{$prefix}order_pay_type.order_id, {$prefix}order_pay_type.value, {$prefix}order_pay_type.price, {$prefix}order_pay_type.disabled_cancel, {$prefix}order_pay_type.pay_status, {$prefix}order_pay_type.fee_money, {$prefix}order_pay_type.payment_order_id")
            ->field("pt.source, pt.name, pt.remark")
            ->leftJoin('pay_type pt', "pt.value = {$prefix}order_pay_type.value")
            ->orderRaw("pt.source, pt.value");
    }

    /**
     * 退款方式
     */
    public function refundDestinations()
    {
        $prefix = env('DB_PREFIX');
        return $this->hasMany('app\\common\\model\\order\\OrderRefundDestination', 'order_id', 'order_id')
            ->field("{$prefix}order_refund_destination.order_id, {$prefix}order_refund_destination.value, {$prefix}order_refund_destination.price, sum({$prefix}order_refund_destination.refund_money) as refund_money")
            ->field("pt.source, pt.name, pt.remark")
            ->leftJoin('pay_type pt', "pt.value = {$prefix}order_refund_destination.value")
            ->group("{$prefix}order_refund_destination.value")
            ->orderRaw("pt.source, pt.value");
    }

    /**
     * 合单子单的父单
     */
    public function subParentOrder()
    {
        return $this->belongsTo('app\\common\\model\\order\\Order', 'parent_id', 'order_id');
    }


    /**
     * 合单子单的父单
     */
    public function parentOrder()
    {
        return $this->belongsTo('app\\common\\model\\order\\Order', 'merge_parent_id', 'order_id');
    }

    /**
     * 退款方式
     */
    public function refundType()
    {
        return $this->hasMany('app\\common\\model\\order\\OrderRefund', 'order_id', 'order_id')->field(['order_id', 'refund_type', 'refund_money']);
    }

    /**
     * 合单子单列表
     */
    public function mergeList()
    {
        return $this->hasMany('app\\common\\model\\order\\Order', 'merge_parent_id', 'order_id');
    }

    /**
     * 付款扣库存的产品
     */
    public function erpInventoryRecords()
    {
        return $this->hasMany('app\\common\\model\\erp\\ErpInventoryRecord', 'order_id', 'order_id');
    }

    /**
     * 用户余额日志
     */
    public function userBalanceLog()
    {
        return $this->hasOne('app\\common\\model\\user\\BalanceLog', 'order_id', 'order_id');
    }

    /**
     * 已锁单时间
     */
    public function getLockElapsedTimeAttr($value, $data)
    {
        if (isset($data['lock_time']) && $data['lock_time'] > 0) {
            return time() - $data['lock_time'];
        }
        return 0;
    }

    /**
     * 订单实付金额
     */
    public function getActualPriceAttr($value, $data)
    {
        $actual_price = $data['actual_price'] ?? 0;
        $refund_money = $data['refund_money'] ?? 0;
        $calculated_price = helper::bcsub($actual_price, $refund_money);
        return helper::number2($calculated_price >= 0 ? $calculated_price : 0);
    }

    /**
     * 订单实际收到金额(pay_price - refund_money)
     */
    public function getActualReceivePriceAttr($value, $data)
    {
        if (($data['is_free'] ?? 0) > 0) {
            return 0;
        }

        $pay_price = isset($data['pay_price']) ? $data['pay_price'] : 0;
        $refund_money = isset($data['refund_money']) ? $data['refund_money'] : 0;

        return helper::bcsub($pay_price, $refund_money);
    }

    /**
     * 订单现金实际收到金额 (不包含找零)
     */
    public function getCashReceivePriceAttr($value, $data)
    {
        $prefix = env('DB_PREFIX');
        if ($cashAmount = $this->payType()->where("{$prefix}order_pay_type.value", OrderPayTypeEnum::CASH)->sum('price')) {
            $cashAmount = helper::bcsub($cashAmount, $data['change_due']);
        }
        return $cashAmount;
    }

    /**
     * 订单生成时间长度
     */
    public function getElapsedTimeAttr($value, $data)
    {
        if (isset($data['create_time'])) {
            // 获取当前时间
            $currentTime = time();
            // 获取订单生成时间
            $generateTime = $data['create_time'];
            // 返回时间长度
            return $currentTime - $generateTime;
        }
        return 0;
    }

    /**
     * 订单自助餐剩余时间
     */
    public function getBuffetRemainingTimeAttr($value, $data)
    {
        if (isset($data['buffet_expired_time'])) {
            $remaining_time = $data['buffet_expired_time'] - time();
            return max($remaining_time, 0);
        }
        return $value;
    }

    /**
     * 免赠标签
     */
    public function orderFree()
    {
        return $this->hasMany(OrderFree::class, 'order_id', 'order_id');
    }

    /**
     * 免单原因
     */
    public function getFreeTagTextAttr($value, $data)
    {
        $free_tag_arr = [];
        $orderFree = json_decode($this->orderFree, true);
        if ($orderFree) {
            foreach ($orderFree as $item) {
                $free_tag_arr[] = extractLanguage($item['free_tag'] ?? '');
            }
        }
        $free_tag_text = implode('、', $free_tag_arr);
        if ($free_tag_text) {
            if (isset($data['free_remark']) && $data['free_remark']) {
                $free_tag_text = $free_tag_text . '、' . $data['free_remark'];
            }
        } else if (isset($data['free_remark']) && $data['free_remark']) {
            $free_tag_text = $data['free_remark'];
        }
        return $free_tag_text;
    }

    /**
     * 支付时间格式化
     */
    public function getPayTimeTextAttr($value, $data)
    {
        return isset($data['pay_time']) && $data['pay_time'] != 0 ? DateHelp::formatTimeHis($data['pay_time']) : '-';
    }

    /**
     * 订单状态文字描述
     * @param $value
     * @param $data
     * @return string
     */
    public function getStateTextAttr($value, $data)
    {

        // 订单状态
        if (isset($data['order_status']) && in_array($data['order_status'], [20, 30])) {
            return OrderStatusEnum::data($data['order_status'])['name'];
        }
        // 付款状态
        if (isset($data['pay_status']) && $data['pay_status'] == 10) {
            return OrderPayStatusEnum::data($data['pay_status'])['name'];
        }
        // 发货状态
        if (isset($data['order_status']) && isset($data['delivery_type']) && $data['order_status'] == 10) {
            return OrderStatusEnum::data($data['order_status'])['name'];
        }

        return $value;
    }

    /**
     * 订单状态文字描述
     * @param $value
     * @param $data
     * @return string
     */
    public function getDeliverTextAttr($value, $data)
    {
        // 订单状态待接单＝1,待取货＝2,配送中＝3,已完成＝4,已取消＝5, 指派单=8
        if (isset($data['order_status']) && in_array($data['order_status'], [20, 30])) {
            return OrderStatusEnum::data($data['order_status'])['name'];
        }
        // 发货状态
        if (isset($data['delivery_status']) && $data['delivery_status'] == 10) {
            return __('待配送');
        }
        // 发货状态
        if (isset($data['delivery_status']) && $data['delivery_status'] == 20) {
            $deliverStatus = [0 => __('待接单'), 1 => __('待接单'), 2 => __('待取货'), 3 => __('配送中'), 4 => __('已完成')];
            return $deliverStatus[$data['deliver_status'] ?? 0] ?? __('待配送');
        }
        return $value;
    }

    /**
     * 订单类型
     * @param $value
     * @return string
     */
    public function getOrderTypeTextAttr($value, $data)
    {
        if (isset($data['order_type'])) {
            return $data['order_type'] == 0 ? __('外卖订单') : __('店内订单');
        }
        return $value;
    }

    /**
     * 订单来源
     * @param $value
     * @return string
     */
    public function getOrderSourceTextAttr($value, $data)
    {
        if (isset($data['order_source'])) {
            return OrderSourceEnum::data($data['order_source'])['name'];
        }
        return $value;
    }

    /**
     * 付款状态
     * @param $value
     * @return array
     */
    public function getPayStatusAttr($value)
    {
        return [
            'text' => OrderPayStatusEnum::data($value)['name'],
            'value' => $value
        ];
    }

    /**
     * 改价金额（差价）
     * @param $value
     * @return array
     */
    public function getUpdatePriceAttr($value)
    {
        return [
            'symbol' => $value < 0 ? '-' : '+',
            'value' => Helper::number2(abs($value))
        ];
    }

    /**
     * 发货状态
     * @param $value
     * @return array
     */
    public function getDeliveryStatusAttr($value)
    {
        $status = [10 => __('待配送'), 20 => __('已配送')];
        return ['text' => $status[$value], 'value' => $value];
    }

    /**
     * 收货状态
     * @param $value
     * @return array
     */
    public function getReceiptStatusAttr($value)
    {
        $status = [10 => __('待收货'), 20 => __('已收货')];
        return ['text' => $status[$value], 'value' => $value];
    }

    /**
     * 收货状态
     * @param $value
     * @return array
     */
    public function getOrderStatusAttr($value)
    {
        return [
            'text' => OrderStatusEnum::data($value)['name'],
            'value' => $value
        ];
    }

    /**
     * 配送方式
     * @param $value
     * @return array
     */
    public function getDeliveryTypeAttr($value)
    {
        if (!$value) {
            return [];
        }
        return [
            'text' => DeliveryTypeEnum::data($value)['name'],
            'value' => $value
        ];
    }

    /**
     * 支付方式名称
     * @param $value
     * @return string
     */
    public function getPayTypeName($value)
    {
        $text = '-';
        if ($value == OrderPayTypeEnum::FREE_PAY) {
            $text = OrderPayTypeEnum::data($value)['name'] ?? '-';
        } else {
            $text = PayType::where('value', $value)->withTrashed()->value('name') ?: '-';
        }
        return $text;
    }

    /**
     * 支付方式名称
     * @param $value
     * @return string
     */
    public function getPayTypeRemark($value)
    {
        $text = '-';
        if ($value == OrderPayTypeEnum::FREE_PAY) {
            $text = OrderPayTypeEnum::data($value)['name'] ?? '-';
        } else {
            $payType = PayType::where('value', $value)->withTrashed()->find();
            if ($payType) {
                $text = $payType->name;
                $text = $payType->remark ?: $text;
            } else {
                $text = OrderPayTypeEnum::data($value)['name'] ?? '-';
            }
        }
        return $text;
    }

    /**
     * 获取订单指定来源的相关产品列表
     * @param string $productSource
     * @param string $type          // 触发送厨类型 kitchen-送厨 payment-结算 accept-接单
     * @param string $batch_no      // 批次号，用于查出扫码端下单的商品
     * @return array
     */
    public function getOrderSourceProductList($productSource = self::CASHIER_PRODUCT_SOURCE, $type = 'kitchen', $batch_no = '')
    {
        // 订单产品列表
        $orderProductList = [];
        $curOrderProductList = $this->product;

        if (!empty($batch_no) && $type == 'accept') {
            foreach ($curOrderProductList as $orderProduc) {
                if ($orderProduc->is_send_kitchen == 1 || $orderProduc->batch_no == $batch_no) {
                    $orderProductList[] = $orderProduc;
                }
            }
        } else if (!empty($batch_no)) {
            foreach ($curOrderProductList as $orderProduc) {
                if ($orderProduc->is_send_kitchen == 1 || $orderProduc->add_source == $productSource || $productSource === false || $orderProduc->batch_no == $batch_no) {
                    $orderProductList[] = $orderProduc;
                }
            }
        } else {
            foreach ($curOrderProductList as $orderProduc) {
                if ($orderProduc->is_send_kitchen == 1 || $orderProduc->add_source == $productSource || $productSource === false) {
                    $orderProductList[] = $orderProduc;
                }
            }
        }
        // 原始产品列表
        $allProductIds = array_values(array_unique(array_column($orderProductList, 'product_id')));
        $allProductList = count($allProductIds) > 0 ? ProductModel::with(['feed.material'])
            ->where('product_id', 'in', $allProductIds)
            ->select()
            ->toArray() : [];
        $allProductList = array_column($allProductList, null, 'product_id');
        // 原始产品规格列表
        $allProductSkuIds = array_values(array_unique(array_column($orderProductList, 'product_sku_id')));
        $allProductSkuList = count($allProductIds) > 0 ? ProductSkuModel::with(['material'])
            ->field('product_sku_id, spec_name, product_price, stock_num')
            ->where('product_sku_id', 'in', $allProductSkuIds)
            ->select()
            ->toArray() : [];
        $allProductSkuList = array_column($allProductSkuList, null, 'product_sku_id');
        // 结果
        return compact('orderProductList', 'allProductList', 'allProductSkuList', 'productSource');
    }

    /**
     * 获取订单的相关产品的税率列表
     * @param array $productIds
     * @param int $productTaxType 产品税率类型  1-食堂税类，2-外带税类
     * @return array
     */
    public function getOrderProductTaxRateList($productIds = [], $productTaxType = 1)
    {
        return $productIds ? ProductTax::alias('pt')
            ->join('tax_category tc', 'pt.tax_category_id = tc.id')
            ->where('pt.product_id', 'in', $productIds)
            ->where('pt.product_tax_type', '=', $productTaxType)
            ->column('tax_rate', 'product_id') : [];
    }

    /**
     * 获取订单的相关自助餐的税率列表
     * @param array $productIds
     * @param int $productTaxType
     * @return array
     */
    public function getBuffetTaxRateList()
    {
        $order = $this;
        if ($order['is_buffet'] == 1) {
            $buffet_ids = (new OrderBuffet)->where('order_id', $this->parent_id ?: $order['order_id'])->column('buffet_id');
            return $buffet_ids ? BuffetTax::alias('bt')->join('tax_category tc', 'bt.tax_category_id = tc.id')
                ->where('bt.buffet_id', 'in', $buffet_ids)
                ->where('bt.buffet_tax_type', '=', 1)
                ->column('tax_rate', 'buffet_id') : [];
        } else {
            return [];
        }
    }

    /**
     * 得到距离
     * @param $value
     * @return int
     */
    public static function getDistance($ulon, $ulat, $slon, $slat)
    {
        // 地球半径
        $R = 6378137;
        // 将角度转为狐度
        $radLat1 = deg2rad($ulat);
        $radLat2 = deg2rad($slat);
        $radLng1 = deg2rad($ulon);
        $radLng2 = deg2rad($slon);
        // 结果
        $s = acos(cos($radLat1) * cos($radLat2) * cos($radLng1 - $radLng2) + sin($radLat1) * sin($radLat2)) * $R;
        // 精度
        $s = round($s * 10000) / 10000;
        return round($s);
    }

    /**
     * 生成订单号
     * @return string
     */
    public function orderNo()
    {
        return OrderService::createOrderNo();
    }

    /**
     * 生成新版订单号
     * @return string
     */
    public function newOrderNo($order_source)
    {
        return OrderService::createNewOrderNo($order_source);
    }

    /**
     * 生成交易号
     * @return string
     */
    public function tradeNo()
    {
        return OrderService::createTradeNo();
    }

    /**
     * 支付方式 - 可退款金额
     */
    public function payTypeCellRefundMoneys()
    {
        $res = [];
        $refundDestination = array_column($this->refundDestinations->toArray(), null, 'value');
        foreach ($this->payType as $pay) {
            $payTypeCellRefundMoney = $pay['price'];
            if ($pay['value'] == OrderPayTypeEnum::CASH) {
                $pay['price'] = floatval(helper::bcsub($payTypeCellRefundMoney, $this->change_due));
                $payTypeCellRefundMoney = floatval(helper::bcsub($payTypeCellRefundMoney, $this->change_due));
            }
            if (isset($refundDestination[$pay['value']])) {
                $payTypeCellRefundMoney = floatval(helper::bcsub($payTypeCellRefundMoney, $refundDestination[$pay['value']]['refund_money']));
            }
            $pay['cell_refund_money'] = $payTypeCellRefundMoney;
            $res[] = $pay->toArray();
        }
        return $res;
    }

    /**
     * 支付方式 - 使用会员余额
     */
    public function payTypeBalance()
    {
        $balance = 0;
        foreach ($this->payType as $pay) {
            if ($pay['value'] == OrderPayTypeEnum::BALANCE) {
                $balance = $pay['price'];
            }
        }
        return $balance;
    }

    /**
     * 当前订单使用余额之后 剩余会员余额
     */
    public function surplusBalance()
    {
        if ($this->user) {
            return $this->user?->total_balance - $this->payTypeBalance();
        }
        return 0;
    }

    /**
     * 获取 设备id或者桌台id 对应的订单id
     */
    public function getDeviceIdOrTableIdToOrderId($device_id = '', $table_id = 0)
    {
        if ($table_id) {
            return $this->getTableIdToOrderId($table_id);
        }
        //
        return self::where('device_id', $device_id)
            ->where('is_stay', 0)
            ->where('is_settled', 0)
            ->where('order_status', OrderStatusEnum::NORMAL)
            ->where('pay_status', OrderPayStatusEnum::PENDING)
            ->where('order_source', OrderSourceEnum::CASHIER)
            ->where('is_merge', 0)
            ->order('order_id', 'desc')
            ->value('order_id') ?: 0;
    }

    /**
     * 获取设桌台id对应的订单id
     */
    public function getTableIdToOrderId($table_id)
    {
        return self::where('table_id', $table_id)
            ->where('order_status', OrderStatusEnum::NORMAL)
            ->value('order_id') ?: 0;
    }

    /**
     * 判断订单是否可操作
     */
    public function validateOrderActionableStatus($type = '')
    {
        if ($type != 'payment' && $this->is_lock == 1) {
            return '订单已被锁定，请解锁后重新操作';
        }
        if ($this->getData('order_status') == OrderStatusEnum::CANCELLED || $this->getData('order_status') == OrderStatusEnum::APPLY_CANCEL) {
            return '订单已取消';
        }
        if ($this->getData('order_status') == OrderStatusEnum::COMPLETED) {
            return '订单已结账';
        }
        return '';
    }

    /**
     * 判断订单是否拆单
     */
    public function isSplitTheOrder($order_id = 0)
    {
        if (self::where('parent_id', $order_id ?: $this->order_id ?: 0)->value('order_id')) {
            return '当前订单已拆单，请前去收银机操作';
        }
        return '';
    }

    /**
     * 检查部分支付
     */
    public function checkPartialPayment()
    {
        if ($this->merge_parent_id && (
            OrderPayType::where('order_id', $this->merge_parent_id)->find() ||
            Order::where('order_id', $this->merge_parent_id)->where('pay_status', OrderPayStatusEnum::SUCCESS)->find()
        )) {
            return '当前订单已被部分支付，不支持取消';
        } else if (
            OrderPayType::where('order_id', $this->order_id)->find() ||
            Order::where('order_id', $this->order_id)->where('pay_status', OrderPayStatusEnum::SUCCESS)->find()
        ) {
            return '当前订单已被部分支付，不支持取消';
        } else if (
            OrderPayType::where('order_id', 'in', self::where('parent_id', $this->order_id)->column('order_id'))->find() ||
            Order::where('order_id', 'in', self::where('parent_id', $this->order_id)->column('order_id'))->where('pay_status', OrderPayStatusEnum::SUCCESS)->find()
        ) {
            return '当前订单已被部分支付，不支持取消';
        }
        return '';
    }

    /**
     * 添加后材料库存是否充足
     * @param int $productId
     * @param int $productSkuId
     * @param array $feed_ids
     * @param int $productNum
     * @param int $orderId
     * @return array
     */
    public function addAndCheckFeedStockIsFull($product, $feed_ids, $productNum, $orderId = 0, $product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        // 添加的需消耗的加料库存
        $addNum = $productNum;  // 添加数
        $addFeedIds = $feed_ids; // 加料ids
        $addConsumed = [];
        if ($feed_ids) {
            foreach ($product['feed'] as $feed_v) {
                if (in_array($feed_v['product_feed_id'], $addFeedIds) && (empty($feed_v['material']) || count($feed_v['material']) == 0)) {
                    $addConsumed[$feed_v['product_feed_id']] = [
                        'stockConsumed' => $addNum,
                        'feed_name_text' => $feed_v['feed_name_text']
                    ];
                }
            }
        }
        // 无消耗直接添加
        if (empty($addConsumed)) {
            return [];
        }

        // 订单当前消耗的
        $orderConsumed = [];
        if ($orderId > 0) {
            $orderProductList = OrderProduct::field('product_id, product_sku_id, feed_ids, total_num')
                ->where('order_id', $orderId)
                ->where(function ($q) use ($product_source) {
                    $q->where('is_send_kitchen', 1);
                    $q->whereOr('add_source', $product_source);
                })
                ->select()
                ->toArray();
            //
            $allProductIds = array_values(array_unique(array_column($orderProductList, 'product_id')));
            $allProductList = ProductModel::with(['sku.material', 'feed'])
                ->where('product_id', 'in', $allProductIds)
                ->where('product_status', '=', 10)
                ->select()->toArray();
            $allProductList = array_column($allProductList, null, 'product_id');
            //
            foreach ($orderProductList as $orderProduct) {
                // 加料
                $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
                $productFeedIds = $productFeedIds ?: [];
                if (!empty($productFeedIds)) {
                    $productD = $allProductList[$orderProduct['product_id']];
                    foreach ($productD['feed'] as $feed_v) {
                        if (in_array($feed_v['product_feed_id'], $productFeedIds) && empty($feed_v['material'])) {
                            $consumedNum = isset($orderConsumed[$feed_v['product_feed_id']]['stockConsumed']) ? $orderConsumed[$feed_v['product_feed_id']]['stockConsumed'] : 0;
                            $consumedNum += $orderProduct['total_num'];
                            $orderConsumed[$feed_v['product_feed_id']] = [
                                'stockConsumed' => $consumedNum,
                                'feed_name_text' => $feed_v['feed_name_text']
                            ];
                        }
                    }
                }
            }
        }

        // 合并消耗的库存
        $totalFeedStockConsumed = OrderProduct::mergeConsumedArr($addConsumed, $orderConsumed, 'stockConsumed');
        $outProductFeedIds = $this->checkFeedStockIsFull($totalFeedStockConsumed, $addFeedIds);
        if (count($outProductFeedIds) > 0) {
            $feed_name = '';
            if (count($addConsumed) > 0) {
                $addConsumed = array_reverse($addConsumed, true);
                foreach ($addConsumed as $product_feed_id => $item) {
                    if (in_array($product_feed_id, $outProductFeedIds)) {
                        $feed_name = isset($item['feed_name_text']) ? $item['feed_name_text'] : '';
                        break;
                    }
                }
            }
            $this->errorData = $product['product_name_text'] . $feed_name;
        }
        return $outProductFeedIds;
    }

    /**
     * 添加后材料库存是否充足
     * @param int $productId
     * @param int $productSkuId
     * @param array $feed_ids
     * @param int $productNum
     * @param int $orderId
     * @return array
     */
    public function addAndCheckMaterialStockIsFull($product, $productSkuId, $feed_ids, $productNum, $orderId = 0, $product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        // 添加的需消耗的材料库存
        $addNum = $productNum;  // 添加数
        $addFeedIds = $feed_ids; // 加料ids
        $addConsumed = [];
        if ($feed_ids) {
            // 添加商品消耗的材料
            foreach ($product['feed'] as $feed_v) {
                // 库存联动材料数
                if (in_array($feed_v['product_feed_id'], $addFeedIds) && !(empty($feed_v['material']) || count($feed_v['material']) == 0)) {
                    foreach ($feed_v['material'] as $material) {
                        $consumedNum = isset($addConsumed[$material['material_id']]['consumed']) ? $addConsumed[$material['material_id']]['consumed'] : 0;
                        $consumedNum += $material['material_num'] * $addNum;
                        $addConsumed[$material['material_id']] = [
                            'consumed' => $consumedNum,
                            'feed_name_text' => $feed_v['feed_name_text']
                        ];
                    }
                }
            }
        }
        // 规格消耗的材料
        $addProductSku = ProductSku::with(['material'])
            ->where('product_id', '=', $product['product_id'])
            ->where('product_sku_id', '=', $productSkuId)
            ->find()->toArray();
        foreach ($addProductSku['material'] as $material) {
            $consumedNum = isset($addConsumed[$material['material_id']]['consumed']) ? $addConsumed[$material['material_id']]['consumed'] : 0;
            $consumedNum += $material['material_num'] * $addNum;
            $addConsumed[$material['material_id']] = ['consumed' => $consumedNum];
        }
        // 无消耗直接添加
        if (empty($addConsumed)) {
            return [];
        }

        // 订单当前已送厨和未送厨消耗的
        $orderConsumed = [];
        if ($orderId > 0) {
            $orderUnSendList = OrderProduct::field('product_id, product_sku_id, feed_ids, total_num')
                ->where('order_id', $orderId)
                ->where(function ($q) use ($product_source) {
                    $q->where('is_send_kitchen', 1);
                    $q->whereOr('add_source', $product_source);
                })
                ->select()
                ->toArray();
            //
            $allProductIds = array_values(array_unique(array_column($orderUnSendList, 'product_id')));
            $allProductList = ProductModel::with(['sku.material', 'feed'])
                ->where('product_id', 'in', $allProductIds)
                ->where('product_status', '=', 10)
                ->select()->toArray();
            $allProductList = array_column($allProductList, null, 'product_id');
            //
            $allProductSkuIds = array_values(array_unique(array_column($orderUnSendList, 'product_sku_id')));
            $allProductSkuList = ProductSku::with(['material'])
                ->field("*, CONCAT(product_id, '_', product_sku_id) as _key_")
                ->where('product_id', 'in', $allProductIds)
                ->where('product_sku_id', 'in', $allProductSkuIds)
                ->select()->toArray();
            $allProductSkuList = array_column($allProductSkuList, null, '_key_');
            //
            foreach ($orderUnSendList as $orderProduct) {
                // 加料消耗的材料
                $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
                $productFeedIds = $productFeedIds ?: [];
                if (!empty($productFeedIds)) {
                    $productD = $allProductList[$orderProduct['product_id']];
                    foreach ($productD['feed'] as $feed_v) {
                        // 库存联动材料数
                        if (in_array($feed_v['product_feed_id'], $productFeedIds) && !empty($feed_v['material'])) {
                            foreach ($feed_v['material'] as $material) {
                                $consumedNum = isset($orderConsumed[$material['material_id']]['consumed']) ? $orderConsumed[$material['material_id']]['consumed'] : 0;
                                $consumedNum += $material['material_num'] * $orderProduct['total_num'];
                                $orderConsumed[$material['material_id']] = ['consumed' => $consumedNum];
                            }
                        }
                    }
                }
                // 规格消耗的材料
                $orderProductSku = $allProductSkuList[$orderProduct['product_id'] . '_' . $orderProduct['product_sku_id']];
                foreach ($orderProductSku['material'] as $material) {
                    $consumedNum = isset($orderConsumed[$material['material_id']]['consumed']) ? $orderConsumed[$material['material_id']]['consumed'] : 0;
                    $consumedNum += $material['material_num'] * $orderProduct['total_num'];
                    $orderConsumed[$material['material_id']] = ['consumed' => $consumedNum];
                }
            }
        }

        // 合并材料消耗
        $totalMaterialConsumed = OrderProduct::mergeConsumedArr($addConsumed, $orderConsumed);
        $outMaterialProductIds = $this->checkMaterialStockIsFull($totalMaterialConsumed);
        if (count($outMaterialProductIds) > 0) {
            $feed_name = '';
            if (count($addConsumed) > 0) {
                $addConsumed = array_reverse($addConsumed, true);
                foreach ($addConsumed as $material_id => $item) {
                    if (in_array($material_id, $outMaterialProductIds)) {
                        $feed_name = isset($item['feed_name_text']) ? $item['feed_name_text'] : '';
                        break;
                    }
                }
            }
            $this->errorData = $product->product_name_text . $addProductSku['spec_name_text'] . $feed_name;
        }
        return $outMaterialProductIds;
    }

    /**
     * 添加后材料库存是否充足
     * @param array $consumed_list
     * @return array
     */
    public function checkMaterialStockIsFull(array $consumed_list)
    {
        return Product::where('product_id', 'in', array_keys($consumed_list))->where('type', Product::TYPE_MATERIAL)->where(function ($q) use ($consumed_list) {
            foreach ($consumed_list as $product_id => $consumed) {
                $q->whereOr(function ($qq) use ($product_id, $consumed) {
                    $qq->where('product_id', $product_id);
                    $qq->where('product_material_stock', '<', $consumed['consumed']);
                });
            }
        })->column('product_id');
    }

    /**
     * 检查加料库存是否充足
     * @param array $consumed_list
     * @return array
     */
    public function checkFeedStockIsFull(array $consumed_list, array $addFeedIds = [])
    {
        return ProductFeed::where('product_feed_id', 'in', array_keys($consumed_list))->where(function ($q) use ($consumed_list, $addFeedIds) {
            foreach ($consumed_list as $product_feed_id => $consumed) {
                if (!$addFeedIds || in_array($product_feed_id, $addFeedIds)) {
                    $q->whereOr(function ($qq) use ($product_feed_id, $consumed) {
                        $qq->where('product_feed_id', $product_feed_id);
                        $qq->where('stock_num', '<', $consumed['stockConsumed']);
                    });
                }
            }
        })->column('product_feed_id');
    }

    /**
     * 并发性判断订单是否可操作
     * @return BaseModelOrder|QueueHelp
     */
    public function concurrencyValidateOrderActionableStatus()
    {
        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queue->while();

        /** @var BaseModelOrder $order */
        $order = self::where('order_id', '=', $this->order_id)->find();
        if (!$order) {
            $this->error = '订单不存在';
            $queue->release();
            return false;
        }

        // 验证订单状态是否可操作
        if ($error = $order->validateOrderActionableStatus()) {
            $this->error = $error;
            $queue->release();
            return false;
        }

        //
        return [$order, $queue];
    }

    /**
     * 处理添加菜品数据
     * @param null $data
     */
    public function processAddToOrderParams($data, $settingData = null)
    {
        $params = [
            'order_id' => ($data['order_id'] ?? 0) ?: 0,
            'table_id' => ($data['table_id'] ?? 0) ?: 0,
            'add_source' => $data['add_source'] ?? 1,
            'product_id' => intval($data['product_id'] ?? 0),
            'attr_ids' => ($data['attr_ids'] ?? []) ?: [],
            'product_sku_id' => intval($data['product_sku_id'] ?? 0),
            'product_num' => intval($data['product_num'] ?? 0),
            'describe' => $data['describe'] ?? '',
            'delivery' => $data['delivery'] ?? 40,
            'is_send_kitchen' => intval($data['is_send_kitchen'] ?? 0),
            'send_kitchen_time' => intval($data['send_kitchen_time'] ?? 0),
            'is_move' => intval($data['is_move'] ?? 0),
            'move_from_table_id' => intval($data['move_from_table_id'] ?? 0),
            'move_from_order_id' => intval($data['move_from_order_id'] ?? 0),
            'remark' => $data['remark'] ?? '',
            'is_change_price' => intval($data['is_change_price'] ?? 0),
            'product_price' => $data['product_price'] ?? 0,
            'free_tag_order_product_id' => $data['free_tag_order_product_id'] ?? 0,
            'feed_uuids' => (array)($data['feed_uuids'] ?? []),
            'feed_ids' => isset($data['feed_uuids']) ? $data['feed_uuids']  : (isset($data['feed_ids']) ? $data['feed_ids'] : []),
            'scheme_id' => $data['scheme_id'] ?? 0,
            'sub_order_id' => $data['sub_order_id'] ?? 0
        ];
        //
        $settingData = $settingData ?: SettingModel::getAll($user['app_id'] ?? 0, $user['shop_supplier_id'] ?? 0);
        // 处理免费商品逻辑
        $params['is_free'] = intval($data['is_free'] ?? 0);
        if ($params['is_free']) {
            $params['is_free'] = ($settingData[SettingEnum::BUSINESS]['values']['gift_method'] ?? 10) == '10' ? 1 : 2;
        }
        $params['free_remark'] = $data['free_remark'] ?? '';
        // 厨显设置
        $params['kitchen_is_open'] = $settingData[SettingEnum::KITCHEN]['values']['is_open'] ?? 1; // 厨显是否开启 1-开启 0-关闭
        //
        return $params;
    }

    /**
     * 验证商品属性
     * @param $productDetail 商品
     * @param array $attrIds 属性ID数组
     * @param array $feedIds 加料ID数组
     * @return array ['status' => bool, 'message' => string]
     */
    public function validateProductAttributes($productDetail, $attrIds, $feedIds)
    {
        // 验证属性
        if (($productAttr = $productDetail->product_attr)) {
            $attrIds = array_filter($attrIds);
            $addAttrIds = array_keys($attrIds);
            if (is_array($productAttr) && !empty($productAttr)) {
                foreach ($productAttr as $item) {
                    // 获取最小和最大可选数量
                    $minSelection = $item['attribute_min_select'] ?? 0;
                    $maxSelection = $item['attribute_max_select'] ?? 0;
                    
                    // 已选择的数量
                    $selectedCount = isset($item['parent_id']) && isset($attrIds[$item['parent_id']]) ? count($attrIds[$item['parent_id']]) : 0;
                    
                    // 必选验证（最小可选数量 > 0 时必须选择）
                    if ($minSelection > 0 && $selectedCount == 0) {
                        $space = checkDetect() == 'zhtw' || checkDetect() == 'zh' ? '' : ' ';
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => __('请选择') . $space . extractLanguage($item['attribute_name'])
                        ];
                    }
                    
                    // 最小可选验证
                    if ($selectedCount > 0 && $minSelection > 0 && $selectedCount < $minSelection) {
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => extractLanguage($item['attribute_name']) . ' ' . __('至少选择') . $minSelection . __('个')
                        ];
                    }
                    
                    // 最多可选验证
                    if ($selectedCount > 0 && $maxSelection > 0 && $selectedCount > $maxSelection) {
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => extractLanguage($item['attribute_name']) . ' ' . __('最多选择') . $maxSelection . __('个')
                        ];
                    }
                }
            }
        }

        // 验证加料
        $productFeedId = $productDetail->feed()->column('product_feed_id');
        $feedIds = array_filter($feedIds);
        $missingFeedIds = array_diff($feedIds, $productFeedId);
        if ($missingFeedIds) {
            return [
                'code' => StatusCode::PRODUCT_ERROR_NOT_EXIST_FEED,
                'status' => false,
                'data' => $missingFeedIds,
                'message' => __('加料已下架')
            ];
        }
        if ($productDetail->feed_min_select > 0 || $productDetail->feed_required == 1) {
            if (($productFeed = $productDetail->product_feed)) {
                if (is_array($productFeed) && !empty($productFeed)) {
                    $feedCount = count($feedIds);
                    $minSelection = $productDetail->feed_min_select ?? 0;
                    $maxSelection = $productDetail->feed_max_select ?? 0;
                    
                    // 必选验证（最小可选数量 > 0 时必须选择）
                    if ($minSelection > 0 && $feedCount == 0) {
                        $space = checkDetect() == 'zhtw' || checkDetect() == 'zh' ? '' : ' ';
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => __('请选择加料')
                        ];
                    }
                    
                    // 最小可选验证
                    if ($feedCount > 0 && $minSelection > 0 && $feedCount < $minSelection) {
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => __('加料至少选择') . $minSelection . __('个')
                        ];
                    }
                    
                    // 最多可选验证
                    if ($feedCount > 0 && $maxSelection > 0 && $feedCount > $maxSelection) {
                        return [
                            'code' => 0,
                            'status' => false,
                            'data' => [],
                            'message' => __('加料最多选择') . $maxSelection . __('个')
                        ];
                    }
                }
            }
        }
        //
        return ['status' => true, 'data' => [], 'message' => ''];
    }

    /**
     * 获取订单自助餐商品列表
     * @param int $order_id
     * @return array
     */
    public static function getOrderBuffetProductArr($order_id)
    {
        $list = (new OrderBuffet)->with(['buffetProduct'])->where('order_id', '=', $order_id)->select();
        $arr = [];
        foreach ($list as $buffet) {
            foreach ($buffet['buffetProduct'] as $product) {
                // 改为自助餐限制数量改为累加
                if (isset($arr[$product['product_id']]) && $arr[$product['product_id']]['limit_num'] != 0) {
                    if ($product['limit_num'] == 0) {
                        $arr[$product['product_id']] = [
                            'product_id' => $product['product_id'],
                            'limit_num' => $product['limit_num'],
                        ];
                    } else {
                        $arr[$product['product_id']] = [
                            'product_id' => $product['product_id'],
                            'limit_num' => $product['limit_num'] + $arr[$product['product_id']]['limit_num'],
                        ];
                    }
                } else if (!isset($arr[$product['product_id']])) {
                    $arr[$product['product_id']] = [
                        'product_id' => $product['product_id'],
                        'limit_num' => $product['limit_num'],
                    ];
                }
            }
        }
        return $arr;
    }

    /**
     * 获取订单自助餐商品限购数
     * @param int $order_id
     * @param int $product_id
     * @return int
     */
    public static function getBuffetProductLimitNum($order_id, $product_id)
    {
        $buffet_ids = (new OrderBuffet)->where('order_id', '=', $order_id)->column('buffet_id');
        $limit_num = (new BuffetProduct)->where('buffet_id', 'in', $buffet_ids)->where('product_id', '=', $product_id)->where('limit_num', '=', 0)->find();
        if ($limit_num) {
            return 0;
        } else {
            return (new BuffetProduct)->where('buffet_id', 'in', $buffet_ids)->where('product_id', '=', $product_id)->sum('limit_num');
        }
    }

    /**
     * 验证产品购买限额和库存
     * @param ProductModel $productDetail Product detail information
     * @param int $productSkuId Product SKU ID
     * @return array ['status' => bool, 'message' => string]
     */
    public function validatePurchaseLimit($orderId, $productSource, $productDetail, $productSkuId, $productNum = 1, $mealNum = 1, $feedIds = [])
    {
        $productId = $productDetail['product_id'];
        // 授权信息
        $license = request()->licenses;

        // 产品已售罄
        if (ProductSoldOut::where('product_id', $productId)->where('product_sku_id', $productSkuId)->find()) {
            return $this->handleError(ProductSkuModel::getNameById($productSkuId) . '库存不足');
        }

        /**
         * 判断限购
         */
        $isBuffet = array_key_exists($productId, self::getOrderBuffetProductArr($orderId)) ? 1 : 0;
        if ($isBuffet == 1) {
            $limitNum = self::getBuffetProductLimitNum($orderId, $productId) * $mealNum;
        } else {
            $limitNum = $productDetail->limit_num;
        }
        if ($limitNum && $productNum > $limitNum) {
            return $this->handleError('超过限购数量');
        }

        // 一次性获取所有订单商品数量信息
        $orderProducts = OrderProduct::where([
            'order_id' => $orderId,
            'product_id' => $productId,
        ])
            ->field("
                SUM(CASE
                    WHEN (merge_from_table_id = 0 AND (is_send_kitchen = 1 OR add_source = $productSource))
                    THEN total_num ELSE 0 END) as limit_num_count,
                SUM(CASE
                    WHEN (product_sku_id = $productSkuId AND add_source = $productSource)
                    THEN total_num ELSE 0 END) as stock_num_count
            ")
            ->find();
        //
        $curNum = $productNum + ($orderProducts->limit_num_count ?: 0);

        // 检查限购数量
        if ($limitNum && ($curNum > $limitNum)) {
            return $this->handleError('超过限购数量');
        }

        // 计算当前订单总数量（用于库存检查）
        $curOrderNum = ($orderProducts->stock_num_count ?: 0) + $productNum;

        // 添加到购物车时，库存应只判断未送厨数量 小于等于 库存即可
        $productSku = (new ProductSku)->where('product_sku_id', $productSkuId)->field('stock_num')->find();
        if (!$productSku) {
            $this->errorData = ['product_sku_id' => $productSkuId];
            return $this->handleError('规格已下架,请选择其他规格', StatusCode::PRODUCT_ERROR_NOT_EXIST_SKU);
        }
        //
        $is_material_sku = (new ProductSkuMaterial())->where('product_sku_id', $productSkuId)->find();
        if (($license['sale'] ?? 0) != 1 || !$is_material_sku) {
            $skuStockNum = $productSku->stock_num;
            if ($skuStockNum == 0) {
                return $this->handleError(ProductSkuModel::getNameById($productSkuId) . __('库存不足'), OrderErrorEnum::STOCK_ERROR);
            }
            if ($curOrderNum > $skuStockNum) {
                return $this->handleError(ProductSkuModel::getNameById($productSkuId) . __('库存不足'), OrderErrorEnum::STOCK_ERROR);
            }
            if ($productDetail->deduct_stock_type == DeductStockTypeEnum::CREATE) {
                $stockStatus = $this->productStockState($productId, $productSkuId, $orderId, $productSource);
                if (!$stockStatus) {
                    return $this->handleError(ProductSkuModel::getNameById($productSkuId) . __('库存不足'), OrderErrorEnum::STOCK_ERROR);
                }
            }
        }

        // 判断加料库存
        if ($feedIds) {
            $outProductFeedIds = $this->addAndCheckFeedStockIsFull($productDetail, $feedIds, $productNum, $orderId, $productSource);
            if (count($outProductFeedIds) > 0) {
                return $this->handleError('加料库存不足', OrderErrorEnum::STOCK_ERROR);
            }
        }

        // 材料库存是否充足
        if (($license['sale'] ?? 0) == 1) {
            $outProductArr = $this->addAndCheckMaterialStockIsFull($productDetail, $productSkuId, $feedIds, $productNum, $orderId, $productSource);
            if (count($outProductArr) > 0) {
                return $this->handleError($this->errorData . __('库存不足'), OrderErrorEnum::STOCK_ERROR);
            }
        }

        //
        return compact('isBuffet');
    }
}
