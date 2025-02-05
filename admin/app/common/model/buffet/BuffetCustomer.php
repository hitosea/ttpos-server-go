<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;

/**
 *
 */
class BuffetCustomer extends BaseModel
{
    protected $name = 'buffet_customer';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'name_text',
    ];

    /**
     * 获取名称
     */
    public function getNameTextAttr($value, $data = [])
    {
        $name = (new CustomerType)->where('id', $data['customer_type_id'])->value('name');
        return extractLanguage($name);
    }
}
