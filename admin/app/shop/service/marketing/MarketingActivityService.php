<?php

namespace app\shop\service\marketing;

use think\facade\Db;
use think\facade\Validate;
use app\common\model\store\MultiLanguageName;
use app\common\model\marketing\MarketingActivity;
use app\common\model\marketing\MarketingActivityPrize;
use app\common\model\marketing\MarketingActivityRecord;

class MarketingActivityService
{
    /**
     * 创建活动
     */
    public function create($data)
    {
        $validate = Validate::rule([
            'name' => 'require|max:2000',
            'description' => 'require|max:5000',
            'start_time' => 'require|string',
            'end_time' => 'require|string|gt:start_time',
            'reward_condition_amount' => 'require|float|egt:0.01',
            'is_open_reward_limit' => 'require|integer|in:0,1',
            'reward_limit' => 'require|integer|egt:1',
            'prize_list' => 'require|array|min:1',
        ]);
        if (!$validate->check($data)) {
            return ['code'=>1, 'msg'=>$validate->getError()];
        }
        // 检查时间合法性
        if ($data['start_time'] < time()) {
            return ['code'=>1, 'msg'=> __('活动开始时间不能早于当前时间')];
        }
        Db::startTrans();
        try {
            $gift = new MarketingActivity();
            $gift->save([
                'uuid' => createUuid(),
                'name' => $data['name'],
                'multi_language_name_uuid' => (new MultiLanguageName)->saveNames($data['name']),
                'description' => $data['description'],
                'multi_language_desc_uuid' => (new MultiLanguageName)->saveNames($data['description']),
                'start_time' => is_numeric($data['start_time']) ? $data['start_time'] : strtotime($data['start_time']),
                'end_time' => is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']),
                'reward_condition_amount' => $data['reward_condition_amount'],
                'is_open_reward_limit' => $data['is_open_reward_limit'],
                'reward_limit' => $data['reward_limit'],
                'is_invalid' => 0,
                'image_base64' => $data['image_base64'] ?? '',
                'create_time' => time(),
                'update_time' => time(),
            ]);
            // 保存奖品
            foreach ($data['prize_list'] as $prize) {
                (new MarketingActivityPrize())->save([
                    'uuid' => createUuid(),
                    'activity_uuid' => $gift->uuid,
                    'prize_type' => $prize['prize_type'],
                    'prize_uuid' => $prize['prize_uuid'],
                    'create_time' => time(),
                    'update_time' => time(),
                ]);
            }
            Db::commit();
            return ['uuid' => $gift->uuid];
        } catch (\Exception $e) {
            Db::rollback();
            return ['code'=>1, 'msg'=> __('创建失败:').$e->getMessage()];
        }
    }

    /**
     * 编辑活动
     */
    public function update($uuid, $data)
    {
        $gift = MarketingActivity::where('uuid', $uuid)->find();
        if (!$gift) {
            return ['code'=>1, 'msg'=> __('活动不存在')];
        }
   
        Db::startTrans();
        try {
            // 活动已开始后  只能修改结束时间，开始时间及其他信息不可修改
            if ($gift->start_time <= time()) {
                $gift->save([
                    'end_time' => is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']) ?? $gift->end_time,
                ]);
            } else {
                $gift->save([
                    'name' => $data['name'] ?? $gift->name,
                    'multi_language_name_uuid' => (new MultiLanguageName)->saveNames($data['name'] ?? $gift->name, $gift->multi_language_name_uuid),
                    'description' => $data['description'] ?? $gift->description,
                    'multi_language_desc_uuid' => (new MultiLanguageName)->saveNames($data['description'] ?? $gift->description, $gift->multi_language_desc_uuid),
                    'start_time' => is_numeric($data['start_time']) ? $data['start_time'] : strtotime($data['start_time']) ?? $gift->start_time,
                    'end_time' => is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']) ?? $gift->end_time,
                    'reward_condition_amount' => $data['reward_condition_amount'] ?? $gift->reward_condition_amount,
                    'is_open_reward_limit' => $data['is_open_reward_limit'] ?? $gift->is_open_reward_limit,
                    'reward_limit' => $data['reward_limit'] ?? $gift->reward_limit,
                    'image_base64' => $data['image_base64'] ?? $gift->image_base64,
                    'update_time' => time(),
                ]);
                // 奖品更新：先软删除原奖品，再插入新奖品
                MarketingActivityPrize::where('activity_uuid', $uuid)->update(['delete_time'=>time()]);
                if (!empty($data['prize_list'])) {
                    foreach ($data['prize_list'] as $prize) {
                        (new MarketingActivityPrize())->save([
                            'uuid' => createUuid(),
                            'activity_uuid' => $uuid,
                            'prize_type' => $prize['prize_type'],
                            'prize_uuid' => $prize['prize_uuid'],
                        ]);
                    }
                }
            }
            Db::commit();
            return ['uuid' => $uuid];
        } catch (\Exception $e) {
            Db::rollback();
            return ['code'=>1, 'msg'=> __('编辑失败:').$e->getMessage()];
        }
    }

    /**
     * 查询活动列表
     */
    public function getList($params)
    {
        $query = MarketingActivity::where('delete_time', 0);
        // 活动名称
        if (isset($params['name'])) {
            $query->where('name', 'like', '%'.$params['name'].'%');
        }
        // 
        $page = $params['page'] ?? 1;
        $pageSize = $params['list_rows'] ?? 10;
        $list = $query->with([
            'prizes' => function($query) {
                $query->with('couponName')->field('activity_uuid, prize_type, prize_uuid')->where('delete_time', 0);
            },
        ])->order('create_time desc')->field('
            uuid, 
            name,
            type,
            start_time, 
            end_time, 
            reward_condition_amount, 
            reward_limit, 
            is_invalid, 
            image_base64, 
            create_time, 
            update_time
        ')->page($page, $pageSize)->select();
        $total = $query->count();
        return [
            'current_page' => $page,
            'data' => $list,
            'per_page' => $pageSize,
            'total' => $total,
            'last_page' => ceil($total / $pageSize),
        ];
    }

    /**
     * 查询活动详情
     */
    public function getDetail($uuid)
    {
        $gift = MarketingActivity::where('uuid', $uuid)->where('delete_time', 0)->find();
        if (!$gift) return null;
        $prizes = MarketingActivityPrize::where('activity_uuid', $uuid)->where('delete_time', 0)->select();
        return array_merge($gift->toArray(), ['prizes'=>$prizes]);
    }

    /**
     * 失效活动
     */
    public function disable($uuid)
    {
        $gift = MarketingActivity::where('uuid', $uuid)->where('delete_time', 0)->find();
        if (!$gift) return false;
        $gift->save(['is_invalid'=>1, 'update_time'=>time()]);
        return true;
    }

    /**
     * 查询会员奖励记录
     */
    public function getRecord($params)
    {
   
        return MarketingActivityRecord::alias('record')
            ->field('record.uuid, record.activity_uuid, record.member_uuid, m.nickname, m.phone, record.reward_count, record.last_reward_time')
            ->join('member m', 'record.member_uuid = m.uuid')
            ->where('record.delete_time', 0)
            ->when(isset($params['activity_uuid']) && !empty($params['activity_uuid']),function($query) use ($params){
                $query->where('record.activity_uuid', $params['activity_uuid']);
            })
            ->when(isset($params['keyword']) && !empty($params['keyword']),function($query) use ($params){
                $query->where(function($query) use ($params){
                    $query->where('m.nickname', 'like', '%'.$params['keyword'].'%')
                        ->whereOr('m.phone', 'like', '%'.$params['keyword'].'%')
                        ->whereOr('m.id', 'like', '%'.$params['keyword'].'%');
                });
            })
            ->order('record.create_time desc')
            ->select();
    }

    /**
     * 发放奖励
     */
    public function issue($data)
    {
        $validate = Validate::rule([
            'activity_uuid' => 'require',
            'member_uuid' => 'require',
            'order_amount' => 'require|float|egt:0.01',
        ]);
        if (!$validate->check($data)) {
            return false;
        }
        $activity = MarketingActivity::where('uuid', $data['activity_uuid'])->where('delete_time', 0)->find();
        if (!$activity) return false;
        // 校验活动有效期
        $now = time();
        if ($activity->start_time > $now || $activity->end_time < $now || $activity->is_invalid == 1) {
            return false;
        }
        // 校验奖励条件
        if ($data['order_amount'] < $activity->reward_condition_amount) {
            return false;
        }
        // 校验奖励次数
        $record = MarketingActivityRecord::where('activity_uuid', $data['activity_uuid'])
            ->where('member_uuid', $data['member_uuid'])
            ->where('delete_time', 0)
            ->find();
        if ($record && $record->reward_count >= $activity->reward_limit) {
            return false;
        }
        // 发放奖励（这里只做记录，实际发券等业务可扩展）
        if ($record) {
            $record->save([
                'reward_count' => $record->reward_count + 1,
                'last_reward_time' => $now,
                'update_time' => $now
            ]);
        } else {
            (new MarketingActivityRecord())->save([
                'uuid' => uniqid('', true),
                'activity_uuid' => $data['activity_uuid'],
                'member_uuid' => $data['member_uuid'],
                'reward_count' => 1,
                'last_reward_time' => $now,
                'create_time' => $now,
                'update_time' => $now,
                'delete_time' => 0,
            ]);
        }
        // TODO: 实际发券、通知等业务
        return true;
    }

    /**
     * 软删除活动
     */
    public function delete($uuid)
    {
        $gift = MarketingActivity::where('uuid', $uuid)->where('delete_time', 0)->find();
        if (!$gift) return false;
        $now = time();
        $gift->save(['delete_time'=>$now]);
        MarketingActivityPrize::where('activity_uuid', $uuid)->where('delete_time', 0)->update(['delete_time'=>$now]);
        MarketingActivityRecord::where('activity_uuid', $uuid)->where('delete_time', 0)->update(['delete_time'=>$now]);
        return true;
    }
} 