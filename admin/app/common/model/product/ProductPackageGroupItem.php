<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\ProductBom;

/**
 * 套餐组商品
 */
class ProductPackageGroupItem extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_package_group_item';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $append = [];

    /**
     * 关联商品
     */
    public function productBom()
    {
        return $this->hasOne(ProductBom::class, 'uuid', 'product_bom_uuid');
    }
}
