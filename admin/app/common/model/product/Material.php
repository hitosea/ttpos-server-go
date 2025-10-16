<?php

namespace app\common\model\product;

use app\common\library\helper;
use help\ValidateHelp;
use app\common\model\BaseModel;
use app\common\model\erp\ErpInventoryRecord;
use app\common\model\erp\ErpMonthlyMaterialStatistics;
use app\common\model\erp\ErpSupplier;
use app\common\model\erp\ErpWarehouseForm;
use app\common\model\erp\ErpWarehouseOutForm;
use app\common\model\erp\ErpWarehouse;
use app\common\model\erp\ErpWarehouseItem;
use app\shop\service\CheckService;
use app\common\model\file\UploadFile;
use app\common\model\product\RelatedMaterial as ProductRelatedMaterial;
use app\common\model\store\MultiLanguageName;
use app\common\model\supplier\Supplier;
use app\shop\model\product\Product;
use app\shop\model\product\RelatedMaterial;
use think\facade\Db;
use think\model\concern\SoftDelete;

/**
 * 原料信息表
 */
class Material extends BaseModel
{
    use SoftDelete;

    protected $name = 'material';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

    /**
     * 追加字段
     */
    protected $append = [
        'type',
        'product_id',
        'product_name',
        'category_id',
        'erp_supplier_id',
        'img_name',
        'unit_id',
        'product_status',
        'product_name_text',
        'product_unit_text',
    ];

    /**
     * 获取类型
     */
    public function getTypeAttr($value, $data = [])
    {
        return 20;
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

    public function getProductStatusAttr($value, $data)
    {
        $value = $this->status ? 10 : 20;
        $status = [10 => __('上架'), 20 => __('下架')];
        return ['text' => $status[$value], 'value' => $value];
    }

    /**
     * 获取商品名称
     */
    public function getProductNameTextAttr($value, $data)
    {
        return extractLanguage($value ?: $data['name'] ?? '');
    }

    /**
     * 关联商品图片表
     */
    public function image()
    {
        return $this->belongsTo('app\\common\\model\\file\\UploadFile', 'image_uuid', 'uuid')->order(['id' => 'asc']);
    }

    /**
     * 关联加料
     */
    public function feed()
    {
        return $this->belongsToMany(Feed::class, RelatedMaterial::class, 'material_uuid', 'related_uuid');
    }

    /**
     * 关联单位
     */
    public function unit()
    {
        return $this->belongsTo(Unit::class, 'unit_uuid', 'uuid');
    }

    /**
     * 关联产品语言
     */
    public function MultiLanguageName()
    {
        return $this->belongsTo(MultiLanguageName::class, 'multi_language_name_uuid', 'uuid');
    }

    /**
     * 关联关联材料
     */
    public function relatedMaterial()
    {
        return $this->hasMany(ProductRelatedMaterial::class, 'material_uuid', 'uuid');
    }

    /**
     * 关联供应商
     */
    public function supplier()
    {
        return $this->belongsTo(Supplier::class, 'supplier_uuid', 'uuid');
    }

    /**
     * 关联erp供应商
     */
    public function erpSupplier()
    {
        return $this->belongsTo(ErpSupplier::class, 'supplier_uuid', 'uuid');
    }

    /**
     * 关联分类
     */
    public function category()
    {
        return $this->belongsTo(Category::class, 'category_uuid', 'uuid');
    }

    /**
     * 关联月度库存记录
     */
    public function erpMonthlyMaterialStatistics()
    {
        return $this->hasMany(ErpMonthlyMaterialStatistics::class, 'material_uuid', 'uuid');
    }

    /**
     * 关联入库记录
     */
    public function erpInventoryRecord()
    {
        return $this->hasMany(ErpInventoryRecord::class, 'material_uuid', 'uuid');
    }

    /**
     * 关联材料单位
     */
    public function materialUnit()
    {
        return $this->belongsTo(MaterialUnit::class, 'unit_uuid', 'uuid')->where('is_default', '=', 1);
    }

    /**
     * 关联仓库物品库存
     */
    public function warehouseItems()
    {
        return $this->hasMany(ErpWarehouseItem::class, 'material_uuid', 'uuid');
    }

    /**
     * 详情
     */
    public static function detail($id, $enableErp = false)
    {
        $with = [
            'image',
            'MultiLanguageName',
            'relatedMaterial',
            'unit',
            'warehouseItems' => function ($query) {
                $query->with('warehouse');
            },
        ];
        if ($enableErp) {
            $with = [
                'image',
                'MultiLanguageName',
                'relatedMaterial',
                'materialUnit' => function ($query) {
                    $query->with('unit');
                },
            ];
        }
        $material = (new static())->with($with)->where('uuid', '=', $id)->find();
        if ($material) {
            // 材料库存
            $materialStock = 0;
            foreach ($material->warehouseItems as $warehouseItem) {
                if ($warehouseItem->warehouse && $warehouseItem->warehouse['type'] == 'normal' && $warehouseItem->warehouse['is_default'] == 1) {
                    $materialStock = $warehouseItem->stock;
                    break;
                }
            }
            $material->stock_num = $materialStock;
            // 材料图片
            $image = $material->image ? [ $material->image ] : [];
            $material->image = $image;
            // 材料单位
            if (!$enableErp) {
                $material->product_unit = $material->unit?->name;
            } else {
                $material->product_unit = $material->materialUnit?->unit?->name;
                $material->product_unit_text = extractLanguage($material->product_unit);
                $material->unit_uuid = $material->materialUnit?->unit?->uuid;
                $material->unit_id = $material->materialUnit?->unit?->uuid;
            }
            // 材料规格
            $material->product_sku = [
                [
                    'purchase_price' => $material->price,
                    'barcode' => $material->barcode_value,
                    'material_stock' => $material->stock_num,
                    'stock_num' => 0,
                    'material' => [],
                ]
            ];
            $material->sku = $material->product_sku;
            $material->productTaxes = [];
        }
        return $material;
    }

    /**
     * 添加原料
     * @return bool
     */
    public function add($data)
    {
        if (!isset($data['type']) || !in_array($data['type'], [Product::TYPE_PRODUCT, Product::TYPE_MATERIAL])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($product_name)) {
            $this->error = '商品名称不能为空';
            return false;
        }
        //
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, 150);
        if ($status === true) {
            $this->error = '商品名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }
        // 商品名称唯一性
        if (CheckService::checkNameExist('product', $product_name, 0)) {
            $this->error = '商品名称已存在';
            return false;
        }
        // 判断图片id是否存在
        $images = isset($data['image']) ? $data['image'] : [];
        $imageIds = array_map(function ($image) {
            return isset($image['file_id']) ? $image['file_id'] : $image['image_id'];
        }, $images);
        $existingImageIds = UploadFile::whereIn('uuid', $imageIds)->column('uuid');
        $missingImageIds = array_diff($imageIds, $existingImageIds);
        if ($missingImageIds) {
            $this->error = '商品图片不存在';
            return false;
        }
        // 条码格式验证，12或13位数字
        $barcode = $data['sku'][0]['barcode'] ?? '';
        if ($barcode && !preg_match('/^[0-9]{1,13}$/', $barcode)) {
            $this->error = '输入条形码不合规，请重新检查';
            return false;
        }
        if ($barcode && CheckService::checkNameExist('product_bom_barcode', $barcode, 0)) {
            $this->error = '商品条码已存在';
            return false;
        }
        $data = $this->sanitizeProductData($data);
        //
        return Db::transaction(function () use($data, $product_name, $imageIds) {
            $data['name'] = $product_name;
            $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($product_name);
            $data['category_uuid'] = $data['category_id'] ?? 0;
            $data['supplier_uuid'] = $data['erp_supplier_id'] ?? 0;
            $data['image_uuid'] = $imageIds[0] ?? 0;
            $data['image_name'] = $data['img_name'] ?? 0;
            $data['unit_uuid'] = $data['unit_id'] ?? 0;
            $data['price'] = $data['sku'][0]['purchase_price'] ?? 0;;
            $data['stock_num'] = $data['sku'][0]['material_stock'] ?? 0; // 库存数量
            $data['barcode_value'] = $data['sku'][0]['barcode'] ?? ''; // 条形码值
            $data['status'] = $data['product_status'] == 10 ? 1 : 0; // 状态, 1-上架 0-下架

            // 保存材料
            if (!$this->save($data)) {
                return false;
            }

            // 查询默认仓库
            $defaultWarehouse = ErpWarehouse::where('type', 'normal')->where('is_default', 1)->find();
            if ($defaultWarehouse) {
                $warehouseItem = new ErpWarehouseItem();
                $warehouseItem->save([
                    'warehouse_uuid' => $defaultWarehouse['uuid'],
                    'material_uuid' => $this['uuid'],
                    'material_code' => $this['code'],
                    'stock' => $data['sku'][0]['material_stock'] ?? 0,
                ]);
            }

            $hasInventoryAuth = (new Product())->hasInventoryAuth();
            if ($hasInventoryAuth) {
                // 创建"添加入库"记录
                self::addWarehouseInForm($this, 1, $data['shop_user_id'], $this['stock_num']);
                
                // erp商品月初库存记录
                ErpMonthlyMaterialStatistics::newMaterialRecord($this['uuid']);
            }

            return true;
        });
    }

    /**
     * 更新原料
     * @return bool
     */
    public function edit($data, $enableErp = false)
    {
        if (!isset($data['type']) || !in_array($data['type'], [Product::TYPE_PRODUCT, Product::TYPE_MATERIAL])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($product_name)) {
            $this->error = '商品名称不能为空';
            return false;
        }
        //
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, 150);
        if ($status === true) {
            $this->error = '商品名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }
        // 商品名称唯一性
        if (CheckService::checkNameExist('product', $product_name, 0, $this['product_id'] ?? 0)) {
            $this->error = '商品名称已存在';
            return false;
        }
        // 判断图片id是否存在
        $images = isset($data['image']) ? $data['image'] : [];
        $imageIds = array_map(function ($image) {
            return isset($image['file_id']) ? $image['file_id'] : $image['image_id'];
        }, $images);
        $existingImageIds = UploadFile::whereIn('uuid', $imageIds)->column('uuid');
        $missingImageIds = array_diff($imageIds, $existingImageIds);
        if ($missingImageIds) {
            $this->error = '商品图片不存在';
            return false;
        }
        // 条码格式验证，12或13位数字
        $barcode = $data['sku'][0]['barcode'] ?? '';
        if ($barcode && !preg_match('/^[0-9]{1,13}$/', $barcode)) {
            $this->error = '输入条形码不合规，请重新检查';
            return false;
        }
        if ($barcode && CheckService::checkNameExist('product_bom_barcode', $barcode, 0, $this['product_id'] ?? 0)) {
            $this->error = '商品条码已存在';
            return false;
        }
        $data = $this->sanitizeProductData($data);
        //

        return Db::transaction(function () use ($data, $product_name, $imageIds, $enableErp) {
            if (!$enableErp) {
                $data['name'] = $product_name;
                $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($product_name, $this['multi_language_name_uuid']);
                $data['category_uuid'] = $data['category_id'] ?? 0;
                $data['supplier_uuid'] = $data['erp_supplier_id'] ?? 0;
                $data['image_uuid'] = $imageIds[0] ?? 0;
                $data['image_name'] = $data['img_name'] ?? 0;
                $data['unit_uuid'] = $data['unit_id'] ?? 0;
                $data['price'] = $data['sku'][0]['purchase_price'] ?? 0;
                $data['stock_num'] = $data['sku'][0]['material_stock'] ?? 0; // 库存数量
                $data['barcode_value'] = $data['sku'][0]['barcode'] ?? ''; // 条形码值
                $data['status'] = $data['product_status'] == 10 ? 1 : 0; // 状态, 1-上架 0-下架
                $oldStockNum = floatval($this->stock_num); // 旧库存
                $newStockNum = floatval($data['stock_num']); // 新库存
                if(!$this->save($data)) {
                    return false;
                }
            } else {
                $data = [
                    'shop_user_id' => $data['shop_user_id'],
                    'price' => $data['sku'][0]['purchase_price'] ?? 0,
                    'stock_num' => $data['sku'][0]['material_stock'] ?? 0, // 库存数量
                    'stock_remark' => $data['stock_remark'] ?? '',
                ];
                $oldStockNum = floatval($this->stock_num); // 旧库存
                $newStockNum = floatval($data['stock_num']); // 新库存
                if(!self::update($data, ['id' => $this['id']])) {
                    return false;
                }
            }

            // 更新仓库物品库存
            $defaultWarehouse = ErpWarehouse::where('type', 'normal')->where('is_default', 1)->find();
            if ($defaultWarehouse) {
                $warehouseItem = ErpWarehouseItem::where('material_uuid', '=', $this['uuid'])->where('warehouse_uuid', '=', $defaultWarehouse['uuid'])->find();
                if (!$warehouseItem) {
                    $warehouseItem = new ErpWarehouseItem();
                    $warehouseItem->save([
                        'warehouse_uuid' => $defaultWarehouse['uuid'],
                        'material_uuid' => $this['uuid'],
                        'material_code' => $this['code'],
                        'stock' => $data['sku'][0]['material_stock'] ?? 0,
                    ]);
                } else {
                    $warehouseItem->save([
                        'stock' => $data['sku'][0]['material_stock'] ?? 0,
                    ]);
                }
                
            }

            $product = new Product();
            if ($product->hasInventoryAuth()) {
                $relatedMaterialUuidList = [];
                foreach ($this->relatedMaterial as $relatedMaterial) {
                    $relatedMaterialUuidList[] = $relatedMaterial->uuid;
                }

                // 更新规格/加料关联材料库存
                RelatedMaterial::updateStock($relatedMaterialUuidList);

                // 出/入库记录
                $diffStockNum = abs(floatval(helper::bcsub($newStockNum, $oldStockNum, 4)));
                // 创建"调整入库"记录
                if ($newStockNum > $oldStockNum) {
                    self::addWarehouseInForm($this, 2, $data['shop_user_id'], $diffStockNum, $data['stock_remark']);
                }
                // 创建"调整出库"记录
                if ($newStockNum < $oldStockNum) {
                    self::addWarehouseOutForm($this, 1, $data['shop_user_id'], $diffStockNum, $data['stock_remark']);
                }
            }

            return true;
        });
    }

    /**
     * 处理数据为负数时，自动转换为0
     */
    private function sanitizeProductData($data)
    {
        $keys = [
            'price',
            'purchase_price',
            'material_num',
            'stock_num',
        ];

        foreach ($keys as $key) {
            if (array_key_exists($key, $data)) {
                $data[$key] = max(0, $data[$key]);
            }
        }
        return $data;
    }

    /**
     * 添加材料入库记录
     * 
     * @param $material 材料
     * @param $scene 入库场景: 0-purchase采购入库 1-add添加入库 2-adjust调整入库
     * @param $operatorUuid 入库操作人uuid
     * @param $num 入库数量
     * @param $remark 入库备注
     */
    public static function addWarehouseInForm($material, $scene, $operatorUuid, $num, $remark = '', $purchaseOrderUuid = 0)
    {
        $formModel = new ErpWarehouseForm();
        return $formModel->save([
            'form_no' => $formModel->generateInCode(),
            'scene' => $scene,
            'num' => $num,
            'material_uuid' => $material['uuid'],
            'operator_uuid' => $operatorUuid,
            'remark' => $remark,
            'purchase_order_uuid' => $purchaseOrderUuid,
        ]);
    }

    /**
     * 添加材料出库记录
     * 
     * @param $material 材料
     * @param $scene 出库场景: 0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库
     * @param $operatorUuid 出库操作人uuid
     * @param $num 出库数量
     * @param $remark 出库备注
     */
    public static function addWarehouseOutForm($material, $scene, $operatorUuid, $num, $remark = '')
    {
        $outFormModel = new ErpWarehouseOutForm();
        return $outFormModel->addOutForm($scene, $operatorUuid, [
            'material_uuid' => $material['uuid'],
            'num' => $num,
            'remark' => $remark
        ]);
    }

    /**
     * 获取材料总库存数
     */
    public static function getTotalStockNum()
    {
        return self::sum('stock_num') ?: 0;
    }
}
