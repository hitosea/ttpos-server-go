<?php

namespace app\common\model\product;

use think\model\concern\SoftDelete;
use think\model\Pivot;

class RelatedMaterial extends Pivot
{
    use SoftDelete;

    protected $name = 'related_material';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;
    
}
