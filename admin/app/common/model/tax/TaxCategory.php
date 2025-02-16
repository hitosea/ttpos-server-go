<?php

namespace app\common\model\tax;

use app\common\library\helper;
use app\common\model\BaseModel;

/**
 *
 */
class TaxCategory extends BaseModel
{
    protected $name = 'tax';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'tax_category_id',
        'name_text',
    ];

    /**
     * 兼容字段
     */
    public function getTaxCategoryIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }

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
