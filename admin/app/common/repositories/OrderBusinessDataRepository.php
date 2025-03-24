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
            ->where('a.is_delete', '=', 0)
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
                ->where('a.is_delete', '=', 0)
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
        $prefix = env('DB_PREFIX');

        // 总营业数据
        $model = $this->baseModel->clone()
            // 合单父级
            ->leftJoin("order merge_parent_order", 'a.merge_parent_id = merge_parent_order.order_id and a.merge_parent_id > 0')
            // 兄弟订单的数量
            ->leftJoin("
                (
                    select
                        opt.merge_parent_id, count(1) as order_count
                    from {$prefix}order opt
                    group by opt.merge_parent_id
                ) o_merge_order
            ", 'a.merge_parent_id = o_merge_order.merge_parent_id and a.merge_parent_id > 0')
            // 子订单的汇总
            ->leftJoin("
                (
                    select
                        sub.parent_id,

                        ifnull(sum(
                            case
                                when sub.is_free=0 || sub.is_free=1 then
                                    if(order_refund_type.mode > 0, 0, sub.service_money)
                                else 0
                            end
                        ), 0) as sub_service_money,

                        ifnull(sum(
                            case
                                when sub.is_free=0 then
                                    sub.total_product_price
                                    + sub.total_product_service_consumption_tax
                                    + if( sub.consumption_tax_type = 2,
                                        sub.total_product_consumption_tax
                                    , 0)
                                    - sub.refund_consumption_tax
                                    + sub.pay_fee_money
                                    - ifnull(rp.free_not_include_product_price, 0)
                                when sub.is_free=1 then
                                    sub.total_product_price
                                    + sub.total_product_service_consumption_tax
                                    + if( sub.consumption_tax_type = 2,
                                        sub.total_product_consumption_tax
                                    , 0)
                                    - sub.refund_consumption_tax
                                    + sub.pay_fee_money
                                else 0
                            end
                        ), 0) as sub_receivable_price,

                        ifnull(sum(
                            case
                                when sub.is_free=0 then sub.discount_money - ifnull(rp.free_not_include_product_discount_money, 0) + sub.checkout_diff_money
                                when sub.is_free=1 then sub.discount_money
                                else 0
                            end
                        ), 0) as sub_discount_money,

                        ifnull(sum(
                            case
                                when sub.is_free=0 then sub.consumption_tax_money - sub.refund_consumption_tax
                                when sub.is_free=1 then sub.consumption_tax_money - sub.refund_consumption_tax
                                else 0
                            end
                        ), 0) as sub_consumption_tax_money

                    from {$prefix}order sub
                    left join (
                        select
                            d.order_id, sum(if(d.refund_type=1, 1, 0)) as mode
                        from {$prefix}order_refund d
                        group by d.order_id
                    ) order_refund_type on order_refund_type.order_id = sub.order_id
                    left join (
                        select
                            p.sub_order_id,
                            sum(if(p.is_free = 2, p.product_discount_money, 0)) as free_not_include_product_discount_money,
                            sum(if(p.is_free = 2, p.total_product_price, 0)) as free_not_include_product_price
                        from {$prefix}order_product p
                        where p.product_id > 0 and p.is_return = 0
                        group by p.sub_order_id
                    ) rp on rp.sub_order_id = sub.order_id
                group by sub.parent_id
                ) sub_order
            ", 'sub_order.parent_id = a.order_id')
            // 产品
            ->leftJoin("
                (
                    select order_id,
                        sum(tax_rate) as tax_rate,
                        sum(total_pay_price) as total_pay_price,
                        sum(refund_money) as refund_money,
                        sum(consumption_tax) as consumption_tax,
                        sum(refund_consumption_tax) as refund_consumption_tax,
                        sum(total_num) as total_num,
                        sum(product_service_fee / total_num * refund_num) as refund_product_service_fee,
                        sum(total_consumption_tax_pay_price) as total_consumption_tax_pay_price,
                        sum(if(is_free > 0, total_num, 0)) as free_num,
                        sum(if(is_free > 0, total_product_price, 0)) as free_price,
                        sum(if(is_free = 2, total_product_price, 0)) as free_not_include_product_price,
                        sum(no_free_product_service_fee) as free_not_include_product_service_fee,
                        sum(if(is_free = 2, product_discount_money, 0)) as free_not_include_product_discount_money,
                        sum(no_free_product_consumption_tax) as free_not_include_product_consumption_tax,
                        round(sum(
                            if (tax_calc_type = 1,
                                total_product_price - product_original_consumption_tax,
                                total_product_price
                            )
                        ), 2) as not_tax_total_product_price
                    from  (
                        select
                            order_id,
                            is_free,
                            tax_rate,
                            product_discount_money,
                            product_service_fee,
                            (no_free_product_consumption_tax + no_free_product_service_consumption_tax) as no_free_product_consumption_tax,
                            no_free_product_service_fee,
                            total_product_price,
                            total_pay_price,
                            refund_money,
                            consumption_tax,
                            refund_consumption_tax,
                            total_num,
                            refund_num,
                            if(tax_calc_type=2, total_pay_price + consumption_tax, total_pay_price) as total_consumption_tax_pay_price,
                            tax_calc_type,
                            product_original_consumption_tax
                        from {$prefix}order_product where product_id > 0 and is_return = 0
                        UNION ALL
                        select order_id,
                            0 as is_free,
                            tax_rate,
                            0 as no_free_product_consumption_tax,
                            0 as product_discount_money,
                            0 as product_service_fee,
                            0 as no_free_product_service_fee,
                            0 as total_product_price,
                            total_pay_price,
                            refund_money,
                            consumption_tax,
                            refund_consumption_tax,
                            num as total_num,
                            0 as refund_num,
                            if(tax_calc_type=2, total_pay_price + consumption_tax, total_pay_price) as total_consumption_tax_pay_price,
                            tax_calc_type,
                            product_original_consumption_tax
                        from {$prefix}order_buffet_customer
                        UNION ALL
                        select
                            order_id,
                            0 as is_free,
                            0 as tax_rate,
                            0 as no_free_product_consumption_tax,
                            0 as product_discount_money,
                            0 as product_service_fee,
                            0 as no_free_product_service_fee,
                            0 as total_product_price,
                            total_price as total_pay_price,
                            refund_money,
                            0 as consumption_tax,
                            0 as refund_consumption_tax,
                            0 as total_num,
                            0 as refund_num,
                            total_price as total_consumption_tax_pay_price,
                            0 as tax_calc_type,
                            0 as product_original_consumption_tax
                        from {$prefix}order_delay
                    ) rp
                    group by order_id
                ) rp
            ", 'a.order_id = rp.order_id and a.is_merge = 0')
            // 付款方式 - 余额收入
            ->leftJoin("
                (
                    select
                        opt.order_id,
                        sum(price) as balance_price,
                        sum(price - ifnull(order_refund_destination.refund_money, 0)) as balance_price_not_exist_refund
                    from {$prefix}order_pay_type opt
                    left join (
                        select
                            d.order_id, sum(refund_money) as refund_money
                        from {$prefix}order_refund_destination d
                        where d.value=10 and d.status=1
                        group by d.order_id
                    ) order_refund_destination on order_refund_destination.order_id = opt.order_id
                    where opt.value=10
                    group by opt.order_id
                ) o_pay_balance
            ", 'o_pay_balance.order_id = a.order_id')
            // 退款方式 - 是否整单退款, order_refund_type.mode > 0 就是整单退款
            ->leftJoin("
                (
                    select
                        d.order_id,
                        sum(if(d.refund_type=1, 1, 0)) as mode
                    from {$prefix}order_refund d
                    group by d.order_id
                ) order_refund_type
            ", 'order_refund_type.order_id = a.order_id');
        // 总销售额 = 应收 = 原商品金额 + 服务费（实收）+ 支付手续费（实收）+ 税费（实收）注：已含税的商品不重复加税 -
        $model = $model->field("
                ifnull(sum(
                    case
                        when a.is_free=0 then
                            a.total_product_price
                            + a.total_product_service_consumption_tax
                            + if( a.consumption_tax_type = 2,
                                a.total_product_consumption_tax
                            , 0)
                            - a.refund_consumption_tax
                            + a.pay_fee_money
                            + (ifnull(merge_parent_order.pay_fee_money, 0) / ifnull(o_merge_order.order_count, 1))
                            - rp.free_not_include_product_price
                        when a.is_free=1 then
                            a.total_product_price
                            + a.total_product_service_consumption_tax
                            + if( a.consumption_tax_type = 2,
                                a.total_product_consumption_tax
                            , 0)
                            - a.refund_consumption_tax
                            + a.pay_fee_money
                            + (ifnull(merge_parent_order.pay_fee_money, 0) / ifnull(o_merge_order.order_count, 1))
                        else sub_order.sub_receivable_price
                    end
                ), 0) as receivable_price
            ");
        // 原商品金额 - 未税金额
        $model = $model->field("ifnull(sum(rp.not_tax_total_product_price), 0) as not_tax_total_product_price");
        // 原商品金额
        $model = $model->field("ifnull(sum(a.total_product_price), 0) as total_product_price");
        // 服务费 - 任务描述: 部分退款之前都不会退固定服务费的，手续费也不会
        // 字段：setting_service_money > 0 等于固定服务费不然就是百分比服务费
        $model = $model->field("
                ifnull(sum(
                    if (a.merge_parent_id != 0 and a.is_merge != 1, 0,
                        case
                            when a.is_free=0 || a.is_free=1 then
                                if(order_refund_type.mode > 0, 0,
                                    if(a.setting_service_money > 0,
                                        a.service_money,
                                        a.service_money - ifnull(rp.refund_product_service_fee, 0)
                                    )
                                )
                            else sub_order.sub_service_money
                        end
                    )
                ), 0) as service_money
            ");
        // 优惠折扣
        $model = $model->field("
                ifnull(sum(
                    case
                        when a.is_free=0 then a.discount_money - ifnull(rp.free_not_include_product_discount_money, 0) + a.checkout_diff_money
                        when a.is_free=1 then a.discount_money
                        else sub_order.sub_discount_money
                    end
                ), 0) as discount_money
            ");
        // 消费税费
        $model = $model->field("
                ifnull(sum(
                    case
                        when a.is_free=0 then a.consumption_tax_money - a.refund_consumption_tax
                        when a.is_free=1 then a.consumption_tax_money - a.refund_consumption_tax
                        else sub_order.sub_consumption_tax_money
                    end
                ), 0) as consumption_tax_money
            ");
        // 支付手续费（v1.0.6） - 任务描述: 部分退款之前都不会退固定服务费的，手续费也不会
        $model = $model->field("
                ifnull(sum(
                    if (a.merge_parent_id != 0 and a.is_merge != 1, 0,
                        if(order_refund_type.mode > 0, 0,
                            a.pay_fee_money
                        )
                    )
                ), 0) as pay_fee_money
            ");
        // 余额消费金额, 不包含退款 任务描述: 部分退款之前都不会退固定服务费的，手续费也不会
        $model = $model->field("
                ifnull( sum(o_pay_balance.balance_price_not_exist_refund) , 0) as received_balance_price
            ");
        // 实收金额,  任务：需求补充-关于会员充值/会员消费的数值显示 ,（v1.0.8）, 实收 = (不包括会员余额消费金额，不包含退款)
        $model = $model->field("
            ifnull(
                sum(
                    if (a.merge_parent_id != 0 and a.is_merge != 1, 0 ,
                        a.pay_price - ifnull(o_pay_balance.balance_price, 0) -
                        if (a.refund_money > ifnull(o_pay_balance.balance_price, 0) , a.refund_money - ifnull(o_pay_balance.balance_price, 0), 0 )
                    )
                )
            , 0) as received_price");

        // 基础统计
        $model = $model->field("ifnull(sum(rp.total_consumption_tax_pay_price - rp.refund_money), 0) sales_price")      // 净销售额
            ->field("ifnull(sum(rp.total_num), 0) as product_num")                                                      // 商品数量
            ->field("ifnull(sum(a.user_discount_money), 0) as user_discount_money")                                     // 会员折扣
            ->field("ifnull(sum(a.refund_money), 0) as refund_money")                                                   // 退款
            ->field("ifnull(sum(a.refund_consumption_tax), 0) as refund_consumption_tax")                               // 退款消费税
            ->field("ifnull(sum(rp.free_price), 0) as free_product_price")                                                         // 赠菜折扣：赠菜的总金额（v1.0.7）
            ->field("ifnull(sum(rp.free_num), 0) as free_product_num")                                                             // 赠菜数量：赠菜的数量（v1.0.7）
            ->field("ifnull(sum(a.free_pay_price), 0) as free_order_price")                                                        // 免单的总金额（v1.0.7）
            ->field("ifnull(sum(if(a.is_free > 0 or a.free_pay_price > 0, 1, 0)), 0) as free_order_num");                          // 免单数量（v1.0.7）
        // 子单
        $subOrder = $model->clone()->where('a.is_merge', 0)->where('a.parent_id', 0);
        // 主单
        $mainOrder = $model->clone()->where('a.is_merge', 1)->where('a.parent_id', 0);

        //  合计
        if ($this->source != 'HomeData') {
            $subOrder = $subOrder->field("count(*) as total_order_num")                                                 // 所有订单数
                ->field("ifnull(sum(if(a.table_id > 0, 1, 0)), 0) as total_table_num")                                  // 总桌数
                ->field("ifnull(sum(a.meal_num), 0) as total_people_num")                                               // 总人数
                ->field("ifnull(min(a.pay_price - a.refund_money), 0) as min_order_price")                              // 最小订单金额
                ->field("ifnull(max(a.pay_price - a.refund_money), 0) as max_order_price")                              // 最大订单金额
                ->field("ifnull(round(avg(a.pay_price - a.refund_money), 2), 0) as avg_order_price")                    // 平均订单金额
                // 桌台方式
                ->field("ifnull(sum(if(a.table_id > 0, 1, 0)), 0) as table_order_num")                                  // 订单数（桌数）
                ->field("ifnull(sum(if(a.table_id > 0, a.meal_num, 0)), 0) as table_people_num")                        // 人数
                ->field("ifnull(min(if(a.table_id > 0, a.pay_price - a.refund_money, 999999999)), 0) as table_min_order_price")    // 最小订单金额
                ->field("ifnull(max(if(a.table_id > 0, a.pay_price - a.refund_money, 0)), 0) as table_max_order_price")            // 最大订单金额
                ->field("ifnull(round(sum(if(a.table_id > 0, a.pay_price - a.refund_money, 0)) / sum(if(a.table_id > 0, 1, 0)), 2), 0) as table_avg_order_price")      // 平均订单金额
                ->field("ifnull(round(sum(if(a.table_id > 0, a.pay_price - a.refund_money, 0)) / sum(if(a.table_id > 0, a.meal_num, 0)), 2), 0) as table_people_avg")  // 桌台人均
                // 收银方式
                ->field("ifnull(sum(if(a.table_id = 0, 1, 0)), 0) as cashier_order_num")                                 // 订单数
                ->field("ifnull(min(if(a.table_id = 0, a.pay_price - a.refund_money, 999999999)), 0) as cashier_min_order_price")   // 最小订单金额
                ->field("ifnull(max(if(a.table_id = 0, a.pay_price - a.refund_money, 0)), 0) as cashier_max_order_price")           // 最大订单金额
                ->field("ifnull(round(sum(if(a.table_id = 0, a.pay_price - a.refund_money, 0)) / sum(if(a.table_id = 0, 1, 0)) , 2), 0) as cashier_avg_order_price"); // 平均订单金额
        }

        // 店内概况统计根据每日显示（导出数据）
        if ($type == 2) {
            $subOrder = $model
                ->leftJoin("
                    (
                        select opt.order_id, group_concat(DISTINCT opt.value) as pay_type
                        from {$prefix}order_pay_type opt
                        group by opt.order_id
                    ) opt
                ", 'a.order_id = opt.order_id')
                ->field("date_format(FROM_UNIXTIME(a.pay_time), '%Y-%m-%d') as date")   // 日期
                ->field("group_concat(DISTINCT opt.pay_type) as pay_type_date")         // 支付方式（当天）
                ->group("date");
            $result = $subOrder->select()->append([])?->toArray();
            foreach ($result as $key => $item) {
                $result[$key] = $this->formattedData($item);
            }
        }
        // 默认
        else {
            $subOrder = $subOrder->find()->append([])?->toArray();
            $mainOrder = $mainOrder->find()->append([])?->toArray();
            //
            $subOrder['service_money'] = $subOrder['service_money'] + $mainOrder['service_money'];
            $subOrder['pay_fee_money'] = $subOrder['pay_fee_money'] + $mainOrder['pay_fee_money'];
            $subOrder['received_price'] = $subOrder['received_price'] + $mainOrder['received_price'];
            $subOrder['received_balance_price'] = $subOrder['received_balance_price'] + $mainOrder['received_balance_price'];
            //
            $result = $this->formattedData($subOrder);
        }
        //
        return $result;
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
        $area_id = $params['area_id'] ?? 0;
        $category_id = $params['category_id'] ?? 0;
        $product_name = $params['product_name'] ?? '';
        $sort_field = $params['sort_field'] ?? 'op.create_time'; // 排序字段
        $sort_type = $params['sort_type'] ?? 'desc'; // 排序方式
        //
        $data = $this->baseModel->clone()
            ->leftJoin('table t', 'a.table_id = t.table_id')
            ->leftJoin('table_area ta', 't.area_id = ta.area_id')
            ->leftJoin('order_product op', 'a.order_id = op.order_id')
            ->leftJoin('product p', 'op.product_id = p.product_id')
            ->leftJoin('category c', 'p.category_id = c.category_id and c.is_button = 0')
            ->leftJoin('category pc', 'c.parent_id = pc.category_id and pc.is_button = 0') // 父类分类
            // 原价销售额（原商品金额+商品实付服务费+商品实付税费（如有折扣，计原价，免单的对应商品销售额服务费/税费为0））
            // consumption_tax - 总的 - 产品 + 产品服务的税费(折后)
            // product_consumption_tax - 商品消费税(折后)
            // product_service_consumption_tax - 商品服务费消费税(折后)
            // product_service_fee - 商品服务费(折后)
            ->field("
                ifnull(sum(op.total_product_price + op.product_service_consumption_tax + op.product_service_fee ), 0) as sales_price
            ")
            // 消费税费
            ->field("
                ifnull(sum(
                    case
                        when a.is_free=0 && op.is_free=0 then op.consumption_tax - op.refund_consumption_tax
                        else 0
                    end
                ), 0) as consumption_tax_money
            ")
            // 实际销售额（指商品原价扣除优惠折扣、会员折扣后的金额（包含商品实付服务费、商品实付税费））免单不计入
            ->field("
                ifnull(sum(
                    case
                        when a.is_free=0 && op.is_free=0 then op.total_pay_price + op.product_service_consumption_tax + op.product_service_fee + if(op.tax_calc_type=2, product_consumption_tax, 0) - op.refund_money
                        else 0
                    end
                ), 0) as received_price
            ")
            ->field("ifnull(sum(op.total_num), 0) as product_num")                                      // 商品数量（销售的数量，赠菜也包括在内）
            ->field("sum(if(a.is_free > 0 || op.is_free > 0, op.total_num, 0)) as free_product_num")    // 赠菜数量（包含普通赠菜、免单赠菜；免单后等于所有的商品都是赠菜；某个商品标记为赠菜，所属订单标记为免单则赠菜记1）
            ->field("ifnull(p.product_id,0) product_id, ifnull(p.product_name,'--') product_name")
            ->field("ifnull(c.category_id,0) category_id, ifnull(c.name,'--') category_name")
            ->field("ifnull(pc.category_id,0) parent_category_id, ifnull(pc.name,'--') parent_category_name") // 父类分类
            ->group("op.product_id")
            ->order($sort_field, $sort_type)
            ->where('op.product_id', '>', 0)
            ->where('op.is_return', 0) // 不包含退货
            ->when($area_id, function ($q) use ($area_id) {
                $q->where('ta.area_id', $area_id);
            })
            ->when($category_id, function ($q) use ($category_id) {
                $q->where('c.category_id', $category_id);
            })
            ->when($product_name, function ($q) use ($product_name) {
                $q->jsonLike('p.product_name', $product_name);
            })
            ->paginate($params)?->append([])->toArray() ?: [];
        //
        foreach ($data['data'] as &$item) {
            // 营业收入 = 实收金额-税费（实收的税费，包含商品税费、服务费税费）
            $item['business_price'] = helper::bcsub($item['received_price'], $item['consumption_tax_money']);
            $item['product_name_text'] = extractLanguage($item['product_name'] ?? '');
            if ($item['parent_category_id'] == 0) {
                $item['category_name_text'] = extractLanguage($item['category_name'] ?? '');
            } else {
                $item['category_name_text'] = extractLanguage($item['parent_category_name'] ?? '') . "/" . extractLanguage($item['category_name'] ?? '');
            }
        }
        return $data;
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
