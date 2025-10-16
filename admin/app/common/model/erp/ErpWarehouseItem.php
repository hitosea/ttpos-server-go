<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class ErpWarehouseItem extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_item';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

    /**
     * 关联仓库
     */
    public function warehouse()
    {
        return $this->belongsTo(ErpWarehouse::class, 'warehouse_uuid', 'uuid');
    }
}