<?php

namespace app\shop\model\product;

use help\StringHelp;
use help\ValidateHelp;
use app\shop\service\CheckService;
use app\common\model\file\UploadFile;
use app\shop\model\product\ProductBom;
use app\common\model\supplier\Printing;
use app\common\model\store\MultiLanguageName;
use app\common\model\supplier\PrintingProduct;
use app\common\model\product\Product as ProductModel;
use app\common\model\Product\Material as MaterialModel;
use app\common\model\product\ProductPackageGroup as ProductPackageGroupModel;
use app\common\model\product\ProductPackageGroupItem as ProductPackageGroupItemModel;

/**
 * 商品模型
 */
class Product extends ProductModel
{
    private const SELLING_POINT_LANGUAGES = ['zh', 'en', 'zhtw', 'th', 'my', 'ja', 'ko', 'tr', 'sv'];

    /**
     * 添加商品
     */
    public function add($data)
    {
        if (!isset($data['type']) || !in_array($data['type'], [ProductModel::TYPE_PRODUCT, ProductModel::TYPE_MATERIAL, ProductModel::TYPE_PACKAGE])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        // 是否套餐
        $isPackage = $data['type'] == ProductModel::TYPE_PACKAGE;
        // 
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($product_name)) {
            $this->error = !$isPackage ? '商品名称不能为空' : '套餐名称不能为空';
            return false;
        }
        //
        $maxLength = 150;
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, $maxLength);
        if ($status === true) {
            $this->error = !$isPackage ? '商品名称长度不能超过150个字符' : '套餐名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }
        $data['content'] = isset($data['content']) ? $data['content'] : '';
        $data['alone_grade_equity'] = isset($data['alone_grade_equity']) ? json_decode($data['alone_grade_equity'], true) : '';
        $data['app_id'] = self::$app_id;
        // 规格
        if (isset($data['sku']) && is_array($data['sku']) && !empty($data['sku'])) {
            // 初始化条码错误数组和条码去重检查
            $firstError = [];
            $barcodes = array_column($data['sku'], 'barcode');
            $barcodeCounts = array_count_values(array_filter($barcodes));
            $duplicateBarcodes = array_filter($barcodeCounts, fn($count) => $count > 1);
            foreach ($data['sku'] as &$info) {
                // 如果同一个条码出现超过1次，标记为错误
                $errorMsg1 = '商品条码不能重复';
                $barcodeError1 = true;
                if (!isset($firstError[$errorMsg1])) {
                    $firstError[$errorMsg1] = [];
                }
                if ($info['barcode'] && isset($duplicateBarcodes[$info['barcode']])) {
                    $barcodeError1 = false;
                }
                $firstError[$errorMsg1][] = $barcodeError1;

                // 条码格式验证，12或13位数字
                $errorMsg2 = '输入条形码不合规，请重新检查';
                $barcodeError2 = true;
                if (!isset($firstError[$errorMsg2])) {
                    $firstError[$errorMsg2] = [];
                }
                if ($info['barcode'] && !preg_match('/^[0-9]{1,13}$/', $info['barcode'])) {
                    $barcodeError2 = false;
                }
                $firstError[$errorMsg2][] = $barcodeError2;

                // 条码唯一性验证
                $errorMsg3 = '商品条码已存在';
                $barcodeError3 = true;
                if (!isset($firstError[$errorMsg3])) {
                    $firstError[$errorMsg3] = [];
                }
                if ($info['barcode'] && CheckService::checkNameExist('product_bom_barcode', $info['barcode'], 0)) {
                    $barcodeError3 = false;
                }
                $firstError[$errorMsg3][] = $barcodeError3;

                // 处理数据超过最大值时，返回提示信息
                if ($text = $this->alertProductData($info)) {
                    $this->error = $text;
                    return false;
                }

                // 处理数据格式
                $info = $this->sanitizeProductData($info);
            }
            unset($info);
            // 处理错误
            $firstError = array_filter($firstError, fn($item) => in_array(false, $item, true));
            if (!empty($firstError)) {
                $this->error = key($firstError);
                $this->errorData = reset($firstError);
                return false;
            }
        }
        // 属性
        if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
            // 验证属性组数量限制
            if (count($data['product_attr']) > 100) {
                $this->error = '属性组数量不能超过100';
                return false;
            }
            
            foreach ($data['product_attr'] as $attr) {
                // 验证最小最大可选范围
                $minSelect = $attr['attribute_min_select'] ?? 0;
                $maxSelect = $attr['attribute_max_select'] ?? 0;
                
                // 验证最大可选 >= 最小可选
                if ($maxSelect > 0 && $maxSelect < $minSelect) {
                    $this->error = '最大可选不可小于最小可选';
                    return false;
                }
                
                // 验证默认勾选数量
                if (isset($attr['default_select']) && is_array($attr['default_select'])) {
                    $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
                        return $item == 1;
                    }));
                    
                    // 默认勾选数量必须在可选范围内
                    if ($defaultSelectCount < $minSelect) {
                        $this->error = '默认勾选数量不能少于最小可选数量';
                        return false;
                    }
                    if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
                        $this->error = '默认勾选数量不能超过最大可选数量';
                        return false;
                    }
                }
                
                // 验证最大可选不能超过属性值数量
                if ($maxSelect > 0 && isset($attr['attribute_value'])) {
                    $attrValueCount = count($attr['attribute_value']);
                    if ($maxSelect > $attrValueCount) {
                        $this->error = '最大可选数量不能超过属性值数量';
                        return false;
                    }
                }
            }
        }
        // 套餐商品组
        if ($isPackage) {
            // 套餐价格
            $packagePrice = $data['package_price'] ?: 0;
            if ($packagePrice <= 0 || $packagePrice > 1000000) {
                $this->error = '套餐价格不能为0或超过1000000';
                return false;
            }
            $packageGroup = $data['package_group'] ?? [];
            if (empty($packageGroup)) {
                $this->error = '套餐分组不能为空';
                return false;
            }
            // 调整限制数量从5改为100
            if (count($packageGroup) > 100) {
                $this->error = '套餐分组最多只能设置100个';
                return false;
            }
            foreach ($packageGroup as $groupIndex => &$item) {
                // 分组名称
                $groupName = $item['group_name'] ?? '';
                [$status, $msg] = ValidateHelp::hasExceedLength($groupName, 150);
                if ($status === true) {
                    $this->error = '分组名称长度不能超过150个字符';
                    $this->errorData = $msg;
                    return false;
                }
                
                // 验证分组可选范围
                $minSelect = $item['group_min_select'] ?? 1;
                $maxSelect = $item['group_max_select'] ?? 1;
                
                // 验证最大可选 >= 最小可选（允许为0）
                if ($maxSelect < $minSelect) {
                    $this->error = sprintf('分组%d最大可选不可小于最小可选', $groupIndex + 1);
                    return false;
                }
                
                // 分组商品
                $groupProductList = $item['product_list'] ?? [];
                if (count($groupProductList) <= 0) {
                    $this->error = '商品不能为空';
                    return false;
                }
                
                // 验证最大可选不能超过分组商品数量
                $groupProductCount = count($groupProductList);
                if ($maxSelect > $groupProductCount) {
                    $this->error = sprintf('分组%d最大可选数量不能超过商品数量', $groupIndex + 1);
                    return false;
                }
                
                $productIds = array_column($groupProductList, 'product_id');
                $productBoms = ProductBom::whereIn('uuid', $productIds)->select();
                foreach ($productBoms as $productBom) {
                    $groupProducts = array_filter($groupProductList, function($product) use ($productBom) {
                        return $product['product_id'] == $productBom->uuid;
                    });
                    $groupProduct = reset($groupProducts); // 取第一个匹配的元素
                    $groupProductNum = $groupProduct['num'] ?: 0;
                    if ($groupProductNum <= 0) {
                        $this->error = '商品数量不能为0';
                        return false;
                    }
                }
            }
            $data['product_type'] = 1; // 商品类型 0-商品 1-套餐
            $data['price'] = $packagePrice; // 套餐价格
            $data['is_show_delivery'] = 2; // 默认不显示外送 1-显示 2-隐藏
            $data['sku'] = [
                [
                    'product_price' => $packagePrice,
                    'spec_name' => $product_name,
                    'stock_num' => $data['package_stock'] ?: 0,
                    'spec_id' => 0,
                    'barcode' => '',
                ]
            ];
        }
        $data = $this->sanitizeProductData($data);
        // 加料
        if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
            // 调整限制数量从10改为100
            if (count($data['product_feed']) > 100) {
                $this->error = '最多可添加100个加料';
                return false;
            }
            
            // 验证最小最大可选范围
            $minSelect = $data['feed_min_select'] ?? 0;
            $maxSelect = $data['feed_max_select'] ?? 0;
            
            // 验证最大可选 >= 最小可选
            if ($maxSelect > 0 && $maxSelect < $minSelect) {
                $this->error = '最大可选不可小于最小可选';
                return false;
            }
            
            // 验证默认勾选数量
            $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
                return isset($item['is_default']) && $item['is_default'] == 1;
            }));
            
            // 默认勾选数量必须在可选范围内
            if ($defaultSelectCount < $minSelect) {
                $this->error = '默认勾选数量不能少于最小可选数量';
                return false;
            }
            if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
                $this->error = '默认勾选数量不能超过最大可选数量';
                return false;
            }
            
            // 验证最大可选不能超过加料数量
            if ($maxSelect > 0 && $maxSelect > count($data['product_feed'])) {
                $this->error = '最大可选数量不能超过加料数量';
                return false;
            }
            
            foreach ($data['product_feed'] as &$item) {
                if (!isset($item['uuid']) || empty($item['uuid'])) {
                    $item['uuid'] = StringHelp::getGuidV4();
                }
                $item = $this->sanitizeProductData($item);
            }
            unset($item);
        }
        //
        if (isset($data['productTaxes']) && is_array($data['productTaxes']) && count($data['productTaxes']) > 2) {
            $this->error = '产品税类只能设置2条';
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
        //
        $dineTaxUuid = 0; // 堂食税类id
        $takeoutTaxUuid = 0; // 外带税类id
        if (!empty($data['productTaxes'])) {
            foreach ($data['productTaxes'] as $tax) {
                if ($tax['product_tax_type'] == 1) {
                    $dineTaxUuid = $tax['tax_category_id'] ?? 0;
                } else if ($tax['product_tax_type'] == 2) {
                    $takeoutTaxUuid = $tax['tax_category_id'] ?? 0;
                }
            }
        }
        $data['name'] = $product_name;
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($product_name);
        $data['category_uuid'] = $data['category_id'] ?? 0; // 分类UUID
        $data['special_category_uuid'] = $data['special_id'] ?? 0; // 特殊类别UUID
        $data['supplier_uuid'] = $data['erp_supplier_id'] ?? 0; // 供应商UUID
        $data['image_file_uuid'] = $imageIds[0] ?? 0; // 图片文件UUID
        $data['image_name'] = $data['img_name'] ?? 0; // 图片名称
        $data['unit_uuid'] = $data['unit_id'] ?? 0; // 单位UUID
        $data['dine_tax_uuid'] = $dineTaxUuid; // 堂食税类UUID
        $data['takeout_tax_uuid'] = $takeoutTaxUuid; // 外带税类UUID
        $data['printer_tag_uuid'] = $data['label_id'] ?? 0; // 打印机标签UUID
        $data['deduct_stock_type'] = $data['deduct_stock_type'] == 10 ? 1 : 0; // 库存计算方法, 0-付款减库存 1-下单减库存 （deduct_stock_type 库存计算方式(10下单减库存 20付款减库存)）
        $data['num_type'] = $data['num_type'] ?? 0; // 数量计算方法, 0-整数 1-小数
        $data['sauce_min_selection'] = $data['feed_min_select'] ?? 0; // 小料最小选择数量
        $data['sauce_max_selection'] = $data['feed_max_select'] ?? 0; // 小料最大选择数量
        // sauce_required 保留但不再使用，兼容旧数据
        $data['sauce_required'] = ($data['feed_min_select'] ?? 0) > 0 ? 1 : 0; // 是否必选小料, 0-否 1-是
        if (!$this->applySellingPointLocales($data)) {
            return false;
        }
        $data['open_discount'] = $data['is_enable_grade'] ?? 0; // 是否开启会员折扣, 0-否 1-是
        $data['price'] = $isPackage ? $data['price'] : ($data['sku'][0]['purchase_price'] ?? 0); // 价格
        $data['stock_num'] = $data['sku'][0]['material_stock'] ?? 0; // 库存数量
        $data['barcode_value'] = $data['sku'][0]['barcode'] ?? ''; // 条形码值
        $data['status'] = $data['product_status'] == 10 ? 1 : 0; // 状态, 1-上架 0-下架
        $data['is_show_cashier'] = $data['is_show_cashier'] != 2 ? 1 : 0;
        $data['is_show_tablet'] = $data['is_show_tablet'] != 2 ? 1 : 0;
        $data['is_show_kitchen'] = $data['is_show_kitchen'] != 2 ? 1 : 0;
        $data['is_show_assistant'] = $data['is_show_assistant'] != 2 ? 1 : 0;
        $data['is_show_h5'] = $data['is_show_h5'] != 2 ? 1 : 0;
        $data['is_show_delivery'] = $data['is_show_delivery'] != 2 ? 1 : 0;
        $data['sort'] = $data['product_sort'] ?? 0;
        $data['open_overall_discount'] = $data['open_overall_discount'] ?? 1; // 是否开启整单折扣 0-否 1-是

        // 开启事务
        $this->startTrans();
        try {
            // 添加商品
            $this->save($data);
            // 商品图片
            $this->addProductImages($data['image']);
            // 商品规格
            ProductBom::addFlavor($data, $this);
            // 商品加料
            ProductBom::addFeed($data, $this);
            // 商品属性
            ProductAttribute::addAttribute($data, $this);
            // 套餐商品组
            if ($isPackage) {
                ProductPackageGroupModel::addPackageGroup($data, $this);
            } else if (isset($data['product_printer_uuids'])) {
                // 验证商品打印机
                if (empty($data['printer_tag_uuid'])) {
                    $isExistPrintProductSelect = Printing::where('uuid', 'in', $data['product_printer_uuids'])->where('print_product_select', 1)->count();
                    if ($isExistPrintProductSelect) {
                        $this->error = '未设置打印标签';
                        return false;
                    }
                }
                // 新增商品包关联打印机
                PrintingProduct::createProductPackagePrinter($this['product_id'], $data['product_printer_uuids']);
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 处理数据超过最大值时，返回提示信息
     */
    private function alertProductData($data)
    {
        $limits = [
            'price' => ['limit' => 100000000, 'message' => '价格不能超过100000000'],
            'product_price' => ['limit' => 100000000, 'message' => '价格不能超过100000000'],
            'stock_num' => ['limit' => 99999999, 'message' => '库存不能超过99999999']
        ];

        foreach ($limits as $key => $value) {
            if (array_key_exists($key, $data) && $data[$key] > $value['limit']) {
                return $value['message'];
            }
        }
        return '';
    }

    /**
     * 处理数据为负数时，自动转换为0
     */
    private function sanitizeProductData($data)
    {
        $keys = [
            'price',
            'product_price',
            'sales_initial',
            'product_sort',
            'line_price',
            'supplier_price',
            'bag_price',
            'cost_price',
            'min_buy',
            'limit_num',
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

    private function applySellingPointLocales(array &$data, ?int $currentUuid = null): bool
    {
        $locales = $this->normalizeSellingPointLocales($data['selling_point_i18n'] ?? []);
        unset($data['selling_point_i18n']);

        $hasContent = false;
        foreach ($locales as $lang => $text) {
            if ($text !== '') {
                $hasContent = true;
            }
            if (mb_strlen($text) > 500) {
                $this->error = sprintf('%s卖点不能超过500个字符', strtoupper($lang));
                return false;
            }
        }
        
        if (!$hasContent) {
            $data['describe'] = $this->resolveSellingPointFallback($locales, $data);
            $data['describe_multi_language_name_uuid'] = $currentUuid ?: 0;
            return true;
        }

        $multiLanguageUuid = (new MultiLanguageName)->saveNames($locales, $currentUuid ?: null);
        $data['describe_multi_language_name_uuid'] = $multiLanguageUuid;
        $data['describe'] = $this->resolveSellingPointFallback($locales, $data);

        return true;
    }

    private function normalizeSellingPointLocales($value): array
    {
        if (empty($value)) {
            return [];
        }

        if (is_string($value)) {
            $decoded = json_decode($value, true);
            if (json_last_error() === JSON_ERROR_NONE) {
                $value = $decoded;
            } else {
                return [];
            }
        }

        if (!is_array($value)) {
            return [];
        }

        $locales = [];
        foreach (self::SELLING_POINT_LANGUAGES as $lang) {
            if (!array_key_exists($lang, $value)) {
                continue;
            }
            $text = $value[$lang];
            if ($text === null) {
                $locales[$lang] = '';
                continue;
            }
            $locales[$lang] = is_string($text) ? trim($text) : trim((string) $text);
        }

        return $locales;
    }

    private function resolveSellingPointFallback(array $locales, array $data): string
    {
        if (!empty($locales['zh'])) {
            return $locales['zh'];
        }
        foreach ($locales as $value) {
            if ($value !== '') {
                return $value;
            }
        }
        if (!empty($data['selling_point'])) {
            $describe = '';
            $sellingPoint = $data['selling_point'];
            if (is_string($sellingPoint)) {
                $decoded = json_decode($sellingPoint, true);
                // 保证解码成功且为数组
                $sellingPoint = (json_last_error() === JSON_ERROR_NONE && is_array($decoded)) ? $decoded : [];
            }
            if (is_array($sellingPoint)) {
                foreach ($sellingPoint as $value) {
                    if (is_string($value)) {
                        $trimmedValue = trim($value);
                        if ($trimmedValue !== '') {
                            $describe = $trimmedValue;
                        }
                    } elseif (!empty($value)) {
                        // 保证非字符串类型也能转为字符串不会报错
                        $descValue = trim((string)$value);
                        if ($descValue !== '') {
                            $describe = $descValue;
                        }
                    }
                }
            }
            return $describe;
        }
        $currentDescribe = $this->describe;
        return is_string($currentDescribe) ? $currentDescribe : '';
    }

    /**
     * 添加商品图片
     */
    private function addProductImages($images)
    {
        foreach ($this->image()->select() as $image) {
            if ($image->uuid == $images[0]['file_id']) {
                continue;
            }
            $image->delete();
        }
    }

    /**
     * 编辑商品
     */
    public function edit($data, $enableErp = false)
    {
        if (!isset($data['type']) || !in_array($data['type'], [ProductModel::TYPE_PRODUCT, ProductModel::TYPE_MATERIAL, ProductModel::TYPE_PACKAGE])) {
            $this->error = '商品类型不能为空';
            return false;
        }
        // 是否套餐
        $isPackage = $data['type'] == ProductModel::TYPE_PACKAGE;
        // 商品名称
        $product_name = isset($data['product_name']) ? $data['product_name'] : '';
        if (ValidateHelp::hasEmptyValue($product_name)) {
            $this->error = !$isPackage ? '商品名称不能为空' : '套餐名称不能为空';
            return false;
        }
        //
        $maxLength = 150;
        [$status, $msg] = ValidateHelp::hasExceedLength($product_name, $maxLength);
        if ($status === true) {
            $this->error = !$isPackage ? '商品名称长度不能超过150个字符' : '套餐名称长度不能超过150个字符';
            $this->errorData = $msg;
            return false;
        }

        $data['spec_type'] = isset($data['spec_type']) ? $data['spec_type'] : $this['spec_type'];
        $data['content'] = isset($data['content']) ? $data['content'] : '';
        $data['alone_grade_equity'] = isset($data['alone_grade_equity']) ? json_decode($data['alone_grade_equity'], true) : '';
        // 规格
        if (isset($data['sku']) && is_array($data['sku']) && !empty($data['sku'])) {
            // 初始化条码错误数组和条码去重检查
            $firstError = [];
            $barcodes = array_column($data['sku'], 'barcode');
            $barcodeCounts = array_count_values(array_filter($barcodes));
            $duplicateBarcodes = array_filter($barcodeCounts, fn($count) => $count > 1);
            foreach ($data['sku'] as &$info) {
                // 如果同一个条码出现超过1次，标记为错误
                $errorMsg1 = '商品条码不能重复';
                $barcodeError1 = true;
                if (!isset($firstError[$errorMsg1])) {
                    $firstError[$errorMsg1] = [];
                }
                if ($info['barcode'] && isset($duplicateBarcodes[$info['barcode']])) {
                    $barcodeError1 = false;
                }
                $firstError[$errorMsg1][] = $barcodeError1;

                // 条码格式验证，12或13位数字
                $errorMsg2 = '输入条形码不合规，请重新检查';
                $barcodeError2 = true;
                if (!isset($firstError[$errorMsg2])) {
                    $firstError[$errorMsg2] = [];
                }
                if ($info['barcode'] && !preg_match('/^[0-9]{1,13}$/', $info['barcode'])) {
                    $barcodeError2 = false;
                }
                $firstError[$errorMsg2][] = $barcodeError2;

                // 条码唯一性验证
                $errorMsg3 = '商品条码已存在';
                $barcodeError3 = true;
                if (!isset($firstError[$errorMsg3])) {
                    $firstError[$errorMsg3] = [];
                }
                if ($info['barcode'] && CheckService::checkNameExist('product_bom_barcode', $info['barcode'], 0, $info['product_id'] ?? 0)) {
                    $barcodeError3 = false;
                }
                $firstError[$errorMsg3][] = $barcodeError3;

                // 处理数据超过最大值时，返回提示信息
                if ($text = $this->alertProductData($info)) {
                    $this->error = $text;
                    return false;
                }

                // 处理数据格式
                $info = $this->sanitizeProductData($info);
            }
            unset($info);
            // 处理错误
            $firstError = array_filter($firstError, fn($item) => in_array(false, $item, true));
            if (!empty($firstError)) {
                $this->error = key($firstError);
                $this->errorData = reset($firstError);
                return false;
            }
        }
        // 属性
        if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
            // 验证属性组数量限制
            if (count($data['product_attr']) > 100) {
                $this->error = '属性组数量不能超过100';
                return false;
            }
            
            foreach ($data['product_attr'] as $attr) {
                // 验证最小最大可选范围
                $minSelect = $attr['attribute_min_select'] ?? 0;
                $maxSelect = $attr['attribute_max_select'] ?? 0;
                
                // 验证最大可选 >= 最小可选
                if ($maxSelect > 0 && $maxSelect < $minSelect) {
                    $this->error = '最大可选不可小于最小可选';
                    return false;
                }
                
                // 验证默认勾选数量
                if (isset($attr['default_select']) && is_array($attr['default_select'])) {
                    $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
                        return $item == 1;
                    }));
                    
                    // 默认勾选数量必须在可选范围内
                    if ($defaultSelectCount < $minSelect) {
                        $this->error = '默认勾选数量不能少于最小可选数量';
                        return false;
                    }
                    if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
                        $this->error = '默认勾选数量不能超过最大可选数量';
                        return false;
                    }
                }
                
                // 验证最大可选不能超过属性值数量
                if ($maxSelect > 0 && isset($attr['attribute_value'])) {
                    $attrValueCount = count($attr['attribute_value']);
                    if ($maxSelect > $attrValueCount) {
                        $this->error = '最大可选数量不能超过属性值数量';
                        return false;
                    }
                }
            }
        }
        // 套餐商品组
        if ($isPackage) {
            // 套餐价格
            $packagePrice = $data['package_price'] ?: 0;
            if ($packagePrice <= 0 || $packagePrice > 100000000) {
                $this->error = '套餐价格不能为0或超过100000000';
                return false;
            }
            $packageGroup = $data['package_group'] ?? [];
            if (empty($packageGroup)) {
                $this->error = '套餐分组不能为空';
                return false;
            }
            // 调整限制数量从5改为100
            if (count($packageGroup) > 100) {
                $this->error = '套餐分组最多只能设置100个';
                return false;
            }
            foreach ($packageGroup as $groupIndex => &$item) {
                // 分组名称
                $groupName = $item['group_name'] ?? '';
                [$status, $msg] = ValidateHelp::hasExceedLength($groupName, 150);
                if ($status === true) {
                    $this->error = '分组名称长度不能超过150个字符';
                    $this->errorData = $msg;
                    return false;
                }
                
                // 验证分组可选范围
                $minSelect = $item['group_min_select'] ?? 1;
                $maxSelect = $item['group_max_select'] ?? 1;
                
                // 验证最大可选 >= 最小可选（允许为0）
                if ($maxSelect < $minSelect) {
                    $this->error = sprintf('分组%d最大可选不可小于最小可选', $groupIndex + 1);
                    return false;
                }
                
                // 分组商品
                $groupProductList = $item['product_list'] ?? [];
                if (count($groupProductList) <= 0) {
                    $this->error = '商品不能为空';
                    return false;
                }
                
                // 验证最大可选不能超过分组商品数量
                $groupProductCount = count($groupProductList);
                if ($maxSelect > $groupProductCount) {
                    $this->error = sprintf('分组%d最大可选数量不能超过商品数量', $groupIndex + 1);
                    return false;
                }
                
                $productIds = array_column($groupProductList, 'product_id');
                $productBoms = ProductBom::whereIn('uuid', $productIds)->select();
                foreach ($productBoms as $productBom) {
                    if ($productBom->status == 0) {
                        $this->error = '商品不能为下架商品';
                        return false;
                    }
                    $groupProducts = array_filter($groupProductList, function($product) use ($productBom) {
                        return $product['product_id'] == $productBom->uuid;
                    });
                    $groupProduct = reset($groupProducts); // 取第一个匹配的元素
                    $groupProductNum = $groupProduct['num'] ?: 0;
                    if ($groupProductNum <= 0) {
                        $this->error = '商品数量不能为0';
                        return false;
                    }
                }
            }
            $data['product_type'] = 1; // 商品类型 0-商品 1-套餐
            $data['price'] = $packagePrice; // 套餐价格
            $data['is_show_delivery'] = 2; // 默认不显示外送 1-显示 2-隐藏
            $data['sku'] = [
                [
                    'product_price' => $packagePrice,
                    'spec_name' => $product_name,
                    'stock_num' => $data['package_stock'] ?: 0,
                    'spec_id' => 0,
                    'barcode' => '',
                    'product_sku_id' => $this['sku'][0]['uuid'],
                ]
            ];
        }
        $data = $this->sanitizeProductData($data);
        // 加料
        if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
            // 调整限制数量从10改为100
            if (count($data['product_feed']) > 100) {
                $this->error = '最多可添加100个加料';
                return false;
            }
            
            // 验证最小最大可选范围
            $minSelect = $data['feed_min_select'] ?? 0;
            $maxSelect = $data['feed_max_select'] ?? 0;
            
            // 验证最大可选 >= 最小可选
            if ($maxSelect > 0 && $maxSelect < $minSelect) {
                $this->error = '最大可选不可小于最小可选';
                return false;
            }
            
            // 验证默认勾选数量
            $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
                return isset($item['is_default']) && $item['is_default'] == 1;
            }));
            
            // 默认勾选数量必须在可选范围内
            if ($defaultSelectCount < $minSelect) {
                $this->error = '默认勾选数量不能少于最小可选数量';
                return false;
            }
            if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
                $this->error = '默认勾选数量不能超过最大可选数量';
                return false;
            }
            
            // 验证最大可选不能超过加料数量
            if ($maxSelect > 0 && $maxSelect > count($data['product_feed'])) {
                $this->error = '最大可选数量不能超过加料数量';
                return false;
            }
            
            foreach ($data['product_feed'] as &$item) {
                if (!isset($item['uuid']) || empty($item['uuid'])) {
                    $item['uuid'] = StringHelp::getGuidV4();
                }
                $item = $this->sanitizeProductData($item);
            }
            unset($item);
        } else {
            $data['feed_min_select'] = 0;
            $data['feed_max_select'] = 0;
            // 兼容旧字段
            $data['feed_required'] = 0;
            $data['feed_open_max_select'] = 0;
        }
        //
        if (isset($data['productTaxes']) && is_array($data['productTaxes']) && count($data['productTaxes']) > 2) {
            $this->error = '产品税类只能设置2条';
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
        //
        return $this->transaction(function () use ($data, $isPackage) {
            $data['product_attr'] = isset($data['product_attr']) ? $data['product_attr'] : '';
            $data['product_feed'] = isset($data['product_feed']) ? $data['product_feed'] : '';
            // 更新产品包
            $this->updateProductPackage($data);
            // 产品图片
            $this->addProductImages($data['image']);
            // 更新产品规格
            ProductBom::updateFlavor($data, $this);
            // 更新产品加料
            ProductBom::updateFeed($data, $this);
            // 更新产品属性
            ProductAttribute::updateAttribute($data, $this);
            // 套餐商品组
            if ($isPackage) {
                ProductPackageGroupModel::updatePackageGroup($data, $this);
            } else {
                // 下架非商品时，同时下架关联的套餐
                if ($data['product_status'] == 20) {
                    $items = ProductPackageGroupItemModel::with([
                        'productPackageGroup' => [
                            'product'
                        ]
                    ])->where('related_uuid', $this['product_id'])->select();
                    foreach ($items as $item) {
                        $package = $item?->productPackageGroup?->product;
                        if ($package) {
                            $package->save(['status' => 0]);
                            ProductBom::where('product_package_uuid', $package->uuid)->update(['status' => 0]);
                        }
                    }
                }
                // 新增商品包关联打印机
                if (isset($data['product_printer_uuids'])) {
                    // 验证商品打印机
                    if ($this->printer_tag_uuid == 0) {
                        $isExistPrintProductSelect = Printing::where('uuid', 'in', $data['product_printer_uuids'])->where('print_product_select', 1)->count();
                        if ($isExistPrintProductSelect) {
                            $this->error = '未设置打印标签';
                            return false;
                        }
                    }
                    // 新增商品包关联打印机
                    PrintingProduct::createProductPackagePrinter($this['product_id'], $data['product_printer_uuids']);
                }
            }
            return true;
        });
    }

    /**
     * 更新产品包
     */
    public function updateProductPackage($data)
    {
        // 处理商品图片
        $fileId = 0;
        if (!empty($data['image'])) {
            $fileId = $data['image'][0]['file_id'];
        }
        // 处理税率
        $dineTaxUuid = 0; // 堂食税类id
        $takeoutTaxUuid = 0; // 外带税类id
        $taxList = $data['productTaxes'] ?? [];
        foreach ($taxList as $item) {
            if ($item['product_tax_type'] == 1) {
                $dineTaxUuid = $item['tax_category_id'];
            } else if ($item['product_tax_type'] == 2) {
                $takeoutTaxUuid = $item['tax_category_id'];
            }
        }

        // 是否显示外送
        $isShowDelivery = isset($data['is_show_delivery']) ? ($data['is_show_delivery'] == 0 ? 2 : $data['is_show_delivery']) : 2;

        $sellingPointPayload = [
            'selling_point' => $data['selling_point'] ?? ($this->getData('describe') ?? ''),
        ];
        if (isset($data['selling_point_i18n'])) {
            $sellingPointPayload['selling_point_i18n'] = $data['selling_point_i18n'];
        }
        if (!$this->applySellingPointLocales($sellingPointPayload, $this->getData('describe_multi_language_name_uuid') ?: 0)) {
            return false;
        }

        $updateData = [
            'name' => $data['product_name'], // 产品包名称
            'image_name' => $data['img_name'] ?? '', // 产品包图片名称
            'image_file_uuid' => $fileId, // 产品包图片文件id
            'deduct_stock_type' => $data['deduct_stock_type'] == 10 ? 1 : 0, // 扣库存类型: 10-下单减库存, 20-付款减库存
            'num_type' => $data['num_type'] ?? 0, // 数量计算方法, 0-整数 1-小数
            'unit_uuid' => $data['unit_id'], // 单位uuid
            'dine_tax_uuid' => $dineTaxUuid, // 堂食税类id
            'takeout_tax_uuid' => $takeoutTaxUuid, // 外带税类id
            'category_uuid' => $data['category_id'], // 分类uuid
            'status' => $data['product_status'] == 10 ? 1 : 0, // 状态: 10-上架, 20-下架
            'is_show_cashier' => $data['is_show_cashier'] != 2 ? 1 : 0, // 是否显示收银台: 10-显示, 20-隐藏
            'is_show_tablet' => $data['is_show_tablet'] != 2 ? 1 : 0, // 是否显示平板: 10-显示, 20-隐藏,
            'is_show_kitchen' => $data['is_show_kitchen'] != 2 ? 1 : 0, // 是否显示厨房: 10-显示, 20-隐藏
            'is_show_assistant' => $data['is_show_assistant'] != 2 ? 1 : 0, // 是否显示助手: 10-显示, 20-隐藏,
            'is_show_h5' => $data['is_show_h5'] != 2 ? 1 : 0, // 是否显示h5: 10-显示, 20-隐藏
            'is_show_delivery' => $isShowDelivery != 2 ? 1 : 0, // 是否显示外送: 1-显示, 0-隐藏
            'sort' => $data['product_sort'], // 排序
            'limit_num' => $data['limit_num'], // 限购数量,
            'sauce_min_selection' => $data['feed_min_select'] ?? 0, // 加料最小选择数量
            'sauce_max_selection' => $data['feed_max_select'] ?? 0, // 加料最大选择数量
            // sauce_required 保留但不再使用，兼容旧数据
            'sauce_required' => ($data['feed_min_select'] ?? 0) > 0 ? 1 : 0, // 是否必选加料: 0-否, 1-是,
            'special_category_uuid' => $data['special_id'], // 热门分类
            'describe' => $sellingPointPayload['describe'], // 卖点
            'describe_multi_language_name_uuid' => $sellingPointPayload['describe_multi_language_name_uuid'] ?? 0,
            'open_discount' => $data['is_enable_grade'], // 是否开启折扣: 0-否, 1-是
            'open_overall_discount' => $data['open_overall_discount'], // 是否开启整单折扣: 0-否, 1-是
            'printer_tag_uuid' => $data['label_id'] ?? 0, // 打印机标签
            'supplier_uuid' => $data['erp_supplier_id'] ?? 0, // 供应商uuid
        ];
        if ($this->product_type == 1) {
            $updateData['price'] = $data['price'];
        }

        // 更新产品包
        $this->save($updateData);
        // 更新产品包多语言
        $multiLanguageName = new MultiLanguageName();
        $multiLanguageName->saveNames($data['product_name'], $this['multi_language_name_uuid']);
    }

    /**
     * 修改商品状态
     */
    public function setStatus($state)
    {
        $this->startTrans();
        try {
            $value = $state == 10 ? 1 : 0;
            // 套餐开启时判断套餐内商品库存是否充足
            if ($this->product_type == 1 && $value == 1) {
                $items = ProductPackageGroupItemModel::with('productBom')->where('related_uuid', $this->uuid)->select();
                foreach ($items as $item) {
                    if ($item->productBom->stock_num <= 0) {
                        $this->error = '套餐商品库存不足';
                        return false;
                    }
                }
            }
            // 更新product_package表status
            $res = $this->save(['status' => $value]);
            if ($res === false) {
                return false;
            }
            // 更新product_bom表status
            $boms = ProductBom::where('product_package_uuid', $this->uuid)->select();
            foreach ($boms as $bom) {
                $bomRes = $bom->save(['status' => $value]);
                if ($bomRes === false) {
                    return false;
                }
            }
            // 更新套餐状态
            if ($this->product_type == 0 && $value == 0) {
                $items = ProductPackageGroupItemModel::with([
                    'productPackageGroup' => [
                        'product'
                    ]
                ])->where('related_uuid', $this->uuid)->select();
                foreach ($items as $item) {
                    $package = $item?->productPackageGroup?->product;
                    if ($package) {
                        $package->save(['status' => 0]);
                        ProductBom::where('product_package_uuid', $package->uuid)->update(['status' => 0]);
                    }
                }
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
        $value = $state == 10 ? 1 : 0;
        return $this->save(['status' => $value]) !== false;
    }

    /**
     * 软删除
     */
    public function setDelete($product_ids, $shop_user_id = 0)
    {
        $product_ids = explode(',', $product_ids);
        if (empty($product_ids)) return false;
        // 开启事务
        $this->startTrans();
        try {
            $multiLanguageName = new MultiLanguageName();
            $hasInventoryAuth = $this->hasInventoryAuth();

            // 删除商品
            $productList = self::with([
                'sku' => [ 'relatedMaterial' ],
                'feed',
                'productAttributeGroup' => [ 'productAttribute' ],
                'buffetProduct',
                'orderSchemeProduct'
            ])
            ->whereIn('uuid', $product_ids)
            ->select();

            foreach ($productList as $product) {
                // 删除规格
                foreach ($product->sku as $sku) {
                    foreach ($sku->relatedMaterial as $relatedMaterial) {
                        $relatedMaterial->delete();
                    }
                    if ($hasInventoryAuth) {
                        // 创建"删除出库"记录
                        ProductBom::addWarehouseOutForm($sku, 4, $shop_user_id, $sku['stock_num']);
                    }
                    // 删除规格
                    $sku->delete();
                }
                // 删除加料
                foreach ($product->feed as $feed) {
                    $feed->delete();
                }
                // 删除产品属性
                foreach ($product->productAttributeGroup as $productAttributeGroup) {
                    foreach ($productAttributeGroup->productAttribute as $productAttribute) {
                        $productAttribute->delete();
                    }   
                    $productAttributeGroup->delete();
                }
                // 删除产品语言
                $multiLanguageName->clearCache($product->multi_language_name_uuid);
                $product->multiLanguageName?->delete();
                // 删除自助餐商品
                foreach ($product->buffetProduct as $buffetProduct) {
                    $buffetProduct->delete();
                }
                // 删除必点方案商品
                foreach ($product->orderSchemeProduct as $orderSchemeProduct) {
                    $orderSchemeProduct->delete();
                }
                // 删除套餐商品组
                ProductPackageGroupModel::deletePackageGroup($product);
                // 删除产品
                if (!$product->delete()) {
                    $this->error = '删除失败';
                    return false;
                }
            }

            // 删除材料
            $materialList = Material::with([
                'multiLanguageName',
                'relatedMaterial'
            ])->whereIn('uuid', $product_ids)->select();
            /** @var MaterialModel $material */
            foreach ($materialList as $material) {
                if ($material->relatedMaterial()->count() > 0) {
                    $this->error = '该材料已被使用，无法删除';
                    return false;
                }
                // 删除材料语言
                $multiLanguageName->clearCache($material->multi_language_name_uuid);
                $material->multiLanguageName->delete();
                // 删除材料关联表
                foreach ($material->relatedMaterial as $relatedMaterial) {
                    $relatedMaterial->delete();
                }
                if ($hasInventoryAuth) {
                    // 创建"删除出库"记录
                    MaterialModel::addWarehouseOutForm($material, 4, $shop_user_id, $material['stock_num']);
                }
                // 删除材料
                if (!$material->delete()) {
                    $this->error = '删除失败';
                    return false;
                }
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 获取商品告急数量总数
     */
    public function getProductStockTotal($shop_supplier_id = 0)
    {
        $query = $this->alias('product')
            ->join('product_sku sku', 'sku.product_id = product.product_id')
            ->where('sku.stock_num', '<', condition: 10)
            ->where('product.type', ProductModel::TYPE_PRODUCT)
            ->where('product.delete_time', '=', 0);

        if ($shop_supplier_id > 0) {
            $query = $query->where('product.shop_supplier_id', $shop_supplier_id);
        }

        return count($query->distinct(true)->column('product.product_id'));
    }

    public function getProductId($search)
    {
        $res = $this->like('product_name', $search)->select()->toArray();
        return array_column($res, 'product_id');
    }

    /**
     * 获取数量
     */
    public function getCount($type, $shop_supplier_id, $product_type = 0)
    {
        $model = $this;
        //已下架
        if ($type == 'lower') {
            $model = $model->where('status', '=', 20);
        }
        return $model->count();
    }

    /**
     * 同步商品到门店
     * @param $data
     */
    public function transmit($data)
    {
        $product_list = $this->where('product_id', 'in', $data['product_id'])->with(['image', 'sku'])->select();
        // 开启事务
        $this->startTrans();
        try {
            foreach ($product_list as $item) {
                foreach ($data['shop_supplier_id'] as $value) {
                    $product = $this->where('product_id', '=', $item['product_id'])->find();
                    if ($product) {
                        $product->update($item);
                        $this->addImages($item['image']);
                        (new ProductSku)->where('product_id', '=', $product['product_id'])->delete();
                        $this->addProductSku($data['sku']);
                    } else {
                        unset($item['product_id']);
                        unset($item['create_time']);
                        unset($item['update_time']);
                        $item['shop_supplier_id'] = $value;
                        $this->save($item);
                        $this->addImages($item['image']);
                        $this->addProductSku($data['sku']);
                    }
                }
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 添加商品图片
     */
    private function addImages($images)
    {
        foreach ($this->image()->select() as $image) {
            $image->delete();
        }
        $data = array_map(function ($images) {
            return [
                'image_id' => $images['image_id'],
            ];
        }, $images);
        return $this->image()->saveAll($data);
    }

    /**
     * 添加商品规格
     */
    private function addProductSku($data, $isUpdate = false, $productSkuIdList = [])
    {
        // 更新模式: 先删除所有规格
        $model = new ProductSku;
        $stock = 0; //总库存
        $materialStock = 0; //总材料库存
        $product_price = 0; //价格
        $cost_price = 0;
        $bag_price = 0;
        // 添加规格数据
        if ($data['spec_type'] == '10') {
            // 单规格
            $this->sku()->save($data['sku']);
            $stock = $data['sku']['stock_num'] ?? 0;
            $materialStock = $data['sku']['material_stock'] ?? 0;
            $product_price = $data['sku']['product_price'] ?? 0;
            $cost_price = $data['sku']['cost_price'] ?? 0;
            $bag_price = $data['sku']['bag_price'] ?? 0;
        } else if ($data['spec_type'] == '20') {
            //更新规格
            (new Spec)->updateSpec($data['sku']);
            // 添加商品sku
            $model->addSkuList($this['product_id'], $data['sku'], $productSkuIdList);
            $product_price = $data['sku'][0]['product_price'] ?? 0;
            $cost_price = $data['sku'][0]['cost_price'] ?? 0;
            $bag_price = $data['sku'][0]['bag_price'] ?? 0;
            foreach ($data['sku'] as $item) {
                $stock += $item['stock_num'] ?? 0;
                $materialStock += $item['material_stock'] ?? 0;
                if ($item['product_price'] < $product_price) {
                    $product_price = $item['product_price'] ?? 0;
                }
                if ($item['cost_price'] < $cost_price) {
                    $cost_price = $item['cost_price'] ?? 0;
                }
                if ($item['bag_price'] < $bag_price) {
                    $bag_price = $item['bag_price'] ?? 0;
                }
            }
        }
        $this->save([
            'product_stock' => $stock,
            'product_material_stock' => $materialStock,
            'product_price' => $product_price,
            'cost_price' => $cost_price,
            'bag_price' => $bag_price
        ]);
    }
}
