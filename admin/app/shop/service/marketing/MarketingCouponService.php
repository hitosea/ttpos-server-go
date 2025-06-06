<?php

namespace app\shop\service\marketing;

use app\common\model\marketing\MarketingCoupon;
use app\common\model\marketing\MarketingCouponRecord;
use think\facade\Db;
use think\facade\Validate;
use app\common\model\store\MultiLanguageName;
use app\common\model\marketing\MarketingActivity;
use app\common\model\marketing\MarketingActivityPrize;
use app\common\model\marketing\MarketingActivityRecord;

class MarketingCouponService
{
    /**
     * 创建活动
     */
    public function create($data)
    {
        $validate = Validate::rule([
            'name' => 'require|max:2000',
            'sort' => 'require|min:1|max:99',
            'type' => 'string|in:deduction',
            'deduction_type' => 'string|in:taxed',
            'amount' => 'require|float|egt:0',
            'count' => 'require|integer|max:999999',
            'day_start_time' => 'require|string',
            'day_end_time' => 'require|string|egt:day_start_time',
            'requirement' => 'require|string|in:none,marketing',
            'valid_start_time' => 'requireIf:requirement,none|string',
            'valid_end_time' => 'requireIf:requirement,none|string|egt:valid_start_time',
            'valid_days' => 'requireIf:requirement,marketing|integer|egt:0',
        ]);
        if (!$validate->check($data)) {
            return ['code' => 1, 'msg' => $validate->getError()];
        }

        $coupon = new MarketingCoupon();
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
            'valid_start_time' => $data['valid_start_time'],
            'valid_end_time' => $data['valid_end_time'],
            'valid_days' => $data['valid_days'],
            'create_time' => time(),
            'update_time' => time(),
        ]);
        return ['uuid' => $coupon->uuid];
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
            'valid_start_time' => $data['valid_start_time'] ?? $coupon->valid_start_time,
            'valid_end_time' => $data['valid_end_time'] ?? $coupon->valid_end_time,
            'valid_days' => $data['valid_days'] ?? $coupon->valid_days,
            'update_time' => time(),
        ]);
        // ToDo 保存操作记录
        return ['uuid' => $uuid];
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
        if (isset($params['type'])) {
            $query->where('type', '=', $params['type']);
        }
        $page = $params['page'] ?? 1;
        $pageSize = $params['list_rows'] ?? 10;
        $list = $query->order('sort asc, create_time desc')->page($page, $pageSize)->select();
        $total = $query->count();
        return [
            'data' => $list,
            'current_page' => $page,
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
        $coupon = MarketingCoupon::where('uuid', $uuid)->where('delete_time', 0)->find();
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
        $query = MarketingCouponRecord::with(['coupon'])->where('delete_time', 0)->field(['coupon_uuid', 'serial_no','type','count','left_count','create_time'])->order('create_time desc');
        $list = $query->page($page, $pageSize)->select();

        foreach ($list as &$item) {
            $item['coupon_name'] = $item->coupon->name;
            unset($item['coupon_uuid']);
            unset($item['coupon']);
        }
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