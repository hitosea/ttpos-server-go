<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 *
 */
class BuffetProduct extends BaseModel
{
    use SoftDelete;
    protected $name = 'buffet_product';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['buffet_id', 'product_id', 'limit_num'];

    /**
     * 兼容字段
     */
    public function getBuffetIdAttr($value, $data = [])
    {
        return $this->buffet_package_uuid ?: 0;
    }
    public function getProductIdAttr($value, $data = [])
    {
        return $this->product_package_uuid ?: 0;
    }
    public function getLimitNumAttr($value, $data = [])
    {
        return $this->limit ?: 0;
    }

    /**
     * 关联自助餐
     */
    public function buffet()
    {
        return $this->belongsTo('app\\common\\model\\buffet\\Buffet', 'uuid', 'buffet_package_uuid');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo('app\\common\\model\\product\\Product', 'product_package_uuid', 'uuid');
    }
}
