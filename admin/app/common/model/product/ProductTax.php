<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use app\common\model\buffet\Buffet;
use app\common\model\tax\TaxCategory;

/**
 *
 */
class ProductTax extends BaseModel
{
    protected $name = 'tax';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'tax_rate',
    ];

    /**
     * 关联税类分类
     */
    public function taxCategory()
    {
        return $this->belongsTo(TaxCategory::class, 'tax_category_id', 'id');
    }

    /**
     * 获取比例
     */
    public function getTaxRateAttr($value, $data = [])
    {
        $taxCategory = TaxCategory::where('id', $data['tax_category_id'] ?? 0)->find();
        if ($taxCategory) {
            return $taxCategory['tax_rate'] ?? 0;
        }
    }

    /**
     * 列表
     *
     * @return object
     */
    public function getList($params)
    {
        return (new self())->paginate($params);
    }

    /**
     * 是否使用税类
     */
    public function isUseTax($tax_category_id)
    {
        $model = new self();
        $data = $model->where('tax_category_id', $tax_category_id)->find();
        if ($data) {
            return true;
        }
        return false;
    }

    /**
     * 没有过税类的商品成品默认选择第一个税类
     */
    public function getProductDefaultTaxCategory()
    {
        // 查询第一个税类
        $taxCategory = TaxCategory::order('id asc')->find();
        $productTax = [];
        $buffetTax = [];

        // 查询没有税类的商品
        $product = new Product();
        $productList = $product->where('dine_tax_uuid', 0)->whereOr('takeout_tax_uuid', 0)->select();        
        foreach ($productList as $productItem) {
            $productTax[] = [
                'id' => $productItem['id'],
                'dine_tax_uuid' => $taxCategory['uuid'],
                'takeout_tax_uuid' => $taxCategory['uuid'],
            ];
        }

        // 查询没有税类的自助餐
        $buffet = new Buffet();
        $buffetList = $buffet->where('tax_uuid', 0)->select();
        foreach ($buffetList as $buffetItem) {
            $buffetTax[] = [
                'id' => $buffetItem->getData('id'),
                'tax_uuid' => $taxCategory['uuid'],
            ];
        }

        if (!empty($productTax)) {
            $product->saveAll($productTax);
        }
        if (!empty($buffetTax)) {
            $buffet->saveAll($buffetTax);
        }
    }
}
