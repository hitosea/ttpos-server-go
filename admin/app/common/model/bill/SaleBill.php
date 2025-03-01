<?php

namespace app\common\model\bill;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class SaleBill extends BaseModel
{
    use SoftDelete;

    const SALE_BILL_STATUS_PENNING = 0; // 等待付款
    const SALE_BILL_STATUS_COMPLETE = 1; // 已付款
    const SALE_BILL_STATUS_CANCEL = 2; // 已取消

    protected $name = 'sale_bill';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}
