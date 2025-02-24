<?php

namespace app\common\model\product;

use think\model\Pivot;

class RelatedMaterial extends Pivot
{
    protected $name = 'related_material';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    
}
