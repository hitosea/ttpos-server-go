<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 外送台账结清数据
 */
class DeliveryLedgerSettle extends BaseModel
{
    use SoftDelete;
    protected $name = 'delivery_ledger_settle';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}
