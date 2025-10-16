<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class ErpWarehouse extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
}