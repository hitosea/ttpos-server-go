<?php

namespace app\scan\model\product;

use think\facade\Db;
use think\facade\Env;
use app\common\model\order\OrderBuffet;
use app\common\model\supplier\Supplier;
use app\common\model\buffet\BuffetProduct;
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
        $order_id = $params['order_id'] ?? 0;
        $product_source = $params['product_source'] ?? 1;     // 1-收银 2-桌台
        //
        $prefix = Env::get('DB_PREFIX');
        //
        $settingInfo = Supplier::getTabletBaseInfo();
        // 过滤商品
        $buffetIds = (new OrderBuffet)->where('order_id', $order_id)->column('buffet_id');
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
                $prefix . 'product.feed_max_select',
                $prefix . 'product.feed_open_max_select',
                $prefix . 'product.feed_required',
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
            ->when($settingInfo['tablet']['is_show_sold_out'] ? 0 : 1, function ($q) use ($settingInfo, $prefix) {
                $q->where('product_sku.sku_stock_num', '>', 0);
            })
            ->when($order_id, function ($q) use ($order_id, $product_source) {
                $q->withSum([
                    'orderProducts' => function ($q) use ($order_id, $product_source) {
                        $q->where('order_id', $order_id)->where(function ($q) use ($product_source) {
                            $q->where('is_send_kitchen', 0); // 扫码端统计未送厨的
                            $q->where('add_source', $product_source);
                            $q->whereOr('is_send_kitchen', 1);
                        });
                    }
                ], 'total_num');
            })
            //不显示在h5端的
            ->when(count($buffetIds) > 0, function ($q) use ($prefix, $buffetIds) {
                $buffetIds = implode(",", $buffetIds);
                $q->leftJoin("
                    (
                        select * from (
                            select bp.product_id, is_show_h5
                            from {$prefix}buffet_product as bp
                            where bp.buffet_id in ($buffetIds)
                            order by bp.is_show_h5
                            limit 99999
                        ) bp
                        group by bp.product_id
                    ) bp
                ", $prefix . 'product.product_id = bp.product_id');
                $q->where("ifnull(bp.is_show_h5,{$prefix}product.is_show_h5) != 2");
            })
            ->when(count($buffetIds) == 0, function ($q) use ($prefix) {
                $q->where("{$prefix}product.is_show_h5 != 2");
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

    /**
     * 获取自助餐商品变更列表
     */
    public function getBuffetChangeList($params)
    {
        Db::connect()->execute("SET SESSION sql_mode = ''");
        $isSpecial = $params['is_special'] ?? 0;
        $categoryId = $params['category_id'] ?? 0;
        $specialId = $isSpecial ? $categoryId : 0;
        $orderId = $params['order_id'] ?? 0;
        $mealNum = $params['meal_num'] ?? 1;
        //
        $prefix = Env::get('DB_PREFIX');
        //
        $settingInfo = Supplier::getTabletBaseInfo();
        $isShowSoldOut = $settingInfo['tablet']['is_show_sold_out'] ? 0 : 1;
        // 过滤商品
        $buffetIds = (new OrderBuffet)->where('order_id', $orderId)->column('buffet_id');
        $productIds = $buffetIds ? (new BuffetProduct)->where('buffet_id', 'in', $buffetIds)->column('product_id') : [];
        if (count($buffetIds) > 0) {
            $buffetIds = implode(",", $buffetIds);
        } else {
            $buffetIds = 0;
        }
        // 筛选条件
        $list = $this->alias($prefix . 'product')
            ->leftJoin('category c', $prefix . 'product.category_id = c.category_id')
            ->leftJoin("
                (
                    select product_sku.product_id, CAST(ifnull(sum(product_sku.stock_num),0) AS SIGNED) as sku_stock_num
                    from {$prefix}product_sku as product_sku
                    where not EXISTS(select 1 from {$prefix}product_sold_out as sold_out where sold_out.product_sku_id = product_sku.product_sku_id)
                    group by product_sku.product_id
                ) product_sku
            ", $prefix . 'product.product_id = product_sku.product_id')
            ->leftJoin("
                    (
                        select * from (
                            select bp.product_id, bp.limit_num, bp.is_show_h5
                            from {$prefix}buffet_product as bp
                            where bp.buffet_id in ($buffetIds)
                            order by bp.is_show_h5
                            limit 99999
                        ) bp
                        group by bp.product_id
                    ) bp
                ", $prefix . 'product.product_id = bp.product_id')
            ->field([
                $prefix . 'product.product_id',
                '0 as product_price', // 自助餐商品，价格固定为0
                'product_sku.sku_stock_num as product_stock',
                "ifnull(bp.limit_num * {$mealNum},{$prefix}product.limit_num * {$mealNum}) as limit_num", // 自助餐限购数量
                "ifnull(bp.is_show_h5,{$prefix}product.is_show_h5) as is_show_h5",
            ])
            // 是否显示售罄商品 0-关闭 1-开启
            ->when($isShowSoldOut, function ($q) {
                $q->where('product_sku.sku_stock_num', '>', 0);
            })
            // 分类筛选
            ->when($specialId, function ($q) use ($specialId, $prefix) {
                $q->where($prefix . 'product.special_id', $specialId);
            })
            ->when($categoryId && $isSpecial == 0, function ($q) use ($categoryId) {
                $q->where(function ($query) use ($categoryId) {
                    $query->where('c.category_id', '=', $categoryId);
                    $query->whereOr('c.parent_id', '=', $categoryId);
                });
            })
            // 自助餐关联商品
            ->when(count($productIds) > 0, function ($q) use ($productIds, $prefix) {
                $q->where($prefix . 'product.product_id', 'in', $productIds);
            })
            //
            ->where('c.status', '=', 1)
            ->where($prefix . 'product.is_delete', '=', 0)
            ->where($prefix . 'product.type', '=', 10)    // 10-成品 20-材料
            ->where($prefix . 'product.product_type', '=', 1)
            ->where($prefix . 'product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where($prefix . 'product.product_status', '=', 10)
            ->order([$prefix . 'product.product_sort', $prefix . 'product.product_id' => 'desc'])
            ->paginate($params)
            ->append([])
            ->toArray();
        //
        return $list;
    }

    /**
     * 获取商品详情
     */
    public static function detail($product_id)
    {
        $model = (new static())->with([
            'category',
            'image.file',
            'sku' => function ($q) {
                $q->alias('sku')->field([
                    'sku.product_id',
                    'sku.product_price',
                    'sku.product_sales',
                    'sku.product_sku_id',
                    'sku.product_weight',
                    'sku.spec_name',
                    'sku.spec_sku_id'
                ])->field([
                    'EXISTS(SELECT id FROM jjjfood_product_sold_out as sold_out WHERE sold_out.product_sku_id = sku.product_sku_id) as is_sold_out',
                    'CAST(if(EXISTS(SELECT 1 FROM jjjfood_product_sold_out WHERE product_sku_id = sku.product_sku_id), 0, sku.stock_num) AS UNSIGNED) as stock_num'
                ])
                    ->hidden(['spec_name']);
            },
            'supplier'
        ])->where('product_id', '=', $product_id)->find();
        if (empty($model)) {
            return $model;
        }
        //
        $model['product_stock'] = array_sum(array_column($model['sku']?->toArray() ?? [], 'stock_num'));
        //
        // 整理商品数据并返回
        return $model->setProductListData($model, false);
    }
}
