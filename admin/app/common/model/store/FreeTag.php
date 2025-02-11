<?php

namespace app\common\model\store;

use think\facade\Db;
use app\common\model\BaseModel;
use app\common\model\store\MultiLanguageName;

/**
 * 门店免单标签
 */
class FreeTag extends BaseModel
{
    // 属性定义
    protected $name = 'free_reason';
    protected $pk   = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['free_tag'];

    /**
     * 关联多语言
     */
    public function multiLanguageName()
    {
        return $this->belongsTo('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 标签名
     */
    public function getFreeTagAttr($value, $data = [])
    {
        $langUuid = $this->multi_language_name_uuid;
        if (!$langUuid) {
            return '';
        }
        return (new MultiLanguageName)->getNames($langUuid);
    }

    /**
     * 列表
     */
    public function getList($app_id, $shop_supplier_id)
    {
        return $this->select();
    }
}
