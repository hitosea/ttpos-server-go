<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;
use think\model\concern\SoftDelete;

/**
 * 库存记录模型
 */
class ErpWarehouseOutFormItem extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_out_form_item';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = false;

    /**
     * 关联商品bom
     */
    public function productBom()
    {
        return $this->belongsTo(ProductBom::class, 'product_bom_uuid', 'uuid');
    }

    /**
     * 关联材料
     */
    public function material()
    {
        return $this->belongsTo(Material::class, 'material_uuid', 'uuid');
    }
}
