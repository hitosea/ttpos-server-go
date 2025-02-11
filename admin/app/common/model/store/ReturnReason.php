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
    protected $append = ['reason_text'];

    /**
     * 原因
     */
    public static function getReasonTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['reason'] ?? '');
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
