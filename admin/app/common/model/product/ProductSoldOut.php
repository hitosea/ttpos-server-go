<?php

namespace app\common\model\product;

use app\common\model\BaseModel;

/**
 * 商品售罄模型
 */
class ProductSoldOut extends BaseModel
{
    protected $name = 'product_sold_out';
    protected $pk = 'id';


    /**
     * 获取商品数据
     */
    public static function getProductNameTextAttr($value, $data)
    {
        return extractLanguage($data['product_name'] ?? '');
    }

    /**
     * 获取单位规格名称
     */
    public function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['spec_name'] ?? '');
    }

    /**
     * 获取商品简易列表
     */
    public function productSkuList($params)
    {
        $search = $params['search'] ?? "";
        //
        return $this->alias('a')
            ->leftJoin('product product', 'product.product_id = a.product_id')
            ->leftJoin('product_sku sku', 'sku.product_sku_id = a.product_sku_id')
            ->field([
                'product.product_id',
                'product.product_name',
                'sku.product_sku_id',
                'sku.spec_name',
            ])
            ->when($search, function ($q) use ($search) {
                $q->jsonLike('product.product_name', trim($search));
            })
            ->where('product.is_delete', '=', 0)
            ->where('product.product_type', '=', 1)
            ->where('product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where('product.product_status', '=', 10)
            ->order(['product.product_sort', 'product.product_id' => 'desc'])
            ->paginate($params)
            ->append(['product_name_text', 'spec_name_text'])
            ->hidden(['product_name', 'product_unit', 'spec_name'])
            ->toArray();
    }

    /**
     * 添加
     */
    public function add($data, $user)
    {
        foreach ($data as $key => $val) {
            $data[$key]['app_id'] = $user['app_id'];
            $data[$key]['shop_supplier_id'] = $user['shop_supplier_id'];
        }
        foreach ($this->select() as $m) {
            $m->delete();
        }
        return $this->saveAll($data);
    }

    /**
     * 添加
     */
    public function addSoldOut($data, $user)
    {
        foreach ($data as $val) {
            if ($val['is_sold_out'] == 1) {
                if (!self::where('product_id', $val['product_id'])->where('product_sku_id', $val['product_sku_id'])->find()) {
                    (new self)->save([
                        'product_id' => $val['product_id'],
                        'product_sku_id' => $val['product_sku_id']
                    ]);
                }
            } else {
                self::where('product_id', $val['product_id'])->where('product_sku_id', $val['product_sku_id'])->delete();
            }
        }
    }

    /**
     * 取消
     */
    public function cancelSoldOut($data)
    {
        return self::where('product_id', $data['product_id'])->where('product_sku_id', $data['product_sku_id'])->delete();
    }

    /**
     * 全部取消
     */
    public function cancelAllSoldOut($shop_supplier_id)
    {
        return self::where('shop_supplier_id', $shop_supplier_id)->delete();
    }
}
