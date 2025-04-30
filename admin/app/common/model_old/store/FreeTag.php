<?php

namespace app\common\model_old\store;

use app\common\model_old\BaseModel;

/**
 * 门店免单标签
 */
class FreeTag extends BaseModel
{
    protected $pk   = 'id';
    protected $name = 'free_tag';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['free_tag_text'];

    /**
     * 标签名
     */
    public static function getFreeTagTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['free_tag'] ?? '');
    }

    /**
     * 列表
     */
    public function getList($app_id, $shop_supplier_id)
    {
        return $this->where('app_id', '=', $app_id)
            ->where('shop_supplier_id', '=', $shop_supplier_id)
            ->select();
    }
}
