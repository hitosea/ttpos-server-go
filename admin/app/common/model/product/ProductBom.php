<?php

namespace app\common\model\product;

use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Category;
use app\common\model\erp\ErpPurchaseDetail;
use app\common\model\erp\ErpDamagedProductRecord;
use app\common\model\erp\ErpWarehouseForm;
use app\common\model\file\UploadFile AS UploadFileModel;
use app\common\model\product\Product AS ProductModel;
use think\facade\Db;

/**
 * 商品BOM模型
 */
class ProductBom extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_bom';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

    /**
     * 追加字段
     */
    protected $append = ['product_sku_id', 'product_id', 'spec_name_text', 'spec_name', 'spec_sku_id', 'barcodeUniqueness', 'barcode', 'product_price'];

    /**
     * 规格SKU ID
     */
    public function getProductSkuIdAttr($value, $data = [])
    {
        return $this->uuid ?: 0;
    }

    /**
     * 产品ID
     */
    public function getProductIdAttr($value, $data = [])
    {
        return $this->product_package_uuid ?: 0;
    }

    /**
     * 规格名称
     */
    public static function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
    }

    /**
     * 规格名称, JSON字符串
     */
    public static function getSpecNameAttr($value, $data)
    {
        return $data['name'] ?? '';
    }

    /**
     * 规格SKU ID
     */
    public static function getSpecSkuIdAttr($value, $data)
    {
        return $data['product_flavor_uuid'] ?? 0;
    }

    /**
     * 条形码唯一性
     */
    public static function getBarcodeUniquenessAttr($value, $data)
    {
        return $data['barcode_value'] ? true : false;
    }

    /**
     * 条形码
     */
    public static function getBarcodeAttr($value, $data)
    {
        return $data['barcode_value'] ?? '';
    }

    /**
     * 产品价格
     */
    public static function getProductPriceAttr($value, $data)
    {
        return $data['price'] ?? 0;
    }

    /**
     * 成品库存
     */
    public function getStockNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 材料库存
     */
    public function getMaterialStockAttr($value)
    {
        return floatval($value);
    }

    /**
     * 销量
     */
    public function getProductSalesAttr($value)
    {
        return floatval($value);
    }

    /**
     * 规格图片
     */
    public function image()
    {
        return $this->hasOne('app\\common\\model\\file\\UploadFile', 'file_id', 'image_id');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo(ProductModel::class, 'product_package_uuid', 'uuid')->with(['image', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
    }

    /**
     * 关联材料
     */
    public function relatedMaterial()
    {
        return $this->hasMany('app\\common\\model\\product\\RelatedMaterial', 'related_uuid', 'uuid');
    }

    /**
     * 通过规格获取商品SKU列表
     */
    public static function getProductBomList($params, $filterHavingMaterial = 0)
    {
        // 商品列表获取条件
        $params = array_merge([
            'material_type' => 0, // 搜索商品类型: 10成品 20材料
            'product_status' => 0, // 搜索商品状态: 10开启 20关闭
            'keyword' => '', // 搜索商品名称/条码
            'page' => 1, // 当前页
            'list_rows' => 15, // 每页数量
            'sort' => '', // 排序库存: asc, desc
            'stock_num' => '', // 搜索库存范围
        ], $params);

        // 规格
        $productBomSql = self::alias('bom')
            ->field(implode(',', [
                '10 as type',
                'bom.create_time',
                'file.file_type',
                'file.file_url',
                'file.file_name',
                'file.storage',
                'file.save_name',
                'file.url_param',
                'p.name as product_name',
                'bom.price as product_price',
                'c1.name as category_name', 
                'c2.name as category_parent_name',
                'c1.uuid as category_uuid',
                'c2.uuid as category_parent_uuid',
                'bom.name as spec_name',
                'bom.actual_sale_num', 
                'bom.stock_num',
                'p.status',
                'bom.uuid',
                'bom.barcode_value',
                'count(rm.uuid) as material_count',
                's.name as supplier_name',
                'bom.update_time as update_time',
            ]))
            ->leftJoin('product_package p', 'bom.product_package_uuid = p.uuid')
            ->leftJoin('supplier s', 'p.supplier_uuid = s.uuid')
            ->leftJoin('file', 'p.image_file_uuid = file.uuid')
            ->leftJoin('product_category c1', 'p.category_uuid = c1.uuid')
            ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
            ->leftJoin('related_material rm', 'bom.uuid = rm.related_uuid')
            ->where('bom.product_flavor_uuid', '>', 0)
            ->group('bom.uuid')
            ->buildSql();

        // 材料
        $materialSql = Material::alias('m')
            ->field(implode(',', [
                '20 as type', 
                'm.create_time',
                'file.file_type',
                'file.file_url',
                'file.file_name',
                'file.storage',
                'file.save_name',
                'file.url_param',
                'm.name as product_name',
                '0 as product_price',
                'c1.name as category_name', 
                'c2.name as category_parent_name',
                'c1.uuid as category_uuid',
                'c2.uuid as category_parent_uuid',
                '"" as spec_name',
                'm.actual_sale_num',
                'm.stock_num', 
                'm.status',
                'm.uuid',
                'm.barcode_value',
                '0 as material_count',
                's.name as supplier_name',
                'm.update_time as update_time',
            ]))
            ->leftJoin('product_category c1', 'm.category_uuid = c1.uuid')
            ->leftJoin('product_category c2', 'c1.parent_uuid = c2.uuid')
            ->leftJoin('file', 'm.image_uuid = file.uuid')
            ->leftJoin('supplier s', 'm.supplier_uuid = s.uuid')
            ->group('m.uuid')
            ->buildSql();

        // 分页
        $offset = ($params['page'] - 1) * $params['list_rows'];
        $limit = $params['list_rows'];

        // 搜索
        $bind = [];
        $whereSql = '';

        // 搜索商品类型
        $materialType = $params['material_type'] ?? 0;
        if (in_array($materialType, [ProductModel::TYPE_MATERIAL, ProductModel::TYPE_PRODUCT])) {
            $whereSql .= " WHERE (type = :type)";
            $bind['type'] = $materialType;
        }

        // 搜索商品状态
        $status = $params['product_status'] ?? 0;
        if (in_array($status, [10, 20])) {
            $where = "(status = :status)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['status'] = $status == 10 ? 1 : 0;
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

        // 搜索库存
        $stockNum = $params['stock_num'] ?? '';
        if ($stockNum) {
            $where = "(stock_num < :stock_num)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['stock_num'] = $stockNum;
        }

        // 搜索商品名称/条码
        $keyword = $params['keyword'] ?? '';
        if ($keyword != '') {
            $where = "(product_name LIKE :product_name OR barcode_value LIKE :barcode_value)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
            $bind['product_name'] = "%{$keyword}%";
            $bind['barcode_value'] = "%{$keyword}%";
        }

        // 过滤有材料的规格
        if ($filterHavingMaterial) {
            $where = "(material_count = 0)";
            if (!$whereSql) {
                $whereSql .= " WHERE {$where}";
            } else {
                $whereSql .= " AND {$where}";
            }
        }

        // 排序
        $sort = $params['sort'] ?? '';
        $orderSql = ' ORDER BY update_time DESC';
        if ($sort && in_array($sort, ['asc', 'desc'])) {
            $orderSql = " ORDER BY stock_num {$sort}";
        }

        $querySql = "SELECT " . implode(',', [
            'type',
            'create_time',
            'file_type',
            'file_url',
            'file_name',
            'storage',
            'save_name',
            'url_param',
            'product_name',
            'product_price',
            'category_name', 
            'category_parent_name', 
            'category_uuid',
            'category_parent_uuid',
            'spec_name',
            'actual_sale_num',
            'stock_num',
            'status',
            'uuid',
            'barcode_value',
            'material_count',
            'supplier_name',
            'update_time',
        ]) . " FROM ($productBomSql UNION ALL $materialSql) AS all_product";
        $pageSql = " LIMIT {$offset}, {$limit}";

        // 执行查询
        $countSql = "SELECT COUNT(*) AS total_count". " FROM ($productBomSql UNION ALL $materialSql) AS all_product";
        $count = Db::connect('shop' . self::$app_id)->query($countSql . $whereSql, $bind);
        $total = $count[0]['total_count'];
        $rows = Db::connect('shop' . self::$app_id)->query($querySql . $whereSql . $orderSql . $pageSql, $bind);

        $file = new UploadFileModel();
        $list = [];
        foreach ($rows as $row) {
            // 图片
            $filePath = $file->getFilePathAttr(null, [
                'file_type' => $row['file_type'],
                'file_url' => $row['file_url'],
                'file_name' => $row['file_name'],
                'storage' => $row['storage'],
                'save_name' => $row['save_name'],
                'url_param' => $row['url_param'],
            ]);
            // 分类
            $pathNameText = extractLanguage($row['category_name']);
            if ($row['category_parent_name']) {
                $pathNameText = extractLanguage($row['category_parent_name']) . '-' . $pathNameText;
            }
            // 规格名称
            $specNameText = '';
            if ($row['type'] == ProductModel::TYPE_PRODUCT) {
                $specNameText = extractLanguage($row['spec_name']);
            }
            // 库存
            $productStock = 0;
            $productMaterialStock = 0;
            $historyPurchaseNum = 0; // 历史进货数
            if ($row['type'] == 10) {
                $productStock = $row['stock_num'];
                $historyPurchaseNum = ErpWarehouseForm::where('product_bom_uuid', $row['uuid'])->where('status', 0)->sum('num') ?: 0;
            } else {
                $productMaterialStock = $row['stock_num'];
                $historyPurchaseNum = ErpWarehouseForm::where('material_uuid', $row['uuid'])->where('status', 0)->sum('num') ?: 0;
            }
            // todo 历史报损数
            $historyLossNum = 0;
            
            $list[] = [
                'product_id' => $row['uuid'],
                'product_sku_id' => $row['uuid'],
                'product' => [
                    'type' => $row['type'],
                    'image' => [
                        [
                            'file_path' => $filePath,
                            'file_name' => $row['file_name'],
                            'file_url' => $row['file_url'],
                        ]
                    ],
                    'product_name_text' => extractLanguage($row['product_name']),
                    'category' => [ 'path_name_text' => $pathNameText ],
                    'product_status' => [ 'value' => $row['status'] === 1 ? 10 : 20 ],
                    'erpSupplier' => [
                        [ 'name' => $row['supplier_name'] ?: '', ]
                    ],
                ],
                'product_price' => $row['product_price'],
                'create_time' => date('Y-m-d H:i:s', $row['create_time']),
                'spec_name_text' => $specNameText,
                'product_sales' => floatval($row['actual_sale_num']),
                'stock_num' => floatval($productStock),
                'material_stock' => floatval($productMaterialStock),
                'product_sku_id' => $row['uuid'],
                'update_time' => date('Y-m-d H:i:s', $row['update_time']),
                'history_purchase_num' => floatval($historyPurchaseNum),
                'history_loss_num' => floatval($historyLossNum),
            ];
        }

        return [
            'current_page' => $params['page'],
            'data' => $list,
            'per_page' => $params['list_rows'],
            'total' => $total,
            'last_page' => ceil($total / $params['list_rows']),
        ];


        // return (new self())->alias('bom')
        //     ->with(['product'])
        //     ->field(['bom.*'])
        //     ->where('bom.product_flavor_uuid', '>', 0)
        //     ->paginate($params);

        // todo 兼容
        // $model = (new ProductSku())->alias('ps')
        //     ->field('ps.*, p.category_id, p.erp_supplier_id, IF(p.type=10, ps.stock_num, ps.material_stock) as stock_num')
        //     ->leftJoin('product p', 'ps.product_id = p.product_id')
        //     ->where('p.is_delete', '=', 0)
        //     ->with(['product' => function ($query) {
        //         $query->with(['image', 'image.file', 'category', 'erpSupplier', 'erpSupplier.purchaser']);
        //     }]);

        // // 过滤产品规格有材料关联的数据
        // if ($filterHavingMaterial) {
        //     $model = $model->leftJoin('product_sku_material psm', 'psm.product_sku_id = ps.product_sku_id')
        //         ->where('psm.product_sku_id', null);
        // }

        // // 规格库存
        // if (isset($params['stock_num']) && $params['stock_num'] > 0) {
        //     $model = $model->where(function ($query) use ($params) {
        //         $query->where(function ($query) use ($params) {
        //             $query->where('p.type', '=', 10)
        //                 ->where('ps.stock_num', '<', $params['stock_num']);
        //         })->whereOr(function ($query) use ($params) {
        //             $query->where('p.type', '<>', 10)
        //                 ->where('ps.material_stock', '<', $params['stock_num']);
        //         });
        //     });
        // }

        // // 类型
        // if (isset($params['material_type']) && $params['material_type'] > 0) {
        //     $model->where('p.type', '=', $params['material_type']);
        // }

        // // 状态
        // if (isset($params['product_status']) && $params['product_status'] > 0) {
        //     $model->where('p.product_status', '=', $params['product_status']);
        // }

        // // 分类
        // if (!empty($params['category_id'])) {
        //     $categoryIds = (new Category)
        //         ->where('category_id', $params['category_id'])
        //         ->whereOr('parent_id', $params['category_id'])
        //         ->column('category_id');
        //     $model->where('p.category_id', 'IN', $categoryIds);
        // }

        // // 商品名称/条形码
        // if (isset($params['keyword']) && $params['keyword'] != '') {
        //     $keyword = trim($params['keyword']);
        //     $model->where(function ($query) use ($keyword) {
        //         $query->jsonLike('p.product_name', $keyword);
        //         $query->orLike('ps.barcode', $keyword);
        //     });
        // }

        // // 规格库存排序
        // if (isset($params['sort']) && $params['sort'] != '') {
        //     $model = $model->order('stock_num', $params['sort']);
        // }
        // $model = $model->order('ps.update_time', 'desc')->order('ps.product_sku_id');
        // $list = $model->paginate($params);

        // //
        // foreach ($list as &$item) {
        //     // 历史进货数
        //     $item['history_purchase_num'] = (new ErpPurchaseDetail())->sumActualPurchaseNum($item['product_sku_id']);
        //     // 历史报损数
        //     $item['history_loss_num'] = (new ErpDamagedProductRecord())->sumDamagedProductNum($item['product_sku_id']);
        // }
        // return $list;
    }

    /**
     * 检查产品条码唯一性
     * @param string $name 条码
     * @param int $uuid product_package唯一标识uuid
     * @return bool
     */
    public function checkProductBarcodeExist($name, $uuid = null)
    {
        $productBomFilter = [
            ['barcode_value', '=', $name],
            ['barcode_value', '<>', ''],
            ['delete_time', '=', 0],
        ];
        $materialFilter = $productBomFilter;
        if (!is_null($uuid) && $uuid != 0) {
            $productBomFilter[] = ['product_package_uuid', '<>', $uuid];
            $materialFilter[] = ['uuid', '<>', $uuid];
        }
        $productBomSql = self::where($productBomFilter)->field('uuid')->buildSql();
        $materialSql = Material::where($materialFilter)->field('uuid')->buildSql();
        $dbName = 'shop' . self::$app_id;
        $results = Db::connect($dbName)->query("SELECT COUNT(*) FROM ($productBomSql UNION ALL $materialSql) AS combined_names");
        $count = array_column($results, 'COUNT(*)')[0] ?? 0;
        return $count > 0 ? true : false;
    }
}
