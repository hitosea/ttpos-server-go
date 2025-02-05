<?php

namespace app\shop\validate;

use app\common\validate\BaseValidate;

/**
 * 桌台导入验证类
 */
class TableImportsValidate extends  BaseValidate
{
    //定义验证规则
    protected $rule = [
        'table_no' => 'require',
        'type_name' => 'require',
        'type_id' => 'require',
        'area_name' => 'require',
        'area_id' => 'require',
        'sort' => 'require',
        'row' => 'require',
    ];

    protected $message = [
        'table_no.require' => '桌位名称不能为空',
        'type_name.require' => '所属类型不能为空',
        'type_id.require' => '所属类型不能为空',
        'area_name.require' => '所属区域不能为空',
        'area_id.require' => '所属区域不能为空',
        'sort.require' => '排序不能为空',
        'row.require' => '行号不能为空',
    ];

    protected $scene = [
        'get' => [
            'table_no', 
            'type_name',
            'area_name',
            'sort',
            'row',
        ],
        'save' => [
            'table_no', 
            'sort',
            'row',
            'type_id',
            'area_id',
        ],
    ];
}
