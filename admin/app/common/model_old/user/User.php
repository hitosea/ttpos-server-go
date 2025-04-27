<?php


namespace app\common\model_old\user;

use help\ValidateHelp;
use app\common\model_old\BaseModel;
use app\common\model_old\order\Order;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\model_old\settings\Setting as SettingModel;
use app\common\model_old\user\PointsLog as PointsLogModel;
use app\shop\model\user\BalanceLog as BalanceLogModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum as SceneEnum;

/**
 * 用户模型
 */
class User extends BaseModel
{
    protected $pk = 'user_id';
    protected $name = 'user';

    /**
     * 默认头像
     */
    public function getAvatarUrlAttr($value)
    {
        return $value ? $value : SettingModel::getItem('store', self::$app_id)['avatarUrl'];
    }

    /**
     * 生日
     */
    public function getBirthdayAttr($value)
    {
        return $value ? date('Y-m-d', $value) : '';
    }

    /**
     * 积分
     */
    public function getPointsAttr($value, $data)
    {
        if (isset($data['points'])) {
            return floatval($data['points']);
        } else {
            return $value;
        }
    }

    /**
     * 余额
     */
    public function getBalanceAttr($value, $data)
    {
        if (isset($data['balance'])) {
            return floatval($data['balance']);
        } else {
            return $value;
        }
    }

    /**
     * 获取总余额（包含赠送金额）
     */
    public function getTotalBalanceAttr($value = null, $data = [])
    {
        if (isset($data['balance']) && isset($data['gift_balance'])) {
            return floatval($data['balance']) + floatval($data['gift_balance']);
        } else if ($this->balance && $this->gift_balance) {
            return floatval($this->balance) + floatval($this->gift_balance);
        } else {
            return '-';
        }
    }

    /**
     * 关联会员等级表
     */
    public function grade()
    {
        return $this->belongsTo('app\\common\\model\\user\\Grade', 'grade_id', 'grade_id');
    }

    /**
     * 关联会员卡表
     */
    public function card()
    {
        return $this->hasOne('app\\common\\model\\user\\Card', 'card_id', 'card_id');
    }

    /**
     * 关联会员卡记录表
     */
    public function cardRecord()
    {
        return $this->hasOne('app\\common\\model\\user\\CardRecord', 'user_id', 'user_id')->where('is_delete', 0);
    }

    /**
     * 关联收货地址表
     */
    public function address()
    {
        return $this->hasOne('app\\common\\model\\user\\UserAddress', 'address_id', 'address_id');
    }

    /**
     * 关联收货地址表 (默认地址)
     */
    public function addressDefault()
    {
        return $this->belongsTo('app\\common\\model\\user\\UserAddress', 'address_id', 'address_id');
    }

    /**
     * 获取用户信息
     */
    public static function detail($where, $includeDeleted = false)
    {
        $model = new static;
        $filter = $includeDeleted ? [] : ['is_delete' => 0];
        $filter = is_array($where) ? array_merge($filter, $where) : array_merge($filter, ['user_id' => (int) $where]);

        $info = $model->field(['*, (balance + gift_balance) as balance'])->where($filter)->with(['address', 'addressDefault', 'grade', 'card'])->find();
        if ($info) {
            $info->password = '';
        }
        return $info;
    }

    /**
     * 获取用户信息
     */
    public static function cardDetail($where)
    {
        $model = new static;
        $filter = ['is_delete' => 0];
        if (is_array($where)) {
            $filter = array_merge($filter, $where);
        } else {
            $filter['user_id'] = (int) $where;
        }
        //
        $info = $model->field(['*, (balance + gift_balance) as balance'])->where($filter)->with(['grade', 'card', 'cardRecord'])->find();
        if ($info) {
            $info->password = '';
        }
        return $info;
    }

    /**
     * 获取用户信息
     */
    public static function detailByUnionid($unionid)
    {
        $model = new static;
        $filter = ['is_delete' => 0];
        $filter = array_merge($filter, ['union_id' => $unionid]);
        //
        $info = $model->where($filter)->with(['address', 'addressDefault', 'grade', 'card'])->find();
        if ($info) {
            $info->password = '';
        }
        return $info;
    }

    /**
     * 指定会员等级下是否存在用户
     */
    public static function checkExistByGradeId($gradeId)
    {
        $model = new static;
        return !!$model->where('grade_id', '=', (int) $gradeId)
            ->where('is_delete', '=', 0)
            ->value('user_id');
    }

    /**
     * 累积用户总消费金额
     */
    public function setIncPayMoney($money)
    {
        return $this->where('user_id', '=', $this['user_id'])->inc('pay_money', $money)->update();
    }

    /**
     * 累积用户总消费金额
     */
    public function setDecPayMoney($money)
    {
        return $this->where('user_id', '=', $this['user_id'])->dec('pay_money', $money)->update();
    }

    /**
     * 累积用户实际消费的金额 (批量)
     */
    public function onBatchIncExpendMoney($data)
    {
        foreach ($data as $userId => $expendMoney) {
            $this->where(['user_id' => $userId])->inc('expend_money', $expendMoney)->update();
            event('UserGrade', $userId);
        }
        return true;
    }
    /**
     * 累积用户实际消费的金额 (批量)
     */
    public function onBatchDecExpendMoney($data)
    {
        foreach ($data as $userId => $expendMoney) {
            $this->where(['user_id' => $userId])->dec('expend_money', $expendMoney)->update();
        }
        return true;
    }

    /**
     * 累积用户实际消费的金额
     */
    public function IncExpendMoney($data)
    {
        $this->where(['user_id' => $data['user_id']])->inc('expend_money', $data['money'])->update();
        event('UserGrade', $data['user_id']);
        return true;
    }

    /**
     * 累积用户的可用积分数量 (批量)
     */
    public function onBatchIncPoints($data)
    {
        foreach ($data as $userId => $expendPoints) {
            $this->where(['user_id' => $userId])->inc('points', $expendPoints)->inc('total_points', $expendPoints)->update();
            event('UserGrade', $userId);
        }
        return true;
    }

    /**
     * 累积用户的可用积分数量 (批量)
     */
    public function onBatchDecPoints($data)
    {
        foreach ($data as $userId => $expendPoints) {
            // 获取当前用户的积分
            $user = $this->where(['user_id' => $userId])->find();
            if ($user) {
                // 计算新的积分值，确保不小于0
                $newPoints = max(0, $user->points - $expendPoints);
                $newTotalPoints = max(0, $user->total_points - $expendPoints);

                // 更新用户的积分
                $this->where(['user_id' => $userId])
                    ->update(['points' => $newPoints, 'total_points' => $newTotalPoints]);

                // 触发用户等级事件
                event('UserGrade', $userId);
            }
        }
        return true;
    }

    /**
     * 累积用户的可用积分
     */
    public function setIncPoints($points, $describe, $scene = PointsLogSceneEnum::ADMIN, $custom_dec = false)
    {
        // 新增积分变动明细
        PointsLogModel::add([
            'scene' => $scene,
            'card_id' => $this['card_id'],
            'user_id' => $this['user_id'],
            'value' => $points,
            'describe' => $custom_dec ? $describe : vsprintf(PointsLogSceneEnum::data()[$scene]['describe'], [$describe]),
            'app_id' => $this['app_id'],
        ]);

        // 更新用户可用积分
        $data['points'] = ($this['points'] + $points <= 0) ? 0 : $this['points'] + $points;
        // 用户总积分
        if ($points > 0) {
            $data['total_points'] = $this['total_points'] + $points;
        }
        $this->where('user_id', '=', $this['user_id'])->update($data);
        event('UserGrade', $this['user_id']);
        return true;
    }

    /**
     * 累计邀请书
     */
    public function setIncInvite($user_id)
    {
        $this->where('user_id', '=', $user_id)->inc('total_invite')->update();
        event('UserGrade', $user_id);
    }

    /**
     * 更新会员卡id
     */
    public function setCardId($cardId)
    {
        return $this->save(['card_id' => $cardId]);
    }

    /*
     * 添加会员
     */
    public function add($data)
    {
        $grade_id = $data['grade_id'] ?? Grade::getDefaultGradeId();
        $password = $data['password'] ?? '';
        if (empty($grade_id)){
            $this->error = '请选择会员等级';
            return false;
        }
        if (!Grade::detail($grade_id)) {
            $this->error = '会员等级不存在';
            return false;
        }
        if (!empty($data['nick_name'])) {
            if (mb_strlen($data['nick_name']) > 50) {
                $this->error = '昵称不能超过50个字符';
                return false;
            }
        }
        if (empty($data['mobile'])) {
            $this->error = '手机号不能为空';
            return false;
        }
        if (mb_strlen($data['mobile']) > 20) {
            $this->error = '手机号不能超过20个字符';
            return false;
        }
        // 密码是否为4-16位纯数字
        if ($password && !ValidateHelp::validateNumber($password)) {
            $this->error = '密码必须为4-16位纯数字';
            return false;
        }
        $user = $this->where('mobile', '=', $data['mobile'])
            ->where('is_delete', '=', 0)
            ->find();

        if (!$user) {
            return $this->save([
                'mobile' => $data['mobile'],
                'password' => $password != '' ? md5($password) : '',
                'reg_source' => 'cashier',
                'app_id' => self::$app_id,
                'grade_id' => $grade_id,
                'nickName' => $data['nick_name']
            ]);
        } else {
            $this->error = '会员已存在';
            return false;
        }
    }

    /**
     * 用户充值
     */
    public function recharge($storeUserName, $source, $data)
    {
        if ($source == 0) {
            return $this->rechargeToBalance($storeUserName, $data);
        } elseif ($source == 1) {
            return $this->rechargeToPoints($storeUserName, $data);
        }
        return false;
    }

    /**
     * 用户充值：余额
     */
    private function rechargeToBalance($storeUserName, $data)
    {
        if (!isset($data['recharge_value']) || $data['recharge_value'] === '' || $data['recharge_value'] < 0) {
            $this->error = '请输入正确的金额';
            return false;
        }
        // 判断是否是正确的数字金额
        if (!is_numeric($data['recharge_value'])) {
            $this->error = '请输入正确的金额';
            return false;
        }
        if ($data['recharge_value'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        // 判断充值方式，计算最终金额
        $money = 0;
        if ($data['mode'] === 'inc') {
            $diffMoney = $this['balance'] + $data['recharge_value'];
            $money = $data['recharge_value'];
        } elseif ($data['mode'] === 'dec') {
            $diffMoney = $this['balance'] - $data['recharge_value'];
            if ($diffMoney < 0) {
                if ($this['balance'] > 0) {
                    $this->error = '减少金额不能大于当前余额';
                } else {
                    $this->error = '余额不能小于0';
                }
                return false;
            }
            $money = -$data['recharge_value'];
        } else {
            $diffMoney = $data['recharge_value'];
            $money = $diffMoney - $this['balance'];
        }
        $maxLimit = 999999999;
        if ($diffMoney > $maxLimit) {
            $this->error = '充值后的余额不能大于' . $maxLimit;
            return false;
        }
        // 更新记录
        $this->transaction(function () use ($storeUserName, $data, $diffMoney, $money) {
            // 新增余额变动记录
            BalanceLogModel::add(SceneEnum::ADMIN, [
                'user_id' => $this['user_id'],
                'card_id' => $this['card_id'],
                'money' => $money,
                'remark' => $data['remark'] ?? '',
            ], [$storeUserName]);
        });
        return true;
    }

    /**
     * 用户充值：积分
     */
    private function rechargeToPoints($storeUserName, $data)
    {
        if (!isset($data['recharge_value']) || $data['recharge_value'] === '' || $data['recharge_value'] < 0) {
            $this->error = '请输入正确的积分数量';
            return false;
        }
        // 判断是否是正确的金额格式（允许两位小数的数字）
        if (!preg_match('/^-?\d+(\.\d{0,2})?$/', $data['value'])) {
            $this->error = '请输入正确的积分数量';
            return false;
        }
        if ($data['recharge_value'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        $points = 0;
        // 判断充值方式，计算最终积分
        if ($data['mode'] === 'inc') {
            $diffMoney = $this['points'] + $data['recharge_value'];
            $points = $data['recharge_value'];
        } elseif ($data['mode'] === 'dec') {
            $diffMoney = $this['points'] - $data['recharge_value'];
            if ($diffMoney < 0) {
                if ($this['points'] > 0) {
                    $this->error = '减少积分不能大于当前积分';
                } else {
                    $this->error = '积分不能小于0';
                }
                return false;
            }
            $points = -$data['recharge_value'];
        } else {
            $diffMoney = $data['recharge_value'];
            $points = $data['recharge_value'] - $this['points'];
        }
        $maxLimit = 999999999;
        if ($diffMoney > $maxLimit) {
            $this->error = '充值后的余额不能大于' . $maxLimit;
            return false;
        }
        // 更新记录
        $this->transaction(function () use ($storeUserName, $data, $diffMoney, $points) {
            $totalPoints = $this['total_points'] + $points <= 0 ? 0 : $this['total_points'] + $points;
            // 更新账户积分
            $this->where('user_id', '=', $this['user_id'])->update([
                'points' => $diffMoney,
                'total_points' => $totalPoints
            ]);
            // 新增积分变动记录
            $scene = $data['mode'] === 'dec' ? PointsLogSceneEnum::DEDUCT : PointsLogSceneEnum::ADMIN;
            PointsLogModel::add([
                'scene' => $scene,
                'user_id' => $this['user_id'],
                'card_id' => $this['card_id'],
                'value' => $points,
                'describe' => vsprintf(PointsLogSceneEnum::data()[$scene]['describe'], [$storeUserName]),
                'remark' => $data['remark'] ?? '',
            ]);
        });
        event('UserGrade', $this['user_id']);
        return true;
    }

    /**
     * 扣减
     */
    public function deduct($storeUserName, $source, $data)
    {
        if ($source == 0) {
            $data['mode'] = 'dec';
            return $this->operationToGiftBalance($storeUserName, $data);
        } elseif ($source == 1) {
            $data['mode'] = 'dec';
            $data['recharge_value'] = $data['value'] ?? 0;
            return $this->rechargeToPoints($storeUserName, $data);
        }
        return false;
    }

    /**
     * 操作：赠送余额
     */
    private function operationToGiftBalance($storeUserName, $data)
    {
        if (!isset($data['value']) || $data['value'] === '' || $data['value'] == 0) {
            $this->error = '请输入正确的金额';
            return false;
        }
        // 判断是否是正确的金额格式（允许两位小数的数字）
        if (!preg_match('/^-?\d+(\.\d{0,2})?$/', $data['value'])) {
            $this->error = '请输入正确的赠送金额';
            return false;
        }
        if ($data['value'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        // 判断充值方式，计算最终金额
        $giftMoney = 0;
        if ($data['mode'] === 'inc') {
            $diffMoney = $this['gift_balance'] + $data['value'];
            $giftMoney = $data['value'];
        } elseif ($data['mode'] === 'dec') {
            $diffMoney = $this['gift_balance'] - $data['value'];
            if ($diffMoney < 0) {
                if ($this['gift_balance'] > 0) {
                    $this->error = '减少金额不能大于当前赠送余额';
                } else {
                    $this->error = '赠送余额不能小于0';
                }
                return false;
            }
            $giftMoney = -$data['value'];
        } else {
            $diffMoney = $data['value'];
            $giftMoney = $diffMoney - $this['gift_balance'];
        }
        $maxLimit = 999999999;
        if ($diffMoney > $maxLimit) {
            $this->error = '充值后的赠送余额不能大于' . $maxLimit;
            return false;
        }
        // 更新记录
        $this->transaction(function () use ($storeUserName, $data, $diffMoney, $giftMoney) {
            // 新增余额变动记录
            $scene = $data['mode'] === 'dec' ? SceneEnum::DEDUCT : SceneEnum::ADMIN;
            BalanceLogModel::add($scene, [
                'user_id' => $this['user_id'],
                'card_id' => $this['card_id'],
                'money' => $giftMoney,
                'gift_money' => $giftMoney,
                'remark' => $data['remark'] ?? '',
            ], [$storeUserName]);
        });
        return true;
    }

    /**
     * 获取用户
     * @param $user_id
     * @return array|mixed
     */
    public static function getUser($user_id)
    {
        return self::where('user_id', $user_id)->find();
    }

    /**
     * 模糊查询会员
     * @param $mobile
     * @return User[]|array|\think\Collection
     */
    public static function search($mobile)
    {
        if (empty($mobile)) {
            return [];
        }
        return (new self)
            ->field("user_id, nickName, mobile")
            ->where(function ($q) use ($mobile) {
                $q->like('mobile', $mobile);
                $q->orLike('user_id', $mobile);
            })
            ->where(['is_delete' => 0])
            ->limit(100) // 最多返回100条 v1.1.0
            ->select();
    }

    /**
     * 会员概括
     */
    public function storeOverview($data)
    {
        $detail = $this->field(['sum(balance) as balance', 'sum(gift_balance) as gift_balance'])
            ->where('is_delete', '=', 0)
            ->find();
        return $detail;
    }

    /**
     * 充值排行榜
     */
    public function getRechargeRank($data)
    {
        // 所需充值场景类型
        $rechargeScenes = [
            BalanceLogSceneEnum::REFUND, // 订单退款
            BalanceLogSceneEnum::RECHARGE,
            BalanceLogSceneEnum::ADMIN,
            BalanceLogSceneEnum::RECHARGE_REVERSE,
            BalanceLogSceneEnum::RECHARGE_REFUND,
            BalanceLogSceneEnum::DEDUCT
        ];
        $rechargeRank = self::alias('a')
            ->leftJoin('user_balance_log ubl', 'a.user_id = ubl.user_id')
            ->field([
                'a.user_id',
                'COALESCE(SUM(ubl.money), 0) as tatal_amount',
                'COALESCE(SUM(ubl.gift_money), 0) as gift_amount',
                '(COALESCE(SUM(ubl.money), 0) - COALESCE(SUM(ubl.gift_money), 0)) as recharge_amount',
                'a.nickName as nickname'
            ])
            ->where('a.is_delete', '=', 0)
            ->whereIn('ubl.scene', $rechargeScenes)
            ->group('a.user_id')
            ->order(['recharge_amount' => 'desc', 'gift_amount' => 'desc', 'user_id' => 'desc'])
            ->paginate($data)?->appends([]);
        return $rechargeRank;
    }

    /**
     * 消费排行榜
     */
    public function getConsumerRank($data)
    {
        $sort = ($data['sort'] ??  0) == 2 ? 'consumption_amount' : 'consumption_num';
        $consumerRank = Order::alias('a')
            ->leftjoin('user b', 'a.user_id = b.user_id')
            ->field('a.user_id, count(a.order_id) as consumption_num, sum(a.pay_price) - sum(a.refund_money) as consumption_amount, b.nickName as nickname')
            ->where('a.order_status', '=', 30)
            ->where('a.pay_status', '=', 20)
            ->where('a.delete_time', '=', 0)
            ->where('a.user_id', '>', 0)
            ->where('b.is_delete', '=', 0)
            ->group('a.user_id')
            ->order($sort, 'desc')
            ->paginate($data)?->append([]);
        return $consumerRank;
    }
}
