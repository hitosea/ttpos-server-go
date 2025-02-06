<?php

namespace app\menu\model\product;

use think\facade\Db;
use think\facade\Env;
use app\common\model\settings\Setting;
use app\common\model\product\Product as ProductModel;

/**
 * 商品模型
 */
class Product extends ProductModel
{
    /**
     * 获取商品列表
     */

    public function list($params)
    {
        Db::connect()->execute("SET SESSION sql_mode = ''");
        $is_special = $params['is_special'] ?? 0;
        $category_id = $params['category_id'] ?? 0;
        $special_id = $is_special ? ($params['category_id'] ?? 0) : 0;
        $search = $params['search'] ?? "";
        $settingInfo = Setting::getItem('cashier', $params['shop_supplier_id']);
        //
        $prefix = Env::get('DB_PREFIX');
        // 筛选条件
        $result = $this->alias($prefix . 'product')
            ->leftJoin('category c', $prefix . 'product.category_id = c.category_id')
            ->leftJoin("
                (
                    select product_sku.product_id, ifnull(sum(product_sku.stock_num),0) as sku_stock_num
                    from {$prefix}product_sku as product_sku
                    where not EXISTS(select 1 from {$prefix}product_sold_out as sold_out where sold_out.product_sku_id = product_sku.product_sku_id)
                    group by product_sku.product_id
                ) product_sku
            ", $prefix . 'product.product_id = product_sku.product_id')
            ->field([
                "IF(c.parent_id > 0, c.parent_id, {$prefix}product.category_id) as category_id",    // 都返回一级分类ID
                $prefix . 'product.limit_num',
                $prefix . 'product.selling_point',
                $prefix . 'product.product_id',
                $prefix . 'product.product_name',
                $prefix . 'product.product_price',
                'product_sku.sku_stock_num as product_stock',
                $prefix . 'product.spec_type',
                $prefix . 'product.special_id',
                $prefix . 'product.sales_actual',
                $prefix . 'product.sales_initial',
                $prefix . 'product.product_unit',
                $prefix . 'product.product_feed',
                $prefix . 'product.product_attr',
                $prefix . 'product.is_show_kitchen',
                $prefix . 'product.is_show_h5',
                $prefix . 'product.product_sort',
            ])->with([
                'image.file',
                'category' => function ($q) {
                    $q->field('category_id,is_special,name,parent_id,type')->hidden(['name']);
                },
                'sku.material' => function ($q) use ($prefix) {
                    $q->alias('sku')->field([
                        'sku.product_id',
                        'sku.product_price',
                        'sku.product_sales',
                        'sku.product_sku_id',
                        'sku.product_weight',
                        'sku.spec_name',
                        'sku.spec_sku_id',
                        'sku.stock_num'
                    ])->field([
                        "IF(
                            EXISTS(SELECT id FROM {$prefix}product_sold_out as sold_out WHERE sold_out.product_sku_id = sku.product_sku_id) OR
                            EXISTS(
                                SELECT id FROM {$prefix}product_sku_material as skm
                                LEFT JOIN {$prefix}product AS pp ON pp.product_id = skm.material_id
                                WHERE skm.product_sku_id = sku.product_sku_id
                                AND skm.material_id = pp.product_id
                                AND pp.product_status = 20
                            )
                            , 1, 0
                        ) as is_sold_out"
                    ]);
                },
                'feed',
            ])
            ->when($special_id, function ($q) use ($special_id, $prefix) {
                $q->where($prefix . 'product.special_id', '=', $special_id);
            })
            ->when($category_id && $special_id == 0, function ($q) use ($category_id) {
                $q->where(function ($query) use ($category_id) {
                    $query->where('c.category_id', '=', $category_id);
                    $query->whereOr('c.parent_id', '=', $category_id);
                });
            })
            ->when($search, function ($q) use ($search, $prefix) {
                $q->like($prefix . 'product.product_name', trim($search));
            })
            ->when($settingInfo['menu_show_sold_out'] ? 0 : 1, function ($q) {
                $q->where('product_sku.sku_stock_num', '>', 0);
            })
            //
            ->where('c.status', '=', 1)
            ->where($prefix . 'product.is_delete', '=', 0)
            ->where($prefix . 'product.type', '=', 10)    // 10-成品 20-材料
            ->where($prefix . 'product.product_type', '=', 1)
            ->where($prefix . 'product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where($prefix . 'product.product_status', '=', 10)
            ->order([$prefix . 'product.product_sort', $prefix . 'product.product_id' => 'desc'])
            ->select()
            ->toArray();

        // 处理列表库存数量
        foreach ($result as $key => $item) {
            // 处理加料
            foreach ($item['feed'] as $feed_k => $feed_v) {
                // 库存联动材料数
                if ($feed_v['material']) {
                    $stock = 0;
                    foreach ($feed_v['material'] as $material) {
                        $material_num = $material['material_num'] <= 0 ? 0.0001 : $material['material_num'];
                        $remaining_num = intval($material['materialProduct']['product_material_stock'] / $material_num); // 材料还能做几份
                        $stock = $stock == 0 ? $remaining_num : min($stock, $remaining_num);
                    }
                    $result[$key]['feed'][$feed_k]['stock_num'] = $stock;
                }
            }
            // 处理规格
            foreach ($item['sku'] as $sku_k => $sku_v) {
                if ($sku_v['is_sold_out'] == 1) {
                    $result[$key]['sku'][$sku_k]['stock_num'] = 0;
                }
                // 库存联动材料数
                if ($sku_v['is_sold_out'] != 1 && $sku_v['material']) {
                    $min_num = [];
                    foreach ($sku_v['material'] as $material) {
                        $material_num = $material['material_num'] <= 0 ? 0.0001 : $material['material_num'];
                        $remaining_num = intval($material['materialProduct']['product_material_stock'] / $material_num); // 材料还能做几份
                        $min_num[] = $remaining_num;
                    }
                    $result[$key]['sku'][$sku_k]['stock_num'] = empty($min_num) ? 0 : min($min_num);
                }
            }
            $result[$key]['product_stock'] = array_sum(array_column($result[$key]['sku'] ?? [], 'stock_num'));
        }
        //
        return $result;
    }
}
