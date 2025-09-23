<?php

namespace app\shop\service\marketing;

use app\common\model\marketing\MarketingCoupon;
use app\common\model\marketing\MarketingCouponRecord;
use think\facade\Validate;

class MarketingCouponService
{
    /**
     * 创建活动
     */
    public function create($data)
    {
        $validate = Validate::rule($this->couponRule());
        if (!$validate->check($data)) {
            return ['code' => 1, 'msg' => $validate->getError()];
        }
        $coupon = new MarketingCoupon();

        $validStartTime = $data['valid_start_time'] ? strtotime($data['valid_start_time']) : 0;
        $validEndTime = $data['valid_end_time'] ? strtotime($data['valid_end_time']) + 86400 - 1 : 0;

        $coupon->save([
            'uuid' => createUuid(),
            'name' => $data['name'],
            'sort' => $data['sort'],
            'type' => $data['type'],
            'deduction_type' => $data['deduction_type'],
            'amount' => $data['amount'],
            'count' => $data['count'],
            'day_start_time' => $data['day_start_time'],
            'day_end_time' => $data['day_end_time'],
            'requirement' => $data['requirement'],
            'valid_start_time' => $validStartTime,
            'valid_end_time' => $validEndTime,
            'valid_days' => $data['valid_days'] ?? 0,
            'create_time' => time(),
            'update_time' => time(),
        ]);
        $this->addRecord(MarketingCouponRecord::RecordTypeCreate, $coupon->uuid, $coupon->count, $coupon->count);
        return ['uuid' => $coupon->uuid];
    }

    private function couponRule(): array
    {
        return [
            'name' => 'require|max:50',
            'sort' => 'require|integer|egt:1|elt:99',
            'type' => 'string|in:deduction',
            'deduction_type' => 'string|in:taxed',
            'amount' => 'require|float|egt:0',
            'count' => 'require|integer|elt:999999',
            'day_start_time' => 'require|dateFormat:H:i',
            'day_end_time' => 'require|dateFormat:H:i|egt:day_start_time',
            'requirement' => 'require|string|in:none,marketing',
            'valid_start_time' => 'requireIf:requirement,none|dateFormat:Y-m-d',
            'valid_end_time' => 'requireIf:requirement,none|dateFormat:Y-m-d|egt:valid_start_time',
            'valid_days' => 'requireIf:requirement,marketing|integer|egt:0',
        ];
    }

    /**
     * 编辑优惠券
     */
    public function update($uuid, $data)
    {
        $coupon = MarketingCoupon::where('uuid', $uuid)->find();
        if (!$coupon) {
            return ['code' => 1, 'msg' => __('优惠券不存在')];
        }
        $validate = Validate::rule($this->couponRule());
        if (!$validate->check($data)) {
            return ['code' => 1, 'msg' => $validate->getError()];
        }
        $oldCount = $coupon->count;


        $validStartTime = $data['valid_start_time'] ? strtotime($data['valid_start_time']) : $coupon->valid_start_time;
        $validEndTime = $data['valid_end_time'] ? strtotime($data['valid_end_time']) + 86400 - 1 : $coupon->valid_end_time;

        $coupon->save([
            'name' => $data['name'] ?? $coupon->name,
            'sort' => $data['sort'] ?? $coupon->sort,
            'type' => $data['type'] ?? $coupon->type,
            'deduction_type' => $data['deduction_type'] ?? $coupon->deduction_type,
            'amount' => $data['amount'] ?? $coupon->amount,
            'count' => $data['count'] ?? $coupon->count,
            'day_start_time' => $data['day_start_time'] ?? $coupon->day_start_time,
            'day_end_time' => $data['day_end_time'] ?? $coupon->day_end_time,
            'requirement' => $data['requirement'] ?? $coupon->requirement,
            'valid_start_time' => $validStartTime,
            'valid_end_time' => $validEndTime,
            'valid_days' => $data['valid_days'] ?? $coupon->valid_days,
            'update_time' => time(),
        ]);
        if ($oldCount > $coupon->count) {
            $this->addRecord(MarketingCouponRecord::RecordTypeDecrease, $coupon->uuid, $oldCount - $coupon->count, $coupon->count);
        } else if ($oldCount < $coupon->count) {
            $this->addRecord(MarketingCouponRecord::RecordTypeIncrease, $coupon->uuid, $coupon->count - $oldCount, $coupon->count);
        }
        return ['uuid' => $uuid];
    }

    public function addRecord($recordType, $couponUuid, $count, $leftCount)
    {
        $lastRecord = MarketingCouponRecord::order('create_time', 'desc')->limit(1)->find();
        $lastRecordSerialNo = null;
        if ($lastRecord && strlen($lastRecord->serial_no) == 16) {
            $lastRecordSerialNo = $lastRecord->serial_no;
        }
        (new MarketingCouponRecord)->save([
            'type' => $recordType,
            'coupon_uuid' => $couponUuid,
            'serial_no' => $this->generateTimeString($lastRecordSerialNo),
            'uuid' => createUuid(),
            'count' => $count,
            'left_count' => $leftCount,
        ]);
    }

    /**
     * 生成时间字符串，格式：yyMMddhhmmssxxxx
     * @param string|null $lastStr 上一个字符串，如果为null则从0001开始
     * @return string 生成的新字符串
     */
    public function generateTimeString($lastStr = null)
    {
        $timeStr = date("ymdHis");
        // 处理序号部分
        $sequence = '0001';
        if ($lastStr !== null) {
            // 提取上一个字符串的序号部分
            $lastSequence = substr($lastStr, -4);
            // 转换为数字并加1
            $nextNum = intval($lastSequence) + 1;
            // 如果超过9999，则从0001重新开始
            if ($nextNum > 9999) {
                $nextNum = 1;
            }
            // 格式化为4位数字，不足补0
            $sequence = str_pad($nextNum, 4, '0', STR_PAD_LEFT);
        }
        // 组合时间字符串和序号
        return $timeStr . $sequence;
    }

    /**
     * 优惠券列表
     */
    public function getList($params)
    {
        $query = MarketingCoupon::where('delete_time', 0);
        // 优惠券名称
        if (isset($params['name'])) {
            $query->where('name', 'like', '%' . $params['name'] . '%');
        }
        // 优惠券类型
        if (isset($params['type']) && in_array($params['type'], ['deduction'])) {
            $query->where('type', '=', $params['type']);
        }
        // requirement
        if (isset($params['requirement'])) {
            $query->where('requirement', '=', $params['requirement']);
        }
        // status
        if (isset($params['status'])) {
            $query->where('status', '=', $params['status']);
        }

        $page = $params['page'] ?? 1;
        $pageSize = $params['list_rows'] ?? 10;
        $list = $query->clone(true)->with([
            'prizes' => function($query) {
                $query->with([
                    'activity' => function($query) {
                        $query->where('delete_time', 0);
                    }
                ])->where('prize_type', 1)->where('delete_time', 0);
            },
        ])->order('sort asc, create_time desc')->page($page, $pageSize)->select();
        $total = $query->count();
        foreach ($list as $item) {
            $result = $this->checkCanDelete($item);
            $item->can_delete = $result['can_delete'];
        }
        return [
            'data' => $list,
            'current_page' => $page,
            'per_page' => $pageSize,
            'total' => $total,
            'last_page' => ceil($total / $pageSize),
        ];
    }

    /**
     * 查询优惠券详情
     */
    public function getDetail($uuid)
    {
        $coupon = MarketingCoupon::where('uuid', $uuid)->find();
        if (!$coupon) return null;
        return $coupon->toArray();
    }

    /**
     * 查询优惠券记录
     */
    public function getRecord($params)
    {
        $page = (int)($params['page'] ?? 1);
        $pageSize = (int)($params['list_rows'] ?? 10);
        // 
        $query = MarketingCouponRecord::alias('record')->field(['record.serial_no', 'record.type as record_type', 'record.count', 'record.left_count', 'record.create_time', 'c.name as coupon_name'])
            ->join('marketing_coupon c', 'c.uuid = record.coupon_uuid')
            ->where('record.delete_time', 0)
            ->when(!empty($params['coupon_name']), function ($query) use ($params) {
                $query->where('c.name', 'like', '%' . $params['coupon_name'] . '%');
            })
            ->when(isset($params['coupon_type']) && in_array($params['coupon_type'], ['deduction']), function ($query) use ($params) {
                $query->where('c.type', '=', $params['coupon_type']);
            })
            ->when(isset($params['record_type']) && in_array((int)$params['record_type'], [
                    MarketingCouponRecord::RecordTypeCreate,
                    MarketingCouponRecord::RecordTypeIncrease,
                    MarketingCouponRecord::RecordTypeDecrease,
                    MarketingCouponRecord::RecordTypeActivityDeduction,
                    MarketingCouponRecord::RecordTypeBonus,
                    MarketingCouponRecord::RecordTypeUsed,
                    MarketingCouponRecord::RecordTypeDelete,
                ]), function ($query) use ($params) {
                $query->where('record.type', '=', $params['record_type']);
            });
        // 检索：注册时间
        if (!empty($params['create_time'][0])) {
            $query = $query->where('record.create_time', 'between', [strtotime($params['create_time'][0]), strtotime($params['create_time'][1]) + 86399]);
        }
        $list = $query->clone(true)->page($page, $pageSize)->order('record.create_time desc, record.left_count asc')->select();
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
     * 删除优惠券
     */
    public function delete($uuid)
    {
        $coupon = MarketingCoupon::with([
            'prizes' => function($query) {
                $query->with([
                    'activity' => function($query) {
                        $query->where('delete_time', 0);
                    }
                ])->where('prize_type', 1)->where('delete_time', 0);
            },
        ])->where('uuid', $uuid)->find();
        if (!$coupon) {
            return ['code' => 1, 'msg' => __('优惠券不存在')];
        }
        $result = $this->checkCanDelete($coupon);
        if (!$result['can_delete']) {
            return ['code' => 1, 'msg' => __($result['msg'])];
        }
        $res = MarketingCoupon::destroy(function($query) use ($uuid) {
            $query->where('uuid', $uuid);
        });
        if (!$res) {
            return ['code' => 1, 'msg' => __('删除优惠券失败')];
        }
        $this->addRecord(MarketingCouponRecord::RecordTypeDelete, $coupon->uuid, $coupon->count, $coupon->count);
    }

    /**
     * 修改优惠券状态
     */
    public function status($uuid, $params)
    {
        if (!in_array($params['status'], [0, 1])) {
            return ['code' => 1, 'msg' => __('状态错误')];
        }
        $coupon = MarketingCoupon::where('uuid', $uuid)->find();
        if (!$coupon) {
            return ['code' => 1, 'msg' => __('优惠券不存在')];
        }
        MarketingCoupon::update(['status' => $params['status']], ['uuid' => $uuid]);
        return ['code' => 0];
    }

    /**
     * 检查优惠券是否可以删除
     *  - 关联奖品，奖品关联活动，活动状态不为2（已结束）则不能删除
     *  - 关联会员优惠券，会员优惠券关联销售订单，销售订单状态为1（已结账）则不能删除
     *  - 关联营销优惠券，营销优惠券关联销售订单，销售订单状态为1（已结账）则不能删除
     */
    public function checkCanDelete(MarketingCoupon $coupon)
    {
        $result = [
            'can_delete' => true,
            'msg' => '',
        ];
        foreach ($coupon->prizes as $prize) {
            if ($prize->activity && $prize->activity->status != 2) {
                $result['can_delete'] = false;
                $result['msg'] = __('活动未结束，不能删除');
                break;
            }
        }
        return $result;
    }
} 