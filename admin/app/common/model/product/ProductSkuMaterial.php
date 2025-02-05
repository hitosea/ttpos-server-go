<?php

namespace app\common\model\product;

use app\common\model\BaseModel;

/**
 * 产品规格材料关联模型
 */
class ProductSkuMaterial extends BaseModel
{
    protected $name = 'product_sku_material';

    protected $updateTime = false;

    /**
     * 使用库存数量
     */
    public function getMaterialNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 关联材料产品
     */
    public function materialProduct()
    {
        return $this->belongsTo('app\\common\\model\\product\\Product', 'material_id', 'product_id')->field('product_id, product_name, product_unit, product_material_stock')->with(['sku']);
    }

    /**
     * 关联材料产品SKU
     */
    public function materialSku()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductSku', 'product_sku_id', 'product_sku_id');
    }
}
