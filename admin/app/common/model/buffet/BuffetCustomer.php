<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 *
 */
class BuffetCustomer extends BaseModel
{
    use SoftDelete;
    protected $name = 'buffet_customer_type_price';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'buffet_id',
        'customer_type_id',
        'name_text',
    ];

    /**
     * 兼容字段
     */
    public function getBuffetIdAttr($value, $data = [])
    {
        return $this->buffet_package_uuid ?: 0;
    }
    public function getCustomerTypeIdAttr($value, $data = [])
    {
        return $this->customer_type_uuid ?: 0;
    }

    /**
     * 获取名称
     */
    public function getNameTextAttr($value, $data = [])
    {
        $name = (new CustomerType)->where('uuid', $data['customer_type_uuid'])->value('name');
        return extractLanguage($name);
    }
}
