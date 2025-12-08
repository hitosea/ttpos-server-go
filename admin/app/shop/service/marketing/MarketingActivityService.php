<?php

namespace app\shop\service\marketing;

use think\facade\Db;
use Endroid\QrCode\QrCode;
use think\facade\Validate;
use Endroid\QrCode\Writer\PngWriter;
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
            'reward_condition_amount' => 'require|float|egt:0',
            'is_open_reward_limit' => 'require|integer|in:0,1',
        ])->message([
            'name.require' => __('活动名称不能为空'),
            'description.require' => __('活动描述不能为空'),
            'start_time.require' => __('开始时间不能为空'),
            'end_time.require' => __('结束时间不能为空'),
            'end_time.gt' => __('结束时间必须大于开始时间'),
        ]);
        if (!$validate->check($data)) {
            return ['code'=>1, 'msg'=>$validate->getError()];
        }
        // 如果开启了奖励次数限制，再校验 reward_limit
        if (isset($data['is_open_reward_limit']) && $data['is_open_reward_limit'] == 1) {
            if (!isset($data['reward_limit']) || !is_numeric($data['reward_limit']) || $data['reward_limit'] < 1) {
                return ['code'=>1, 'msg'=>'奖励次数限制必须大于等于1'];
            }
        }
        // 检查时间合法性
        if ($data['start_time'] < time()) {
            return ['code'=>1, 'msg'=> __('活动开始时间不能早于当前时间')];
        }

        // 检查是否存在同类型的正在进行中的活动
        $startTime = is_numeric($data['start_time']) ? $data['start_time'] : strtotime($data['start_time']);
        $endTime = is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']);
        $exists = MarketingActivity::where('delete_time', 0)
            ->where('is_invalid', 0)
            ->where('type', $data['type'] ?? 0)
            ->where('end_time', '>', time())
            ->where(function ($query) use ($startTime, $endTime) {
                $query->where(function ($q) use ($startTime, $endTime) {
                    // 新活动的开始时间在已有活动的时间范围内
                    $q->where('start_time', '<=', $endTime)->where('end_time', '>=', $startTime);
                })->whereOr(function ($q) use ($startTime, $endTime) {
                    // 新活动的结束时间在已有活动的时间范围内
                    $q->where('start_time', '<=', $endTime)->where('end_time', '>=', $startTime);
                })->whereOr(function ($q) use ($startTime, $endTime) {
                    // 新活动的时间范围完全包含已有活动的时间范围
                    $q->where('start_time', '>=', $startTime)->where('end_time', '<=', $endTime);
                });
            })
            ->find();
        if ($exists) {
            return ['code'=>1, 'msg'=> __('同一个时间内只可有一个同类型活动进行')];
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
                'start_time' => $startTime,
                'end_time' => $endTime,
                'reward_condition_amount' => $data['reward_condition_amount'],
                'is_open_reward_limit' => $data['is_open_reward_limit'],
                'reward_limit' => $data['reward_limit'],
                'is_invalid' => 0,
                'image_base64' => $data['image_base64'] ?? '',
                'reward_type' => $data['reward_type'] ?? 0,
                'reward_value' => $data['reward_value'] ?? 0,
                'is_send_sms' => $data['is_send_sms'] ?? 0,
                'create_time' => time(),
                'update_time' => time(),
            ]);
            // 保存奖品
            foreach ($data['prize_list'] ?? [] as $prize) {
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
            // 
            $appId = request()->appId;
            $url = env('MEMBER_BASE_URL') . "/login?cid={$appId}&aid={$gift->uuid}";
            $qrCode = new QrCode($url);
            return [
                'uuid' => $gift->uuid,
                'qr_code_url' => $url,
                'qr_code' => (new PngWriter())->write($qrCode)->getDataUri()
            ];
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
            // 活动已开始后  只能修改结束时间，开始时间及其他信息不可修改
            if ($gift->start_time <= time()) {
                $endTime = is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']) ?? $gift->end_time;
                // 检查是否存在同类型的正在进行中的活动（排除当前活动）
                $exists = MarketingActivity::where('delete_time', 0)
                    ->where('is_invalid', 0)
                    ->where('type', $gift->type)
                    ->where('uuid', '<>', $uuid)
                    ->where('end_time', '>', time())
                    ->where(function ($query) use ($gift, $endTime) {
                        $query->where(function ($q) use ($gift, $endTime) {
                            // 新活动的开始时间在已有活动的时间范围内
                            $q->where('start_time', '<=', $endTime)
                                ->where('end_time', '>=', $gift->start_time);
                        })->whereOr(function ($q) use ($gift, $endTime) {
                            // 新活动的结束时间在已有活动的时间范围内
                            $q->where('start_time', '<=', $endTime)
                                ->where('end_time', '>=', $gift->start_time);
                        })->whereOr(function ($q) use ($gift, $endTime) {
                            // 新活动的时间范围完全包含已有活动的时间范围
                            $q->where('start_time', '>=', $gift->start_time)
                                ->where('end_time', '<=', $endTime);
                        });
                    })
                    ->find();
                if ($exists) {
                    return ['code'=>1, 'msg'=> __('同一个时间内只可有一个同类型活动进行')];
                }
                $gift->save([
                    'end_time' => $endTime,
                ]);
            } else {
                $startTime = is_numeric($data['start_time']) ? $data['start_time'] : strtotime($data['start_time']) ?? $gift->start_time;
                $endTime = is_numeric($data['end_time']) ? $data['end_time'] : strtotime($data['end_time']) ?? $gift->end_time;
                // 检查是否存在同类型的正在进行中的活动（排除当前活动）
                $exists = MarketingActivity::where('delete_time', 0)
                    ->where('is_invalid', 0)
                    ->where('type', $gift->type)
                    ->where('uuid', '<>', $uuid)
                    ->where('end_time', '>', time())
                    ->where(function ($query) use ($startTime, $endTime) {
                        $query->where(function ($q) use ($startTime, $endTime) {
                            // 新活动的开始时间在已有活动的时间范围内
                            $q->where('start_time', '<=', $endTime)
                                ->where('end_time', '>=', $startTime);
                        })->whereOr(function ($q) use ($startTime, $endTime) {
                            // 新活动的结束时间在已有活动的时间范围内
                            $q->where('start_time', '<=', $endTime)
                                ->where('end_time', '>=', $startTime);
                        })->whereOr(function ($q) use ($startTime, $endTime) {
                            // 新活动的时间范围完全包含已有活动的时间范围
                            $q->where('start_time', '>=', $startTime)
                                ->where('end_time', '<=', $endTime);
                        });
                    })
                    ->find();
                if ($exists) {
                    return ['code'=>1, 'msg'=> __('同一个时间内只可有一个同类型活动进行')];
                }

                // 
                $validate = Validate::rule([
                    'name' => 'require|max:2000',
                    'description' => 'require|max:5000',
                    'start_time' => 'require|string',
                    'end_time' => 'require|string|gt:start_time',
                    'reward_condition_amount' => 'require|float|egt:0',
                    'is_open_reward_limit' => 'require|integer|in:0,1',
                ])->message([
                    'name.require' => __('活动名称不能为空'),
                    'description.require' => __('活动描述不能为空'),
                    'start_time.require' => __('开始时间不能为空'),
                    'end_time.require' => __('结束时间不能为空'),
                    'end_time.gt' => __('结束时间必须大于开始时间'),
                ]);
                if (!$validate->check($data)) {
                    return ['code'=>1, 'msg'=>$validate->getError()];
                }

                $gift->save([
                    'name' => $data['name'] ?? $gift->name,
                    'multi_language_name_uuid' => (new MultiLanguageName)->saveNames($data['name'] ?? $gift->name, $gift->multi_language_name_uuid),
                    'description' => $data['description'] ?? $gift->description,
                    'multi_language_desc_uuid' => (new MultiLanguageName)->saveNames($data['description'] ?? $gift->description, $gift->multi_language_desc_uuid),
                    'start_time' => $startTime,
                    'end_time' => $endTime,
                    'reward_condition_amount' => $data['reward_condition_amount'] ?? $gift->reward_condition_amount,
                    'is_open_reward_limit' => $data['is_open_reward_limit'] ?? $gift->is_open_reward_limit,
                    'reward_limit' => $data['reward_limit'] ?? $gift->reward_limit,
                    'image_base64' => $data['image_base64'] ?? $gift->image_base64,
                    'reward_type' => $data['reward_type'] ?? $gift->reward_type,
                    'reward_value' => $data['reward_value'] ?? $gift->reward_value,
                    'is_send_sms' => $data['is_send_sms'] ?? $gift->is_send_sms,
                    'update_time' => time(),
                ]);
                // 奖品更新：先软删除原奖品，再插入新奖品
                MarketingActivityPrize::where('activity_uuid', $uuid)->update(['delete_time'=>time()]);
                if (isset($data['prize_list']) && !empty($data['prize_list'])) {
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
        $list = $query->clone()->with([
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
            is_open_reward_limit,
            reward_type,
            reward_value,
            is_send_sms,
            create_time, 
            update_time,
            headquarter_uuid
        ')->page($page, $pageSize)->select();
        // 
        $total = $query->count();
        // 
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
        $prizes = MarketingActivityPrize::where('activity_uuid', $uuid)
            ->where('delete_time', 0)
            ->with('couponName')
            ->select();
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
        $page = (int)($params['page'] ?? 1);
        $pageSize = (int)($params['list_rows'] ?? 10);
        // 
        $query =  MarketingActivityRecord::alias('record')
            ->field('
                record.uuid, 
                record.activity_uuid, 
                record.member_uuid, 
                record.reward_count, 
                record.last_reward_time, 
                record.create_time, 
                record.update_time, 
                record.reward_value,
                m.nickname, 
                m.phone, 
                m.id
            ')
            ->join('member m', 'record.member_uuid = m.uuid')
            ->where('record.delete_time', 0)
            ->when(isset($params['activity_uuid']) && !empty($params['activity_uuid']),function($query) use ($params){
                $query->where('record.activity_uuid', $params['activity_uuid']);
            })
            ->when(isset($params['keyword']) && !empty($params['keyword']),function($query) use ($params){
                $query->where(function($query) use ($params){
                    $query->where('m.nickname', 'like', '%'.$params['keyword'].'%')
                        ->whereOr('m.phone', 'like', '%'.$params['keyword'].'%')
                        ->whereOr('m.id', 'like', '%'.$params['keyword'].'%')
                        ->whereOr('m.member_card_no', 'like', '%'.$params['keyword'].'%');
                });
            });
        // 
        $list = $query->clone(true)->order('record.create_time desc')->page($page, $pageSize)->select();
        $total = $query->count();
        return [
            'current_page' => $page,
            'data' => $list,
            'per_page' => $pageSize,
            'total' => $total,
            'last_page' => ceil($total / $pageSize),
        ];
    }

} 