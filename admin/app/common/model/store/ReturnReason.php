<?php

namespace app\common\model\store;

use app\common\model\BaseModel;

/**
 * 门店退菜原因
 */
class ReturnReason extends BaseModel
{
    protected $name = 'return_food_reason';
    protected $pk   = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['reason'];

    /**
     * 关联多语言
     */
    public function multiLanguageName()
    {
        return $this->belongsTo('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 原因
     */
    public function getReasonAttr($value, $data = [])
    {
        $langUuid = $this->multi_language_name_uuid;
        if (!$langUuid) {
            return '';
        }
        return (new MultiLanguageName)->getNames($langUuid);
    }

    /**
     * 原因 set
     * @param mixed $value
     * @return string
     */
    public function setReasonAttr($value)
    {
        return is_array($value) ? json_encode($value, JSON_UNESCAPED_UNICODE) : (string) $value;
    }

    /**
     * 列表
     */
    public function getList($app_id, $shop_supplier_id)
    {
        return $this->select();
    }
}
