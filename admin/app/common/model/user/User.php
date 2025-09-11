<?php


namespace app\common\model\user;

use help\ValidateHelp;
use app\common\model\BaseModel;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\user\PointsLog as PointsLogModel;
use app\shop\model\user\BalanceLog as BalanceLogModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum as SceneEnum;
use app\common\library\helper;
use app\common\model\bill\SaleBill;
use app\common\model\order\RechargeOrder;
use think\model\concern\SoftDelete;

/**
 * 用户模型
 */
class User extends BaseModel
{
    use SoftDelete;
    protected $name = 'member';
    protected $pk = 'id';
    protected $defaultSoftDelete = 0;
    protected $delete_time = 'delete_time';
    protected $autoWriteTimestamp = true;

    /**
     * 追加属性
     */
    protected $append = ['user_id', 'points', 'mobile', 'grade_id', 'card_id', 'can_delete', 'pay_money'];

    /**
     * 兼容字段
     */
    public function getUserIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getPointsAttr($value, $data)
    {
        return floatval(helper::bcadd($data['point'] ?? 0, $data['frozen_point'] ?? 0, 2));
    }
    public function getMobileAttr()
    {
        return $this->phone ?: '';
    }
    public function getGradeIdAttr()
    {
        return $this->member_level_uuid ?: 0;
    }
    public function getCardIdAttr()
    {
        return $this->member_card_uuid ?: 0;
    }
    public function getBalanceAttr($value)
    {
        return floatval(helper::bcadd($value ?: 0, $this->frozen_balance ?: 0));
    }
    public function getGiftBalanceAttr($value)
    {
        return floatval(helper::bcadd($value ?: 0, $this->frozen_gift_balance ?: 0));
    }
    public function getCanDeleteAttr($value,$data)
    {
        $memberId = $data['uuid'];
        // 是否有进行中的用餐订单 
        $saleOrderCount = SaleBill::where('uuid', 'IN', function($q) use ($memberId) {
            $prefix = env('DB_PREFIX');
            $q->table($prefix . 'sale_order')->where('consumer_uuid', $memberId)->field('DISTINCT sale_bill_uuid');
        })
        ->where('finish_time', 0)
        ->count();

        // 是否有进行中的充值订单
        $rechargeOrderCount = RechargeOrder::where('member_uuid', $memberId)->where('status', 0)->count();

        return ($saleOrderCount + $rechargeOrderCount) <= 0;
    }

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

    public function getPayMoneyAttr($value, $data)
    {
        return $data['accumulated_consumption_amount'];
    }

    /**
     * 关联会员等级表
     */
    public function grade()
    {
        return $this->belongsTo('app\\common\\model\\user\\Grade', 'member_level_uuid', 'uuid');
    }

    public function memberCard()
    {
        return $this->hasOne('app\\common\\model\\user\\MemberCard', 'member_uuid', 'uuid');
    }

    /**
     * 关联会员卡记录表
     */
    public function cardRecord()
    {
        return $this->hasOne('app\\common\\model\\user\\CardRecord', 'user_id', 'user_id');
    }

    /**
     * 关联会员余额变动记录表
     */
    public function memberBalanceLog()
    {
        return $this->hasMany('app\\common\\model\\user\\BalanceLog', 'member_uuid', 'uuid');
    }

    /**
     * 获取用户信息
     */
    public static function detail($where, $includeDeleted = false)
    {
        $model = new static;
        $filter = $includeDeleted ? [] : ['delete_time' => 0];
        $filter = is_array($where) ? array_merge($filter, $where) : array_merge($filter, ['uuid' => (int) $where]);

        $model = $model->field(['*, (balance + gift_balance) as balance'])->where($filter)->with([
            'grade',
            'memberCard' => [
                'card'
            ]
        ]);
        if ($includeDeleted) {
            $model = $model->withTrashed();
        }
        $info = $model->find();
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
        $filter = ['delete_time' => 0];
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
        $filter = ['delete_time' => 0];
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
        return !!$model->where('member_level_uuid', '=', (int) $gradeId)->value('uuid');
    }

    /**
     * 累积用户总消费金额
     */
    public function setIncPayMoney($money)
    {
        return $this->where('uuid', '=', $this['user_id'])->inc('pay_money', $money)->update();
    }

    /**
     * 累积用户总消费金额
     */
    public function setDecPayMoney($money)
    {
        return $this->where('uuid', '=', $this['user_id'])->dec('pay_money', $money)->update();
    }

    /**
     * 累积用户实际消费的金额 (批量)
     */
    public function onBatchIncExpendMoney($data)
    {
        foreach ($data as $userId => $expendMoney) {
            $this->where(['uuid' => $userId])->inc('expend_money', $expendMoney)->update();
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
            $this->where(['uuid' => $userId])->dec('expend_money', $expendMoney)->update();
        }
        return true;
    }

    /**
     * 累积用户实际消费的金额
     */
    public function IncExpendMoney($data)
    {
        $this->where(['uuid' => $data['user_id']])->inc('expend_money', $data['money'])->update();
        event('UserGrade', $data['user_id']);
        return true;
    }

    /**
     * 累积用户的可用积分数量 (批量)
     */
    public function onBatchIncPoints($data)
    {
        foreach ($data as $userId => $expendPoints) {
            $this->where(['uuid' => $userId])->inc('point', $expendPoints)->update();
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
            $user = $this->where(['uuid' => $userId])->find();
            if ($user) {
                // 计算新的积分值，确保不小于0
                $newPoints = max(0, $user->point - $expendPoints);

                // 更新用户的积分
                $this->where(['uuid' => $userId])
                    ->update(['point' => $newPoints, 'accumulated_get_point' => $user->accumulated_get_point + $expendPoints]);

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
            'member_uuid' => $this['uuid'],
            'value' => $points,
            'describe' => $custom_dec ? $describe : vsprintf(PointsLogSceneEnum::data()[$scene]['describe'], [$describe]),
            'processed' => 1,
        ]);

        // 更新用户可用积分
        $data['point'] = ($this['point'] + $points <= 0) ? 0 : $this['point'] + $points;
        $this->where('uuid', '=', $this['uuid'])->update($data);
        event('UserGrade', $this['uuid']);
        return true;
    }

    /**
     * 累计邀请书
     */
    public function setIncInvite($user_id)
    {
        $this->where('uuid', '=', $user_id)->inc('total_invite')->update();
        event('UserGrade', $user_id);
    }

    /**
     * 更新会员卡id
     */
    public function setMemberCardId($memberCardUuid)
    {
        return $this->save(['member_card_uuid' => $memberCardUuid]);
    }

    /**
     * 关联推荐人
     */
    public function referrer()
    {
        return $this->hasOne('app\\common\\model\\user\\User', 'uuid', 'referrer_uuid');
    }

    /*
     * 添加会员
     */
    public function add($data)
    {
        $grade_id = $data['grade_id'] ?? Grade::getDefaultGradeId();
        $password = $data['password'] ?? '';
        if (empty($grade_id)) {
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
        $user = $this->where('mobile', '=', $data['mobile'])->find();

        if (!$user) {
            return $this->save([
                'mobile' => $data['mobile'],
                'password' => $password != '' ? md5($password) : '',
                'reg_source' => 'cashier',
                'grade_id' => $grade_id,
                'nickname' => $data['nick_name']
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
                'member_uuid' => $this['uuid'],
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
            $diffMoney = $this['point'] + $data['recharge_value'];
            $points = $data['recharge_value'];
        } elseif ($data['mode'] === 'dec') {
            $diffMoney = $this['point'] - $data['recharge_value'];
            if ($diffMoney < 0) {
                if ($this['point'] > 0) {
                    $this->error = '减少积分不能大于当前积分';
                } else {
                    $this->error = '积分不能小于0';
                }
                return false;
            }
            $points = -$data['recharge_value'];
        } else {
            $diffMoney = $data['recharge_value'];
            $points = $data['recharge_value'] - $this['point'];
        }
        $maxLimit = 999999999;
        if ($diffMoney > $maxLimit) {
            $this->error = '充值后的余额不能大于' . $maxLimit;
            return false;
        }
        // 更新记录
        $this->transaction(function () use ($storeUserName, $data, $diffMoney, $points) {
            // 更新账户积分
            $this->where('uuid', '=', $this['uuid'])->update([
                'point' => $diffMoney,
                'accumulated_get_point' => $this['accumulated_get_point'] + $points,
            ]);
            // 新增积分变动记录
            $scene = $data['mode'] === 'dec' ? PointsLogSceneEnum::DEDUCT : PointsLogSceneEnum::ADMIN;
            PointsLogModel::add([
                'scene' => $scene,
                'member_uuid' => $this['uuid'],
                'value' => $points,
                'describe' => vsprintf(PointsLogSceneEnum::data()[$scene]['describe'], [$storeUserName]),
                'remark' => $data['remark'] ?? '',
                'processed' => 1,
            ]);
        });
        event('UserGrade', $this['uuid']);
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
        } elseif ($source == 2) {
            $data['mode'] = 'dec';
            return $this->operationToMainBalance($storeUserName, $data);
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
                'member_uuid' => $this['uuid'],
                'money' => $giftMoney,
                'gift_money' => $giftMoney,
                'remark' => $data['remark'] ?? '',
                'processed' => 1,
            ], [$storeUserName]);
        });
        return true;
    }

    /**
     * 操作：主账户余额
     */
    private function operationToMainBalance($storeUserName, $data)
    {
        if (!isset($data['value']) || $data['value'] === '' || $data['value'] == 0) {
            $this->error = '请输入正确的金额';
            return false;
        }
        // 判断是否是正确的金额格式（允许两位小数的数字）
        if (!preg_match('/^-?\d+(\.\d{0,2})?$/', $data['value'])) {
            $this->error = '请输入正确的主账户余额';
            return false;
        }
        if ($data['value'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        $diffMoney = $this['balance'] - $data['value'];
        if ($diffMoney < 0) {
            if ($this['balance'] > 0) {
                $this->error = '减少金额不能大于当前主账户余额';
            } else {
                $this->error = '主账户余额不能小于0';
            }
            return false;
        }
        $mainMoney = -$data['value'];
        $maxLimit = 999999999;
        if ($diffMoney > $maxLimit) {
            $this->error = '减少后的主账户余额不能大于' . $maxLimit;
            return false;
        }
        // 更新记录
        $this->transaction(function () use ($storeUserName, $data, $diffMoney, $mainMoney) {
            // 新增余额变动记录
            $scene = $data['mode'] === 'dec' ? SceneEnum::DEDUCT : SceneEnum::ADMIN;
            BalanceLogModel::add($scene, [
                'member_uuid' => $this['uuid'],
                'money' => $mainMoney,
                'remark' => $data['remark'] ?? '',
                'processed' => 1,
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
        return self::where('uuid', $user_id)->find();
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
            ->field("uuid as user_id, nickname as nickName, phone as mobile")
            ->where(function ($q) use ($mobile) {
                $q->like('phone', $mobile);
                $q->orLike('uuid', $mobile);
            })
            ->where(['delete_time' => 0])
            ->limit(100) // 最多返回100条 v1.1.0
            ->select();
    }

    /**
     * 会员概括统计
     * @param array $data 查询条件
     * @return array 统计结果
     */
    public function storeOverview($data = [])
    {
        // 获取统计数据
        $stats = $this->where([
            ['delete_time', '=', 0],
        ])->field([
            'COALESCE(SUM(balance), 0) as balance',
            'COALESCE(SUM(gift_balance), 0) as gift_balance',
        ])->findOrEmpty();

        return [
            'balance' => round(floatval($stats['balance'] ?? 0), 2),
            'gift_balance' => round(floatval($stats['gift_balance'] ?? 0), 2),
        ];
    }

    /**
     * 充值排行榜
     */
    public function getRechargeRank($data)
    {
        // 所需充值场景类型
        $rechargeScenes = [
            BalanceLogSceneEnum::RECHARGE,
            BalanceLogSceneEnum::ADMIN,
            BalanceLogSceneEnum::RECHARGE_REVERSE,
            BalanceLogSceneEnum::RECHARGE_REFUND,
            BalanceLogSceneEnum::DEDUCT
        ];
        $rechargeRank = self::alias('a')
            ->leftJoin('member_balance_log ubl', 'a.uuid = ubl.member_uuid and ubl.delete_time = 0')
            ->field([
                'a.id',
                'a.uuid',
                'COALESCE(SUM(ubl.money), 0) as tatal_amount',
                'COALESCE(IF(SUM(ubl.gift_money) > 0, SUM(ubl.gift_money), 0), 0) as gift_amount',
                '(COALESCE(SUM(ubl.money), 0) - COALESCE(SUM(ubl.gift_money), 0)) as recharge_amount',
                'a.nickName as nickname',
                'a.accumulated_consumption_amount'
            ])
            ->withAttr('user_id', function ($value, $data) {
                return $data['id'];
            })
            ->whereIn('ubl.scene', $rechargeScenes)
            ->group('a.uuid')
            ->order(['recharge_amount' => 'desc', 'gift_amount' => 'desc', 'uuid' => 'desc'])
            ->paginate($data)?->appends([]);
        return $rechargeRank;
    }

    /**
     * 消费排行榜
     */
    public function getConsumerRank($data)
    {
        $sort = ($data['sort'] ??  0) == 2 ? 'accumulated_consumption_amount' : 'consumption_count';
        $consumerRank = self::field('id,uuid,nickname,accumulated_consumption_amount,accumulated_consumption_amount as consumption_amount,consumption_count as consumption_num')
            ->where('consumption_count', '>', 0)
            ->withAttr('user_id', function ($value, $data) {
                return $data['id'];
            })
            ->order($sort, 'desc')
            ->paginate($data);
        return $consumerRank;
    }
}
