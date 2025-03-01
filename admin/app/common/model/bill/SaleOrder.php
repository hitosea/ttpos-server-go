<?php

namespace app\common\model\bill;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class SaleOrder extends BaseModel {
    use SoftDelete;

    protected $name = 'sale_order';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 关联订单
     */
    public function saleBill()
    {
        return $this->belongsTo(SaleBill::class, 'sale_bill_uuid', 'uuid');
    }
}