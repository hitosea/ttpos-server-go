<?php

namespace app\common\model_old\product;

use app\common\model_old\BaseModel;

/**
 * 产品加料材料关联模型
 */
class ProductFeedMaterial extends BaseModel
{
    protected $name = 'product_feed_material';

    /**
     * 使用库存数量
     */
    public function getMaterialNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 材料信息（产品表）
     */
    public function materialProduct()
    {
        return $this->belongsTo('app\\common\\model_old\\product\\Product', 'material_id', 'product_id')->field('product_id, product_name, product_unit, product_material_stock');
    }

    /**
     * 关联材料产品SKU
     */
    public function materialFeed()
    {
        return $this->belongsTo('app\\common\\model_old\\product\\ProductFeed', 'product_feed_id', 'product_feed_id');
    }
}
