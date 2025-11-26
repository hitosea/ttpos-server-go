<?php

namespace app\common\model\order;

use app\cashier\model\order\Order as OrderModel;
use think\Model;
use help\QueueHelp;
use help\ClientHelp;
use think\facade\Db;
use think\facade\Cache;
use app\common\library\helper;
use app\common\model\delay\Delay;
use app\common\model\store\Table;
use app\common\model\shop\Account;
use app\common\model\buffet\Buffet;
use app\common\model\store\FreeTag;
use app\common\model\store\PayType;
use think\db\exception\DbException;
use app\common\enum\http\StatusCode;
use app\common\model\BaseModelOrder;
use app\common\model\product\Product;
use app\common\model\store\TakeOrder;
use app\common\model\user\BalanceLog;
use app\common\exception\BaseException;
use app\common\model\supplier\Printing;
use app\common\enum\order\OrderTypeEnum;
use app\common\model\product\ProductSku;
use app\common\model\store\ReturnReason;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\settings\SettingEnum;
use app\common\model\buffet\CustomerType;
use app\common\enum\order\OrderStatusEnum;
use app\common\model\payment\PaymentOrder;
use app\common\service\order\OrderService;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\buffet\BuffetCustomer;
use app\common\model\buffet\BuffetDiscount;
use app\common\model\user\User as UserModel;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\OrderOperationLog;
use think\db\exception\DataNotFoundException;
use think\db\exception\ModelNotFoundException;
use app\common\model\store\Table as TableModel;
use app\common\enum\product\DeductStockTypeEnum;
use app\common\service\order\OrderPrinterService;
use app\common\model\order\OrderRefundDestination;
use app\common\service\order\OrderCompleteService;
use app\common\model\store\PayType as PayTypeModel;
use app\common\model\product\Product as ProductModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\service\product\factory\ProductFactory;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model\user\CardRecord as CardRecordModel;
use app\common\repositories\OrderBusinessDataRepository;
use app\common\model\product\ProductSku as ProductSkuModel;
use app\tablet\model\product\Product as TabletProductModel;
use app\common\model\order\OrderProduct as OrderProductModel;
use app\assistant\model\product\Product as AssistantProductModel;
use app\cashier\service\order\settled\CashierOrderSettledService;
use app\common\model\order\OrderBuffetCustomer as OrderBuffetCustomerModel;
use app\common\model\order\OrderDelay as OrderDelayModel;
use app\common\model\order\OrderBuffet as OrderBuffetModel;
use help\HttpHelp;

/**
 * 订单模型模型
 */
class Order extends BaseModelOrder
{

    /**
     * 订单详情
     * @param $where
     * @param array $with
     * @param string[] $field
     * @return array|Model|null|self
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function detail($where, $with = ['user', 'address', 'buffet', 'buffetCustomerType', 'buffetDiscount', 'delay', 'product' => ['image', 'productSku.material'], 'extract', 'supplier', 'cashier', 'payType'], $field = ['*'])
    {
        is_array($where) ? $filter = $where : $filter['order_id'] = (int) $where;
        return self::with($with)->where($filter)->order('order_id', 'desc')->field($field)->find();
    }

    /**
     * 订单详情
     * @param $where
     * @param $with
     * @param $field
     * @return false|object
     */
    public function underwayDetail($where, $with = ['user', 'address', 'buffet', 'buffetCustomerType', 'buffetDiscount', 'delay', 'product' => ['image', 'productSku.material'], 'extract', 'supplier', 'cashier', 'payType'], $field = ['*'])
    {
        is_array($where) ? $filter = $where : $filter['order_id'] = (int) $where;
        $order = self::with($with)->where($filter)->where('order_status', OrderStatusEnum::NORMAL)->order('order_id', 'desc')->field($field)->find();
        if (!$order) {
            $this->error = '订单不存在';
            return false;
        }
        if ($order->is_lock == 1) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }
        return $order;
    }

    /**
     * 订单详情（包含删除的订单记录信息）
     * @param $where
     * @param string[] $with
     * @return array|Model|null
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function detailWithTrashed($where, $orderField = null, $orderAppendField = [])
    {
        $filter = is_array($where) ? $where : ['order_id' => (int) $where];
        //
        $orderField = $orderField ?: [
            'order_id',
            'parent_id',
            'order_name',
            'extra_times',
            'eat_type',
            'delivery_type',
            'meal_num',
            'order_type',
            'pay_status',
            'order_status',
            'is_merge',
            'merge_id',
            'merge_parent_id',
            'order_no',
            'table_id',
            'table_no',
            'is_buffet',
            'total_price',
            'total_product_price',
            'order_price',
            'pay_price',
            'points_bonus',
            'surplus_balance',
            'refund_money',
            'actual_price',
            'pay_time',
            'create_time',
            'user_id',
            'order_source',
            'call_no',
            'app_id',
            'cashier_id',
            'shop_supplier_id',
            'device_id',
            'settle_device_id',
            'cancel_remark',
            'is_free',
            'free_remark',
            'table_remark',
        ];
        $orderProductField = [
            'order_product_id',
            'order_id',
            'sub_order_id',
            'product_id',
            'feed_ids',
            'product_sku_id',
            'image_id',
            'product_name',
            'product_attr',
            'product_price',
            'tax_rate',
            'tax_calc_type',
            'consumption_tax',
            'total_num',
            'total_pay_price',
            'total_price',
            'total_product_price',
            'total_price',
            'refund_money',
            'refund_num',
            'is_send_kitchen',
            'delete_time',
            'add_source',
            'is_change_price',
            'is_buffet_product',
            'is_move',
            'is_free',
            'free_remark',
            'product_consumption_tax',
            'product_service_fee',
            'product_service_consumption_tax',
            'product_original_consumption_tax',
            'product_original_service_fee',
            'product_original_service_consumption_tax',
            'product_discount_money',
            'is_return',
        ];
        $buffetCustomerTypeField = [
            'id',
            'order_id',
            'sub_order_id',
            'buffet_id',
            'buffet_name',
            'customer_type_name',
            'tax_calc_type',
            'tax_rate',
            'consumption_tax',
            'num',
            'price',
            'total_pay_price',
            'total_price',
            'refund_money',
            'refund_num',
            'product_consumption_tax',
            'product_service_fee',
            'product_service_consumption_tax',
            'product_original_consumption_tax',
            'product_original_service_fee',
            'product_original_service_consumption_tax',
        ];
        // 以下查询field指定返回了字段，可能会漏某些地方的字段导致问题 TODO
        return self::with([
            'user' => function ($query) {
                $query->field(['user_id', 'nickname as nickName', 'balance', 'gift_balance', 'grade_id', 'card_id']);
            },
            'address',
            'buffet' => function ($query) {
                $query->field(['id', 'order_id', 'name', 'total_price']);
            },
            'buffetCustomerType' => function ($query) use ($buffetCustomerTypeField) {
                $query->field($buffetCustomerTypeField);
            },
            'delay' => function ($query) {
                $query->field(['id', 'order_id', 'sub_order_id', 'name', 'num', 'price', 'total_price', 'refund_money', 'refund_num']);
            },
            'payType',
            'parentOrder.payType',
            'refundType',
            'extract',
            'cashier',
            'product' => function ($query) use ($orderProductField) {
                $query->withTrashed()
                    ->field($orderProductField)
                    ->where('is_send_kitchen', 1)
                    ->with(['image', 'productReturn']);
            },
            'mergeList' => function ($query) use ($orderField, $orderProductField, $buffetCustomerTypeField) {
                $query->with([
                    'user' => function ($query) {
                        $query->field(['user_id', 'nickname as nickName', 'balance']);
                    },
                    'address',
                    'buffet' => function ($query) {
                        $query->field(['id', 'order_id', 'name', 'total_price']);
                    },
                    'buffetCustomerType' => function ($query) use ($buffetCustomerTypeField) {
                        $query->field($buffetCustomerTypeField);
                    },
                    'delay' => function ($query) {
                        $query->field(['id', 'order_id', 'name', 'num', 'price', 'total_price', 'refund_money', 'refund_num']);
                    },
                    'payType',
                    'refundType',
                    'extract',
                    'cashier',
                    'product' => function ($query) use ($orderProductField) {
                        $query->withTrashed()
                            ->field($orderProductField)
                            ->where('is_send_kitchen', 1)
                            ->with(['image', 'productReturn']);
                    }
                ])->field($orderField);
            },
            'subOrder' => function ($query) use ($orderField, $orderProductField, $buffetCustomerTypeField) {
                $query->with([
                    'user' => function ($query) {
                        $query->field(['user_id', 'nickname as nickName', 'balance']);
                    },
                    'address',
                    'buffet' => function ($query) {
                        $query->field(['id', 'order_id', 'name', 'total_price']);
                    },
                    'buffetCustomerType' => function ($query) use ($buffetCustomerTypeField) {
                        $query->field($buffetCustomerTypeField);
                    },
                    'delay' => function ($query) {
                        $query->field(['id', 'order_id', 'name', 'num', 'price', 'total_price', 'refund_money', 'refund_num']);
                    },
                    'payType',
                    'refundType',
                    'extract',
                    'cashier',
                    'product' => function ($query) use ($orderProductField) {
                        $query->withTrashed()
                            ->field($orderProductField)
                            ->where('is_send_kitchen', 1)
                            ->with(['image', 'productReturn']);
                    }
                ])
                    ->field($orderField)
                    ->order('order_id', 'asc');
            },
        ])
            ->field($orderAppendField)
            ->field($orderField)
            ->where($filter)
            ->find();
    }

    /**
     * 获取桌台进行中订单
     * @param $table_id
     * @return array|Model|null
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function getTableUnderwayOrder($table_id)
    {
        return self::with(['user', 'product.productSku.material', 'buffet', 'buffetCustomerType'])->where([
            ['table_id', '=', $table_id],
            ['table_id', '<>', 0],
            ['order_status', '=', OrderStatusEnum::NORMAL],
            ['is_merge', '=', 0]
        ])->order('order_id', 'desc')->find();
    }

    /**
     * 扫码端获取桌台进行中订单
     * @param $table_id
     * @return array|Model|null
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function getScanTableUnderwayOrder($table_id)
    {
        $tableOrder = self::with(['user', 'product.productSku.material', 'buffet', 'buffetCustomerType'])->where([
            ['table_id', '=', $table_id],
            ['table_id', '<>', 0],
            ['order_status', '=', OrderStatusEnum::NORMAL],
            ['is_merge', '=', 0]
        ])->order('order_id', 'desc')->find();
        if (!$tableOrder) {
            $table = Table::where('table_id', $table_id)->where('status', 30)->find();
            if ($table) {
                throw new BaseException(['msg' => '桌台已结帐', 'code' => StatusCode::CHECKOUT_ERROR]);
            } else {
                throw new BaseException(['msg' => '桌台已关闭', 'code' => StatusCode::VISIT_ERROR]);
            }
        }
        return $tableOrder;
    }

    /**
     * 自助餐到期提醒
     * @param mixed $tid
     * @param mixed $buffet_expired_time
     * @param mixed $tablet_end_time_minute
     * @return int
     */
    public static function buffetTimeRemind($tid, $buffet_expired_time, $tablet_end_time_minute, $app_id = 0)
    {
        $second = $tablet_end_time_minute * 60;
        $now_timestamp = time();
        $buffet_remaining_time = $buffet_expired_time - $now_timestamp;
        $lock = Cache::get("{$app_id}_remind::{$tid}");
        if (($buffet_remaining_time < $second) && !$lock) {
            Cache::set("{$app_id}_remind::{$tid}", 1, $second);
            return 1;
        }
        return 0;
    }

    /**
     * 订单详情
     * @param $where
     * @param string[] $with
     * @return array|Model|null
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function detailByNo($order_no, $with = ['user', 'address', 'product' => ['image', 'refund'], 'extract', 'express', 'extractStore.logo', 'extractClerk', 'supplier'])
    {
        return self::with($with)->where('order_no', '=', $order_no)->find();
    }

    /**
     * 批量获取订单列表
     * @param $orderIds
     * @param array $with
     * @return array
     */
    public function getListByIds($orderIds, $with = [])
    {
        $data = $this->getListByInArray('order_id', $orderIds, $with);
        return helper::arrayColumn2Key($data, 'order_id');
    }

    /**
     * 批量更新订单
     * @param $orderIds
     * @param $data
     * @return bool
     */
    public function onBatchUpdate($orderIds, $data)
    {
        return $this->where('order_id', 'in', $orderIds)->save($data);
    }

    /**
     * 批量获取订单列表
     * @param $field
     * @param $data
     * @param array $with
     * @return \think\Collection
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    private function getListByInArray($field, $data, $with = [])
    {
        return $this->with($with)->where($field, 'in', $data)->select();
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
     * 确认核销（自提订单）
     * @param $extractClerkId
     * @return bool|mixed
     */
    public function verificationOrder()
    {
        if ($this['pay_status']['value'] != 20 || in_array($this['order_status']['value'], [20, 30])) {
            $this->error = '该订单不满足核销条件';
            return false;
        }
        return $this->transaction(function () {
            // 更新订单状态：已发货、已收货
            $status = $this->save([
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
     * 获取已付款订单总数 (可指定某天)
     */
    public function getOrderData($startDate, $endDate, $type, $shop_supplier_id, $order_type = -1)
    {
        $model = $this;

        if (!is_null($startDate)) {
            $model = $model->where('pay_time', '>=', strtotime($startDate));
            $endDate = $endDate ?? $startDate;
            $model = $model->where('pay_time', '<', strtotime($endDate) + 86400);
        }

        if ($order_type >= 0) {
            $model = $model->where('order_type', '=', $order_type);
        }

        $model = $model->where('pay_status', '=', 20)->where('order_status', '<>', 20);

        if ($type != 'order_refund_money' && $type != 'order_refund_total') {
            $model = $model->where('is_merge', '=', 0);
        }

        switch ($type) {
            case 'order_total': // 订单数量
                return $model->count();
            case 'order_total_price': // 订单数量
                return $model->sum('pay_price');
            case 'order_user_total': // 支付用户数
                return count($model->distinct(true)->column('user_id'));
            case 'order_refund_money': // 退款金额
                return $model->sum('refund_money');
            case 'order_refund_total': // 退款订单数
                return $model->where('refund_money', '>', 0)->count();
            case 'order_discount_money': // 折扣总金额
                return Helper::bcadd($model->sum('discount_money'), $model->sum('user_discount_money'));
            case 'discount_money': // 优惠折扣
                return $model->sum('discount_money');
            case 'user_discount_money': // 会员折扣
                return $model->sum('user_discount_money');
            case 'income_price': // 预计收入
                return Helper::bcsub($model->sum('pay_price'), $model->sum('refund_money'));
            default:
                return 0;
        }
    }

    /**
     * 交易记录列表
     */
    public function getRecordList($data, $type = 0)
    {
        $model = $this;

        //订单状态
        if (isset($data['order_status']) && $data['order_status']) {
            switch ($data['order_status']) {
                case '1': //待支付
                    $model = $model->where('pay_status', '=', 10)->where('order_status', '<>', 20);
                    break;
                case '2': //进行中
                    $model = $model->where('pay_status', '=', 20)->where('order_status', '=', 10);
                    break;
                case '3': //已取消
                    $model = $model->where('order_status', '=', 20);
                    break;
                case '4': //已完成
                    $model = $model->where('pay_status', '=', 20)->where('order_status', '=', 30);
                    break;
            }
        }
        //订单类型
        if (isset($data['order_type']) && $data['order_type'] >= 0) {
            $model = $model->where('order_type', '=', $data['order_type']);
        }
        //支付方式
        if (isset($data['pay_type']) && $data['pay_type']) {
            $model = $model->where('pay_type', '=', $data['pay_type']);
        }
        //查询日期
        switch ($data['type']) {
            case '1': //今天
                $model = $model->where('create_time', '>=', strtotime(date('Y-m-d')));
                break;
            case '2': //近7天
                $model = $model->where('create_time', '>=', strtotime(-6 . ' days', strtotime(date('Y-m-d'))));
                break;
            case '3': //近15天
                $model = $model->where('create_time', '>=', strtotime(-14 . ' days', strtotime(date('Y-m-d'))));
                break;
            case '4': //自定义
                $start = strtotime($data['time'][0]);
                $end = strtotime($data['time'][1]) + 86399;
                $model = $model->where('create_time', 'between', "$start,$end");
                break;
            default:
                $model = $model->where('create_time', '>=', strtotime(date('Y-m-d')));
                break;
        }
        // 获取数据列表
        if ($type == 0) {
            return $model->order(['create_time' => 'desc'])
                ->paginate($data);
        } else {
            return $model->order(['create_time' => 'desc'])
                ->select();
        }
    }

    /**
     * 获取各类型总销售额
     */
    public function getOrderTotalMoney($order_type, $shop_supplier_id, $data = [])
    {
        $model = $this;
        $userModel = UserModel::where('delete_time', '=', 0);
        if (isset($data['type']) && $data['type']) {
            switch ($data['type']) {
                case '1': //今天
                    $model = $model->where('create_time', '>=', strtotime(date('Y-m-d')));
                    break;
                case '2': //近7天
                    $model = $model->where('create_time', '>=', strtotime(-6 . ' days', strtotime(date('Y-m-d'))));
                    break;
                case '3': //近15天
                    $model = $model->where('create_time', '>=', strtotime(-14 . ' days', strtotime(date('Y-m-d'))));
                    break;
                case '4': //自定义
                    $start = strtotime($data['time'][0]);
                    $end = strtotime($data['time'][1]) + 86399;
                    $model = $model->where('create_time', 'between', "$start,$end");
                    $userModel = $userModel->where('create_time', 'between', "$start,$end");
                    break;
            }
        }

        $model = $model->where('pay_status', '=', 20)
            ->where('order_status', '<>', 20)
            ->where('order_type', '=', $order_type);
        $detail['express_price'] = helper::number2($model->sum('express_price') ?: 0); //配送费
        $detail['bag_price'] = helper::number2($model->sum('bag_price') ?: 0); //包装费
        $detail['product_price'] = helper::number2($model->sum('total_price') ?: 0); //商品总金额
        $detail['refund_money'] = helper::number2($model->sum('refund_money') ?: 0); //退款金额
        $detail['total_price'] = helper::number2($model->sum('pay_price') ?: 0); //订单总金额（营业总额）
        $detail['income_money'] = helper::number2(round($detail['total_price'] - $detail['refund_money'], 2)); //预计收入
        $detail['order_count'] = $model->count(); //有效订单数量
        // 有效用户数量
        $detail['user_count'] = $userModel->count();
        // 折扣总额(优惠折扣 + 会员折扣)
        $discount_money = $model->sum('discount_money') ?: 0;
        $user_discount_money = $model->sum('user_discount_money') ?: 0;
        $detail['total_discount_money'] = helper::bcadd($discount_money, $user_discount_money, 2);
        return $detail;
    }

    /**
     * 店内概况统计
     */
    public function storeOverview($params)
    {
        $repository = new OrderBusinessDataRepository($this, $params);
        [$startTime, $endTime] = $repository->getTimes();
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/business', [
            'query_start_time' => $startTime,
            'query_end_time' => $endTime,
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $this->error = $res['message'] ?? '请求失败';
            return false;
        }

        $data = $res['data'];

        $incomes = [];
        $paymentMethodIncomes = $data['payment_method_incomes'];
        foreach ($paymentMethodIncomes as $paymentMethodIncome) {
            $incomes[] = [
                'pay_type' => $paymentMethodIncome['code'],
                'pay_type_name' => $paymentMethodIncome['name'],
                'price' => $paymentMethodIncome['amount'],
                'order_num' => $paymentMethodIncome['order_num'],
            ];
        }

        return [
            'receivable_price' => $data['total_sales'],
            'received_price' => $data['total_received_price'],
            'product_num' => $data['total_product_num'],
            'user_count' => $data['member_data']['user_count'],
            'user_discount_money' => $data['total_user_discount_money'],
            'business_price' => $data['total_pay_price'],
            'service_money' => $data['total_service_money'],
            'pay_fee_money' => $data['total_pay_fee_money'],
            'consumption_tax_money' => $data['total_tax_money'],
            'refund_money' => $data['total_refund_money'],
            'discount_money' => $data['total_discount_money'],
            'discount_ratio' => $data['total_discount_ratio'] . '%',
            'free_product_price' => $data['total_give_product_price'],
            'free_product_num' => $data['total_give_product_num'],
            'free_order_price' => $data['total_free_order_price'],
            'free_order_num' => $data['total_free_order_num'],
            'recharge_amount' => $data['member_data']['recharge_amount'],
            'delivery_order_amount' => $data['total_takeout_sale_amount'], // 外送订单总额
            'delivery_order_revenue' => $data['total_takeout_business_amount'], // 外送营收
            'delivery_order_refund_amount' => $data['total_takeout_refund_amount'], // 外送订单退款
            'delivery_fee' => $data['total_takeout_delivery_fee'], // 配送费
            'delivery_order_num' => $data['all_takeout_order_num'], // 外送订单数
            'delivery_min_order_price' => $data['all_takeout_min_order_price'], // 外送最小金额
            'delivery_max_order_price' => $data['all_takeout_max_order_price'], // 外送最大金额
            'delivery_avg_order_price' => $data['all_takeout_avg_order_price'], // 外送平均订单金额
            'total_order_num' => $data['total_order_num'],
            'min_order_price' => $data['min_order_price'],
            'max_order_price' => $data['max_order_price'],
            'avg_order_price' => $data['avg_order_price'],
            'table_order_num' => $data['all_table_order_num'],
            'table_people_num' => $data['all_table_people_num'],
            'table_min_order_price' => $data['all_table_min_order_price'],
            'table_max_order_price' => $data['all_table_max_order_price'],
            'table_avg_order_price' => $data['all_table_avg_order_price'],
            'cashier_order_num' => $data['all_cashier_order_num'],
            'cashier_min_order_price' => $data['all_cashier_min_order_price'],
            'cashier_max_order_price' => $data['all_cashier_max_order_price'],
            'cashier_avg_order_price' => $data['all_cashier_avg_order_price'],
            'takeaway_order_num' => $data['all_takeaway_order_num'],
            'takeaway_min_order_price' => $data['all_takeaway_min_order_price'],
            'takeaway_max_order_price' => $data['all_takeaway_max_order_price'],
            'takeaway_avg_order_price' => $data['all_takeaway_avg_order_price'],
            'incomes' => $incomes,
        ];
    }

    /**
     * 商品销售统计
     */
    public function productSales($params)
    {
        $repository = new OrderBusinessDataRepository($this, $params);
        //
        return $repository->getProductData($params);
    }

    /**
     * 店内概况统计根据每日显示（导出数据）
     */
    public function storeOverviewByDate($params)
    {
        $start_time = isset($params['date'][0]) ? $params['date'][0] : 0;
        $end_time = isset($params['date'][1]) ? $params['date'][1] : 0;

        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/export', [
            'query_start_time' => strtotime($start_time),
            'query_end_time' => strtotime($end_time) + 86399,
        ], [
            'Authorization: Bearer ' . $params['token'],
            'Accept-Language: ' . $params['language'],
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $this->error = $res['message'] ?? '请求失败';
            return false;
        }
        $data = $res['data'];
        return $data;
    }

    /**
     * 获取商品销量Top10 1
     */
    public function getProductRank($type, $product_type, $shop_supplier_id = 0, $data = [])
    {
        $start_time = isset($data['date'][0]) ? $data['date'][0] : 0;
        $end_time = isset($data['date'][1]) ? $data['date'][1] : 0;

        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/product_rank', [
            'query_start_time' => intval(strtotime($start_time)),
            'query_end_time' => intval(strtotime($end_time) + 86399),
            'rank_type' => intval($type + 1),
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $this->error = $res['message'] ?? '请求失败';
            return false;
        }
        $data = $res['data'];
        $list = [];
        foreach ($data['ranks'] as $item) {
            $list[] = [
                'product_name_text' => $item['product_name'],
                'total_num' => $item['sales_num'],
                'total_price' => $item['sales_price'],
            ];
        }

        return $list;
    }

    /**
     * 获取区域数据
     */
    public function regionData($params = [])
    {
        $repository =  new OrderBusinessDataRepository($this, $params);
        return $repository->getRegionData();
    }


    /**
     * 修改桌台就餐人数
     */
    public function updateMealNum($meal_num)
    {
        $this->startTrans();
        try {
            // 检查桌台状态
            if ($this['order_status']['value'] != OrderStatusEnum::NORMAL) {
                $this->error = '订单已结束';
                return false;
            }
            if ($this['is_buffet']) {
                $this->updateBuffetMealNum($this['order_id'], $meal_num);
                $this->updateDelayMealNum($this['order_id'], $meal_num);
            }
            $old_meal_num = $this->meal_num;
            $this->save(['meal_num' => $meal_num]);
            // 如果有子单，则同时修改子单就餐人数
            if ($this->subOrder->count() > 0) {
                foreach ($this->subOrder as $sub_order) {
                    $sub_order->save(['meal_num' => $meal_num]);
                }
            }
            $this->reloadPrice($this['order_id']);
            // 添加操作记录
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_UPDATE_MEAL_NUM, [
                'old_meal_num' => $old_meal_num,
                'new_meal_num' => $meal_num,
            ], '修改桌台就餐人数');
            //
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 取消订单
     * @param $extractClerkId
     * @return bool|mixed
     */
    public function cancels()
    {
        if ($this->pay_status['value'] == 20) {
            $this->error = "订单已付款，不允许取消";
            return false;
        }
        if ($this->order_status['value'] != 10) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        // 关闭桌台
        if ($this->table_id) {
            TableModel::close($this->table_id);
        }
        return $this->save(['order_status' => 20]);
    }

    /**
     * 删除订单
     * @param $extractClerkId
     * @return bool|mixed
     */
    public function remove()
    {
        if ($this->pay_status['value'] == 20) {
            $this->error = "订单已付款，不允许删除";
            return false;
        }
        if ($this->order_status['value'] != 20) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        $this->delete_time = 1;
        $this->save();
        return $this->delete($this->order_id);
    }

    /**
     * 删除订单
     */
    public function orderDelete()
    {
        if ($this->pay_status['value'] == 20) {
            $this->error = "订单已付款，不允许删除";
            return false;
        }
        if ($this->order_status['value'] != 20) {
            $this->error = "订单状态错误，不允许取消";
            return false;
        }
        $this->startTrans();
        try {
            $this->delete_time = 1;
            $this->save();
            $this->delete();
            $this->product()->delete();  // 删除订单商品
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 订单 已送厨的
     * 重新计算订单价格信息（服务费+消费税+会员折扣+自助餐+加钟费）
     * @param $order_id
     * @param $sub_order     // 子订单实例(用于拆单子单计算)
     * @return Order
     */
    public function reloadPrice($order_id, $sub_order = null, $setting_param = [])
    {
        $order_query_field = [
            // 订单基本信息
            'order_id',
            'parent_id',
            'order_no',
            'order_name',
            'order_source',
            'pay_status',

            // 关联ID
            'app_id',
            'shop_supplier_id',
            'user_id',
            'table_id',

            // 订单类型和状态
            'delivery_type',
            'is_buffet',
            'is_free',

            // 价格相关
            'discount_ratio',
            'discount_method',
            'discount_change_price',
            'discount_money',
            'free_pay_price',
            'actual_price',
            'original_price',
            'is_change_price',

            // 抹零设置
            'small_discount_type',
            'checkout_discount_type',
            'checkout_diff_money',

            // 其他
            'meal_num'
        ];
        $order_product_query_field = ['product_id, is_enable_grade, is_alone_grade, alone_grade_equity, alone_grade_type, is_points_gift, sales_initial, sales_actual'];
        /** @var OrderModel $order */
        $order = $sub_order ?: (isset($setting_param['masterOrder']) ? $setting_param['masterOrder'] : self::detail($order_id, [
            'product' => function ($q) use ($order_product_query_field) {
                $q->withoutField('product_name')->with([
                    'product' => function ($q) use ($order_product_query_field) {
                        $q->field($order_product_query_field);
                    }
                ]);
            },
            'supplier' => function ($q) {
                $q->field('shop_supplier_id, app_id, settle_type, service_type, service_money');
            },
            'user' => function ($q) {
                $q->field('user_id, grade_id');
            },
            'buffetCustomerType',
            'subOrder' => function ($q) use ($order_query_field) {
                $q->field($order_query_field);
            },
        ], $order_query_field));

        /**
         * 算价基础数据
         */
        // 商品
        $delivery_type = is_numeric($order['delivery_type']) ? $order['delivery_type'] : $order['delivery_type']['value'];
        $allProductIds = isset($setting_param['allProductIds']) ? $setting_param['allProductIds'] : array_unique(array_column($order->product?->toArray(), 'product_id'));
        $allProductTaxList = isset($setting_param['allProductTaxList']) ? $setting_param['allProductTaxList'] : $order->getOrderProductTaxRateList($allProductIds, $delivery_type == 40 ? 1 : 2);
        $allBuffetTaxList = $order->getBuffetTaxRateList();

        // 设置
        $setting = isset($setting_param['setting']) ? $setting_param['setting'] : SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        $serviceFee = $setting[SettingEnum::SERVICE_CHARGE]['values'];      // 服务费设置
        $pointsSetting = $setting[SettingEnum::POINTS]['values'];           // 积分设置
        $consumptionTaxSetting = $setting[SettingEnum::TAX_RATE]['values']; // 消费税设置
        $is_buffet = $order['is_buffet'];
        // 是否存在拆单子单
        if (!$sub_order && count($order['subOrder']) > 0) {
            $totals = [
                'discount_money' => 0,
                'total_price' => 0,
                'total_product_price' => 0,
                'order_price' => 0,
                'pay_price' => 0,
                'points_bonus' => 0,
                'service_money' => 0,
                'setting_service_money' => 0,
                'consumption_tax_money' => 0,
                'original_consumption_tax_money' => 0,
                'user_discount_money' => 0,
                'small_diff_money' => 0,
                'pay_fee_money' => 0,
                'total_product_service_fee' => 0,
                'total_product_service_consumption_tax' => 0,
                'total_product_consumption_tax' => 0,
                'free_pay_price' => 0,
                'actual_price' => 0,
                'original_price' => 0,
                'checkout_diff_money' => 0,
            ];
            //
            $is_change_price = 0;
            foreach ($order['subOrder'] as $sub_order_item) {
                $sub_order = $this->reloadPrice($sub_order_item->order_id, $sub_order_item, $setting_param);
                // 检查子单是否改价
                if ($sub_order['is_change_price'] == 1) {
                    $is_change_price = 1;
                }
                foreach ($totals as $key => &$value) {
                    $value += $sub_order->$key;
                }
            }
            //
            if (($order['table_id'] ?? 0) > 0) {
                Cache::set($order['table_id'] . '_table_price' . $order['app_id'], floatval($totals['pay_price']));
            }
            //
            if ($is_change_price) {
                $totals['is_change_price'] = 1;
            }
            $order->save($totals);
            //
            return $order;
        }
        //
        $service_charge_rate = $serviceFee['service_charge_rate'] ?: 0; // 商品服务费率
        //
        $pay_money = 0;
        $order_price = 0;
        $points_bonus = 0;
        $user_discount_money = 0;
        $total_product_discount_money = 0;
        $consumption_tax = 0;                           // 商品消费税 + 商品服务费
        $o_consumption_tax = 0;                         // 商品消费税 + 商品服务费（原价）
        $total_product_consumption_tax = 0;             // 商品消费税
        $total_product_service_fee = 0;                 // 商品服务费
        $o_total_product_service_fee = 0;               // 商品服务费（原价）
        $total_product_service_consumption_tax = 0;     // 商品服务费消费税
        $o_total_product_service_consumption_tax = 0;   // 商品服务费消费税（原价）
        $meal_num = $order['meal_num'] ?? 0;            // 就餐人数
        // 已送厨+未送 商品的累计
        $all_pay_money = 0;
        $all_consumption_tax = 0;
        $all_total_product_consumption_tax = 0;
        $all_total_product_service_consumption_tax = 0;
        $all_total_product_service_fee = 0;
        // 订单门店服务费
        $setting_service_money = ($serviceFee['is_open'] && $serviceFee['charge_type'] == 1) ? $serviceFee['service_charge'] : 0;
        //
        $productUpdateArr = [];

        /**
         * 子单算价数据
         */
        $order_product = $sub_order ? $order->getSubOrderProduct($order['product'], $sub_order->order_id) : $order['product'];
        $orderBuffetCustomerType = $sub_order ? $order->getSubOrderBuffetCustomerType($order['buffetCustomerType'], $sub_order->order_id) : $order['buffetCustomerType'];
        $order = $sub_order ?: $order;
        $order_id = $sub_order ? $sub_order->order_id : $order_id;

        // 是否优惠折扣比例
        $discount_ratio = 1;
        $discount_method = $order['discount_method'] ?? 10; // 折扣计算方式 10-按百分比 20-直接减免
        if ($order['discount_ratio'] > 0) {
            if ($discount_method == 20) {
                $discount_ratio = (100 - $order['discount_ratio']) / 100;
            } else {
                $discount_ratio = $order['discount_ratio'] / 100;
            }
        }

        foreach ($order_product as $product) {
            // 标记参与会员折扣
            $is_user_grade = false;
            // 会员等级抵扣的金额
            $grade_ratio = 0;
            // 会员折扣的商品单价
            $grade_product_price = 0;
            // 会员折扣的总额差
            $grade_total_money = 0;
            $user = null;
            $unit_pay_price = $product['product_price'];
            if ($product['product']['is_enable_grade'] && $product['total_price'] > 0) {
                $user = $order['user'];
                if ($user) {
                    $discount = (new CardRecordModel)->getDiscount($user['user_id']);
                } else {
                    $discount = 0;
                }
                $alone_grade_type = 10;
                // 商品单独设置了会员折扣  （折扣类型 alone_grade_type 10-百分比 20-固定金额）
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
                    //
                    $is_user_grade = !($discountRatio == 1);
                    $grade_ratio = $discountRatio == 1 ? 0 : $discountRatio;
                    // 订单的会员折扣 （原商品总价 - 折扣后
                    $grade_total_money = helper::number2(helper::bcsub($product['product_price'] * $product['total_num'], $gradeTotalPrice, 3));
                    // 商品应付单价
                    $unit_pay_price = $grade_product_price;
                }
            }
            //
            $product_points_bonus = 0;
            if ($pointsSetting['is_shopping_gift']) {
                // 积分赠送比例
                $ratio = $pointsSetting['gift_ratio'] / 100;
                // 计算抵扣积分数量
                $product_points_bonus = !$product['product']['is_points_gift'] ? 0 : helper::bcmul($unit_pay_price * $product['total_num'], $ratio, 2);
            }
            //
            $total_product_price = $product['product_price'] * $product['total_num'];   // 商品总原价
            $o_order_product_total_price = $unit_pay_price * $product['total_num'];     // 商品使用收银【优惠折扣】前总价
            // 商品应付纯单价
            $unit_pay_price = round($unit_pay_price * $discount_ratio, 2);
            $order_product_total_price = helper::bcmul($unit_pay_price, $product['total_num']);
            // 消费税
            $product_rate = isset($allProductTaxList[$product['product_id']]) ? $allProductTaxList[$product['product_id']] : 0;
            $product_consumption_tax = 0;
            $o_product_consumption_tax = 0;
            if ($consumptionTaxSetting['is_open']) {
                // 商品折扣后的消费税
                $unit_product_consumption_tax = ProductModel::getConsumptionTax($product_rate, $unit_pay_price, $consumptionTaxSetting['calc_type']); // 应付税费单价
                $product_consumption_tax = helper::bcmul($unit_product_consumption_tax, $product['total_num']); // 应付税费总价
                // 商品未折扣的消费税
                $unit_o_product_consumption_tax = ProductModel::getConsumptionTax($product_rate, $product['product_price'], $consumptionTaxSetting['calc_type']); // 原税费单价
                $o_product_consumption_tax = helper::bcmul($unit_o_product_consumption_tax, $product['total_num']); // 原税费总价
            }
            // 商品服务费
            $o_product_service_price = 0;
            $product_service_fee = 0;   // 商品服务费
            $product_service_consumption_tax = 0;   // 商品服务费的消费税
            $o_product_service_consumption_tax = 0;
            if ($serviceFee['is_open'] && $serviceFee['charge_type'] == 2) {
                // 折扣后的商品服务费
                $unit_product_service_price = ProductModel::getProductServiceFee($unit_pay_price, $service_charge_rate, $consumptionTaxSetting['calc_type'], $product_rate);  // 商品单价服务费
                $product_service_fee = helper::bcmul($unit_product_service_price, $product['total_num']); // 应付商品服务费
                // 未折扣的商品服务费
                $o_unit_product_service_price = ProductModel::getProductServiceFee($product['product_price'], $service_charge_rate, $consumptionTaxSetting['calc_type'], $product_rate);  // 商品单价服务费(未折)
                $o_product_service_price = helper::bcmul($o_unit_product_service_price, $product['total_num']); // 应付商品服务费(未折)
                // VAT 开启才计算消费税
                if ($consumptionTaxSetting['is_open'] && $serviceFee['is_open_tax']) {
                    // 折扣后商品服务费消费税
                    $unit_product_service_consumption_tax = ProductModel::getProductServiceConsumptionTax($unit_product_service_price, $product_rate);  // 商品单价服务费的消费税
                    $product_service_consumption_tax = helper::bcmul($unit_product_service_consumption_tax, $product['total_num']);  // 商品total服务费的消费税
                    // 未折扣后商品服务费消费税(未折)
                    $o_unit_product_service_consumption_tax = ProductModel::getProductServiceConsumptionTax($o_unit_product_service_price, $product_rate);  // 商品单价服务费的消费税(未折)
                    $o_product_service_consumption_tax = helper::bcmul($o_unit_product_service_consumption_tax, $product['total_num']);  // 商品total服务费的消费税(未折)
                }
            }
            // 免单处理
            if ($product['is_free']) {
                $no_free_order_product_total_price = $order_product_total_price;
                $no_free_product_service_fee = $product_service_fee;
                $no_free_product_service_consumption_tax = $product_service_consumption_tax;
                $no_free_product_consumption_tax = $product_consumption_tax;
                $product_points_bonus = 0;
                $order_product_total_price = 0;
                $product_service_fee = 0;
                $product_service_consumption_tax = 0;
                $product_consumption_tax = 0;
                $o_order_product_total_price = $total_product_price; // 免单用纯原价计算优化折扣
                $grade_total_money = 0; // 免单会员折扣置零
            } else {
                $no_free_order_product_total_price = 0;
                $no_free_product_service_fee = 0;
                $no_free_product_service_consumption_tax = 0;
                $no_free_product_consumption_tax = 0;
            }
            // 优惠金额
            $product_discount_money = helper::bcsub($o_order_product_total_price, $order_product_total_price, 2);
            $product_points_bonus = $product_points_bonus * $discount_ratio;
            // 主表order数据累加
            if ($product->is_send_kitchen == 1 && $product->is_return == 0) {
                $points_bonus += $product_points_bonus; // 积分
                $pay_money += floatval($order_product_total_price);  // 应付金额
                $order_price += $total_product_price;  // 商品原价
                $user_discount_money += $grade_total_money; // 会员商品优惠金额
                $total_product_discount_money += floatval($product_discount_money);
                // 总消费税（折后）
                $consumption_tax = helper::bcadd($consumption_tax, $product_consumption_tax);
                $consumption_tax = helper::bcadd($consumption_tax, $product_service_consumption_tax);
                // 总消费税（未折）
                $o_consumption_tax = helper::bcadd($o_consumption_tax, $o_product_service_consumption_tax);
                $o_consumption_tax = helper::bcadd($o_consumption_tax, $o_product_consumption_tax);
                // 商品消费税、商品服务费、商品服务费消费税（未折)
                $total_product_consumption_tax = helper::bcadd($total_product_consumption_tax, $product_consumption_tax);
                $total_product_service_consumption_tax = helper::bcadd($total_product_service_consumption_tax, $product_service_consumption_tax);   // 总商品服务费消费税
                $total_product_service_fee = helper::bcadd($total_product_service_fee, $product_service_fee); // 总商品服务费
                // 商品消费税、商品服务费、商品服务费消费税（折后)
                $o_total_product_service_consumption_tax = helper::bcadd($o_total_product_service_consumption_tax, $o_product_service_consumption_tax);   // 总商品服务费消费税(未折)
                $o_total_product_service_fee = helper::bcadd($o_total_product_service_fee, $o_product_service_price);   // 总商品服务费消费税(未折)
            }
            // 收银送厨和未送厨订单价格记录
            if ($product->is_return == 0 && ($product->is_send_kitchen == 1 || $product->add_source == self::CASHIER_PRODUCT_SOURCE)) {
                //
                $all_pay_money += floatval($order_product_total_price);  // 应付金额
                // 总消费税（折后）
                $all_consumption_tax = helper::bcadd($all_consumption_tax, $product_consumption_tax);
                $all_consumption_tax = helper::bcadd($all_consumption_tax, $product_service_consumption_tax);
                // 商品消费税、商品服务费、商品服务费消费税（折后)
                $all_total_product_consumption_tax = helper::bcadd($all_total_product_consumption_tax, $product_consumption_tax);
                $all_total_product_service_consumption_tax = helper::bcadd($all_total_product_service_consumption_tax, $product_service_consumption_tax);   // 总商品服务费消费税
                $all_total_product_service_fee = helper::bcadd($all_total_product_service_fee, $product_service_fee); // 总商品服务费
            }
            //
            $merge_consumption_tax = helper::bcadd($product_consumption_tax, $product_service_consumption_tax); // 商品消费税 + 商品服务费消费税
            $newData = [
                'order_product_id' => $product['order_product_id'],
                'user_id' => $order['user_id'],
                'total_price' => $order_product_total_price,                   // 商品总价(数量×单价)应付
                'total_pay_price' => $order_product_total_price,
                'total_product_price' => $total_product_price,              // 商品总价(数量×单价)原价
                'points_bonus' => $product_points_bonus,                    // 奖励积分
                'is_user_grade' => (int) $is_user_grade,                    // 是否存在会员等级折扣
                'grade_ratio' => $grade_ratio,                              // 会员折扣比例(0-10)
                'grade_product_price' => $user ? $grade_product_price : 0,  // 会员折扣后的商品单价
                'grade_total_money' => $grade_total_money,                  // 会员折扣的总额差 （商品总价 - 商品折扣后总价）
                'product_discount_money' => $product_discount_money,        // 优惠折扣后与原价总差额
                'tax_rate' => $product_rate,                                // 当前消费税率
                'consumption_tax' => $merge_consumption_tax,                    // 商品消费税 + 商品服务费消费税
                'tax_calc_type' => $consumptionTaxSetting['calc_type'],         // 含税类型 0-关闭 1-已含税 2-未含税
                'product_service_rate' => $service_charge_rate,                             // 商品服务费率
                'product_service_fee' => $product_service_fee,                              // 商品服务费
                'product_service_consumption_tax' => $product_service_consumption_tax,      // 商品服务费的消费税
                'product_consumption_tax' => $product_consumption_tax,                      // 商品消费税
                'product_original_service_fee' => $o_product_service_price,                              // 商品服务费(原价)
                'product_original_service_consumption_tax' => $o_product_service_consumption_tax,      // 商品服务费的消费税(原价)
                'product_original_consumption_tax' => $o_product_consumption_tax,                      // 商品消费税(原价)
                'no_free_total_pay_price' => $no_free_order_product_total_price,                            // 商品应付(免单前)
                'no_free_product_consumption_tax' => $no_free_product_consumption_tax,                      // 商品消费税(免单前)
                'no_free_product_service_fee' => $no_free_product_service_fee,                              // 商品服务费(免单前)
                'no_free_product_service_consumption_tax' => $no_free_product_service_consumption_tax,      // 商品服务费的消费税(免单前)
            ];
            foreach ($newData as $key => $val) {
                if ($val != $product[$key]) {
                    $productUpdateArr[] = $newData;
                    break;
                }
            }
        }

        // 构建批量更新的 SQL 语句
        if (count($productUpdateArr) > 0) {
            $orderProduct = new OrderProduct;
            $sql = 'UPDATE ' . $orderProduct->getConfig('prefix') . $orderProduct->getName() . ' SET ';
            foreach (array_keys($productUpdateArr[0] ?? []) as $key) {
                if ($key != 'order_product_id') {
                    $sql .= "{$key} = CASE order_product_id ";
                    foreach ($productUpdateArr as $item) {
                        $sql .= "WHEN {$item['order_product_id']} THEN {$item[$key]} ";
                    }
                    $sql .= 'END,';
                }
            }
            $sql = rtrim($sql, ',');
            $sql .= ' WHERE order_product_id IN (' . implode(',', array_column($productUpdateArr, 'order_product_id')) . ')';
            Db::connect($orderProduct->getConnection())->execute($sql);
        }

        //
        $total_price = round($pay_money, 2); // 订单商品纯总价（折后）
        $all_total_price = round($all_pay_money, 2); // 订单商品纯总价（折后）(送厨+未送)
        // 自助餐
        $buffetPrice = 0; // 总原价
        $buffetTotalPayPrice = 0;   // 总应付
        $buffetTotalDiscountMoney = 0;   // 总优惠
        $order_buffet_consumption_tax_money = 0; // 消费税
        $o_order_buffet_consumption_tax_money = 0; // 原消费税
        // $discount_ratio = $order['discount_ratio'] > 0 ? $order['discount_ratio'] / 100 : 1;  // 是否优惠折扣比例 - 使用上面存在的赋值 v1.1.0
        $buffetCustomerUpdateArr = [];
        if ($is_buffet) {
            foreach ($orderBuffetCustomerType as $item) {
                $o_buffet_consumption_tax = 0;                              // 商品消费税(原价)
                $o_buffet_product_service_price = 0;                        // 商品服务费(原价)
                $o_buffet_product_service_consumption_tax = 0;              // 商品服务费的消费税(原价)
                $buffet_consumption_tax = 0;
                $buffet_rate = isset($allBuffetTaxList[$item['buffet_id']]) ? $allBuffetTaxList[$item['buffet_id']] : 0;
                $buffet_price = round($item['total_price'] * $discount_ratio, 2); // 受优惠折扣影响
                //
                if ($consumptionTaxSetting['is_open']) {
                    // 折扣后 消费税
                    $unit_price = helper::bcdiv($buffet_price, $item['num'] ?: 1, 2);   // 应收单价
                    $unit_buffet_consumption_tax = Buffet::getConsumptionTax($buffet_rate, $unit_price, $consumptionTaxSetting['calc_type']); // 单价应付税费
                    $buffet_consumption_tax = helper::bcmul($unit_buffet_consumption_tax, $item['num']); // 共应付税费
                    $order_buffet_consumption_tax_money = helper::bcadd($order_buffet_consumption_tax_money, $buffet_consumption_tax);
                    $total_product_consumption_tax = helper::bcadd($total_product_consumption_tax, $buffet_consumption_tax);
                    $all_total_product_consumption_tax = helper::bcadd($all_total_product_consumption_tax, $buffet_consumption_tax);
                    // 未折扣后 消费税累加
                    $o_unit_price = helper::bcdiv($item['total_price'], $item['num'], 2);   // 原价单价
                    $unit_o_buffet_consumption_tax = Buffet::getConsumptionTax($buffet_rate, $o_unit_price, $consumptionTaxSetting['calc_type']); // 原税费单价
                    $o_buffet_consumption_tax = helper::bcmul($unit_o_buffet_consumption_tax, $item['num']); // 原税费总价
                    $o_order_buffet_consumption_tax_money = helper::bcadd($o_order_buffet_consumption_tax_money, $o_buffet_consumption_tax);
                }
                // 自助餐商品服务费
                $buffet_product_service_fee = 0;   // 商品服务费
                $buffet_product_service_consumption_tax = 0;   // 商品服务费的消费税
                if ($serviceFee['is_open'] && $serviceFee['charge_type'] == 2) {
                    $unit_price = helper::bcdiv($buffet_price, $item['num'] ?: 1, 2);   // 应收单价
                    $o_unit_price = helper::bcdiv($item['total_price'], $item['num'], 2);           // 原价单价
                    // 折扣后的商品服务费
                    $buffet_unit_product_service_price = ProductModel::getProductServiceFee($unit_price, $service_charge_rate, $consumptionTaxSetting['calc_type'], $buffet_rate);  // 商品单价服务费
                    $buffet_product_service_fee = helper::bcmul($buffet_unit_product_service_price, $item['num']); // 应付商品服务费
                    // 未折扣的商品服务费
                    $o_buffet_unit_product_service_price = ProductModel::getProductServiceFee($o_unit_price, $service_charge_rate, $consumptionTaxSetting['calc_type'], $buffet_rate);  // 商品单价服务费(未折)
                    $o_buffet_product_service_price = helper::bcmul($o_buffet_unit_product_service_price, $item['num']); // 应付商品服务费(未折)
                    // VAT 开启才计算消费税
                    if ($consumptionTaxSetting['is_open'] && $serviceFee['is_open_tax']) {
                        // 折扣后商品服务费消费税
                        $unit_buffet_product_service_consumption_tax = ProductModel::getProductServiceConsumptionTax($buffet_unit_product_service_price, $buffet_rate);  // 商品单价服务费的消费税
                        $buffet_product_service_consumption_tax = helper::bcmul($unit_buffet_product_service_consumption_tax, $item['num']);  // 商品total服务费的消费税
                        // 未折扣后商品服务费消费税(未折)
                        $o_unit_buffet_product_service_consumption_tax = ProductModel::getProductServiceConsumptionTax($o_buffet_unit_product_service_price, $buffet_rate);  // 商品单价服务费的消费税(未折)
                        $o_buffet_product_service_consumption_tax = helper::bcmul($o_unit_buffet_product_service_consumption_tax, $item['num']);  // 商品total服务费的消费税(未折)
                        // 数据累计到主表
                        $consumption_tax = helper::bcadd($consumption_tax, $buffet_product_service_consumption_tax);   // 总消费税（折后）
                        $o_consumption_tax = helper::bcadd($o_consumption_tax, $o_buffet_product_service_consumption_tax);     // 总消费税（未折）
                        $total_product_service_consumption_tax = helper::bcadd($total_product_service_consumption_tax, $buffet_product_service_consumption_tax);   // 总商品服务费消费税
                        $o_total_product_service_consumption_tax = helper::bcadd($o_total_product_service_consumption_tax, $o_buffet_product_service_consumption_tax);   // 总商品服务费消费税(未折)
                        // 送+未送累计
                        $all_consumption_tax = helper::bcadd($all_consumption_tax, $buffet_product_service_consumption_tax);
                        $all_total_product_service_consumption_tax = helper::bcadd($all_total_product_service_consumption_tax, $buffet_product_service_consumption_tax);
                    }
                    // 消费税的累计到主表
                    $total_product_service_fee = helper::bcadd($total_product_service_fee, $buffet_product_service_fee); // 总商品服务费
                    $all_total_product_service_fee = helper::bcadd($all_total_product_service_fee, $buffet_product_service_fee);
                    $o_total_product_service_fee = helper::bcadd($o_total_product_service_fee, $o_buffet_product_service_price);   // 总商品服务费消费税(未折)
                }
                $merge_consumption_tax = helper::bcadd($buffet_consumption_tax, $buffet_product_service_consumption_tax); // 商品消费税 + 商品服务费消费税

                // 自助餐累计
                $buffetDiscountMoney = helper::bcsub($item['total_price'], $buffet_price, 2);  // 折扣后差价
                $buffetTotalDiscountMoney = helper::bcadd($buffetTotalDiscountMoney, $buffetDiscountMoney);
                $buffetTotalPayPrice = helper::bcadd($buffetTotalPayPrice, $buffet_price);
                $buffetPrice = helper::bcadd($buffetPrice, $item['total_price']);
                //
                $newData = [
                    'id' => $item['id'],
                    'tax_rate' => $buffet_rate,                                                                 // 当前消费税率
                    'consumption_tax' => $merge_consumption_tax,                                                // 商品消费税 + 商品服务费消费税
                    'total_pay_price' => $buffet_price,                                                         // 应收价格（惠折扣后）
                    'tax_calc_type' => $consumptionTaxSetting['calc_type'],                                     // 含税类型 0-关闭 1-已含税 2-未含税
                    'product_service_rate' => $service_charge_rate,                                             // 商品服务费率
                    'product_service_fee' => $buffet_product_service_fee,                                       // 商品服务费
                    'product_service_consumption_tax' => $buffet_product_service_consumption_tax,               // 商品服务费的消费税
                    'product_consumption_tax' => $buffet_consumption_tax,                                       // 商品消费税
                    'product_original_service_fee' => $o_buffet_product_service_price,                          // 商品服务费(原价)
                    'product_original_service_consumption_tax' => $o_buffet_product_service_consumption_tax,    // 商品服务费的消费税(原价)
                    'product_original_consumption_tax' => $o_buffet_consumption_tax,                            // 商品消费税(原价)
                ];
                foreach ($newData as $key => $val) {
                    if ($val != $item[$key]) {
                        $buffetCustomerUpdateArr[] = $newData;
                        break;
                    }
                }
            }
        }
        $buffetPayPrice = $buffetTotalPayPrice; // 应付
        $total_product_discount_money += floatval($buffetTotalDiscountMoney);   // 累计入优惠金额

        // 构建批量更新的 SQL 语句
        if (count($buffetCustomerUpdateArr) > 0) {
            $orderBuffetCustomer = new OrderBuffetCustomer();
            $sql = 'UPDATE ' . $orderBuffetCustomer->getConfig('prefix') . $orderBuffetCustomer->getName() . ' SET ';
            foreach (array_keys($buffetCustomerUpdateArr[0] ?? []) as $key) {
                if ($key != 'id') {
                    $sql .= "{$key} = CASE id ";
                    foreach ($buffetCustomerUpdateArr as $item) {
                        $sql .= "WHEN {$item['id']} THEN {$item[$key]} ";
                    }
                    $sql .= 'END,';
                }
            }
            $sql = rtrim($sql, ',');
            $sql .= ' WHERE id IN (' . implode(',', array_column($buffetCustomerUpdateArr, 'id')) . ')';
            Db::connect($orderBuffetCustomer->getConnection())->execute($sql);
        }
        // 总消费税
        $consumption_tax = $order_buffet_consumption_tax_money + $consumption_tax;
        $consume_fee = round($consumption_tax, 2);
        //
        $all_consumption_tax = $order_buffet_consumption_tax_money + $all_consumption_tax;
        $all_consume_fee = round($all_consumption_tax, 2);
        //
        $o_consumption_tax = $o_order_buffet_consumption_tax_money + $o_consumption_tax;
        $o_consume_fee = floatval($o_consumption_tax);
        // 加钟费用
        if ($sub_order) {
            $delayPrice = Order::getSubDelayPrice($sub_order->order_id);
        } else {
            if ($order && $order['parent_id'] > 0) {
                $delayPrice = Order::getSubDelayPrice($order['order_id']);
            } else {
                $delayPrice = Order::getDelayPrice($order_id);
                $delayPrice = helper::bcmul($delayPrice, $meal_num, 3);
                $delayPrice = round($delayPrice, 2);
            }
        }

        // 应付
        // 1-已含税 2-未含税
        if ($consumptionTaxSetting['calc_type'] == 1) {
            // 含税不需要关联消费税到pay_price
            $pay_price = $total_price + $setting_service_money + $total_product_service_fee + $buffetPayPrice + $delayPrice + $total_product_service_consumption_tax; // 应付金额 = 商品折扣总价（会员折扣） + 固定服务费用 + 商品服务费 + 自助餐 + 加钟费 + 消费税(商品服务费消费税)
            $all_pay_price = $all_total_price + $setting_service_money + $all_total_product_service_fee + $buffetPayPrice + $delayPrice + $all_total_product_service_consumption_tax;
        } else {
            $pay_price = $total_price + $setting_service_money + $total_product_service_fee + $buffetPayPrice + $delayPrice + $consume_fee; // 应付金额 = 商品折扣总价（会员折扣） + 固定服务费用 + 商品服务费 + 自助餐 + 加钟费 + 消费税(商品消费税和商品服务费消费税)
            $all_pay_price = $all_total_price + $setting_service_money + $all_total_product_service_fee + $buffetPayPrice + $delayPrice + $all_consume_fee;
        }
        // 合计
        $total_price = $total_price + $buffetPayPrice + $delayPrice;
        // 原价合计
        $total_product_price = $order_price + $buffetPrice + $delayPrice;
        // 优惠折扣
        if ($order['discount_ratio'] == -1) {    // 折扣 0%
            $discount_money = $pay_price;
            $pay_price = 0;
            $all_pay_price = 0;
        } else {
            // 无折扣或折扣大于0
            $discount_money = $total_product_discount_money;
        }
        $discount_money = round($discount_money, 2);

        // 订单原价
        if ($consumptionTaxSetting['calc_type'] == 1) { // 1-已含税 2-未含税
            // 含税不需要再关联费税
            $order_price = $order_price + $setting_service_money + $o_total_product_service_fee + $buffetPrice + $delayPrice + $o_total_product_service_consumption_tax; // 订单总额 = 商品原始总价 + 桌台服务费 + 固定服务费用 + 商品服务费 + 自助餐费用 + 加钟费用 + 消费税（商品原价服务费消费税）
        } else {
            $order_price = $order_price + $setting_service_money + $o_total_product_service_fee + $buffetPrice + $delayPrice + $o_consume_fee; // 订单总额 = 商品原始总价 + 桌台服务费 + 固定服务费用 + 商品服务费 + 自助餐费用 + 加钟费用 + 消费税（商品原价消费税和服务费消费税）
        }
        //
        $small_discount_money = 0;
        $small_diff_money = 0;    // 抹零后与pay_price差值
        // 改价（直接决定pay_price）
        if ($order['discount_change_price'] > 0 || $order['discount_change_price'] == -1) { // -1 改价为0元
            $order_discount_change_price = $order['discount_change_price'] == -1 ? 0 : $order['discount_change_price'];
            $change_discount_money = $pay_price - $order_discount_change_price;    // 应付 - 改价的优惠
            $discount_money = $change_discount_money + $discount_money; // 加上之前的优惠
            $pay_price = $order_discount_change_price;
            $all_pay_price = $order_discount_change_price;
        } else {
            // 抹零
            if ($order['small_discount_type'] == 1) { //抹分
                $after_total_pay_price = floor($pay_price * 10) / 10;
                $small_discount_money = floatval(helper::bcsub($pay_price, $after_total_pay_price));
                $pay_price = $after_total_pay_price;
                $small_diff_money = $small_discount_money;
                //
                $all_pay_price = floor($all_pay_price * 10) / 10;
            } elseif ($order['small_discount_type'] == 2) { //抹角
                $after_total_pay_price = (int) $pay_price;
                $small_discount_money = floatval(helper::bcsub($pay_price, $after_total_pay_price));
                $pay_price = $after_total_pay_price;
                $small_diff_money = $small_discount_money;
                //
                $all_pay_price = (int) $all_pay_price;
            } elseif ($order['small_discount_type'] == 3) { //四舍五入到角
                $after_total_pay_price = round($pay_price, 1);
                $small_discount_money = floatval(helper::bcsub($pay_price, $after_total_pay_price));
                $pay_price = $after_total_pay_price;
                $small_diff_money = $small_discount_money;
                //
                $all_pay_price = round($all_pay_price, 1);
            } elseif ($order['small_discount_type'] == 4) { //四舍五入到元
                $after_total_pay_price = round($pay_price);
                $small_discount_money = floatval(helper::bcsub($pay_price, $after_total_pay_price));
                $pay_price = $after_total_pay_price;
                $small_diff_money = $small_discount_money;
                //
                $all_pay_price = round($all_pay_price);
            }
            $discount_money += $small_discount_money;
        }

        // 结账抹零 v1.1.0
        $checkout_discount_type = $order['checkout_discount_type'] ?? 0; // 结账抹零：1-抹分 2-抹角 5-抹元
        $checkout_diff_money = 0;    // 抹零后与pay_price差值
        $checkout_pay_price = 0.0;  // 初始化新的应付金额
        if ($checkout_discount_type == 1) {
            $checkout_pay_price = floor($pay_price * 10) / 10; // 抹去分位，保留一位小数
        } elseif ($checkout_discount_type == 2) {
            $checkout_pay_price = (int)$pay_price; // 抹去角位，保留整数部分
        } elseif ($checkout_discount_type == 5) {
            $checkout_pay_price = floor($pay_price / 10) * 10; // 抹去元位，保留10的倍数
        } else {
            $checkout_pay_price = $pay_price;
        }
        $checkout_diff_money = helper::bcsub($pay_price, $checkout_pay_price); // 抹零后与pay_price差值

        // 积分奖励按照应付计算
        if ($pointsSetting['is_shopping_gift']) {
            // 积分赠送比例
            $ratio = $pointsSetting['gift_ratio'] / 100;
        } else {
            $ratio = 0;
        }
        $points_bonus = helper::bcmul($pay_price - $checkout_diff_money, $ratio, 3);
        $points_bonus = round($points_bonus, 2);

        // 支付方式手续费
        $total_fee_money = OrderPayType::where('order_id', $order_id)->sum('fee_money');
        $pay_price = helper::bcadd($pay_price, $total_fee_money);
        if ($total_fee_money > 0) {
            $checkout_diff_money = 0; // 支付方式手续费时结账抹零金额为0（v1.1.0 - 遗漏）
        }

        // 结账抹零后的应收金额，结账完成后才更新 v1.1.0
        if ($order->getData('pay_status') == OrderPayStatusEnum::SUCCESS) {
            $pay_price = helper::bcsub($pay_price, $checkout_diff_money);
        }

        // 如果免单，pay_price设为0，is_free - 是否免单 0-否 1-免单，计入总销售额、优惠折扣 2-免单，不计入总销售额、优惠折扣 (v1.1.1) - 产品需求变更
        if ($order['is_free']) {
            $discount_money = $order['is_free'] == 1 ? helper::bcadd($discount_money, $order['free_pay_price'] ?? 0) : $discount_money;
            $pay_price = 0;
            $checkout_diff_money = 0; // 免单时结账抹零金额为0
        }

        // 记录桌台送+未送 订单价格(不包含支付手续费)
        if ($order['table_id'] > 0) {
            Cache::set($order['table_id'] . '_table_price' . $order['app_id'], $all_pay_price);
        }

        //
        $consumption_tax_type = $consumptionTaxSetting['is_open'] ? $consumptionTaxSetting['calc_type'] : 0;
        $updateOrderArr = [
            'discount_money' => $discount_money,  // 折扣优惠
            'total_price' => $total_price,
            'total_product_price' => $total_product_price,
            'order_price' => $order_price,  // 订单原价总额
            'original_price' => $order_price, // 订单原价总额
            'pay_price' => $pay_price,  // 应付
            'points_bonus' => $points_bonus,
            'service_money' => helper::bcadd($setting_service_money, $total_product_service_fee),   // 总的服务费
            'meal_num' => $meal_num,
            'setting_service_money' => $setting_service_money,
            'consumption_tax_money' => $consume_fee,
            'original_consumption_tax_money' => $o_consume_fee,
            'user_discount_money' => $user_discount_money,
            'consumption_tax_type' => $consumption_tax_type,
            'small_diff_money' => $small_diff_money,
            'checkout_diff_money' => $checkout_diff_money,
            'pay_fee_money' => $total_fee_money,
            'total_product_service_fee' => $total_product_service_fee,
            'total_product_service_consumption_tax' => $total_product_service_consumption_tax,
            'total_product_consumption_tax' => $total_product_consumption_tax,
        ];
        $order->save($updateOrderArr);
        //
        return $order;
    }

    /**
     * 商品直接加入订单
     * @param $data
     *   int order_id 订单产品ID，选填
     *   int table_id 赠菜标签ID，选填
     *   int add_source 添加来源 1-收银 2-平板 3-扫码h5 ，选填
     *   int product_id 产品ID 必填
     *   int product_sku_id 产品规格ID 必填
     *   int product_num 产品数量 必填
     *   string describe 商品规格信息（用于唯一值） 必填
     *   int delivery 就餐方式 30-打包 40-堂食
     *   array feed_uuids 加料uuid
     *   array attr_ids 属性ID组
     *   int sub_order_id 子订单ID，选填
     * @return bool 如果操作成功返回true，如果商品不存在或操作失败返回false
     * @param $user
     * @param $device_id
     * @param $product_source
     * @return false|int|mixed
     */
    public function addToOrder($data, $user, $device_id = '', $product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        // 设置
        $settingData = SettingModel::getAll($user['app_id'] ?? 0, $user['shop_supplier_id'] ?? 0);

        // 参数
        $params = $this->processAddToOrderParams($data, $settingData);
        $productId = $params['product_id'];
        $productSkuId = $params['product_sku_id'];

        // 取得订单id
        $orderId = $params['order_id'] ?: $this->getDeviceIdOrTableIdToOrderId($device_id, $params['table_id']);

        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . request()->appId . '_' . $orderId);
        $queue->while();

        /**
         * 获取并判断商品是否下架
         */
        $productDetail = ProductModel::where('product_id', '=', $productId)
            ->where('product_status', '=', 10)
            ->find();
        if (!$productDetail) {
            $this->errorData = ['product_id' => $productId];
            return $this->handleError('商品已下架', StatusCode::PRODUCT_ERROR_NOT_EXIST, $queue);
        }

        // 商品属性检查
        $validationResult = $this->validateProductAttributes($productDetail, $params['attr_ids'], $params['feed_ids']);
        if (!$validationResult['status']) {
            $this->errorData = $validationResult['data'] ?? [];
            return $this->handleError($validationResult['message'], $validationResult['code'] ?? 0, $queue);
        }

        // 各种验证
        $this->startTrans();
        try {
            // 购买人数
            $mealNum = 1;
            // 检查订单状态
            $detail = null;
            if ($orderId > 0) {
                $detail = self::detail([
                    ['order_id', '=', $orderId],
                    ['order_status', '=', OrderStatusEnum::NORMAL]
                ]);
                if (!$detail) {
                    return $this->handleError('订单不存在', 0, $queue);
                }
                // 验证订单状态是否可操作
                if ($error = $detail->validateOrderActionableStatus()) {
                    return $this->handleError($error, 0, $queue);
                }
                // 检查自助餐商品可添加状态
                if ($detail['is_buffet'] == 1 && $detail['buffet_expired_time'] != -1 && $detail['buffet_expired_time'] < time()) {
                    if (($settingData[SettingEnum::BUFFET]['values']['is_buy_continue'] ?? 0) != 1) {
                        return $this->handleError('用餐时间已到，无法添加商品', 0, $queue);
                    }
                }
                //
                $orderId = $detail['order_id'];
                $mealNum = $detail['meal_num'];
            } else {
                // 实例化订单service
                $orderService = new CashierOrderSettledService($user, [], ['eat_type' => 10]);
                // 初始化订单信息
                $orderInfo = $orderService->settlementCashier();
                if ($orderService->hasError()) {
                    return $this->handleError($orderService->getError(), 0, $queue);
                }
                // 获取消费税类型
                $consumeFee = $settingData[SettingEnum::TAX_RATE]['values'];
                $orderInfo['consumption_tax_type'] = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
                // 基本信息
                $orderInfo['device_id'] = $device_id;
                $orderInfo['delivery'] = $params['delivery'];
                // 创建订单
                $orderId = $orderService->createOrder($orderInfo);
                if (!$orderId) {
                    return $this->handleError($orderService->getError(), 0, $queue);
                }
                if (!$detail) {
                    $detail = (new Order)->where('order_id', $orderId)->find();
                }
            }

            /**
             * 判断限购
             */
            if (($purchaseLimit = $this->validatePurchaseLimit($detail['parent_id'] ?: $orderId, $product_source, $productDetail, $productSkuId, $params['product_num'], $mealNum, $params['feed_ids'])) === false) {
                return $this->handleError($this->getError(), 0, $queue);
            }
            $isBuffet = $purchaseLimit['isBuffet'];

            /**
             * 判断是否存在拆单 - 往拆单1加，按照原规格往拆单1加
             */
            if ($params['sub_order_id'] == 0) {
                $params['sub_order_id'] = $this->where('parent_id', $orderId)->order('order_id')->value('order_id') ?: 0;
            }
            if ($params['sub_order_id']) {
                if ($error = $this->where('order_id', $params['sub_order_id'])->find()?->validateOrderActionableStatus()) {
                    return $this->handleError($error, 0, $queue);
                }
            }

            // 是否存在该商品
            $query = (new OrderProduct)
                ->where('order_id', $orderId)
                ->where('product_attr', $params['describe'])
                ->where('product_sku_id', $productSkuId)
                ->where('is_free', 0)
                ->where('is_change_price', 0)
                ->where('remark', '')
                ->where('add_source', $product_source)
                ->where('is_change_price', 0)
                ->where('scheme_id', $params['scheme_id'])
                ->where('is_send_kitchen', 0)
                ->where('batch_no', '');   // 过滤下单批次的
            /** @var OrderProduct $exist_product */
            $exist_product = $query->clone()->where('sub_order_id', $params['sub_order_id'])->find();
            if ($exist_product && !$params['is_move']) {
                $exist_product->save(['total_num' => ++$exist_product->total_num]);
            } else {
                // 保存商品
                $line_price = (new ProductSku())->where('product_sku_id', $data['product_sku_id'])->value('product_price');
                $price = $isBuffet ? 0 : $line_price;
                $feed_price = $productDetail->getFeedPrice($params['feed_ids']);
                $price = helper::bcadd($price, $feed_price);
                if ($params['is_change_price']) {
                    $price = $params['product_price'];
                }
                // 是否属于必点商品
                $mustProductIds = $detail->getSchemeMustProductIds();
                //
                $inArr = [
                    'order_id' => $orderId,
                    'product_id' => $productDetail['product_id'],
                    'product_name' => $productDetail['product_name'],
                    'image_id' => isset($productDetail['logo']['image_id']) ? $productDetail['logo']['image_id'] : 0,
                    'deduct_stock_type' => $productDetail['deduct_stock_type'],
                    'spec_type' => $productDetail['spec_type'],
                    'content' => $productDetail['content'],
                    'product_sku_id' => $data['product_sku_id'] ?? 0,
                    'product_attr' => $params['describe'],
                    'product_price' => $price,
                    'line_price' => $line_price,
                    'total_num' => $params['product_num'],
                    'total_price' => $totalPrice = $params['product_num'] * $price,
                    'total_pay_price' => $totalPrice,
                    'is_buffet_product' => $isBuffet,
                    'feed_price' => $feed_price,
                    'feed_uuids' => json_encode($params['feed_uuids']),
                    'attr_ids' => json_encode($params['attr_ids']),
                    'feed_ids' => json_encode($params['feed_ids']),
                    'add_source' => $params['add_source'],
                    'kitchen_is_open' => $params['kitchen_is_open'],
                    'is_send_kitchen' => $params['is_send_kitchen'],
                    'send_kitchen_time' => $params['send_kitchen_time'],
                    'is_free' => $params['is_free'],
                    'free_remark' => $params['free_remark'],
                    'is_move' => $params['is_move'],
                    'move_from_table_id' => $params['move_from_table_id'],
                    'move_from_order_id' => $params['move_from_order_id'],
                    'remark' => $params['remark'],
                    'is_change_price' => $params['is_change_price'],
                    'is_require' => in_array($productDetail['product_id'], $mustProductIds) ? 1 : 0,
                    'scheme_id' => $params['scheme_id'],
                    'batch_time' => $product_source == self::SCAN_PRODUCT_SOURCE ? 0 : 1,   // 扫码端需要接单
                    'sub_order_id' => $params['sub_order_id'],
                ];
                $newOrderProductModel = new OrderProductModel;
                $newOrderProductModel->save($inArr);
                if ($params['free_tag_order_product_id'] > 0) {
                    OrderProductFree::where('order_product_id', $params['free_tag_order_product_id'])->update(['order_product_id' => $newOrderProductModel->order_product_id, 'order_id' => $orderId]);
                }
                OrderProductModel::where('order_product_id', $newOrderProductModel->order_product_id)->update([
                    'main_order_product_id' => $query->clone()->value('main_order_product_id') ?: $newOrderProductModel->order_product_id
                ]);
            }
            //
            (new self)->reloadPrice($orderId);
            //
            $this->commit();
            //
            if ($queue) $queue->release();
            //
            return $orderId;
        } catch (BaseException $e) {
            $this->rollback();
            return $this->handleError($e->getMessage(), 0, $queue);
        }
    }

    /**
     * @param $data
     * @param $order_id
     * @param $schemeMustProductIds     // 方案必点商品的ID
     * @param $product_source           // 商品来源
     * @return array
     */
    public function addToTableOrder($data, $order_id, $schemeMustProductIds, $product_source = self::TABLET_PRODUCT_SOURCE)
    {
        $add_source = $product_source;  // 添加来源 1-收银 2-平板 3-扫码
        $productId = intval($data['product_id'] ?? 0);
        $productSkuId = intval($data['product_sku_id'] ?? 0);
        $productNum = intval($data['product_num'] ?? 0);
        $describe = $data['describe'] ?? '';
        $feed_uuids = ($data['feed_uuids'] ?? []) ?: [];
        $feed_uuids = is_array($feed_uuids) ? $feed_uuids : [$feed_uuids];
        if (isset($data['feed_uuids'])) {
            $feed_ids = $data['feed_uuids'];
        } elseif ((isset($data['feed_ids']))) {
            $feed_ids = $data['feed_ids'];
        } else {
            $feed_ids = [];
        }
        $feed_ids = is_array($feed_ids) ? $feed_ids : [$feed_ids];
        $isBuffet = array_key_exists($productId, Order::getOrderBuffetProductArr($order_id)) ? 1 : 0;

        // 判断商品
        $productDetail = ProductModel::where('product_id', '=', $productId)
            ->where('product_status', '=', 10)
            ->find();
        if (!$productDetail) {
            $this->error = __('商品') . ' ' . ($data['product_name_text'] ?? '') . ' ' . __('已下架，请选择其他商品');
            $this->errorData = ['product_id' => $productId];
            $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST;
            return false;
        }

        // 判断规格
        $productSku = (new ProductSku())->where('product_sku_id', $productSkuId)->field('product_price')->find();
        if (!$productSku) {
            $this->error = __('规格') . ' ' . ($data['product_name_text'] ?? '') . ' ' . __('已下架，请选择其他规格');
            $this->errorData = ['product_id' => $productId, 'product_sku_id' => $productSkuId];
            $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST;
            return false;
        }

        // 保存商品
        $price = $isBuffet ? 0 : $productSku->product_price;
        $feed_price = $productDetail->getFeedPrice($feed_ids);
        $price = helper::bcadd($price, $feed_price);
        $price = floatval($price);
        $front_price = floatval($data['price']);
        $is_require = in_array($productDetail['product_id'], $schemeMustProductIds) ? 1 : 0;

        /**
         * 判断是否存在拆单 - 往拆单1加
         */
        $sub_order_id = $this->where('parent_id', $order_id)->order('order_id')->value('order_id') ?: 0;
        if ($sub_order_id) {
            if ($error = $this->where('order_id', $sub_order_id)->find()?->validateOrderActionableStatus()) {
                return $this->handleError($error, 0);
            }
        }

        //
        $inArr = [
            'order_id' => $order_id,
            'product_id' => $productDetail['product_id'],
            'product_name' => $productDetail['product_name'],
            'image_id' => isset($productDetail['logo']['image_id']) ? $productDetail['logo']['image_id'] : 0,
            'deduct_stock_type' => $productDetail['deduct_stock_type'],
            'spec_type' => $productDetail['spec_type'],
            'content' => $productDetail['content'],
            'product_sku_id' => $productSkuId,
            'product_attr' => $describe,
            'product_price' => $price,
            'line_price' => $productDetail['product_price'],
            'total_num' => $productNum,
            'total_price' => $totalPrice = $productNum * $price,
            'total_pay_price' => $totalPrice,
            'is_buffet_product' => $isBuffet,
            'feed_price' => $feed_price,
            'feed_uuids' => json_encode($feed_uuids),
            'feed_ids' => json_encode($feed_ids),
            'add_source' => $add_source,
            'is_require' => $is_require,
            'remark' => $data['remark'] ?? '',
            'sub_order_id' => $sub_order_id,
        ];
        $orderProductModel = (new OrderProductModel);
        $orderProductModel->save($inArr);
        //
        OrderProductModel::where('order_product_id', $orderProductModel->order_product_id)->update(['main_order_product_id' => $orderProductModel->order_product_id]);
        //
        if ($price != $front_price) {
            $data['tablet_product_name_text'] = ProductSkuModel::getNameById($data['product_sku_id']);
            return $data;
        }
        return [];
    }

    /**
     * 判断商品是否下架
     * @param $product_id
     * @return int
     */
    public function productState($product_id)
    {
        return (new ProductModel)->where('product_id', '=', $product_id)
            ->where('product_status', '=', 10)
            ->count();
    }

    /**
     * 判断商品库存
     * @param int $product_id
     * @param int $product_sku_id
     * @param int $order_id
     * @param string $product_source
     * @return int
     */
    public function productStockState($product_id, $product_sku_id, $order_id, $product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        //
        $deductStockType = ProductModel::where('product_id', $product_id)->value('deduct_stock_type');
        $orderProductNum = !$order_id ? 0 : OrderProductModel::where('order_id', $order_id)
            // 下单减库存
            ->when($deductStockType == DeductStockTypeEnum::CREATE, function ($q) use ($product_source) {
                $q->where('is_send_kitchen', 0)->where('add_source', $product_source);
            })
            //
            ->where('product_id', '=', $product_id)
            ->where('product_sku_id', '=', $product_sku_id)
            ->sum('total_num');

        //
        return (new ProductSkuModel)->where('product_id', '=', $product_id)
            ->where('product_sku_id', '=', $product_sku_id)
            ->where("stock_num", '>', $orderProductNum)
            ->count();
    }

    /**
     * 查询桌号订单未送厨商品
     * @param int $table_id
     * @param string $product_source
     * @return self|null
     */
    public function getUnSendKitchen($table_id, $product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        return $this->with([
            'product',
            'buffet',
            'delay',
            'unSendKitchenProduct' => function ($q) use ($product_source) {
                $q->where('add_source', $product_source)->where('batch_time', '=', 0);
            }
        ])
            ->where('table_id', '=', $table_id)
            ->where('order_status', '=', OrderStatusEnum::NORMAL)
            ->order('order_id desc')
            ->find();
    }

    /**
     * 查询桌号订单已送厨商品
     * @param int $table_id
     * @return self|null
     */
    public function getSendKitchen($table_id)
    {
        $appName = app('http')->getName();
        //
        return $this->with([
            'sendKitchenProduct' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('main_order_product_id, is_free, remark, product_price')->field('
                            *,
                            sum(total_num) as total_num,
                            sum(finish_num) as finish_num,
                            sum(total_price) as total_price,
                            sum(total_product_price) as total_product_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num,
                            sum(tax_rate) as tax_rate,
                            sum(consumption_tax) as consumption_tax
                        ');
                }
            },
            'buffetCustomerType' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('buffet_id, customer_type_id, buffet_customer_id')->field('
                            *,
                            sum(num) as num,
                            sum(total_price) as total_price,
                            sum(total_pay_price) as total_pay_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num,
                            sum(tax_rate) as tax_rate,
                            sum(consumption_tax) as consumption_tax
                        ');
                }
            },
            'delay' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('main_id, delay_id, delay_time, price, name')->field('
                            *,
                            sum(num) as num,
                            sum(total_price) as total_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num
                        ');
                }
            }
        ])
            ->where('table_id', '=', $table_id)
            ->where('order_status', '=', OrderStatusEnum::NORMAL)
            ->order('order_id desc')
            ->find();
    }

    /**
     * 查询桌号订单已送厨商品
     * @param int $table_id
     * @return self|null
     */
    public function getSendAndBatchKitchen($table_id)
    {
        $appName = app('http')->getName();
        //
        return $this->with([
            'sendAndBatchKitchenProduct' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('main_order_product_id, is_free, remark, product_price')->field('
                            *,
                            sum(total_num) as total_num,
                            sum(finish_num) as finish_num,
                            sum(total_price) as total_price,
                            sum(total_product_price) as total_product_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num,
                            sum(tax_rate) as tax_rate,
                            sum(consumption_tax) as consumption_tax
                        ');
                }
            },
            'buffetCustomerType' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('buffet_id, customer_type_id, buffet_customer_id')->field('
                            *,
                            sum(num) as num,
                            sum(total_price) as total_price,
                            sum(total_pay_price) as total_pay_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num,
                            sum(tax_rate) as tax_rate,
                            sum(consumption_tax) as consumption_tax
                        ');
                }
            },
            'delay' => function ($query) use ($appName) {
                if ($appName != 'cashier') {
                    $query->group('main_id, delay_id, delay_time, price, name')->field('
                            *,
                            sum(num) as num,
                            sum(total_price) as total_price,
                            sum(refund_money) as refund_money,
                            sum(refund_num) as refund_num
                        ');
                }
            }
        ])
            ->where('table_id', '=', $table_id)
            ->where('order_status', '=', OrderStatusEnum::NORMAL)
            ->order('order_id desc')
            ->find();
    }

    /**
     * 创建订单自助餐关联信息
     * @param int $order_id
     * @param array $buffet_ids
     * @param array $buffet_customer_type
     * @param int $shop_supplier_id
     * @return array
     */
    public static function createOrderBuffet($order_id, array $buffet_ids, array $buffet_customer_type, $shop_supplier_id = 0)
    {
        $time_limit = 0;
        $meal_num = array_sum(array_column($buffet_customer_type, 'num'));
        foreach ($buffet_ids as $buffet_id) {
            $orderBuffetTotalPrice = 0;
            $orderBuffetMealNum = 0;
            $buffet = (new Buffet)->where('status', '=', 1)->where('id', '=', $buffet_id)->find();
            if ($buffet) {
                $customer_type_num = 0;
                // 自助餐顾客类型价格
                foreach ($buffet_customer_type as $customer_item) {
                    $buffet_customer = (new BuffetCustomer)->alias('bc')
                        ->leftJoin('customer_type ct', 'ct.id = bc.customer_type_id')
                        ->where('bc.buffet_id', '=', $buffet_id)
                        ->where('ct.id', '=', $customer_item['customer_type_id'])
                        ->find();
                    if ($buffet_customer && $customer_item['num'] > 0) {
                        $buffet_customer_total_price = round(helper::bcmul($buffet_customer['price'], $customer_item['num'], 3), 2);
                        $inArr = [
                            'order_id' => $order_id,
                            'buffet_customer_id' => $buffet_customer['id'],
                            'buffet_id' => $buffet_id,
                            'customer_type_id' => $buffet_customer['customer_type_id'],
                            'buffet_name' => $buffet['name'],
                            'price' => $buffet_customer['price'],
                            'num' => $customer_item['num'],
                            'total_price' => $buffet_customer_total_price,
                            'customer_type_name' => (new CustomerType)->where('id', $buffet_customer['customer_type_id'])->value('name'),
                        ];
                        (new OrderBuffetCustomer())->save($inArr);
                        $customer_type_num++;
                        $orderBuffetTotalPrice += $buffet_customer_total_price;
                        $orderBuffetMealNum += $customer_item['num'];
                    }
                }

                if ($customer_type_num > 0) {
                    $inArr = [
                        'order_id' => $order_id,
                        'buffet_id' => $buffet_id,
                        'name' => $buffet['name'],
                        'price' => $buffet['price'],
                        'num' => $orderBuffetMealNum,
                        'total_price' => $orderBuffetTotalPrice,
                        'buy_limit_status' => $buffet['buy_limit_status'],
                        'is_comb' => $buffet['is_comb'],
                        'time_limit' => $buffet['time_limit'],
                    ];
                    if ($time_limit != -1) {
                        if ($buffet['time_limit'] == 0) {
                            $time_limit = -1;
                        } else {
                            $time_limit = max($time_limit, $buffet['time_limit']);
                        }
                    }
                    (new OrderBuffet)->save($inArr);
                }
            }
        }
        return [$time_limit, $meal_num];
    }

    /**
     * 点餐商品列表按自助餐优惠显示
     * @param array $product_list
     * @param array $buffet_arr
     * @param int $meal_num
     * @return array
     */
    public static function handleBuffetProductIndex($product_list, $buffet_arr, $meal_num)
    {
        foreach ($product_list as &$product) {
            // 已购买商品数量
            $current_add_num = $product['order_products_sum'] ?? 0;
            //
            if (array_key_exists($product['product_id'], $buffet_arr)) {
                $product['is_buffet'] = 1;
                $product['buffet_limit_num'] = $buffet_arr[$product['product_id']]['limit_num'] * $meal_num;
                $product['product_price'] = 0;
                $product['current_add_num'] = $current_add_num;
                foreach ($product['sku'] as &$item) {
                    $item['product_price'] = 0;
                }
                if ($product['buffet_limit_num'] == 0) {
                    $product['limit_num_status'] = 0;
                } else {
                    $product['limit_num_status'] = $current_add_num >= $product['buffet_limit_num'] ? 1 : 0;
                }
            } else {
                $product['is_buffet'] = 0;
                $product['buffet_limit_num'] = 0;
                $product['current_add_num'] = $current_add_num;
                if ($product['limit_num'] == 0) {
                    $product['limit_num_status'] = 0;
                } else {
                    $product['limit_num_status'] = $current_add_num >= $product['limit_num'] ? 1 : 0;
                }
            }
        }
        return $product_list;
    }

    /**
     * 商品详情按自助餐优惠显示
     * @param array $product
     * @param array $buffet_arr
     * @return array
     */
    public static function handleBuffetProductDetail($product, $buffet_arr)
    {
        if (array_key_exists($product['product_id'], $buffet_arr)) {
            $product['is_buffet'] = 1;
            $product['buffet_limit_num'] = $buffet_arr[$product['product_id']]['limit_num'];
            $product['product_price'] = 0;
            foreach ($product['sku'] as &$item) {
                $item['product_price'] = 0;
            }
        } else {
            $product['is_buffet'] = 0;
            $product['buffet_limit_num'] = 0;
        }

        return $product;
    }

    /**
     * 获取自助餐订单剩余就餐时间
     * @param int $buffet_expired_time
     * @return int
     */
    public static function getBuffetRemainingTime($buffet_expired_time)
    {
        $remaining_time = $buffet_expired_time - time();
        return max($remaining_time, 0);
    }

    /**
     * 订单加钟
     * @param int $order_id
     * @param array $delay_ids
     * @return int
     */
    public function addDelay($delay_ids)
    {
        // 禁止并发操作
        [$order, $queue] = $this->concurrencyValidateOrderActionableStatus();
        if (!$order) {
            return false;
        }

        $i = 0;
        $delay_time = 0;
        $this->startTrans();
        try {
            foreach ($delay_ids as $delay_id) {
                $delay = (new Delay)->where('status', '=', 1)->where('id', '=', $delay_id)->find();
                if ($delay) {
                    $orderId = $this->parent_id > 0 ? $this->parent_id : $this->order_id;
                    $subOrderId = $this->parent_id > 0 ? $this->order_id : 0;
                    $inArr = [
                        'order_id' => $orderId,
                        'sub_order_id' => $subOrderId,
                        'delay_id' => $delay_id,
                        'name' => $delay['name'],
                        'price' => $delay['price'],
                        'num' => $order['meal_num'],
                        'total_price' => round(helper::bcmul($delay['price'], $order['meal_num'], 3), 2),
                        'delay_time' => $delay['delay_time'],
                    ];
                    $delay = new OrderDelay();
                    $delay->save($inArr);
                    OrderDelayModel::where('id', $delay->id)->update(['main_id' => $delay->id]);
                    $delay_time = max($delay_time, $delay['delay_time']);
                    $i++;
                }
            }
            // 更新主单加钟时间
            $now_timestamp = time();
            $delay_time_second = $delay_time * 60;
            if ($order['buffet_expired_time'] >= $now_timestamp) {
                $buffet_expired_time = $order['buffet_expired_time'] + $delay_time_second;
            } else {
                $buffet_expired_time = $now_timestamp + $delay_time_second;
            }
            if ($order['parent_id'] == 0) {
                $order->save(['buffet_expired_time' => $buffet_expired_time]);
            } else {
                $subOrderList = (new Order)->where('order_id', $order['parent_id'])
                    ->whereOr('parent_id', $order['parent_id'])
                    ->select();
                foreach ($subOrderList as $subOrder) {
                    $subOrder->save(['buffet_expired_time' => $buffet_expired_time]);
                }
            }
            $this->commit();
            $queue->release();
            //
            return $i;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            $queue->release();
            return false;
        }
    }

    /**
     * 订单自助餐费用
     * @param int $order_id
     * @return int
     */
    public static function getBuffetPrice($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffet)->where($orderField, '=', $order_id)->sum('price');
    }

    /**
     * 订单自助餐顾客类型费用
     * @param int $order_id
     * @return int
     */
    public static function getBuffetCustomerPrice($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffetCustomer())->where($orderField, '=', $order_id)->sum('total_price');
    }

    /**
     * 订单自助餐顾客类型总消费税（商品+服务）
     * @param int $order_id
     * @return int
     */
    public static function getBuffetCustomerTotalConsumptionTax($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffetCustomer())->where($orderField, '=', $order_id)->sum('consumption_tax');
    }

    /**
     * 订单自助餐顾客类型商品的消费税
     * @param $order_id
     * @return float
     */
    public static function getBuffetCustomerTotalProductConsumptionTax($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffetCustomer())->where($orderField, '=', $order_id)->sum('product_consumption_tax');
    }

    /**
     * 订单自助餐顾客类型商品服务费的消费税
     * @param $order_id
     * @return float
     */
    public static function getBuffetCustomerTotalProductServiceConsumptionTax($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffetCustomer())->where($orderField, '=', $order_id)->sum('product_service_consumption_tax');
    }

    /**
     * 订单自助餐顾客类型消费税
     * @param int $order_id
     * @return int
     */
    public static function getBuffetCustomerTotalProductServiceFee($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffetCustomer())->where($orderField, '=', $order_id)->sum('product_service_fee');
    }

    /**
     * 订单自助餐数量
     * @param int $order_id
     * @return int
     */
    public static function getBuffetNum($order_id, $orderField = 'order_id')
    {
        return (new OrderBuffet)->where($orderField, '=', $order_id)->sum('num');
    }

    /**
     * 订单自助餐顾客类型数量
     * @param int $order
     * @return int
     */
    public static function getBuffetCustomerNum($order)
    {
        $query = (new OrderBuffetCustomer);
        if ($order['parent_id'] == 0) {
            return (new OrderBuffetCustomer)->where('order_id', '=', $order['order_id'])->sum('num');
        } else {
            return (new OrderBuffetCustomer)->where('sub_order_id', '=', $order['order_id'])->sum('num');
        }
    }

    /**
     * 订单加钟费用
     * @param int $order_id
     * @return int
     */
    public static function getDelayPrice($order_id, $orderField = 'order_id')
    {
        return (new OrderDelay())->where($orderField, '=', $order_id)->sum('price');
    }

    /**
     * 订单子单加钟费用
     * @param int $sub_order_id
     * @return int
     */
    public static function getSubDelayPrice($sub_order_id)
    {
        return (new OrderDelay())->where('sub_order_id', '=', $sub_order_id)->value('sum(price * num)') ?: 0;
    }

    /**
     * 订单自助餐优惠数量
     * @param int $order_id
     * @return int
     */
    public static function getBuffetDiscountNum($order_id)
    {
        return (new OrderBuffetDiscount())->where('order_id', '=', $order_id)->sum('num');
    }

    /**
     * 订单自助餐其中一种优惠数量
     * @param int $order_id
     * @param int $buffet_id
     * @return int
     */
    public static function getOrderBuffetDiscoun($order_id, $buffet_id)
    {
        return (new OrderBuffetDiscount())->where('order_id', '=', $order_id)
            ->where('buffet_id', '=', $buffet_id)
            ->sum('num');
    }

    /**
     * 订单自助餐数量
     * @param int $order_id
     * @return int
     */
    public static function getDelayNum($order)
    {
        if ($order['parent_id'] == 0) {
            return (new OrderDelay)->where('order_id', '=', $order['order_id'])->sum('num');
        } else {
            return (new OrderDelay)->where('sub_order_id', '=', $order['order_id'])->sum('num');
        }
    }

    /**
     * 更新订单自助餐人数
     * @param int $order_id
     * @param int $meal_num
     * @return void
     */
    public function updateBuffetMealNum($order_id, $meal_num)
    {
        $list = (new OrderBuffet)->where('order_id', '=', $order_id)->select();
        foreach ($list as $item) {
            $updateArr = [
                'num' => $meal_num,
                'total_price' => round(helper::bcmul($item['price'], $meal_num, 3), 2),
            ];
            $item->save($updateArr);
        }
    }

    /**
     * 更新订单加钟人数
     * @param int $order_id
     * @param int $meal_num
     * @return void
     */
    public function updateDelayMealNum($order_id, $meal_num)
    {
        $list = (new OrderDelay())->where('order_id', '=', $order_id)->select();
        foreach ($list as $item) {
            $updateArr = [
                'num' => $meal_num,
                'total_price' => round(helper::bcmul($item['price'], $meal_num, 3), 2),
            ];
            $item->save($updateArr);
        }
    }

    /**
     * 订单已送厨商品数量
     * @param int $order_id
     * @param int $product_id
     * @return int
     */
    public static function getSendKitchenNum($order_id, $product_id)
    {
        return (new OrderProduct)
            ->where('order_id', '=', $order_id)
            ->where('product_id', '=', $product_id)
            ->where('is_send_kitchen', '=', 1)
            ->sum('total_num');
    }

    /**
     * 订单未送出商品数量
     * @param int $order_id
     * @param int $product_id
     * @return int
     */
    public static function getUnSendKitchenNum($order_id, $product_id)
    {
        return (new OrderProduct)
            ->where('order_id', '=', $order_id)
            ->where('product_id', '=', $product_id)
            ->where('is_send_kitchen', '=', 0)
            ->sum('total_num');
    }

    /**
     * 添加订单自助折扣
     * @param int $buffet_id
     * @param array $buffet_discount_list
     * @return boolean
     */
    public function addOrderBuffetDiscount($buffet_id, $buffet_discount_list)
    {
        // 禁止并发操作
        [$order, $queue] = $this->concurrencyValidateOrderActionableStatus();
        if (!$order) {
            return false;
        }
        //
        $buffet = (new Buffet)->where('status', '=', 1)->where('id', '=', $buffet_id)->find();
        if (!$buffet) {
            $this->error = '自助餐不存在';
            $queue->release();
            return false;
        }
        $total_discount_num = 0;
        foreach ($buffet_discount_list as $item) {
            $total_discount_num += $item['num'];
        }
        if ($total_discount_num > $this->meal_num) {
            $this->error = '自助餐优惠数量不能大于就餐人数';
            $queue->release();
            return false;
        }

        $this->startTrans();
        try {
            foreach ($buffet_discount_list as $item) {
                $buffetDiscount = (new BuffetDiscount)->where('id', '=', $item['id'])->find();
                if (!$buffetDiscount) {
                    $this->error = '自助餐优惠不存在';
                    $queue->release();
                    return false;
                }
                if ($buffetDiscount->discount_type == 1) {  // 比例
                    $price = helper::bcmul($buffet->price, (100 - $buffetDiscount->discount_ratio) / 100);
                } else {
                    $price = $buffetDiscount->discount_price > $buffet->price ? $buffet->price : $buffetDiscount->discount_price;
                }

                $saveArr = [
                    'order_id' => $this->order_id,
                    'buffet_id' => $buffet->id,
                    'buffet_name' => $buffet->name,
                    'buffet_price' => $buffet->price,
                    'buffet_discount_id' => $buffetDiscount->id,
                    'buffet_discount_name' => $buffetDiscount->name,
                    'discount_type' => $buffetDiscount->discount_type,
                    'discount_ratio' => $buffetDiscount->discount_ratio,
                    'discount_price' => $buffetDiscount->discount_price,
                    'price' => $price,
                    'num' => $item['num'],
                    'total_price' => helper::bcmul($price, $item['num']),
                ];
                (new OrderBuffetDiscount)->save($saveArr);
                $after_total_num = (new OrderBuffetDiscount)->where('order_id', '=', $this->order_id)->where('buffet_id', '=', $buffet_id)->sum('num');
                if ($after_total_num > $this->meal_num) {
                    $this->error = '自助餐优惠数量不能大于就餐人数';
                    $queue->release();
                    return false;
                }
            }
            $this->commit();
            $queue->release();
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            $queue->release();
            return false;
        }
        return true;
    }

    /**
     * 更新自助订单折扣数目
     * @param int $order_buffet_discount_id
     * @param int $num
     * @return boolean
     */
    public function updateOrderBuffetDiscountNum($order_buffet_discount_id, $num)
    {
        if ($this->is_lock) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }
        $this->startTrans();
        try {
            $orderBuffetDiscount = (new OrderBuffetDiscount)->where('id', '=', $order_buffet_discount_id)->find();
            $updateArr = [
                'num' => $num,
                'total_price' => helper::bcmul($orderBuffetDiscount->price, $num),
            ];
            $orderBuffetDiscount->save($updateArr);
            $after_total_num = (new OrderBuffetDiscount)->where('order_id', '=', $this->order_id)->where('buffet_id', '=', $orderBuffetDiscount->buffet_id)->sum('num');

            if ($after_total_num > $this->meal_num) {
                $this->error = '自助餐优惠数量不能大于就餐人数';
                return false;
            }
            $this->commit();
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
        return true;
    }

    /**
     * 自助订餐折扣
     * @param int $order_buffet_discount_id
     * @return void
     */
    public function delOrderBuffetDiscount($order_buffet_discount_id)
    {
        if ($this->is_lock) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }

        return (new OrderBuffetDiscount)->where('id', '=', $order_buffet_discount_id)->delete();
    }

    /**
     * 更新自助餐
     * @param array $buffet_ids
     * @param array $buffet_customer_type
     * @return boolean
     */
    public function updateBuffet(array $buffet_ids, $buffet_customer_type)
    {
        // 禁止并发操作
        $result = $this->concurrencyValidateOrderActionableStatus();
        if (!$result) {
            return false;
        }
        //
        if ($this->isSplitTheOrder()) {
            $this->error = '当前订单已拆单，无法调整自助餐';
            return false;
        }
        //
        $orderUpdateArr = [];
        $now_timestamp = time();
        // 当前是否超时
        $is_time_out = $this['buffet_expired_time'] != -1 && $this['buffet_expired_time'] < $now_timestamp ? 1 : 0;
        // 当前用餐时间
        if ((new OrderBuffet)->where('order_id', $this['order_id'])->where('time_limit', 0)->find()) {
            $old_buffet_max_time_limit = 0;
        } else {
            $old_buffet_max_time_limit = (new OrderBuffet)->where('order_id', $this['order_id'])->max('time_limit');
        }
        // 是否新增
        $old_buffet_ids = (new OrderBuffet)->where('order_id', $this['order_id'])->column('buffet_id');
        $is_add = array_diff($buffet_ids, $old_buffet_ids);
        // 是否去除了原套餐
        $is_remove_old = !empty(array_diff($old_buffet_ids, $buffet_ids));

        // 只修改人数
        $re1 = array_diff($buffet_ids, $old_buffet_ids);
        $re2 = array_diff($old_buffet_ids, $buffet_ids);
        if (empty($re1) && empty($re2)) {
            $this->startTrans();
            try {
                // 删除原来的 orderBuffet、orderBuffetCustomer
                (new OrderBuffet)->where('order_id', $this['order_id'])->delete();
                (new OrderBuffetCustomer())->where('order_id', $this['order_id'])->delete();
                // 创建新的 orderBuffet、orderBuffetCustomer
                [$buffet_time_limit, $meal_num] = self::createOrderBuffet($this['order_id'], $buffet_ids, $buffet_customer_type);
                $this->updateDelayMealNum($this['order_id'], $meal_num);
                $orderUpdateArr['meal_num'] = $meal_num;
                $this->save($orderUpdateArr);
                $this->commit();
                return true;
            } catch (BaseException $e) {
                $this->error = $e->getMessage();
                $this->rollback();
                return false;
            }
        }

        // 修改了套餐
        $this->startTrans();
        try {
            // 删除原来的 orderBuffet、orderBuffetCustomer
            (new OrderBuffet)->where('order_id', $this['order_id'])->delete();
            (new OrderBuffetCustomer())->where('order_id', $this['order_id'])->delete();
            // 创建新的 orderBuffet、orderBuffetCustomer
            [$buffet_time_limit, $meal_num] = self::createOrderBuffet($this['order_id'], $buffet_ids, $buffet_customer_type);
            $new_buffet_time_limit_second = $buffet_time_limit * 60;
            // 做废已过期的加钟时间
            if ($this['buffet_expired_time'] != -1 && $this['buffet_expired_time'] < $now_timestamp) {
                (new OrderDelay)->where('order_id', $this['order_id'])->where('expired_time', 0)->update(['expired_time' => $this['buffet_expired_time']]);
            }

            if ($buffet_time_limit == -1) {
                // -1 不限时
                $buffet_expired_time = -1;
                if (!(new OrderBuffet)->where('order_id', $this['order_id'])->where('time_limit', '<>', 0)->find()) {
                    $orderUpdateArr['last_buffet_time_limit'] = 0;
                }
                if ($is_time_out && $is_add) {
                    // 无剩余时间
                    $orderUpdateArr['last_buffet_time_limit'] = 0;
                    $orderUpdateArr['buffet_start_time'] = $now_timestamp;
                }
            } else {
                if ($is_time_out && $is_add) {
                    // 无剩余时间
                    $buffet_expired_time = $now_timestamp + ($buffet_time_limit * 60) - ($old_buffet_max_time_limit * 60);
                    $orderUpdateArr['last_buffet_time_limit'] = $old_buffet_max_time_limit;
                    $orderUpdateArr['buffet_start_time'] = $now_timestamp;
                } else {
                    // 有剩余时间
                    /**
                    $buffet_remaining_time // 自助餐剩余时长（秒）
                    $delay_remaining_time  // 加钟剩余时长（秒）
                    $buffet_expend_time // 自助餐消耗时长（秒）
                    $delay_expend_time  // 加钟消耗时长（秒）
                     */
                    $order_delay_time = (new OrderDelay)->where('order_id', $this['order_id'])->where('expired_time', 0)->sum('delay_time');    // 当加钟总时长（分）
                    $order_delay_time_second = $order_delay_time * 60;
                    if ($this['buffet_remaining_time'] != 0 && $this['buffet_remaining_time'] < $order_delay_time_second && $order_delay_time_second > 0) {
                        // 只剩加钟时间
                        $delay_remaining_time = $this['buffet_remaining_time'];  // 加钟剩余时长
                        $buffet_expend_time = $old_buffet_max_time_limit * 60;
                        $delay_expend_time = $order_delay_time_second - $delay_remaining_time;
                    } else if ($this['buffet_remaining_time'] >= $order_delay_time_second && $order_delay_time_second > 0) {
                        // 剩余自助餐时间和加钟时间
                        $buffet_remaining_time = $this['buffet_remaining_time'] - $order_delay_time_second;  // 加钟剩余时长
                        $buffet_expend_time = $old_buffet_max_time_limit * 60 - $buffet_remaining_time;
                        $delay_expend_time = 0;
                    } else {
                        $buffet_remaining_time = $this['buffet_remaining_time'];
                        $buffet_expend_time = $old_buffet_max_time_limit * 60 - $buffet_remaining_time;
                        $delay_expend_time = 0;
                    }
                    // 调整类型 1 => 2
                    if ($is_add && $is_remove_old) {
                        $orderUpdateArr['last_buffet_time_limit'] = $old_buffet_max_time_limit;
                        $orderUpdateArr['buffet_start_time'] = $now_timestamp;
                    }
                    // 从不限时转到限时
                    if ($buffet_expend_time == 0) {
                        $buffet_expend_time = $now_timestamp - $this['buffet_start_time'];
                    }
                    $new_buffet_remaining_time = max($new_buffet_time_limit_second - $buffet_expend_time, 0);
                    $buffet_expired_time = $now_timestamp + $new_buffet_remaining_time + $order_delay_time_second - $delay_expend_time;
                }
            }

            // 更新订单信息
            $orderUpdateArr['buffet_expired_time'] = $buffet_expired_time;
            $orderUpdateArr['meal_num'] = $meal_num;
            $this->save($orderUpdateArr);

            // 订单商品价格变动
            if (
                !(empty(array_diff($old_buffet_ids, $buffet_ids)) && empty(array_diff($buffet_ids, $old_buffet_ids)))
            ) {
                $buffet_ids_arr = Buffet::getBuffetProductIds($buffet_ids);
                $orderProductList = (new OrderProduct)->where('order_id', $this['order_id'])->select();
                foreach ($orderProductList as $orderProduct) {
                    if ($orderProduct->is_buffet_product == 1) {
                        $this->error = '请先清除自助餐套餐内商品';
                        return false;
                    }
                    if (in_array($orderProduct->product_id, $buffet_ids_arr)) {
                        $product_price = helper::bcadd(0, $orderProduct->feed_price);
                        $updateArr = [
                            'product_price' => $product_price,
                            'total_price' => $totalPrice = $orderProduct->total_num * $product_price,
                            'total_pay_price' => $totalPrice,
                            'is_buffet_product' => 1,
                        ];
                        $orderProduct->save($updateArr);
                    }
                }
            }
            $this->updateDelayMealNum($this['order_id'], $meal_num);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 当前支付方式是否可以用 付款类型  10-余额收款 40-现金收款
     * @param int $shop_supplier_id
     * @param int $app_id
     * @param array $pay_type
     * @return boolean
     */
    public function checkPayType($shop_supplier_id, $app_id, $pay_type)
    {
        $allowPayTypeArr = [];
        $settingData = SettingModel::getAll($app_id, $shop_supplier_id);
        $payment = $settingData[SettingEnum::PAYMENT]['values'] ?? [];
        $list = PayTypeModel::getEnableListAll($shop_supplier_id, $app_id);
        //
        if ($payment['is_cash'] == 1) {
            $allowPayTypeArr[] = 40;
        }
        if ($payment['is_balance'] == 1) {
            $allowPayTypeArr[] = 10;
        }
        if ($payment['is_other'] == 1) {
            foreach ($list as $item) {
                if ($item['status'] == 1) {
                    $allowPayTypeArr[] = $item['value'];
                }
            }
        }
        //
        return in_array($pay_type, $allowPayTypeArr);
    }

    /**
     * 判断未送厨商品材料是否充足
     * @param array $product_source
     * @param array $sourceProductList
     * @return array
     */
    public function checkOrderProductIsFull($product_source = Order::CASHIER_PRODUCT_SOURCE, $sourceProductList = [])
    {
        // 订单产品列表
        $sourceProductList = $sourceProductList ?: $this->getOrderSourceProductList($product_source);
        $allProductList = $sourceProductList['allProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];
        $orderAllProductList = $sourceProductList['orderProductList'];

        // 加料消耗的材料
        $orderConsumed = [];
        foreach ($orderAllProductList as $orderProduct) {
            $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
            $productFeedIds = $productFeedIds ?: [];
            if (!empty($productFeedIds)) {
                $product = $allProductList[$orderProduct['product_id']];
                foreach ($product['feed'] as $feed_v) {
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
            $materials = $allProductSkuList[$orderProduct['product_sku_id']]['material'] ?? [];
            foreach ($materials as $material) {
                $consumedNum = isset($orderConsumed[$material['material_id']]['consumed']) ? $orderConsumed[$material['material_id']]['consumed'] : 0;
                $consumedNum += $material['material_num'] * $orderProduct['total_num'];
                $orderConsumed[$material['material_id']] = ['consumed' => $consumedNum];
            }
        }

        // 材料是否充足
        $outMaterialProductIds = $this->checkMaterialStockIsFull($orderConsumed);
        $outProductList = [];
        $outMaterialProductIdsNum = count($outMaterialProductIds);
        // 检出未送厨不足的商品
        $orderUnSendList = array_filter($orderAllProductList, function ($product) {
            return $product['is_send_kitchen'] == 0;
        });

        if ($outMaterialProductIdsNum > 0) {
            $orderUnSendList = array_reverse($orderUnSendList, true);
            foreach ($orderUnSendList as $k => $orderProduct) {
                // 是否加料材料不足
                $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
                $productFeedIds = $productFeedIds ?: [];
                if (!empty($productFeedIds)) {
                    $product = $allProductList[$orderProduct['product_id']];
                    $productSku = $allProductSkuList[$orderProduct['product_sku_id']];
                    //
                    foreach ($product['feed'] ?? [] as $feed_v) {
                        // 查找加料
                        if (in_array($feed_v['product_feed_id'], $productFeedIds) && !empty($feed_v['material'])) {
                            // 查找材料
                            foreach ($feed_v['material'] as $material) {
                                if (in_array($material['material_id'], $outMaterialProductIds)) {
                                    $feed_name = $feed_v['feed_name_text'];
                                    $product_name = $orderProduct['product']['product_name_text'] . ($productSku['spec_name_text'] ?? '') . $feed_name;
                                    $tablet_product_name = $orderProduct['product']['product_name_text'] . ' （' . ($productSku['spec_name_text'] ?? '') . '）' . $feed_name;
                                    $outProductList[] = [
                                        'product_name_text' => $product_name,
                                        'tablet_product_name_text' => $tablet_product_name
                                    ];
                                    unset($orderUnSendList[$k]);
                                    break 2;
                                }
                            }
                        }
                    }
                }
                // 是否规格材料不足
                if (!empty($orderProduct['productSku']['material'])) {
                    foreach ($orderProduct['productSku']['material'] as $material) {
                        if (in_array($material['material_id'], $outMaterialProductIds)) {
                            $product_name = $orderProduct['product']['product_name_text'] . $orderProduct['productSku']['spec_name_text'];
                            $tablet_product_name = $orderProduct['product']['product_name_text'] . ' （' . $orderProduct['productSku']['spec_name_text'] . '）';
                            $outProductList[] = [
                                'product_name_text' => $product_name,
                                'tablet_product_name_text' => $tablet_product_name
                            ];
                            unset($orderUnSendList[$k]);
                        }
                    }
                }
                if (count($outProductList) >= $outMaterialProductIdsNum) {
                    break;
                }
            }
        }
        //
        return $outProductList;
    }

    /**
     * 判断订单未送厨商品加料库存是否充足
     * @param array $product_source
     * @param array $sourceProductList
     * @return array
     */
    public function checkOrderFeedIsFull($product_source = Order::CASHIER_PRODUCT_SOURCE, $sourceProductList = [])
    {
        $sourceProductList = $sourceProductList ?: $this->getOrderSourceProductList($product_source);
        $orderProductList = $sourceProductList['orderProductList'];
        $allProductList = $sourceProductList['allProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];

        // 订单当前消耗的加料
        $orderConsumed = [];
        foreach ($orderProductList as $orderProduct) {
            $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
            $productFeedIds = $productFeedIds ?: [];
            if (!empty($productFeedIds)) {
                $product = $allProductList[$orderProduct['product_id']];
                foreach ($product['feed'] ?? [] as $feed_v) {
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

        // 检出超出库存的加料
        $outProductFeedIds = $this->checkFeedStockIsFull($orderConsumed);
        $outProductFeedIdsNum = count($outProductFeedIds);

        // 检出不足的商品
        $outProductList = [];
        if ($outProductFeedIdsNum > 0) {
            $orderProductList = array_reverse($orderProductList, true);
            foreach ($orderProductList as $k => $orderProduct) {
                $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
                $productFeedIds = $productFeedIds ?: [];
                if (!empty($productFeedIds)) {
                    $product = $allProductList[$orderProduct['product_id']];
                    foreach ($product['feed'] ?? [] as $feed_v) {
                        // 查找加料
                        if (in_array($feed_v['product_feed_id'], $outProductFeedIds) && empty($feed_v['material'])) {
                            $feed_name = $feed_v['feed_name_text'];
                            $product_spec_name_text = $allProductSkuList[$orderProduct['product_sku_id']]['spec_name_text'] ?? '';
                            $product_name = $product['product_name_text'] . $product_spec_name_text . $feed_name;
                            $tablet_product_name = $product_name;
                            $outProductList[] = [
                                'product_name_text' => $product_name,
                                'tablet_product_name_text' => $tablet_product_name
                            ];
                        }
                    }
                }

                if (count($outProductList) >= $outProductFeedIdsNum) {
                    break;
                }
            }
        }

        // 结果
        return $outProductList;
    }

    /**
     * 取消收银台上所有订单
     * @return boolean
     */
    public static function delStayOrder($deviceUUid = 0)
    {
        $builder = Order::where('desk_uuid', 0)->where('status', 0);
        if ($deviceUUid > 0) {
            $builder->where('device_uuid', $deviceUUid);
        }
        $list = $builder->select();
        foreach ($list as $item) {
            /** @var Order $item */
            $item->delStay($item->uuid);
        }
        return true;
    }

    /**
     * 取消收银桌台所有订单
     * @return boolean
     */
    public static function delStayTableOrder()
    {
        $list = Order::where('desk_uuid', '>', 0)->where('status', 0)->select();
        foreach ($list as $item) {
            /** @var Order $item */
            $item->delStay($item->uuid);
        }
        return true;
    }

    /**
     * 整单取消
     * @param $saleBillUuid
     * @param $remark
     * @return bool
     */
    public function delStay($saleBillUuid, $remark = '')
    {
        // 请求获取充值订单列表接口
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/order/cancel', json_encode([
            'sale_bill_uuid' => intval($saleBillUuid),
            'cancel_reason' => $remark ?? '',
            'not_need_password' => true,
        ]), [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $result = json_decode($res, true);
        if (($result['code'] ?? -1) != 0) {
            $this->error = $result['message'] ?? '请求失败';
            return false;
        }
        return true;
    }

    /**
     * 订单支付结果信息
     * @param int $order_id
     * @return array
     */
    public static function getPayRes($order_id)
    {
        $order = self::where('order_id', $order_id)->find();
        return [
            'order_id' => $order->order_id,
            'pay_price' => $order->pay_price,                   //  应收金额
            'pay_type' => self::getPayTypeList($order_id),      //  支付方式
            'actual_price' => $order->actual_price,             //  实收金额
            'change_due' => $order->change_due,                 //  找零
            'is_free' => $order->is_free,                       //  是否免单
            'discount_money' => $order->free_pay_price,         //  免单金额
        ];
    }

    /**
     * 获取订单自助餐的提醒信息
     * @param $order_id
     * @param $buffetIds  // 已查传入提升效率
     * @return array
     */
    public static function getBuffetRemain($order_id, $buffetIds = [])
    {
        if (empty($buffetIds)) {
            $buffetIds = OrderBuffet::where('order_id', $order_id)->column('buffet_id');
        }
        if (Buffet::whereIn('id', $buffetIds)->where('is_remain_continue', 0)->find()) {
            // 存在一条[关闭]平板/扫码H5时间的自助餐就不限制
            $is_remain_continue = 1;    // 超时平板是否可继续点餐开关 0-关闭 1-开启
            $remain_continue_notice_time = 0;   // 剩余xx分提醒不可继续点餐
            $remain_continue_time = 0;  // 剩余xx分不可继续点餐
        } else {
            $is_remain_continue = Buffet::whereIn('id', $buffetIds)->max('is_remain_continue') ?? 0;
            $remain_continue_notice_time = Buffet::whereIn('id', $buffetIds)->where('remain_continue_notice_time', '>', 0)->min('remain_continue_notice_time') ?? 0;
            $remain_continue_time = Buffet::whereIn('id', $buffetIds)->where('remain_continue_notice_time', '>', 0)->min('remain_continue_time') ?? 0;
        }

        return [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time];
    }

    /**
     * 获取（已送厨+未送厨）订单应付
     * @param $product_source
     * @return float
     */
    public function getBackPayPrice($product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        $order = $this;
        //
        $meal_num = $order['meal_num'] ?? 0; //就餐人数
        //
        $setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        //
        $consumptionTaxSetting = $setting[SettingEnum::TAX_RATE]['values'];
        $serviceFee = $setting[SettingEnum::SERVICE_CHARGE]['values'];
        //
        if ($this['discount_change_price'] > 0 || $this['discount_change_price'] == -1) {
            // 存在改价
            $order_discount_change_price = $this['discount_change_price'];
            $order_discount_change_price = $order_discount_change_price == -1 ? 0 : $order_discount_change_price;
            $total_pay_price = $order_discount_change_price;
        } else {
            // 自助餐
            $buffetPrice = 0;
            $delayPrice = 0;
            $order_buffet_consumption_tax_money = 0; // 自助餐的消费税
            if ($order && $order['is_buffet'] == 1) {
                // 自助餐顾客费用
                $buffetPrice = Order::getBuffetCustomerPrice($order['order_id']);    // 原价
                // 是否优惠折扣比例
                $discount_ratio = 1;
                if ($order['discount_ratio'] > 0) {
                    $discount_ratio = $order['discount_ratio'] / 100;
                }
                $buffetDiscountPrice = $buffetPrice * $discount_ratio;    // 折扣后
                $buffetPrice = $buffetDiscountPrice;
                // 加钟费用
                $delayPrice = Order::getDelayPrice($order['order_id']);
                $delayPrice = helper::bcmul($delayPrice, $meal_num, 3);
                $delayPrice = round($delayPrice, 2);
                if ($consumptionTaxSetting['is_open']) {
                    foreach ($order['buffetCustomerType'] as $item) {
                        $order_buffet_consumption_tax_money = helper::bcadd($order_buffet_consumption_tax_money, $item['consumption_tax'] ?: 0);
                    }
                }
            }
            //
            $order_product_pay_price = 0;    // 订单商品(已送厨)实付价钱
            $cart_product_pay_price = 0;    // 购物车商品实付价钱
            $total_consumption_tax = 0;     // 总消费税
            $total_product_service_fee = Order::getTotalProductServiceFee($order['order_id'], $product_source);                             // 总商品服务费
            $total_product_service_consumption_tax = Order::getTotalProductServiceConsumptionTax($order['order_id'], $product_source);      // 总商品服务费消费税
            foreach ($order['product'] as $product) {
                if ($product['is_return'] == 0) {
                    if (
                        $consumptionTaxSetting['is_open']
                        && ($product['is_send_kitchen'] == 1 || ($product['is_send_kitchen'] == 0 && $product['add_source'] == $product_source))
                    ) {
                        $total_consumption_tax = helper::bcadd($total_consumption_tax, $product['consumption_tax']);
                    }
                    if ($product['is_send_kitchen'] == 0 && $product['add_source'] == $product_source) {
                        $cart_product_pay_price += $product['total_price'];
                    }
                    if ($product['is_send_kitchen'] == 1) {
                        $order_product_pay_price = helper::bcadd($order_product_pay_price, $product['total_price']);
                    }
                }
            }
            // 服务费
            $total_service_money = 0;
            if ($serviceFee['is_open']) {
                if ($serviceFee['charge_type'] == 1) {
                    $total_service_money = $serviceFee['service_charge'];
                } else if ($serviceFee['charge_type'] == 2) {
                    $total_service_money = $total_product_service_fee;
                }
            }
            // 不含消费税的订单商品应付
            $orderNoTaxPayPrice = $order_product_pay_price + $buffetPrice + $delayPrice;
            //
            $total_consumption_tax += $order_buffet_consumption_tax_money;
            // 应付
            // 1-已含税 2-未含税
            if ($consumptionTaxSetting['calc_type'] == 1) {
                // 含税不需要关联消费税到pay_price
                $total_pay_price = $orderNoTaxPayPrice + $cart_product_pay_price + $total_service_money + $total_product_service_consumption_tax; // 不含消费税的订单应付 + 购物车商品总价 + 商品服务费 + 商品服务费消费税
            } else {
                $total_pay_price = $orderNoTaxPayPrice + $cart_product_pay_price + $total_service_money + $total_consumption_tax; // 不含消费税的订单应付 + 购物车商品总价 + 商品服务费 + 消费税(订单商品、自助餐套餐)
            }
        }
        // 抹零
        $total_pay_price = round($total_pay_price, 2);
        if ($order['small_discount_type'] == 1) { //抹分
            $after_total_pay_price = floor($total_pay_price * 10) / 10;
            $total_pay_price = $after_total_pay_price;
        } elseif ($order['small_discount_type'] == 2) { //抹角
            $after_total_pay_price = (int) $total_pay_price;
            $total_pay_price = $after_total_pay_price;
        } elseif ($order['small_discount_type'] == 3) { //四舍五入到角
            $after_total_pay_price = round($total_pay_price, 1);
            $total_pay_price = $after_total_pay_price;
        } elseif ($order['small_discount_type'] == 4) { //四舍五入到元
            $after_total_pay_price = round($total_pay_price);
            $total_pay_price = $after_total_pay_price;
        }
        // 支付方式手续费
        $total_fee_money = OrderPayType::where('order_id', $order['order_id'])->sum('fee_money');
        $total_pay_price = helper::bcadd($total_pay_price, $total_fee_money);
        $total_pay_price = round($total_pay_price, 2);
        // 记录桌台送+未送 订单价格
        if ($order['table_id'] > 0) {
            Cache::set($order['table_id'] . '_table_price' . $order['app_id'], $total_pay_price);
        }
        return $total_pay_price;
    }

    /**
     * 获取（已送厨+未送厨）订单原价
     * @param $product_source
     * @return float
     */
    public function getBackOrderPrice($product_source = self::CASHIER_PRODUCT_SOURCE)
    {
        $order = $this;
        //
        $meal_num = $order['meal_num'] ?? 0; //就餐人数
        //
        $setting = SettingModel::getAll($order['app_id'], $order['shop_supplier_id']);
        //
        $consumptionTaxSetting = $setting[SettingEnum::TAX_RATE]['values'];
        $serviceFee = $setting[SettingEnum::SERVICE_CHARGE]['values'];

        // 自助餐
        $buffetPrice = 0;
        $delayPrice = 0;
        $order_buffet_consumption_tax_money = 0; // 自助餐的消费税
        if ($order && $order['is_buffet'] == 1) {
            // 自助餐顾客费用
            $buffetPrice = Order::getBuffetCustomerPrice($order['order_id']);    // 原价
            // 加钟费用
            $delayPrice = Order::getDelayPrice($order['order_id']);
            $delayPrice = helper::bcmul($delayPrice, $meal_num, 3);
            $delayPrice = round($delayPrice, 2);
            if ($consumptionTaxSetting['is_open']) {
                foreach ($order['buffetCustomerType'] as $item) {
                    $unit_buffet_price = helper::bcdiv($item['total_price'], $item['num'] ?: 1);
                    $unit_consumption_tax = Buffet::getConsumptionTax($item['tax_rate'], $unit_buffet_price, $item['tax_calc_type']);
                    $total_unit_consumption_tax = helper::bcmul($unit_consumption_tax, $item['num']);
                    $order_buffet_consumption_tax_money = helper::bcadd($order_buffet_consumption_tax_money, $total_unit_consumption_tax);
                }
            }
        }
        //
        $order_product_order_price = 0;    // 订单商品(已送厨)实付价钱
        $cart_product_order_price = 0;    // 购物车商品实付价钱
        $total_consumption_tax = 0;     // 总消费税
        $total_product_original_service_fee = Order::getTotalProductOriginalServiceFee($order['order_id'], $product_source);     // 总商品服务费
        $total_product_original_service_consumption_tax = Order::getTotalProductOriginalServiceConsumptionTax($order['order_id'], $product_source);     // 总商品服务费消费税
        $total_product_original_consumption_tax = Order::getTotalProductOriginalConsumptionTax($order['order_id'], $product_source);     // 总商品消费税
        foreach ($order['product'] as $product) {
            if ($product['is_return'] == 0) {
                if (
                    $consumptionTaxSetting['is_open']
                    && ($product['is_send_kitchen'] == 1 || ($product['is_send_kitchen'] == 0 && $product['add_source'] == $product_source))
                ) {
                    $unit_consumption_tax = ProductModel::getConsumptionTax($product['tax_rate'], $product['product_price'], $product['tax_calc_type']);
                    $total_unit_consumption_tax = helper::bcmul($unit_consumption_tax, $product['total_num']);
                    $total_consumption_tax = helper::bcadd($total_consumption_tax, $total_unit_consumption_tax);
                }
                if ($product['is_send_kitchen'] == 0 && $product['add_source'] == $product_source) {
                    $cart_product_order_price += $product['total_product_price'];
                }
                if ($product['is_send_kitchen'] == 1) {
                    $order_product_order_price = helper::bcadd($order_product_order_price, $product['total_product_price']);
                }
            }
        }
        // 服务费
        $total_service_money = 0;
        if ($serviceFee['is_open']) {
            if ($serviceFee['charge_type'] == 1) {
                $total_service_money = $serviceFee['service_charge'];
            } else if ($serviceFee['charge_type'] == 2) {
                $total_service_money = $total_product_original_service_fee;
            }
        }
        // 不含消费税的订单原价
        $orderNoTaxOrderPrice = $order_product_order_price + $buffetPrice + $delayPrice;
        //
        $total_consumption_tax += $order_buffet_consumption_tax_money;
        $total_consumption_tax = round($total_consumption_tax, 2);
        // 订单原价
        // 1-已含税 2-未含税
        if ($consumptionTaxSetting['calc_type'] == 1) {
            // 含税不需要关联消费税
            $total_order_price = $orderNoTaxOrderPrice + $cart_product_order_price + $total_service_money + $total_product_original_service_consumption_tax; // 不含消费税的订单原价 + 购物车商品总价 + 商品服务费 + 商品服务费消费税
        } else {
            $total_order_price = $orderNoTaxOrderPrice + $total_product_original_consumption_tax + $cart_product_order_price + $total_service_money + $total_product_original_service_consumption_tax; // 不含商品消费税的订单原价 + 商品消费税 + 购物车商品总价 + 商品消费税 + 商品服务费消费税
        }

        return round($total_order_price, 2);
    }

    /**
     * 更改打包状态
     * @param string $delivery
     * @return boolean
     */
    public function updateDelivery($delivery)
    {
        $this->startTrans();
        try {
            // 打包状态
            $this->save(['delivery_type' => $delivery]);
            OrderProductModel::where('order_id', $this['order_id'])->save(['delivery' => $delivery]);
            $this->reloadPrice($this['order_id']);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 获取订单商品库存不足的
     * @param $productSource
     * @param $orderProductList
     * @param $allProductSkuList
     * @return array
     */
    public function getStockInsufficientProduct($productSource, $orderProductList, $allProductSkuList)
    {
        // 付款减库存-判断库存
        $productArray = [];
        foreach ($orderProductList as $orderProduct) {
            if ($orderProduct['deduct_stock_type'] != DeductStockTypeEnum::CREATE || ($orderProduct['is_send_kitchen'] == 0 && $orderProduct['add_source'] == $productSource)) {
                $key = $orderProduct['product_id'] . '_' . $orderProduct['product_sku_id'];
                $productArray[$key] = [
                    'product_id' => $orderProduct['product_id'],
                    'product_sku_id' => $orderProduct['product_sku_id'],
                    'total_num' => ($productArray[$key]['total_num'] ?? 0) + $orderProduct['total_num']
                ];
            }
        }
        // 查出库存不足的数据
        $stockInsufficient = $productArray ? (new ProductSkuModel)->where(function ($q) use ($productArray) {
            foreach ($productArray as $product) {
                $q->whereOr(function ($qq) use ($product) {
                    $qq->where('product_id', '=', $product['product_id']);
                    $qq->where('product_sku_id', '=', $product['product_sku_id']);
                    $qq->where('stock_num', '<', $product['total_num']);
                });
            }
        })->column('product_sku_id', 'product_id') : [];
        //
        $result = [];
        foreach ($orderProductList as $orderProduct) {
            $key = $orderProduct->product_id . '_' . $orderProduct->product_sku_id;
            if (!in_array($key, $result)) {
                foreach ($stockInsufficient as $product_id => $product_sku_id) {
                    if ($orderProduct['product_id'] == $product_id && $orderProduct['product_sku_id'] == $product_sku_id) {
                        $specNameText = $allProductSkuList[$product_sku_id]['spec_name_text'] ?? '';
                        $productNameText = $orderProduct->product_name_text;
                        if ($specNameText) {
                            $productNameText = $productNameText . ' （' . $specNameText . '）';
                        }
                        $result[$key] = [
                            'order_product_id' => $orderProduct['order_product_id'],
                            'product_id' => $product_id,
                            'product_sku_id' => $product_sku_id,
                            'total_num' => $orderProduct['total_num'],
                            'product_name_text' => $productNameText,
                        ];
                    }
                }
            }
        }
        //
        return array_values($result);
    }

    /**
     * 获取当前订单的百分比对象列表
     * @return array
     */
    public function getPercentageList()
    {
        $orderIds = $this->parent_id ?: $this->order_id;
        if ($this->is_merge == 1) {
            $orderIds = self::where('delete_time', 0)->where('merge_parent_id', '=', $this->order_id)->column('order_id');
            $orderIds = implode(',', $orderIds);
        }
        //
        $prefix = env('DB_PREFIX');
        //
        $where = '';
        if ($this->parent_id > 0) {
            $where = ' and rp.sub_order_id = ' . $this->order_id;
        }
        $percentageList = Db::connect($this->getConnection())->query("
            select rp.tax_rate
                ,ifnull(sum(rp.total_pay_price), 0) as total_price
                ,round(ifnull(sum(rp.consumption_tax), 0), 2) as consumption_tax
            from (
                select order_id, sub_order_id, tax_rate, total_pay_price, consumption_tax from {$prefix}order_product where product_id > 0 and is_return = 0
                UNION ALL
                select order_id, sub_order_id, tax_rate, total_pay_price, consumption_tax from {$prefix}order_buffet_customer
                UNION ALL
                select order_id, sub_order_id, 0 as tax_rate, total_price as total_pay_price, 0 as consumption_tax from {$prefix}order_delay
            ) rp
            where rp.order_id in ($orderIds) and rp.tax_rate > 0 $where
            group by rp.tax_rate
        ");
        //
        foreach ($percentageList as $key => &$data) {
            $percentageList[$key]['tax_rate'] = floatval($data['tax_rate']);
            $percentageList[$key]['consumption_tax'] = floatval($data['consumption_tax']);
            $percentageList[$key]['total_price'] = floatval($data['total_price']);
            if ($this->consumption_tax_type == 2) {
                $percentageList[$key]['total_price'] += $percentageList[$key]['consumption_tax'];
            }
        }
        //
        return $percentageList;
    }

    /**
     * 获取当前订单商品列表 - 结账单打印模版用
     * @return array
     */
    public function getPrinterProduct($order)
    {
        $products = [];
        foreach ($order->product as $product) {
            if (($product['is_buffet_product'] ?? 0) == 1 && $product['total_product_price'] <= 0) {
                continue;
            }
            if (($product['is_return'] ?? 0) == 1) {
                continue;
            }
            $key = $product['product_id'] . $product['product_sku_id'] . $product['product_attr'] . ($product['is_free'] ? 1 : 0) . strval($product['product_price']);
            $products[$key] = [
                'product_name' => $product['product_name_text'],
                'product_price' => $product['product_price'],
                'is_free' => $product['is_free'] ? 1 : 0,
                'product_attr' => $product['product_attr'] ? '(' . $product['product_attr'] . ')'  : '',
                'product_discount_money' => helper::bcadd($product['product_discount_money'], $products[$key]['product_discount_money'] ?? 0, 2),
                "total_num" => helper::bcadd($product['total_num'], $products[$key]['total_num'] ?? 0, 0),
                "total_product_price" => helper::bcadd($product['total_product_price'], $products[$key]['total_product_price'] ?? 0)
            ];
        }
        return [
            'no' => $order->call_no ?: $order->table_no ?: 0,
            'meal_num' => $order->meal_num ?: 0,
            'products' => $products,
            'buffetCustomerType' => $order->buffetCustomerType,
            'buffetDiscount' => $order->buffetDiscount,
            'delay' => $order->delay
        ];
    }

    /**
     * 获取当前订单的商品列表（包含兄弟） - 结账单打印模版用
     * @return array
     */
    public function getPrinterProductsList()
    {
        $allProducts = [];
        if ($this->is_merge == 0) {
            $allProducts[] = $this->getPrinterProduct($this);
        } else {
            $subOrders = $this->field('order_id, call_no, table_no, meal_num')->where('delete_time', 0)->where('merge_parent_id', '=', $this->order_id)->select();
            foreach ($subOrders as $order) {
                $allProducts[] = $this->getPrinterProduct($order);
            }
        }
        return $allProducts;
    }

    /**
     * 合并订单
     * @param Order $order
     * @param $table_id
     * @param array $table_ids
     * @param $user_id
     * @return bool
     */
    public function mergeOrdersByTableIds(Order $order, $table_id, array $table_ids, $user_id)
    {
        $merge_id = $order->merge_id ?: $this->generateMergeId();
        $master_order_id = $order->merge_parent_id;

        //
        if ($order->merge_id && !empty($table_ids)) {
            $cur_table_ids = Order::where('merge_id', $order->merge_id)->column('table_id');
            $add_after_table_ids = array_merge([$table_id], $table_ids);
            if (count(array_diff($cur_table_ids, $add_after_table_ids)) === 0 && count(array_diff($add_after_table_ids, $cur_table_ids)) === 0) {
                // 无变化
                return true;
            }
        }
        // 检查本桌是否部分支付
        if ($master_order_id && OrderPayType::where('order_id', $master_order_id)->find()) {
            $this->error = '当前桌台已被部分支付，不支持合单';
            return false;
        } else if (OrderPayType::where('order_id', $order->order_id)->find()) {
            $this->error = '当前桌台已被部分支付，不支持合单';
            return false;
        }

        $this->startTrans();
        try {
            //
            Order::where('merge_id', $merge_id)->update(['merge_id' => '', 'merge_parent_id' => 0]);
            //
            if ($master_order_id) {
                $existMaster = Order::where('order_id', $master_order_id)->find();
                if ($existMaster) {
                    $existMaster->force()->delete();
                }
            }
            //
            if (!empty($table_ids)) {
                $table_ids = array_merge([$table_id], $table_ids);
                foreach ($table_ids as $table_id) {
                    $order = Order::getTableUnderwayOrder($table_id);
                    if (!$order || $order->merge_id) {
                        $this->error = '桌台信息变动，请重新查看';
                        $this->rollback();
                        return false;
                    }
                    $order->save([
                        'merge_id' => $merge_id,
                        'user_id' => $user_id,
                        // 重置改价折扣抹零
                        'discount_ratio' => 0,
                        'discount_money' => 0,
                        'discount_change_price' => 0,
                        'is_change_price' => 0,
                        'small_discount_type' => 0,
                        'small_diff_money' => 0,
                    ]);
                    $order->reloadPrice($order->order_id);
                }
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
     * 生成合并订单ID
     * @return string
     */
    public function generateMergeId()
    {
        return time() . sprintf("%06d", mt_rand(0, 999999));
    }

    /**
     * 取消订单合并
     * @param $merge_id
     * @param $master_order_id
     * @return bool
     */
    public function cancelOrderMerge($merge_id, $master_order_id)
    {
        // 开启事务
        $this->startTrans();
        try {
            if ($master_order_id) {
                if (OrderPayType::where('order_id', $master_order_id)->find()) {
                    $this->error = '当前桌台已被部分支付，不支持拆单';
                    return false;
                }
                $existMaster = Order::where('order_id', $master_order_id)->find();
                if ($existMaster) {
                    $existMaster->force()->delete();
                }
            }
            $orderList = Order::where('merge_id', $merge_id)->select();
            foreach ($orderList as $subOrder) {
                $subOrder->where('merge_id', $merge_id)->update([
                    'merge_id' => '',
                    'merge_parent_id' => 0, // 解除合单关系
                    'user_id' => 0, // 会员重置
                ]);
                $subOrder->reloadPrice($subOrder->order_id);
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
     * 订单合并单的桌台列表
     * @return Order[]|array|\think\Collection
     */
    public function mergeTableList()
    {
        $merge_id = $this['merge_id'];
        $order_id = $this['order_id'];
        return (new self)->where(function ($q) use ($merge_id, $order_id) {
            $q->where('merge_id', $merge_id)->where('order_id', '<>', $order_id);
        })
            ->select();
    }

    /**
     * 合并子单生成待支付主单
     * @param $merge_id
     * @param $param  []
     * @return Order
     */
    public function generateMasterMergeOrder($merge_id, $param)
    {
        $mergeOrderList = self::where('merge_id', $merge_id)->select();
        $orderIds = [];
        $user_id = 0;
        foreach ($mergeOrderList as $order) {
            $orderIds[] = $order['order_id'];
            $user_id = $order['user_id'];
        }
        // 若存在旧的主单先删掉重新生成
        $old_master_order_id = 0;
        foreach ($mergeOrderList as $subOrder) {
            if ($subOrder->merge_parent_id) {
                $re = self::where('order_id', $subOrder->merge_parent_id)->find();
                if ($re) {
                    $old_master_order_id = $re->order_id;
                    $re->force()->delete();
                    break;
                }
            }
        }
        // 清除别端未送厨商品
        $otherUnsendOrderProduct = OrderProduct::whereIn('order_id', $orderIds)->where('is_send_kitchen', 0)->where('add_source', '<>', Order::CASHIER_PRODUCT_SOURCE)->select();
        foreach ($otherUnsendOrderProduct as $p) {
            $p->force()->delete();
        }

        //
        $order_no = $this->newOrderNo(10);
        $extra_times = 0;   // 送厨次数
        $total_price = 0;   // 商品总金额
        $total_product_price = 0; // 订单商品总价(原价)
        $order_price = 0; // 订单总金额
        $pay_price = 0; // 实际应收金额（不包含退款）
        $total_pay_fee_money = 0; // 支付方式手续费
        $consumption_tax_money = 0; // 订单消费税
        $service_money = 0; // 服务费
        $setting_service_money = 0; //
        $user_discount_money = 0; // 会员折扣金额
        $discount_money = 0; // 折扣优惠金额
        $pay_type = 0; // 支付方式
        $pay_time = 0; // 支付时间
        $points_bonus = 0; // 积分
        $user_id = isset($param['user_id']) ? $param['user_id'] : $user_id; // 会员ID
        $cashier_id = isset($param['cashier_id']) ? $param['cashier_id'] : 0; // 收银员ID
        $meal_num = 0; // 就餐人数
        $settle_device_id = isset($param['settle_device_id']) ? $param['settle_device_id'] : '';
        $device_id = isset($param['device_id']) ? $param['device_id'] : '';
        $app_id = isset($param['app_id']) ? $param['app_id'] : 0;
        $shop_supplier_id = isset($param['shop_supplier_id']) ? $param['shop_supplier_id'] : 0;
        if (isset($param['consumption_tax_type'])) {
            $consumption_tax_type = $param['consumption_tax_type']; // 消费税类型：0关闭, 1商品已含税, 2商品未含
        } else {
            $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $this['shop_supplier_id'], $this['app_id']);
            $consumption_tax_type = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
        }

        foreach ($mergeOrderList as $subOrder) {
            $extra_times = helper::bcadd($extra_times, $subOrder->extra_times);
            $total_price = helper::bcadd($total_price, $subOrder->total_price);
            $total_product_price = helper::bcadd($total_product_price, $subOrder->total_product_price);
            $order_price = helper::bcadd($order_price, $subOrder->order_price);
            $pay_price = helper::bcadd($pay_price, $subOrder->pay_price);
            $consumption_tax_money += $subOrder->consumption_tax_money;
            $service_money = helper::bcadd($service_money, $subOrder->service_money);
            $setting_service_money = helper::bcadd($setting_service_money, $subOrder->setting_service_money);
            $user_discount_money = helper::bcadd($user_discount_money, $subOrder->user_discount_money);
            $discount_money = helper::bcadd($discount_money, $subOrder->discount_money);
            $points_bonus = helper::bcadd($points_bonus, $subOrder->points_bonus);
            $meal_num = helper::bcadd($meal_num, $subOrder->meal_num);
        }
        $masterOrder = (new self);
        $masterOrder->save([
            'is_merge' => 1,
            'pay_status' => OrderPayStatusEnum::PENDING,
            'order_status' => OrderStatusEnum::NORMAL,
            'order_no' => $order_no,
            'extra_times' => $extra_times == 0 ? 1 : $extra_times,
            'total_price' => $total_price,
            'total_product_price' => $total_product_price,
            'order_price' => $order_price,
            'pay_price' => $pay_price,
            'consumption_tax_type' => $consumption_tax_type,
            'consumption_tax_money' => $consumption_tax_money,
            'service_money' => $service_money,
            'setting_service_money' => $setting_service_money,
            'user_discount_money' => $user_discount_money,
            'discount_money' => $discount_money,
            'pay_type' => $pay_type,
            'pay_time' => $pay_time,
            'points_bonus' => $points_bonus,
            'user_id' => $user_id,
            'cashier_id' => $cashier_id,
            'meal_num' => $meal_num,
            'settle_device_id' => $settle_device_id,
            'device_id' => $device_id,
            'buyer_remark' => '',
            'order_type' => 1,  // 用餐方式 0-外卖 1-店内
            'delivery_type' => 40,  //30-打包 40-堂食
            'eat_type' => 10, // 10-堂食 20-快餐
        ]);
        // 子单重新关联主单
        foreach ($mergeOrderList as $subOrder) {
            $subOrder->save(['merge_parent_id' => $masterOrder->order_id]);
        }
        // 已支付方式重新关联主单
        if ($old_master_order_id > 0) {
            OrderPayType::where('order_id', $old_master_order_id)->update(['order_id' => $masterOrder->order_id]);
            $total_pay_fee_money = OrderPayType::where('order_id', $masterOrder->order_id)->sum('fee_money');
            // 支付手续费
            if ($total_pay_fee_money > 0) {
                $pay_price = helper::bcadd($pay_price, $total_pay_fee_money);
                $masterOrder->save(['pay_fee_money' => $total_pay_fee_money, 'pay_price' => $pay_price]);
            }
        }

        return $masterOrder;
    }

    /**
     * 重新统计主单数据
     * @param $master_order_id
     * @param $merge_id
     * @param $param
     * @return array|mixed
     */
    public function reloadMasterMergeOrder($master_order_id, $merge_id, $param = [])
    {
        $updateData = [];
        //
        $masterOrder = self::where('order_id', $master_order_id)->find();
        $mergeOrderList = self::where('merge_id', $merge_id)->select();
        //
        $extra_times = 0;   // 送厨次数
        $total_price = 0;   // 商品总金额
        $total_product_price = 0; // 订单商品总价(原价)
        $order_price = 0; // 订单总金额
        $pay_price = 0; // 实际应收金额（不包含退款）
        $consumption_tax_money = 0; // 订单消费税
        $service_money = 0; // 服务费
        $setting_service_money = 0; // 现在的服务费
        $user_discount_money = 0; // 会员折扣金额
        $discount_money = 0; // 折扣优惠金额
        $actual_price = 0; // 客户实付金额
        $change_due = 0; // 找零
        $pay_type = 0; // 支付方式
        $points_bonus = 0; // 积分
        $meal_num = 0; // 就餐人数
        $free_pay_price = 0; // 免单前应付
        // 会员ID
        if (isset($param['user_id'])) {
            $updateData['user_id'] = $param['user_id'];
        }
        // 收银员ID
        if (isset($param['cashier_id'])) {
            $updateData['user_id'] = $param['cashier_id'];
        }
        if (isset($param['consumption_tax_type'])) {
            $updateData['consumption_tax_type'] = $param['consumption_tax_type']; // 消费税类型：0关闭, 1商品已含税, 2商品未含税
        }

        foreach ($mergeOrderList as $subOrder) {
            $extra_times = helper::bcadd($extra_times, $subOrder->extra_times);
            $total_price = helper::bcadd($total_price, $subOrder->total_price);
            $total_product_price = helper::bcadd($total_product_price, $subOrder->total_product_price);
            $order_price = helper::bcadd($order_price, $subOrder->order_price);
            $pay_price = helper::bcadd($pay_price, $subOrder->pay_price);
            $consumption_tax_money += $subOrder->consumption_tax_money;
            $service_money = helper::bcadd($service_money, $subOrder->service_money);
            $setting_service_money = helper::bcadd($setting_service_money, $subOrder->setting_service_money);
            $user_discount_money = helper::bcadd($user_discount_money, $subOrder->user_discount_money);
            $discount_money = helper::bcadd($discount_money, $subOrder->discount_money);
            $actual_price = helper::bcadd($actual_price, $subOrder->actual_price);
            $points_bonus = helper::bcadd($points_bonus, $subOrder->points_bonus);
            $meal_num = helper::bcadd($meal_num, $subOrder->meal_num);
            $free_pay_price = helper::bcadd($free_pay_price, $subOrder->free_pay_price);
        }
        // 支付方式手续费用
        $total_fee_money = OrderPayType::where('order_id', $master_order_id)->sum('fee_money');
        $pay_price = helper::bcadd($pay_price, $total_fee_money);
        //
        $updateData['is_merge'] = 1;
        $updateData['extra_times'] = $extra_times == 0 ? 1 : $extra_times;
        $updateData['total_price'] = $total_price;
        $updateData['total_product_price'] = $total_product_price;
        $updateData['order_price'] = $order_price;
        $updateData['pay_price'] = $pay_price;
        $updateData['pay_fee_money'] = $total_fee_money;
        $updateData['consumption_tax_money'] = $consumption_tax_money;
        $updateData['service_money'] = $service_money;
        $updateData['setting_service_money'] = $setting_service_money;
        $updateData['user_discount_money'] = $user_discount_money;
        $updateData['discount_money'] = $discount_money;
        $updateData['actual_price'] = $actual_price;
        $updateData['change_due'] = $change_due;
        $updateData['pay_type'] = $pay_type;
        $updateData['points_bonus'] = $points_bonus;
        $updateData['meal_num'] = $meal_num;
        $updateData['free_pay_price'] = $free_pay_price;
        $updateData['buyer_remark'] = '';
        $masterOrder->save($updateData);

        return $masterOrder;
    }

    /**
     * 主单支付完成
     * @param $master_order_id
     * @param $param
     * @return array|mixed
     */
    public function updateMasterMergeOrderPayComplete($master_order_id, $param)
    {
        $masterOrder = self::where('order_id', $master_order_id)->find();
        //
        $actual_price = $param['actual_price'] ?? 0; // 客户实付金额
        $change_due = $param['change_due'] ?? 0; // 找零
        $user_id = isset($param['user_id']) ? $param['user_id'] : 0; // 会员ID
        $pay_time = time(); // 支付时间
        $cashier_id = isset($param['cashier_id']) ? $param['cashier_id'] : 0; // 收银员ID
        $settle_device_id = isset($param['settle_device_id']) ? $param['settle_device_id'] : 0;
        $table_id = isset($param['table_id']) ? $param['table_id'] : 0;
        $is_free = isset($param['is_free']) ? $param['is_free'] : 0;
        $free_remark = isset($param['free_remark']) ? $param['free_remark'] : '';

        if (isset($param['consumption_tax_type'])) {
            $consumption_tax_type = $param['consumption_tax_type']; // 消费税类型：0关闭, 1商品已含税, 2商品未含税
        } else {
            $consumeFee = SettingModel::getSupplierItem(SettingEnum::TAX_RATE, $this['shop_supplier_id'], $this['app_id']);
            $consumption_tax_type = (int) ($consumeFee['is_open'] == 0 ? 0 : $consumeFee['calc_type']);
        }

        $masterOrder->save([
            'pay_status' => OrderPayStatusEnum::SUCCESS,
            'order_status' => OrderStatusEnum::COMPLETED,
            'consumption_tax_type' => $consumption_tax_type,
            'actual_price' => $actual_price,
            'change_due' => $change_due,
            'pay_time' => $pay_time,
            'user_id' => $user_id,
            'table_id' => $table_id,
            'is_settled' => 1,
            'cashier_id' => $cashier_id,
            'settle_device_id' => $settle_device_id,
            'buyer_remark' => '',
            'is_free' => $is_free,
            'free_remark' => $free_remark,
        ]);

        return $masterOrder;
    }

    /**
     * 获取订单支付方式
     * @param $order_id
     * @return OrderPayType[]|array|\think\Collection
     */
    public static function getPayTypeList($order_id)
    {
        return (new OrderPayType)->field(['id', 'value', 'price', 'disabled_cancel', 'fee_money'])->where('order_id', $order_id)->select();
    }

    /**
     * 添加支付方式
     * @param $value
     * @param $price
     * @return bool
     */
    public function addPayType($value, $price, $payment_order_id = 0)
    {
        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queue->while();

        // 判断在线支付订单
        $paymentOrder = PaymentOrder::where('id', $payment_order_id)->find();
        if (in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) && !$paymentOrder) {
            $this->error = '支付订单不存在';
            $queue->release();
            return false;
        }

        // 重命名
        $order = $this;

        //
        $orderPayType = (new OrderPayType);
        if ($payment_order_id > 0) {
            // 在线支付订单已支付
            $orderPayTypeRecord = $orderPayType->where('payment_order_id', $payment_order_id)->find();
            if ($orderPayTypeRecord && $orderPayTypeRecord->pay_status == 1) {
                $queue->release();
                return true;
            }
        }

        /** @var PayType $payType */
        $payType = PayType::where('value', $value)->where('status', 1)->find();
        if (!(new Order)->checkPayType($order['shop_supplier_id'], $order['app_id'], $value)) {
            $this->error = '支付方式未开启';
            $queue->release();
            return false;
        }
        $pay_type_fee_money = $payType ? $payType->calFeeMoney($price) : 0;
        $fee = $payType ? $payType->fee : 0;
        // 判断会员余额支付
        if ($value == 10) {
            if ($order['user_id'] == 0) {
                $this->error = '使用会员才能进行余额支付';
                return false;
            }
            $user = UserModel::where('uuid', $order['user_id'])->find();
            $user_balance = $user->balance;
            $user_gift_balance_balance = $user->gift_balance;
            if ($price > $user_balance + $user_gift_balance_balance) {
                $this->error = '会员余额不足，请先充值';
                $queue->release();
                return false;
            }
        }
        // 判断金额
        if ($price < 0.01) {
            $this->error = '支付金额最低0.01';
            $queue->release();
            return false;
        }
        //
        $amount_paid = $orderPayType->where('order_id', $order['order_id'])->sum('price');
        if ($amount_paid >= $order['pay_price']) {
            $this->error = '当前已足额';
            $queue->release();
            return false;
        }
        // 非现金不能超订单应付金额
        $amount_paid = helper::bcadd($amount_paid, $price);
        if ($value != 40 && $amount_paid > $order['pay_price']) {
            $this->error = '非现金支付不能大于应收';
            $queue->release();
            return false;
        }

        // 最终支付金额 = 支付 + 手续费
        $price = helper::bcadd($price, $pay_type_fee_money);
        //
        $exist = $orderPayType->where('order_id', $order['order_id'])->where('value', $value)->find();
        //
        $this->startTrans();
        try {
            // 存在更新否则添加
            if ($exist) {
                $orderPayTypeId = $exist->id;
                $exist->where('order_id', $order['order_id'])
                    ->where('value', $value)
                    ->update(['price' => $price, 'fee_money' => $pay_type_fee_money]);

                // 更新主订单支付方式
                if ($order['parent_id'] > 0) {
                    $parentExists = (new OrderPayType())->where('order_id', $order['parent_id'])->where('sub_id', $exist->id)->where('value', $value)->find();
                    if ($parentExists) {
                        $parentExists->price = $price;
                        $parentExists->fee_money = $pay_type_fee_money;
                        $parentExists->save();
                        (new Order())->reloadPrice($order['parent_id']);
                    }
                }
            } else {
                $saveArr = [
                    'order_id' => $order['order_id'],
                    'payment_order_id' => $payment_order_id,
                    'value' => $value,
                    'price' => $price,
                    'fee' => $fee,
                    'fee_money' => $pay_type_fee_money,
                    'pay_status' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 0 : 1, //  手动添加在线支付标为0
                    'disabled_cancel' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 1 : 0, //  在线支付不可撤销
                    'pay_hash' => OrderPayType::generateUniqueHash($order['order_id'] . '_0', $value),
                ];
                $orderPayType->save($saveArr);
                $orderPayTypeId = $orderPayType->id;

                // 添加主订单支付方式
                if ($order['parent_id'] > 0) {
                    $parentSaveArr = [
                        'order_id' => $order['parent_id'],
                        'sub_id' => $orderPayType->id,
                        'payment_order_id' => $payment_order_id,
                        'value' => $value,
                        'price' => $price,
                        'fee' => $fee,
                        'fee_money' => $pay_type_fee_money,
                        'pay_status' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 0 : 1, //  手动添加在线支付标为0
                        'disabled_cancel' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 1 : 0, //  在线支付不可撤销
                        'pay_hash' => OrderPayType::generateUniqueHash($order['parent_id'] . '_' . $orderPayType->id, $value),
                    ];
                    (new OrderPayType())->save($parentSaveArr);
                    (new Order())->reloadPrice($order['parent_id']);
                }
            }
            if ($pay_type_fee_money > 0) {
                if ($order->is_merge) {
                    $subOrder = Order::where('merge_parent_id', $order->order_id)->find();
                    $order->reloadMasterMergeOrder($order->order_id, $subOrder->merge_id);
                } else {
                    $order->reloadPrice($order->order_id);
                }
            }
            $this->commit();
            //
            $queue->release();
            //
            return $orderPayTypeId;
        } catch (BaseException $e) {
            $queue->release();
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 添加在线支付方式
     * @param PaymentOrder $paymentOrderModel
     * @return bool
     */
    public function addOnlinePayType(PaymentOrder $paymentOrderModel, $merge_parent_id = 0)
    {
        if ($paymentOrderModel->order_id > 0) {
            $orderPayType = (new OrderPayType);
            //
            $existOrderPayType = $orderPayType->where('payment_order_id', $paymentOrderModel->id)->find();
            if ($existOrderPayType) {
                // 手动添加的变为已支付
                if ($existOrderPayType->pay_status == 1) {
                    return true;
                } else {
                    return $existOrderPayType->save(['pay_status' => 1]);
                }
            } else {
                $order_id = $merge_parent_id ?: $paymentOrderModel->order_id;
                $inArr = [
                    'order_id' => $merge_parent_id ?: $paymentOrderModel->order_id,
                    'payment_order_id' => $paymentOrderModel->id,
                    'value' => $paymentOrderModel->pay_type_value,
                    'price' => $paymentOrderModel->order_amount,
                    'fee' => $paymentOrderModel->pay_type_fee,
                    'fee_money' => $paymentOrderModel->pay_type_fee_money,
                    'pay_status' => 1,
                    'disabled_cancel' => 1,
                    'pay_hash' => OrderPayType::generateUniqueHash($order_id, $paymentOrderModel->pay_type_value),
                ];
                //
                return $orderPayType->save($inArr);
            }
        } else if ($paymentOrderModel->recharge_order_id > 0) {
            $orderPayType = (new UserRechargeOrderPayType());
            //
            $existOrderPayType = $orderPayType->where('payment_order_id', $paymentOrderModel->id)->find();
            if ($existOrderPayType) {
                // 手动添加的变为已支付
                if ($existOrderPayType->pay_status == 1) {
                    return true;
                } else {
                    return $existOrderPayType->save(['pay_status' => 1]);
                }
            } else {
                $inArr = [
                    'order_id' => $paymentOrderModel->recharge_order_id,
                    'payment_order_id' => $paymentOrderModel->id,
                    'value' => $paymentOrderModel->pay_type_value,
                    'price' => $paymentOrderModel->order_amount,
                    'fee' => $paymentOrderModel->pay_type_fee,
                    'fee_money' => $paymentOrderModel->pay_type_fee_money,
                    'pay_status' => 1,
                    'disabled_cancel' => 1,
                ];
                //
                return $orderPayType->save($inArr);
            }
        } else {
            return false;
        }
    }

    /**
     * 撤销支付方式
     * @param $id
     * @return false
     */
    public function cancelPayType($id)
    {
        $opt = (new OrderPayType)->where('id', $id)->find();
        if (!$opt) {
            $this->error = '支付方式不存在';
            return false;
        }
        if ($opt->disabled_cancel == 1) {
            $this->error = '支付方式不可撤销';
            return false;
        }
        $this->startTrans();
        try {
            $opt->delete();
            if ($opt->fee_money > 0) {
                if ($this->merge_parent_id > 0) {
                    $this->reloadMasterMergeOrder($this->merge_parent_id, $this->merge_id);
                } else {
                    $this->reloadPrice($this->order_id);
                }
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
     * 预结账送厨
     * @param $merge_id
     * @param $cashier
     * @return array|false
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public function preSendKitchen($merge_id, $cashier)
    {
        $mergeOrderList = self::where('merge_id', $merge_id)->select();
        $pay_price = 0;
        $order_price = 0;

        $this->startTrans();
        try {
            foreach ($mergeOrderList as $order) {
                $model = new OrderProduct();
                $unSendNum = $model::where('order_id', $order->order_id)->where('is_send_kitchen', 0)->count();
                if ($unSendNum > 0) {
                    // 清除别端未送厨商品
                    $otherOrderProductModel = (new OrderProduct);
                    $otherOrderProduct = $otherOrderProductModel->where('order_id', $order->order_id)->where('add_source', '<>', Order::CASHIER_PRODUCT_SOURCE)->select();
                    $ids = $otherOrderProduct->column('order_product_id');
                    $otherOrderProductModel->destroy($ids, true);
                    //
                    if (!$model->sendKitchen($order->order_id, 'kitchen', true, $order->delivery_type['value'])) {
                        $this->error = $model->getError() ?: '送厨失败';
                        $this->errorData = $model->getErrorData();
                        $this->errorCode = $model->getErrorCode();
                        return false;
                    }
                }
                $pay_price = helper::bcadd($pay_price, $order->getBackPayPrice());
                $order_price = helper::bcadd($order_price, $order->getBackOrderPrice());
            }
            // 生成主单
            $param = [
                'settle_device_id' => $cashier['device_id'],
                'device_id' => $cashier['device_id'],
            ];
            $masterOrder = $this->generateMasterMergeOrder($merge_id, $param);
            $pay_order_id = $masterOrder->order_id;
            $this->commit();
            return [
                'pay_order_id' => $pay_order_id,
                'pay_price' => $pay_price,
                'order_price' => $order_price,
            ];
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 订单终端未送厨商品条数
     * @param $order_id
     * @param $product_source
     * @return int
     */
    public function orderProductUnSendCount($order_id, $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        return OrderProduct::where('order_id', $order_id)->where('is_send_kitchen', 0)->where('add_source', $product_source)->count();
    }

    /**
     * 获取订单支付方式总额
     * @param $order_id
     * @return float
     */
    public function getPayTypeTotalPrice($order_id)
    {
        return OrderPayType::where('order_id', $order_id)->sum('price');
    }

    /**
     * 获取订单非现金支付方式总额
     * @param $order_id
     * @return float
     */
    public function getNonCashPayTypeTotalPrice($order_id)
    {
        return OrderPayType::where('order_id', $order_id)->where('value', '<>', PayType::CASH_VALUE)->sum('price');
    }

    /**
     * 结账检查组合支付
     * @return bool
     */
    public function checkPayTypeList()
    {
        $balancePay = OrderPayType::where('order_id', $this['order_id'])->where('value', OrderPayTypeEnum::BALANCE)->find();
        // 验证余额支付时用户余额是否满足
        if ($balancePay) {
            if ($this['user_id'] == 0) {
                $this->error = '请先选择会员登录';
                return false;
            }
            if ($this->user['balance'] + $this->user['gift_balance'] < $balancePay['price']) {
                $this->error = '会员余额不足，请先充值';
                return false;
            }
        }
        return true;
    }

    /**
     * 结账处理组合支付
     * @return void
     */
    public function handlePayTypeList()
    {
        $balancePay = OrderPayType::where('order_id', $this['order_id'])->where('value', OrderPayTypeEnum::BALANCE)->find();
        if ($balancePay) {
            // 累积用户总消费金额
            $this['uuid'] && $this->user->setIncPayMoney($balancePay['price']);
            // 更新用户余额
            BalanceLog::add(BalanceLogSceneEnum::CONSUME, [
                'order_id' => $this['order_id'],
                'member_uuid' => $this->user['uuid'],
                'money' => -$balancePay['price'],
            ], ['order_no' => $this['order_no']]);
        }
    }

    /**
     * 订单商品列表
     * @return OrderProduct[]|array|\think\Collection
     */
    public function getOrderProductList()
    {
        $order = $this;
        // 是否合单主单
        if ($order['is_merge'] == 1) {
            $order_ids = self::where('merge_parent_id', $order['order_id'])->column('order_id');
            $productList = OrderProduct::whereIn('order_id', $order_ids)->where('total_price', '<>', 0)->where('is_return', 0)->select();
            $buffetList = OrderBuffetCustomer::whereIn('order_id', $order_ids)->where('total_pay_price', '<>', 0)->select();
            $delayList = OrderDelay::whereIn('order_id', $order_ids)->where('total_price', '<>', 0)->select();
        } elseif ($order['parent_id'] > 0) {
            $productList = OrderProduct::whereIn('sub_order_id', $order['order_id'])->where('total_price', '<>', 0)->where('is_return', 0)->select();
            $buffetList = OrderBuffetCustomer::whereIn('sub_order_id', $order['order_id'])->where('total_pay_price', '<>', 0)->select();
            $delayList = OrderDelay::whereIn('sub_order_id', $order['order_id'])->where('total_price', '<>', 0)->select();
        } else {
            $productList = OrderProduct::where('order_id', $order['order_id'])->where('total_price', '<>', 0)->where('is_return', 0)->select();
            $buffetList = OrderBuffetCustomer::where('order_id', $order['order_id'])->where('total_pay_price', '<>', 0)->select();
            $delayList = OrderDelay::where('order_id', $order['order_id'])->where('total_price', '<>', 0)->select();
        }
        return compact('productList', 'buffetList', 'delayList');
    }

    /**
     * 订单退款 1.0.5
     * 1.实际退款不能大于应付（改价抹零可能会大于）
     * 2.系统退款：余额支付时显示系统退款和线下退款按钮（线下退款只做记录，余额/积分不做变动）
     * 3.线下退款：组合支付及非余额支付只能线下退款（只有此按钮）
     * 4.部分退款后再次退款只能线下退款（只有此按钮）
     * @param $params
     * @return bool
     */
    public function orderRefund($params)
    {
        $order = $this->toArray();
        $refund_method = 3; // 1.0.9  3-原路返还
        $refund_type = $params['refund_type'] ?: 0;
        $refund_product = isset($params['refund_product']) ? $params['refund_product'] : [];
        $refund_buffet = isset($params['refund_buffet']) ? $params['refund_buffet'] : [];
        $refund_delay = isset($params['refund_delay']) ? $params['refund_delay'] : [];
        $cashier_id = isset($params['cashier_id']) ? $params['cashier_id'] : 0;
        $bank_code = isset($params['bank_code']) ? $params['bank_code'] : 0;
        $account_no = isset($params['account_no']) ? $params['account_no'] : '';
        $account_name = isset($params['account_name']) ? $params['account_name'] : '';

        // 判断订单是否有效
        if ($this['pay_status']['value'] != 20) {
            $this->error = '该订单不合法';
            return false;
        }
        if ($this['is_free']) {
            $this->error = '免单订单不能退款';
            return false;
        }
        //
        $reloadOrderIds = [];                   // 部分退款id
        $total_refund_money = 0;                // 累计退款金额
        $total_refund_consumption_tax = 0;      // 累计退消费税
        $orderRefundLog = null;
        $this->startTrans();
        try {
            /**
             * 退款类型[refund_type] 1-整单退款 2-部分退款
             */
            if ($refund_type == 1) {
                // 1.整单退款可用金额
                $available_pay_price = helper::bcsub($order['pay_price'], $order['refund_money']);
                $total_refund_money = $available_pay_price;
                if ($total_refund_money <= 0) {
                    $this->error = '无法退款';
                    return false;
                }
                $saveData = [
                    'order_id' => $order['order_id'],
                    'refund_type' => $refund_type,
                    'refund_method' => $refund_method,
                    'refund_money' => $total_refund_money,
                    'cashier_id' => $cashier_id,
                ];
                $orderRefundLog = (new OrderRefund)->create($saveData);

                // 2.处理所有订单商品为退款状态
                if ($order['is_merge']) {
                    $orderIds = Order::where('merge_parent_id', $order['order_id'])->column('order_id');
                    // 商品(送厨和未退菜的)
                    $orderProductList = OrderProduct::whereIn('order_id', $orderIds)->where('is_send_kitchen', 1)->where('is_return', 0)->select();
                    foreach ($orderProductList as $item) {
                        $refund_money = $item->total_consumption_tax_pay_price;
                        $refund_consumption_tax = $item->consumption_tax;
                        $item->save([
                            'refund_money' => $refund_money,
                            'refund_num' => $item->total_num,
                            'refund_consumption_tax' => $refund_consumption_tax,
                        ]);
                    }
                    // 自助餐
                    $orderBuffetCustomerList = OrderBuffetCustomer::whereIn('order_id', $orderIds)->select();
                    foreach ($orderBuffetCustomerList as $item) {
                        $refund_money = $item->total_consumption_tax_pay_price;
                        $refund_consumption_tax = $item->consumption_tax;
                        $item->save([
                            'refund_money' => $refund_money,
                            'refund_num' => $item->num,
                            'refund_consumption_tax' => $refund_consumption_tax,
                        ]);
                    }
                    // 加钟
                    $orderDelayList = OrderDelay::whereIn('order_id', $orderIds)->select();
                    foreach ($orderDelayList as $item) {
                        $item->save(['refund_money' => $item->total_price, 'refund_num' => $item->num]);
                    }
                } else {
                    // 商品(送厨和未退菜的)
                    $orderId = $order['order_id'] ?? 0;
                    $orderField = $order['parent_id'] > 0 ? 'sub_order_id' : 'order_id'; // 订单产品表记录为主单order_id，子单sub_order_id
                    $orderProductList = OrderProduct::where($orderField, $orderId)->where('is_send_kitchen', 1)->where('is_return', 0)->select();
                    foreach ($orderProductList as $item) {
                        $refund_money = $item->total_consumption_tax_pay_price;
                        $refund_consumption_tax = $item->consumption_tax;
                        $item->save([
                            'refund_money' => $refund_money,
                            'refund_num' => $item->total_num,
                            'refund_consumption_tax' => $refund_consumption_tax,
                        ]);
                    }
                    // 自助餐
                    $orderBuffetCustomerList = OrderBuffetCustomer::where($orderField, $orderId)->select();
                    foreach ($orderBuffetCustomerList as $item) {
                        $refund_money = $item->total_consumption_tax_pay_price;
                        $refund_consumption_tax = $item->consumption_tax;
                        $item->save([
                            'refund_money' => $refund_money,
                            'refund_num' => $item->num,
                            'refund_consumption_tax' => $refund_consumption_tax,
                        ]);
                    }
                    // 加钟
                    $orderDelayList = OrderDelay::where($orderField, $orderId)->select();
                    foreach ($orderDelayList as $item) {
                        $item->save(['refund_money' => $item->total_price, 'refund_num' => $item->num]);
                    }
                }
                //
            }
            // 部分退款
            else if ($refund_type == 2) {
                if (empty($refund_product) && empty($refund_buffet) && empty($refund_delay)) {
                    $this->error = '退款商品不能为空';
                    return false;
                }
                // 商品
                foreach ($refund_product as $k => $item) {
                    $op = (new OrderProduct)->where('order_product_id', $item['order_product_id'])->find();
                    if ($item['refund_num'] > ($op->total_num - $op->refund_num)) {
                        $this->error = '商品退款数量错误';
                        return false;
                    }
                    // 订单商品表保存退款数据
                    $product_price = helper::bcdiv($op->total_price, $op->total_num);                                                                   // 商品应付单价
                    $consumption_tax = helper::bcdiv($op->consumption_tax, $op->total_num);                                                             // 商品所有税单价
                    $unit_product_consumption_tax = helper::bcdiv($op->product_consumption_tax, $op->total_num);                                        // 商品消费税单价
                    $unit_product_service_fee = helper::bcdiv($op->product_service_fee, $op->total_num);                                                // 商品服务费单价
                    $unit_product_service_consumption_tax = helper::bcdiv($op->product_service_consumption_tax, $op->total_num);                        // 商品服务费消费税单价
                    $product_price = $op->tax_calc_type == 2 ? helper::bcadd($product_price, $unit_product_consumption_tax) : $product_price;           // 含税单价
                    $product_price = helper::bcadd($product_price, helper::bcadd($unit_product_service_fee, $unit_product_service_consumption_tax));    // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
                    //
                    $refund_product[$k]['product_name'] = $op->product_name;
                    $refund_product[$k]['product_attr'] = $op->getData('product_attr');
                    $refund_product[$k]['product_price'] = $product_price;
                    //
                    $total_op_refund_money = helper::bcmul($product_price, $item['refund_num']);  // total商品应付总退
                    $total_op_refund_consumption_tax = helper::bcmul($consumption_tax, $item['refund_num']);  // total商品消费税总退
                    $op_refund_num = helper::bcadd($op->refund_num, $item['refund_num']);
                    $op_refund_money = helper::bcadd($op->refund_money, $total_op_refund_money);
                    $op_consumption_tax = helper::bcadd($op->refund_consumption_tax, $total_op_refund_consumption_tax);
                    $opSaveData = [
                        'refund_num' => $op_refund_num,
                        'refund_money' => $op_refund_money,
                        'refund_consumption_tax' => $op_consumption_tax,
                    ];
                    $op->save($opSaveData);
                    // 主表累计
                    $total_refund_money = helper::bcadd($total_refund_money, $total_op_refund_money);
                    $total_refund_consumption_tax = helper::bcadd($total_refund_consumption_tax, $total_op_refund_consumption_tax);
                    // 子单累计
                    if (isset($reloadOrderIds[$op['order_id']])) {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = helper::bcadd($reloadOrderIds[$op['order_id']]['refund_money'], $total_op_refund_money);
                        $reloadOrderIds[$op['order_id']]['refund_consumption_tax'] = helper::bcadd($reloadOrderIds[$op['order_id']]['refund_consumption_tax'], $total_op_refund_consumption_tax);
                    } else {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = $total_op_refund_money;
                        $reloadOrderIds[$op['order_id']]['refund_consumption_tax'] = $total_op_refund_consumption_tax;
                    }
                }
                // 自助餐
                foreach ($refund_buffet as $k => $item) {
                    $op = (new OrderBuffetCustomer)->where('id', $item['id'])->find();
                    if ($item['refund_num'] > ($op->num - $op->refund_num)) {
                        $this->error = '商品退款数量错误';
                        return false;
                    }
                    $refund_buffet[$k]['name'] = $op->buffet_name;
                    $refund_buffet[$k]['customer_type_name'] = $op->customer_type_name;
                    $refund_buffet[$k]['price'] = $op->price;
                    // 订单商品表保存退款数据
                    $buffet_price = helper::bcdiv($op->total_pay_price, $op->num);                                                                      // 商品应付单价
                    $consumption_tax = helper::bcdiv($op->consumption_tax, $op->num);                                                                   // 商品所有税单价
                    $unit_buffet_consumption_tax = helper::bcdiv($op->product_consumption_tax, $op->num);                                               // 商品消费税单价
                    $unit_buffet_service_fee = helper::bcdiv($op->product_service_fee, $op->num);                                                       // 商品服务费单价
                    $unit_buffet_service_consumption_tax = helper::bcdiv($op->product_service_consumption_tax, $op->num);                               // 商品服务费消费税单价
                    $buffet_price = $op->tax_calc_type == 2 ? helper::bcadd($buffet_price, $unit_buffet_consumption_tax) : $buffet_price;               // 含税单价
                    $buffet_price = helper::bcadd($buffet_price, helper::bcadd($unit_buffet_service_fee, $unit_buffet_service_consumption_tax));        // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
                    //
                    $total_op_refund_money = helper::bcmul($buffet_price, $item['refund_num']);
                    $total_op_refund_consumption_tax = helper::bcmul($consumption_tax, $item['refund_num']);  // total 消费税总退
                    $op_refund_num = helper::bcadd($op->refund_num, $item['refund_num']);
                    $op_refund_money = helper::bcadd($op->refund_money, $total_op_refund_money);
                    $op_consumption_tax = helper::bcadd($op->refund_consumption_tax, $total_op_refund_consumption_tax);
                    $opSaveData = [
                        'refund_num' => $op_refund_num,
                        'refund_money' => $op_refund_money,
                        'refund_consumption_tax' => $op_consumption_tax,
                    ];
                    $op->save($opSaveData);
                    //
                    // 主表累计
                    $total_refund_money = helper::bcadd($total_refund_money, $total_op_refund_money);
                    $total_refund_consumption_tax = helper::bcadd($total_refund_consumption_tax, $total_op_refund_consumption_tax);
                    // 子单累计
                    if (isset($reloadOrderIds[$op['order_id']])) {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = helper::bcadd($reloadOrderIds[$op['order_id']]['refund_money'], $total_op_refund_money);
                        $reloadOrderIds[$op['order_id']]['refund_consumption_tax'] = helper::bcadd($reloadOrderIds[$op['order_id']]['refund_consumption_tax'], $total_op_refund_consumption_tax);
                    } else {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = $total_op_refund_money;
                        $reloadOrderIds[$op['order_id']]['refund_consumption_tax'] = $total_op_refund_consumption_tax;
                    }
                }
                // 加钟
                foreach ($refund_delay as $k => $item) {
                    $op = (new OrderDelay)->where('id', $item['id'])->find();
                    if ($item['refund_num'] > ($op->num - $op->refund_num)) {
                        $this->error = '商品退款数量错误';
                        return false;
                    }
                    $refund_delay[$k]['name'] = $op->name;
                    $refund_delay[$k]['price'] = $op->price;
                    //
                    $delay_price = helper::bcdiv($op->total_price, $op->num);   // 商品应付单价
                    $total_op_refund_money = helper::bcmul($delay_price, $item['refund_num']);
                    // 订单商品表保存退款数据
                    $op_refund_num = helper::bcadd($op->refund_num, $item['refund_num']);
                    $op_refund_money = helper::bcadd($op->refund_money, $total_op_refund_money);
                    $opSaveData = [
                        'refund_num' => $op_refund_num,
                        'refund_money' => $op_refund_money,
                    ];
                    $op->save($opSaveData);
                    // 主表累计
                    $total_refund_money = helper::bcadd($total_refund_money, $total_op_refund_money);
                    // 子单累计
                    if (isset($reloadOrderIds[$op['order_id']])) {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = helper::bcadd($reloadOrderIds[$op['order_id']]['refund_money'], $total_op_refund_money);
                    } else {
                        $reloadOrderIds[$op['order_id']]['refund_money'] = $total_op_refund_money;
                    }
                }
                $refundable_amount = floatval(helper::bcsub($order['pay_price'], $order['refund_money']));
                if ($total_refund_money > $refundable_amount) {
                    $this->error = '退款金额不能大于实付金额' . $refundable_amount;
                    return false;
                }

                if ($total_refund_money <= 0) {
                    $this->error = '无法退款';
                    return false;
                }
                $saveData = [
                    'order_id' => $order['order_id'],
                    'refund_type' => $refund_type,
                    'refund_method' => $refund_method,
                    'refund_money' => $total_refund_money,
                    'cashier_id' => $cashier_id,
                ];
                $orderRefundLog = (new OrderRefund)->create($saveData);
            } else {
                $this->error = '退款类型错误';
                return false;
            }

            // 页面操作退款金额
            $initRefundMoney = 0;
            $initRefundMoney = $total_refund_money;

            // 更新订单状态
            // 合单处理
            if ($order['is_merge']) {
                /**
                 * 部分退款
                 */
                if ($refund_type == 2) {
                    // 处理主单
                    $total_refund_money = helper::bcadd($total_refund_money, $order['refund_money']);
                    $order_refund_consumption_tax = helper::bcadd($total_refund_consumption_tax, $order['refund_consumption_tax']);
                    $this->save([
                        'refund_money' => $total_refund_money,
                        'refund_consumption_tax' => $order_refund_consumption_tax,
                    ]);
                    // 处理子单
                    foreach ($reloadOrderIds as $sub_order_id => $sub_item) {
                        $subOrder = Order::where('order_id', $sub_order_id)->find();
                        $s_total_refund_money = helper::bcadd($sub_item['refund_money'], $subOrder['refund_money']);
                        $s_total_consumption_tax = helper::bcadd($sub_item['refund_consumption_tax'], $subOrder['refund_consumption_tax']);
                        $subOrder->save([
                            'refund_money' => $s_total_refund_money,
                            'refund_consumption_tax' => $s_total_consumption_tax,
                        ]);
                    }
                }
                /**
                 *  整单退款
                 */
                else {
                    // 处理主单
                    $total_refund_money = helper::bcadd($total_refund_money, $order['refund_money']);
                    $this->save([
                        'refund_money' => $total_refund_money,
                        'refund_consumption_tax' => $order['consumption_tax_money'],
                    ]);
                    // 处理子单
                    foreach ($order['mergeList'] as $item) {
                        $subOrder = Order::where('order_id', $item['order_id'])->find();
                        $subOrder->save([
                            'refund_money' => $subOrder['pay_price'],
                            'refund_consumption_tax' => $subOrder['consumption_tax_money'],
                        ]);
                    }
                }
            }
            // 独单处理
            else {
                /**
                 * 部分退款
                 */
                if ($refund_type == 2) {
                    $total_refund_money = helper::bcadd($total_refund_money, $order['refund_money']);
                    $order_refund_consumption_tax = helper::bcadd($total_refund_consumption_tax, $order['refund_consumption_tax']);
                    $this->save([
                        'refund_money' => $total_refund_money,
                        'refund_consumption_tax' => $order_refund_consumption_tax,
                    ]);
                    // 拆单主单处理
                    if ($order['parent_id'] > 0) {
                        $parentOrder = Order::where('order_id', $order['parent_id'])->find();
                        $parentOrder->save([
                            'refund_money' => helper::bcadd($parentOrder['refund_money'], $total_refund_money),
                            'refund_consumption_tax' => $order_refund_consumption_tax,
                        ]);
                    }
                } else {
                    $total_refund_money = $order['pay_price'];
                    $this->save([
                        'refund_money' => $total_refund_money,
                        'refund_consumption_tax' => $order['consumption_tax_money'],
                    ]);
                    // 拆单主单处理
                    if ($order['parent_id'] > 0) {
                        $parentOrder = Order::where('order_id', $order['parent_id'])->find();
                        $parentOrder->save([
                            'refund_money' => helper::bcadd($parentOrder['refund_money'], $total_refund_money),
                            'refund_consumption_tax' => $order['consumption_tax_money'],
                        ]);
                    }
                }
            }

            // 处理主单高峰时间段退款记录 - 等待所有数据处理完成再记入峰值，保证最终数值正确(v1.1.1版本)
            $PeakTimeOrderId = $this->parent_id > 0 ? $this->parent_id : $this->order_id;
            (new OrderPeakTime)->record('desc', $PeakTimeOrderId, $initRefundMoney);

            // 添加退款记录目的地
            $refundDestination = OrderRefundDestination::createData($this, $orderRefundLog, $refund_type, $refund_method, $cashier_id, $bank_code, $account_no, $account_name);
            $orderProductReturnList = $refundDestination['orderProductReturnList'];
            $cashRefundMoney = $refundDestination['cashRefundMoney'];

            // 添加操作记录
            OrderOperationLog::createLog($order['order_id'], OrderOperationLog::ACTION_REFUND, [
                'pay_type' => $orderProductReturnList,          //  支付方式
                'refund_type' => $refund_type,                  //  退款类型
                'refund_method' => $refund_method,              //  退款方式
                'refund_product' => $refund_product,            //  退款产品
                'refund_buffet' => $refund_buffet,              //  退款自助餐
                'refund_delay' => $refund_delay,                //  退款加钟
                'parent_id' => $order['parent_id'],             //  拆单主单ID
                'order_name' => $order['order_name'],           //  订单名称
            ], '退款');

            // 添加店铺账户余额
            if (ClientHelp::verifyClientVersion('1.0.8', '>')) {
                if ($cashRefundMoney > 0) {
                    Account::updateAmount(0, $cashRefundMoney, $order['order_no'], $cashier_id, $order['shop_supplier_id'], $order['app_id'], 'order-refund');
                }
            } else {
                if ($orderRefundLog && $orderRefundLog->refund_method == 2) {
                    Account::updateAmount(0, $orderRefundLog->refund_money, $order['order_no'], $cashier_id, $order['shop_supplier_id'], $order['app_id'], 'order-refund');
                }
            }

            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->errorData = ['code' => $e->getCode()];
            $this->rollback();
            return false;
        }
    }

    // 获取主单信息
    public static function getPreOrderInfo($order_id)
    {
        $order = self::where('order_id', $order_id)->find();
        if (!$order) {
            return [
                'pay_order_id' => $order_id,
                'pay_price' => 0,
                'final_price' => 0,
                'pay_fee_money' => 0,
                'order_price' => 0,
                'pay_type' => [],
                'user' => null,
            ];
        }
        $pay_price = helper::bcsub($order->pay_price, $order->pay_fee_money);

        // 最终应收应该减去结账抹零金额 v1.1.0
        if ($order->checkout_discount_type > 0) {
            $final_price = helper::bcsub($order->pay_price, $order->checkout_diff_money);
        } else {
            $final_price = $order->pay_price;
        }

        return [
            'pay_order_id' => $order_id,            // 主单order_id, 或独单的order_id
            'pay_price' => floatval($pay_price),    // 合计应收（不包含支付手续费）
            'final_price' => $final_price,     // 最终应收
            'pay_fee_money' => $order->pay_fee_money,
            'checkout_diff_money' => $order->checkout_diff_money,       // 结账抹零金额
            'checkout_discount_type' => $order->checkout_discount_type, // 结账抹零类型
            'order_price' => $order->order_price,
            'pay_type' => self::getPayTypeList($order_id),
            'user' => UserModel::detail($order->user_id),
        ];
    }

    /**
     * 是否能取消进行中订单
     * @param $order_id
     * @return bool
     */
    public function isCancelUnderwayOrder($order_id)
    {
        /** @var self $detail */
        $detail = self::with(['payType'])->where(['order_id' => $order_id, 'pay_status' => OrderPayStatusEnum::PENDING, 'order_status' => OrderStatusEnum::NORMAL])->find();
        if (!$detail) {
            $this->error = '订单不存在';
            return false;
        }
        // 检查部分支付
        if ($error = $detail->checkPartialPayment()) {
            $this->error = $error;
            return false;
        }
        return true;
    }

    /**
     * 取消开台订单
     * @param $remark
     * @return bool
     */
    public function CashierOrderCancels($remark = '')
    {
        if ($this->pay_status['value'] == OrderPayStatusEnum::SUCCESS) {
            $this->error = "订单已付款，不允许取消";
            return false;
        }
        // 变更子订单状态
        self::where('parent_id', $this->order_id)->update(['order_status' => OrderStatusEnum::CANCELLED]);

        return $this->save([
            'cancel_remark' => $remark,
            'order_status' => OrderStatusEnum::CANCELLED,
            'merge_id' => '', // 退出合单
            'merge_parent_id' => '',
        ]);
    }

    /**
     * 订单赠菜
     * @param array $param
     *   int order_product_id   订单产品ID，必填
     *   array free_tag_ids        赠菜标签ID，必填
     *   string free_remark     赠菜备注，选填
     * @return bool
     */
    public function productFree($param)
    {
        $order_product_id = $param['order_product_id'] ?? 0;
        $free_tag_ids = $param['free_tag_ids'] ?? [];
        $free_remark = $param['free_remark'] ?? '';
        if (!$free_tag_ids && !$free_remark) {
            $this->error = "赠菜原因不能为空";
            return false;
        }
        /** @var OrderProduct $orderProduct */
        $orderProduct = OrderProduct::where('order_product_id', $order_product_id)->find();
        if (!$orderProduct) {
            $this->error = "商品不存在";
            return false;
        }

        if (!$orderProduct->toFree($free_tag_ids, $free_remark)) {
            $this->error = $orderProduct->getError();
            return false;
        }
        // 兼容拆单订单信息
        $splitOrder = $orderProduct->orderM()->field(['order_id', 'parent_id', 'order_name'])->find();
        // 添加操作记录
        OrderOperationLog::createLog($splitOrder->order_id, OrderOperationLog::ACTION_PRODUCT_FREE, [
            'order_product_id' => $orderProduct->order_product_id,
            'product_id' => $orderProduct->product_id,
            'product_name' => $orderProduct->product_name,
            'product_attr' => $orderProduct->getData('product_attr'),
            'product_price' => $orderProduct->product_price,
            'total_num' => $orderProduct->total_num,
            'total_price' => $orderProduct->total_price,
            'free_tag_ids' => $free_tag_ids,
            'free_remark' => $free_remark,
            'parent_id' => $splitOrder->parent_id,   // 拆单主单ID
            'order_name' => $splitOrder->order_name, // 订单名称
        ], '订单赠菜');

        // 重新计算价格
        $this->reloadPrice($orderProduct->order_id);
        return true;
    }

    /**
     * 订单-取消赠菜（1.1.0）
     * @param array $param
     *   int order_product_id   订单产品ID，必填
     * @return bool
     */
    public function cancelProductFree($order_product_id)
    {
        /** @var OrderProduct $orderProduct */
        $orderProduct = OrderProduct::where('order_product_id', $order_product_id)->find();
        if (!$orderProduct) {
            $this->error = "商品不存在";
            return false;
        }
        if ($orderProduct->is_free == 0) {
            $this->error = "非赠菜商品不能取消";
            return false;
        }
        //
        $this->startTrans();
        try {
            //
            $orderProduct->is_free = 0;
            $orderProduct->free_remark = '';
            $orderProduct->save();
            // 兼容拆单订单信息
            $splitOrder = $orderProduct->orderM()->field(['order_id', 'parent_id', 'order_name'])->find();
            // 添加操作记录
            OrderOperationLog::createLog($splitOrder->order_id, OrderOperationLog::ACTION_CANCEL_PRODUCT_FREE, [
                'order_product_id' => $orderProduct->order_product_id,
                'product_id' => $orderProduct->product_id,
                'product_name' => $orderProduct->product_name,
                'product_attr' => $orderProduct->getData('product_attr'),
                'product_price' => $orderProduct->product_price,
                'total_num' => $orderProduct->total_num,
                'total_price' => $orderProduct->no_free_total_pay_price,
                'parent_id' => $splitOrder->parent_id,   // 拆单主单ID
                'order_name' => $splitOrder->order_name, // 订单名称
            ], '订单取消赠菜');
            //
            $this->reloadPrice($orderProduct['order_id']);
            //
            $this->commit();
            //
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
        //
        return true;
    }

    /**
     * 订单转菜
     * @param array $param
     *   int order_product_id   订单产品ID，必填
     *   int move_to_order_id        赠菜标签ID，必填
     * @return bool
     */
    public function moveProduct($param, $cashier)
    {
        $order_product_id = $param['order_product_id'] ?? 0;
        $move_to_order_id = $param['move_to_order_id'] ?? 0;
        if (!$move_to_order_id) {
            $this->error = "move_to_order_id必填";
            return false;
        }
        /** @var Order $toOrder */
        $toOrder = self::where(['order_id' => $move_to_order_id, 'order_status' => OrderStatusEnum::NORMAL])->find();
        if (!$toOrder) {
            $this->error = '目标桌台订单不可操作';
            return false;
        }
        if ($toOrder->is_lock == 1) {
            $this->error = '目标桌台订单已被锁定，请解锁后重新操作';
            return false;
        }
        // 如果是拆单，转到拆单1，并需要判断是否是未支付状态
        if ($toOrder->subOrder->count() > 1) {
            $toOrder = self::where('parent_id', $toOrder->order_id)->order('order_id', 'asc')->find();
            if ($toOrder->getData('pay_status') == OrderPayStatusEnum::SUCCESS) {
                $this->error = '订单已结账';
                return false;
            }
        }
        /** @var OrderProduct $orderProduct */
        $orderProduct = OrderProduct::where('order_product_id', $order_product_id)->find();
        if (!$orderProduct) {
            $this->error = "当前桌订单商品不存在";
            return false;
        }
        // 转菜用
        $free_tag_order_product_id = OrderProductFree::where('order_product_id', $orderProduct->order_product_id)->find() ? $orderProduct->order_product_id : 0;
        //
        $data = [
            'order_id' => $toOrder->parent_id == 0 ? $toOrder->order_id : $toOrder->parent_id,
            'sub_order_id' => $toOrder->parent_id == 0 ? 0 : $toOrder->order_id,
            'add_source' => 1,
            'product_id' => $orderProduct->product_id,
            'product_sku_id' => $orderProduct->product_sku_id,
            'product_num' => $orderProduct->total_num,
            'describe' => $orderProduct->product_attr,
            'delivery' => 40,
            'feed_uuids' => json_decode($orderProduct->feed_uuids),
            'is_send_kitchen' => $orderProduct->is_send_kitchen,
            'send_kitchen_time' => $orderProduct->send_kitchen_time,
            'is_free' => $orderProduct->is_free,
            'free_remark' => $orderProduct->free_remark,
            'is_move' => 1,
            'remark' => $orderProduct->remark,
            'finish_num' => $orderProduct->finish_num,
            'finish_time' => $orderProduct->finish_time,
            'move_from_table_id' => $orderProduct->table_id,
            'move_from_order_id' => $orderProduct->order_id,
            'is_change_price' => $orderProduct->is_change_price,
            'product_price' => $orderProduct->product_price,
            'free_tag_order_product_id' => $free_tag_order_product_id,
        ];
        $this->startTrans();
        try {
            if (!$toOrder->addToOrder($data, $cashier)) {
                $this->error = $toOrder->getError();
                return false;
            }
            // 兼容拆单订单信息
            $splitOrder = $orderProduct->orderM()->field(['order_id', 'parent_id', 'order_name'])->find();
            // 添加操作记录
            OrderOperationLog::createLog($splitOrder->order_id, OrderOperationLog::ACTION_PRODUCT_MOVE, [
                'order_product_id' => $orderProduct->order_product_id,
                'product_id' => $orderProduct->product_id,
                'product_name' => $orderProduct->product_name,
                'product_attr' => $orderProduct->getData('product_attr'),
                'total_num' => $orderProduct->total_num,
                'to_order_id' => $toOrder->order_id,
                'to_table_id' => $toOrder->table_id,
                'to_table_no' => $toOrder->table_no,
                'parent_id' => $splitOrder->parent_id,   // 拆单主单ID
                'order_name' => $splitOrder->order_name, // 订单名称
            ], '订单转菜');
            //
            $orderProduct->force()->delete();
            $this->reloadPrice($param['move_from_order_id']);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 免单标签
     * @return array
     */
    public static function freeTag()
    {
        return FreeTag::field(['id', 'free_tag'])->select()->toArray();
    }

    /**
     * 退菜标签
     * @return array
     */
    public static function returnTag()
    {
        return ReturnReason::field(['id', 'reason'])->select()->toArray();
    }

    /**
     * 订单免单
     * @param $param
     *   int order_product_id   订单产品ID，必填
     *   array free_tag_ids     赠菜标签ID，必填
     *   string free_remark     赠菜备注，选填
     * @param $cashier
     * @return bool
     */
    public function orderToFree($param, $cashier, $deviceId = '')
    {
        $final_price = $param['final_price'] ?? 0;
        $free_tag_ids = $param['free_tag_ids'] ?? [];
        $free_remark = $param['free_remark'] ?? '';
        $param = [
            'delivery' => 40,
            'final_price' => $final_price,
            'is_free' => 1,
            'free_remark' => $free_remark
        ];

        // 禁止并发操作
        $queue = new QueueHelp('ORDER_FREE_' . $this->app_id . '_' . $this->order_id);
        $queue->while();

        //
        $this->startTrans();
        try {
            // 判断合单
            if ($this->merge_id) {
                // 添加免单支付方式
                /** @var Order $masterOrder */
                $masterOrder = Order::where('order_id', $this->merge_parent_id)->where('order_status', OrderStatusEnum::NORMAL)->find();
                if (!$masterOrder) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '未生成主单';
                    return false;
                }
                // 部分支付后不能使用免单
                if (OrderPayType::where('order_id', $masterOrder->order_id)->find()) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '当前订单已部分支付，无法操作免单';
                    return false;
                }
                if (!$masterOrder->toFreePayType()) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '添加免单支付失败';
                    return false;
                }
                // 添加操作记录
                OrderOperationLog::createLog($masterOrder->order_id, OrderOperationLog::ACTION_DISCOUNT, [
                    'price' => $masterOrder->pay_price,
                    'discount_type' =>  4,                       //折扣类型： 1-改价 2-折扣 3-抹零 4-免单
                    'parent_id' => $masterOrder->parent_id,      // 拆单主单ID
                    'order_name' => $masterOrder->order_name,    // 订单名称
                ], '优惠折扣');
                // 订单合并支付
                if ($this->orderMergePay($this->merge_id, $param, $cashier, $deviceId)) {
                    // 订单免单记录
                    if ($free_tag_ids) {
                        $freeTagList = FreeTag::whereIn('id', $free_tag_ids)->select()->toArray();
                        if ($freeTagList) {
                            $saveAllArr = [];
                            foreach ($freeTagList as $item) {
                                $saveAllArr[] = [
                                    'free_tag_id' => $item['id'],
                                    'free_tag' => $item['free_tag'],
                                    'order_id' => $masterOrder->order_id,
                                ];
                            }
                            (new OrderFree)->saveAll($saveAllArr);
                        }
                    }
                    $this->commit();
                    $queue->release();
                    return true;
                }
            } else {
                // 添加免单支付方式
                /** @var Order $subOrder */
                $subOrder = Order::where('order_id', $this->order_id)->where('order_status', OrderStatusEnum::NORMAL)->find();
                if (!$subOrder) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '当前订单不可修改';
                    return false;
                }
                // 部分支付后不能使用免单
                if (OrderPayType::where('order_id', $subOrder->order_id)->find()) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '当前订单已部分支付，无法操作免单';
                    return false;
                }
                if (!$subOrder->toFreePayType()) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '添加免单支付失败';
                    return false;
                }
                // 添加操作记录
                OrderOperationLog::createLog($subOrder->order_id, OrderOperationLog::ACTION_DISCOUNT, [
                    'price' => $subOrder->pay_price,
                    'discount_type' =>  4,                       //折扣类型： 1-改价 2-折扣 3-抹零 4-免单
                    'parent_id' => $subOrder->parent_id,         // 拆单主单ID
                    'order_name' => $subOrder->order_name,       // 订单名称
                ], '优惠折扣');
                //
                if ($this->orderPay($param, $cashier, $deviceId, false)) {
                    // 订单免单记录
                    OrderFree::where('order_id', $subOrder->order_id)->delete();
                    if ($free_tag_ids) {
                        $freeTagList = FreeTag::whereIn('id', $free_tag_ids)->select()->toArray();
                        if ($freeTagList) {
                            $saveAllArr = [];
                            foreach ($freeTagList as $item) {
                                $saveAllArr[] = [
                                    'free_tag_id' => $item['id'],
                                    'free_tag' => $item['free_tag'],
                                    'order_id' => $subOrder->order_id,
                                ];
                            }
                            (new OrderFree)->saveAll($saveAllArr);
                        }
                    }
                    //
                    $this->commit();
                    $queue->release();
                    //
                    event('CashierPaySuccess', $this);
                    //
                    return true;
                }
            }
            $this->rollback();
            //
            $queue->release();
            return false;
        } catch (BaseException $e) {
            $queue->release();
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 使用免单支付
     * @return bool
     */
    public function toFreePayType()
    {
        $value = -1;
        $saveArr = [
            'order_id' => $this->order_id,
            'pay_status' => 1,
            'value' => $value,
            'price' => $this->pay_price,
            'pay_hash' => OrderPayType::generateUniqueHash($this->order_id . '_0', $value),
        ];
        $orderPayType = (new OrderPayType);
        $res = $orderPayType->save($saveArr);
        if (!$res) {
            return false;
        }

        // 更新主单免单支付方式
        if ($this->parent_id > 0) {
            $parentSaveArr = [
                'order_id' => $this->parent_id,
                'sub_id' => $orderPayType->id,
                'pay_status' => 1,
                'value' => $value,
                'price' => $this->pay_price,
                'pay_hash' => OrderPayType::generateUniqueHash($this->order_id . '_' . $orderPayType->id, $value),
            ];
            $res = (new OrderPayType())->save($parentSaveArr);
            if (!$res) {
                return false;
            }
        }
        return true;
    }

    /**
     * 获取订单（送厨+未送厨）总商品服务费(折后)
     * @param $order_id
     * @param $product_source
     * @return float
     */
    public static function getTotalProductServiceFee($order_id, $product_source)
    {
        $total_order_product_service_fee = OrderProduct::where('order_id', $order_id)->where('add_source', $product_source)->sum('product_service_fee');
        $total_buffet_service_fee = OrderBuffetCustomer::where('order_id', $order_id)->sum('product_service_fee');
        return floatval(helper::bcadd($total_order_product_service_fee, $total_buffet_service_fee));
    }

    /**
     * 获取订单（送厨+未送厨）总商品服务费消费税(折后)
     * @param $order_id
     * @param $product_source
     * @return float
     */
    public static function getTotalProductServiceConsumptionTax($order_id, $product_source)
    {
        $total_order_product_service_fee = OrderProduct::where('order_id', $order_id)->where('add_source', $product_source)->sum('product_service_consumption_tax');
        $total_buffet_service_fee = OrderBuffetCustomer::where('order_id', $order_id)->sum('product_service_consumption_tax');
        return floatval(helper::bcadd($total_order_product_service_fee, $total_buffet_service_fee));
    }

    /**
     * 获取订单（送厨+未送厨）总商品服务费(原价)
     * @param $order_id
     * @param $product_source
     * @return float
     */
    public static function getTotalProductOriginalServiceFee($order_id, $product_source)
    {
        $total_order_product_original_service_fee = OrderProduct::where('order_id', $order_id)->where('add_source', $product_source)->sum('product_original_service_fee');
        $total_buffet_original_service_fee = OrderBuffetCustomer::where('order_id', $order_id)->sum('product_original_service_fee');
        return floatval(helper::bcadd($total_order_product_original_service_fee, $total_buffet_original_service_fee));
    }

    /**
     * 获取订单（送厨+未送厨）总商品服务费消费税(原价)
     * @param $order_id
     * @param $product_source
     * @return float
     */
    public static function getTotalProductOriginalServiceConsumptionTax($order_id, $product_source)
    {
        $total_order_product_original_consumption_tax = OrderProduct::where('order_id', $order_id)->where('add_source', $product_source)->sum('product_original_service_consumption_tax');
        $total_buffet_original_consumption_tax = OrderBuffetCustomer::where('order_id', $order_id)->sum('product_original_service_consumption_tax');
        return floatval(helper::bcadd($total_order_product_original_consumption_tax, $total_buffet_original_consumption_tax));
    }

    /**
     * 获取订单（送厨+未送厨）总商品消费税(原价)
     * @param $order_id
     * @param $product_source
     * @return float
     */
    public static function getTotalProductOriginalConsumptionTax($order_id, $product_source)
    {
        $total_order_product_original_consumption_tax = OrderProduct::where('order_id', $order_id)->where('add_source', $product_source)->sum('product_original_consumption_tax');
        $total_buffet_original_consumption_tax = OrderBuffetCustomer::where('order_id', $order_id)->sum('product_original_consumption_tax');
        return floatval(helper::bcadd($total_order_product_original_consumption_tax, $total_buffet_original_consumption_tax));
    }

    /**
     * 订单解锁
     * @return bool
     */
    public function unlock()
    {
        if ($this->parent_id == 0) {
            $this->is_lock = 0;
            return $this->save();
        } else {
            return self::where('order_id', $this->parent_id)
                ->whereOr('parent_id', $this->parent_id)
                ->save(['is_lock' => 0, 'lock_time' => time()]);
        }
    }

    /**
     * 收银序号：每天从0001开始自增，如果存在边界值，需要查出昨天订单的最大值，然后加1
     *
     * @return string
     */
    public function getTableNumber($create_time)
    {
        $today = date('Ymd');
        $create_ymd = is_numeric($create_time)
            ? date('Ymd', $create_time)
            : date('Ymd', strtotime($create_time));

        // 如果是历史日期，直接查询历史数据
        if ($today != $create_ymd) {
            $start_time = strtotime(date('Y-m-d 00:00:00', strtotime($create_ymd)));
            $end_time = strtotime(date('Y-m-d 23:59:59', strtotime($create_ymd)));
            $order = Order::where('table_id', 0)
                ->where('call_no', '<>', '')
                ->whereBetweenTime('create_time', $start_time, $end_time)
                ->order('create_time', 'desc')
                ->field('call_no')
                ->limit(1)
                ->find();

            return str_pad(($order ? intval($order->call_no) : 0) + 1, 4, '0', STR_PAD_LEFT);
        }

        // 处理今天的序号
        $key = self::$app_id . "_table_number_" . $today;
        $number = Cache::get($key);

        if (!$number) {
            $number = 1;
        } else {
            // 防止缓存丢失，查询今日最后一个订单号
            $order = Order::where('table_id', 0)
                ->where('call_no', '<>', '')
                ->whereBetweenTime(
                    'create_time',
                    strtotime(date('Y-m-d 00:00:00')),
                    strtotime(date('Y-m-d 23:59:59'))
                )
                ->order('create_time', 'desc')
                ->field('call_no')
                ->limit(1)
                ->find();

            $number = $order ? max(intval($order->call_no) + 1, $number + 1) : $number + 1;
        }

        // 更新缓存并返回格式化的序号
        Cache::set($key, $number, 86400);
        return str_pad($number, 4, '0', STR_PAD_LEFT);
    }

    /**
     * 合并桌台
     * @param $order_list
     * @return bool
     */
    public function mergeOrders($merge_table_ids)
    {
        if ($this->isSplitTheOrder()) {
            $this->error = '当前桌台已拆单，不支持合并桌台';
            return false;
        }
        //
        $merge_time = time();
        $meal_num = 0;
        //
        $order_list = [];
        $paid_table_error = [];
        $paid_table_split_the_order_error = [];
        foreach ($merge_table_ids as $table_id) {
            $detail = self::getTableUnderwayOrder($table_id);
            if (!$detail) {
                $this->error = '桌台有变动，请重新选择';
                return false;
            }
            if ($detail->is_buffet) {
                $this->error = '自助餐桌台不符合并台';
                return false;
            }
            if ($detail->isSplitTheOrder()) {
                $paid_table_split_the_order_error[] = $detail->table_no;
            }
            if (OrderPayType::where('order_id', $detail->order_id)->find()) {
                $paid_table_error[] = $detail->table_no;
            }
            $order_list[] = $detail;
        }
        if (!empty($paid_table_error)) {
            $this->error = '以下桌台已被部分支付，不支持合并桌台';
            $this->errorData = $paid_table_error;
            return false;
        }
        if (!empty($paid_table_split_the_order_error)) {
            $this->error = '以下桌台已被拆单，不支持合并桌台';
            $this->errorData = $paid_table_split_the_order_error;
            return false;
        }
        //
        $this->startTrans();
        try {
            $tableNos = [];
            foreach ($order_list as $item_order) {
                //
                $tableNos[] = $item_order->table_no;
                //
                $meal_num += $item_order->meal_num;
                // 订单商品改变
                $updateArr = [
                    'order_id' => $this->order_id,
                    'merge_from_table_id' => $item_order->table_id,
                    'create_time' => $merge_time++,
                ];
                OrderProduct::where('order_id', $item_order->order_id)->update($updateArr);
                // 被并桌待接单记录变更为本桌
                TakeOrder::where('order_id', $item_order->order_id)->update(['order_id' => $this->order_id, 'table_id' => $this->table_id]);
                // 清台
                Table::close($item_order->table_id);
                // 清除订单
                $item_order->force()->delete();
            }
            if ($this->small_discount_type > 0 || $this->discount_change_price > 0 || $this->discount_change_price == -1 || $this->discount_ratio != 0) {
                $this->errorData = 1;
            } else {
                $this->errorData = 0;
            }
            // 更新本桌订单人数、改价\抹零\折扣\会员折扣\重置
            $updateCurOrderArr = [
                'meal_num' => $this->meal_num + $meal_num,
                'discount_change_price' => 0,
                'small_discount_type' => 0,
                'discount_ratio' => 0,
                'is_change_price' => 0,
            ];
            $this->save($updateCurOrderArr);
            //
            $this->reloadPrice($this->order_id);
            //
            OrderOperationLog::createLog($this->order_id, OrderOperationLog::ACTION_MERGE_TABLE, $tableNos, '并台');
            //
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 退菜 1.0.9 [退菜后回保留退菜商品在列表]
     * @param $order_product_id
     * @param $num
     * @param $return_reason
     * @return bool
     */
    public function returnProduct($order_product_id, $num, $return_ids = [], $custom_reason = '')
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单已完成,不允许退菜";
            return false;
        }

        $orderProduct = OrderProduct::where('order_product_id', $order_product_id)->find();
        if (!$orderProduct) {
            $this->error = "当前状态不可操作";
            return false;
        }
        if ($orderProduct->is_return == 1) {
            $this->error = "已退商品不可操作";
            return false;
        }

        if ($orderProduct['total_num'] < $num) {
            $this->error = "退菜数量不能大于当前商品数量";
            return false;
        }
        // 退菜原因
        $returnList = ReturnReason::where('id', 'in', $return_ids)->select()->toArray();
        $return_reason = [];
        foreach ($returnList as $item) {
            $return_reason[] = $item['reason'];
        }
        $return_reason = implode(';', $return_reason);
        // 退菜记录 - 订单id 产品id 产品规格 退菜原因 条件要一致
        $returnOrderProduct = OrderProduct::alias('op')
            ->leftJoin('order_product_return opr', 'opr.order_product_id = op.order_product_id')
            ->where('op.product_id', $orderProduct->product_id)
            ->where('op.order_id', $orderProduct->order_id)
            ->where('op.sub_order_id', $orderProduct->sub_order_id)
            ->where('op.product_attr', $orderProduct->getData('product_attr'))
            ->where('op.is_return', 1)
            ->where('op.is_send_kitchen', 1)
            ->where('op.merge_from_table_id', 0)
            ->where('op.is_change_price', $orderProduct->is_change_price)
            ->where('op.product_price', $orderProduct->product_price)
            ->where('op.is_free', $orderProduct->is_free)
            ->where('op.remark', $orderProduct->remark)
            ->where('opr.reason', $return_reason)
            ->where('opr.custom_reason', $custom_reason)
            ->find();
        //
        $this->startTrans();
        try {
            $isPay = $this['pay_status']['value'] == 20 ? 1 : 0;
            // 退回商品库存
            ProductFactory::getFactory($this['order_source'])->backProductStock([$orderProduct], $isPay);
            // 退菜的列处理
            if ($returnOrderProduct) {
                $returnOrderProduct->save([
                    'total_num' => $returnOrderProduct->total_num + $num,
                ]);
            } else {
                // 如果只有一个商品，直接退菜，不用新增
                if ($orderProduct['total_num'] != $num) {
                    $allowFields = [
                        'order_id',
                        'merge_from_table_id',
                        'product_id',
                        'delivery',
                        'is_buffet_product',
                        'product_name',
                        'send_kitchen_time',
                        'is_send_kitchen',
                        'is_return',
                        'is_free',
                        'is_move',
                        'send_kitchen_source',
                        'add_source',
                        'product_price',
                        'total_num',
                        'total_price',
                        'total_product_price',
                        'total_pay_price',
                        'no_free_total_pay_price',
                        'line_price',
                        'finish_num',
                        'finish_time',
                        'image_id',
                        'deduct_stock_type',
                        'spec_type',
                        'spec_sku_id',
                        'product_sku_id',
                        'product_attr',
                        'content',
                        'product_no',
                        'product_discount_money',
                        'tax_rate',
                        'consumption_tax',
                        'tax_calc_type',
                        'product_original_consumption_tax',
                        'product_original_service_consumption_tax',
                        'product_original_service_fee',
                        'product_consumption_tax',
                        'product_service_consumption_tax',
                        'product_service_fee',
                        'product_service_rate',
                        'no_free_product_service_consumption_tax',
                        'no_free_product_service_fee',
                        'no_free_product_consumption_tax',
                        'is_change_price',
                        'free_remark',
                        'move_from_table_id',
                        'move_from_order_id',
                        'product_weight',
                        'is_user_grade',
                        'grade_ratio',
                        'grade_product_price',
                        'grade_total_money',
                        'coupon_money_sys',
                        'coupon_money',
                        'points_money',
                        'points_num',
                        'points_bonus',
                        'feed_price',
                        'feed_ids',
                        'attr_ids',
                        'feed_uuids',
                        'refund_money',
                        'refund_consumption_tax',
                        'refund_num',
                        'supplier_money',
                        'sys_money',
                        'is_comment',
                        'remark',
                        'user_id',
                        'bag_price',
                        'discount_money',
                        'extra_times',
                        'kitchen_is_open',
                        'app_id',
                        'create_time',
                        'delete_time',
                        'sub_order_id',
                    ];
                    $orderProductArr = $orderProduct->toArray();
                    $returnOrderProduct = new OrderProduct;
                    foreach ($orderProductArr as $key => $value) {
                        if (in_array($key, $allowFields)) {
                            if ($key == 'total_num') {
                                $returnOrderProduct->$key = $num;
                            } elseif ($key == 'is_return') {
                                $returnOrderProduct->$key = 1;
                            } elseif ($key == 'create_time') {
                                $returnOrderProduct->$key = strtotime($value);
                            } elseif ($key == 'is_return') {
                                $returnOrderProduct->$key = 1;
                            } elseif ($key == 'product_attr') {
                                $returnOrderProduct->$key = $orderProduct->getData('product_attr'); // 原生数据
                            } elseif ($key == 'attr_ids') {
                                $returnOrderProduct->$key = $orderProduct->getData('attr_ids'); // 原生数据 解决获取器数据问题，不然会报：undefined array key 0
                            } elseif ($key == 'feed_ids') {
                                $returnOrderProduct->$key = $orderProduct->getData('feed_ids'); // 原生数据
                            } else {
                                $returnOrderProduct->$key = $value;
                            }
                        }
                    }
                    $returnOrderProduct->save();
                }
            }
            //
            $return_order_product_id = $returnOrderProduct ? $returnOrderProduct->order_product_id : $orderProduct['order_product_id'];
            if ($orderProduct['total_num'] == $num) {
                // 不要删除，不要变更订单产品id，不然无法统计同一个商品的退菜次数
                $orderProduct->save([
                    'is_return' => 1,
                ]);
                // 如果有拆分的记录，之前的删除
                if ($returnOrderProduct) {
                    $orderProduct->force()->delete();
                }
            } else {
                //
                $total_num = $orderProduct['total_num'] - $num;
                $orderProduct->save([
                    'total_num' => $total_num,
                ]);
            }
            // 退菜记录
            if ($num > 0) {
                // 相同订单商品退菜原因不用再次添加记录，只需更新数量
                $returnOrderProductReturn = OrderProductReturn::where('order_product_id', $return_order_product_id)
                    ->where('product_id', $orderProduct['product_id'])
                    ->where('reason', $return_reason)
                    ->where('custom_reason', $custom_reason)
                    ->find();
                if (!$returnOrderProductReturn) {
                    OrderProductReturn::add([
                        'order_id' => $this['order_id'],
                        'order_product_id' => $return_order_product_id,
                        'product_id' => $orderProduct['product_id'],
                        'num' => $num,
                        'reason' => $return_reason,
                        'custom_reason' => $custom_reason,
                    ]);
                }
            }
            // 添加操作记录
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_REFUND_PRODUCT, [
                'order_product_id' => $return_order_product_id,
                'product_id' => $orderProduct['product_id'],
                'product_name' => $orderProduct['product_name'],
                'product_attr' => $orderProduct->getData('product_attr'),
                'num' => $num,
                'reason' => $return_reason,
                'custom_reason' => $custom_reason,
                'parent_id' => $this['parent_id'],         // 拆单主单ID
                'order_name' => $this['order_name'],       // 订单名称
                'remark' => $orderProduct['remark'],       // 商品备注
            ], '退菜');
            //
            $this->reloadPrice($this['order_id']);
            // 重算主单
            if ($this['merge_parent_id']) {
                $this->reloadMasterMergeOrder($this['merge_parent_id'], $this['merge_id']);
            }
            //
            $this->commit();
            // 打印退菜单
            $printOrderProduct = $returnOrderProduct ? $returnOrderProduct : $orderProduct;
            $this['product'] = [$printOrderProduct];
            (new OrderPrinterService)->printProductTicket($this, Printing::PRINT_TYPE_BACK_FOOD);
            //
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 退菜 1.1.1 [取消退菜]
     * @param $order_product_id
     * @return bool
     */
    public function cancelReturnProduct($order_product_id)
    {
        if ($this['order_status']['value'] != 10) {
            $this->error = "订单已完成,不允许取消退菜";
            return false;
        }
        //
        $orderProduct = OrderProduct::where('order_product_id', $order_product_id)->find();
        if (!$orderProduct) {
            $this->error = "当前状态不可操作";
            return false;
        }
        if ($orderProduct->is_return == 0) {
            $this->error = "未退商品不可操作";
            return false;
        }
        //
        $this->startTrans();
        try {
            $orderProduct->is_return = 0;
            $orderProduct->is_send_kitchen = 0;
            $orderProduct->finish_num = 0;
            $orderProduct->finish_time = 0;
            $orderProduct->add_source = OrderProduct::CASHIER_ADD_PRODUCT;
            $orderProduct->save();
            // 兼容拆单订单信息
            $splitOrder = $orderProduct->orderM()->field(['order_id', 'parent_id', 'order_name'])->find();
            $orderProductReturn = $orderProduct->productReturn()->field(['reason', 'custom_reason'])->find();
            // 删除退菜记录
            OrderProductReturn::where('order_product_id', $order_product_id)->delete();
            //
            OrderOperationLog::createLog($this['order_id'], OrderOperationLog::ACTION_CANCEL_REFUND_PRODUCT, [
                'order_product_id' => $order_product_id,
                'product_id' => $orderProduct['product_id'],
                'product_name' => $orderProduct['product_name'],
                'product_attr' => $orderProduct->getData('product_attr'),
                'reason' => $orderProductReturn->getData('reason'),                 // 退菜原因
                'custom_reason' => $orderProductReturn->getData('custom_reason'),   // 自定义退菜原因
                'num' => $orderProduct->total_num,
                'parent_id' => $splitOrder['parent_id'],         // 拆单主单ID
                'order_name' => $splitOrder['order_name'],       // 订单名称
                'remark' => $orderProduct['remark'],             // 商品备注
            ], '取消退菜');
            //
            $this->commit();
            //
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 收银、助手、扫码获取必点方案接口
     * @return array
     */
    public function getSchemeBaseProductList()
    {
        $order = $this;
        //
        $orderSchemeProductList = [];
        // 生效的套餐
        if ($order['table_id'] > 0) {
            $meal_num = $order['meal_num'];
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $meal_num = 1;
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }
        //
        $scheme_product_ids = [];           // 方案的所有商品
        $scheme_must_page_product_ids = [];      // 必点页的所有商品
        if (!empty($OrderSchemeList)) {
            foreach ($OrderSchemeList as $scheme_item) {
                foreach ($scheme_item['product_ids'] as $product_item) {
                    $scheme_product_ids[$product_item['product_id']] = 1;
                }
            }
            // 方案单规格无属性无加料商品
            $schemeSingleSkuProductList = Product::alias('p')
                ->leftJoin('product_sku ps', 'p.product_id = ps.product_id')
                ->whereIn('p.product_id', array_keys($scheme_product_ids))
                ->where('p.delete_time', 0)
                ->where('p.product_status', 10)
                ->where('p.product_attr', '[]')       // 无属性
                ->where('p.product_feed', '[]')       // 无加料
                ->group('p.product_id')
                ->having('count(ps.product_sku_id)=1')  // 单规格
                ->column('p.product_id');
            // 方案中已加入购物车的必点商品
            $cartMustProductList = [];
            if (isset($order['product'])) {
                foreach ($order['product'] as $order_product_item) {
                    if ($order_product_item['scheme_id'] > 0) {
                        if (isset($cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']])) {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                    }
                }
            }
            //
            foreach ($OrderSchemeList as $scheme_item) {
                $product_num = 0; // 方案商品数量
                $scheme = [];
                foreach ($scheme_item['product_ids'] as $product_item) {
                    // 1、过滤固选择的单规格无属性无加料商品 2、过滤不自动加购的
                    if (($scheme_item['must_rule'] == 1 && in_array($product_item['product_id'], $schemeSingleSkuProductList)) || $scheme_item['auto_cart'] == 0) {
                        continue;
                    }
                    $product_num++;
                    //
                    $scheme['product_list'][] = [
                        'product_id' => $product_item['product_id'],
                        'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                        'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                    ];
                    //
                    $scheme_must_page_product_ids[$product_item['product_id']] = 1;
                }
                if (!$product_num) {
                    continue;
                }
                if ($scheme_item['must_rule'] == 2) {   // 可选
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $meal_num : 1; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                } else {
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $product_num * $meal_num : $product_num; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                }
                $scheme['scheme_id'] = $scheme_item['id'];                             // 方案ID
                $scheme['must_rule'] = $scheme_item['must_rule'];                      // must_rule 1-固定商品  2-可选商品
                $scheme['must_type'] = $scheme_item['must_type'];                      // must_type 1-每人一份 2-每单一份
                $scheme['auto_change'] = $scheme_item['auto_change'];                  // 顾客可修改必点数量 0-否 1-是
                $scheme['auto_check'] = $scheme_item['auto_check'];                    // 下单(送厨)检查 0-否 1-是
                $scheme['auto_checkout'] = $scheme_item['auto_checkout'];              // 结账检查 0-否 1-是
                $scheme['auto_cart'] = $scheme_item['auto_cart'];                      // 自动加入购物车 0-否 1-是
                $scheme['name'] = $scheme_item['name'];                                // 方案名
                //
                $orderSchemeProductList[] = $scheme;
            }
        }
        // 商品全信息
        $productList = [];
        if (!empty($scheme_product_ids)) {
            $order_id = 0;
            $list = (new ProductModel)->getBaseList(['order_id' => $order_id], false, array_keys($scheme_must_page_product_ids));
            $buffetProductArr = Order::getOrderBuffetProductArr($order_id);
            $productList = Order::handleBuffetProductIndex($list, $buffetProductArr, $meal_num);
        }
        $param['list_rows'] = 100;
        $model = new AssistantProductModel;
        $param['order_id'] = $order->order_id;
        $param['meal_num'] = $order->meal_num;
        $param['product_source'] = Order::TABLET_PRODUCT_SOURCE;
        $getParamChangeList = [];
        if ($order->is_buffet) {
            $list = $model->getBuffetChangeList($param, false);
            $deleteIds = $model->getDeleteProductIds();
            $getParamChangeList = compact('list', 'deleteIds');
        }
        return [
            'schemeProductList' => $orderSchemeProductList,
            'productBaseList' => $productList,
            'getParamChangeList' => $getParamChangeList,        // 自助餐变动
        ];
    }

    /**
     * 获取订单方案必点商品列表（返回 固定商品多规格\可选商品单规格\可选商品多规）
     * @param $need_scheme_product_ids  // 是否需要返回方案商品ID组
     * @return array
     */
    public function getSchemeProductList($need_scheme_product_ids = false)
    {
        $order = $this;
        //
        $orderSchemeProductList = [];
        // 生效的套餐
        if ($order['table_id'] > 0) {
            $meal_num = $order['meal_num'];
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $meal_num = 1;
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }

        // 存在方案
        $scheme_product_ids = [];   // 方案的所有商品
        if (!empty($OrderSchemeList)) {
            foreach ($OrderSchemeList as $scheme_item) {
                foreach ($scheme_item['product_ids'] as $product_item) {
                    $scheme_product_ids[$product_item['product_id']] = 1;
                }
            }
            // 方案单规格无属性无加料商品
            $schemeSingleSkuProductList = Product::alias('p')
                ->leftJoin('product_sku ps', 'p.product_id = ps.product_id')
                ->whereIn('p.product_id', array_keys($scheme_product_ids))
                ->where('p.delete_time', 0)
                ->where('p.product_status', 10)
                ->where('p.product_attr', '[]')       // 无属性
                ->where('p.product_feed', '[]')       // 无加料
                ->group('p.product_id')
                ->having('count(ps.product_sku_id)=1')  // 单规格
                ->column('p.product_id');
            // 方案中已加入购物车的必点商品
            $cartMustProductList = [];
            if (isset($order['product'])) {
                foreach ($order['product'] as $order_product_item) {
                    if ($order_product_item['scheme_id'] > 0) {
                        if (isset($cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']])) {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                    }
                }
            }
            //
            foreach ($OrderSchemeList as $scheme_item) {
                $product_num = 0;       // 必点页方案商品数量
                $product_auto_num = 0;  // 加购页方案商品数量
                $scheme = [
                    'product_auto_list' => [],
                    'product_list' => [],
                ];
                foreach ($scheme_item['product_ids'] as $product_item) {    // must_rule 1-固定商品 2-可选商品
                    // 1、过滤固选择的单规格无属性无加料商品 2、过滤不自动加购的(不自动加购不用出现在必点页)
                    if (($scheme_item['must_rule'] == 1 && in_array($product_item['product_id'], $schemeSingleSkuProductList)) || $scheme_item['auto_cart'] == 0) {
                        $scheme['product_auto_list'][] = [
                            'product_id' => $product_item['product_id'],
                            'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                            'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                        ];
                        $product_auto_num++;
                    } else {
                        $product_num++;
                        //
                        $scheme['product_list'][] = [
                            'product_id' => $product_item['product_id'],
                            'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                            'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                        ];
                    }
                }
                if ($product_num == 0 && $product_auto_num == 0) {
                    continue;
                }
                if ($scheme_item['must_rule'] == 2) {   // 可选
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $meal_num : 1; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                } else {
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $product_num * $meal_num : $product_num; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                }
                $scheme['scheme_id'] = $scheme_item['id'];                             // 方案ID
                $scheme['must_rule'] = $scheme_item['must_rule'];                      // must_rule 1-固定商品  2-可选商品
                $scheme['must_type'] = $scheme_item['must_type'];                      // must_type 1-每人一份 2-每单一份
                $scheme['auto_change'] = $scheme_item['auto_change'];                  // 顾客可修改必点数量 0-否 1-是
                $scheme['auto_check'] = $scheme_item['auto_check'];                    // 下单(送厨)检查 0-否 1-是
                $scheme['auto_checkout'] = $scheme_item['auto_checkout'];              // 结账检查 0-否 1-是
                $scheme['auto_cart'] = $scheme_item['auto_cart'];                      // 自动加入购物车 0-否 1-是
                $scheme['name'] = $scheme_item['name'];                                // 方案名
                //
                $orderSchemeProductList[] = $scheme;
            }
        }

        if ($need_scheme_product_ids) {
            $scheme_product_ids = array_keys($scheme_product_ids);
            return compact('orderSchemeProductList', 'scheme_product_ids');
        } else {
            return $orderSchemeProductList;
        }
    }

    /**
     * @param $return_base_data  // 是否返回商品基础信息
     * @return array
     */
    public function getSchemeAllProductList($return_base_data = false)
    {
        $order = $this;
        //
        $productList = [];
        $orderSchemeProductList = [];
        // 生效的套餐
        if ($order['table_id'] > 0) {
            $meal_num = $order['meal_num'];
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $meal_num = 1;
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }

        // 必点商品
        $scheme_product_ids = [];   // 方案的所有商品
        if (!empty($OrderSchemeList)) {
            //
            foreach ($OrderSchemeList as $scheme_item) {
                foreach ($scheme_item['product_ids'] as $product_item) {
                    $scheme_product_ids[$product_item['product_id']] = 1;
                }
            }
            // 方案单规格无属性无加料商品
            $schemeSingleSkuProductList = Product::alias('p')
                ->leftJoin('product_sku ps', 'p.product_id = ps.product_id')
                ->whereIn('p.product_id', array_keys($scheme_product_ids))
                ->where('p.delete_time', 0)
                ->where('p.product_status', 10)
                ->where('p.product_attr', '[]')       // 无属性
                ->where('p.product_feed', '[]')       // 无加料
                ->group('p.product_id')
                ->having('count(ps.product_sku_id)=1')  // 单规格
                ->column('p.product_id');
            // 方案中已加入购物车的必点商品
            $cartMustProductList = [];
            if (isset($order['product'])) {
                foreach ($order['product'] as $order_product_item) {
                    if ($order_product_item['scheme_id'] > 0) {
                        if (isset($cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']])) {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                    }
                }
            }
            //
            foreach ($OrderSchemeList as $scheme_item) {
                $product_num = 0; // 方案商品数量
                $scheme = [
                    'product_list' => [],
                    'product_auto_cart_list' => []
                ];
                foreach ($scheme_item['product_ids'] as $product_item) {
                    // 单规格无属性无加料商品 自动加入购物车
                    if ($scheme_item['must_rule'] == 1 && in_array($product_item['product_id'], $schemeSingleSkuProductList)) {
                        $scheme['product_auto_cart_list'][] = [
                            'product_id' => $product_item['product_id'],
                            'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                            'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                        ];
                    } else {
                        $scheme['product_list'][] = [
                            'product_id' => $product_item['product_id'],
                            'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                            'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                        ];
                        $product_num++;
                    }
                }
                if (empty($scheme['product_list']) && empty($scheme['product_auto_cart_list'])) {
                    continue;
                }
                if ($scheme_item['must_rule'] == 2) {   // 可选
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $meal_num : 1; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                } else {
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $product_num * $meal_num : $product_num; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                }
                $scheme['scheme_id'] = $scheme_item['id'];                             // 方案ID
                $scheme['must_rule'] = $scheme_item['must_rule'];                      // must_rule 1-固定商品  2-可选商品
                $scheme['must_type'] = $scheme_item['must_type'];                      // must_type 1-每人一份 2-每单一份
                $scheme['auto_change'] = $scheme_item['auto_change'];                  // 顾客可修改必点数量 0-否 1-是
                $scheme['auto_check'] = $scheme_item['auto_check'];                    // 下单(送厨)检查 0-否 1-是
                $scheme['auto_checkout'] = $scheme_item['auto_checkout'];              // 结账检查 0-否 1-是
                $scheme['auto_cart'] = $scheme_item['auto_cart'];                      // 自动加入购物车 0-否 1-是
                $scheme['name'] = $scheme_item['name'];                                // 方案名
                //
                $orderSchemeProductList[] = $scheme;
            }
            // 商品全信息
            if ($return_base_data) {
                $order_id = 0;
                $list = (new ProductModel)->getBaseList(['order_id' => $order_id], false, array_keys($scheme_product_ids));
                $buffetProductArr = Order::getOrderBuffetProductArr($order_id);
                $productList = Order::handleBuffetProductIndex($list, $buffetProductArr, $meal_num);
            }
        }
        $param['list_rows'] = 100;
        $model = new TabletProductModel;
        $param['order_id'] = $order->order_id;
        $param['meal_num'] = $order->meal_num;
        $param['product_source'] = Order::TABLET_PRODUCT_SOURCE;
        $param['filter_product_ids'] = array_keys($scheme_product_ids);
        $getParamChangeList = [];
        if ($order->is_buffet) {
            $list = $model->getBuffetChangeList($param, false);
            $deleteIds = $model->getDeleteProductIds();
            $getParamChangeList = compact('list', 'deleteIds');
        }

        return [
            'schemeProductList' => $orderSchemeProductList,     // 方案
            'productBaseList' => $productList,                  // 商品基础信息
            'getParamChangeList' => $getParamChangeList,        // 自助餐变动
        ];
    }

    /**
     * 必点商品下单送厨检查
     * @param $check_type   1-送厨  2-结账
     * @return bool
     */
    public function checkSchemeMustProduct($check_type = 1)
    {
        $order = $this;
        // 生效的套餐
        if ($order['table_id'] > 0) {
            $meal_num = $order['meal_num'];
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $meal_num = 1;
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }

        if (!empty($OrderSchemeList)) {
            // 已加入购物车的必点商品
            $cartSchemeMustProductList = [];    // 按方案ID分已添加数量
            $cartMustProductList = [];          // 按商品ID分已添加数量
            if (isset($order['product'])) {
                foreach ($order['product'] as $order_product_item) {
                    if ($order_product_item['is_require'] == 1) {
                        if (isset($cartSchemeMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']])) {
                            $cartSchemeMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartSchemeMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                        //
                        if (isset($cartMustProductList[$order_product_item['product_id']])) {
                            $cartMustProductList[$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartMustProductList[$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                    }
                }
            }
            // 方案必点商品要求
            $orderSchemeProductList = [];
            foreach ($OrderSchemeList as $scheme_item) {
                $scheme = [];
                // check_type 1-送厨  2-结账
                if ($check_type == 1 && $scheme_item['auto_check'] == 0) {
                    continue;
                }
                if ($check_type == 2 && $scheme_item['auto_checkout'] == 0) {
                    continue;
                }
                $product_num = 0; // 方案商品数量
                foreach ($scheme_item['product_ids'] as $product_item) {
                    $product_num++;
                    //
                    $scheme['product_list'][] = [
                        'product_id' => $product_item['product_id'],
                        'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                    ];
                }
                if (!$product_num) {
                    continue;
                }
                if ($scheme_item['must_rule'] == 2) {   // 可选
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $meal_num : 1; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                } else {
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $product_num * $meal_num : $product_num; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                }
                $scheme['scheme_id'] = $scheme_item['id'];                             // 方案ID
                $scheme['must_rule'] = $scheme_item['must_rule'];                      // must_rule 1-固定商品  2-可选商品
                $scheme['must_type'] = $scheme_item['must_type'];                      // must_type 1-每人一份 2-每单一份
                $scheme['auto_change'] = $scheme_item['auto_change'];                  // 顾客可修改必点数量 0-否 1-是
                $scheme['auto_check'] = $scheme_item['auto_check'];                    // 下单(送厨)检查 0-否 1-是
                $scheme['auto_checkout'] = $scheme_item['auto_checkout'];              // 结账检查 0-否 1-是
                $scheme['auto_cart'] = $scheme_item['auto_cart'];                      // 自动加入购物车 0-否 1-是
                $scheme['name'] = $scheme_item['name'];                                // 方案名
                //
                $orderSchemeProductList[] = $scheme;
            }

            // 检查方案的必点商品是否已满足
            $error_must_product = [];
            foreach ($orderSchemeProductList as $scheme_item) {
                $miss_num = 0;  // 不足数量
                // 固选商品（需要每种商品都点）
                if ($scheme_item['must_rule'] == 1) {
                    foreach ($scheme_item['product_list'] as $product_item) {
                        // 购物车数量是否大于需求数量
                        $cart_num = isset($cartSchemeMustProductList[$scheme_item['scheme_id']][$product_item['product_id']]) ? $cartSchemeMustProductList[$scheme_item['scheme_id']][$product_item['product_id']] : 0;
                        $remaining_num = $cart_num - $product_item['need_num'];
                        // 不足继续从[cartSchemeMustProductList]没有方案ID的购物车商品匹配
                        if ($remaining_num < 0) {
                            $all_cart_num = isset($cartSchemeMustProductList[0][$product_item['product_id']]) ? $cartSchemeMustProductList[0][$product_item['product_id']] : 0;
                            $remaining_num = $cart_num + $all_cart_num - $product_item['need_num'];
                            // 最终依然不足记录
                            if ($remaining_num < 0) {
                                $expend_num = $all_cart_num;
                                // 累计方案不足数量
                                $miss_num += abs($remaining_num);
                            } else {
                                $expend_num = $product_item['need_num'] - $cart_num;
                            }
                            // 消耗 cartSchemeMustProductList 购物车商品数量
                            if (isset($cartSchemeMustProductList[0][$product_item['product_id']])) {
                                $cartSchemeMustProductList[0][$product_item['product_id']] = $cartSchemeMustProductList[0][$product_item['product_id']] - $expend_num;
                            }
                        }
                    }
                }
                // 可选商品
                else {
                    $total_cart_num = 0;
                    $total_need_num = $scheme_item['total_need_num'];
                    $expend_arr = [];
                    //
                    // 累计购物车已点的
                    if (isset($cartSchemeMustProductList[$scheme_item['scheme_id']])) {
                        foreach ($cartSchemeMustProductList[$scheme_item['scheme_id']] as $product_id => $num) {
                            $total_cart_num += $num;
                        }
                    }
                    $remaining_num =  $total_cart_num - $total_need_num;
                    // 不足继续从[cartSchemeMustProductList]没有方案ID的购物车商品匹配
                    if ($remaining_num < 0) {
                        foreach ($scheme_item['product_list'] as $product_item) {
                            // 累计购物车的可选商品总数量
                            $cart_num = isset($cartSchemeMustProductList[0][$product_item['product_id']]) ? $cartSchemeMustProductList[0][$product_item['product_id']] : 0;
                            $last_total_cart_num = $total_cart_num;
                            $total_cart_num += $cart_num;
                            // 需要选择的商品最终总数量total_cart_num大于total_need_num）
                            if ($total_cart_num >= $total_need_num) {
                                $expend_num = $total_need_num - $last_total_cart_num;
                                $expend_arr[$product_item['product_id']] = $expend_num;
                                break;
                            } else {
                                $expend_arr[$product_item['product_id']] = $cart_num;
                            }
                        }
                        // 最终依然不足累计方案不足数量
                        $remaining_num =  $total_cart_num - $total_need_num;
                        if ($remaining_num < 0) {
                            // 累计方案不足数量
                            $miss_num += abs($remaining_num);
                        }
                        // 消耗 $cartSchemeMustProductList 购物车商品数量
                        foreach ($expend_arr as $e_product_id => $e_num) {
                            if (isset($cartSchemeMustProductList[0][$e_product_id])) {
                                $cartSchemeMustProductList[0][$e_product_id] = $cartSchemeMustProductList[0][$e_product_id] - $e_num;
                            }
                        }
                    }
                }
                //
                if ($miss_num > 0) {
                    $error_must_product[] = [
                        'name' => $scheme_item['name'],
                        'num' => $miss_num
                    ];
                }
            }
            //
            if (!empty($error_must_product)) {
                $this->error = "当前订单的商品未选择必点商品，确定要继续结账吗？";
                $this->errorData = $error_must_product;
                $this->errorCode = OrderErrorEnum::MISS_MUST_PRODUCT;
                return false;
            }
        }

        return true;
    }

    /**
     * 检查多规格必点商品
     * @param $check_source
     * @return bool
     */
    public function checkSkuMustProduct($check_source = '')
    {
        $order = $this;
        //
        $orderSchemeProductList = [];
        // 生效的套餐
        if ($order['table_id'] > 0) {
            $meal_num = $order['meal_num'];
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $meal_num = 1;
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }

        // 必点商品
        if (!empty($OrderSchemeList)) {
            // 方案的所有商品
            $scheme_product_ids = [];
            foreach ($OrderSchemeList as $scheme_item) {
                foreach ($scheme_item['product_ids'] as $product_item) {
                    $scheme_product_ids[$product_item['product_id']] = 1;
                }
            }
            // 方案单规格无属性无加料商品
            $schemeSingleSkuProductList = Product::alias('p')
                ->leftJoin('product_sku ps', 'p.product_id = ps.product_id')
                ->whereIn('p.product_id', array_keys($scheme_product_ids))
                ->where('p.delete_time', 0)
                ->where('p.product_status', 10)
                ->where('p.product_attr', '[]')       // 无属性
                ->where('p.product_feed', '[]')       // 无加料
                ->group('p.product_id')
                ->having('count(ps.product_sku_id)=1')  // 单规格
                ->column('p.product_id');
            // 方案中已加入购物车的必点商品
            $cartMustProductList = [];
            if (isset($order['product'])) {
                foreach ($order['product'] as $order_product_item) {
                    if ($order_product_item['scheme_id'] > 0) {
                        if (isset($cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']])) {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] += $order_product_item['total_num'];
                        } else {
                            $cartMustProductList[$order_product_item['scheme_id']][$order_product_item['product_id']] = $order_product_item['total_num'];
                        }
                    }
                }
            }
            //
            foreach ($OrderSchemeList as $scheme_item) {
                $product_num = 0; // 方案商品数量
                $scheme = [];
                foreach ($scheme_item['product_ids'] as $product_item) {
                    // 1、过滤单规格无属性无加料且非可选的商品 2、过滤不自动加入购物车的
                    if (($scheme_item['must_rule'] != 2 && in_array($product_item['product_id'], $schemeSingleSkuProductList)) || $scheme_item['auto_cart'] == 0) {
                        continue;
                    }
                    $product_num++;
                    //
                    $scheme['product_list'][] = [
                        'product_id' => $product_item['product_id'],
                        'cur_num' => isset($cartMustProductList[$scheme_item['id']][$product_item['product_id']]) ? $cartMustProductList[$scheme_item['id']][$product_item['product_id']] : 0,
                        'need_num' => $scheme_item['must_type'] == 1 ? $meal_num : 1,
                    ];
                }
                if (!$product_num) {
                    continue;
                }
                if ($scheme_item['must_rule'] == 2) {   // 可选
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $meal_num : 1; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                } else {
                    $scheme['total_need_num'] = $scheme_item['must_type'] == 1 ? $product_num * $meal_num : $product_num; // 方案商品需求数量（受must_type影响 1-每人一份 2-每单一份）
                }
                $scheme['scheme_id'] = $scheme_item['id'];                             // 方案ID
                $scheme['must_rule'] = $scheme_item['must_rule'];                      // must_rule 1-固定商品  2-可选商品
                $scheme['name'] = $scheme_item['name'];
                $scheme['id'] = $scheme_item['id'];
                //
                $orderSchemeProductList[] = $scheme;
            }
        }

        // 检查
        $error_must_product = [];
        foreach ($orderSchemeProductList as $scheme_item) {
            $miss_num = 0;
            if ($scheme_item['must_rule'] == 1) {
                foreach ($scheme_item['product_list'] as $p) {
                    $dif_num = $p['need_num'] - $p['cur_num'];
                    $miss_num += max($dif_num, 0);
                }
            } else {
                $total_cur_num = 0;
                foreach ($scheme_item['product_list'] as $p) {
                    $total_cur_num += $p['cur_num'];
                }
                $dif_num = $scheme_item['total_need_num'] - $total_cur_num;
                $miss_num += max($dif_num, 0);
            }
            //
            if ($miss_num > 0) {
                $error_must_product[] = [
                    'scheme_id' => $scheme_item['id'],
                    'name' => $scheme_item['name'],
                    'num' => $miss_num
                ];
            }
        }
        //
        if (!empty($error_must_product)) {
            if ($check_source == 'confirmMust') {
                $this->error = "必点商品未选，请选择对应商品";
            } else {
                $this->error = "当前订单的商品未选择必点商品，确定要继续结账吗？";
            }
            $this->errorData = $error_must_product;
            $this->errorCode = OrderErrorEnum::MISS_MUST_PRODUCT;
            return false;
        }

        return true;
    }

    /**
     * 获取订单必点商品ids
     * @return array
     */
    public function getSchemeMustProductIds()
    {
        $order = $this;
        $product_ids = [];
        // 生效的套餐
        if ($order['table_id'] > 0) {
            // 桌台获得区域ID
            $area_id = Table::where('table_id', $order['table_id'])->value('area_id');
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_TABLE])
                ->whereRaw('FIND_IN_SET(?, table_area_ids)', [$area_id])
                ->where('status', 1)
                ->select()->toArray();
        } else {
            $OrderSchemeList = OrderScheme::whereRaw('FIND_IN_SET(?, use_channel)', [OrderScheme::USE_CHANNEL_ORDER])
                ->where('status', 1)
                ->select()->toArray();
        }
        foreach ($OrderSchemeList as $scheme_item) {
            foreach ($scheme_item['product_ids'] as $product_item) {
                $product_ids[$product_item['product_id']] = 1;
            }
        }

        return array_keys($product_ids);
    }

    /**
     * 订单确认多规格必点商品
     * @return bool
     */
    public function confirmMust()
    {
        // 多规格必点商品
        if (!$this->checkSkuMustProduct('confirmMust')) {
            $this->error = $this->getError();
            $this->errorData = $this->getErrorData();
            $this->errorCode = $this->getErrorCode();
            return false;
        }
        $update = [
            'is_must_notice' => 0
        ];
        return $this->save($update);
    }

    /**
     * 获取桌台的订单价格（已送厨+未送厨）
     * @param $order_id
     * @param $table_id
     * @return mixed
     */
    public function getTablePrice()
    {
        $cache_price = Cache::get($this->table_id . '_table_price' . ($this['app_id'] ?? 0));
        if ($cache_price == null) {
            return $this->getBackPayPrice();
        } else {
            return $cache_price;
        }
    }

    // 创建子单
    public function createSubOrder()
    {
        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queue->while();
        $this->startTrans();
        try {
            //
            $subCount = self::where('parent_id', $this->order_id)->count();
            if ($subCount >= 10) {
                $this->error = '拆单数量已达最大限制';
                $queue->release();
                return false;
            }
            if ($this->table_id == 0 && $this->call_no == '' && $this->parent_id == 0) {
                $this->call_no = $this->getTableNumber($this->create_time);
                $this->save();
            }

            //
            $resOrderid = $this->order_id;
            if ($subCount == 0) {
                // 第一次生成子单，重置所有优惠
                $businessSetting = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this->shop_supplier_id, $this->app_id);
                $smallAuto = isset($businessSetting['zeroing_method']) && in_array($businessSetting['zeroing_method'], [1, 2, 3, 4]);
                $appendData = [
                    'user_id' => 0,
                ];
                // 如果开启自动抹零，并且主单开启了自动抹零，则子单也开启自动抹零
                if ($smallAuto && $this->small_auto == 1) {
                    $appendData['small_auto'] = 1;
                    $appendData['small_diff_money'] = $this->small_diff_money;
                    $appendData['small_discount_type'] = $this->small_discount_type;
                }
                $this->handleCancelDiscount($appendData);
                /**
                 * 第一次生成一条主单[子单1]和一[子单2]
                 */
                // [子单1]
                $res = $this->createDefaultSubOrder();
                $resOrderid = $res->order_id;
                // [子单2]
                $this->createEmptySubOrder();
            } else {
                $this->createEmptySubOrder();
            }
            $reloadPriceOrder = $this->reloadPrice($this->order_id);
            $this->commit();
            $queue->release();
            return [$resOrderid, $reloadPriceOrder];
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $queue->release();
            $this->rollback();
            return false;
        }
    }

    // 创建主单的默认子单
    public function createDefaultSubOrder()
    {
        $cloneOrderData = $this->toArray();
        unset($cloneOrderData['order_id']);
        $cloneOrderData['order_no'] = $this->newOrderNo($this->order_source);
        $cloneOrderData['parent_id'] = $this->order_id;
        $cloneOrderData['table_id'] = 0;                // 会存在多个相同table_id进行中订单, 所以子单不能用主单table_id
        $cloneOrderData['extra_times'] = 0;             // 送厨对整单
        $cloneOrderData['is_buffet'] = $this->is_buffet;
        $cloneOrderData['is_must_notice'] = 0;
        $cloneOrderData['order_name'] = '1';
        $cloneOrderData['discount_method'] = $this->discount_method;
        $cloneOrderData['delivery_type'] = $this->delivery_type;
        $newCloneSubOrder = new Order;
        $newCloneSubOrder->save($cloneOrderData);
        // 标记订单商品
        OrderProduct::where('order_id', '=', $this->order_id)->update(['sub_order_id' => $newCloneSubOrder->order_id]);
        // 标记自助餐订单
        OrderBuffet::where('order_id', '=', $this->order_id)->update(['sub_order_id' => $newCloneSubOrder->order_id]);
        // 标记订单自助餐顾客
        OrderBuffetCustomer::where('order_id', '=', $this->order_id)->update(['sub_order_id' => $newCloneSubOrder->order_id]);
        // 标记订单加钟
        OrderDelay::where('order_id', '=', $this->order_id)->update(['sub_order_id' => $newCloneSubOrder->order_id]);
        return $newCloneSubOrder;
    }

    // 创建主单的空子单
    public function createEmptySubOrder()
    {
        $subOrderData['order_no'] = $this->newOrderNo($this->order_source);
        $subOrderData['parent_id'] = $this->order_id;
        $subOrderData['order_name'] = (string)($this->subOrder()->count() + 1);
        $subOrderData['table_id'] = 0;                          // 会存在多个相同table_id进行中订单, 所以子单不能用主单table_id
        $subOrderData['table_no'] = $this->table_no;
        $subOrderData['extra_times'] = $this->extra_times;      // 送厨对整单
        $subOrderData['is_buffet'] = $this->is_buffet;
        $subOrderData['is_must_notice'] = 0;
        $subOrderData['buyer_remark'] = '';
        $subOrderData['table_no'] = $this->table_no;
        $subOrderData['eat_type'] = $this->eat_type;
        $subOrderData['cashier_id'] = $this->cashier_id;
        $subOrderData['device_id'] = $this->device_id;
        $subOrderData['shop_supplier_id'] = $this->shop_supplier_id;
        $subOrderData['app_id'] = $this->app_id;
        $subOrderData['call_no'] = $this->call_no;
        $subOrderData['meal_num'] = $this->meal_num;
        $subOrderData['small_discount_type'] = $this->small_discount_type;
        $subOrderData['checkout_discount_type'] = $this->checkout_discount_type;
        $subOrderData['buffet_expired_time'] = $this->buffet_expired_time;
        $subOrderData['order_source'] = $this->order_source;
        $subOrderData['order_type'] = $this->order_type;
        $subOrderData['discount_method'] = $this->discount_method;
        $subOrderData['delivery_type'] = $this->delivery_type;
        $newSubOrder = new Order;
        $newSubOrder->save($subOrderData);
        return $newSubOrder;
    }

    /**
     * 删除子单
     * @param $params array 列表订单id order_id主单id, sub_order_id子单id, list_order_id列表订单id
     * @return bool|mixed
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public function deleteSubOrder($params)
    {
        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' . $this->order_id);
        $queue->while();
        //
        $reloadPriceOrder = null;
        // 开启事务
        $this->startTrans();
        try {
            $outputOrderId = $params['list_order_id'] ?? $this->order_id;
            // 查询子单详情
            $subOrder = (new self)->detail([
                'order_id' => $params['sub_order_id'],
                'parent_id' => $this->order_id
            ], ['product']);
            if (!$subOrder) {
                $this->error = '子单不存在';
                $queue->release();
                return false;
            }
            // 当前订单已结账，无法撤销
            if (self::checkOrderPaid($subOrder->order_id)) {
                $this->error = '当前订单已结账，无法撤销';
                $queue->release();
                return false;
            }

            // 查询子单1订单id，子单列表中最小id
            $minOrderId = (new self)->where('parent_id', $this->order_id)->where('order_id', '<=', $subOrder->order_id)->min('order_id');
            if ($subOrder->order_id == $minOrderId) {
                $this->error = '子单1不能删除';
                $queue->release();
                return false;
            }
            // 拆单1已结账，请结账当前拆单或删除商品后再删除拆单
            $minSubOrder = (new self)->detail([
                'order_id' => $minOrderId,
                'parent_id' => $this->order_id
            ], []);
            if ($minSubOrder->pay_status['value'] == OrderPayStatusEnum::SUCCESS && $subOrder->product()->count() > 0) {
                $this->error = '拆单1已结账，请结账当前拆单或删除商品后再删除拆单';
                $queue->release();
                return false;
            }

            // 商品转移到子单1
            foreach ($subOrder->product as $item) {
                $item->sub_order_id = $minOrderId;
                OrderProductModel::addOrIncNum($item, $item->total_num);
                $item->force()->delete();
            }
            // 自助餐转移到子单1
            $buffetPrice = 0;
            foreach ($subOrder->buffetCustomerType as $item) {
                $buffetPrice = $item->price;
                $item->sub_order_id = $minOrderId;
                OrderBuffetCustomerModel::addOrIncNum($item, $item->num);
                $item->force()->delete();
            }
            // 自助餐订单转移到子单1
            foreach ($subOrder->buffet as $item) {
                $item->sub_order_id = $minOrderId;
                OrderBuffetModel::addOrIncNum($item, $buffetPrice, $item->num);
                $item->force()->delete();
            }
            // 加钟转移到子单1
            foreach ($subOrder->delay as $item) {
                $item->sub_order_id = $minOrderId;
                OrderDelayModel::addOrIncNum($item, $item->num);
                $item->force()->delete();
            }

            // 删除子单
            if (!$subOrder->force()->delete()) {
                $this->error = '删除子单失败';
                $queue->release();
                return false;
            }
            // 重新生成order_name
            $orderCompleted = true;
            $subOrders = $this->subOrder()->field('order_id, parent_id, order_name,pay_status')->order('order_id', 'asc')->select();
            foreach ($subOrders as $key => $item) {
                if ($item->pay_status['value'] != OrderPayStatusEnum::SUCCESS) {
                    $orderCompleted = false;
                }
                $item->order_name = (string)($key + 1);
                $item->save();
            }

            // 如果有所子单都已结账，则完结整单
            if ($orderCompleted) {
                // 获取最后一个结账的子单
                $lastSubOrder = $this->subOrder()->where('pay_status', OrderPayStatusEnum::SUCCESS)->order('pay_time', 'desc')->find();
                // 完结整单
                self::handleOrderCompleted($this->order_id, [
                    'pay_time' => $lastSubOrder->pay_time,
                    'settle_type' => $lastSubOrder->settle_type,
                    'auto_close' => $lastSubOrder->auto_close,
                    'close_time' => $lastSubOrder->close_time,
                    'delivery_time' => $lastSubOrder->delivery_time,
                    'receipt_time' => $lastSubOrder->receipt_time,
                ]);
                $outputOrderId = $this->order_id;
            }

            // 如果只有一个子单，并且未结账，则撤销拆单
            if ($this->subOrder()->count() == 1) {
                $oneSubOrder = $this->subOrder()->where('pay_status', '<>', OrderPayStatusEnum::SUCCESS)->find();
                if ($oneSubOrder) {
                    // 撤销拆单，已重新计算订单价格，无需重复计算
                    if (!$this->revokeSubOrder(0)) {
                        $queue->release();
                        return false;
                    }
                } else {
                    // 重新计算订单价格
                    $reloadPriceOrder = $this->reloadPrice($this->order_id);
                }
                //
                $outputOrderId = $this->order_id;
            } else {
                // 重新计算订单价格
                $reloadPriceOrder = $this->reloadPrice($this->order_id);
            }
            // 提交事务
            $this->commit();
            $queue->release();
            return [$outputOrderId, $reloadPriceOrder];
        } catch (BaseException $e) {
            // 回滚事务
            $this->error = $e->getMessage();
            $queue->release();
            $this->rollback();
            return false;
        }
    }

    // 订单商品返回对应子单数据
    public function getSubOrderProduct($list, $sub_order_id)
    {
        return $list->filter(function ($item) use ($sub_order_id) {
            return $item['sub_order_id'] == $sub_order_id;
        });
    }

    // 自助餐商品返回对应子单数据
    public function getSubOrderBuffetCustomerType($list, $sub_order_id)
    {
        return $list->filter(function ($item) use ($sub_order_id) {
            return $item['sub_order_id'] == $sub_order_id;
        });
    }

    /**
     * 拆单-商品到子单
     * - data['main_order_id'] 主订单ID
     * - data['from_order_id'] 原订单ID
     * - data['to_order_id'] 目标订单ID
     * - data['product_list'] 商品列表
     */
    public function addToSubOrder(array $data)
    {
        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $this->app_id . '_' .  $data['to_order_id']);
        $queue->while();

        // 开启事务
        $this->startTrans();
        try {
            foreach ($data['product_list'] as $item) {
                $type = $item['type'] ?? 0; // 商品类型 0-商品 1-自助餐 2-加钟
                switch ($type) {
                    case 0:
                        // 商品
                        OrderProductModel::addToSubOrder($item['order_product_id'], $data['from_order_id'], $data['to_order_id'], $item['num']);
                        break;
                    case 1:
                        // 自助餐客户
                        $orderBuffetCustomer = OrderBuffetCustomerModel::addToSubOrder($item['order_product_id'], $data['from_order_id'], $data['to_order_id'], $item['num']);
                        // 自助餐订单
                        OrderBuffetModel::addToSubOrder($data['main_order_id'], $data['from_order_id'], $data['to_order_id'], $orderBuffetCustomer->price, $item['num']);
                        break;
                    case 2:
                        // 加钟
                        OrderDelayModel::addToSubOrder($item['order_product_id'], $data['from_order_id'], $data['to_order_id'], $item['num']);
                        break;
                    default:
                        $this->error = '商品类型错误';
                        $queue->release();
                        return false;
                }
            }
            // 重新计算订单价格
            $reloadPriceOrder = $this->reloadPrice($data['main_order_id']);
            // 提交事务
            $this->commit();
            $queue->release();
            return [$data['to_order_id'], $reloadPriceOrder];
        } catch (BaseException $e) {
            // 回滚事务
            $this->error = $e->getMessage();
            $queue->release();
            $this->rollback();
            return false;
        }
    }

    /**
     * 撤销子单
     */
    public function revokeSubOrder($user_id)
    {
        // 禁止并发操作
        $queue = new QueueHelp("REVOKE_SUBORDER_" . $this->app_id . '_' . $this->order_id);
        $queue->while();
        // 开启事务
        $this->startTrans();
        try {
            // 当前订单已结账，无法撤销
            foreach ($this->subOrder as $item) {
                if (self::checkOrderPaid($item->order_id)) {
                    $this->error = '当前订单已结账，无法撤销';
                    $queue->release();
                    return false;
                }
            }

            // 更新所有订单商品sub_order_id = 0
            OrderProductModel::where('order_id', $this->order_id)->update(['sub_order_id' => 0]);
            // 更新所有自助餐订单sub_order_id = 0
            OrderBuffetModel::where('order_id', $this->order_id)->update(['sub_order_id' => 0]);
            // 更新所有自助餐顾客sub_order_id = 0
            OrderBuffetCustomerModel::where('order_id', $this->order_id)->update(['sub_order_id' => 0]);
            // 更新所有加钟sub_order_id = 0
            OrderDelayModel::where('order_id', $this->order_id)->update(['sub_order_id' => 0]);

            // 删除所有子单
            foreach ($this->subOrder as $item) {
                /**  @var Order $item */
                $item->force()->delete();
            }

            // 使用会员
            if ($user_id > 0) {
                $this->useMember($user_id);
            }

            // 重载订单价格
            $this->reloadPrice($this->order_id);

            // 添加操作记录
            OrderOperationLog::createLog($this->order_id, OrderOperationLog::ACTION_CANCEL_SPLIT_ORDER, [], '撤销拆单');

            // 提交事务
            $this->commit();
            $queue->release();
            return true;
        } catch (BaseException $e) {
            // 回滚事务
            $this->error = $e->getMessage();
            $queue->release();
            $this->rollback();
            return false;
        }
    }

    /**
     * 获取子单列表
     */
    public static function getSubOrderList($parentId)
    {
        $list = [];
        $subOrderList = self::where('parent_id', $parentId)->order('order_id', 'asc')->select();
        foreach ($subOrderList as $order) {
            $list[] = [
                'order_id' => $order->order_id,     // 子单id
                'order_name' => $order->order_name, // 子单名称
                'pay_price' => $order->pay_price,   // 子单价格
                'discount_change_price' => $order->discount_change_price,   // 子单价格
                'order_disabled' => $order->pay_status['value'] == OrderPayStatusEnum::SUCCESS, // 子单状态：是否禁用
                'request_time' => time()            // 前端用来判断新旧数据
            ];
        }
        return $list;
    }

    /**
     * 检查订单是否已完成
     */
    public static function checkOrderComplete($orderId)
    {
        $mainOrder = self::where('order_id', $orderId)->find();
        return $mainOrder &&
            (
                $mainOrder->order_status['value'] == OrderStatusEnum::CANCELLED ||
                $mainOrder->order_status['value'] == OrderStatusEnum::COMPLETED
            );
    }

    /**
     * 检查订单是否已送厨
     */
    public static function checkOrderSendKitchen($orderId)
    {
        return OrderProductModel::where('order_id', $orderId)->where('is_send_kitchen', 0)->count() == 0;
    }

    /**
     * 处理订单已完成
     *
     * @param $orderId
     * @param $data
     * @return void
     * @throws DataNotFoundException
     * @throws DbException
     * @throws ModelNotFoundException
     */
    public static function handleOrderCompleted($orderId, $data)
    {
        $noPayCount = OrderModel::where('parent_id', $orderId)
            ->where('pay_status', '<>', OrderPayStatusEnum::SUCCESS)
            ->count();
        if ($noPayCount == 0) {
            $mainOrder = OrderModel::detail($orderId);
            $mainOrder->save([
                'pay_status' => OrderPayStatusEnum::SUCCESS,
                'pay_time' => $data['pay_time'],
                'settle_type' => $data['settle_type'],
                'auto_close' => $data['auto_close'],
                'close_time' => $data['close_time'],
                'delivery_status' => 20,
                'delivery_time' => $data['delivery_time'],
                'receipt_status' => 20,
                'receipt_time' => $data['receipt_time'],
                'order_status' => OrderStatusEnum::COMPLETED,
                'is_settled' => 1,
                'cashier_id' => $data['cashier_id'] ?? 0,
            ]);
            // 桌台订单关闭桌台
            if ($mainOrder['table_id'] > 0) {
                // 支付后是否清台
                $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $mainOrder['shop_supplier_id'], $mainOrder['app_id']);
                if ($store['no_clear_table'] == 0) {
                    TableModel::close($mainOrder['table_id']);
                }
                // 清除接单数据
                $orderIds = array_merge([$orderId], $mainOrder->subOrder()->column('order_id'));

                /** @var TakeOrder $takeOrders */
                $takeOrders = TakeOrder::where('order_id', 'in', $orderIds)->where('status', 0)->select();
                foreach ($takeOrders as $takeOrder) {
                    $takeOrder->reject();
                }
            }
        }
    }

    /**
     * 处理订单已完成 - 高峰时间段记入处理（v1.1.1）
     *
     * @param $orderId
     * @return void
     */
    public static function handleOrderCompletedPeakTime($orderId)
    {
        $noPayCount = OrderModel::where('parent_id', $orderId)
            ->where('pay_status', '<>', OrderPayStatusEnum::SUCCESS)
            ->count();
        if ($noPayCount == 0) {
            (new OrderPeakTime)->record('inc', $orderId);
        }
    }

    /**
     * 撤销折扣 v1.1.0
     */
    public function cancelDiscount($data): bool
    {
        // 获取订单详情
        if (isset($data['order_id']) && $data['order_id'] > 0) {
            /** @var Order $detail */
            $detail = self::detail([
                ['order_id', '=', $data['order_id']],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
        } elseif (isset($data['table_id']) && $data['table_id'] > 0) {
            /** @var Order $detail */
            $detail = self::getTableUnderwayOrder($data['table_id']);
        } else {
            $detail = null;
        }

        // 检查订单状态
        if (!$detail) {
            $this->error = '订单不存在或状态异常';
            return false;
        }

        // 检查支付状态
        if ($error = $detail->validateOrderActionableStatus()) {
            $this->error = $error;
            return false;
        }

        $this->startTrans();
        try {
            // 重置折扣相关字段
            $detail->handleCancelDiscount();

            // 重新计算订单价格
            $this->reloadPrice($detail['order_id']);

            // 处理合并订单
            if ($detail['merge_parent_id'] ?? 0) {
                $detail->reloadMasterMergeOrder($detail['merge_parent_id'], $detail['merge_id']);
            }

            // 添加操作日志
            OrderOperationLog::createLog(
                $detail['order_id'],
                OrderOperationLog::ACTION_CANCEL_DISCOUNT,
                [
                    'parent_id' => $detail['parent_id'],         // 拆单主单ID
                    'order_name' => $detail['order_name'],       // 订单名称
                ],
                '撤销优惠折扣'
            );

            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 结账抹零 v1.1.0
     */
    public function checkoutDiscount($data)
    {
        // 抹零类型（1：抹分，2：抹角，5：抹元）
        $checkoutDiscountType = $data['checkout_discount_type'] ?? 0;
        // 获取订单详情
        if (isset($data['order_id']) && $data['order_id'] > 0) {
            $detail = self::detail([
                ['order_id', '=', $data['order_id']],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
        } else {
            $detail = null;
        }

        // 检查订单状态
        if (!$detail) {
            $this->error = '订单不存在或状态异常';
            return false;
        }

        // 未送厨订单不允许结账抹零
        if ($detail['extra_times'] == 0) {
            $this->error = '未送厨订单不允许结账抹零';
            return false;
        }

        // 检查支付状态
        if ($error = $detail->validateOrderActionableStatus('payment')) {
            $this->error = $error;
            return false;
        }

        // 有含手续费的支付方式的订单不允许抹零
        if ($detail['pay_fee_money'] > 0) {
            $this->error = '使用含手续费的支付方式，手动抹零已失效';
            return false;
        }

        $newPayPrice = 0.0; // 初始化新的应付金额
        $checkoutDiffMoney = 0.0; // 初始化结账抹零差额（与pay_price）
        //
        $this->startTrans();
        try {
            // 根据抹零类型计算新的应付金额和抹零差额
            $getlast = (new self())->reloadPrice($detail['order_id']);  // 获取最新
            $payPrice = isset($getlast['pay_price']) ? (float)$getlast['pay_price'] : 0.0; // 原始支付金额
            switch ($checkoutDiscountType) {
                case 0: // 实款实收
                    $newPayPrice = $payPrice;
                    break;
                case 1: // 抹分
                    $newPayPrice = floor($payPrice * 10) / 10; // 抹去分位，保留一位小数
                    break;
                case 2: // 抹角
                    $newPayPrice = (int)$payPrice; // 抹去角位，保留整数部分
                    break;
                case 5: // 抹元，例如268.25->260
                    $newPayPrice = floor($payPrice / 10) * 10; // 抹去元位，保留10的倍数
                    break;
                default:
                    $this->error = '无效的抹零类型';
                    $this->rollback();
                    return false;
            }
            // 新的结账抹零差额（与pay_price）
            $checkoutDiffMoney = floatval(helper::bcsub($payPrice, $newPayPrice));
            //
            $detail->save([
                'checkout_discount_type' => $checkoutDiscountType,
                'checkout_diff_money'    => $checkoutDiffMoney,
            ]);

            // 添加操作日志
            OrderOperationLog::createLog(
                $detail['order_id'],
                OrderOperationLog::ACTION_CHECKOUT_DISCOUNT,
                [
                    'operation' => 'add',
                    'rounding_type' => $checkoutDiscountType,   //抹零类型：0-实款实收 1-抹分 2-抹角 5-抹元
                    'special_discount' => $checkoutDiffMoney,   //抹零差额
                    'parent_id' => $detail['parent_id'],        // 拆单主单ID
                    'order_name' => $detail['order_name'],      // 订单名称
                ],
                '结账抹零'
            );

            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 自动取消结账抹零
     */
    public function cancelCheckoutDiscount($order_id)
    {
        // 获取订单详情
        $detail = self::detail([
            ['order_id', '=', $order_id],
            ['order_status', '=', OrderStatusEnum::NORMAL]
        ]);
        if (!$detail) {
            $this->error = "订单不存在或状态异常";
            return false;
        }
        // 检查支付状态
        if ($error = $detail->validateOrderActionableStatus()) {
            $this->error = $error;
            return false;
        }
        // 检查是否存在结账抹零
        if ($detail['checkout_discount_type'] == 0) {
            return true;
        }
        $this->startTrans();
        try {
            // 重置结账抹零相关
            $updateData = [
                'checkout_discount_type' => 0,   // 抹零类型
                'checkout_diff_money' => 0,      // 抹零差额
            ];

            // 更新订单数据
            if (!$detail->save($updateData)) {
                throw new BaseException(['msg' => '取消结账抹零失败']);
            }
            // 添加操作日志
            OrderOperationLog::createLog(
                $detail['order_id'],
                OrderOperationLog::ACTION_CHECKOUT_DISCOUNT,
                [
                    'operation' => 'cancel',
                    'remark' => '选择含手续费的支付方式',
                    'parent_id' => $detail['parent_id'],         // 拆单主单ID
                    'order_name' => $detail['order_name'],       // 订单名称
                ],
                '取消结账抹零'
            );
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 是否显示按钮 1-显示 0-不显示
     */
    public function getButtonStatus($data)
    {
        $parentId = $data['parent_id'] ?? 0;
        $mergeParentId = $data['merge_parent_id'] ?? 0;
        $isFree = $data['is_free'] ?? 0;
        $refundMoney = $data['refund_money'] ?? 0;
        $payPrice = $data['pay_price'] ?? 0;
        $orderStatusValue = $data['order_status']['value'] ?? 0;
        $subOrderCount = count($data['subOrder'] ?? []);
        //
        $payTime = $data['pay_time'] ?? 0;
        $isSettled = $data['is_settled'] ?? 0;
        $orderCashierId = $data['cashier_id'] ?? 0;
        $nowCashierId = $data['now_cashier_id'] ?? 0;
        $cashierLoginTime = $data['cashier_login_time'] ?? 0;

        // 退款按钮 1-显示 0-不显示
        $isRefundButton = ($mergeParentId == 0 &&
            $isFree == 0 &&
            $refundMoney < $payPrice &&
            $orderStatusValue == 30 &&
            ($subOrderCount == 0 || $parentId > 0)) ? 1 : 0;

        // 取消按钮 1-显示 0-不显示 如果subOrder子单中，有一个子单已结账，则不显示取消按钮
        if ($parentId == 0 && $subOrderCount > 0) {
            $hasSuccessStatus = false;
            foreach ($data['subOrder']->toArray() ?? [] as $item) {
                if ($item['order_status']['value'] == OrderStatusEnum::COMPLETED) {
                    $hasSuccessStatus = true;
                    break;
                }
            }
            $isCancelButton = ($mergeParentId == 0 && $orderStatusValue == 10 && !$hasSuccessStatus) ? 1 : 0;
        } else {
            $isCancelButton = ($mergeParentId == 0 && $orderStatusValue == 10 && $parentId == 0) ? 1 : 0;
        }

        // 反结账按钮 1-显示 0-不显示
        $isReverseSettleButton = ($nowCashierId == $orderCashierId &&
            $payTime > $cashierLoginTime &&
            $isSettled == 1 &&
            $refundMoney == 0 &&
            $orderStatusValue == 30 &&
            $parentId == 0) ? 1 : 0;

        return [$isRefundButton, $isCancelButton, $isReverseSettleButton];
    }

    /**
     * 获取使用优惠的会员列表
     */
    public function getUsedMemberList($orderId)
    {
        /** @var OrderModel $order */
        $order = OrderModel::detail($orderId, [
            'user',
            'subOrder' => function ($query) {
                $query->with(['user']);
            },
        ]);
        if (!$order) {
            $this->error = '订单不存在';
            return false;
        }
        // 当前订单已结账，无法撤销
        foreach ($order->subOrder as $item) {
            if (self::checkOrderPaid($item->order_id)) {
                $this->error = '当前订单已结账，无法撤销';
                return false;
            }
        }
        $tmp = [];
        $memberList = [];
        $isDiscountExpire = false;

        // 子订单用户
        foreach ($order->subOrder as $subOrder) {
            if ($subOrder->user && !in_array($subOrder->user_id, $tmp)) {
                $tmp[] = $subOrder->user_id;
                $memberList[] = [
                    'user_id' => $subOrder->user_id,
                    'nickName' => $subOrder->user->nickName,
                    'mobile' => $subOrder->user->mobile,
                ];
            }
            if (!$isDiscountExpire && $subOrder->is_change_price > 0) {
                $isDiscountExpire = true;
            }
        }

        return [
            'memberList' => $memberList,
            'isDiscountExpire' => $isDiscountExpire,
        ];
    }

    /**
     * 创建拆单操作记录
     */
    public function createSplitOrderLog($mainOrderId, $subOrderList)
    {
        if (empty($subOrderList)) {
            return;
        }
        $subOrderArr = [];
        foreach ($subOrderList as $subOrder) {
            // 子单订单金额
            $subOrderId = $subOrder['order_id'] ?? 0;
            $orderName = $subOrder['order_name'] ?? '';
            $shopSupplierId = $subOrder['shop_supplier_id'] ?? 0;
            $appId = $subOrder['app_id'] ?? 0;
            $tableId = $subOrder['table_id'] ?? 0;
            //
            $cache_price = Cache::get('order_pay_price_' . $subOrderId . '_' . $appId);
            if ($cache_price == null) {
                $subOrderPrice =  (new \app\cashier\model\order\Cart())->getOrderCartDetail([], $tableId, $subOrderId);
                $cache_price =  $subOrderPrice['sumInfo']['total_pay_price'] ?? 0;
            }
            //
            $subOrderArr[] = [
                'order_id' => $subOrderId,
                'order_name' => $orderName,
                'pay_price' => $cache_price, // 订单金额
            ];
        }
        //
        OrderOperationLog::createLog($mainOrderId, OrderOperationLog::ACTION_SPLIT_ORDER, ['split_order' => $subOrderArr], '拆单操作');
    }

    /**
     * 处理撤销优惠折扣
     * @param array $appendData 附加数据
     */
    public function handleCancelDiscount($appendData = [])
    {
        // 重置折扣相关字段
        $updateData = [
            'discount_ratio' => 0,           // 折扣率
            'discount_money' => 0,           // 折扣金额
            'discount_change_price' => 0,    // 改价金额
            'small_discount_type' => 0,      // 抹零类型
            'small_diff_money' => 0,         // 抹零差额
            'small_auto' => 0,               // 是否自动抹零
            'is_change_price' => 0,          // 是否改价
        ];
        if ($appendData) {
            $updateData = array_merge($updateData, $appendData);
        }

        // 更新订单数据
        if (!$this->save($updateData)) {
            throw new BaseException(['msg' => '撤销优惠折扣失败']);
        }
    }

    /**
     * 检查订单是否已结账
     */
    public static function checkOrderPaid($orderId)
    {
        $order = self::detail($orderId, []);
        $payTypeCount = OrderPayType::where('order_id', $orderId)->where('pay_status', 1)->count();
        if ($order->getData('pay_status') == OrderPayStatusEnum::SUCCESS || $payTypeCount > 0) {
            return true;
        }
        return false;
    }
}
