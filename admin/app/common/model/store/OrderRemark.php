<?php

namespace app\common\model\store;

use think\facade\Db;
use app\common\model\BaseModel;
use app\common\model\store\MultiLanguageName;
use think\model\concern\SoftDelete;

/**
 * 整单备注
 */
class OrderRemark extends BaseModel
{
    // 属性定义
    use SoftDelete;
    protected $name = 'order_remark';
    protected $pk   = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;
    
    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['order_remark'];

    /**
     * 关联多语言
     */
    public function multiLanguageName()
    {
        return $this->belongsTo('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 整单备注名
     */
    public function getOrderRemarkAttr($value, $data = [])
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
