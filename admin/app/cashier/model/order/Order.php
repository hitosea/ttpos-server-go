<?php

namespace app\cashier\model\order;

use help\QueueHelp;
use think\facade\Db;
use app\common\library\helper;
use app\common\model_old\shop\Account;
use app\common\model_old\buffet\Buffet;
use app\common\enum\http\StatusCode;
use app\shop\model\product\Category;
use app\common\model_old\order\OrderFree;
use app\common\model_old\product\Product;
use app\common\model_old\store\TakeOrder;
use app\common\exception\BaseException;
use app\common\model_old\order\OrderBuffet;
use app\common\model_old\supplier\Printing;
use app\common\model_old\supplier\Supplier;
use app\common\enum\order\OrderTypeEnum;
use app\common\model_old\order\OrderPayType;
use app\common\model_old\order\OrderProduct;
use app\common\model_old\product\ProductSku;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\settings\SettingEnum;
use app\common\model_old\order\OrderPeakTime;
use app\common\enum\order\OrderSourceEnum;
use app\common\enum\order\OrderStatusEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model_old\buffet\BuffetCustomer;
use app\common\model_old\order\OrderAbnormalLog;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model_old\order\OrderOperationLog;
use app\cashier\model\order\Cart as CartModel;
use app\common\model_old\order\OrderProductReturn;
use app\common\model_old\order\Order as OrderModel;
use app\common\model_old\order\OrderBuffetCustomer;
use app\cashier\model\store\Table as TableModel;
use app\common\enum\product\DeductStockTypeEnum;
use app\common\service\order\OrderPrinterService;
use app\common\service\order\OrderCompleteService;
use app\common\model_old\settings\Setting as SettingModel;
use app\common\service\product\factory\ProductFactory;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model_old\user\BalanceLog as BalanceLogModel;
use app\common\repositories\OrderBusinessDataRepository;
use app\common\model_old\order\OrderBuffet as OrderBuffetModel;
use app\common\model_old\order\OrderProduct as OrderProductModel;
use app\cashier\service\order\paysuccess\type\MasterPaySuccessService;

/**
 * 普通订单模型
 */
class Order extends OrderModel
{
    /**
     * 隐藏字段
     * @var array
     */
    protected $hidden = [
        'update_time',
    ];


    /**
     * 取消订单
     */
    public function cancel($table_id)
    {
        $detail = $this->where('table_id', '=', $table_id)
            ->where('order_status', '=', 10)
            ->where('eat_type', '=', 10)
            ->find();

        if (!$detail) {
            TableModel::close($table_id);
            $this->error = "订单不存在";
            return false;
        }

        if ($detail['pay_status']['value'] == 20) {
            TableModel::close($table_id);
            $this->error = "订单已付款，不允许取消";
            return false;
        }
        if ($detail['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        return $detail->save(['order_status' => 20]);
    }

    /**
     * 待支付订单详情
     */
    public static function getPayDetail($orderNo)
    {
        $model = new static();
        return $model->where(['order_no' => $orderNo, 'pay_status' => 10, 'is_delete' => 0])->with(['product', 'user', 'supplier'])->find();
    }

    /**
     * 设置错误信息
     */
    protected function setError($error)
    {
        empty($this->error) && $this->error = $error;
    }

    /**
     * 是否存在错误
     */
    public function hasError()
    {
        return !empty($this->error);
    }

    //查询桌号信息
    public static function getTableInfo($table_id)
    {
        return (new static())->where('table_id', '=', $table_id)
            ->where('is_delete', '=', 0)
            ->order('order_id desc')
            ->find();
    }

    //查询桌号订单信息
    public static function getOrderInfo($table_id)
    {
        return (new static())->with('product')
            ->where('table_id', '=', $table_id)
            ->where('order_status', '=', 10)
            ->where('is_delete', '=', 0)
            ->order('order_id', 'desc')
            ->find();
    }

    /**
     * 折扣抹零
     */
    public function changeMoney($user, $data)
    {
        if (isset($data['order_id']) && $data['order_id'] > 0) {
            $detail = OrderModel::detail([
                ['order_id', '=', $data['order_id']],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
        } else if (isset($data['table_id']) && $data['table_id'] > 0) {
            $detail = self::getTableUnderwayOrder($data['table_id']);
        } else {
            $detail = null;
        }

        // 检查订单状态
        if (!$detail) {
            $this->error = '当前状态不可操作';
            return false;
        }

        // 检查支付状态
        if ($error = $detail->validateOrderActionableStatus()) {
            $this->error = $error;
            return false;
        }
        $this->startTrans();
        try {
            $discount_money = 0;
            $discount_ratio = 0;
            switch ($data['type']) {
                case '1': //改价
                    if ($data['money'] > 999999999 || $data['money'] < 0) {
                        $this->error = "价格范围错误";
                        return false;
                    }
                    $discount_money = round($detail['order_price'] - $data['money'], 2);
                    $discount_money = max($discount_money, 0);
                    break;
                case '2': //折扣
                    if ($data['rate'] < 1 || $data['rate'] > 99) {
                        $this->error = "请输入合理的折扣";
                        return false;
                    }
                    if ($data['rate'] < 1) {
                        $discount_ratio = -1;
                    } else {
                        $discount_ratio = $data['rate'];
                    }

                    break;
                case '3': //抹零

                    break;
            }

            if ($data['type'] == 2) { // 折扣
                // 重置改价和抹零
                $detail->save([
                    'discount_ratio' => $discount_ratio,
                    'is_change_price' => 1,
                    'discount_change_price' => 0,
                    'small_discount_type' => 0,
                    'small_auto' => 0,
                ]);
                (new OrderModel())->reloadPrice($detail['order_id']);
            } else if ($data['type'] == 3) { // 抹零
                // 重置改价
                $detail->save([
                    //                    'discount_ratio' => 0,  // TODO 不重置折扣
                    'discount_money' => 0,
                    'discount_change_price' => 0,   // 重置改价
                    'small_auto' => 0,
                ]);
                $o = (new OrderModel())->reloadPrice($detail['order_id']);  // TODO 这里现在会返回主单
                if ($data['discountType'] == 1) { //抹分
                    $discount_money = round($o['order_price'] - intval(($o['pay_price'] - $o['pay_fee_money']) * 10) / 10, 2);
                } elseif ($data['discountType'] == 2) { //抹角
                    $discount_money = round($o['order_price'] - intval(($o['pay_price'] - $o['pay_fee_money'])), 2);
                } elseif ($data['discountType'] == 3) { //四舍五入到角
                    $discount_money = round($o['order_price'] - round(($o['pay_price'] - $o['pay_fee_money']), 1), 2);
                } elseif ($data['discountType'] == 4) { //四舍五入到元
                    $discount_money = round($o['order_price'] - round(($o['pay_price'] - $o['pay_fee_money']), 0), 2);
                }
                //
                $pay_price = round($o['order_price'] - $discount_money, 2);
                // 积分奖励按照应付计算
                $setting = SettingModel::getSupplierItem(SettingEnum::POINTS, $detail['shop_supplier_id'], $detail['app_id']);
                if ($setting['is_shopping_gift']) {
                    // 积分赠送比例
                    $ratio = $setting['gift_ratio'] / 100;
                } else {
                    $ratio = 0;
                }
                $points_bonus = helper::bcmul($pay_price, $ratio, 3);
                $points_bonus = round($points_bonus, 2);
                //  减去会员优惠金额
                $discount_money = floatval(helper::bcsub($discount_money, $o['user_discount_money']));
                //
                $o->save([
                    'discount_money' => $discount_money < 0 ? 0 : $discount_money,
                    'pay_price' => $pay_price,
                    'points_bonus' => $points_bonus,
                    'is_change_price' => 1,
                    'small_discount_type' => $data['discountType'],
                    'small_auto' => 0,
                ]);
                (new OrderModel())->reloadPrice($detail['order_id']);
            } else {
                // 改价
                if ($data['money'] > $detail['order_price']) {
                    $pay_price = $data['money'];
                } else {
                    $pay_price = round($detail['order_price'] - $discount_money, 2);
                }
                if ($pay_price <= 0) {
                    $pay_price = 0;
                }

                //
                if ($data['type'] == 1) {   // 改价
                    // 积分奖励按照应付计算
                    $setting = SettingModel::getSupplierItem(SettingEnum::POINTS, $detail['shop_supplier_id'], $detail['app_id']);
                    if ($setting['is_shopping_gift']) {
                        // 积分赠送比例
                        $ratio = $setting['gift_ratio'] / 100;
                    } else {
                        $ratio = 0;
                    }
                    $points_bonus = helper::bcmul($pay_price, $ratio, 3);
                    $points_bonus = round($points_bonus, 2);
                    // 加上支付方式手续费
                    $total_fee_money = OrderPayType::where('order_id', $detail['order_id'])->sum('fee_money');
                    $pay_price = helper::bcadd($pay_price, $total_fee_money);
                    //
                    $detail->save([
                        'discount_money' => $discount_money < 0 ? 0 : $discount_money,
                        'pay_price' => $pay_price,
                        'discount_ratio' => $discount_ratio,
                        //                        'user_discount_money' => 0,  // TODO  改价不清除会员折扣了
                        'points_bonus' => $points_bonus,
                        'is_change_price' => 1,
                        'discount_change_price' => $data['money'] == 0 ? -1 : $data['money'],
                        'small_discount_type' => 0,
                        'small_diff_money' => 0,
                        'small_auto' => 0,
                    ]);
                } else {
                    $detail->save([
                        'discount_money' => $discount_money < 0 ? 0 : $discount_money,
                        'pay_price' => $pay_price,
                        'is_change_price' => 1,
                        'small_auto' => 0,
                    ]);
                }
                (new OrderModel())->reloadPrice($detail['order_id']);
            }
            // 是否合单
            if ($detail && $detail['merge_parent_id']) {
                $detail->reloadMasterMergeOrder($detail['merge_parent_id'], $detail['merge_id']);
            }

            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    //查询桌号信息
    public function changeTable($table_id)
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许转台";
            return false;
        }
        $orderInfo = self::getTableInfo($table_id);
        if ($orderInfo) {
            if ($orderInfo['order_source'] == 10) { //小程序下单
                if ($orderInfo['pay_status']['value'] == 20 && $orderInfo['order_status']['value'] == 10) {
                    $this->error = "台号已被使用";
                    return false;
                }
            } else { //收银台下单
                if ($orderInfo['order_status']['value'] == 10) {
                    $this->error = "台号已被使用";
                    return false;
                }
            }
        }
        return $this->save(['table_id' => $table_id]);
    }

    /**
     * 交换桌台(转台)
     * @param $old_table_id
     * @param $new_table_id
     * @return bool
     */
    public function exchangeTable($old_table_id, $new_table_id)
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许转台";
            return false;
        }

        // 禁止并发操作 - 开台/转台
        $queue = new QueueHelp('TABLE_ORDER_ALL_' . request()->appId . '_' . $new_table_id);
        $queue->while();

        $oldTable = TableModel::detail($old_table_id);
        $newTable = TableModel::detail($new_table_id);
        if (!$newTable) {
            $queue->release();
            $this->error = "新的桌台不存在";
            return false;
        }
        $this->startTrans();
        try {
            $this->save(['table_id' => $new_table_id, 'table_no' => $newTable['table_no']]);
            if ($this->subOrder()->count() > 0) {
                foreach ($this->subOrder as $subOrder) {
                    $subOrder->table_no = $newTable['table_no'];
                    $subOrder->save();
                }
            }
            TableModel::open($new_table_id);
            TableModel::close($oldTable->table_id);
            // 处理待接单数据桌台
            TakeOrder::where('order_id', $this->order_id)->update(['table_id' => $new_table_id]);
            //
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_CHANGE_TABLE, [
                'old' => [
                    'table_id' => $oldTable->table_id,
                    'table_no' => $oldTable->table_no,
                ],
                'new' => [
                    'table_id' => $new_table_id,
                    'table_no' => $newTable['table_no'],
                ],
            ], '转台');
            //
            $this->commit();
            $queue->release();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $queue->release();
            $this->rollback();
            return false;
        }
    }

    // 订单预结检查
    public function orderPayCheck($product_source = OrderModel::CASHIER_PRODUCT_SOURCE, $ignore_must = 0, $sub_order_id = 0)
    {
        // 送厨必点商品检查
        if ($ignore_must == 0 && !$this->checkSchemeMustProduct(2)) {
            $this->errorCode = $this->getErrorCode();
            $this->error = $this->getError();
            $this->errorData = $this->getErrorData();
            return false;
        }
        //
        $sourceProductList = $this->getOrderSourceProductList($product_source);
        $orderProductList = $sourceProductList['orderProductList'];
        $allProductList = $sourceProductList['allProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];

        // 判断是否下架 - 得到未送厨的产品
        $unSendKitchenProduct = array_values(array_filter($orderProductList, function ($orderProduct) {
            return $orderProduct['is_send_kitchen'] == 0;
        }));
        foreach ($unSendKitchenProduct as $orderProduct) {
            // 判断商品是否下架
            foreach ($allProductList as $product) {
                if ($product['is_delete'] == 1 || $product['product_status']['value'] == 20) {
                    if ($orderProduct['product_id'] == $product['product_id']) {
                        $this->error = __('商品') . ' ' . $product['product_name_text'] . ' ' . __('已下架，请选择其他商品');
                        $this->errorData = ['product_id' => $product['product_id']];
                        $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST;
                        return false;
                    }
                }
            }
            // 判断规格是否下架
            if (!isset($allProductSkuList[$orderProduct['product_sku_id']])) {
                $this->error = __('规格') . ' ' . $orderProduct['product_name_text'] . '-' . $orderProduct['product_attr'] . ' ' . __('已下架，请选择其他规格');
                $this->errorData = ['product_id' => $orderProduct['product_id'], 'product_sku_id' => $orderProduct['product_sku_id']];
                $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST_SKU;
                return false;
            }
        }

        // 付款减库存-判断库存
        $productArray = [];
        foreach ($orderProductList as $orderProduct) {
            if ($sub_order_id && $orderProduct['sub_order_id'] != $sub_order_id) {
                continue;
            }
            //
            $productSku = $allProductSkuList[$orderProduct['product_sku_id']] ?? [];
            if (
                empty($productSku['material'])
                && $orderProduct['is_return'] == 0
                && (
                    $orderProduct['deduct_stock_type'] == DeductStockTypeEnum::PAYMENT ||
                    (
                        $orderProduct['deduct_stock_type'] == DeductStockTypeEnum::CREATE
                        && $orderProduct['is_send_kitchen'] == 0
                    )
                )
            ) {
                $productArray[] = $orderProduct;
            }
        }
        if ($productArray) {
            $result = $this->getStockInsufficientProduct($product_source, $productArray, $allProductSkuList);
            if (!empty($result)) {
                $this->error = "以下商品库存不足，请删除后再下单";
                $this->errorData = $result;
                return false;
            }
        }

        //
        $this->startTrans();
        try {
            // 校验订单商品价格
            if (!$this->reloadOrderProductPrice($product_source, $sourceProductList)) {
                $this->error = '订单价格有变动，请重新确认后结账';
                $this->errorCode = OrderErrorEnum::RELOAD_PRICE;
                $this->reloadPrice($this['order_id']);
                $this->errorData = ['order_id' => $this->order_id, 'res' => CartModel::getHallCartOrderDetail(['shop_supplier_id' => $this->shop_supplier_id], 0, $this->order_id)];
                $this->commit();
                return false;
            }
            // 订单商品送厨前验证 - 库存，限购等相关
            $modelSendKitchen = new OrderProductModel;
            if (!$modelSendKitchen->sendKitchenBeforeVerify($this, 'payment', $sourceProductList)) {
                $this->error = $modelSendKitchen->getError();
                $this->errorData = $modelSendKitchen->getErrorData();
                $this->errorCode = $modelSendKitchen->getErrorCode();
                $this->rollback();
                return false;
            }
            //
            $this->rollback();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 订单支付
     * @param $data
     * @param $cashier
     * @param $device_id
     * @return bool
     */
    public function orderPay($data, $cashier = null, $device_id = '', $isCashierPaySuccess = true)
    {
        Db::connect($this->getConnection())->execute("SET SESSION sql_mode = ''");
        //
        if ($this['pay_status']['value'] != 10) {
            $this->error = "订单已支付";
            return false;
        }

        // 禁止并发操作 - 所有 - （添加商品,送厨等)
        $queueAll = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queueAll->while();

        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ORDER_PAY_' . $this->app_id);
        $queue->while();

        // 检查未送厨房
        if ($this->orderProductUnSendCount($this['order_id'])) {
            return $this->handleError("存在未送厨商品，请重新查看订单", 0, $queue, $queueAll);
        }

        // 前后端是否不一致
        $front_pay_price = isset($data['final_price']) ? $data['final_price'] : 0;
        $back_pay_price = helper::bcsub($this['pay_price'], $this['checkout_diff_money'] ?? 0);   // 订单标的 pay_price= 应付 + 手续费（也就是最终金额）- 结账抹零
        if ($front_pay_price != $back_pay_price) {
            $this->reloadPrice($this['order_id']);
            return $this->handleError("当前订单价格数据有变动，请重新查看订单", 0, $queue, $queueAll);
        }
        // 后端最后重新计算一次是否正确（可能最后一步后台配置又有变化） TODO
        $reloadOrder = $this->reloadPrice($this['order_id']);
        if ($reloadOrder->pay_price != $this['pay_price']) {
            return $this->handleError("当前订单价格数据有变动，请重新查看订单", 0, $queue, $queueAll);
        }
        // 获取保存数据
        $saveData = [];

        // 是否免单
        if (!OrderPayType::where(['order_id' => $this['order_id'], 'value' => -1])->find()) {
            // 检查付款金额
            $total_pay_type_price = $this->getPayTypeTotalPrice($this['order_id']);
            if ($total_pay_type_price < $back_pay_price) {
                return $this->handleError("实收金额不能小于应收金额", 0, $queue, $queueAll);
            }
            // 检查非现金超额
            $total_non_cash_pay_type_price = $this->getNonCashPayTypeTotalPrice($this['order_id']);
            if ($total_non_cash_pay_type_price > $back_pay_price) {
                return $this->handleError("收款金额大于最终应收，请先修改收款金额", 0, $queue, $queueAll);
            }

            // 检查支付方式
            if (!$this->checkPayTypeList()) {
                return $this->handleError($this->getError(), 0, $queue, $queueAll);
            }
            $change_due = helper::bcsub($total_pay_type_price, $back_pay_price);
            $saveData['actual_price'] = $total_pay_type_price;
            $saveData['change_due'] = $change_due;
        }

        $is_free = 0;
        if (isset($data['is_free'])) {
            if ($this->parent_id > 0) {
                $is_free = $saveData['is_free'] = self::where('order_id', $this->parent_id)->value('is_free');
                if (!$is_free) {
                    $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this->shop_supplier_id, $this->app_id);
                    $is_free = $saveData['is_free'] = $store['free_method'] == '10' ? 1 : 2;
                }
            } else {
                $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this->shop_supplier_id, $this->app_id);
                $is_free = $saveData['is_free'] = $store['free_method'] == '10' ? 1 : 2;
            }
        }
        if (isset($data['free_remark'])) {
            $saveData['free_remark'] = $data['free_remark'];
        }
        if (isset($data['user_id']) && $data['user_id'] > 0) {
            $saveData['user_id'] = $data['user_id'];
        }
        if ($cashier && isset($cashier['cashier_id'])) {
            $saveData['cashier_id'] = $cashier['cashier_id'];
        }

        // 并发检查
        $nowOrder = Order::where('order_id', $this['order_id'])->find();
        if (!$nowOrder) {
            return $this->handleError("订单不存在", 0, $queue, $queueAll);
        }
        if ($nowOrder->order_status['value'] == OrderStatusEnum::COMPLETED) {
            return $this->handleError("订单已支付", 0, $queue, $queueAll);
        }
        if ($nowOrder->order_status['value'] == OrderStatusEnum::CANCELLED) {
            return $this->handleError("订单已取消", 0, $queue, $queueAll);
        }
        // 执行更新
        $this->startTrans();
        try {
            // 更新消费税类型
            $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $this['shop_supplier_id'], $this['app_id']);
            $saveData['consumption_tax_type'] = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
            $saveData['settle_device_id'] = $device_id;

            //
            $this->save($saveData);
            $res = $this->onPayment($this['order_no'], $data['delivery'], false);
            if (!$res) {
                if ($this->errorCode = OrderErrorEnum::RELOAD_PRICE) {
                    $this->commit();
                }
                //
                return $this->handleError($this->getError(), 0, $queue, $queueAll);
            }

            // 处理组合支付
            $this->handlePayTypeList();

            // 执行订单完成后的操作（累计消费、积分、）
            $detail = Order::where('order_id', $this['order_id'])->find();
            $orderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
            $orderCompleteService->complete([$detail], $detail['app_id']);

            // 桌台订单关闭桌台
            if ($detail['table_id'] > 0 || $detail['parent_id'] > 0) {
                // 支付后是否清台
                if ($detail['parent_id'] == 0) {
                    $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this['shop_supplier_id'], $this['app_id']);
                    if ($store['no_clear_table'] == 0) {
                        TableModel::close($detail['table_id']);
                    }
                }
                // 清除接单数据
                if ($detail['parent_id'] == 0 || ($detail['parent_id'] > 0 && $detail['order_name'] == '1')) {
                    $takeOrders = TakeOrder::where('status', 0)->where(function ($q) use ($detail) {
                        $q->where('order_id', $detail['order_id']);
                        if ($detail['parent_id'] > 0) {
                            $q->whereOr('order_id', 'in', Order::where('parent_id', $this['parent_id'])->whereOr('order_id', $this['parent_id'])->column('order_id'));
                        } else {
                            $q->whereOr('order_id', 'in', Order::where('parent_id', $this['order_id'])->column('order_id'));
                        }
                    })->select();
                    /** @var TakeOrder $takeOrder */
                    foreach ($takeOrders as $takeOrder) {
                        $takeOrder->reject();
                    }
                }
            }

            //
            if ($is_free) {
                $freeSave['pay_price'] = 0;
                $freeSave['points_bonus'] = 0;
                $freeSave['discount_money'] = $this['pay_price'];
                $freeSave['free_pay_price'] = $this['pay_price'];
                $this->save($freeSave);
                if ($this->parent_id > 0) {
                    self::where('order_id', $this->parent_id)->update(['is_free' => $is_free]);
                }
            }

            // 添加操作记录
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_SETTLE, [
                'order_price' => $this->order_price,                    //  订单金额
                'pay_price' => $back_pay_price,                         //  应收金额 - 结账抹零
                'pay_type' => $this->payType,                           //  支付方式
                'actual_price' => $this->actual_price,                  //  实收金额
                'change_due' => $this->change_due,                      //  找零
                'is_free' => $this->is_free,                            //  是否免单
                'discount_money' => $this->free_pay_price,              //  免单金额
                'parent_id' => $this->parent_id,                        //  拆单主单ID
                'order_name' => $this->order_name,                      //  订单名称
            ], '结账');

            // 拆单-处理主单
            if ($detail->parent_id > 0) {
                // 完结主单
                self::handleOrderCompleted($detail->parent_id, [
                    'pay_time' => $detail->pay_time,
                    'settle_type' => $detail->settle_type,
                    'auto_close' => $detail->auto_close,
                    'close_time' => $detail->close_time,
                    'delivery_time' => $detail->delivery_time,
                    'receipt_time' => $detail->receipt_time,
                    'cashier_id' => $saveData['cashier_id'] ?? 0,
                ]);
            }

            // 重新算价
            $this->reloadPrice($this->parent_id > 0 ? $this->parent_id : $this->order_id, null, []);

            // 订单高峰时间段记录，主单才加峰值 - 等待所有数据处理完成再记入峰值，保证最终数值正确(v1.1.1版本)
            $PeakTimeOrderId = $detail->parent_id > 0 ? $detail->parent_id : $detail->order_id;
            self::handleOrderCompletedPeakTime($PeakTimeOrderId);

            //
            $this->commit();
            //
            $queue->release();
            $queueAll->release();
            //
            if ($isCashierPaySuccess) {
                event('CashierPaySuccess', $this);
            }
            //
            return $res;
        } catch (BaseException $e) {
            $this->rollback();
            return $this->handleError($e->getMessage(), 0, $queue, $queueAll);
        }
    }

    /**
     * 订单合并支付
     * @param $merge_id
     * @param $data
     * @param $cashier
     * @param $device_id
     * @return bool
     */
    public function orderMergePay($merge_id, $data, $cashier = null, $device_id = '')
    {
        Db::connect($this->getConnection())->execute("SET SESSION sql_mode = ''");
        //
        if ($this['pay_status']['value'] != 10) {
            $this->error = "订单已支付";
            return false;
        }
        //
        if (!$this['merge_parent_id']) {
            $this->error = "未生成合单主单";
            return false;
        }

        // 禁止并发操作 - 所有 - （添加商品,送厨等)
        $queueAll = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queueAll->while();

        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ORDER_PAY_' . $this['shop_supplier_id']);
        $queue->while();

        $masterOrder = OrderModel::with(['user'])->where('order_id', $this['merge_parent_id'])->find();
        if (!$masterOrder) {
            $this->error = "合单数据有变动，请重新查看订单";
            $queue->release();
            $queueAll->release();
            return false;
        }
        //
        $mergeOrderList = OrderModel::where('merge_id', $merge_id)->select();
        // 检查未送厨房
        foreach ($mergeOrderList as $order_item) {
            if ($this->orderProductUnSendCount($order_item['order_id'])) {
                $this->error = "存在未送厨商品，请重新查看订单";
                $queue->release();
                $queueAll->release();
                return false;
            }
        }
        // 前后端是否不一致
        $front_pay_price = isset($data['final_price']) ? $data['final_price'] : 0;
        $back_pay_price = $masterOrder['pay_price'];
        if ($front_pay_price != $back_pay_price) {
            foreach ($mergeOrderList as $order_item) {
                $order_item->reloadPrice($order_item['order_id']);
            }
            $this->reloadMasterMergeOrder($masterOrder['order_id'], $merge_id);
            $this->error = "订单价格有变动，请重新查看订单";
            $this->errorCode = OrderErrorEnum::RELOAD_PRICE;
            $queue->release();
            $queueAll->release();
            return false;
        }

        // 主单更新数据
        $updateMasterData = [];

        // 是否免单
        if (!OrderPayType::where(['order_id' => $masterOrder['order_id'], 'value' => -1])->find()) {
            // 检查付款金额
            $total_pay_type_price = $this->getPayTypeTotalPrice($masterOrder['order_id']);
            if ($total_pay_type_price < $back_pay_price) {
                $this->error = "实收金额不能小于应收金额";
                $queue->release();
                $queueAll->release();
                return false;
            }
            // 检查支付方式
            if (!$masterOrder->checkPayTypeList()) {
                $this->error = $masterOrder->error;
                $queue->release();
                $queueAll->release();
                return false;
            }
            $updateMasterData['actual_price'] = $total_pay_type_price;
            $updateMasterData['change_due'] = helper::bcsub($total_pay_type_price, $back_pay_price);
        }

        $is_free = isset($data['is_free']) ? $data['is_free'] : 0;
        if ($is_free) {
            $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this->shop_supplier_id, $this->app_id);
            $updateMasterData['is_free'] = $is_free = $store['free_method'] == '10' ? 1 : 2;
        }
        if (isset($data['free_remark'])) {
            $updateMasterData['free_remark'] = $data['free_remark'];
        }
        $user_id = $masterOrder->user_id ?: 0;
        if (isset($data['user_id']) && $data['user_id'] > 0) {
            $user_id = $data['user_id'];
            $updateMasterData['user_id'] = $data['user_id'];    //
        } else {
            $updateMasterData['user_id'] = $user_id;
        }
        if ($cashier && isset($cashier['cashier_id'])) {
            $updateMasterData['cashier_id'] = $cashier['cashier_id'];
        }
        // 更新消费税类型
        $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $this['shop_supplier_id'], $this['app_id']);
        $updateMasterData['consumption_tax_type'] = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
        $updateMasterData['settle_device_id'] = $device_id;
        $updateMasterData['table_id'] = $this['table_id'];

        // 支付后是否清台
        $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this['shop_supplier_id'], $this['app_id']);

        //
        $detail = null;
        $this->startTrans();
        try {
            // 处理子订单
            $k = 0;
            $is_reload = 0;
            foreach ($mergeOrderList as $orderItem) {
                $saveData = [
                    'actual_price' => $orderItem['pay_price'],
                    'user_id' => $user_id,
                    'cashier_id' => $cashier['cashier_id'],
                    'settle_device_id' => $device_id,
                    'consumption_tax_type' => $updateMasterData['consumption_tax_type'],
                    'is_free' => $is_free,
                    'free_remark' => isset($data['free_remark']) ? $data['free_remark'] : '',
                    'pay_price' => $is_free ? 0 : $orderItem['pay_price'],
                    'free_pay_price' => $is_free ? $orderItem['pay_price'] : 0,
                    'discount_money' => $is_free ? $orderItem['order_price'] : $orderItem['discount_money'],
                    'points_bonus' => $is_free ? 0 : $orderItem['points_bonus'],
                ];
                $res = $orderItem->onPayment($orderItem['order_no'], $data['delivery'], false);
                if ($orderItem->getErrorCode() == OrderErrorEnum::RELOAD_PRICE) {
                    $is_reload = 1;
                    continue;
                }
                if (!$res) {
                    $this->errorData = [
                        'table_no' => $orderItem['table_no'],
                        'errorData' => $orderItem->errorData,
                    ];
                    $this->error = $orderItem->error;
                    $this->rollback();
                    $queue->release();
                    $queueAll->release();
                    return false;
                }
                $orderItem->save($saveData);
                $k++;
            }
            if ($is_reload) {
                // 重算主单
                if ($this['merge_parent_id']) {
                    $this->reloadMasterMergeOrder($this['merge_parent_id'], $this['merge_id']);
                }
                $this->errorData = [
                    'table_no' => $orderItem['table_no'],
                    'errorData' => $orderItem->errorData,
                ];
                $this->errorCode = OrderErrorEnum::RELOAD_PRICE;
                $this->errorData = ['res' => CartModel::getHallCartOrderDetail(['shop_supplier_id' => $this['shop_supplier_id']], 0, $this['order_id']), 'order_id' => $this['order_id']];
                $this->error = $orderItem->error;
                $this->commit();
                $queue->release();
                $queueAll->release();
                return false;
            }
            // 处理组合支付
            $masterOrder->handlePayTypeList();
            //
            if ($is_free) {
                $this->reloadMasterMergeOrder($this['merge_parent_id'], $this['merge_id']);
            }
            // 更新主单数据
            $detail = $this->updateMasterMergeOrderPayComplete($masterOrder['order_id'], $updateMasterData);
            // 执行订单完成后的操作（累计消费、积分、）
            $orderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
            if (!$orderCompleteService->complete([$detail], $detail['app_id'])) {
                $this->error = "执行订单完成操作失败";
                $queue->release();
                $queueAll->release();
                return false;
            }
            // 关闭桌台
            if ($store['no_clear_table'] == 0) {
                foreach ($mergeOrderList as $order_item) {
                    TableModel::close($order_item['table_id']);
                }
            }
            // 更新店铺账户余额
            if ($cashAmount = $detail->getCashReceivePriceAttr(null, $detail)) {
                Account::updateAmount(1, $cashAmount, $detail->order_no, $cashier['cashier_id'], $detail->shop_supplier_id, $detail->app_id, 'order-pay');
            }
            //
            $this->commit();
            $queue->release();
            $queueAll->release();
        } catch (BaseException $e) {
            $queue->release();
            $queueAll->release();
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
        //
        if ($detail) {
            event('CashierPaySuccess', $detail);
        }
        //
        return true;
    }

    /**
     * 退菜
     * @param $order_product_id
     * @param $num
     * @param $return_reason
     * @return bool
     */
    public function removeProduct($order_product_id, $num, $return_reason = '')
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单已完成,不允许退菜";
            return false;
        }

        $orderProduct = OrderProduct::detail($order_product_id);
        if (!$orderProduct) {
            $this->error = "当前状态不可操作";
            return false;
        }

        if ($orderProduct['total_num'] < $num) {
            $this->error = "退菜数量不能大于当前商品数量";
            return false;
        }
        $this->startTrans();
        try {
            $isPay = $this['pay_status']['value'] == 20 ? 1 : 0;
            // 退回商品库存
            ProductFactory::getFactory($this['order_source'])->backProductStock([$orderProduct], $isPay);
            if ($orderProduct['total_num'] == $num) {
                $orderProduct->force()->delete();
            } else {
                $total_num = $orderProduct['total_num'] - $num;
                $orderProduct->save([
                    'total_num' => $total_num,
                ]);
            }
            // 退菜记录
            if ($num > 0) {
                OrderProductReturn::add([
                    'order_id' => $this['order_id'],
                    'order_product_id' => $order_product_id,
                    'product_id' => $orderProduct['product_id'],
                    'num' => $num,
                    'reason' => $return_reason,
                ]);
            }
            //
            $this->reloadPrice($this['order_id']);
            // 重算主单
            if ($this['merge_parent_id']) {
                $this->reloadMasterMergeOrder($this['merge_parent_id'], $this['merge_id']);
            }
            // 添加操作记录
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_REFUND_PRODUCT, [
                'order_product_id' => $order_product_id,
                'product_id' => $orderProduct['product_id'],
                'product_name' => $orderProduct['product_name'],
                'product_attr' => $orderProduct->getData('product_attr'),
                'num' => $num,
                'reason' => $return_reason,
                'custom_reason' => "",
                'parent_id' => $this['parent_id'],      // 拆单主单ID
                'order_name' => $this['order_name'],    // 订单名称
                'remark' => $orderProduct['remark'],    // 商品备注
            ], '退菜');
            //
            $this->commit();
            // 打印退菜单
            $this['product'] = [$orderProduct];
            (new OrderPrinterService)->printProductTicket($this, Printing::PRINT_TYPE_BACK_FOOD);
            //
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    //结账完成
    public function settle()
    {
        if ($this['pay_status']['value'] != 20) {
            $this->error = "订单未付款，不允许操作";
            return false;
        }
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单状态错误，不允许操作";
            return false;
        }
        return $this->transaction(function () {
            // 更新订单状态：已发货、已收货
            $status = $this->save([
                'delivery_status' => 20,
                'delivery_time' => time(),
                'receipt_status' => 20,
                'receipt_time' => time(),
                'order_status' => OrderStatusEnum::COMPLETED
            ]);
            // 执行订单完成后的操作
            $OrderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
            $OrderCompleteService->complete([$this], static::$app_id);
            return $status;
        });
    }

    /**
     * 营业数据
     */
    public function businessData($params)
    {
        Db::connect($this->getConnection())->execute("SET SESSION sql_mode = ''");
        //
        $queryMode = $params['mode'] ?? 0;
        $categoryType = $params['category_type'] ?? 1;
        $shopSupplierId = $params['shop_supplier_id'] ?? 0;
        $shopCashierId = $params['cashier_id'] ?? 0;
        //
        $repository =  new OrderBusinessDataRepository($this, $params);
        $model = $repository->getBaseModel($this, $params);
        [$startTime, $endTime] = $repository->getTimes();
        //
        $all = [];
        $categorys = [];
        $products = [];
        $incomes = [];
        //
        switch ($queryMode) {
                // -------
                // 全部
                // -------
            case 0:
                $all = $repository->getBusinessData();
                // 高峰时间段
                $all['peak_hour_list'] = OrderPeakTime::getMaxRecord($startTime, $endTime, $shopCashierId);
                // 税收百分比对象列表
                $all['percentage_list'] = $repository->getPercentageList();
                break;
                // -------
                // 按商品分类
                // -------
            case 2:
                $categorys = $model->clone()
                    ->leftJoin('order_product rp', 'a.order_id = rp.order_id')
                    ->leftJoin('product p', 'p.product_id = rp.product_id')
                    ->leftJoin('category c', 'c.category_id = p.category_id')
                    ->when($categoryType, function ($q) use ($categoryType) {
                        if ($categoryType == 1) {
                            $q->leftJoin('category cc', 'cc.category_id = IF(c.parent_id = 0, c.category_id, c.parent_id)');
                            $q->where('cc.parent_id', 0);
                            $q->group('cc.category_id');
                            $q->field('cc.category_id, cc.name');
                        } else {
                            $q->where('c.parent_id', '>', 0);
                            $q->group('c.category_id');
                            $q->field('c.category_id, c.name');
                        }
                    })
                    ->where('a.is_merge', 0)
                    ->where('rp.is_return', 0) // 不包含退货
                    ->field("sum(rp.total_num) as sales, sum(rp.total_pay_price) as prices")
                    ->select()
                    ->append([])?->toArray();
                foreach ($categorys as $key => &$data) {
                    $data['parent_id'] = $categoryType == 1 ? 0 : Category::where('category_id', $data['category_id'])->value('parent_id');
                    $categorys[$key]['name_text'] = Category::getPathNameTextAttr($data['name'] ?: '', $data);
                    $categorys[$key]['prices'] = Helper::number2($data['prices']);
                }
                break;
                // -------
                // 商品列表
                // -------
            case 3:
                $products = $model->clone()
                    ->leftJoin('order_product rp', 'a.order_id = rp.order_id')
                    ->leftJoin('product p', 'p.product_id = rp.product_id')
                    ->leftJoin('product_sku sku', 'sku.product_sku_id = rp.product_sku_id and p.product_id = sku.product_id')
                    ->field("sum(rp.total_num) as sales, sum(rp.total_price) as prices, p.product_name, rp.product_id, sku.product_price, sku.spec_name")
                    ->group("rp.product_sku_id")
                    ->where("rp.product_id", '>', 0)
                    ->where('rp.is_return', 0) // 不包含退货
                    ->select()
                    ->append([])?->toArray();
                foreach ($products as $key => &$data) {
                    $specNameText = ProductSku::getSpecNameTextAttr($data['spec_name'] ?: '', $data);
                    $products[$key]['product_name_text'] = Product::getProductNameTextAttr($data['product_name'] ?: '', $data) . ($specNameText ? " ($specNameText)" : '');
                    $products[$key]['prices'] = Helper::number2($data['prices']);
                    $products[$key]['product_price'] = Helper::number2($data['product_price']);
                }
                break;
        }
        // 收入列表
        $totalIncomeAll = 0;
        if ($queryMode != 3) {
            $incomes = $repository->getIncomesList();
            foreach ($incomes as $key => &$income) {
                if ($income['pay_type'] == -1) {
                    if ($queryMode == 0) {
                        unset($incomes[$key]);
                    } else {
                        $income['pay_type_way'] = $income['pay_type_name'] = __('免单金额');
                    }
                }
            }
            $incomes = array_values($incomes);
            // 计算显示支付方式 - 总收入（除去免单）
            $totalIncomeAll = array_sum(array_column($incomes, 'price'));
        }
        //
        $refundAmount = Helper::number2($model->clone()->where('a.is_merge', 0)->sum("refund_money")); // 退款金额：支付方式/商品分类
        $totalAmount = Helper::number2($model->clone()->where('a.is_merge', 0)->sum("pay_price")); // 实收金额：支付方式/商品分类
        return [
            'all' => $all,
            'categorys' => $categorys,
            'incomes' => $incomes,
            'total_income_all' => Helper::number2($totalIncomeAll),
            'products' => $products,
            'sales_num' => $model->clone()->where('a.is_merge', 0)->count(),
            'refund_amount' => $refundAmount,
            'total_amount' => helper::bcsub($totalAmount, $refundAmount),
            'times' => [$startTime, $endTime],
            'supplier' => Supplier::field('shop_supplier_id,business_id,name,address,description,link_name,link_phone,logo,app_id')
                ->where('shop_supplier_id', $shopSupplierId)
                ->find()?->toArray(),
        ];
    }

    // 修改订单商品价格
    public function changeProductPrice($order_product_id, $money)
    {
        $this->startTrans();
        try {
            if ($money < 0) {
                $this->error = "价格错误";
                return false;
            }
            $p = OrderProduct::where('order_product_id', '=', $order_product_id)->find();
            if (!$p) {
                $this->error = "商品不存在";
                return false;
            }
            $p->product_price = $money;
            $p->total_price = helper::bcmul($money, $p->total_num);
            if ($p->save()) {
                // 更新
                $this->reloadPrice($this['order_id']);
                $this->commit();
                return true;
            } else {
                $this->error = "商品不存在";
                return false;
            }
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    // 更新商品总价
    public function updateTotalPrice()
    {
        // 商品总价 - 优惠抵扣
        $total_price = 0;
        foreach ($this['product'] as $product) {
            $total_price = helper::bcadd($total_price, $product['total_price']);
        }
        $order_price = helper::bcadd($total_price, $this['service_money']);
        $pay_price = round(helper::bcsub($order_price, $this['discount_money']), 2);
        return $this->save(['total_price' => $total_price, 'order_price' => $order_price, 'pay_price' => $pay_price]);
    }

    // 订单使用会员
    public function useMember($user_id)
    {
        $this->startTrans();
        try {
            $user_id = !empty($user_id) ? $user_id : 0;
            // 订单表更新user_id
            if ($this['merge_id']) {
                // 已有支付方式不能在使用会员
                if ($this['merge_parent_id'] > 0) {
                    $have_paid = OrderPayType::where('order_id', $this['merge_parent_id'])->find();
                    if ($have_paid) {
                        $this->error = $this->table_id > 0 ? '当前桌台已被部分支付，不支持变更会员信息' : '当前订单已被部分支付，不支持变更会员信息';
                        return false;
                    }
                }
                // 对所有兄弟单使用会员
                $orderList = Order::where('merge_id', $this['merge_id'])->select();
                foreach ($orderList as $subOrder) {
                    if ($subOrder->small_discount_type > 0 || $subOrder->discount_change_price > 0 || $subOrder->discount_change_price == -1) {
                        $this->errorCode = OrderErrorEnum::RESET_DISCOUNT_NOTICE;
                    }
                    $updateData = [
                        'user_id' => $user_id,
                        'small_discount_type' => 0,
                        //                    'discount_ratio' => 0,    // 不重置折扣
                        'discount_money' => 0,
                        'discount_change_price' => 0,
                    ];
                    $updateData['is_change_price'] = $subOrder->discount_ratio > 0 ? 1 : 0; // 不重置折扣
                    $subOrder->save($updateData);
                    $subOrder->reloadPrice($subOrder->order_id);
                }
                // 删掉余额支付方式
                if ($user_id == 0) {
                    OrderPayType::where('order_id', $this['merge_parent_id'])->where('value', 10)->delete();
                }
                if ($this['merge_parent_id']) {
                    // 重算主单
                    $param = [
                        'user_id' => $user_id,
                    ];
                    $this->reloadMasterMergeOrder($this['merge_parent_id'], $this['merge_id'], $param);
                }
            } else {
                // 已有支付方式不能在使用会员
                $have_paid = OrderPayType::where('order_id', $this['order_id'])->find();
                if ($have_paid) {
                    $this->error = $this->table_id > 0 ? '当前桌台已被部分支付，不支持变更会员信息' : '当前订单已被部分支付，不支持变更会员信息';
                    return false;
                }
                // 删掉余额支付方式
                if ($user_id == 0) {
                    OrderPayType::where('order_id', $this['order_id'])->where('value', 10)->delete();
                }
                if ($this->small_discount_type > 0 || $this->discount_change_price > 0 || $this->discount_change_price == -1) {
                    $this->errorCode = OrderErrorEnum::RESET_DISCOUNT_NOTICE;
                }
                $updateData = [
                    'user_id' => $user_id,
                    'small_discount_type' => 0,
                    //                    'discount_ratio' => 0,    // 不重置折扣
                    'discount_money' => 0,
                    'discount_change_price' => 0,
                ];
                $updateData['is_change_price'] = $this->discount_ratio > 0 ? 1 : 0; // 不重置折扣
                $this->save($updateData);
                // 重载订单价格信息
                $this->reloadPrice($this['order_id']);
                // 使用折扣+会员折扣，改价/抹零优惠折扣次数需要重置 v1.1.0
                $mainOrderId = $this['parent_id'] > 0 ? $this['parent_id'] : $this['order_id'];
                $subOrderId = $this['parent_id'] > 0 ? $this['order_id'] : 0;
                (new OrderAbnormalLog)->resetDiscount($mainOrderId, $subOrderId, [1, 3]);
            }
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 挂单列表
     * @return array
     */
    public function getStayList()
    {
        $list = $this->with(['product'])
            ->where('is_stay', '=', 1)
            ->where('order_status', '=', OrderStatusEnum::NORMAL)
            ->select()->toArray();
        foreach ($list as &$order) {
            /** @var Order $o */
            $o = (new Order)->with(['product'])
                ->where('order_id', $order['order_id'])
                ->find();
            $productList = $o->product()
                ->group('main_order_product_id, is_free, remark, product_price')
                ->field('
                    *,
                    sum(total_num) as total_num,
                    sum(finish_num) as finish_num,
                    sum(total_price) as total_price,
                    sum(total_product_price) as total_product_price,
                    sum(refund_money) as refund_money,
                    sum(refund_num) as refund_num,
                    sum(tax_rate) as tax_rate,
                    sum(consumption_tax) as consumption_tax
                ')
                ->select();
            $order['product'] = $productList;
            $order['total_price'] = $o->getBackPayPrice();
        }
        return $list;
    }

    /**
     * 订单挂单
     * @param $order_id
     * @return Order|false
     */
    public function stayOrder($order_id)
    {
        // 检查订单状态
        $detail = OrderModel::detail([
            ['order_id', '=', $order_id],
            ['order_status', '=', OrderStatusEnum::NORMAL]
        ]);
        if (!$detail) {
            $this->error = '当前状态不可操作';
            return false;
        }
        // 添加操作记录
        OrderOperationLog::createLog($order_id, OrderOperationLog::ACTION_STAY_ORDER, '', '挂单');
        //
        return $this->where('order_id', '=', $order_id)->update(['is_stay' => 1, 'stay_time' => time()]);
    }

    /**
     * 订单取单
     * @param $order_id
     * @return Order|false
     */
    public function pickOrder($order_id)
    {
        // 检查订单状态
        $detail = OrderModel::detail([
            ['order_id', '=', $order_id],
            ['order_status', '=', OrderStatusEnum::NORMAL]
        ]);
        if (!$detail) {
            $this->error = '当前状态不可操作';
            return false;
        }
        // 添加操作记录
        OrderOperationLog::createLog($order_id, OrderOperationLog::ACTION_PICK_ORDER, '', '取单');
        //
        if (!$this->where('order_id', '=', $order_id)->update(['is_stay' => 0])) {
            $this->error = '取单失败';
            return false;
        }
        $orderId = $detail->order_id;
        if ($detail->subOrder->count() > 0) {
            foreach ($detail->subOrder as $subOrder) {
                if ($subOrder->getData('pay_status') == OrderPayStatusEnum::PENDING) {
                    $orderId = $subOrder->order_id;
                    break;
                }
            }
        }
        return $orderId;
    }

    /**
     * 订单挂单数量
     * @return int
     */
    public function stayOrderNum()
    {
        return $this->where('is_stay', '=', 1)->where('order_status', '=', 10)->count();
    }

    /**
     * 反结账
     * @param $new_table_id
     * @param $to_order_id
     * @param $device_id
     * @return bool
     */
    public function reverseSettle($new_table_id = 0, $to_order_id = 0, $device_id = '')
    {
        // 检查订单状态
        if ($this->pay_status['value'] !== 20) {
            $this->error = '当前状态不可操作';
            return false;
        }
        //
        $this->startTrans();
        try {
            // 是否拆单
            $isSplitOrder = $this->subOrder()->count() > 0;
            // 先挂单再放进去
            if ($this->order_source == OrderSourceEnum::CASHIER) {
                $to_order_id && $this->stayOrder($to_order_id);
            } else {
                // 再次开台
                if ($new_table_id) {
                    $newTable = TableModel::detail($new_table_id);
                    if ($newTable->status != 10) {
                        $this->error = '当前桌台被占用，请选择其他桌台打开';
                        return false;
                    }
                    TableModel::open($new_table_id);
                    // 更新主单桌台信息
                    $this->save(['table_id' => $new_table_id, 'table_no' => $newTable['table_no']]);
                    // 更新子单桌台信息
                    if ($this->subOrder()->count() > 0) {
                        OrderModel::where('parent_id', $this->order_id)->update(['table_no' => $newTable['table_no']]);
                    }
                } else {
                    if ($this->tables?->status != 10 && $this->tables?->underwayOrder && $this->tables?->underwayOrder?->order_id != $this->order_id) {
                        $this->error = '当前桌台有待付款订单，请选择其他桌台打开';
                        return false;
                    }
                    if ($this->tables?->status != 10) {
                        $this->error = '当前桌台被占用，请选择其他桌台打开';
                        return false;
                    }
                    TableModel::open($this->table_id);
                }
            }
            // 反商品库存、销量
            ProductFactory::getFactory($this->order_source)->updateStockSales($this->getOrderSourceProductList(), 'inc');
            // 反自助餐销量
            if ($this->is_buffet == 1) {
                $orderBuffets = OrderBuffetModel::where('order_id', $this->order_id)->select();
                if ($orderBuffets) {
                    foreach ($orderBuffets as $orderBuffet) {
                        foreach ($orderBuffet->buffet()->select() as $buffet) {
                            $buffet->save(['sale_num' => ['dec', $orderBuffet->num]]);
                        }
                    }
                }
            }
            // 反材料库存
            (new Product)->salesOutReverse($this->order_id);
            // 反会员 消费金额
            if (!$isSplitOrder) {
                if ($this->user_id && $this->user) {
                    $price = helper::bcsub($this->pay_price, $this->refund_money);
                    // 累积用户总消费金额
                    $this->user->setDecPayMoney($price);
                    // 余额支付-反结账返还
                    $balancePay = OrderPayType::where('order_id', $this->order_id)->where('value', OrderPayTypeEnum::BALANCE)->find();
                    if ($balancePay) {
                        BalanceLogModel::add(BalanceLogSceneEnum::REVERSE, [
                            'order_id' => $this->order_id,
                            'user_id' => $this->user?->user_id,
                            'card_id' => $this->user?->card_id,
                            'money' => $balancePay->price,
                        ], ['order_no' => $this->order_no]);
                    }
                }
            } else {
                foreach ($this->subOrder as $subOrder) {
                    if ($subOrder->user_id && $subOrder->user) {
                        $price = helper::bcsub($subOrder->pay_price, $subOrder->refund_money);
                        // 累积用户总消费金额
                        $subOrder->user->setDecPayMoney($price);
                        // 余额支付-反结账返还
                        $balancePay = OrderPayType::where('order_id', $subOrder->order_id)->where('value', OrderPayTypeEnum::BALANCE)->find();
                        if ($balancePay) {
                            BalanceLogModel::add(BalanceLogSceneEnum::REVERSE, [
                                'order_id' => $subOrder->order_id,
                                'user_id' => $subOrder->user?->user_id,
                                'card_id' => $subOrder->user?->card_id,
                                'money' => $balancePay->price,
                            ], ['order_no' => $subOrder->order_no]);
                        }
                    }
                }
            }

            // 反结算
            if (!$isSplitOrder) {
                if ($this->eat_type == 10 && ($this->settle_type == 10 || $this->settle_type == 20)) {
                    $orderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
                    $orderCompleteService->settled([$this], 'reverse');
                }
            } else {
                foreach ($this->subOrder as $subOrder) {
                    if ($subOrder->eat_type == 10 && ($subOrder->settle_type == 10 || $subOrder->settle_type == 20)) {
                        $orderCompleteService = new OrderCompleteService(OrderTypeEnum::MASTER);
                        $orderCompleteService->settled([$subOrder], 'reverse');
                    }
                }
                $this->onBatchUpdate([$this->order_id], ['is_settled' => 0]);
            }

            // 订单高峰时间段记录
            (new OrderPeakTime)->record('desc', $this->order_id);
            // 反订单相关状态
            $this->save([
                'order_status' => OrderStatusEnum::NORMAL,
                'device_id' => $device_id,
                'pay_status' => OrderPayStatusEnum::PENDING,
                'pay_time' => 0,
                'delivery_status' => 10,
                'delivery_time' => 0,
                'receipt_status' => 10,
                'receipt_time' => 0,
                'is_stay' => 0,
                'is_lock' => 0,
                'refund_money' => 0,
                'actual_price' => 0,
                'is_free' => 0,
                'free_remark' => '',
                'discount_money' => $this->is_free > 0 ? 0 : $this->discount_money,
                'pay_price' => $this->is_free > 0 ? $this->discount_money : $this->pay_price,
                'free_pay_price' => 0,
            ]);
            if ($isSplitOrder) {
                foreach ($this->subOrder as $subOrder) {
                    $subOrder->save([
                        'order_status' => OrderStatusEnum::NORMAL,
                        'device_id' => $device_id,
                        'pay_status' => OrderPayStatusEnum::PENDING,
                        'pay_time' => 0,
                        'delivery_status' => 10,
                        'delivery_time' => 0,
                        'receipt_status' => 10,
                        'receipt_time' => 0,
                        'is_stay' => 0,
                        'is_lock' => 0,
                        'refund_money' => 0,
                        'actual_price' => 0,
                        'is_free' => 0,
                        'free_remark' => '',
                        'discount_money' => $subOrder->is_free > 0 ? 0 : $subOrder->discount_money,
                        'pay_price' => $subOrder->is_free > 0 ? $subOrder->discount_money : $subOrder->pay_price,
                        'free_pay_price' => 0,
                    ]);
                }
            }
            // 返店铺账户余额
            if ($cashAmount = $this->getCashReceivePriceAttr(null, $this)) {
                Account::updateAmount(0, $cashAmount, $this->order_no, $this->cashier_id, $this->shop_supplier_id, $this->app_id, 'order-reverse-settle');
            }
            // 删除之前的组合支付
            OrderPayType::where('order_id', $this->order_id)->delete();
            if ($isSplitOrder) {
                foreach ($this->subOrder as $subOrder) {
                    OrderPayType::where('order_id', $subOrder->order_id)->delete();
                }
            }
            // 清除之前的免单记录
            OrderFree::where('order_id', $this->order_id)->delete();
            if ($isSplitOrder) {
                foreach ($this->subOrder as $subOrder) {
                    OrderFree::where('order_id', $subOrder->order_id)->delete();
                }
            }
            // 重算订单（清除支付方式手续费价格会变动）
            $this->reloadPrice($this->order_id);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 重计算价格
     */
    public function reloadOrderProductPrice($productSource = OrderModel::CASHIER_PRODUCT_SOURCE, $sourceProductList = [])
    {
        $sourceProductList = $sourceProductList ?: $this->getOrderSourceProductList($productSource);
        // 当前的订单产品
        $orderProductList = $sourceProductList['orderProductList'];
        $orderProductList = array_filter($orderProductList, function ($orderProduct) {
            return $orderProduct['is_change_price'] == 0;
        });
        // 取得所有的产品
        $productDetailAll = $sourceProductList['allProductList'];
        // 取得所有的产品sku
        $productSkuListAll = $sourceProductList['allProductSkuList'];
        // 取得自助餐商品ids
        if ($this->parent_id == 0) {
            $now_buffet_ids = (new OrderBuffet)->where('order_id', $this->order_id)->column('buffet_id');
        } else {
            $now_buffet_ids = (new OrderBuffet)->where('order_id', $this->parent_id)->column('buffet_id');
        }
        $now_buffet_product_ids_arr = $now_buffet_ids ? Buffet::getBuffetProductIds($now_buffet_ids) : [];
        //
        $diffNum = 0;
        // 检查商品
        foreach ($orderProductList as $orderProduct) {
            // 取得原始数据
            $productDetail = $productDetailAll[$orderProduct->product_id];
            //
            $feed_price = 0;
            $is_exist_feed = false;
            if ($orderProduct->feed_ids) {
                $feedIds = is_array($orderProduct->feed_ids) ? $orderProduct->feed_ids : json_decode($orderProduct->feed_ids);
                foreach ($productDetail['feed'] as $feed) {
                    if (in_array($feed['product_feed_id'], $feedIds)) {
                        $feed_price = helper::bcadd($feed_price, $feed['price']);
                        $is_exist_feed = true;
                    }
                }
                $feed_price = floatval($feed_price);
            }
            if ($feed_price == 0 && !$is_exist_feed) {
                $feed_price = $orderProduct->feed_price ?: 0;
            }

            //
            if (in_array($orderProduct->product_id, $now_buffet_product_ids_arr)) {
                // 自助餐商品
                if ($feed_price != $orderProduct->product_price) {
                    $updateArr = [
                        'product_price' => $feed_price,
                        'total_price' => $totalPrice = $orderProduct->total_num * $feed_price,
                        'total_pay_price' => $totalPrice,
                        'is_buffet_product' => 1,
                    ];
                    $orderProduct->save($updateArr);
                    $diffNum++;
                }
            } else {
                // 非餐商品
                if (isset($productSkuListAll[$orderProduct->product_sku_id])) {
                    $product_price = $productSkuListAll[$orderProduct->product_sku_id]['product_price'] ?? 0;
                    $product_price = helper::bcadd($product_price, $feed_price);
                    //
                    if ($product_price != $orderProduct->product_price) {
                        $updateArr = [
                            'product_price' => $product_price,
                            'total_price' => $totalPrice = $orderProduct->total_num * $product_price,
                            'total_pay_price' => $totalPrice,
                            'is_buffet_product' => 0,
                        ];
                        $orderProduct->save($updateArr);
                        $diffNum++;
                    }
                }
            }
        }

        // 检查自助餐结价格
        if ($this->parent_id == 0) {
            $orderBuffetCustomerList = (new OrderBuffetCustomer)->where('order_id', $this->order_id)->select();
        } else {
            $orderBuffetCustomerList = (new OrderBuffetCustomer)->where('sub_order_id', $this->parent_id)->select();
        }

        foreach ($orderBuffetCustomerList as $orderBuffetCustomer) {
            // 取得原始数据
            $buffetCustomerDetail = BuffetCustomer::where('buffet_id', '=', $orderBuffetCustomer->buffet_id)->where('customer_type_id', '=', $orderBuffetCustomer->customer_type_id)->find();
            if ($buffetCustomerDetail && $orderBuffetCustomer->price != $buffetCustomerDetail->price) {
                $updateArr = [
                    'price' => $buffetCustomerDetail->price,
                    'total_price' => $orderBuffetCustomer->num * $buffetCustomerDetail->price,
                ];
                $orderBuffetCustomer->save($updateArr);
                $diffNum++;
            }
        }
        // 结果
        return $diffNum > 0 ? false : true;
    }

    /**
     * 打印小票
     */
    public function printSmall($deviceId = '', $userId = 0, $data = '', $isLock = true)
    {
        $printLang = is_array($data) ? ($data['print_lang'] ?? '') : $data;
        // 发送打印
        if ($printLang) {
            request()->language = $printLang;
        } else {
            $printerConfig = SettingModel::getSupplierItem('printer', $this->shop_supplier_id, $this->app_id);
            request()->language = $printerConfig['default_language'] ?? '';
        }
        // 更新消费税类型
        $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $this['shop_supplier_id'], $this['app_id']);
        $allProductInfo = (new CartModel())->getOrderCartDetail(['shop_supplier_id' => $this->shop_supplier_id], $this->table_id, $this->order_id);
        //
        $order = (clone $this);
        if ($order->pay_status['value'] == OrderPayStatusEnum::PENDING) {
            $order->pay_price = helper::bcsub($allProductInfo['perOrderInfo']['pay_price'], $allProductInfo['perOrderInfo']['checkout_diff_money'] ?? 0);
            $order->total_product_price = $allProductInfo['sumInfo']['subtotal'];
            $order->consumption_tax_money = $allProductInfo['sumInfo']['total_consumption_tax_money'];
            $order->consumption_tax_type = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
            $order->settle_device_id = $deviceId;
            $order->service_money = $allProductInfo['sumInfo']['service_money'];
            $order->discount_money = $allProductInfo['sumInfo']['special_discount'];
            $order->user_discount_money = $allProductInfo['sumInfo']['total_user_discount_money'];
            if ($userId) {
                $order->cashier_id = $userId;
                $order->cashier = $order->cashier()->find();
            }
        }
        /** @var OrderPrinterService $printerService */
        $printerService = (new OrderPrinterService);
        $printerData = $printerService->printTicket($order, false, $deviceId, $data, true, true);
        //
        request()->language = '';
        //
        if (!$printerData) {
            $this->error = $printerService->getError() ?: '打印失败，未连接打印机';
            return false;
        }
        // 锁定主单和子单
        if ($isLock && $this->pay_status['value'] != OrderPayStatusEnum::SUCCESS) {
            if ($this->parent_id == 0) {
                $this->is_lock = 1;
                $this->lock_time = time();
                $this->save();
            } else {
                self::where('order_id', $this->parent_id)
                    ->whereOr('parent_id', $this->parent_id)
                    ->save(['is_lock' => 1, 'lock_time' => time()]);
            }
        }
        //
        return [
            'order_id' => $this->order_id,
            'printer_data' => $printerData,
        ];
    }
}
