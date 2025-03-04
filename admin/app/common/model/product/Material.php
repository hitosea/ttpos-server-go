<?php

namespace app\common\model\product;

use help\ValidateHelp;
use app\common\model\BaseModel;
use app\common\model\erp\ErpWarehouseForm;
use app\shop\service\CheckService;
use app\common\model\file\UploadFile;
use app\common\model\product\RelatedMaterial as ProductRelatedMaterial;
use app\common\model\store\MultiLanguageName;
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
        return $this->hasMany(ProductRelatedMaterial::class, 'related_uuid', 'uuid');
    }

    /**
     * 详情
     */
    public static function detail($id)
    {
        $material = (new static())->with([
            'image',
            'unit',
            'MultiLanguageName',
            'relatedMaterial',
        ])->where('uuid', '=', $id)->find();
        if ($material) {
            // 材料图片
            $image = $material->image ? [ $material->image ] : [];
            $material->image = $image;
            // 材料单位
            $material->product_unit = $material->unit->name;
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
            $data['barcode_value'] = $data['sku'][0]['barcode'] ?? 0; // 条形码值
            $data['status'] = $data['product_status'] == 10 ? 1 : 0; // 状态, 1-上架 0-下架
            if (!$this->save($data)) {
                return false;
            }

            // 添加入库记录
            if ((new Product())->hasInventoryAuth()) {
                $formModel = new ErpWarehouseForm();
                return $formModel->save([
                    'form_no' => $formModel->generateInCode(),
                    'scene' => 1,
                    'num' => $this->stock_num,
                    'material_uuid' => $this->uuid,
                    'operator_uuid' => $data['shop_user_id'],
                ]);
            }

            return true;
        });
    }

    /**
     * 更新原料
     * @return bool
     */
    public function edit($data)
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
        $data = $this->sanitizeProductData($data);
        //
        $data['name'] = $product_name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($product_name, $this['multi_language_name_uuid']);
        $data['category_uuid'] = $data['category_id'] ?? 0;
        $data['supplier_uuid'] = $data['erp_supplier_id'] ?? 0;
        $data['image_uuid'] = $imageIds[0] ?? 0;
        $data['image_name'] = $data['img_name'] ?? 0;
        $data['unit_uuid'] = $data['unit_id'] ?? 0;
        $data['price'] = $data['sku'][0]['purchase_price'] ?? 0;;
        $data['stock_num'] = $data['sku'][0]['material_stock'] ?? 0; // 库存数量
        $data['barcode_value'] = $data['sku'][0]['barcode'] ?? 0; // 条形码值
        $data['status'] = $data['product_status'] == 10 ? 1 : 0; // 状态, 1-上架 0-下架

        $product = new Product();
        if ($product->hasInventoryAuth()) {
            $relatedMaterialUuidList = [];
            foreach ($this->relatedMaterial as $relatedMaterial) {
                $relatedMaterialUuidList[] = $relatedMaterial->uuid;
            }
            RelatedMaterial::updateStock($relatedMaterialUuidList);
        }

        return $this->save($data);
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
}
