<?php

namespace app\common\model_old\erp;

use app\common\model_old\BaseModel;
use app\common\model_old\product\Product;
use app\common\model_old\supplier\Supplier as SupplierModel;

/**
 * 月度商品记录模型
 */
class ErpMonthlyProductStatistics extends BaseModel
{
    protected $name = 'erp_monthly_product_statistics';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 月记录类型 1-月初 2-月末
     */
    const MONTH_START = 1;
    const MONTH_END = 2;

    /**
     * 记录月初库存
     */
    public function recordStart()
    {
        $year = date("Y");
        $month = date("m");
        //
        Product::alias('a')
            ->field('a.*')
            ->leftJoin('erp_monthly_product_statistics s', "s.product_id = a.product_id and s.month = $month and s.year = $year and s.record_type=" . self::MONTH_START)
            ->whereNull('s.id')
            ->chunk(500, function ($list) use ($year, $month) {
                foreach ($list as $product) {
                    $stock = 0;
                    if ($product['type'] == 10) {
                        $stock = $product->product_stock;
                    }
                    if ($product['type'] == 20) {
                        $stock = $product->product_material_stock;
                    }
                    //
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['stock'] = $stock;
                    $data['record_type'] = self::MONTH_START;
                    $data['product_id'] = $product->product_id;
                    $data['shop_supplier_id'] = $product->shop_supplier_id;
                    $data['app_id'] = $product->app_id;
                    (new self)->save($data);
                }
            }, 'a.product_id');
    }

    /**
     * 记录月末库存
     */
    public function recordEnd()
    {
        $year = date("Y");
        $month = date("m");

        // 判断当前日期是否是月末最后一小时
        $lastDayOfMonth = date('Y-m-t');
        $currentDate = date('Y-m-d');
        if (
            $currentDate === $lastDayOfMonth &&
            !self::where('year', $year)->where('month', $month)->where('record_type', self::MONTH_END)->find()
        ) {
            Product::alias('a')
                ->field('a.*')
                ->leftJoin('erp_monthly_product_statistics s', "s.product_id = a.product_id and s.month = $month and s.year = $year and s.record_type=" . self::MONTH_END)
                ->whereNull('s.id')
                ->chunk(500, function ($list) use ($year, $month) {
                    foreach ($list as $product) {
                        $stock = 0;
                        if ($product['type'] == 10) {
                            $stock = $product->product_stock;
                        }
                        if ($product['type'] == 20) {
                            $stock = $product->product_material_stock;
                        }
                        // 记录月初数据
                        $data['year'] = $year;
                        $data['month'] = $month;
                        $data['record_type'] = self::MONTH_END;
                        $data['product_id'] = $product->product_id;
                        $data['stock'] = $stock;
                        $data['shop_supplier_id'] = $product->shop_supplier_id;
                        $data['app_id'] = $product->app_id;
                        (new self)->save($data);
                    }
                }, 'a.product_id');
        }
    }

    // 新商品记录
    public function newProductRecord($product_id)
    {
        $supplier = SupplierModel::where('is_main', '=', 1)->find();
        $shop_supplier_id = $supplier['shop_supplier_id'];
        $app_id = $supplier['app_id'];

        $year = date("Y");
        $month = date("m");
        $saveArr['year'] = $year;
        $saveArr['month'] = $month;
        $saveArr['record_type'] = self::MONTH_START;
        $saveArr['product_id'] = $product_id;
        $saveArr['stock'] = 0;
        $saveArr['shop_supplier_id'] = $shop_supplier_id;
        $saveArr['app_id'] = $app_id;
        $this->save($saveArr);
    }

    /**
     * 月库存记录更新
     */
    public function recordUpdate()
    {
        $year = date("Y");
        $month = date("m");

        Product::order('id ASC')->chunk(500, function ($list) use ($year, $month) {
            foreach ($list as $product) {
                // 当月月数据
                $monthStartRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('product_id', $product->product_id)
                    ->where('record_type', self::MONTH_START)->find();
                $monthEndRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('product_id', $product->product_id)
                    ->where('record_type', self::MONTH_END)->find();
                // 当月初数据更新或添加
                if ($monthStartRecord) {
                    $data['stock'] = Product::getProductStockById($product->product_id);
                    $monthStartRecord->save($data);
                } else {
                    $detail = SupplierModel::where('is_main', '=', 1)->find();
                    $shop_supplier_id = $detail['shop_supplier_id'];
                    $app_id = $detail['app_id'];
                    $model = new self;
                    //
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['record_type'] = self::MONTH_START;
                    $data['product_id'] = $product->product_id;
                    $data['stock'] = Product::getProductStockById($product->product_id);
                    $data['shop_supplier_id'] = $shop_supplier_id;
                    $data['app_id'] = $app_id;
                    $model->save($data);
                }
                // 当月末数据删除
                $monthEndRecord?->delete();
            }
        });
    }
}
