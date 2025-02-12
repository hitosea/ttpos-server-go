<?php

namespace app\common\model\buffet;

use app\common\model\BaseModel;
use app\common\model\tax\TaxCategory;

/**
 *
 */
class BuffetTax extends BaseModel
{
    protected $name = 'buffet_tax';
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
     * 没有过税类的自助餐默认选择第一个税类
     */
    public function getBuffetDefaultTaxCategory()
    {
        // 查询所有自助餐
        $buffet = new Buffet();
        $buffetList = $buffet->where('status', 1)->select();
        // 如果没有自助餐，直接返回
        if (count($buffetList) == 0) {
            return;
        }
        // 查询第一个税类
        $taxCategory = TaxCategory::order('id asc')->find();
        // 遍历自助餐，找到没有过税类的商品
        $buffetTax = [];
        foreach ($buffetList as $buffet) {
            $existingTax = $this->where('buffet_id', $buffet['id'])->where('tax_category_id', 0)->find();
            if ($existingTax) {
                $this->where('buffet_id', $buffet['id'])->update(['tax_category_id' => $taxCategory['id']]);
                continue;
            }
            if ($this->where('buffet_id', $buffet['id'])->where('tax_category_id', '>', 0)->count()) {
                continue;
            }
            // 如果没有过税类，则选择第一个税类
            $buffetTax[] = [
                'buffet_id' => $buffet['id'],
                'buffet_tax_type' => 1, //自助餐税率类型，1-堂食税类
                'tax_category_id' => $taxCategory['id'],
            ];
        }
        $this->saveAll($buffetTax);
    }
}
