<?php

namespace app\common\model\supplier;

use app\common\model\BaseModel;
use app\common\model\product\Product;
use think\model\concern\SoftDelete;

class PrintingProduct extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_printer_product_item';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;

    /**
     * 关联商品
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_package_uuid', 'uuid');
    }
}