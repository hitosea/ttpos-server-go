<?php

namespace app\common\model\product;

use think\facade\Db;

use think\facade\Env;
use app\common\model\app\App;
use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\erp\ErpSupplier;
use app\common\model\file\UploadFile;
use app\common\model\product\Material;
use app\common\model\order\OrderBuffet;
use app\common\model\order\OrderProduct;
use app\common\model\product\ProductFeed;
use app\common\model\buffet\BuffetProduct;
use app\common\service\websocket\Websocket;
use app\common\model\erp\ErpInventoryRecord;
use app\common\model\store\MultiLanguageName;
use app\common\model\order\OrderSchemeProduct;
use app\common\model\supplier\PrintingProduct;
use app\common\model\product\ProductSkuMaterial;
use app\common\model\product\ProductFeedMaterial;

/**
 * 商品模型
 */
class Product extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_package';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

    /**
     * 兼容字段
     */
    protected $append = [
        'type',
        'product_id',
        'product_name',
        'product_price',
        'category_id',
        'special_id',
        'erp_supplier_id',
        'img_name',
        'unit_id',
        'label_id',
        'feed_required',
        'feed_open_max_select',
        'feed_max_select',
        'selling_point',
        'selling_point_i18n',
        'is_enable_grade',
        'is_alone_grade',
        'product_sort',
        'product_status',
        'product_sales',
        'product_name_text',
        'product_unit_text',
    ];

    private const SELLING_POINT_LANGUAGES = ['zh', 'en', 'zhtw', 'th', 'my', 'ja', 'ko', 'tr', 'sv'];

    /*
     * 类型 10-成品 20-材料
     */
    const TYPE_PRODUCT = 10;
    const TYPE_MATERIAL = 20;
    const TYPE_PACKAGE = 30;

    /**
     * 商品更新后推送通知
     */
    public static function onAfterWrite(Product $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 商品删除后推送通知
     */
    public static function onAfterDelete(Product $model)
    {
        
        $msgData = [
            'type' => 'delete',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 兼容字段
     */
    public function getTypeAttr($value, $data = [])
    {
        return $this->product_type == 0 ? self::TYPE_PRODUCT : self::TYPE_PACKAGE;
    }
    public function getProductIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }
    public function getProductNameAttr($value, $data = [])
    {
        return $this->getData('name') ?: '';
    }
    public function getCategoryIdAttr($value, $data = [])
    {
        return $this->category_uuid ?: 0;
    }
    public function getSpecialIdAttr($value, $data = [])
    {
        return $this->special_category_uuid ?: 0;
    }
    public function getErpSupplierIdAttr($value, $data = [])
    {
        return $this->supplier_uuid ?: 0;
    }
    public function getImgNameAttr($value, $data = [])
    {
        return $this->image_name ?: '';
    }
    public function getUnitIdAttr($value, $data = [])
    {
        return $this->unit_uuid ?: 0;
    }
    public function getLabelIdAttr($value, $data = [])
    {
        return $this->printer_tag_uuid ?: 0;
    }
    public function getDeductStockTypeAttr($value, $data = [])
    {
        return $value ? 10 : 20;
    }
    public function getFeedRequiredAttr($value, $data = [])
    {
        return $this->sauce_required ?: 0;
    }
    public function getFeedOpenMaxSelectAttr($value, $data = [])
    {
        return $this->sauce_max_selection ? 1 : 0;
    }
    public function getFeedMaxSelectAttr($value, $data = [])
    {
        return $this->sauce_max_selection ?: 0;
    }
    public function getSellingPointAttr($value, $data = [])
    {
        $describe = $this->getData('describe');
        return is_string($describe) ? $describe : '';
    }

    public function getSellingPointI18nAttr($value, $data = [])
    {
        $default = [
            'zh' => '',
            'en' => '',
            'zhtw' => '',
            'th' => '',
            'my' => '',
            'ja' => '',
            'ko' => '',
            'tr' => '',
            'sv' => '',
        ];
        $uuid = $data['describe_multi_language_name_uuid'] ?? 0;
        if (empty($uuid)) {
            return json_encode($default);
        }
        $names = (new MultiLanguageName)->getNames($uuid);
        if (empty($names)) {
            return json_encode($default);
        }
        $decoded = json_decode($names, true);
        return is_array($decoded) ? json_encode($decoded) : json_encode($default);
    }
    public function getIsEnableGradeAttr($value, $data = [])
    {
        return $this->open_discount ?: 0;
    }
    public function getProductStatusAttr($value, $data)
    {
        $value = $this->status ? 10 : 20;
        $status = [10 => __('上架'), 20 => __('下架')];
        return ['text' => $status[$value], 'value' => $value];
    }
    public function getProductSortAttr($value, $data)
    {
        return $this->sort ?: 0;
    }
    public function getIsAloneGradeAttr($value, $data)
    {
        return 0;
    }

    /**
     * 商品价格
     */
    public function getProductPriceAttr($value)
    {
        return floatval($value);
    }

    /**
     * 材料库存
     */
    public function getProductMaterialStockAttr($value)
    {
        return floatval($value);
    }

    /**
     * 商品名称
     */
    public static function getProductNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name'] ?? '');
    }

    /**
     * 获取单位
     */
    public function getProductUnitTextAttr($value, $data)
    {
        return extractLanguage($value ?: $data['product_unit'] ?? '');
    }

    /**
     * 获取单位规格名称
     */
    public function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['spec_name'] ?? '');
    }

    /**
     * 属性
     */
    public function getProductAttrAttr($value)
    {
        $datas = $value ? (json_decode($value, true) ?: []) : [];
        foreach ($datas as $key => $data) {
            if (isset($datas[$key]['attribute_name'])) {
                $datas[$key]['attribute_name_text'] = extractLanguage($datas[$key]['attribute_name']);
                if (isset($datas[$key]['attribute_value'])) {
                    foreach ($datas[$key]['attribute_value'] as $k => $v) {
                        $datas[$key]['attribute_value_text'][$k] = extractLanguage($v);
                    }
                }
            }
        }
        return $datas;
    }

    /**
     * 加料
     */
    public function getProductFeedAttr($value)
    {
        $datas = $value ? (json_decode($value, true) ?: []) : [];
        foreach ($datas as $key => $data) {
            $datas[$key]['feed_name_text'] = extractLanguage($datas[$key]['feed_name']);
        }
        return $datas;
    }

    /**
     * 实际销量
     */
    public function getSalesActualAttr($value)
    {
        return floatval($value);
    }

    /**
     * 计算显示销量 (初始销量 + 实际销量)
     */
    public function getProductSalesAttr($value, $data)
    {
        $salesInitial = isset($data['sales_initial']) && is_numeric($data['sales_initial']) ? $data['sales_initial'] : 0;
        $salesActual = isset($data['sales_actual']) && is_numeric($data['sales_actual']) ? $data['sales_actual'] : 0;
        return floatval($salesInitial + $salesActual);
    }

    /**
     * 属性配置
     */
    public function setProductAttrAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 属性加料
     */
    public function setProductFeedAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 获取器：单独设置折扣的配置
     */
    public function getAloneGradeEquityAttr($value)
    {
        return $value ? json_decode($value, true) : '';
    }

    /**
     * 修改器：单独设置折扣的配置
     */
    public function setAloneGradeEquityAttr($value)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 关联商品分类表
     */
    public function category()
    {
        return $this->belongsTo(Category::class, 'category_uuid', 'uuid');
    }

    /**
     * 关联打印标签表
     */
    public function label()
    {
        return $this->hasOne('app\\common\\model\\product\\Label', 'label_id', 'label_id');
    }

    /**
     * 关联商品规格表
     */
    public function sku()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductBom', 'product_package_uuid', 'uuid')->where('product_sauce_uuid', '=', 0)->order(['uuid' => 'asc']);
    }

    /**
     * 关联加料表
     */
    public function feed()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductBom', 'product_package_uuid', 'uuid')->where('product_sauce_uuid', '>', 0)->order(['uuid' => 'asc']);
    }

    /**
     * 关联商品图片表
     */
    public function image()
    {
        return $this->hasMany('app\\common\\model\\file\\UploadFile', 'uuid', 'image_file_uuid')->order(['id' => 'asc']);
    }

    /**
     * 关联商品图片表
     */
    public function logo()
    {
        return $this->hasOne('app\\common\\model\\product\\ProductImage')->order(['id' => 'asc']);
    }

    /**
     * 关联供应商表
     */
    public function supplier()
    {
        return $this->belongsTo('app\\common\\model\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id')->field(['shop_supplier_id', 'name', 'address', 'logo']);
    }

    /**
     * 关联订单商品
     */
    public function orderProducts()
    {
        return $this->hasMany(OrderProduct::class, 'product_id', 'product_id');
    }

    /**
     * 关联erp供应商
     */
    public function erpSupplier()
    {
        return $this->belongsTo(ErpSupplier::class, 'supplier_uuid', 'uuid');
    }

    /**
     * 规格关联材料
     */
    public function skuProduct()
    {
        return $this->belongsToMany(Product::class, 'product_sku_material', 'product_id', 'material_id')->field(['product_id', 'product_name']);
    }

    /**
     * 加料关联材料
     */
    public function feedProduct()
    {
        return $this->belongsToMany(Product::class, 'product_feed_material', 'product_id', 'material_id')->field(['product_id', 'product_name']);
    }

    /**
     * 关联堂食税类
     */
    public function dineTax()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductTax', 'dine_tax_uuid', 'uuid');
    }

    /**
     * 关联外卖税类
     */
    public function takeoutTax()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductTax', 'takeout_tax_uuid', 'uuid');
    }

    /**
     * 关联产品套餐属性组
     */
    public function productAttributeGroup()
    {
        return $this->hasMany(ProductAttributeGroup::class, 'product_package_uuid', 'uuid');
    }

    /**
     * 关联产品语言
     */
    public function MultiLanguageName()
    {
        return $this->belongsTo(MultiLanguageName::class, 'multi_language_name_uuid', 'uuid');
    }

    /**
     * 关联自助餐商品
     */
    public function buffetProduct()
    {
        return $this->hasMany(BuffetProduct::class, 'product_package_uuid', 'uuid');
    }

    /**
     * 关联必点方案商品
     */
    public function orderSchemeProduct()
    {
        return $this->hasMany(OrderSchemeProduct::class, 'product_package_uuid', 'uuid');
    }

    /**
     * 关联套餐分组
     */
    public function productPackageGroup()
    {
        return $this->hasMany(ProductPackageGroup::class, 'product_package_uuid', 'uuid');
    }

    /**
     * 关联商品打印机表
     */
    public function productPrinters()
    {
        return $this->hasMany(PrintingProduct::class, 'product_package_uuid', 'uuid');
    }


    /**
     * 获取商品缓存基础列表 attribute_name
     * @param $params
     * @param $product_ids  // 查询指定商品
     * @return array
     */
    public function getBaseList($params, $is_page = true, $product_ids = [])
    {
        Db::connect()->execute("SET SESSION sql_mode = ''");
        $isSpecial = $params['is_special'] ?? 0;
        $categoryId = $params['category_id'] ?? 0;
        $specialId = $isSpecial ? $categoryId : 0;
        $orderId = $params['order_id'] ?? 0;
        $productSource = $params['product_source'] ?? 1;     // 1-收银 2-桌台
        $prefix = Env::get('DB_PREFIX');
        //
        $model = $this->alias($prefix . 'product')
            ->leftJoin('category c', $prefix . 'product.category_id = c.category_id')
            ->field([
                $prefix . 'product.category_id',
                $prefix . 'product.limit_num',
                $prefix . 'product.product_id',
                $prefix . 'product.product_name',
                $prefix . 'product.product_price',
                $prefix . 'product.product_stock',
                $prefix . 'product.spec_type',
                $prefix . 'product.special_id',
                $prefix . 'product.sales_actual',
                $prefix . 'product.sales_initial',
                $prefix . 'product.product_unit',
                $prefix . 'product.product_sort',
                $prefix . 'product.product_attr',
                $prefix . 'product.is_show_cashier',
                $prefix . 'product.is_show_tablet',
                $prefix . 'product.is_show_assistant',
                $prefix . 'product.is_show_kitchen',
                $prefix . 'product.is_show_h5',
                $prefix . 'product.feed_required',
                $prefix . 'product.feed_open_max_select',
                $prefix . 'product.feed_max_select',
                'c.parent_id as category_parent_id',
            ])->with([
                'image.file',
                'sku.material' => function ($q) use ($prefix) {
                    $q->alias('sku')->field([
                        'sku.product_id',
                        'sku.product_price',
                        'sku.product_sales',
                        'sku.product_sku_id',
                        'sku.product_weight',
                        'sku.spec_name',
                        'sku.spec_sku_id',
                        'sku.stock_num',
                        'sku.barcode',
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
                        ) as is_sold_out",
                    ]);
                },
                'feed',
            ])
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
            // 根据订单来源送厨关联显示
            ->when($orderId, function ($q) use ($orderId, $productSource) {
                $q->withSum(['orderProducts' => function ($q) use ($orderId, $productSource) {
                    $q->where('order_id', $orderId)->where(function ($q) use ($productSource) {
                        $q->where('is_send_kitchen', 1);
                        $q->whereOr('add_source', $productSource);
                    });
                }], 'total_num');
            })
            // 查找部分商品
            ->when(!empty($product_ids), function ($q) use ($prefix, $product_ids) {
                $q->whereIn($prefix . 'product.product_id', $product_ids);
            })
            //
            ->where('c.status', '=', 1)
            ->where($prefix . 'product.delete_time', '=', 0)
            ->where($prefix . 'product.type', '=', 10)    // 10-成品 20-材料
            ->where($prefix . 'product.product_type', '=', 1)
            ->where($prefix . 'product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where($prefix . 'product.product_status', '=', 10)
            ->order([$prefix . 'product.product_sort', $prefix . 'product.product_id' => 'desc']);

        if ($is_page) {
            $result = $model->paginate($params)->toArray();
        } else {
            $result = $model->select()->toArray();
        }

        if (!$is_page) {
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
        } else {
            // 处理列表库存数量
            foreach ($result['data'] as $key => $item) {
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
                        $result['data'][$key]['feed'][$feed_k]['stock_num'] = $stock;
                    }
                }
                // 处理规格
                foreach ($item['sku'] as $sku_k => $sku_v) {
                    if ($sku_v['is_sold_out'] == 1) {
                        $result['data'][$key]['sku'][$sku_k]['stock_num'] = 0;
                    }
                    // 库存联动材料数
                    if ($sku_v['is_sold_out'] != 1 && $sku_v['material']) {
                        $min_num = [];
                        foreach ($sku_v['material'] as $material) {
                            $material_num = $material['material_num'] <= 0 ? 0.0001 : $material['material_num'];
                            $remaining_num = intval($material['materialProduct']['product_material_stock'] / $material_num); // 材料还能做几份
                            $min_num[] = $remaining_num;
                        }
                        $result['data'][$key]['sku'][$sku_k]['stock_num'] = empty($min_num) ? 0 : min($min_num);
                    }
                }
                $result['data'][$key]['product_stock'] = array_sum(array_column($result['data'][$key]['sku'] ?? [], 'stock_num'));
            }
        }
        //
        return $result;
    }

    /**
     * 获取商品缓存基础列表 attribute_name
     * @param $params
     * @param $product_ids  // 查询指定商品
     * @return array
     */
    public function getBaseListOptimize($params, $is_page = true, $product_ids = [])
    {
        Db::connect()->execute("SET SESSION sql_mode = ''");
        $isSpecial = $params['is_special'] ?? 0;
        $categoryId = $params['category_id'] ?? 0;
        $specialId = $isSpecial ? $categoryId : 0;
        $orderId = $params['order_id'] ?? 0;
        $productSource = $params['product_source'] ?? 1;     // 1-收银 2-桌台
        $prefix = Env::get('DB_PREFIX');
        //
        $model = $this->alias($prefix . 'product')
            ->leftJoin('category c', $prefix . 'product.category_id = c.category_id')
            ->field([
                $prefix . 'product.category_id',
                $prefix . 'product.limit_num',
                $prefix . 'product.product_id',
                $prefix . 'product.product_name',
                $prefix . 'product.product_price',
                $prefix . 'product.product_stock',
                $prefix . 'product.spec_type',
                $prefix . 'product.special_id',
                $prefix . 'product.sales_actual',
                $prefix . 'product.sales_initial',
                $prefix . 'product.product_unit',
                $prefix . 'product.product_sort',
                $prefix . 'product.product_attr',
                $prefix . 'product.is_show_cashier',
                $prefix . 'product.is_show_tablet',
                $prefix . 'product.is_show_assistant',
                $prefix . 'product.is_show_kitchen',
                $prefix . 'product.is_show_h5',
                $prefix . 'product.feed_required',
                $prefix . 'product.feed_open_max_select',
                $prefix . 'product.feed_max_select',
                'c.parent_id as category_parent_id',
            ])->with([
                'image.file',
                'sku.material' => function ($q) use ($prefix) {
                    $q->alias('sku')->field([
                        'sku.product_id',
                        'sku.product_price',
                        'sku.product_sales',
                        'sku.product_sku_id',
                        'sku.product_weight',
                        'sku.spec_name',
                        'sku.spec_sku_id',
                        'sku.stock_num',
                        'sku.barcode',
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
                        ) as is_sold_out",
                    ]);
                },
                'feed',
            ])
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
            // 根据订单来源送厨关联显示
            ->when($orderId, function ($q) use ($orderId, $productSource) {
                $q->withSum(['orderProducts' => function ($q) use ($orderId, $productSource) {
                    $q->where('order_id', $orderId)->where(function ($q) use ($productSource) {
                        $q->where('is_send_kitchen', 1);
                        $q->whereOr('add_source', $productSource);
                    });
                }], 'total_num');
            })
            // 查找部分商品
            ->when(!empty($product_ids), function ($q) use ($prefix, $product_ids) {
                $q->whereIn($prefix . 'product.product_id', $product_ids);
            })
            //
            ->where('c.status', '=', 1)
            ->where($prefix . 'product.delete_time', '=', 0)
            ->where($prefix . 'product.type', '=', 10)    // 10-成品 20-材料
            ->where($prefix . 'product.product_type', '=', 1)
            ->where($prefix . 'product.shop_supplier_id', '=', $params['shop_supplier_id'])
            ->where($prefix . 'product.product_status', '=', 10)
            ->order([$prefix . 'product.product_sort', $prefix . 'product.product_id' => 'desc']);

        if ($is_page) {
            $result = $model->paginate($params)->toArray();
        } else {
            $result = $model->select()->toArray();
        }

        if (!$is_page) {
            // 处理列表库存数量
            foreach ($result as $key => $item) {
                $result[$key]['product_name'] = $this->parseJsonValue($item['product_name']);
                $result[$key]['product_unit'] = $this->parseJsonValue($item['product_unit']);
                // 处理属性
                if (isset($item['product_attr'])) {
                    $productAttr = &$result[$key]['product_attr'];
                    foreach ($productAttr as &$attr) {
                        $fields = ['attribute_name', 'parent_name'];
                        foreach ($fields as $field) {
                            if (isset($attr[$field])) {
                                $attr[$field] = $this->parseJsonValue($attr[$field]);
                            }
                        }
                        if (isset($attr['attribute_value'])) {
                            $attr['attribute_value'] = array_map([$this, 'parseJsonValue'], $attr['attribute_value']);
                        }
                    }
                    unset($attr);
                }
                // 处理加料
                foreach ($item['feed'] as $feed_k => $feed_v) {
                    $result[$key]['feed'][$feed_k]['feed_name'] = $this->parseJsonValue($feed_v['feed_name']);
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
                    $result[$key]['sku'][$sku_k]['spec_name'] = $this->parseJsonValue($sku_v['spec_name']);
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
        } else {
            // 处理列表库存数量
            foreach ($result['data'] as $key => $item) {
                $result['data'][$key]['product_name'] = $this->parseJsonValue($item['product_name']);
                $result['data'][$key]['product_unit'] = $this->parseJsonValue($item['product_unit']);
                // 处理属性
                if (isset($item['product_attr'])) {
                    $productAttr = &$result['data'][$key]['product_attr'];
                    foreach ($productAttr as &$attr) {
                        $fields = ['attribute_name', 'parent_name'];
                        foreach ($fields as $field) {
                            if (isset($attr[$field])) {
                                $attr[$field] = $this->parseJsonValue($attr[$field]);
                            }
                        }
                        if (isset($attr['attribute_value'])) {
                            $attr['attribute_value'] = array_map([$this, 'parseJsonValue'], $attr['attribute_value']);
                        }
                    }
                    unset($attr);
                }
                // 处理加料
                foreach ($item['feed'] as $feed_k => $feed_v) {
                    $result['data'][$key]['feed'][$feed_k]['feed_name'] = $this->parseJsonValue($feed_v['feed_name']);
                    // 库存联动材料数
                    if ($feed_v['material']) {
                        $stock = 0;
                        foreach ($feed_v['material'] as $material) {
                            $material_num = $material['material_num'] <= 0 ? 0.0001 : $material['material_num'];
                            $remaining_num = intval($material['materialProduct']['product_material_stock'] / $material_num); // 材料还能做几份
                            $stock = $stock == 0 ? $remaining_num : min($stock, $remaining_num);
                        }
                        $result['data'][$key]['feed'][$feed_k]['stock_num'] = $stock;
                    }
                }
                // 处理规格
                foreach ($item['sku'] as $sku_k => $sku_v) {
                    $result['data'][$key]['sku'][$sku_k]['spec_name'] = $this->parseJsonValue($sku_v['spec_name']);
                    if ($sku_v['is_sold_out'] == 1) {
                        $result['data'][$key]['sku'][$sku_k]['stock_num'] = 0;
                    }
                    // 库存联动材料数
                    if ($sku_v['is_sold_out'] != 1 && $sku_v['material']) {
                        $min_num = [];
                        foreach ($sku_v['material'] as $material) {
                            $material_num = $material['material_num'] <= 0 ? 0.0001 : $material['material_num'];
                            $remaining_num = intval($material['materialProduct']['product_material_stock'] / $material_num); // 材料还能做几份
                            $min_num[] = $remaining_num;
                        }
                        $result['data'][$key]['sku'][$sku_k]['stock_num'] = empty($min_num) ? 0 : min($min_num);
                    }
                }
                $result['data'][$key]['product_stock'] = array_sum(array_column($result['data'][$key]['sku'] ?? [], 'stock_num'));
            }
        }
        //
        return $result;
    }

    /**
     * 将JSON字符串转换为数组格式
     * @param string $value 待转换的JSON字符串
     * @return array|string 转换后的数组或原始字符串
     */
    public function parseJsonValue($value)
    {
        if (empty($value)) {
            return [];
        }
        $decoded = json_decode($value, true);
        if (json_last_error() === JSON_ERROR_NONE && is_array($decoded)) {
            return $decoded;
        }
        return $value;
    }

    /**
     * 获取商品列表
     */
    public function getList($param, $page = 0)
    {
        // 商品列表获取条件
        $params = array_merge([
            'type' => 'sell',         // 商品状态
            'category_id' => 0,     // 分类id
            'sortType' => 'all',    // 排序类型
            'list_rows' => 15,       // 每页数量
            'special_id' => 0,        //特殊分类id
        ], $param);

        // 成品
        $productSql = self::alias('p')
            ->field(implode(',', [
                'file.file_type',
                'file.file_url',
                'file.file_name',
                'file.storage',
                'file.save_name',
                'file.url_param',
                'p.create_time',
                'p.status',
                'p.uuid as product_id', 
                'p.name as product_name', 
                'p.actual_sale_num as sales_actual',
                'p.sort as product_sort',
                'c1.multi_language_name_uuid as category_multi_language_name_uuid', 
                'c2.multi_language_name_uuid as category_parent_multi_language_name_uuid', 
                '(SELECT sum(stock_num) FROM ttpos_product_bom WHERE product_package_uuid = p.uuid AND delete_time = 0 AND product_sauce_uuid = 0) AS product_stock',
                '(SELECT min(price) FROM ttpos_product_bom WHERE product_package_uuid = p.uuid AND product_sauce_uuid = 0 AND delete_time = 0) AS product_price',
                'CASE WHEN product_type = 1 THEN 30 ELSE 10 END as type',
                'c1.uuid as category_uuid',
                'c2.uuid as category_parent_uuid',
                '0 as is_material_used',
                'p.sort as sort',
                'p.id as id',
                'pu.multi_language_name_uuid as product_unit_multi_language_name_uuid',
                'bom.is_open_stock',
                '(SELECT count(uuid) FROM ttpos_product_package_group_item WHERE related_uuid = p.uuid AND delete_time = 0) as is_package_used',
            ]))
            ->leftJoin('product_category c1', 'p.category_uuid = c1.uuid')
            ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
            ->leftJoin('product_bom bom', 'p.uuid = bom.product_package_uuid')
            ->leftJoin('file', 'p.image_file_uuid = file.uuid')
            ->leftJoin('product_unit pu', 'p.unit_uuid = pu.uuid')
            ->leftJoin('product_package_group_item pgItem', 'pgItem.related_uuid = p.uuid')
            ->where('bom.product_sauce_uuid', '=', 0)
            ->group('p.uuid')   
            ->order('p.sort', 'asc')
            ->order('p.id', 'desc')
            ->buildSql();

        // 材料
        $enableErp = App::detail(self::$app_id)->isEnableErp();
        $materialSqlBuilder = Material::alias('m')
            ->field(implode(',', [
                'file.file_type',
                'file.file_url',
                'file.file_name',
                'file.storage',
                'file.save_name',
                'file.url_param',
                'm.create_time',
                'm.status',
                'm.uuid as product_id', 
                'm.name as product_name', 
                'm.actual_sale_num as sales_actual',
                '0 as product_sort',
                'c1.multi_language_name_uuid as category_multi_language_name_uuid', 
                !$enableErp ? 'c2.multi_language_name_uuid as category_parent_multi_language_name_uuid' : '0 as category_parent_multi_language_name_uuid', 
                'wi.stock as product_stock',
                '0 as product_price',
                '20 as type', 
                'c1.uuid as category_uuid',
                !$enableErp ? 'c2.uuid as category_parent_uuid' : '0 as category_parent_uuid',
                'count(rm.uuid) as is_material_used',
                '0 as sort',
                'm.id as id',
                'pu.multi_language_name_uuid as product_unit_multi_language_name_uuid',
                '1 as is_open_stock',
                '0 as is_package_used',
            ]));

            if (!$enableErp) {
                $materialSqlBuilder
                    ->leftJoin('product_category c1', 'm.category_uuid = c1.uuid')
                    ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
                    ->leftJoin('product_unit pu', 'm.unit_uuid = pu.uuid');
            } else {
                $materialSqlBuilder
                    ->leftJoin('material_category c1', 'm.category_uuid = c1.uuid')
                    ->leftJoin('material_unit mu', 'm.unit_uuid = mu.uuid')
                    ->leftJoin('product_unit pu', 'mu.unit_uuid = pu.uuid');
            }
            
            $materialSqlBuilder
                ->leftJoin('file', 'm.image_uuid = file.uuid')
                ->leftJoin('related_material rm', 'm.uuid = rm.material_uuid AND rm.delete_time = 0')
                ->leftJoin('warehouse_item wi', 'm.uuid = wi.material_uuid AND wi.warehouse_uuid = (SELECT uuid FROM ttpos_warehouse WHERE type = "normal" AND is_default = 1 LIMIT 1)')
                ->group('m.uuid')
                ->order('m.id', 'desc');

        $materialSql = $materialSqlBuilder->buildSql();

        // 分页
        $offset = ($params['page'] - 1) * $params['list_rows'];
        $limit = $params['list_rows'];

        // 搜索
        $bind = [];
        $whereSql = '';

        // 搜索商品类型
        $materialType = $params['material_type'] ?? 0;
        if (in_array($materialType, [self::TYPE_MATERIAL, self::TYPE_PRODUCT, self::TYPE_PACKAGE])) {
            $whereSql .= " WHERE (type = :type)";
            $bind['type'] = $materialType;
        }

        // 搜索商品分类
        $categoryId = $params['category_id'] ?? 0;
        if ($categoryId > 0) {
            $where = "(category_uuid = :category_uuid OR category_parent_uuid = :category_parent_uuid)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['category_uuid'] = $categoryId;
            $bind['category_parent_uuid'] = $categoryId;
        }

        // 搜索商品库存
        $stock = $params['stock'] ?? 0;
        if ($stock > 0) {
            $where = "(product_stock < :stock)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['stock'] = $stock;
        }

        // 搜索商品状态
        $status = $params['type'] ?? '';
        if (in_array($status, ['sell', 'lower'])) {
            $where = "(status = :status)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['status'] = [
                'sell' => 1,
                'lower' => 0,
            ][$status];
        }

        // 搜索商品名称
        $lang = request()->header('language') ?: 'zh';
        $productName = $params['product_name'] ?? '';
        if ($productName != '') {
            $where = "(JSON_UNQUOTE(JSON_EXTRACT(product_name, '$.$lang')) LIKE :product_name)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['product_name'] = "%{$productName}%";
        }
        
        // 搜索商品uuid列表
        $productIds = $params['product_ids'] ?? '';
        if ($productIds != '') {
            $where = "(FIND_IN_SET(product_id, :product_ids))";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['product_ids'] = $productIds;
        }

        $querySql = "SELECT " . implode(',', [
            'create_time',
            'category_multi_language_name_uuid', 
            'category_parent_multi_language_name_uuid', 
            'file_type',
            'file_url',
            'file_name',
            'product_id', 
            'product_name', 
            'product_sort',
            'product_price',
            'product_stock',
            'sales_actual',
            'status',
            'storage',
            'save_name',
            'type', 
            'url_param',
            'category_uuid',
            'category_parent_uuid',
            'is_material_used',
            'sort',
            'id',
            'product_unit_multi_language_name_uuid',
            'is_open_stock',
            'is_package_used',
        ]) . " FROM ($productSql UNION ALL $materialSql) AS all_product";
        $orderSql = ' ORDER BY sort ASC, create_time DESC';
        $pageSql = " LIMIT {$offset}, {$limit}";

        // 执行查询
        $total = 0;
        if ($page == 1) {
            $rows = Db::connect('shop' . self::$app_id)->query($querySql . $whereSql . $orderSql, $bind);
        } else {
            $countSql = "SELECT COUNT(*) AS total_count". " FROM ($productSql UNION ALL $materialSql) AS all_product";
            $count = Db::connect('shop' . self::$app_id)->query($countSql . $whereSql, $bind);
            $total = $count[0]['total_count'];
            $rows = Db::connect('shop' . self::$app_id)->query($querySql . $whereSql . $orderSql . $pageSql, $bind);
        }

        $file = new UploadFile();
        $list = [];
        
        foreach ($rows as $row) {
            // 分类
            $categoryNames = (new MultiLanguageName())->getNames($row['category_multi_language_name_uuid'] ?? 0);
            $pathNameText = extractLanguage($categoryNames);
            if ($row['category_parent_multi_language_name_uuid']) {
                $cayegoryParentNames = (new MultiLanguageName())->getNames($row['category_parent_multi_language_name_uuid']);
                $pathNameText = extractLanguage($cayegoryParentNames) . '-' . $pathNameText;
            }
            // 库存
            $productStock = 0;
            $productMaterialStock = 0;
            if ($row['type'] == 10 || $row['type'] == 30) {
                $productStock = $row['product_stock'];
            } else {
                $productMaterialStock = $row['product_stock'];
            }
            // 图片
            $filePath = $file->getFilePathAttr(null, [
                'file_type' => $row['file_type'],
                'file_url' => $row['file_url'],
                'file_name' => $row['file_name'],
                'storage' => $row['storage'],
                'save_name' => $row['save_name'],
                'url_param' => $row['url_param'],
            ]);
            // 规格
            $sku = [];
            if ($row['type'] == self::TYPE_MATERIAL) {
                $sku[] = [
                    'material_stock' => floatval($productMaterialStock),
                ];
            }

            // 单位
            $productUnit = '';
            $productUnitText = '';
            if ($row['product_unit_multi_language_name_uuid']) {
                $productUnit = (new MultiLanguageName())->getNames($row['product_unit_multi_language_name_uuid']);
                $productUnitText = extractLanguage($productUnit);
            }
            
            $list[] = [
                'category' => [ 'path_name_text' => $pathNameText ],
                'create_time' => date('Y-m-d H:i:s', $row['create_time']),
                'image' => [
                    [
                        'file_path' => $filePath,
                        'file_name' => $row['file_name'],
                        'file_url' => $row['file_url'],
                    ]
                ],
                'product_id' => $row['product_id'],
                'product_material_stock' => floatval($productMaterialStock),
                'product_name' => $row['product_name'],
                'product_name_text' => extractLanguage($row['product_name']),
                'product_price' => $row['product_price'] ?: 0,
                'product_sort' => $row['product_sort'],
                'product_status' => [
                    'text' => $row['status'] === 1 ? __('上架') : __('下架'),
                    'value' => $row['status'] === 1 ? 10 : 20,
                ],
                'product_stock' => floatval($productStock),
                'sales_actual' => floatval($row['sales_actual']),
                'type' => $row['type'],
                'sku' => $sku,
                'is_material_used' => $row['is_material_used'] > 0 ? 1 : 0,
                'product_unit' => $productUnit,
                'product_unit_text' => $productUnitText,
                'is_open_stock' => $row['is_open_stock'],
                'is_package_used' => $row['is_package_used'] > 0 ? 1 : 0,
            ];
        }

        if ($page == 1) {
            return $list;
        } else {
            return [
                'current_page' => $params['page'],
                'data' => $list,
                'per_page' => $params['list_rows'],
                'total' => $total,
                'last_page' => ceil($total / $params['list_rows']),
            ];
        }
    }

    /**
     * 获取商品列表
     */
    public function getLists($param)
    {
        // 商品列表获取条件
        $params = array_merge([
            'product_status' => 10,         // 商品状态
            'category_id' => 0,     // 分类id
        ], $param);
        // 筛选条件
        $model = $this;
        if ($params['category_id'] > 0) {
            $arr = Category::getSubCategoryId($params['category_id']);
            $model = $model->where('category_id', 'IN', $arr);
        }
        if (isset($params['product_name']) && $params['product_name'] != '') {
            $model = $model->like('product_name', trim($params['product_name']));
        }
        if (isset($params['search']) && $params['search'] != '') {
            $model = $model->like('product_name', trim($params['search']));
        }
        $list = $model
            ->with(['category', 'image.file'])
            ->where('product_status', '=', $params['product_status'])
            ->paginate($params);
        // 整理列表数据并返回
        return $this->setProductListData($list, true);
    }

    /**
     * 设置商品展示的数据
     */
    protected function setProductListData($data, $isMultiple = true, callable $callback = null)
    {
        if (!$isMultiple)
            $dataSource = [&$data];
        else
            $dataSource = &$data;
        // 整理商品列表数据
        foreach ($dataSource as &$product) {
            // 商品主图
            $product['product_image'] = $product['image'][0]['file_path'] ?? '';
            // 商品默认规格
            $product['product_sku'] = [];
            if (!$product->product_type == 1) {
                $product['product_sku'] = self::getShowSku($product);
            }
            // 套餐分组
            $product['package'] = self::getPackageGroup($product);
            // 材料是否被使用
            $product['is_material_used'] = 0;
            if ($product['type'] == self::TYPE_MATERIAL) {
                $product['is_material_used'] = $this->checkMaterialUsed($product['product_id']) ? 1 : 0;
            }
            // 商品属性
            $product['product_attr'] = self::getProductAttr($product);
            // 商品加料
            $product['feed'] = self::getProductFeed($product);
            // 税类
            $product['productTaxes'] = self::getProductTaxes($product);

            unset($product['productAttributeGroup']);

            // 回调函数
            is_callable($callback) && call_user_func($callback, $product);
        }

        return $data;
    }

    /**
     * 检测材料是否被使用
     * @param mixed $product_id
     * @return bool
     */
    public function checkMaterialUsed($product_id)
    {
        $sku_exists = ProductFeedMaterial::where('product_feed_id', '>', 0)->where('material_id', $product_id)->find();
        $feed_exists = ProductSkuMaterial::where('product_sku_id', '>', 0)->where('material_id', $product_id)->find();
        return $sku_exists || $feed_exists;
    }

    /**
     * 根据商品id集获取商品列表
     */
    public function getListByIds($productIds, $status = null)
    {
        $model = $this;
        $filter = [];
        // 筛选条件
        $status > 0 && $filter['product_status'] = $status;
        if (!empty($productIds)) {
            $model = $model->orderRaw('field(product_id, ' . implode(',', $productIds) . ')');
        }
        // 获取商品列表数据
        $data = $model->with(['category', 'image.file', 'sku'])
            ->where($filter)
            ->where('product_id', 'in', $productIds)
            ->select();

        // 整理列表数据并返回
        return $this->setProductListData($data, true);
    }

    /**
     * 商品多规格信息
     */
    public function getManySpecData($specRel, $skuData)
    {
        // spec_attr
        $specAttrData = [];
        foreach ($specRel as $item) {
            if (!isset($specAttrData[$item['spec_id']])) {
                $specAttrData[$item['spec_id']] = [
                    'group_id' => $item['spec']['spec_id'],
                    'group_name' => $item['spec']['spec_name'],
                    'spec_items' => [],
                ];
            }
            $specAttrData[$item['spec_id']]['spec_items'][] = [
                'item_id' => $item['spec_value_id'],
                'spec_value' => $item['spec_value'],
            ];
        }
        // spec_list
        $specListData = [];
        foreach ($skuData as $item) {
            $image = (isset($item['image']) && !empty($item['image'])) ? $item['image'] : ['file_id' => 0, 'file_path' => ''];
            $specListData[] = [
                'product_sku_id' => $item['product_sku_id'],
                'spec_sku_id' => $item['spec_sku_id'],
                'rows' => [],
                'spec_form' => [
                    'image_id' => $image['file_id'],
                    'image_path' => $image['file_path'],
                    'product_no' => $item['product_no'],
                    'product_price' => $item['product_price'],
                    'product_weight' => $item['product_weight'],
                    'line_price' => $item['line_price'],
                    'stock_num' => $item['stock_num'],
                    'supplier_price' => $item['supplier_price'],
                ],
            ];
        }
        return ['spec_attr' => array_values($specAttrData), 'spec_list' => $specListData];
    }

    /**
     * 获取商品详情
     * @param $product_id
     */
    public static function detail($product_id)
    {
        /** @var Product $model */
        $model = (new static())->alias('p')
        ->field(['p.*', 'unit.name as product_unit'])
        ->leftJoin('product_unit unit', 'unit.uuid = p.unit_uuid')
        ->with([
            'category',
            'image',
            'sku' => [
                'relatedMaterial' => [
                    'material' => [
                        'unit'
                    ]
                ],
                'productPackageGroupItem',
            ],
            'productAttributeGroup' => function ($q) {
                $q->with([
                    'productAttribute' => function ($q) {
                        $q->with([
                            'attribute'
                        ])->order('sort', 'asc');
                    }
                ])->order('sort', 'asc');
            },
            'feed' => [
                'productSauce' => [
                    'relatedMaterial' => [
                        'material'
                    ]
                ]
            ],
            'dineTax',
            'takeoutTax',
            'productPackageGroup' => [
                'productPackageGropItem' => function ($q) {
                    $q->with([
                        'productBom' => [
                            'product'
                        ]
                    ])->order('sort', 'asc');
                }
            ],
            'productPrinters' => function ($q) {
                $q->field('product_package_uuid,product_printer_uuid,uuid')->with('printerBindNameAndStatus');
            },
        ])->where('p.uuid', '=', $product_id)->find();
        if (empty($model)) {
            return $model;
        }

        $model->is_show_cashier = $model->is_show_cashier != 0 ? 1 : 2;
        $model->is_show_tablet = $model->is_show_tablet != 0 ? 1 : 2;
        $model->is_show_kitchen = $model->is_show_kitchen != 0 ? 1 : 2;
        $model->is_show_assistant = $model->is_show_assistant != 0 ? 1 : 2;
        $model->is_show_h5 = $model->is_show_h5 != 0 ? 1 : 2;

        // 整理商品数据并返回
        return $model->setProductListData($model, false);
    }

    /**
     * 显示的sku，目前取最低价
     */
    public static function getShowSku($product)
    {
        $result = [];

        //如果是单规格
        if ($product['spec_type'] == 10) {
            if (!empty($product['sku'][0] ?? '')) {
                if (count($product['sku'][0]['productPackageGroupItem']) > 0) {
                    $product['sku'][0]['is_package_used'] = 1;
                } else {
                    $product['sku'][0]['is_package_used'] = 0;
                }
                $result[] = $product['sku'][0];
            }
        } else {
            //多规格返回最低价
            foreach ($product['sku'] as &$sku) {
                if ($product['product_price'] == $sku['product_price']) {
                    $result[] = $sku;
                }
                // 显示采购单价为空，如果为0，则为null
                $sku['purchase_price'] = $sku['purchase_price'] > 0 ? $sku['purchase_price'] : 0;
                $material = [];
                foreach ($sku['relatedMaterial'] as $relatedMaterial) {
                    $material[] = [
                        'material_id' => $relatedMaterial['material_uuid'],
                        'material_num' => $relatedMaterial['num'],
                        'materialProduct' => [
                            'product_name_text' => $relatedMaterial['material']['product_name_text'] ?? '',
                            'product_unit_text' => $relatedMaterial['material']['unit']['unit_name_text'] ?? '',
                            'product_material_stock' => $relatedMaterial['material']['stock_num'] ?? 0,
                        ],
                    ];
                }
                $sku['is_package_used'] = 0;
                if (count($sku['productPackageGroupItem']) > 0) {
                    $sku['is_package_used'] = 1;
                }
                $sku['material'] = $material;
                unset($sku['relatedMaterial']);
            }
        }

        // 兼容历史数据，如果找不到返回第一个
        if (empty($result) && !empty($product['sku'][0] ?? '')) {
            $result[] = $product['sku'][0];
        }

        return $result;
    }

    /**
     * 获取商品属性
     */
    public static function getProductAttr($product)
    {
        $productAttr = [];
        foreach ($product->productAttributeGroup as $group) {
            $attributeValue = [];
            $defaultSelect = [];
            $attributeIds = [];
            $attributeValueText = [];
            foreach ($group->productAttribute as $attribute) {
                $attributeValue[] = $attribute->attribute->attribute_name;
                $defaultSelect[] = $attribute->is_default_selected;
                $attributeIds[] = $attribute->attribute_uuid;
                $attributeValueText[] = $attribute->attribute->attribute_name_text;

            }
            $attr = [
                'parent_id' => $group['product_attribute_group_uuid'],
                'parent_name' => $group->attribute->attribute_name_text,
                'attribute_name' => $group->attribute->attribute_name,
                'attribute_value' => $attributeValue,
                'default_select' => $defaultSelect,
                'attribute_ids' => $attributeIds,
                'attribute_max_select' => $group->max_selection,
                'attribute_open_max_select' => $group->max_selection > 0 ? 1 : 0,
                'attribute_required' => $group->is_must,
                'attribute_name_text' => $group->attribute->attribute_name_text,
                'attribute_value_text' => $attributeValueText,
            ];
            $productAttr[] = $attr;
        }

        return $productAttr;
    }

    /**
     * 获取商品加料
     */
    public static function getProductFeed($product)
    {
        $productFeed = [];
        foreach ($product->feed as $feed) {
            $material = [];
            foreach ($feed->productSauce->relatedMaterial as $relatedMaterial) {
                $material[] = [
                    'id' => $relatedMaterial['uuid'],
                    'material_id' => $relatedMaterial['material_uuid'],
                    'material_num' => $relatedMaterial['num'],
                    'product_feed_id' => $feed->product_sauce_uuid,
                    'materialProduct' => [
                        'product_id' => $relatedMaterial['material_uuid'],
                        'product_material_stock' => $relatedMaterial['material']['stock_num'],
                        'product_name' => $relatedMaterial['material']['product_name'],
                        'product_name_text' => $relatedMaterial['material']['product_name_text'],
                        'product_unit' => $relatedMaterial['material']['unit']['unit_name'],
                        'product_unit_text' => $relatedMaterial['material']['unit']['unit_name_text'],
                    ]
                ];
            }
            $productFeed[] = [
                'feed_name_text' => $feed->spec_name_text,
                'feed_id' => $feed->product_sauce_uuid,
                'feed_name' => $feed->spec_name,
                'price' => $feed->product_price,
                'stock_num' => $feed->stock_num,
                'default_select' => $feed->is_default_select,
                'material' => $material,
            ];
        }

        return $productFeed;
    }

    /**
     * 获取商品税类
     */
    public static function getProductTaxes($product)
    {
        return [
            [
                'product_tax_type' => 1,
                'tax_category_id' => $product->dine_tax_uuid,
                'tax_rate' => $product->dine_tax_rate ?? 0,
                'product_id' => $product->uuid,
            ],
            [
                'product_tax_type' => 2,
                'tax_category_id' => $product->takeout_tax_uuid,
                'tax_rate' => $product->takeout_tax_rate ?? 0,
                'product_id' => $product->uuid,
            ],
        ];
    }

    /**
     * 获取套餐分组
     */
    public static function getPackageGroup($product)
    {
        $result = [
            'package_price' => 0,
            'package_stock' => 0,
            'package_group' => [],
        ];
        if ($product['product_type'] == 1) {
            $result['package_price'] = $product['sku'][0]['price']; // 套餐价格
            $result['package_stock'] = $product['sku'][0]['stock_num']; // 套餐库存
            $result['is_open_stock'] = $product['sku'][0]['is_open_stock']; // 是否开启库存
            foreach ($product['productPackageGroup'] as $group) {
                $productList = [];
                foreach ($group['productPackageGropItem'] as $item) {
                    $productList[] = [
                        'item_id' => $item['uuid'], // 套餐分组商品uuid
                        'product_id' => $item['product_bom_uuid'], // 商品package_bom_uuid
                        'product_name_text' => $item['productBom']['product']['product_name_text'], // 商品名称
                        'spec_name_text' => $item['productBom']['spec_name_text'], // 规格名称
                        'product_price' => $item['productBom']['price'], // 商品价格
                        'stock_num' => $item['productBom']['stock_num'], // 商品库存
                        'num' => intval($item['num']), // 商品数量
                        'sort' => $item['sort'], // 排序
                        'add_price' => floatval($item['add_price'] ?? 0), // 加价金额
                        'is_required' => intval($item['is_required'] ?? 0), // 是否必选 0-否 1-是
                        'is_default' => intval($item['is_default'] ?? 0), // 是否默认 0-否 1-是
                    ];
                }
                $result['package_group'][] = [
                    'group_id' => $group['uuid'], // 套餐分组uuid
                    'group_name' => $group['name'], // 套餐分组名称
                    'group_name_text' => $group['group_name_text'], // 套餐分组名称
                    'group_type' => intval($group['group_type'] ?? 0), // 分组类型 0-固定 1-可选
                    'optional_count' => intval($group['optional_count'] ?? 0), // 可选数量
                    'product_list' => $productList, // 套餐分组商品列表
                ];
            }
        }

        return $result;
    }

    /**
     * 获取当前商品总数
     */
    public function getProductTotal($where = [])
    {
        return $this->where($where)->count();
    }

    /**
     * 获取当前商品总数
     */
    public function getSupplierProductTotal($shop_supplier_id, $product_type, $product_status = 0)
    {
        $model = $this;
        if ($product_type >= 0) {
            $model = $model->where('product_type', '=', $product_type);
        }
        if ($product_status > 0) {
            $model = $model->where('product_status', '=', $product_status);
        }
        return $model->count();
    }

    /**
     * 获取店铺销量
     */
    public static function getProductSales($shop_supplier_id)
    {
        $model = new static();
        $sales_initial = $model->sum('sales_initial');
        $sales_actual = $model->sum('sales_actual');
        return $sales_initial + $sales_actual;
    }

    /**
     * 获取商品销量Top10
     */
    public function getProductRank($data, $type = 0)
    {
        $model = new OrderProduct;
        if ($data['sort'] == 1) {
            $order = 'total_num desc';
        } else {
            $order = 'total_price desc';
        }
        if ($data['product_type'] >= 0) {
            $model = $model->where('p.product_type', '=', $data['product_type']);
        }
        if ($data['shop_supplier_id']) {
            $model = $model->where('p.shop_supplier_id', '=', $data['shop_supplier_id']);
        }
        switch ($data['type']) {
            case '1': //今天
                $model = $model->where('o.create_time', '>=', strtotime(date('Y-m-d')));
                break;
            case '2': //近7天
                $model = $model->where('o.create_time', '>=', strtotime(-6 . ' days', strtotime(date('Y-m-d'))));
                break;
            case '3': //近15天
                $model = $model->where('o.create_time', '>=', strtotime(-14 . ' days', strtotime(date('Y-m-d'))));
                break;
            case '4': //自定义
                $start = strtotime($data['time'][0]);
                $end = strtotime($data['time'][1]) + 86399;
                $model = $model->where('o.create_time', 'between', "$start,$end");
                break;
            default:
                $model = $model->where('o.create_time', '=', strtotime(date('Y-m-d')));
                break;
        }
        $model = $model->alias('op')
            ->where('o.pay_status', '=', 20)
            ->where('o.order_status', '<>', 20)
            ->join('order o', 'op.order_id=o.order_id')
            ->join('product p', 'p.product_id=op.product_id')
            ->field('p.product_name,p.product_price,sum(op.total_price) as total_price,sum(total_num) as total_num')
            ->group('op.product_id')
            ->order($order);
        if ($type) {
            $list = $model->select();
        } else {
            $list = $model->paginate($data);
        }
        return $list;
    }


    //判断商品每单限购
    public static function getProductLimitNum($product_id)
    {
        return (new self)->where('product_id', '=', $product_id)
            ->where('product_status', '=', 10)

            ->value('limit_num');
    }

    // 获取加料价格
    public function getFeedPrice($feed_ids)
    {
        $feed_price = 0;
        foreach ($this->feed as $feed) {
            if (in_array($feed['product_feed_id'], $feed_ids)) {
                $feed_price = helper::bcadd($feed_price, $feed['price']);
            }
        }
        return floatval($feed_price);
    }

    /**
     * 过滤商品列表
     * @param $order_id
     * @param $device_type // 当前设备 cashier-收银 tablet-平板 kitchen-厨显
     * @return array
     */
    public static function filterProductList($order_id, $device_type)
    {
        $showBuffetProductIdsArr = [];
        $hideBuffetProductIdsArr = [];
        if ($order_id) {
            $showBuffetProductIdsArr = (new OrderBuffet)->alias('ob')
                ->leftJoin('buffet_product bp', 'ob.buffet_id = bp.buffet_id')
                ->when($device_type == 'cashier', function ($q) {
                    $q->where('is_show_cashier', 1);
                })
                ->when($device_type == 'tablet', function ($q) {
                    $q->where('is_show_tablet', 1);
                })
                ->when($device_type == 'kitchen', function ($q) {
                    $q->where('is_show_kitchen', 1);
                })
                ->distinct(true)
                ->where('order_id', $order_id)->column('product_id');
            $hideBuffetProductIdsArr = (new OrderBuffet)->alias('ob')
                ->leftJoin('buffet_product bp', 'ob.buffet_id = bp.buffet_id')
                ->when($device_type == 'cashier', function ($q) {
                    $q->where('is_show_cashier', 2);
                })
                ->when($device_type == 'tablet', function ($q) {
                    $q->where('is_show_tablet', 2);
                })
                ->when($device_type == 'kitchen', function ($q) {
                    $q->where('is_show_kitchen', 2);
                })
                ->distinct(true)
                ->where('order_id', $order_id)->column('product_id');
            // 显示优先于不显示
            $hideBuffetProductIdsArr = array_diff($hideBuffetProductIdsArr, $showBuffetProductIdsArr);
        }
        $showProductIdsArr = (new Product)
            ->when($device_type == 'cashier', function ($q) {
                $q->where('is_show_cashier', 1);
            })
            ->when($device_type == 'tablet', function ($q) {
                $q->where('is_show_tablet', 1);
            })
            ->when($device_type == 'kitchen', function ($q) {
                $q->where('is_show_kitchen', 1);
            })
            ->distinct(true)->column('product_id');
        return array_diff(array_merge($showBuffetProductIdsArr, $showProductIdsArr), $hideBuffetProductIdsArr);
    }

    /***
     * 销售出库处理（规格，规格材料，加料，加料材料）
     * @param array $sourceProductList 数据源
     * @param int $shopUserId 操作人
     * @param int $shopSupplierId
     * @return bool
     */
    public function salesOut($sourceProductList, $shopUserId = 0, $shopSupplierId = 0)
    {
        // 判断是否开启授权进销存
        if (!$this->hasInventoryAuth()) {
            return true;
        }
        //
        $orderProductList = $sourceProductList['orderProductList'];
        $allProductList = $sourceProductList['allProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];
        //
        $materialIds = [];
        $salesOutInventoryRecords = [];
        foreach ($orderProductList as $orderProduct) {
            if ($orderProduct['is_return'] == 1) {
                continue;
            }
            //
            $num = $orderProduct->total_num;
            $orderId = $orderProduct->order_id;
            $productSkuId = $orderProduct->product_sku_id;
            $productFeedIds = is_array($orderProduct['feed_ids']) ? $orderProduct['feed_ids'] : json_decode($orderProduct['feed_ids']);
            $productFeedIds = $productFeedIds ?: [];
            $product = $allProductList[$orderProduct->product_id];
            $productSku = $allProductSkuList[$productSkuId] ?? [];

            // 判断是否成品
            if ($product['type'] == Product::TYPE_PRODUCT) {
                // 规格库存减少记录
                $salesOutInventoryRecords[] = $this->salesOutInventoryRecord($orderProduct->product_id, $productSkuId, $num, $orderId, $shopUserId, $shopSupplierId, $productSku['spec_name'] ?? '');
                // 减少规格材料库存
                if ($productSku) {
                    foreach ($productSku['material'] as $material) {
                        $skuMaterialNum = $material['material_num'] * $num; // 消耗材料的数量 = 消耗材料数 * 订单数量
                        //
                        $materialIds = array_merge($materialIds, [$material['material_id']]);
                        $materialProductSkuId = ProductSku::where('product_id', '=', $material['material_id'])->value('product_sku_id');
                        Product::where('product_id', '=', $material['material_id'])->update(['product_material_stock' => Db::raw('product_material_stock - ' . $skuMaterialNum), 'sales_actual' => Db::raw('sales_actual + ' . $skuMaterialNum)]);
                        ProductSku::where('product_sku_id', '=', $materialProductSkuId)->update(['material_stock' => Db::raw('material_stock - ' . $skuMaterialNum), 'product_sales' => Db::raw('product_sales + ' . $skuMaterialNum)]); // 材料规格单一，主表的product_id = 规格表的product_id
                        // 规格材料库存减少记录
                        $salesOutInventoryRecords[] = $this->salesOutInventoryRecord($material['material_id'], $materialProductSkuId, $skuMaterialNum, $orderId, $shopUserId, $shopSupplierId, $productSku['spec_name']);
                    }
                }
                // 减少加料材料库存
                foreach ($productFeedIds as $productFeedId) {
                    $productFeed = ProductFeed::detail($productFeedId);
                    if ($productFeed) {
                        $productFeed->where(['product_feed_id' => $productFeedId])->dec('stock_num', $num)->update();
                        foreach ($productFeed->material as $material) {
                            $feedMaterialNum = $material['material_num'] * $num;
                            //
                            $materialIds = array_merge($materialIds, [$material['material_id']]);
                            $materialProductSkuId = ProductSku::where('product_id', '=', $material['material_id'])->value('product_sku_id');
                            Product::where('product_id', '=', $material['material_id'])->update(['product_material_stock' => Db::raw('product_material_stock - ' . $feedMaterialNum), 'sales_actual' => Db::raw('sales_actual + ' . $feedMaterialNum)]);
                            ProductSku::where('product_sku_id', '=', $materialProductSkuId)->update(['material_stock' => Db::raw('material_stock - ' . $feedMaterialNum), 'product_sales' => Db::raw('product_sales + ' . $feedMaterialNum)]); // 材料规格单一，主表的product_id = 规格表的product_id
                            // 加料材料库存减少记录
                            $salesOutInventoryRecords[] = $this->salesOutInventoryRecord($material['material_id'], $materialProductSkuId, $feedMaterialNum, $orderId, $shopUserId, $shopSupplierId, $productSku['spec_name']);
                        }
                    }
                }
            }
        }
        //
        if ($salesOutInventoryRecords) {
            (new ErpInventoryRecord)->limit(100)->replace()->insertAll($salesOutInventoryRecords);
        }
        // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
        if ($materialIds) {
            $this->reCalProductStock(array_unique($materialIds));
        }
        //
        return true;
    }

    /**
     * 销售出库记录
     */
    public function salesOutInventoryRecord($productId, $productSkuId, $num, $orderId, $shopUserId, $shopSupplierId, $productSkuName)
    {
        $inventoryRecordData = [
            'product_id' => $productId,
            'product_sku_id' => $productSkuId,
            'product_sku_name' => $productSkuName,
            'order_id' => $orderId,
            'operator_id' => $shopUserId ?? 0,
            'remark' => '',
        ];

        // 减少库存 出库记录
        $inventoryRecordData['num'] = $num;
        $inventoryRecordData['type'] = ErpInventoryRecord::TYPE_SALE_OUT;
        $inventoryRecordData['status'] = ErpInventoryRecord::STATUS_OUT;
        return (new ErpInventoryRecord)->addNew(ErpInventoryRecord::INVENTORY_TYPE_OUT, $inventoryRecordData, false);
    }

    /**
     * 反结账销售出库处理（规格，规格材料，加料，加料材料）
     * @param $orderId 订单id
     */
    public function salesOutReverse($orderId)
    {
        $erpInventoryRecord = ErpInventoryRecord::where('order_id', $orderId)->where('status', ErpInventoryRecord::STATUS_OUT)->select();
        $materialIds = [];
        foreach ($erpInventoryRecord as $record) {
            if ($record->inventory_type == ErpInventoryRecord::INVENTORY_TYPE_OUT) {
                // 出库撤销操作，回滚增加库存
                switch ($record->product['type']) {
                    case Product::TYPE_MATERIAL:
                        $materialIds = array_merge($materialIds, [$record->product_id]);
                        Product::where('product_id', '=', $record->product_id)->update(['product_material_stock' => Db::raw('product_material_stock + ' . $record->num), 'sales_actual' => Db::raw('sales_actual - ' . $record->num)]);
                        ProductSku::where('product_sku_id', '=', $record->product_sku_id)->update(['material_stock' => Db::raw('material_stock + ' . $record->num), 'product_sales' => Db::raw('product_sales - ' . $record->num)]);
                        break;
                }
            }
            $record->status = ErpInventoryRecord::STATUS_REVOKED;
            $record->revoke_time = time();
            $record->save();
        }
        // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
        if (!empty($materialIds)) {
            $this->reCalProductStock(array_unique($materialIds));
        }
    }

    /**
     * 判断是否有进销存权限
     */
    public function hasInventoryAuth()
    {
        $licenses = request()->licenses;
        if (isset($licenses['sale']) && $licenses['sale'] == 1) {
            return true;
        }
        return false;
    }

    /**
     * 重新根据材料计算产品总库存、产品规格库存、加料库存
     */
    public function reCalProductStock($materialIds)
    {
        // 如果是进销存，则重新计算产品总库存
        if (!$this->hasInventoryAuth() || empty($materialIds)) {
            return;
        }

        $prefix = Env::get('DB_PREFIX');
        $materialIds = implode(',', $materialIds);

        $db = Db::connect($this->getConnection());
        $db->startTrans();
        try {
            // 更新product_sku表的库存
            $db->execute("
                UPDATE {$prefix}product_sku AS ps
                JOIN (
                    SELECT psm.product_sku_id, LEAST(FLOOR(MIN(p.product_material_stock / psm.material_num)), 99999999) AS min_stock_num
                    FROM {$prefix}product_sku_material AS psm
                    JOIN {$prefix}product AS p ON psm.material_id = p.product_id
                    GROUP BY psm.product_sku_id
                ) AS sub ON ps.product_sku_id = sub.product_sku_id
                SET ps.stock_num = sub.min_stock_num
                WHERE ps.product_sku_id IN (
                    SELECT product_sku_id
                    FROM {$prefix}product_sku_material
                    WHERE material_id IN ({$materialIds})
                );
            ");
            $db->execute("
                DELETE FROM {$prefix}sync WHERE (type = 'product_sku' and like_id IN (
                    SELECT product_sku_id
                    FROM {$prefix}product_sku_material
                    WHERE material_id IN ({$materialIds})
                )) or (type = 'product_sku_material' and like_id IN ({$materialIds}));
            ");

            // 更新product_feed表的库存
            $db->execute("
                UPDATE {$prefix}product_feed AS pf
                JOIN (
                    SELECT pfm.product_feed_id, LEAST(FLOOR(MIN(p.product_material_stock / pfm.material_num)), 99999999) AS min_stock_num
                    FROM {$prefix}product_feed_material AS pfm
                    JOIN {$prefix}product AS p ON pfm.material_id = p.product_id
                    GROUP BY pfm.product_feed_id
                ) AS sub ON pf.product_feed_id = sub.product_feed_id
                SET pf.stock_num = sub.min_stock_num
                WHERE pf.product_feed_id IN (
                    SELECT product_feed_id
                    FROM {$prefix}product_feed_material
                    WHERE material_id IN ({$materialIds})
                );
            ");
            $db->execute("
                DELETE FROM {$prefix}sync WHERE (type = 'product_feed' and like_id IN (
                    SELECT product_feed_id
                    FROM {$prefix}product_feed_material
                    WHERE material_id IN ({$materialIds})
                )) or (type = 'product_feed_material' and like_id IN ({$materialIds}));
            ");

            // 更新product表的库存
            $db->execute("
                UPDATE {$prefix}product AS p
                SET p.product_stock = (
                    SELECT SUM(ps.stock_num)
                    FROM {$prefix}product_sku AS ps
                    WHERE p.product_id = ps.product_id
                )
                WHERE p.product_id IN (
                    SELECT ps.product_id
                    FROM {$prefix}product_sku AS ps
                    WHERE ps.product_sku_id IN (
                        SELECT psm.product_sku_id
                        FROM {$prefix}product_sku_material AS psm
                        WHERE psm.material_id IN ({$materialIds})
                    )
                );
            ");
            $db->execute("
                DELETE FROM {$prefix}sync WHERE (type = 'product' and like_id IN (
                    SELECT ps.product_id
                    FROM {$prefix}product_sku AS ps
                    WHERE ps.product_sku_id IN (
                        SELECT psm.product_sku_id
                        FROM {$prefix}product_sku_material AS psm
                        WHERE psm.material_id IN ({$materialIds})
                    )
                )) or (type = 'product' and like_id IN ({$materialIds}));
            ");

            //
            $db->commit();
        } catch (\Exception $e) {
            // 出现异常时，回滚事务
            $db->rollback();
            // 记录或处理异常
            trace('Error: ' . $e->getMessage());
        }
    }

    // 获取商品库存
    public static function getProductStockById($product_id)
    {
        $product = self::where('product_id', $product_id)->find();

        $stock = 0;
        if ($product['type'] == 10) {
            $stock = $product->product_stock;
        }
        if ($product['type'] == 20) {
            $stock = $product->product_material_stock;
        }
        return $stock;
    }

    /**
     * 获取商品消费税
     * @param $rate
     * @param $product_price
     * @param $is_tax           // 是否含税 1-已含税 2-未含税
     * @return float|int
     */
    public static function getConsumptionTax($rate, $product_price, $is_tax)
    {
        if (!$rate) {
            return 0;
        }
        if ($is_tax == 1) {
            /**
             * 商品价格含税
             */
            // 商品税前价 = 商品价格 / （1 + 税率）
            $original_price = helper::bcdiv($product_price, helper::bcadd(1, helper::bcdiv($rate, 100, 7), 7), 3);  //  $product_price / (1 + $rate/100)
            $original_price = round($original_price, 2);    // 四舍五入保留两位
            // 消费税 = 商品价格 - 商品税前价
            $tax_price = helper::bcsub($product_price, $original_price);
        } else {
            /**
             * 商品价格不含税
             */
            // 消费税 = 商品价格 * 税率
            $tax_price = helper::bcmul($product_price,  helper::bcdiv($rate, 100, 7), 3);
            $tax_price = round($tax_price, 2);  // 四舍五入保留两位
        }
        return floatval($tax_price);
    }

    /**
     * 获取删除后的商品ids
     *
     * @return array
     */
    public static function getDeleteProductIds()
    {
        // 删除或者下架的商品
        $ids = self::where('delete_time', 1)->whereOr('product_status', 20)->order('update_time', 'desc')->limit(100)->column('product_id');
        return $ids;
    }

    /**
     * 获取商品服务费
     * @param $product_price
     * @param $product_service_rate // 商品服务费率
     * @param $is_tax
     * @param $rate
     * @return float
     */
    public static function getProductServiceFee($product_price, $product_service_rate, $is_tax, $rate)
    {
        if ($product_service_rate <= 0) {
            return 0;
        }
        $no_tax_unit_product_price = self::getNoTaxUnitPrice($rate, $product_price, $is_tax);
        $product_service = helper::bcmul($no_tax_unit_product_price,  helper::bcdiv($product_service_rate, 100, 7), 3);
        return round($product_service, 2);  // 四舍五入保留两位
    }

    /**
     * 获取商品无税价
     * @param $rate
     * @param $product_price
     * @param $is_tax   // 是否含税 1-已含税 2-未含税
     * @return float
     */
    public static function getNoTaxUnitPrice($rate, $product_price, $is_tax)
    {
        if (!$rate) {
            return $product_price;
        }
        if ($is_tax == 1) {
            /**
             * 商品价格含税
             */
            // 商品税前价 = 商品价格 / （1 + 税率）
            $original_price = helper::bcdiv($product_price, helper::bcadd(1, helper::bcdiv($rate, 100, 7), 7), 3);  //  $product_price / (1 + $rate/100)
            return round($original_price, 2);
        }
        return $product_price;
    }

    /**
     * 商品服务费的消费税
     * @param $product_service_fee
     * @param $rate
     * @return float
     */
    public static function getProductServiceConsumptionTax($product_service_fee, $rate)
    {
        if (!$rate) {
            return 0;
        }
        $product_service_consumption_tax = helper::bcmul($product_service_fee, helper::bcdiv($rate, 100, 7), 3);  //  $product_price * ($rate/100)
        return round($product_service_consumption_tax, 2);
    }

    /**
     * 批量修改商品分类
     */
    public function batchUpdateCategory($params)
    {
        $product_ids = $params['product_ids'] ?? '';
        $category_id = $params['category_id'] ?? 0;
        //
        if (empty($product_ids)) {
            $this->error = '请选择商品';
            return false;
        }
        if (!Category::detail($category_id)) {
            $this->error = '请选择商品分类';
            return false;
        }

        $productUuidList = []; // 商品uuid列表
        $materialUuidList = []; // 材料uuid列表
        
        foreach ($product_ids as $product_id) {
            $product = self::detail($product_id);
            if ($product) {
                $productUuidList[] = $product['uuid'];
                continue;
            }
            $material = Material::detail($product_id);
            if ($material) {
                $materialUuidList[] = $material['uuid'];
            }
        }
        // 更新商品分类
        if (!empty($productUuidList)) {
            $this->where('uuid', 'in', $productUuidList)->update(['category_uuid' => $category_id]);
        }
        // 更新材料分类
        if (!empty($materialUuidList)) {
            Material::where('uuid', 'in', $materialUuidList)->update(['category_uuid' => $category_id]);
        }

        return true;
    }

    /**
     * 批量修改商品分类
     */
    public function batchUpdateTax($params)
    {
        $product_ids = $params['product_ids'] ?? '';
        $product_taxes = $params['productTaxes'] ?? [];
        //
        if (empty($product_ids)) {
            $this->error = '请选择商品';
            return false;
        }
        if (empty($product_taxes)) {
            $this->error = '请选择税务信息';
            return false;
        }
        //
        foreach ($product_taxes as $item) {
            $product_tax_type = $item['product_tax_type'];
            $tax_category_id = $item['tax_category_id'];
            if ($product_tax_type == 1) {
                Product::where('uuid', 'in', $product_ids)->update(['dine_tax_uuid' => $tax_category_id]);
            } else {
                Product::where('uuid', 'in', $product_ids)->update(['takeout_tax_uuid' => $tax_category_id]);
            }
        }
        return true;
    }

    /**
     * 批量修改整单折扣
     * 
     * 逻辑说明：
     * 1. 将参数中的商品uuid列表（product_ids）对应的商品整单折扣(open_overall_discount)字段设置为0（关闭整单折扣）。
     * 2. 将不在参数uuid列表中，并且整单折扣(open_overall_discount)字段为0的商品整单折扣(open_overall_discount)字段设置为1（开启整单折扣）。
     * 3. 操作需使用模型事务，保证数据一致性。
     * 4. 若未选择商品，返回错误。
     */
    public function batchUpdateOpenOverallDiscount($params)
    {
        $product_ids = $params['product_ids'] ?? [];
        // 开启事务，保证批量操作的原子性
        Db::startTrans();
        try {
            // 1. 关闭选中商品的整单折扣（open_overall_discount=0）
            if (!empty($product_ids)) {
                $this->whereIn('uuid', $product_ids)
                    ->update(['open_overall_discount' => 0]);
            }

            // 2. 开启未选中商品的整单折扣（open_overall_discount=1）
            $builder = $this->where('open_overall_discount', 0);
            if (!empty($product_ids)) {
                $builder = $builder->whereNotIn('uuid', $product_ids);
            }
            $builder->update(['open_overall_discount' => 1]);
                

            // 提交事务
            Db::commit();
            return true;
        } catch (\Exception $e) {
            // 回滚事务
            Db::rollback();
            return false;
        }
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$lang'))"), '=', $name],
            'delete_time' => 0,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        $productPackageSql = self::where($filter)->field('name')->buildSql();
        $materialSql = Material::where($filter)->field('name')->buildSql();
        $dbName = 'shop' . self::$app_id;
        $results = Db::connect($dbName)->query("SELECT COUNT(*) FROM ($productPackageSql UNION ALL $materialSql) AS combined_names");
        $count = array_column($results, 'COUNT(*)')[0] ?? 0;
        return $count > 0 ? true : false;
    }

    /**
     * 检查产品图片名称唯一性
     */
    public function checkProductImgExist($name, $id = null)
    {
        $filter = [
            'image_name' => $name,
            'delete_time' => 0,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        $productPackageSql = self::where($filter)->field('image_name')->buildSql();
        $materialSql = Material::where($filter)->field('image_name')->buildSql();
        $dbName = 'shop' . self::$app_id;
        $results = Db::connect($dbName)->query("SELECT COUNT(*) FROM ($productPackageSql UNION ALL $materialSql) AS combined_names");
        $count = array_column($results, 'COUNT(*)')[0] ?? 0;
        return $count > 0 ? true : false;
    }

    /**
     * 获取所有图片
     */
    public static function getAllImg()
    {
        $data = Product::alias('a')
            ->leftJoin('product_image img', 'img.product_id = a.product_id')
            ->leftJoin('upload_file file', 'file.file_id = img.image_id')
            ->where('img.id', '>', 0)
            ->where('file.file_id', '>', 0)
            ->where('a.delete_time', '=', 0)
            ->where('a.type', '=', 10)    // 10-成品 20-材料
            ->where('a.product_type', '=', 1)
            ->where('a.product_status', '=', 10)
            ->field('file.storage, file.url_param, file.save_name, file.file_url, file.file_name')
            ->select();
        //
        $result = [];
        $uploadFile = new UploadFile;
        foreach ($data as $value) {
            $result[] = $uploadFile->getFilePathAttr(null, $value);
        }
        //
        return $result;
    }
}
