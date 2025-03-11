<?php

namespace app\common\model\supplier;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class PrintingRegion extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_printer_region';
    protected $pk = 'id';
    protected $defaultSoftDelete = 0;
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
}