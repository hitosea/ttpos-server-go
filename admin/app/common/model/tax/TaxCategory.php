<?php

namespace app\common\model\tax;

use app\common\library\helper;
use app\common\model\BaseModel;

/**
 *
 */
class TaxCategory extends BaseModel
{
    protected $name = 'tax_category';

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
        return extractLanguage($value ?: $data['name']);
    }

    /**
     * 获取税率
     */
    public function getTaxRateAttr($value, $data = [])
    {
        $taxRate = $value ?: $data['tax_rate'];
        return helper::number2($taxRate);
    }

    /**
     * 列表
     *
     * @return object
     */
    public function getList()
    {
        return (new self())->select();
    }
}
