<?php

namespace app\common\model_old\order;

use help\DateHelp;
use help\QueueHelp;
use think\facade\Cache;
use app\common\library\helper;
use app\common\model_old\BaseModel;
use app\common\model_old\user\User;
use app\common\model_old\shop\Account;
use app\common\model_old\store\PayType;
use app\common\exception\BaseException;
use app\common\model_old\payment\PaymentOrder;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\service\payment\RefundService;
use app\shop\service\order\UserRechargeExportService;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model_old\user\BalanceLog as BalanceLogModel;
use app\common\model_old\order\UserRechargeOrderOperationLog;

/**
 * 会员充值订单
 */
class UserRechargeOrder extends BaseModel
{
    protected $name = 'user_recharge_order';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'pay_time_text',
        'order_status_text',
        'order_type_text',
        'actual_price', // 实收金额
    ];

    //  order_status 订单状态 -1 -已取消 0-进行中 1-已完成
    const ORDER_STATUS_CANCEL = -1;
    const ORDER_STATUS_PROCESS = 0;
    const ORDER_STATUS_COMPLETE = 1;
    const ORDER_STATUS = [
        self::ORDER_STATUS_CANCEL => ['text' => '已取消', 'value' => self::ORDER_STATUS_CANCEL],
        self::ORDER_STATUS_PROCESS => ['text' => '进行中', 'value' => self::ORDER_STATUS_PROCESS],
        self::ORDER_STATUS_COMPLETE => ['text' => '已完成', 'value' => self::ORDER_STATUS_COMPLETE],
    ];

    // 订单来源 10-充值订单
    const ORDER_TYPE_RECHARGE = 10;
    const ORDER_TYPE = [
        self::ORDER_TYPE_RECHARGE => ['text' => '充值订单', 'value' => self::ORDER_TYPE_RECHARGE],
    ];

    /**
     * 支付方式
     */
    public function payType()
    {
        return $this->hasMany('app\\common\\model\\order\\UserRechargeOrderPayType', 'order_id', 'id')->field(['id', 'order_id', 'value', 'price', 'disabled_cancel', 'pay_status', 'fee_money']);
    }

    /**
     * 关联收银员表
     */
    public function cashier()
    {
        return $this->belongsTo('app\\common\\model\\shop\\User', 'cashier_id', 'shop_user_id');
    }

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\user\\User', 'user_id', 'user_id');
    }

    /**
     * 余额日志
     */
    public function balanceLog()
    {
        return $this->hasOne('app\\common\\model\\user\\BalanceLog', 'recharge_order_id', 'id')->order(['create_time' => 'desc']);
    }

    /**
     * 获取实收金额
     */
    public function getActualPriceAttr($value, $data)
    {
        $payPrice = is_numeric($data['pay_price'] ?? null) ? (float)$data['pay_price'] : 0.0;
        $refundMoney = is_numeric($data['refund_money'] ?? null) ? (float)$data['refund_money'] : 0.0;
        $calculatedPrice = helper::bcsub($payPrice, $refundMoney);
        return helper::number2($calculatedPrice >= 0 ? $calculatedPrice : 0);
    }
    /**
     * 应付
     */
    public function getPayPriceAttr($value)
    {
        return floatval($value);
    }

    /**
     * 订单金额
     */
    public function getOrderPriceAttr($value)
    {
        return floatval($value);
    }

    /**
     * 充值金额
     */
    public function getRechargeMoneyAttr($value)
    {
        return floatval($value);
    }

    /**
     * 赠送金额
     */
    public function getGiftMoneyAttr($value)
    {
        return floatval($value);
    }

    /**
     * 赠送积分
     */
    public function getGiftPointAttr($value)
    {
        return floatval($value);
    }

    /**
     * 支付时间格式化
     */
    public function getPayTimeTextAttr($value, $data)
    {
        return isset($data['pay_time']) && $data['pay_time'] != 0 ? DateHelp::formatTimeHis($data['pay_time']) : '-';
    }

    /**
     * 订单状态
     */
    public function getOrderStatusTextAttr($value, $data)
    {
        return isset($data['order_status']) ? __(self::ORDER_STATUS[$data['order_status']]['text'] ?? '-') : '-';
    }

    /**
     * 订单来源
     */
    public function getOrderTypeTextAttr($value, $data)
    {
        return isset($data['order_type']) ? __(self::ORDER_TYPE[$data['order_type']]['text'] ?? '-') : '-';
    }

    /**
     * 从设备号获取进行中订单
     * @param $device_id
     * @return mixed
     */
    public static function getPendingOrder($device_id)
    {
        return self::with(['payType', 'user' => function ($q) {
            $q->field(['*, (balance + gift_balance) as balance'])->with(['grade', 'card']);
        }])->where('order_status', 0)->where('device_id', $device_id)->find();
    }

    /**
     * 充值订单详情
     * @param $recharge_order_id
     * @return array|mixed
     */
    public static function detail($recharge_order_id)
    {
        return self::alias('uro')
            ->with([
                'payType',
                'user' => function ($query) {
                    $query->field(['user_id', 'nickName', 'balance', 'gift_balance', 'points']);
                },
                'cashier' => function ($query) {
                    $query->field(['shop_user_id', 'IF(real_name != "", real_name, user_name) as real_name']);
                },
            ])
            ->where('id', $recharge_order_id)
            ->find();
    }

    /**
     * 订单现金实际收到金额 (不包含找零)
     */
    public function getCashReceivePriceAttr($value, $data)
    {
        if ($cashAmount = $this->payType()->where('value', OrderPayTypeEnum::CASH)->sum('price')) {
            $cashAmount = helper::bcsub($cashAmount, $data['change_due']);
        }
        return $cashAmount;
    }

    /**
     * @param $param
     *  int user_id             会员ID，必填
     *  float recharge_money    充值金额，必填
     *  float gift_money        赠送金额，选填
     *  float gift_point        赠送积分，选填
     * @param $cashier
     * @return array|false|mixed
     */
    public function createRechargeOrder($param, $cashier)
    {
        $recharge_order_id = isset($param['recharge_order_id']) ? $param['recharge_order_id'] : 0;
        $device_id = isset($cashier['device_id']) ? $cashier['device_id'] : '';
        $app_id = isset($cashier['user']['app_id']) ? $cashier['user']['app_id'] : '';
        $shop_supplier_id = isset($cashier['user']['shop_supplier_id']) ? $cashier['user']['shop_supplier_id'] : '';
        $cashier_id = isset($cashier['user']['cashier_id']) ? $cashier['user']['cashier_id'] : '';
        if (!$device_id || !$app_id || !$shop_supplier_id) {
            $this->error = '基础信息缺失';
            return false;
        }
        $user_id = $param['user_id'] ?? 0;
        if (!User::getUser($user_id)) {
            $this->error = '会员不存在';
            return false;
        }
        $recharge_money = isset($param['recharge_money']) && is_numeric($param['recharge_money']) ? $param['recharge_money'] : 0;
        if (!$recharge_money || $recharge_money < 0) {
            $this->error = '充值金额错误';
            return false;
        }
        if ($recharge_money < 1) {
            $this->error = '最小充值金额为1';
            return false;
        }
        $gift_money = isset($param['gift_money']) && is_numeric($param['gift_money']) ? $param['gift_money'] : 0;
        if ($gift_money < 0) {
            $this->error = '赠送金额错误';
            return false;
        }
        $gift_point = $param['gift_point'] ?? 0;
        if (!is_numeric($gift_point) || $gift_point < 0) {
            $this->error = '赠送积分错误';
            return false;
        }
        // 修改订单
        if ($recharge_order_id > 0) {
            /** @var UserRechargeOrder $order */
            $order = UserRechargeOrder::where('order_status', 0)->where('id', $recharge_order_id)->find();
            if ($order) {
                $old_recharge_money = $order->recharge_money;
                $this->startTrans();
                try {
                    $updateData = [
                        'user_id' => $user_id,
                        'recharge_money' => $recharge_money,
                        'gift_money' => $gift_money,
                        'gift_point' => $gift_point,
                        'device_id' => $device_id,
                        'cashier_id' => $cashier_id,
                        'app_id' => $app_id,
                        'shop_supplier_id' => $shop_supplier_id,
                    ];
                    $order->save($updateData);
                    $reloadOrder = $order->reloadRechargePrice($order->id);
                    // 添加操作记录
                    UserRechargeOrderOperationLog::createLog(
                        $order->id,
                        UserRechargeOrderOperationLog::ACTION_CHANGE_AMOUNT,
                        [
                            'recharge_money' => $recharge_money,
                            'old_recharge_money' => $old_recharge_money,
                        ],
                        '变更充值金额'
                    );
                    $this->commit();
                    return $reloadOrder;
                } catch (BaseException $e) {
                    $this->error = $e->getMessage();
                    $this->rollback();
                    return false;
                }
            }
        }
        // 默认返回已存在进行中订单
        $pendingOrder = UserRechargeOrder::getPendingOrder($device_id);
        if ($pendingOrder) {
            return $pendingOrder;
        }
        $this->startTrans();
        try {
            $newData = [
                'order_no' => self::createRechargeOrderNo(),
                'user_id' => $user_id,
                'recharge_money' => $recharge_money,
                'gift_money' => $gift_money,
                'gift_point' => $gift_point,
                'device_id' => $device_id,
                'cashier_id' => $cashier_id,
                'app_id' => $app_id,
                'shop_supplier_id' => $shop_supplier_id,
                'order_status' => 0,
            ];
            $order = new self;
            $order->save($newData);
            $reloadOrder = $order->reloadRechargePrice($order->id);
            // 添加操作记录
            UserRechargeOrderOperationLog::createLog(
                $order->id,
                UserRechargeOrderOperationLog::ACTION_GENERATE_ORDER,
                [],
                '生成订单'
            );
            $this->commit();
            return $reloadOrder;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 生成充值订单号 第九位1是收銀/2是桌台/3是充值
     */
    public static function createRechargeOrderNo()
    {
        $type_num = 3;
        // 获取时间
        $micro_time = microtime(true);
        $date_part = date('Ymd', (int)$micro_time);
        $micro_seconds = sprintf("%.6f", $micro_time - floor($micro_time));
        $time_part = substr(str_replace('.', '', $micro_seconds), 0, 6); // 6位微秒
        //
        $n = substr(implode('', array_map('ord', str_split(substr(uniqid(), 7, 13), 1))), 0, 3);
        $orderNo = $date_part . $type_num . $time_part . $n;
        if (Cache::get('__CREATE_NEW_RC_ORDERNO__' . $orderNo)) {
            return self::createRechargeOrderNo();
        }
        Cache::set('__CREATE_NEW_RC_ORDERNO__' . $orderNo, 1, 5);
        //
        return $orderNo;
    }

    /**
     * 会员充值订单添加支付方式
     * @param $value
     * @param $price
     * @return bool
     */
    public function addPayType($value, $price, $payment_order_id = 0)
    {
        // 判断在线支付订单
        $paymentOrder = PaymentOrder::where('id', $payment_order_id)->find();
        if (in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) && !$paymentOrder) {
            $this->error = '支付订单不存在';
            return false;
        }

        $order = $this;
        //
        $orderPayType = (new UserRechargeOrderPayType);
        if ($payment_order_id > 0) {
            // 在线支付订单已支付
            $orderPayTypeRecord = $orderPayType->where('payment_order_id', $payment_order_id)->find();
            if ($orderPayTypeRecord && $orderPayTypeRecord->pay_status == 1) {
                return true;
            }
        }

        /** @var PayType $payType */
        $payType = PayType::where('value', $value)->where('status', 1)->find();
        if (!(new Order)->checkPayType($order['shop_supplier_id'], $order['app_id'], $value)) {
            $this->error = '支付方式未开启';
            return false;
        }
        $pay_type_fee_money = $payType ? $payType->calFeeMoney($price) : 0;
        $fee = $payType ? $payType->fee : 0;
        //
        if ($value == 10) {
            $this->error = '不能使用余额支付充值';
            return false;
        }
        // 判断金额
        if ($price < 0.01) {
            $this->error = '支付金额最低0.01';
            return false;
        }
        //
        $amount_paid = $orderPayType->where('order_id', $order['id'])->sum('price');
        if ($amount_paid >= $order['pay_price']) {
            $this->error = '当前已足额';
            return false;
        }
        // 非现金不能超订单应付金额
        $amount_paid = helper::bcadd($amount_paid, $price);
        if ($value != 40 && $amount_paid > $order['pay_price']) {
            $this->error = '非现金支付不能大于应收';
            return false;
        }
        // 最终支付金额 = 支付 + 手续费
        $price = helper::bcadd($price, $pay_type_fee_money);
        //
        $exist = $orderPayType->where('order_id', $order['id'])->where('value', $value)->find();
        //
        $this->startTrans();
        try {
            // 存在更新否则添加
            if ($exist) {
                $exist->where('order_id', $order['id'])->where('value', $value)->update(['price' => $price, 'fee_money' => $pay_type_fee_money]);
            } else {
                $saveArr = [
                    'order_id' => $order['id'],
                    'payment_order_id' => $payment_order_id,
                    'value' => $value,
                    'price' => $price,
                    'fee' => $fee,
                    'fee_money' => $pay_type_fee_money,
                    'app_id' => $order['app_id'],
                    'shop_supplier_id' => $order['shop_supplier_id'],
                    'pay_status' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 0 : 1, //  手动添加在线支付标为0
                    'disabled_cancel' => in_array($value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY]) ? 1 : 0, //  在线支付不可撤销
                ];
                $orderPayType->save($saveArr);
            }
            $order->reloadRechargePrice($order->id);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 会员充值订单撤销支付方式
     * @param $id
     * @return false
     */
    public function cancelPayType($id)
    {
        $opt = (new UserRechargeOrderPayType)->where('id', $id)->find();
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
            $this->reloadRechargePrice($this->id);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 计算充值订单金额
     * @param $recharge_order_id
     * @return array|mixed
     */
    public function reloadRechargePrice($recharge_order_id)
    {
        $order = self::detail($recharge_order_id);
        // 支付方式手续费
        $total_fee_money = UserRechargeOrderPayType::where('order_id', $recharge_order_id)->sum('fee_money');
        //
        $order_price = $order->recharge_money;
        $pay_price = helper::bcadd($order_price, $total_fee_money);
        //
        $updateOrderArr = [
            'order_price' => $order_price,
            'pay_fee' => $total_fee_money,
            'pay_price' => $pay_price,
        ];
        $order->save($updateOrderArr);
        //
        return $order;
    }

    /**
     * 充值订单结账
     * @return bool
     */
    public function confirm($user_id, $cashier = [])
    {
        $cashier_id = isset($cashier['user']['cashier_id']) ? $cashier['user']['cashier_id'] : '';
        /** @var User $user */
        $user = (new User)::detail($user_id);
        if (!$user) {
            $this->error = '会员不存在';
            return false;
        }
        $cashier_name = $cashier['user']['user_name'] ? $cashier['user']['user_name'] : $cashier['user']['real_name'];
        $this->startTrans();
        try {
            // 当班编号
            $this->duty_no = isset($cashier['user']['duty_no']) ? $cashier['user']['duty_no'] : '';
            // 处理订单状态
            if ($this->order_status != self::ORDER_STATUS_PROCESS) {
                $this->error = '当前状态不可操作';
                return false;
            }
            $total_pay_type_price = UserRechargeOrderPayType::where('order_id', $this->id)->sum('price');
            if ($this->order_price < 1) {
                $this->error = '最小充值金额为1';
                return false;
            }
            // 检查非现金超额
            $total_non_cash_pay_type_price = $this->getNonCashPayTypeTotalPrice($this->id);
            if ($total_non_cash_pay_type_price > $this->pay_price) {
                $this->error = "收款金额大于充值金额，请先修改收款金额";
                return false;
            }
            if ($total_pay_type_price < $this->pay_price) {
                $this->error = '未足额支付';
                return false;
            }
            $change_due = helper::bcsub($total_pay_type_price, $this->pay_price);
            $pay_time = time();
            $orderSaveArr = [
                'user_id' => $user_id,
                'order_status' => 1,
                'change_due' => $change_due,
                'pay_time' => $pay_time,
            ];
            $this->save($orderSaveArr);
            // 赠送积分
            if ($this->gift_point > 0) {
                $user->setIncPoints($this->gift_point, vsprintf("收银机管理员充值赠送操作 [%s]", [$cashier_name]), PointsLogSceneEnum::RECHARGE_GIVE, true);
            }
            // 处理会员余额
            $gift_money = $this->gift_money;
            $total_money = helper::bcadd($this->recharge_money, $gift_money);
            if ($total_money > 0) {
                BalanceLogModel::add(BalanceLogSceneEnum::RECHARGE, [
                    'recharge_order_id' => $this->id,
                    'user_id' => $user['user_id'],
                    'money' => $total_money,
                    'gift_money' => $gift_money,
                ], vsprintf("收银机管理员操作 [%s]", [$cashier_name]), true);
            }
            // 更新店铺账户余额
            if ($cashAmount = $this->getCashReceivePriceAttr(null, $this)) {
                Account::updateAmount(1, $cashAmount, $this->order_no, $cashier_id, $this->shop_supplier_id, $this->app_id, 'recharge-pay');
            }
            // 添加操作记录
            UserRechargeOrderOperationLog::createLog($this['id'], UserRechargeOrderOperationLog::ACTION_RECHARGE, [
                'order_price' => $this->order_price,                        //  订单金额
                'recharge_money' => $this->recharge_money,                  //  充值金额
                'pay_price' => $this->pay_price,                            //  订单应付
                'gift_money' => $this->gift_money,                          //  赠送金额
                'gift_point' => $this->gift_point,                          //  赠送积分
                'pay_fee' => $this->pay_fee,                                //  支付手续费
                'change_due' => $this->change_due,                          //  找零
                'pay_type' => $this->payType,                               //  支付方式
            ], '充值');
            //
            $this->commit();
            return $this;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 结账小计
     * @param $order_id
     * @return array
     */
    public static function getPayRes($recharge_order_id)
    {
        $order = self::where('id', $recharge_order_id)->find();
        return [
            'recharge_order_id' => $order->id,
            'pay_price' => $order->pay_price,                   //  应收金额
            'pay_type' => self::getPayTypeList($recharge_order_id),      //  支付方式
            'actual_price' => self::getActualPrice($recharge_order_id),             //  实收金额
            'change_due' => $order->change_due,                 //  找零
        ];
    }

    /**
     * 获取订单支付方式
     * @param $order_id
     * @return UserRechargeOrderPayType[]|array|\think\Collection
     */
    public static function getPayTypeList($order_id)
    {
        return (new UserRechargeOrderPayType())->field(['id', 'value', 'price', 'disabled_cancel', 'fee_money'])->where('order_id', $order_id)->select();
    }

    /**
     * 实收金额
     * @param $order_id
     * @return float
     */
    public static function getActualPrice($order_id)
    {
        return UserRechargeOrderPayType::where('order_id', $order_id)->sum('price');;
    }

    /**
     * 获取订单非现金支付方式总额
     * @param $order_id
     * @return float
     */
    public function getNonCashPayTypeTotalPrice($order_id)
    {
        return UserRechargeOrderPayType::where('order_id', $order_id)->where('value', '<>', PayType::CASH_VALUE)->sum('price');
    }

    /**
     * 订单列表
     */
    public function getList($dataType, $data = null)
    {
        $model = $this;

        // 检索查询条件
        $model = $model->setWhere($model, $data);
        // 获取数据列表
        return $model
            ->with([
                'payType',
                'user' => function ($query) {
                    $query->field(['user_id', 'nickName', 'balance', 'gift_balance', 'points']);
                },
                'cashier' => function ($query) {
                    $query->field(['shop_user_id', 'IF(real_name != "", real_name, user_name) as real_name']);
                },
            ])
            ->where($this->transferDataType($dataType))
            ->order(['create_time' => 'desc'])
            ->paginate($data);
    }

    /**
     * 转义数据类型条件
     */
    private function transferDataType($dataType)
    {
        $filter = [];
        // 订单数据类型 -1-已取消 0-进行中 1-已完成
        switch ($dataType) {
            case 'all':
                $filter[] = ['id', '>', 0];
                break;
            case 'payment';
                $filter[] = ['order_status', '=', 0];
                break;
            case 'complete';
                $filter[] = ['pay_time', '>', 0];
                $filter[] = ['order_status', '=', 1];
                break;
            case 'cancel';
                $filter[] = ['order_status', '=', -1];
                break;
        }
        return $filter;
    }

    /**
     * 获取订单总数
     */
    public function getCount($type, $data)
    {
        $model = $this;
        $model = $model->setWhere($model, $data);
        return $model->where($this->transferDataType($type))->count();
    }

    /**
     * 设置检索查询条件
     */
    private function setWhere($model, $data)
    {
        // 时间类型 0-全都 1-今天 2-昨天 3-周
        $startTime = 0;
        $endTime = 0;
        if (isset($data['time_type']) && $data['time_type']) {
            switch ($data['time_type'] ?? 1) {
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
            }
        }
        // 搜索订单号
        if (isset($data['order_no']) && $data['order_no'] != '') {
            $model = $model->like('order_no', trim($data['order_no']));
        }
        // 搜索时间段
        if (isset($data['time_mode']) && is_array($data['time_mode']) && count($data['time_mode']) == 1 && $data['time']) {
            $time_mode = $data['time_mode'][0];
            $timeField = $time_mode == 1 ? 'pay_time' : 'create_time';
            $data[$timeField] = $data['time'] ?? '';
            //
            $timeFields = ['create_time', 'pay_time'];
            foreach ($timeFields as $field) {
                if (isset($data[$field]) && is_array($data[$field])) {
                    $startTime = isset($data[$field][0]) && $data[$field][0] ? strtotime($data[$field][0]) : null;
                    $endTime = isset($data[$field][1]) && $data[$field][1] ? strtotime($data[$field][1]) + 86399 : null;
                    // 开始时间 + 结束时间
                    if ($startTime && $endTime) {
                        $model = $model->where($field, 'between', [$startTime, $endTime]);
                    } elseif ($startTime) {
                        // 只有开始时间
                        $model = $model->where($field, '>', $startTime);
                    } elseif ($endTime) {
                        // 只有结束时间
                        $model = $model->where($field, '<', $endTime)->where($field, '>', 0);
                    }
                }
            }
        } else if (isset($data['time_mode']) && is_array($data['time_mode']) && count($data['time_mode']) == 2 && $data['time']) {
            $startTime = isset($data['time'][0]) && $data['time'][0] ? strtotime($data['time'][0]) : null;
            $endTime = isset($data['time'][1]) && $data['time'][1] ? strtotime($data['time'][1]) + 86399 : null;
            // 开始时间 + 结束时间
            if ($startTime && $endTime) {
                $model = $model->where('create_time', 'between', [$startTime, $endTime])->whereOr('pay_time', 'between', [$startTime, $endTime]);
            } else if ($startTime) {
                // 只有开始时间
                $model = $model->where('create_time', '>', $startTime)->whereOr('pay_time', '>', $startTime);
            } else if ($endTime) {
                // 只有结束时间
                $model = $model->where(function ($query) use ($endTime) {
                    $query->where(function ($subQuery) use ($endTime) {
                        $subQuery->where('create_time', '<', $endTime)
                            ->where('create_time', '>', 0);
                    })->whereOr(function ($subQuery) use ($endTime) {
                        $subQuery->where('pay_time', '<', $endTime)
                            ->where('pay_time', '>', 0);
                    });
                });
            }
        } else if ($startTime && $endTime) {
            // 没有时间范围才按 time_type 查询
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }
        return $model;
    }

    /**
     * 充值订单导出
     */
    public function exportList($dataType, $query)
    {
        // 获取订单列表
        try {
            $list = $this->getList($dataType, $query);
            if (count($list) > 1000) {
                $this->error = '请选择具体时间段，最多可导出1000条以下的数据';
                return false;
            }
        } catch (\Throwable $th) {
            $this->error = '请选择具体时间段，最多可导出1000条以下的数据';
            return false;
        }
        // 导出excel文件
        return (new UserRechargeExportService)->orderList($list?->toArray()['data'] ?: []);
    }

    /**
     * 取消
     */
    public function cancel($params)
    {
        if ($this->order_status != self::ORDER_STATUS_PROCESS) {
            $this->error = '当前状态不可操作';
            return false;
        }
        //
        if (count($this->payType) > 0) {
            $this->error = '当前订单已被部分支付，不支持取消';
            return false;
        }
        $this->startTrans();
        try {
            $this->save(['order_status' => -1, 'cancel_remark' => $params['cancel_remark'] ?? '']);
            // 添加操作记录
            UserRechargeOrderOperationLog::createLog(
                $this->id,
                UserRechargeOrderOperationLog::ACTION_ORDER_CANCEL,
                [],
                '取消'
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
     * 退款-支付方式
     */
    public function refundPayType()
    {
        return $this->hasMany('app\\common\\model\\order\\UserRechargeOrderPayType', 'order_id', 'id')
            ->alias('uropt')
            ->field("uropt.order_id, uropt.value, uropt.price, uropt.disabled_cancel, uropt.pay_status, uropt.fee_money, uropt.payment_order_id")
            ->field("pt.source, pt.name, pt.remark")
            ->leftJoin('pay_type pt', "pt.value = uropt.value")
            ->orderRaw("pt.source, pt.value");
    }

    /**
     * 退款-退款方式
     */
    public function refundDestinations()
    {
        return $this->hasMany('app\\common\\model\\order\\UserRechargeOrderRefundDestination', 'order_id', 'id')
            ->alias('urord')
            ->field("urord.order_id, urord.value, urord.price, sum(urord.refund_money) as refund_money")
            ->field("pt.source, pt.name, pt.remark")
            ->leftJoin('pay_type pt', "pt.value = urord.value")
            ->group("urord.value")
            ->orderRaw("pt.source, pt.value");
    }

    /**
     * 退款-支付记录（包含可退款金额）
     */
    public function payTypeCellRefundMoneys()
    {
        $res = [];
        $refundDestination = array_column($this->refundDestinations->toArray(), null, 'value');
        foreach ($this->refundPayType as $pay) {
            $payTypeCellRefundMoney = $pay['price'];
            if ($pay['value'] == OrderPayTypeEnum::CASH) {
                $pay['price'] = floatval(helper::bcsub($payTypeCellRefundMoney, $this->change_due));
                $payTypeCellRefundMoney = floatval(helper::bcsub($payTypeCellRefundMoney, $this->change_due));
            }
            if (isset($refundDestination[$pay['value']])) {
                $payTypeCellRefundMoney = floatval(helper::bcsub($payTypeCellRefundMoney, $refundDestination[$pay['value']]['refund_money']));
            }
            $pay['cell_refund_money'] = helper::number2($payTypeCellRefundMoney);
            $res[] = $pay->toArray();
        }
        return $res;
    }

    /**
     * 退款
     */
    public function refund(array $params)
    {
        $order = $this->toArray();
        $refundMethod = 3; // 退款方式: 1-系统退款, 2-线下退款, 3-原路返还
        $refundType = $params['refund_type'] ?? 0; // 退款类型: 1-整单退款, 2-部分退款
        $refundMoney = $params['refund_money'] ?? 0; // 部分退款金额
        $cashierId = $params['cashier_id'] ?? 0;
        $bankCode = $params['bank_code'] ?? '';
        $accountNo = $params['account_no'] ?? '';
        $accountName = $params['account_name'] ?? '';

        // 检查订单是否有效
        if ($this->order_status != self::ORDER_STATUS_COMPLETE) {
            $this->error = '该订单不合法';
            return false;
        }

        // 是否用户存在
        if (!$user = User::find($order['user_id'])) {
            $this->error = '用户不存在';
            return false;
        }

        $totalRefundMoney = 0; // 累计退款金额
        $userRechargeOrderRefundLog = null;
        $this->startTrans();

        try {
            /**
             * 退款类型: 1-整单退款, 2-部分退款
             */
            if ($refundType == 1) {
                $refundMoney = (float) helper::bcsub($order['pay_price'], $order['refund_money']);
                // 计算整单退款可用金额
                $availablePayPrice = helper::bcsub($order['pay_price'], $order['refund_money']);
                $totalRefundMoney = $availablePayPrice;

                if ($totalRefundMoney <= 0) {
                    $this->error = '无法退款';
                    return false;
                }

                $saveData = [
                    'order_id'         => $order['id'],
                    'refund_type'      => $refundType,
                    'refund_method'    => $refundMethod,
                    'refund_money'     => $totalRefundMoney,
                    'shop_supplier_id' => $order['shop_supplier_id'],
                    'app_id'           => $order['app_id'],
                    'cashier_id'       => $cashierId,
                ];

                $userRechargeOrderRefundLog = UserRechargeOrderRefund::create($saveData);

                // 更新订单退款金额
                $this->save([
                    'refund_money' => $order['pay_price'],
                ]);
            } elseif ($refundType == 2) {
                $totalRefundMoney = helper::bcadd($totalRefundMoney, $refundMoney); // 主表累计
                $refundableAmount = (float) helper::bcsub($order['pay_price'], $order['refund_money']);

                if ($totalRefundMoney > $refundableAmount) {
                    $this->error = '退款金额不能大于实付金额 ' . $refundableAmount;
                    return false;
                }

                if ($totalRefundMoney <= 0) {
                    $this->error = '无法退款';
                    return false;
                }

                $saveData = [
                    'order_id'         => $order['id'],
                    'refund_type'      => $refundType,
                    'refund_method'    => $refundMethod,
                    'refund_money'     => $totalRefundMoney,
                    'shop_supplier_id' => $order['shop_supplier_id'],
                    'app_id'           => $order['app_id'],
                    'cashier_id'       => $cashierId,
                ];

                $userRechargeOrderRefundLog = UserRechargeOrderRefund::create($saveData);

                $totalRefundMoney = helper::bcadd($totalRefundMoney, $order['refund_money']);
                $this->save([
                    'refund_money' => $totalRefundMoney,
                ]);
            } else {
                $this->error = '退款类型错误';
                return false;
            }

            // 当前会员主账户余额不足以退款
            if ($refundMoney > $user['balance']) {
                $this->error = '当前会员主账户余额不足以退款';
                return false;
            }

            // 添加退款记录目的地
            $refundDestination = UserRechargeOrderRefundDestination::createData(
                $this,
                $userRechargeOrderRefundLog,
                $refundType,
                $refundMethod,
                $cashierId,
                $bankCode,
                $accountNo,
                $accountName
            );

            $orderProductReturnList = $refundDestination['orderProductReturnList'];
            $cashRefundMoney = $refundDestination['cashRefundMoney'];

            // 添加操作记录
            UserRechargeOrderOperationLog::createLog(
                $order['id'],
                UserRechargeOrderOperationLog::ACTION_REFUND,
                [
                    'pay_type'      => $orderProductReturnList, // 支付方式
                    'refund_type'   => $refundType,             // 退款类型
                    'refund_method' => $refundMethod,           // 退款方式
                    'refund_money'  => $refundMoney,            // 退款金额
                ],
                '退款'
            );

            // 退款金额大于0时，减少用户余额
            if ($refundMoney > 0) {
                // 更新用户余额记录
                $this->transaction(function () use ($refundMoney) {
                    BalanceLogModel::add(BalanceLogSceneEnum::RECHARGE_REFUND, [
                        'recharge_order_id' => $this->id,
                        'user_id' => $this->user_id,
                        'money' => -$refundMoney,
                    ], [$this->order_no]);
                });
            }

            // 添加店铺账户余额
            if ($cashRefundMoney > 0) {
                Account::updateAmount(
                    0,
                    $cashRefundMoney,
                    $order['order_no'],
                    $cashierId,
                    $order['shop_supplier_id'],
                    $order['app_id'],
                    'recharge-refund'
                );
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

    /**
     * 重新退款
     * @param $payment_order_id
     * @param $refund_money
     * @param $refund_destination_id
     * @return bool
     */
    public function refundAgain($payment_order_id, $refund_money, $refund_destination_id, $bank_code = '', $account_no = '', $account_name = '')
    {
        if (!$payment_order_id || !$refund_money) {
            $this->error = '参数错误';
            return false;
        }
        //
        $orderRefundDestination = UserRechargeOrderRefundDestination::where('id', $refund_destination_id)->find();
        if (!$orderRefundDestination) {
            $this->error = '退款记录不存在';
            return false;
        }

        // 禁止并发操作
        $queue = new QueueHelp('CASHIER_REFUNDAGAIN_' . $orderRefundDestination->app_id . '-' . $orderRefundDestination->id);
        $queue->while();

        // 原退款信息
        $old_bank_code = $orderRefundDestination->bank_code;
        $old_account_no = $orderRefundDestination->account_no;
        $old_account_name = $orderRefundDestination->account_name;

        //
        $refundService = (new RefundService);
        //
        if ($orderRefundDestination->refund_order_id === 0) {
            $queue->release();
            $this->error = '该订单正在进行退款，无法重复操作';
            return false;
        }
        // 已发起过退款 - 验证退款状态
        if ($orderRefundDestination->refund_order_id) {
            $refund = $refundService->refund($payment_order_id, $refund_money, $orderRefundDestination->refund_order_id, $orderRefundDestination->merchant_refund_order_no, $old_bank_code, $old_account_no, $old_account_name);
            if (!$refund) {
                $this->error = $refundService->getError() ?: '请求支付服务失败';
                $queue->release();
                return false;
            }
            // 判断退款状态
            if (isset($refund['refund_status'])) {
                if ($refund['refund_status'] == 'RP') {
                    $queue->release();
                    $this->error = '该订单正在进行退款，无法重复操作';
                    return false;
                }
                if ($refund['refund_status'] == 'RS') {
                    $orderRefundDestination->save(['status' => 1]);
                    $queue->release();
                    $this->error = '该订单已成功退款，无法重复退款';
                    return false;
                }
            }
            // 更新银行信息 - 重新发起退款
            if (($bank_code && $bank_code != $old_bank_code) || ($account_no && $account_no != $old_account_no) || ($account_name && $account_name != $old_account_name)) {
                $merchant_refund_order_no = (new RefundService)->generateMerchantRefundOrderNo();    // 生成退款单号
                $orderRefundDestination->bank_code = $bank_code ?: $old_bank_code;
                $orderRefundDestination->account_no = $account_no ?: $old_account_no;
                $orderRefundDestination->account_name = $account_name ?: $old_account_name;
                $orderRefundDestination->status = 0;
                $orderRefundDestination->merchant_refund_order_no = $merchant_refund_order_no;
                //
                $refund = $refundService->refund($payment_order_id, $refund_money, '', $merchant_refund_order_no, $orderRefundDestination->bank_code, $orderRefundDestination->account_no, $orderRefundDestination->account_name);
                if (!$refund) {
                    $this->error = $refundService->getError() ?: '请求支付服务失败';
                    $queue->release();
                    return false;
                }
                //
                $orderRefundDestination->refund_order_id = $refund['refund_order_id'];
                $orderRefundDestination->save();
            }
        }
        // 更新银行信息 - 发起退款
        else {
            if (($bank_code && $bank_code != $old_bank_code) || ($account_no && $account_no != $old_account_no) || ($account_name && $account_name != $old_account_name)) {
                $orderRefundDestination->bank_code = $bank_code ?: $old_bank_code;
                $orderRefundDestination->account_no = $account_no ?: $old_account_no;
                $orderRefundDestination->account_name = $account_name ?: $old_account_name;
            }
            //
            $refund = $refundService->refund($payment_order_id, $refund_money, '', $orderRefundDestination->merchant_refund_order_no, $orderRefundDestination->bank_code, $orderRefundDestination->account_no, $orderRefundDestination->account_name);
            if (!$refund) {
                $this->error = $refundService->getError() ?: '请求支付服务失败';
                $queue->release();
                return false;
            }
            // 记录 refund_order_id 因为是新发的请求，不能存状态，因为刚发起肯定是处理中,记录refund_order_id就行
            $orderRefundDestination->refund_order_id = $refund['refund_order_id'];
            $orderRefundDestination->status = 0;
            $orderRefundDestination->save();
        }

        //
        $queue->release();
        return true;
    }

    /**
     * 反结账
     * @param $device_id
     * @return bool
     */
    public function reverseSettle($device_id = '')
    {
        // 检查订单状态
        if ($this->order_status != self::ORDER_STATUS_COMPLETE) {
            $this->error = '当前状态不可操作';
            return false;
        }
        // 获取用户信息
        $userId = $this->user_id;
        /** @var  User $user*/
        $user = User::find($userId);
        if (!$user) {
            $this->error = '会员不存在';
            return false;
        }
        //
        $this->startTrans();
        try {
            // 处理会员积分
            if ($this->gift_point > 0) {
                $user->setIncPoints(-$this->gift_point, $this->order_no, PointsLogSceneEnum::RECHARGE_REVERSE);
            }
            // 处理会员余额
            $gift_money = $this->gift_money;
            $recharge_money = $this->recharge_money;
            $total_money = helper::bcadd($recharge_money, $gift_money);
            if ($total_money > 0) {
                BalanceLogModel::add(BalanceLogSceneEnum::RECHARGE_REVERSE, [
                    'recharge_order_id' => $this->id,
                    'user_id' => $this->user_id,
                    'money' => -$total_money,
                    'gift_money' => -$gift_money,
                ], [$this->order_no]);
            }
            // 返店铺账户余额
            if ($cashAmount = $this->getCashReceivePriceAttr(null, $this)) {
                Account::updateAmount(0, $cashAmount, $this->order_no, $this->cashier_id, $this->shop_supplier_id, $this->app_id, 'order-reverse-settle');
            }
            // 添加操作记录
            UserRechargeOrderOperationLog::createLog($this->id, UserRechargeOrderOperationLog::ACTION_REVERSE_SETTLE, [
                'pay_price' => $this->pay_price,                           //  应收金额
                'pay_type' => $this->payType,                              //  支付方式
                'change_due' => $this->change_due,                         //  找零
                'discount_money' => $this->free_pay_price,                 //  免单金额
            ], '反结账');
            // 删除之前的组合支付
            $this->payType()->delete();
            // 最后更新反结账订单相关信息
            $this->save([
                'order_status' => self::ORDER_STATUS_PROCESS,
                'device_id' => $device_id,
                'pay_time' => 0,
                'refund_money' => 0,
                'pay_fee' => 0,
                'pay_price' => 0,
            ]);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }
}
