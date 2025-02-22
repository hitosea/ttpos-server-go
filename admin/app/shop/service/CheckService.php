<?php

namespace app\shop\service;

/**
 * 验证服务
 */
class CheckService
{
    /**
     * 检查名称唯一性
     * @param string $source 数据源类型
     * @param array $names 名称数组
     * @param int $shop_supplier_id 店铺供应商ID
     * @param int $id 唯一标识
     * @param string $lang 语言
     * @return array 检查结果数组
     */
    public static function checkNameExist(string $source, array|string $names, int $shop_supplier_id, int $id = null, int $parent_id = 0): bool|array
    {
        $result = [];
        $names = is_string($names) ? [$names] : $names;
        foreach ($names as $lang => $name) {
            $unique = true;
            switch ($source) {
                case 'product_barcode':
                    $unique = (new \app\common\model\product\ProductSku)->checkProductBarcodeExist($name, $id);
                    break;
                case 'product_bom_barcode':
                    $unique = (new \app\common\model\product\ProductBom)->checkProductBarcodeExist($name, $id);
                    break;
                case 'product_img':
                    $unique = (new \app\common\model\product\Product)->checkProductImgExist($name, $shop_supplier_id, $id);
                    break;
                case 'product':
                    $unique = (new \app\common\model\product\Product)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'category':
                    $unique = (new \app\common\model\product\Category)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'sku':
                    $unique = (new \app\common\model\product\Spec)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'attribute':
                    $unique = (new \app\common\model\product\Attribute)->checkNameExist($name, $shop_supplier_id, $id, $lang, $parent_id);
                    break;
                case 'feed':
                    $unique = (new \app\common\model\product\Feed)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'unit':
                    $unique = (new \app\common\model\product\Unit)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'label':
                    $unique = (new \app\common\model\product\Label)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'buffet':
                    $unique = (new \app\common\model\buffet\Buffet)->checkNameExist($name, $shop_supplier_id, $id, $lang);
                    break;
                case 'table':
                    $unique = (new \app\common\model\store\Table)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'table_area':
                    $unique = (new \app\common\model\store\TableArea)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'table_type':
                    $unique = (new \app\common\model\store\TableType)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'printer':
                    $unique = (new \app\common\model\settings\Printer)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'supplier_printing':
                    $unique = (new \app\common\model\supplier\Printing)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'pay_type':
                    $unique = (new \app\common\model\store\PayType)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
                case 'order_scheme':
                    $unique = (new \app\common\model\order\OrderScheme)->checkNameExist($name, $shop_supplier_id, $id);
                    break;
            }
            $result[$lang] = $unique;
        }
        //
        if (isset($result[0])) {
            return $result[0];
        }
        return $result;
    }
}
