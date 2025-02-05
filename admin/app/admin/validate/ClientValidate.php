<?php

namespace app\admin\validate;

use app\common\validate\BaseValidate;


class ClientValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule = [
        'id|ID' => 'require',
        'uuid|UUID' => 'require',
        'type|类型' => 'require|in:1,2,3,4,5',
        'brand|品牌' => 'require|in:1,2,3,4,5',
        'version_number|版本号' => 'string',
        'version_name|版本名称' => 'string',
        'forced_update|强制更新' => 'require|in:0,1',
        'package_url|包地址' => 'require',
        'update_log|更新日志' => 'require',
    ];

    protected $message = [
        'type.require' => '请指定类型',
        'forced_update.require' => '强制更新必填',
        'update_log.require' => '更新日志必填',
    ];

    protected $scene = [
        'id' => ['id'],
        'uuid' => ['uuid'],
        'add' => ['type', 'brand', 'version_number', 'forced_update', 'package_url', 'update_log', 'version_name'],
        'publish' => ['id' ,'update_log', 'forced_update'],
    ];
}
