<?php

namespace app\common\model_old\shop;

use help\QueueHelp;
use app\common\library\helper;
use app\common\model_old\BaseModel;

/**
 * 店铺金额变更记录表
 */
class AccountLog extends BaseModel
{
    protected $name = 'shop_account_log';
    protected $pk = 'id';
}
