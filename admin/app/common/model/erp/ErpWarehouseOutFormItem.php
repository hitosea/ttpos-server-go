<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
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
}
