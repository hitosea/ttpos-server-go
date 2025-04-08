<?php

namespace app\common\model\order;

use app\common\model\BaseModel;

class RechargeOrder extends BaseModel
{
    protected $name = 'member_recharge_order';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}