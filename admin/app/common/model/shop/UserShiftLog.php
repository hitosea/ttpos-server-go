<?php

namespace app\common\model\shop;

use help\DateHelp;
use think\facade\Cache;
use app\common\library\helper;
use app\common\model\BaseModel;
use app\shop\model\product\Category;
use app\common\exception\BaseException;
use app\common\model\order\OrderRefund;
use app\common\enum\order\OrderStatusEnum;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\Order as OrderModel;
use app\common\repositories\OrderBusinessDataRepository;
use app\common\model\order\OrderProduct as OrderProductModel;
use help\HttpHelp;

/**
 * 用户交班记录模型
 */
class UserShiftLog extends BaseModel
{

    protected $name = 'staff_shift_log';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;
    protected $deleteTime = 'delete_time';

    /**
     * 生成编号
     * @return string
     */
    public static function generateNumber()
    {
        $datePart = date('Ymd'); // 获取当前日期
        $fixedPart = '01'; // 固定部分
        $randomPart = str_pad(rand(0, 99999999), 8, '0', STR_PAD_LEFT); // 生成一个8位的随机数
        $no = $datePart . $fixedPart . $randomPart;
        if (Cache::get('__USERSHIFTLOG_GENERATENUMBER__' . $no)) {
            return self::generateNumber();
        }
        Cache::set('__USERSHIFTLOG_GENERATENUMBER__' . $no, 1, 3600 * 24);
        //
        return $no;
    }

    /**
     *  设置incomes配置
     */
    public function setIncomesAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     *  获取incomes 配置
     */
    public function getIncomesAttr($value, $data)
    {
        $value = $value ? json_decode($value, true) : [];
        if ($value) {
            foreach ($value as $key => $v) {
                // if ($v['pay_type'] == OrderPayTypeEnum::FREE_PAY) {
                //     $value[$key]['pay_type_name'] = OrderPayTypeEnum::data($v['pay_type'], 2)['name'] ?? '';
                // }
                $value[$key]['pay_type_name'] = __($value[$key]['name']);
            }
        }
        return $value;
    }

    /**
     *  获取异常数据
     */
    public function getAbnormalAttr($value)
    {
        if (is_string($value)) {
            $decoded = json_decode($value, true);
            return is_array($decoded) ? $decoded : [];
        } elseif (is_array($value)) {
            return $value;
        }
        return [];
    }

    /**
     *  获取总营业额 = 所有收入减去退款
     */
    public function getTotalBusinessAttr($value, $data)
    {
        if ($value > 0) {
            return $value;
        }
        return helper::bcsub($data['total_income'] ??  0, $data['refund_amount'] ??  0);
    }

    /**
     * 关联用户表
     */
    public function user()
    {
        return $this->belongsTo('app\\common\\model\\shop\\User', 'staff_uuid', 'uuid');
    }

    /**
     * 关联快照表
     */
    public function snapshot()
    {
        return $this->belongsTo('app\\common\\model\\shop\\UserShiftSnapshot', 'id', 'shift_log_id');
    }

    /**
     * 获取列表记录
     */
    public function getList($params)
    {
        $username = $params['user_name'] ?? '';
        $userId = $params['user_id'] ?? 0;
        $startTime = isset($params['create_time'][0]) ? strtotime($params['create_time'][0]) : 0;
        $endTime = isset($params['create_time'][1]) ? strtotime($params['create_time'][1] . ' 23:59:59') : 0;
        $model = $this;
        $model = $model->alias('a')->leftJoin('staff su', 'a.staff_uuid = su.uuid');

        if ($username) {
            $model = $model->where(function ($q) use ($username) {
                $q->like('su.username|su.real_name', $username);
            });
        }

        if ($userId) {
            $model = $model->where('su.uuid', '=', $userId);
        }

        if ($startTime && $endTime) {
            $model = $model->where('a.create_time', 'between', [$startTime, $endTime]);
        }

        $orderSort = ['a.create_time' => 'desc'];
        $list = $model->with(['user' => function ($query) {
            $query->field('uuid, uuid as shop_user_id, username as user_name, IF(real_name = "", username, real_name) as real_name');
        }])
            ->field("a.*")
            ->where('a.status', 1)
            ->order($orderSort)
            ->paginate($params);
        foreach ($list as $key => $item) {
            // 时间处理
            $list[$key]['shift_start_time'] = $item['shift_start_time'] ? DateHelp::formatTimeHis($item['shift_start_time']) : '-';
            $list[$key]['shift_end_time'] = $item['shift_end_time'] ? DateHelp::formatTimeHis($item['shift_end_time']) : '-';
        }
        return $list;
    }

    /**
     * 获取详情
     */
    public function detail($id)
    {
        $detail = $this->with('user')->where('id', $id)->find();
        //
        $detail['shift_start_time'] = $detail['shift_start_time'] ? DateHelp::formatTimeHis($detail['shift_start_time']) : '-';
        $detail['shift_end_time'] = $detail['shift_end_time'] ? DateHelp::formatTimeHis($detail['shift_end_time']) : '-';
        //
        return $detail;
    }

    /**
     * 获取销售信息
     */
    public function getSalesInfo($shift_user_id, $shop_supplier_id, $startTime, $endTime)
    {
        $datas = OrderProductModel::alias('op')
            ->distinct(true)
            ->join('order a', 'op.order_id = a.order_id', 'left')
            ->join('product p', 'op.product_id = p.product_id', 'left')
            ->join('category c2', 'p.category_id = c2.category_id', 'left')
            ->join('category c', 'c.category_id = IF(c2.parent_id = 0, c2.category_id, c2.parent_id)', 'left')
            ->where('a.pay_status', '=',  OrderPayStatusEnum::SUCCESS)
            ->where('a.order_status', '=', OrderStatusEnum::COMPLETED)
            ->where('a.eat_type', '<>', 0)
            ->where('a.extra_times', '>', 0) // 已送厨
            ->where('c.parent_id', '=', 0) // 只查询一级分类
            ->where('a.shop_supplier_id', '=', $shop_supplier_id)
            ->where('a.cashier_id', '=', $shift_user_id)
            ->where('op.is_return', '=', 0)
            ->where('a.pay_time', 'between', [is_int($startTime) ? $startTime : strtotime($startTime),  is_int($endTime) ? $endTime : strtotime($endTime)])
            ->group("c.category_id")
            ->field("c.name, sum(op.total_num) as sales, sum(op.total_pay_price) as prices")
            ->select()
            ->append([])?->toArray();
        //
        foreach ($datas as $key => $data) {
            $datas[$key]['name_text'] = Category::getNameTextAttr($data['name'] ?: '');
        }
        //
        return $datas;
    }

    /**
     * 获取交班信息
     *
     * @param array $params
     * @return bool
     */
    public function getShiftInfo($params, User $shopUser = null): array
    {
        $shop_user_id = $params['shop_user_id'] ?? 0;
        $cash_taken_out = $params['cash_taken_out'] ?? '0.00';
        $cash_left = $params['cash_left'] ?? '0.00';
        //
        if (!$shopUser) {
            $shopUser = User::where('shop_user_id', '=', $shop_user_id)->find();
            if (!$shopUser) {
                throw new BaseException(['msg' => '收银员不存在', 'code' => 0]);
            }
        }

        // 如果当班记录不存在，则添加一条
        if (!$shopUser->working) {
            $workingLog = (new UserShiftLog)->createWorkingLog($shopUser);
            $shopUser->duty_no = $workingLog->shift_no;
            $shopUser->cashier_login_time = $workingLog->shift_start_time;
            $shopUser->save();
        }

        //
        $startTime = $shopUser->cashier_login_time;
        $endTime = time();
        $params['cashier_id'] = $shop_user_id;
        $params['time'] = [date('Y-m-d H:i:s', $startTime) , date('Y-m-d H:i:s', $endTime)];

        // 上一班遗留备用金
        $previous_shift_cash = $this->order('id', 'desc')->where('status', 1)->value('cash_left') ?: 0;

        // 营业数据
        $repository = (new OrderBusinessDataRepository(new OrderModel, $params));
        $incomesList = $repository->getIncomesList();
        $businessData = $repository->getBusinessData();
        $abnormalData = $repository->getAbnormalData(['duty_no' => $shopUser->duty_no]);

        // 总收入
        $totalIncome = helper::number2($businessData['business_price']);
        $totalBusiness = helper::number2($businessData['receivable_price']);
        $cashIncome = 0;
        foreach ($incomesList as &$value) {
            if ($value['pay_type'] == 40) {
                $cashIncome = helper::number2(helper::bcadd($cashIncome, $value['refund_included_price']));
            }
            if ($value['pay_type'] == -1) {
                $value['pay_type_way'] = $value['pay_type_name'] = __('免单金额');
            }
        }

        // 本班退款金额 - 兼容老数据
        if ($startTime > strtotime('2024-08-20')) {
            $refund_amount = $businessData['refund_money'];
        } else {
            $refund_amount = OrderRefund::where('cashier_id', $shopUser->shop_user_id)
                ->where('refund_method', 2)
                ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                    $q->where('create_time', 'between', [$startTime, $endTime]);
                })
                ->sum('refund_money');
        }

        //
        return [
            'shift_user_id' => $shop_user_id,                                       // 交班人id
            'shift_no' => $shopUser->duty_no,                                       // 交班编号
            'previous_shift_cash' => $previous_shift_cash,                          // 上一班遗留备用金
            'current_cash_total' => Account::getAmount($shopUser, $cashIncome),     // 当前钱箱现金
            'incomes' => $incomesList,                                              // 支付列表
            'total_income' => $totalIncome,                                         // 本班总收入 (所有支付方式)（不包含退款）- 税费 = 营业收入 = 实收金额-税费
            'refund_amount' => $refund_amount,                                      // 本班退款金额 (所有产生的退款)
            'cash_income' => $cashIncome,                                           // 本班收入现金 (已不包含退款)
            'total_business' => $totalBusiness,                                     // 本班营业总额
            'cash_taken_out' => $cash_taken_out,                                    // 本班取出现金
            'cash_left' => $cash_left,                                              // 本班遗留备用金
            'withdraw_cash' => $shopUser->working?->withdraw_cash,                  // 中途取出现金
            'deposit_cash' => $shopUser->working?->deposit_cash,                    // 中途存入现金
            'exception_remark' => $shopUser->working?->exception_remark,            // 异常报备
            'remark' => $params['remark'] ?? '',                                    // 备注
            'cashier_login_time' => $startTime,
            'abnormalData' => $abnormalData,                                        // 异常数据
        ];
    }

    /**
     * 确定交班
     *
     * @param array $params
     */
    public function shiftLog(User $staff)
    {
        $res = HttpHelp::postRequest('http://nginx/api/v1/shop/shift', json_encode([
            'is_background' => true,
            'staff_uuid' => $staff->uuid,
        ]), [
            'Content-Type' => 'application/json',
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
        ]);
        if (!$res) {
            $this->error = "请求失败";
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
     * 创建当班记录
     * @param array $params
     */
    public function createWorkingLog($shopUser)
    {
        $previousShiftCash = $this->order('id', 'desc')->where('status', 1)->value('cash_left') ?: 0;
        $data = [
            'shift_user_id' => $shopUser->shop_user_id,                // 交班人id
            'shift_no' => $this->generateNumber(),                    // 交班编号
            'previous_shift_cash' => $previousShiftCash,              // 上一班遗留备用金
            'current_cash_total' => $previousShiftCash,               // 当前钱箱现金总计(现金收入+上一班遗留备用金)
            'cash_left' => $previousShiftCash,                        // 本班遗留备用金
            'shift_start_time' => $shopUser->cashier_login_time ?: time(),
            'shift_end_time' => 0,
            'status' => 0,
        ];
        $this->save($data);
        return $this;
    }

    /**
     * 获取存在的用户列表
     */
    public function getExistUserList($limit = 1000)
    {
        return (new User)->alias('u')
            ->field('u.uuid, u.uuid as shop_user_id, u.username as user_name, IF(u.real_name = "", u.username, u.real_name) as real_name')
            ->join('staff_shift_log s', 'u.uuid = s.staff_uuid')
            ->where('s.id', '>', 0)
            ->where('s.status', '>', 0)
            ->group('u.uuid')
            ->order(['u.create_time' => 'desc'])
            ->paginate($limit);
    }
}
