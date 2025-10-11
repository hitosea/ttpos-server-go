<?php

namespace app\common\model\supplier;

use app\common\model\BaseModel;
use app\common\model\product\Product;
use think\model\concern\SoftDelete;

class PrintingProduct extends BaseModel
{
    use SoftDelete;

    protected $name = 'product_printer_product_item';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;

    /**
     * 关联打印机（绑定字段到父级）
     */
    public function printerBindNameAndStatus()
    {
        return $this->belongsTo(Printing::class, 'product_printer_uuid', 'uuid')
            ->bind([
                'name',      // 打印机名称
                'status',  // 打印机状态
            ]);
    }

    /**
     * 关联商品
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_package_uuid', 'uuid');
    }

    // 创建商品包关联打印机（先删除旧关联，再批量插入新关联）
    public function CreateProductPackagePrinter($productPackageUuid, $productPrinterUuids) : bool {
        self::where('product_package_uuid', $productPackageUuid)->delete();
        if (empty($productPrinterUuids)) {
            return false;
        }
        foreach ($productPrinterUuids as $productPrinterUuid) {
            self::create([
                'product_package_uuid' => $productPackageUuid,
                'product_printer_uuid' => $productPrinterUuid,
            ]);
        }
        return true;
    }

}