<?php

namespace app\common\repositories;

use app\common\library\helper;
use app\common\model\order\Order;
use app\common\model\user\PointsLog;
use app\common\model\user\BalanceLog;
use app\common\enum\order\OrderStatusEnum;
use app\common\model\order\OrderAbnormalLog;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\OrderOperationLog;
use app\common\model\order\UserRechargeOrder;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model\order\UserRechargeOrderOperationLog;
use help\HttpHelp;

/**
 * 订单业务数据仓库
 */
class OrderBusinessDataRepository
{
    // 用于链调用
    protected $model;
    protected $baseModel;

    // 开始时间 - 结束时间
    protected $startTime = 0;
    protected $endTime = 0;
    protected $shopSupplierId = 0;
    protected $shopCashierId = 0;

    // 来源
    protected $source = 0;


    // 构造器
    public function __construct(Order $model, array $params, $timeType = 0)
    {
        $params['time'] = $params['time'] ?? $params['date'] ?? [];
        //
        $this->shopSupplierId = $shopSupplierId = $params['shop_supplier_id'] ?? 0;
        $this->shopCashierId = $shopCashierId = $params['cashier_id'] ?? 0;
        $timeType = $params['time_type'] ?? $timeType;
        //
        $startTime = 0;
        $endTime = 0;
        //查询类型查询时间
        switch ($timeType) {
            case '1': //今天
                $startTime = strtotime(date('Y-m-d'));
                $endTime = $startTime + 86399;
                break;
            case '2': //昨天
                $startTime = strtotime("-1 days", strtotime(date('Y-m-d')));
                $endTime = $startTime + 86399;
                break;
            case '3': //本周
                $startTime = strtotime('monday this week'); // 本周第一天的时间戳
                $endTime = strtotime('sunday this week +23 hours +59 minutes +59 seconds'); // 本周最后一天的时间戳
                break;
            case '4': //本月
                $year = date('Y');
                $month = date('m');
                $startTime = strtotime(date('Y-m-01 00:00:00', strtotime("$year-$month")));
                $endTime = strtotime(date('Y-m-t 23:59:59', strtotime("$year-$month")));
                break;
        }

        // 按指定时间查询
        if (isset($params['time']) && $params['time'] && ($params['time'][0] ?? '') && ($params['time'][1] ?? '')) {
            $startTime = strtotime($params['time'][0]);
            if (strpos($params['time'][1], ':') !== false) {
                $endTime = strtotime($params['time'][1]);
            } else {
                $endTime = strtotime($params['time'][1]) + 86399;
            }
        }

        //
        $this->startTime = $startTime;
        $this->endTime = $endTime;
        //
        $this->model = $model;
        $this->baseModel = $this->model->alias('a')
            ->where('a.pay_status', '=', OrderPayStatusEnum::SUCCESS)
            ->where('a.order_status', '=', OrderStatusEnum::COMPLETED)
            ->where('a.delete_time', '=', 0)
            ->when($shopSupplierId, function ($q) use ($shopSupplierId) {
                $q->where('a.shop_supplier_id', '=', $shopSupplierId);
            })
            ->when($shopCashierId, function ($q) use ($shopCashierId) {
                $q->where('a.cashier_id', '=', $shopCashierId);
            })
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('a.pay_time', 'between', [$startTime, $endTime]);
            });
    }

    /**
     * 获取基础模型
     * @return Order
     */
    public function getBaseModel()
    {
        return $this->baseModel->clone();
    }

    /**
     * 获取时间
     * @return array
     */
    public function getTimes()
    {
        return [
            0 => $this->startTime,
            1 => $this->endTime,
        ];
    }

    /**
     * 格式化数据
     * @return array
     */
    private function formattedData($data)
    {
        // 用户充值相关
        $endTime = $this->endTime;
        $startTime = $this->startTime;
        $shopCashierId = $this->shopCashierId;
        $shopSupplierId =  $this->shopSupplierId;
        // 充值 - 支付手续费
        $balancePayFee = UserRechargeOrder::alias('a')
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('a.pay_time', 'between', [$startTime, $endTime]);
            })
            ->when($shopSupplierId, function ($q) use ($shopSupplierId) {
                $q->where('a.shop_supplier_id', '=', $shopSupplierId);
            })
            ->when($shopCashierId, function ($q) use ($shopCashierId) {
                $q->where('a.cashier_id', '=', $shopCashierId);
            })
            ->where('a.order_status', 1)
            ->value('sum(a.pay_fee)') ?: 0;
        // 支付手续费 = 订单支付手续费 + 充值支付手续费
        if ($balancePayFee) {
            $data['pay_fee_money'] = helper::bcadd($data['pay_fee_money'], $balancePayFee);
        }

        // 充值 - 金额
        $balanceLog = BalanceLog::alias('a')
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('a.create_time', 'between', [$startTime, $endTime]);
            })
            ->where(function ($q) {
                $q->where(function ($q) {
                    $q->where('a.version', BalanceLog::VERSION);
                    $q->where('a.scene', BalanceLogSceneEnum::RECHARGE);
                });
                $q->whereOr(function ($q) {
                    $q->where('a.version', BalanceLog::VERSION_107);
                    $q->where('a.scene', BalanceLogSceneEnum::ADMIN);
                });
                // 充值反结账
                $q->whereOr(function ($q) {
                    $q->where('a.version', BalanceLog::VERSION);
                    $q->where('a.scene', BalanceLogSceneEnum::RECHARGE_REVERSE);
                });
                // 充值退款
                $q->whereOr(function ($q) {
                    $q->where('a.version', BalanceLog::VERSION);
                    $q->where('a.scene', BalanceLogSceneEnum::RECHARGE_REFUND);
                });
            })
            ->field("
                SUM(a.gift_money) AS gift_money,
                SUM(a.money - a.gift_money) AS recharge_amount,
                SUM(
                    CASE
                        WHEN a.scene = '" . BalanceLogSceneEnum::RECHARGE_REFUND . "' THEN ABS(a.money)
                        ELSE 0
                    END
                ) AS recharge_refund_total
            ")
            ->find();
        // 充值 - 赠送积分
        $data['gift_points'] = PointsLog::where('scene', PointsLogSceneEnum::RECHARGE_GIVE)
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('create_time', 'between', [$startTime, $endTime]);
            })
            ->sum('value');
        $data['gift_money'] = $balanceLog['gift_money'] ?? 0;
        $data['recharge_amount'] = $balanceLog['recharge_amount'] ?? 0;
        $data['recharge_refund_total'] = $balanceLog['recharge_refund_total'] ?? 0;

        // 未结账数据
        $data = array_merge(
            $data,
            $this->model->alias('a')
                ->where('a.pay_status', '=', OrderPayStatusEnum::PENDING)
                ->where('a.order_status', '=', OrderStatusEnum::NORMAL)
                ->where('a.delete_time', '=', 0)
                ->where('a.parent_id', '=', 0)
                ->when($shopSupplierId, function ($q) use ($shopSupplierId) {
                    $q->where('a.shop_supplier_id', '=', $shopSupplierId);
                })
                ->when($shopCashierId, function ($q) use ($shopCashierId) {
                    $q->where('a.cashier_id', '=', $shopCashierId);
                })
                ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                    $q->where('a.create_time', 'between', [$startTime, $endTime]);
                })
                ->field("count(*) as not_settled_total_order_num")
                ->field("ifnull(sum(pay_price), 0) as not_settled_total_price")
                ->find()->append([])?->toArray()
        );

        // 实收金额：指订单原价扣除优惠折扣、会员折扣后的金额，包含支付手续费 + 收银机会员充值金额（不包括会员余额消费金额）
        $data['received_price'] = helper::bcadd(helper::bcadd($data['received_price'], $data['recharge_amount']), $balancePayFee);
        // 总销售额 = 原商品金额+页面的服务费+页面的支付手续费+页面税费（v1.1.1版本）- 产品需求
        $data['receivable_price'] = helper::bcadd($data['receivable_price'], $data['service_money']);
        // 总销售额加上手续费
        $data['receivable_price'] = helper::bcadd($data['receivable_price'], $balancePayFee);
        // 优惠占比 =  优惠 / 总销售额
        $data['discount_ratio'] = ($data['receivable_price'] > 0 ? (round($data['discount_money'] / $data['receivable_price'], 4) * 100) : 0) . '%';
        // 营业收入（v1.0.7） = 实收金额 - 税费 + 会员余额消费金额 + 会员余额消费
        $data['business_price'] = helper::bcadd(helper::bcsub($data['received_price'], $data['consumption_tax_money']), $data['received_balance_price']);
        // 退款金额（v1.1.0） = 退款金额 + 充值退款金额
        $data['refund_money'] = helper::bcadd($data['refund_money'], $data['recharge_refund_total']);

        // 格式化数据
        if (isset($data['min_order_price']) && $data['min_order_price'] < 0) {
            $data['min_order_price'] = 0;
        }
        if (isset($data['table_min_order_price']) && $data['table_min_order_price'] > $data['table_max_order_price']) {
            $data['table_min_order_price'] = $data['table_max_order_price'];
        }
        if (isset($data['table_min_order_price']) && $data['table_min_order_price'] < 0) {
            $data['table_min_order_price'] = 0;
        }
        if (isset($data['cashier_min_order_price']) && $data['cashier_min_order_price'] > $data['cashier_max_order_price']) {
            $data['cashier_min_order_price'] = $data['cashier_max_order_price'];
        }
        if (isset($data['cashier_min_order_price']) && $data['cashier_min_order_price'] < 0) {
            $data['cashier_min_order_price'] = 0;
        }
        foreach ($data as $key => $value) {
            if ($key == 'pay_type_date') {
                $array = explode(",", $value ?? ''); // 将字符串拆分为数组
                $uniqueArray = array_unique($array); // 对数组进行去重
                $data[$key] = implode(",", $uniqueArray);
            } else if (is_numeric($value)) {
                $data[$key] = floatval($value);
            }
        }

        //
        return $data;
    }

    /**
     * 获取营业数据
     * @return self
     */
    public function setSource($source)
    {
        $this->source = $source;
        return $this;
    }
    public function getBusinessData($type = 1)
    {
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/home', [
            'query_start_time' => $this->startTime,
            'query_end_time' => $this->endTime,
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            return [];
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            return [];
        }

        return $res['data'];
    }

    /**
     * 获取百分比列表
     * @return array
     */
    public function getPercentageList()
    {
        $prefix = env('DB_PREFIX');
        //
        $percentageList = $this->baseModel->clone()
            ->leftJoin("
                (
                    select order_id, tax_rate, total_pay_price, refund_money, consumption_tax, product_service_fee, product_service_consumption_tax , refund_consumption_tax from {$prefix}order_product where product_id > 0 and is_return = 0
                    UNION ALL
                    select order_id, tax_rate, total_pay_price, refund_money, consumption_tax, product_service_fee, product_service_consumption_tax , refund_consumption_tax from {$prefix}order_buffet_customer
                    UNION ALL
                    select order_id, 0 as tax_rate, total_price as total_pay_price, refund_money, 0 consumption_tax, 0 product_service_fee, 0 product_service_consumption_tax, 0 refund_consumption_tax from {$prefix}order_delay
                ) rp
            ", 'a.order_id = rp.order_id and a.is_merge = 0')
            ->field("
                ifnull(
                    sum(
                        if( a.consumption_tax_type = 2,
                            rp.total_pay_price + round(rp.consumption_tax, 2) - if(rp.refund_money > 0, rp.refund_money - rp.product_service_fee, 0),
                            rp.total_pay_price - if(rp.refund_money > 0, (rp.refund_money - rp.product_service_fee - rp.product_service_consumption_tax), 0)
                        )
                    )
            , 0) total_price")
            ->field("rp.tax_rate, round(ifnull(sum(round(rp.consumption_tax, 2) - rp.refund_consumption_tax), 0), 2) as consumption_tax")
            ->group("rp.tax_rate")
            ->where('a.is_merge', 0)
            ->select()
            ->append([])?->toArray();
        foreach ($percentageList as $key => &$data) {
            $percentageList[$key]['tax_rate'] = floatval($data['tax_rate']);
            $percentageList[$key]['total_price'] = floatval($data['total_price']);
            $percentageList[$key]['consumption_tax'] = floatval($data['consumption_tax']);
        }
        //
        return $percentageList;
    }

    /**
     * 获取收入列表
     * @return array
     */
    public function getIncomesList()
    {
        $shopSupplierId =  $this->shopSupplierId;
        $shopCashierId = $this->shopCashierId;
        $startTime = $this->startTime;
        $endTime = $this->endTime;
        $prefix = env('DB_PREFIX');

        // 商品订单
        $productOrders = $this->baseModel->clone()
            ->leftJoin('order_pay_type opt', 'a.order_id = opt.order_id')
            ->leftJoin("
                (
                    select rp.merge_parent_id, count(*) as count_num
                    from {$prefix}order rp
                    group by merge_parent_id
                ) sub
            ", 'a.order_id = sub.merge_parent_id')
            ->leftJoin("
                (
                    select ord.order_id, ord.value, ifnull(sum(ord.refund_money), 0) as refund_money
                    from {$prefix}order_refund_destination ord
                    where ord.status = 1
                    group by ord.order_id, ord.value
                ) ord
            ", 'a.order_id = ord.order_id AND opt.value = ord.value')
            ->whereNotNull('opt.id')
            ->group("opt.value")
            // 支付类型
            ->field("opt.value")
            // 订单数量
            ->field("if(sub.count_num > 0, sum(ifnull(sub.count_num, 0)), count(distinct a.order_id))  as order_num")
            // 当前收入（不包含找零）
            ->field("sum(if(opt.value = 40, opt.price - a.change_due - ifnull(ord.refund_money, 0), opt.price - ifnull(ord.refund_money, 0))) as price")
            // 当前收入（不包含找零）(已不包含退款)
            ->field("sum(if(opt.value = 40, opt.price - a.change_due - ifnull(ord.refund_money, 0), opt.price - ifnull(ord.refund_money, 0))) as refund_included_price")
            // 当前支付方式退款成功总金额
            ->field("ifnull(ord.refund_money, 0) as refund_destination_money")
            // 拆单后 只获取主单已经支付的
            ->where('a.parent_id', '=', 0)
            ->select()?->append([])->toArray() ?? [];

        // 用户充值订单
        $userRechargeOrders = UserRechargeOrder::alias('a')
            ->leftJoin('user_recharge_order_pay_type opt', 'a.id = opt.order_id')
            ->leftJoin("
                (
                    select urord.order_id, urord.value, ifnull(sum(urord.refund_money), 0) as refund_money
                    from {$prefix}user_recharge_order_refund_destination urord
                    where urord.status = 1
                    group by urord.order_id, urord.value
                ) urord
            ", 'a.id = urord.order_id AND opt.value = urord.value')
            ->where('a.order_status', '=', 1)
            ->when($shopSupplierId, function ($q) use ($shopSupplierId) {
                $q->where('a.shop_supplier_id', '=', $shopSupplierId);
            })
            ->when($shopCashierId, function ($q) use ($shopCashierId) {
                $q->where('a.cashier_id', '=', $shopCashierId);
            })
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('a.pay_time', 'between', [$startTime, $endTime]);
            })
            ->whereNotNull('opt.id')
            ->group("opt.value")
            ->field("opt.value, sum(1) order_num")
            // 当前收入（不包含找零）
            ->field("sum(if(opt.value = 40, opt.price - a.change_due - ifnull(urord.refund_money, 0), opt.price - ifnull(urord.refund_money, 0))) as price")
            // 当前收入（不包含找零）(已不包含退款)
            ->field("sum(if(opt.value = 40, opt.price - a.change_due - ifnull(urord.refund_money, 0), opt.price - ifnull(urord.refund_money, 0))) as refund_included_price")
            // 当前支付方式退款成功总金额
            ->field("ifnull(urord.refund_money, 0) as refund_destination_money")
            //
            ->select()?->append([])->toArray() ?? [];

        // 合并
        $results = array_merge($productOrders, $userRechargeOrders);

        // 格式化
        $incomes = [];
        foreach ($results as &$value) {
            if ($value['price'] > 0 || $value['value'] == -1 || $value['refund_destination_money'] > 0) {
                $incomes[$value['value']] = [
                    'pay_type' => $value['value'],
                    'pay_type_name' => (new Order)->getPayTypeRemark($value['value']),
                    'pay_type_way' => (new Order)->getPayTypeName($value['value']),
                    'price' => $value['price'] + ($incomes[$value['value']]['price'] ?? 0),
                    'order_num' => $value['order_num'] + ($incomes[$value['value']]['order_num'] ?? 0),
                    'refund_included_price' => $value['refund_included_price'] + ($incomes[$value['value']]['refund_included_price'] ?? 0),
                ];
            }
        }

        // 结果
        return array_values($incomes);
    }

    /**
     * 获取区域数据
     * @return array
     */
    public function getRegionData($dates = [])
    {
        $startTime = $this->startTime;
        $endTime = $this->endTime;
        if ($dates) {
            $startTime = $dates[0];
            $endTime = $dates[1];
        }

        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/area', [
            'query_start_time' => $startTime,
            'query_end_time' => $endTime,
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json; charset=utf-8',
        ]);
        if (!$res) {
            return [];
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            return [];
        }
        $data = $res['data'];
        $regionData = [];
        foreach ($data['areas'] as $item) {
            $regionData[] = [
                'area_name' => $item['name'],
                'sales_price' => $item['total_sales'],
                'business_price' => $item['total_received_price'],
                'product_num' => $item['total_product_num'],
            ];
        }
        return $regionData;
    }

    /**
     * 获取商品销售数据
     * @return array
     */
    public function getProductData($params = [])
    {
        $areaId = (!isset($params['area_id']) || !$params['area_id']) ? 0 : $params['area_id'];
        $categoryId = (!isset($params['category_id']) || !$params['category_id']) ? 0 : $params['category_id'];
        $productName = (!isset($params['product_name']) || !$params['product_name']) ? '' : $params['product_name'];
        $sortField = (!isset($params['sort_field']) || !$params['sort_field']) ? 'complete_time' : $params['sort_field'];
        $sortType = (!isset($params['sort_type']) || !$params['sort_type']) ? 'desc' : $params['sort_type'];
        $token = (!isset($params['token']) || !$params['token']) ? '' : $params['token'];
        $language = (!isset($params['language']) || !$params['language']) ? 'zh-CN' : $params['language'];

        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/product_sales', [
            'product_name' => $productName,
            'area_uuid' => $areaId,
            'category_uuid' => $categoryId,
            'query_start_time' => $this->startTime,
            'query_end_time' => $this->endTime,
            'sort_type' => ['complete_time' => 0, 'product_num' => 1, 'sales_price' => 2][$sortField ?: 'complete_time'],
            'sort_direction' => ['asc' => 1, 'desc' => 2][$sortType],
            'page_no' => $params['page'] ?? 1,
            'page_size' => $params['list_rows'] ?? 10,
        ], [
            'Authorization: Bearer ' . (request()->header('token') ?: $token),
            'Accept-Language: ' . (request()->header('language') ?: $language),
            'Content-Type: application/json; charset=utf-8',
        ]);

        if (!$res) {
            return [];
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            return [];
        }
        $data = $res['data'];
        $meta = $data['meta'];
        $list = [];
        foreach ($data['list'] as $item) {
            $list[] = [
                'product_name_text' => $item['product_name'],
                'category_name_text' => $item['category_name'],
                'product_num' => $item['sales_num'],
                'sales_price' => $item['original_sales_price'],
                'free_product_num' => $item['give_product_num'],
                'received_price' => $item['sales_price'],
                'business_price' => $item['total_pay_price'],
            ];
        }

        return [
            'data' => $list,
            'current_page' => $meta['page_no'],
            'per_page' => $meta['page_size'],
            'total' => $meta['total'],
            'last_page' => ceil($meta['total'] / $meta['page_size']),
        ];
    }

    /**
     * 获取异常信息
     */
    public function getAbnormalData($params = [])
    {
        $shopSupplierId =  $this->shopSupplierId;
        $shopCashierId = $this->shopCashierId;
        $startTime = $this->startTime;
        $endTime = $this->endTime;
        $dutyNo = $params['duty_no'] ?? ''; // 当班编号
        // 异常日志
        $orderAbnormalLog = (new OrderAbnormalLog)
            // 退菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次，如果有取消退菜，则要减去取消退菜次数）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_REFUND_PRODUCT . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS refund_product_times
            ")
            // 取消退菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_CANCEL_REFUND_PRODUCT . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS cancel_refund_times
            ")
            // 赠菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_PRODUCT_FREE . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS product_free_times
            ")
            // 取消赠菜次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_CANCEL_PRODUCT_FREE . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS cancel_product_free_times
            ")
            // 退款次数（操作了几次，数量就是几，不是按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_REFUND . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS refund_times
            ")
            // 转菜次数（操作了几次，数量就是几，不是按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_PRODUCT_MOVE . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS product_move_times
            ")
            // 单品改价次数（操作了几次，数量就是几，按照订单来，对一个商品反复操作，记录为1次）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_CHANGE_PRICE . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS change_price_times
            ")
            // 整单改价次数（操作了几次，数量就是几，按照订单来）
            ->field("
            COUNT(DISTINCT CASE
                WHEN action = '" . OrderOperationLog::ACTION_DISCOUNT . "'
                    AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    AND sub_action = '1'
                THEN order_id
            END) AS change_order_price_times
            ")
            // 整单折扣次数（操作了几次，数量就是几，按照订单来）
            ->field("
            COUNT(DISTINCT CASE
                WHEN action = '" . OrderOperationLog::ACTION_DISCOUNT . "'
                    AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    AND sub_action = '2'
                THEN order_id
            END) AS discount_order_times
            ")
            // 整单优惠折扣抹零次数（操作了几次，数量就是几，按照订单来，包含结账抹零）
            ->field("
            COUNT(DISTINCT CASE
                WHEN action = '" . OrderOperationLog::ACTION_DISCOUNT . "'
                    AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    AND sub_action = '3'
                THEN order_id
            END) AS round_order_times
            ")
            // 免单次数（操作了几次，数量就是几，按照订单来）
            ->field("
            COUNT(DISTINCT CASE
                WHEN action = '" . OrderOperationLog::ACTION_DISCOUNT . "'
                    AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    AND sub_action = '4'
                THEN order_id
            END) AS free_order_times
            ")
            // 反结账次数（操作了几次，数量就是几，不是按照订单来）
            ->field("
             COUNT(CASE
                 WHEN action = '" . OrderOperationLog::ACTION_REVERSE_SETTLE . "'
                     AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                 THEN 1
             END) AS reverse_settle_times
         ")
            // 撤销优惠折扣抹零次数（操作了几次，数量就是几，按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_CANCEL_DISCOUNT . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS round_order_cancel_times
            ")
            // 结账抹零次数（操作了几次，数量就是几，按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . OrderOperationLog::ACTION_CHECKOUT_DISCOUNT . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_ORDER . "'
                    THEN 1
                END) AS checkout_round_order_times
            ")
            // 充值 - 退款次数（操作了几次，数量就是几，不是按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . UserRechargeOrderOperationLog::ACTION_REFUND . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_RECHARGE . "'
                    THEN 1
                END) AS recharge_refund_times
            ")
            // 充值 - 反结账次数（操作了几次，数量就是几，不是按照订单来）
            ->field("
                COUNT(CASE
                    WHEN action = '" . UserRechargeOrderOperationLog::ACTION_REVERSE_SETTLE . "'
                        AND source = '" . OrderAbnormalLog::SOURCE_RECHARGE . "'
                    THEN 1
                END) AS recharge_reverse_settle_times
            ")
            ->when($shopCashierId, function ($q) use ($shopCashierId) {
                $q->where('shop_user_id', '=', $shopCashierId);
            })
            ->when($dutyNo && $dutyNo != '', function ($q) use ($dutyNo) {
                $q->where('duty_no', $dutyNo);
            })
            ->when($startTime && $endTime && $dutyNo == '', function ($q) use ($startTime, $endTime) {
                $q->where('create_time', 'BETWEEN', [$startTime, $endTime]);
            })
            ->select()?->append([])->toArray()[0] ?: [];
        // 退款次数 注：需要结账后操作退款的次数数据，整单退款/部分退款几次就是几次（用餐订单/充值订单）
        $orderAbnormalLog['refund_times'] = max(0, ($orderAbnormalLog['refund_times'] ?? 0) + ($orderAbnormalLog['recharge_refund_times'] ?? 0));
        // 反结账 注：需要结账后，操作反结账的次数数据（用餐订单/充值订单）
        $orderAbnormalLog['reverse_settle_times'] = max(0, ($orderAbnormalLog['reverse_settle_times'] ?? 0) + ($orderAbnormalLog['recharge_reverse_settle_times'] ?? 0));
        // 抹零次数 = 整单优惠折扣抹零次数 + 结账抹零次数（补充）
        $orderAbnormalLog['round_order_times'] = max(0, ($orderAbnormalLog['round_order_times'] ?? 0) + ($orderAbnormalLog['checkout_round_order_times'] ?? 0));
        return $orderAbnormalLog;
    }
}
