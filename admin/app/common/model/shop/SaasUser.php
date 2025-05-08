<?php

namespace app\common\model\shop;

use think\Model;
use think\facade\Cache;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 商家用户模型
 */
class SaasUser extends BaseModel
{
    use SoftDelete;
    protected $name = 'company_staff';
    protected $pk = 'id';
    protected $append = ['shop_user_id'];
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
}
