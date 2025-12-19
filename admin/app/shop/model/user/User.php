<?php

namespace app\shop\model\user;

use help\ValidateHelp;
use app\common\model\user\User as UserModel;
use app\common\enum\user\grade\ChangeTypeEnum;
use app\common\model\user\Grade as GradeModel;
use app\shop\model\user\GradeLog as GradeLogModel;
use app\shop\model\user\PointsLog as PointsLogModel;
use app\shop\model\user\Card as CardModel;
use app\common\enum\user\pointsLog\PointsLogSceneEnum;
use app\shop\model\user\BalanceLog as BalanceLogModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum as SceneEnum;
use app\common\library\helper;
use app\common\model\user\MemberCard;
use app\shop\model\marketing\MarketingActivity;

/**
 * 用户模型
 */
class User extends UserModel
{
    /**
     * 获取当前用户总数
     */
    public function getUserTotal($day = null)
    {
        $model = $this;
        if (!is_null($day)) {
            $startTime = strtotime($day);
            $model = $model->where('create_time', '>=', $startTime)
                ->where('create_time', '<', $startTime + 86400);
        }
        return $model->count();
    }

    /**
     * 获取用户id
     * @return \think\Collection
     */
    public function getUsers($where = null)
    {
        // 获取用户列表
        return $this->where($where)->order(['uuid' => 'asc'])->field(['uuid'])->select();
    }

    /**
     * 获取用户列表
     */
    public static function getList($data)
    {
        $model = new static();
        $data['keyword'] = $data['keyword'] ?? '';
        // 搜索关键词
        if ($data['keyword'] !== '') {
            $keyword = trim($data['keyword']);
            $model = $model->where(function ($query) use ($keyword) {
                $query->like('id|phone|nickname|member_card_no', $keyword);
            });
        }
        // 检索：会员等级
        if (isset($data['grade_id']) && $data['grade_id'] > 0) {
            $model = $model->where('member_level_uuid', '=', (int)$data['grade_id']);
        }
        // 检索：注册时间
        if (!empty($data['reg_date'][0])) {
            $model = $model->where('create_time', 'between', [strtotime($data['reg_date'][0]), strtotime($data['reg_date'][1]) + 86399]);
        }
        // 检索：性别
        if (isset($data['gender']) && $data['gender'] > -1) {
            $model = $model->where('gender', '=', (int)$data['gender']);
        }
        // 检索：排除用户ID
        if (isset($data['exclude_user_id']) && intval($data['exclude_user_id']) > 0) {
            $model = $model->where('uuid', '<>', (int)$data['exclude_user_id']);
        }
        // 获取用户列表
        $paginate = $model->with(['grade', 'memberCard' => ['card'], 'memberBalanceLog', 'referrer'])
            ->field('*, nickname as nickName')
            ->order(['create_time' => 'desc'])
            ->where('is_visitor', 0)
            ->hidden(['open_id', 'union_id', 'password'])
            ->paginate($data);

        foreach ($paginate as $item) {
            $item['card'] = [];
            if ($item['memberCard'] && isset($item['memberCard']['card'])) {
                $item['card'] = $item['memberCard']['card'];
            }
            // 累计充值金额
            $rechargeMoney = '0.00'; 
            // 净充值金额
            $netRechargeMoney = '0.00';
            foreach ($item['memberBalanceLog'] as $log) {
                // 客户端充值、后台充值
                if ($log['scene']['value'] == SceneEnum::RECHARGE || $log['scene']['value'] == SceneEnum::ADMIN) {
                    $addMoney = helper::bcsub($log['money'], $log['gift_money'], 2);
                    // 累计充值金额
                    $rechargeMoney = helper::bcadd($rechargeMoney, $addMoney, 2);
                    // 净充值金额
                    $netRechargeMoney = helper::bcadd($netRechargeMoney, $addMoney, 2);
                }
                // 反结账
                if ($log['scene']['value'] == SceneEnum::RECHARGE_REVERSE) {
                    // 转为正数
                    $money = helper::bcmul($log['money'], -1, 2);
                    $giftMoney = helper::bcmul($log['gift_money'], -1, 2);
                    // 扣减金额
                    $deductMoney = helper::bcsub($money, $giftMoney, 2);
                    // 累计充值金额
                    $rechargeMoney = helper::bcsub($rechargeMoney, $deductMoney, 2);
                    // 净充值金额
                    $netRechargeMoney = helper::bcsub($netRechargeMoney, $deductMoney, 2);
                }
                // 退款
                if ($log['scene']['value'] == SceneEnum::RECHARGE_REFUND) {
                    // 转为正数
                    $deductMoney = helper::bcmul($log['money'], -1, 2);
                    // 净充值金额扣减
                    $netRechargeMoney = helper::bcsub($netRechargeMoney, $deductMoney, 2);
                }
            }
            $item['recharge_money'] = $rechargeMoney;
            $item['net_recharge_money'] = $netRechargeMoney;

            unset($item['memberBalanceLog']);
        }

        return $paginate;
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        if (!$this->can_delete) {
            return false;
        }
        return $this->transaction(function () {
            return $this->delete();
        });
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        $nickName = $data['nick_name'] ?? null;
        $mobile = $data['mobile'] ?? null;
        $password = $data['password'] ?? '';
        $gender = $data['gender'] ?? null;
        $gradeId = $data['grade_id'] ?? GradeModel::getDefaultGradeId();
        $birthday = $data['birthday'] ?? null;
        $cardUuid = $data['card_uuid'] ?? '';
        $cardNumber = $data['card_number'] ?? '';
        $referrerUuid = $data['referrer_uuid'] ?? 0;

        if (!$nickName || !$mobile) {
            $this->error = !$nickName ? '昵称不能为空' : '手机号不能为空';
            return false;
        }

        // 最长20位，不限制格式
        if (mb_strlen($mobile) > 20) {
            $this->error = '手机号长度不能超过20位';
            return false;
        }
        // 密码是否为4-16位纯数字
        if ($password && !ValidateHelp::validateNumber($password)) {
            $this->error = '密码必须为4-16位纯数字';
            return false;
        }
        if (!$cardUuid && $cardNumber) {
            $this->error = '请选择会员卡';
            return false;
        }

        // 检查会员是否已存在
        $user = $this->where('phone', '=', $mobile)->find();
        if ($user) {
            $this->error = '会员已存在';
            return false;
        }

        // 检查会员卡是否存在
        $card = null;
        if ($cardUuid) {
            $card = CardModel::detail($cardUuid);
            if (!$card) {
                $this->error = '会员卡不存在';
                return false;
            }
        }

        // 检查推荐人是否存在
        $activityUuid = 0;
        if ($referrerUuid) {
            $referrer = $this->where('uuid', '=', $referrerUuid)->find();
            if (!$referrer) {
                $this->error = '推荐人不存在';
                return false;
            }
            // 查询当前有效活动
            $activity = (new MarketingActivity())
                ->where('delete_time', 0)
                ->where('is_invalid', 0)
                ->where('start_time', '<=', time())
                ->where('end_time', '>=', time())
                ->order('start_time', 'asc')
                ->order('id', 'asc')
                ->find();
            if ($activity) {
                $activityUuid = $activity['uuid'];
            }
        }

        $this->startTrans();
        try {
            $userUuid = createUuid();
            $user = $this->save([
                'uuid' => $userUuid,
                'nickname' => $nickName,
                'phone' => $mobile,
                'password' => $password ? md5($password) : '',
                'gender' => $gender, //性别
                'member_level_uuid' => $gradeId, //默认等级
                'birthday' => $birthday ? strtotime($birthday) : 0, //生日
                'referrer_uuid' => $referrerUuid, //推荐人
                'activity_uuid' => $activityUuid, //活动
            ]);
            if (!$user) {
                throw new \Exception('添加失败');
            }
            if ($card) {
                $model = new CardModel();
                $result = $model->put([
                    'card_id' => $card['uuid'],
                    'user_ids' => [
                        [ 'uuid' => $userUuid, 'card_number' => $cardNumber ]
                    ],
                ]);
                if (!$result) {
                    throw new \Exception($model->getError());
                }
            }
            $this->commit();
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }

        return true;
    }

    /**
     * 修改记录
     */
    public function editUser($data)
    {
        $password = $data['password'] ?? '';
        // 密码是否为4-16位纯数字
        if ($password && !ValidateHelp::validateNumber($password)) {
            $this->error = "密码必须为4-16位纯数字";
            return false;
        }
        if ($data['mobile']) {
            $mobile = $this->where('phone', '=', $data['mobile'])->where('uuid', '<>', $this['uuid'])->count();
            if ($mobile) {
                $this->error = "手机号已存在";
                return false;
            }
        }
        if ($password) {
            $data['password'] = md5($password);
            $this['password'] = $data['password'];
        } else {
            unset($data['password']);
            unset($this['password']);
        }
        if ($data['birthday']) {
            $data['birthday'] = strtotime($data['birthday']);
        }
        $data['nickname'] = isset($data['nick_name']) ? $data['nick_name'] : $this['nickName'];
        //
        $data['phone'] = $data['mobile'];
        $data['member_level_uuid'] = $data['grade_id'];
        // 
        $cardUuid = $data['card_uuid'] ?? '';
        $cardNumber = $data['card_number'] ?? '';
        if (!$cardUuid && $cardNumber) {
            $this->error = '请选择会员卡';
            return false;
        }
        // 检查会员卡是否存在
        $card = null;
        if ($cardUuid) {
            $card = CardModel::detail($cardUuid);
            if (!$card) {
                $this->error = '会员卡不存在';
                return false;
            }
        }
        if ($cardNumber) {
            $record = (new User)::where('uuid', '<>', $this['uuid'])->where('member_card_no', '=', $cardNumber)->find();
            if ($record) {
                $this->error = '卡号重复';
                return false;
            }
        }

        // 检查推荐人
        $referrerUuid = $data['referrer_uuid'] ?? 0;
        if ($referrerUuid) {
            if ($this->referrer_uuid > 0 && $this->referrer_uuid != $referrerUuid) {
                $this->error = '当前会员已有推荐人';
                return false;
            }
            if ($referrerUuid == $this['uuid']) {
                $this->error = '推荐人不能是自己';
                return false;
            }
            $referrer = $this->where('uuid', '=', $referrerUuid)->find();
            if (!$referrer) {
                $this->error = '推荐人不存在';
                return false;
            }
        }
        return $this->transaction(function () use ($card, $cardNumber, $data, $referrerUuid) {
            if ($card) {
                // 会员未拥有此会员卡，发卡
                $isExist = (new MemberCard())->checkExistByUserId($this['uuid']);
                if ($isExist?->isEmpty() || $card['uuid'] != $isExist['card_type_uuid']) {
                    $model = new CardModel();
                    $result = $model->put([
                        'card_id' => $card['uuid'],
                        'user_ids' => [
                            [ 'uuid' => $this['uuid'], 'card_number' => $cardNumber ]
                        ],
                    ]);
                    if (!$result) {
                        $this->error = $model->getError();
                        return false;
                    }
                }
            }
            $data['member_card_no'] = $cardNumber;
            $data['referrer_uuid'] = $referrerUuid;
            // 使用 force(true) 强制更新所有字段，确保 member_card_no 被更新
            return $this->force(true)->save($data);
        });
    }

    /**
     * 修改用户等级
     */
    public function updateGrade($data)
    {
        if (!isset($data['remark'])) {
            $data['remark'] = '';
        }
        // 变更前的等级id
        $oldGradeId = $this['grade_id'];
        return $this->transaction(function () use ($oldGradeId, $data) {
            // 更新用户的等级
            $status = $this->save(['grade_id' => $data['grade_id']]);
            // 新增用户等级修改记录
            if ($status) {
                (new GradeLogModel)->save([
                    'member_uuid' => $this['uuid'],
                    'old_grade_id' => $oldGradeId,
                    'new_grade_id' => $data['grade_id'],
                    'change_type' => ChangeTypeEnum::ADMIN_USER,
                    'remark' => $data['remark'],
                ]);
            }
            return $status !== false;
        });
    }

    /**
     * 消减用户的实际消费金额
     */
    public function setDecUserExpend($userId, $expendMoney)
    {
        return $this->where(['uuid' => $userId])->dec('expend_money', $expendMoney)->update();
    }

    /**
     * 用户充值
     */
    public function recharge($storeUserName, $source, $data)
    {
        $result = false;
        if ($source == 0) {
            $result = $this->rechargeToBalance($storeUserName, $data['balance']);
        } elseif ($source == 1) {
            $result = $this->rechargeToPoints($storeUserName, $data['points']);
        } elseif ($source == 2) {
            $result = ($this->rechargeToBalance($storeUserName, $data['balance']) && $this->rechargeToPoints($storeUserName, $data['points']));
            if ($result && $this['phone']) {
                // 发送充值短信
                $data['uuid'] = $this['uuid'];
                $data['token'] = request()->header('token');
                $data['language'] = request()->header('language');
                event('UserRecharge', $data);
            }
        }
        return $result;
    }

    /**
     * 用户充值：余额
     */
    private function rechargeToBalance($storeUserName, $data)
    {
        if (!isset($data['money']) || $data['money'] === '' || $data['money'] < 1 || $data['money'] == 0) {
            $this->error = '请输入正确的金额';
            return false;
        }
        // 判断是否是正确的数字金额
        if (!is_numeric($data['money'])) {
            $this->error = '请输入正确的金额';
            return false;
        }
        // 赠送金额
        $data['gift_balance'] = isset($data['gift_balance']) && $data['gift_balance'] > 0 ? $data['gift_balance'] : 0;
        if (!isset($data['gift_balance']) || $data['gift_balance'] < 0) {
            $this->error = '请输入正确的赠送金额';
            return false;
        }
        if (!is_numeric($data['gift_balance'])) {
            $this->error = '请输入正确的赠送金额';
            return false;
        }
        if ($data['money'] > 100000000 || $data['gift_balance'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        // 判断充值方式，计算最终金额
        $money = 0;
        if ($data['mode'] === 'inc') {
            $diffMoney = $this['balance'] + $data['money'] + $data['gift_balance'];
            $money = $data['money'] + $data['gift_balance'];
        } elseif ($data['mode'] === 'dec') {
            if ($this['balance'] == 0) {
                $this->error = '余额不能小于0';
                return false;
            }
            if ($this['balance'] - $data['money'] < 0) {
                $this->error = '减少金额不能大于当前余额';
                return false;
            }
            $diffMoney = $this['balance'] - $data['money'] <= 0 ? 0 : $this['balance'] - $data['money'];
            $money = -$data['money'];
        } else {
            $diffMoney = $data['money'];
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
                'gift_money' => $data['gift_balance'] ?? 0,
                'remark' => $data['remark'],
                'processed' => 1,
            ], [$storeUserName]);
        });
        return true;
    }

    /**
     * 用户充值：积分
     */
    private function rechargeToPoints($storeUserName, $data)
    {
        // 如果充值为0，则直接返回
        if (isset($data['value']) && ($data['value'] == 0 || $data['value'] == '')) {
            return true;
        }
        $data['value'] = isset($data['value']) && $data['value'] > 0 ? $data['value'] : 0;
        if (!isset($data['value']) || $data['value'] < 0) {
            $this->error = '请输入正确的积分数量';
            return false;
        }
        // 判断是否是正确的数字金额
        if (!is_numeric($data['value'])) {
            $this->error = '请输入正确的金额';
            return false;
        }
        if ($data['value'] > 100000000) {
            $this->error = '不能大于100000000';
            return false;
        }
        $points = 0;
        // 判断充值方式，计算最终积分
        if ($data['mode'] === 'inc') {
            $diffMoney = $this['point'] + $data['value'];
            $points = $data['value'];
        } elseif ($data['mode'] === 'dec') {
            if ($this['point'] == 0) {
                $this->error = '积分不能小于0';
                return false;
            }
            if ($this['point'] - $data['value'] < 0) {
                $this->error = '减少积分不能大于当前积分';
                return false;
            }
            $diffMoney = $this['point'] - $data['value'] <= 0 ? 0 : $this['point'] - $data['value'];
            $points = -$data['value'];
        } else {
            $diffMoney = $data['value'];
            $points = $data['value'] - $this['point'];
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
            PointsLogModel::add([
                'member_uuid' => $this['uuid'],
                'scene' => PointsLogSceneEnum::RECHARGE_GIVE,
                'value' => $points,
                'describe' => vsprintf("后台管理员 [%s] 操作", [$storeUserName]),
                'remark' => $data['remark'],
                'processed' => 1,
            ]);
        });
        event('UserGrade', $this['uuid']);
        return true;
    }


    /**
     * 获取用户统计数量
     */
    public function getUserData($startDate, $endDate, $type)
    {
        $model = $this;
        if (!is_null($startDate)) {
            $model = $model->where('create_time', '>=', strtotime($startDate));
        }
        if (is_null($endDate)) {
            $model = $model->where('create_time', '<', strtotime($startDate) + 86400);
        } else {
            $model = $model->where('create_time', '<', strtotime($endDate) + 86400);
        }
        if ($type == 'user_total' || $type == 'user_add') {
            return $model->count();
        } else if ($type == 'user_pay') {
            return $model->where('pay_money', '>', '0')->count();
        } else if ($type == 'user_no_pay') {
            return $model->where('pay_money', '=', '0')->count();
        }
        return 0;
    }

    /**
     * 提现打款成功：累积提现余额
     */
    public static function totalMoney($user_id, $money)
    {
        $model = self::detail($user_id);
        return $model->save([
            'freeze_money' => $model['freeze_money'] - $money,
            'cash_money' => $model['cash_money'] + $money,
        ]);
    }

    /**
     * 提现驳回：解冻用户余额
     */
    public static function backFreezeMoney($user_id, $money)
    {
        $model = self::detail($user_id);
        return $model->save([
            'balance' => $model['balance'] + $money,
            'freeze_money' => $model['freeze_money'] - $money,
        ]);
    }
}
