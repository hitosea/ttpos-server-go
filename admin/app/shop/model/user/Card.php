<?php

namespace app\shop\model\user;

use app\common\library\helper;
use app\common\model\user\CardRecord;
use app\common\model\user\MemberCard;
use app\common\model\user\Card as CardModel;
use app\shop\model\user\CardRecord as CardRecordModel;
use app\common\enum\user\balanceLog\BalanceLogSceneEnum;
use app\common\model\user\BalanceLog as BalanceLogModel;

/**
 * 会员卡模型
 */
class Card extends CardModel
{
    /**
     * 获取列表记录
     */
    public function getList($data)
    {
        $model = $this;
        if (isset($data['card_name']) && $data['card_name'] != '') {
            $model = $model->like('name', $data['card_name']);
        }
        if ($data['status'] >= 0) {
            $model = $model->where('status', '=', $data['status']);
        }
        $list = $model->where('delete_time', '=', 0)->order(['create_time' => 'desc'])->paginate($data);
        return $list;
    }

    /**
     * 获取列表记录
     */
    public function getDeleteList($data)
    {
        $model = $this->onlyTrashed();

        if (isset($data['card_name']) && $data['card_name']) {
            $model = $model->like('name', $data['card_name']);
        }

        $list = $model->order(['create_time' => 'desc'])->paginate($data);
        return $list;
    }

    /**
     * 发卡
     */
    public function put($data)
    {
        $userIds = $data['user_ids'];
        if (empty($userIds)) {
            $this->error = "请选择会员";
            return false;
        }
        $userIdsArr = array_unique(explode(',', $userIds));
        foreach ($userIdsArr as $userId) {
            $isExist = (new MemberCard())->checkExistByUserId($userId);
            if (!$isExist?->isEmpty()) {
                if ($data['card_id'] == $isExist['card_type_uuid']) {
                    $this->error = "会员已拥有此会员卡";
                    return false;
                    continue;
                }

                // 删除会员的现有会员卡和记录
                MemberCard::destroy(function($query) use ($userId) {
                    $query->where('member_uuid', '=', $userId);
                });
                CardRecordModel::destroy(function($query) use ($userId) {
                    $query->where('member_uuid', '=', $userId);
                });
            }

            $detail = self::detail($data['card_id']);
            $user = (new User)::detail($userId);
            $this->startTrans();
            try {
                //添加会员卡
                $record = [
                    'member_uuid' => $userId,
                    'member_card_type_uuid' => $data['card_id'],
                    'expire' => $detail['expire'] ? (time() + $detail['expire'] * 86400 * 30) : 0,
                    'price' => $detail['money'],
                    'discount' => $detail['discount'] > 0 ? $detail['discount'] : 0,
                    'member_name' => $user['nickname'] ?? '',
                    'member_phone' => $user['phone'] ?? '',
                    'member_no' => $user['member_no'] ?? '',
                    'member_card_type_name' => $detail['name'],
                ];
                $CardRecordModel = new CardRecordModel;
                $CardRecordModel->save($record);
                //
                $memberCardUuid = createUuid();
                $id = (new MemberCard)->add([
                    'uuid' => $memberCardUuid,
                    'card_type_uuid' => $data['card_id'],
                    'member_uuid' => $userId,
                    'expire_time' => $detail['expire'] ? (time() + $detail['expire'] * 86400 * 30) : 0,
                    'discount' => $detail['discount'] > 0 ? $detail['discount'] : 0,
                ]);
                // 会员卡id
                if ($id) {
                    /** @var User $user */
                    $user->setMemberCardId($memberCardUuid);
                }
                // 赠送积分
                if ($detail['open_point'] && $detail['open_point_num']) {
                    /** @var User $user */
                    $user->setIncPoints($detail['open_point_num'], '发会员卡获取积分');
                }
                // 赠送余额
                if ($detail['open_money'] && $detail['open_money_num']) {
                    BalanceLogModel::add(BalanceLogSceneEnum::ADMIN, [
                        'member_uuid' => $user['user_id'],
                        'money' => helper::bcadd($detail['open_money_num'], 0), // v1.0.8显示余额明细
                        'gift_money' => $detail['open_money_num'], // v1.0.8影响到赠送余额，而且不是主账户
                    ], ['order_no' => '后台发放会员卡赠送']);
                }
                $this->commit();
            } catch (\Exception $e) {
                $this->error = $e->getMessage();
                $this->rollback();
                return false;
            }
        }
        return true;
    }

    /**
     * 撤销
     */
    public function cancel($data)
    {
        $CardRecordModel = new CardRecordModel;
        $detail = $CardRecordModel::where('id', '=', $data['order_id'])->find();
        //
        if (!$detail || $detail['delete_time'] != 0) {
            $this->error = "记录不存在";
            return false;
        }
        if ($this->checkUserConsumeRecord($detail['user_id'], $detail['card_id'])) {
            $this->error = "会员卡已使用，无法撤销";
            return false;
        }
        //
        $user = (new User)::detail($detail['member_uuid'], true);
        $memberCard = (new MemberCard())::where('member_uuid', '=', $detail['member_uuid'])->find();
        $this->startTrans();
        try {
            $memberCard->delete();
            $detail->delete();
            // 撤销会员卡id
            /** @var User $user */
            $user?->setMemberCardId(0);
            // 撤销积分
            if ($detail['open_point'] && $detail['open_point_num']) {
                $user->setIncPoints(-$detail['open_point_num'], '撤销会员卡减少积分');
            }
            if ($detail['open_money'] && $detail['open_money_num']) {
                BalanceLogModel::add(BalanceLogSceneEnum::ADMIN, [
                    'member_uuid' => $user['user_id'],
                    'money' => -$detail['open_money_num'],
                    'gift_money' => -$detail['open_money_num'], // v1.0.8影响到赠送余额，而且不是主账户
                ], ['order_no' => '撤销会员卡减少余额']);
            }

            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 处理数据超过最大值时，返回提示信息
     */
    private function alertCardNumber($data)
    {
        $limits = [
            'open_points_num' => ['min' => 0, 'min_message' => '积分不能为负数', 'limit' => 100000000, 'message' => '积分最大可设置为100000000'],
            'open_money_num' => ['min' => 0, 'min_message' => '余额不能为负数', 'limit' => 100000000, 'message' => '余额最大可设置为100000000'],
            'money' => ['min' => 0, 'min_message' => '价格不能为负数', 'limit' => 100000000, 'message' => '价格最大可设置为100000000'],
            'sort' => ['min' => 0, 'min_message' => '排序不能为负数', 'limit' => 999, 'message' => '排序最大可设置为999'],
            'expire' => ['min' => 0, 'min_message' => '有效期不能为负数', 'limit' => 999, 'message' => '有效期最大可设置为999'],
        ];

        foreach ($limits as $key => $value) {
            if (array_key_exists($key, $data)) {
                if ($data[$key] > $value['limit']) {
                    return $value['message'];
                }
                if ($data[$key] < $value['min']) {
                    return $value['min_message'];
                }
            }
        }
        return '';
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        $data = !is_array($data) ? json_decode($data, true) : $data;
        //
        if (mb_strlen($data['card_name']) < 1) {
            $this->error = '会员卡名称不能为空';
            return false;
        }
        if (mb_strlen($data['card_name']) > 50) {
            $this->error = '会员卡名称最大长度限制为50个字符';
            return false;
        }
        if ($text = $this->alertCardNumber($data)) {
            $this->error = $text;
            return false;
        }
        if (mb_strlen($data['content']) < 1) {
            $this->error = '使用须知不能为空';
            return false;
        }
        if (mb_strlen($data['content']) > 200) {
            $this->error = '使用须知最大长度限制为200个字符';
            return false;
        }
        // 校验折扣 discount 值：1、discount参数值大于100 2、discount参数值小于1 3、discount参数值为负数 4、discount参数值为字符串
        if (($data['is_discount'] ?? 0) == 1 && (!isset($data['discount']) || !is_numeric($data['discount']) || $data['discount'] < 1 || $data['discount'] > 100)) {
            $this->error = '折扣参数错误';
            return false;
        }
        if (($data['is_discount'] ?? 0) == 0) {
            $data['discount'] = 0;
        }
        //
        $data['name'] = $data['card_name'] ?? '';
        $data['open_point'] = $data['open_points'] ?? 0;
        $data['open_point_num'] = $data['open_point'] ? ($data['open_points_num'] ?? 0) : 0;
        $data['open_money_num'] = ($data['open_money'] ?? 0) ? ($data['open_money_num'] ?? 0) : 0;
        $data['describe'] = $data['content'] ?? '';
        $data['price'] = $data['money'] ?? 0;
        return $this->save($data);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        $data = !is_array($data) ? json_decode($data, true) : $data;
        //
        if (mb_strlen($data['card_name']) < 1) {
            $this->error = '会员卡名称不能为空';
            return false;
        }
        if (mb_strlen($data['card_name']) > 50) {
            $this->error = '会员卡名称最大长度限制为50个字符';
            return false;
        }
        if ($text = $this->alertCardNumber($data)) {
            $this->error = $text;
            return false;
        }
        if (mb_strlen($data['content']) < 1) {
            $this->error = '使用须知不能为空';
            return false;
        }
        if (mb_strlen($data['content']) > 200) {
            $this->error = '使用须知最大长度限制为200个字符';
            return false;
        }
        // 校验折扣 discount 值：1、discount参数值大于100 2、discount参数值小于1 3、discount参数值为负数 4、discount参数值为字符串
        if (($data['is_discount'] ?? 0) == 1 && (!isset($data['discount']) || !is_numeric($data['discount']) || $data['discount'] < 1 || $data['discount'] > 100)) {
            $this->error = '折扣参数错误';
            return false;
        }
        if (($data['is_discount'] ?? 0) == 0) {
            $data['discount'] = 0;
        }
        unset($data['create_time']);
        unset($data['update_time']);
        //
        $data['name'] = $data['card_name'] ?? '';
        $data['open_point'] = $data['open_points'] ?? 0;
        $data['open_point_num'] = $data['open_point'] ? ($data['open_points_num'] ?? 0) : 0;
        $data['open_money_num'] = $data['open_money'] ? ($data['open_money_num'] ?? 0) : 0;
        $data['describe'] = $data['content'] ?? '';
        $data['price'] = $data['money'] ?? 0;
        return $this->save($data);
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        // 判断该卡下是否存在会员
        if (CardRecordModel::checkExistByRecordId($this['uuid'])) {
            return false;
        }
        return $this->delete();
    }

    /**
     * 设置状态
     */
    public function setStatus($status)
    {
        return $this->save(['status' => $status]);
    }
}
