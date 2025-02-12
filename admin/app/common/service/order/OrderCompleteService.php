<?php

namespace app\common\service\order;

use app\common\library\helper;
use app\common\enum\order\OrderTypeEnum;
use app\common\enum\settings\SettingEnum;
use app\common\model\user\User as UserModel;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\user\PointsLog as PointsLogModel;
use app\common\model\supplier\Supplier as SupplierModel;
use app\common\model\order\OrderSettled as OrderSettledModel;

/**
 * 已完成订单结算服务类
 */
class OrderCompleteService
{
    // 订单类型
    private $orderType;

    /**
     * 订单模型类
     * @var array
     */
    private $orderModelClass = [
        OrderTypeEnum::MASTER => 'app\common\model\order\Order',
    ];

    // 模型
    private $model;

    /* @var UserModel $model */
    private $UserModel;

    private $supplierModel;

    /**
     * 构造方法
     */
    public function __construct($orderType = OrderTypeEnum::MASTER)
    {
        $this->orderType = $orderType;
        $this->model = $this->getOrderModel();
        $this->UserModel = new UserModel;
        $this->supplierModel = new SupplierModel();
    }

    /**
     * 初始化订单模型类
     */
    private function getOrderModel()
    {
        $class = $this->orderModelClass[$this->orderType];
        return new $class;
    }

    /**
     * 执行订单完成后的操作
     */
    public function complete($orderList, $appId)
    {
        $this->settled($orderList);
        return true;
    }

    /**
     * 执行订单结算
     */
    public function settled($orderList, $type = 'normal')
    {
        // 订单id集
        $orderIds = helper::getArrayColumn($orderList, 'order_id');
        //
        if ($type == 'normal') {
            // 累积用户实际消费金额
            $this->setUserExpend($orderList);
            // 处理订单赠送的积分
            $this->setGiftPointsBonus($orderList);
            // 将订单设置为已结算
            $this->model->onBatchUpdate($orderIds, ['is_settled' => 1]);
        } else {
            // 累积用户实际消费金额
            $this->setUserExpend($orderList, 'dec');
            // 处理订单赠送的积分
            $this->setGiftPointsBonus($orderList, 'dec');
            // 将订单设置为未结算
            $this->model->onBatchUpdate($orderIds, ['is_settled' => 0]);
        }
        //
        return true;
    }

    /**
     * 供应商金额=支付金额-运费
     */
    private function setIncSupplierMoney($orderList)
    {
        // 计算并累积实际消费金额(需减去售后退款的金额)
        $supplierData = [];
        $supplierCapitalData = [];
        // 订单结算记录
        $orderSettledData = [];
        foreach ($orderList as $order) {
            if ($order['shop_supplier_id'] == 0 || $order['is_settled'] == 1) {
                continue;
            }
            // 供应价格+运费
            $supplierMoney = $order['pay_price'] - $order['refund_money'];
            //线下支付不累积余额
            if (in_array($order['pay_type']['value'], [10, 20, 30])) {
                $orderSettledData[] = [
                    'order_id' => $order['order_id'],
                    'order_money' => $order['pay_price'],
                    'pay_money' => $order['pay_price'],
                    'express_money' => $order['express_price'],
                    'supplier_money' => $supplierMoney,
                    'real_supplier_money' => $supplierMoney,
                    'sys_money' => $order['sys_money'],
                    'refund_money' => $order['refund_money'],
                ];
                // 商家结算记录
                $supplierCapitalData[] = [
                    'money' => $supplierMoney,
                    'describe' => '订单结算，订单号：' . $order['order_no'],
                ];
                !isset($supplierData[$order['shop_supplier_id']]) && $supplierData[$order['shop_supplier_id']] = 0.00;
                $supplierMoney > 0 && $supplierData[$order['shop_supplier_id']] += $supplierMoney;
            }
        }
        // 累积到供应商表记录
        $supplierData && $this->supplierModel->onBatchIncSupplierMoney($supplierData);
        // 修改平台结算金额
        $orderSettledData && (new OrderSettledModel())->saveAll($orderSettledData);
        return true;
    }

    /**
     * 处理订单赠送的积分
     */
    public function setGiftPointsBonus($orderList, $type = 'inc')
    {
        // 计算用户所得积分
        $userData = [];
        $logData = [];
        foreach ($orderList as $order) {
            if ($order['user_id'] == 0) {
                continue;
            }
            // 积分设置 是否开启购物送积分
            if ($type == 'inc') {
                $setting = SettingModel::getSupplierItem(SettingEnum::POINTS, $order['shop_supplier_id'], $order['app_id']);
                if (!$setting['is_shopping_gift']) {
                    continue;
                }
            }
            // 订单是否产生过积分赠送
            $hasOrderPoints = PointsLogModel::where('order_id', $order['order_id'])->where('scene', PointsLogSceneEnum::CONSUME)->count();
            $hasOrderPoints2 = PointsLogModel::where('order_id', $order['order_id'])->where('scene', PointsLogSceneEnum::REVERSE)->count();
            if (($type == 'dec' && !$hasOrderPoints) || ($type == 'dec' && $hasOrderPoints <= $hasOrderPoints2)) {
                continue;
            }
            // 计算用户所得积分
            $pointsBonus = $order['is_free'] ? 0 : $order['points_bonus'];
            //            if ($pointsBonus <= 0) continue;  // 1.0.7 要求使用会员免单0积分也要记录一条
            //
            !isset($userData[$order['user_id']]) && $userData[$order['user_id']] = 0;
            // 减法时 - 减去已退款的积分
            if ($type != 'inc' && $order['refund_money'] > 0) {
                $ratio = helper::bcdiv($order['points_bonus'], $order['pay_price']);
                $points = helper::bcmul($order['refund_money'], $ratio);
                $pointsBonus = helper::bcsub($pointsBonus, $points);
            }
            $userData[$order['user_id']] += $pointsBonus;
            // 整理用户积分变动明细
            $logData[] = [
                'scene' => $type == 'inc' ? PointsLogSceneEnum::CONSUME : PointsLogSceneEnum::REVERSE,
                'user_id' => $order['user_id'],
                'card_id' => UserModel::detail($order['user_id'])?->card_id,
                'value' => $type == 'inc' ? +$pointsBonus : -$pointsBonus,
                'describe' => $type == 'inc' ? vsprintf(PointsLogSceneEnum::data()[PointsLogSceneEnum::CONSUME]['describe'], [$order['order_no']]) : vsprintf(PointsLogSceneEnum::data()[PointsLogSceneEnum::REVERSE]['describe'], [$order['order_no']]),
                'order_id' => $order['order_id'],
            ];
        }
        if (!empty($userData)) {
            // 累积到会员表记录
            if ($type == 'inc') {
                $this->UserModel->onBatchIncPoints($userData);
            } else {
                $this->UserModel->onBatchDecPoints($userData);
            }
            // 批量新增积分明细记录
            (new PointsLogModel)->onBatchAdd($logData);
        }
        return true;
    }

    /**
     * 累积用户实际消费金额
     */
    public function setUserExpend($orderList, $type = 'inc')
    {
        // 计算并累积实际消费金额(需减去售后退款的金额)
        $userData = [];
        foreach ($orderList as $order) {
            if ($order['user_id'] == 0 || $order['is_free']) {
                continue;
            }
            // 订单实际支付金额
            $expendMoney = $order['pay_price'];
            // 减去订单退款的金额
            $expendMoney = $expendMoney - $order['refund_money'];
            if ($expendMoney <= 0) {
                continue;
            }
            !isset($userData[$order['user_id']]) && $userData[$order['user_id']] = 0.00;
            $userData[$order['user_id']] += $expendMoney;
        }
        // 累积到会员表记录
        if (!empty($userData)) {
            if ($type == 'inc') {
                $this->UserModel->onBatchIncExpendMoney($userData);
            } else {
                $this->UserModel->onBatchDecExpendMoney($userData);
            }
        }
        //
        return true;
    }
}
