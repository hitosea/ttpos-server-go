<?php
namespace app\common\model\marketing;

use app\common\model\BaseModel;
use app\common\model\store\MultiLanguageName;

/**
 * 邀请有礼活动模型
 */
class MarketingActivity extends BaseModel
{
    // 设置表名
    protected $name = 'marketing_activity';
    
    // 追加属性
    protected $append = [
        'status',
        'status_text'
    ];

    /**
     * 获取状态文字
     */
    public function getStatusAttr($value, $data)
    {
        if ($data['is_invalid'] == 1) {
            return 2;
        }
        $now = time();
        if ($data['start_time'] > $now) {
            return 0;
        } else if ($data['end_time'] < $now) {
            return 2;
        } else {
            return 1;
        }
    }

    /**
     * 获取状态文字
     */
    public function getStatusTextAttr($value, $data)
    {
        if ($data['is_invalid'] == 1) {
            return __('已结束');
        }
        $now = time();
        if ($data['start_time'] > $now) {
            return __('未开始');
        } else if ($data['end_time'] < $now) {
            return __('已结束');
        } else {
            return __('进行中');
        }
    }
    
    /**
     * 关联奖品
     */
    public function prizes()
    {
        return $this->hasMany('MarketingActivityPrize', 'activity_uuid', 'uuid');
    }
    
    /**
     * 关联记录
     */
    public function records()
    {
        return $this->hasMany('MarketingActivityRecord', 'activity_uuid', 'uuid');
    }
    
    /**
     * 关联活动名称
     */
    public function multiLanguageName()
    {
        return $this->hasOne(MultiLanguageName::class, 'uuid', 'multi_language_name_uuid')->field('uuid, name')->bind(['multi_language_name' => 'name']);
    }

    /**
     * 关联活动描述
     */
    public function multiLanguageDesc()
    {
        return $this->hasOne(MultiLanguageName::class, 'uuid', 'multi_language_desc_uuid')->field('uuid, name')->bind(['multi_language_desc' => 'name']);
    }
} 