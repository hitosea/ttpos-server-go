<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
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
        // 查询所有商品成品
        $product = new Product();
        $productList = $product->where('type', Product::TYPE_PRODUCT)->select();
        // 如果没有商品成品，直接返回
        if (count($productList) == 0) {
            return;
        }
        // 查询第一个税类
        $taxCategory = TaxCategory::order('id asc')->find();
        // 遍历商品成品，找到没有过税类的商品
        $productTax = [];
        foreach ($productList as $product) {
            $existingTax = $this->where('product_id', $product['product_id'])->where('tax_category_id', 0)->find();
            if ($existingTax) {
                $this->where('product_id', $product['product_id'])->update(['tax_category_id' => $taxCategory['id']]);
                continue;
            }
            if ($this->where('product_id', $product['product_id'])->where('tax_category_id', '>', 0)->count()) {
                continue;
            }
            // 如果没有过税类，则选择第一个税类
            $productTax[] = [
                'product_id' => $product['product_id'],
                'product_tax_type' => 1, //产品税率类型，1-堂食税类，2-外带税类
                'tax_category_id' => $taxCategory['id'],
            ];
            $productTax[] = [
                'product_id' => $product['product_id'],
                'product_tax_type' => 2, //产品税率类型，1-堂食税类，2-外带税类
                'tax_category_id' => $taxCategory['id'],
            ];
        }
        $this->saveAll($productTax);
    }
}
