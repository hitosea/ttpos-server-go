<?php

namespace app\common\model\erp;

use help\StringHelp;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\ProductSku;

/**
 * 库存记录模型
 */
class ErpWarehouseOutFormItem extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_out_form_item';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}
