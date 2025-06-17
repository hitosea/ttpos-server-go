<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;

/**
 * 邀请有礼记录模型
 */
class MarketingActivityRecord extends BaseModel
{
    protected $name = 'marketing_activity_record';

    // 格式化
    public function getLastRewardTimeAttr($value, $data)
    {
        return date('Y-m-d H:i:s', $data['last_reward_time'] ?? 0);
    }

    /**
     * 关联活动
     */
    public function activity()
    {
        return $this->belongsTo('MarketingActivity', 'activity_uuid', 'uuid');
    }
    
    /**
     * 关联会员
     */
    public function member()
    {
        return $this->belongsTo('app\common\model\user\User', 'member_uuid', 'uuid');
    }
} 