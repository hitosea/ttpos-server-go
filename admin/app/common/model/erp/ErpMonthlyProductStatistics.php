<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;

/**
 * 月度商品记录模型
 */
class ErpMonthlyProductStatistics extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_monthly_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 月记录类型 0-月初 1-月末
     */
    const MONTH_START = 0;
    const MONTH_END = 1;

    /**
     * 记录月初库存
     */
    public function recordStart()
    {
        $year = date("Y");
        $month = date("m");
        // 商品规格
        ProductBom::alias('a')
            ->field('a.*')
            ->leftJoin('warehouse_monthly_form s', "s.material_uuid = a.product_package_uuid and s.month = $month and s.year = $year and s.scene=" . self::MONTH_START)
            ->whereNull('s.id')
            ->chunk(500, function ($list) use ($year, $month) {
                foreach ($list as $product) {
                    $stock = $product->stock_num ?: 0;
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['stock'] = $stock;
                    $data['scene'] = self::MONTH_START;
                    $data['material_uuid'] = $product->product_package_uuid;
                    (new self)->save($data);
                }
            }, 'a.product_package_uuid');

        // 原料
        Material::alias('a')
            ->field('a.*')
            ->leftJoin('warehouse_monthly_form s', "s.material_uuid = a.uuid and s.month = $month and s.year = $year and s.scene=" . self::MONTH_START)
            ->whereNull('s.id')
            ->chunk(500, function ($list) use ($year, $month) {
                foreach ($list as $product) {
                    $stock = $product->stock_num ?: 0;
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['stock'] = $stock;
                    $data['scene'] = self::MONTH_START;
                    $data['material_uuid'] = $product->uuid;
                    (new self)->save($data);
                }
            }, 'a.uuid');
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
            !self::where('year', $year)->where('month', $month)->where('scene', self::MONTH_END)->find()
        ) {
            // 商品规格
            ProductBom::alias('a')
                ->field('a.*')
                ->leftJoin('warehouse_monthly_form s', "s.material_uuid = a.product_package_uuid and s.month = $month and s.year = $year and s.scene=" . self::MONTH_END)
                ->whereNull('s.id')
                ->chunk(500, function ($list) use ($year, $month) {
                    foreach ($list as $product) {
                        $stock = $product->stock_num ?: 0;
                        $data['year'] = $year;
                        $data['month'] = $month;
                        $data['scene'] = self::MONTH_END;
                        $data['material_uuid'] = $product->product_package_uuid;
                        $data['stock'] = $stock;
                        (new self)->save($data);
                    }
                }, 'a.product_package_uuid');

            // 原料
            Material::alias('a')
                ->field('a.*')
                ->leftJoin('warehouse_monthly_form s', "s.material_uuid = a.uuid and s.month = $month and s.year = $year and s.scene=" . self::MONTH_END)
                ->whereNull('s.id')
                ->chunk(500, function ($list) use ($year, $month) {
                    foreach ($list as $product) {
                        $stock = $product->stock_num ?: 0;
                        $data['year'] = $year;
                        $data['month'] = $month;
                        $data['scene'] = self::MONTH_END;
                        $data['material_uuid'] = $product->uuid;
                        $data['stock'] = $stock;
                        (new self)->save($data);
                    }
                }, 'a.uuid');
        }
    }

    // 新商品记录
    public function newProductRecord($product_id)
    {
        $year = date("Y");
        $month = date("m");
        $saveArr['year'] = $year;
        $saveArr['month'] = $month;
        $saveArr['scene'] = self::MONTH_START;
        $saveArr['material_uuid'] = $product_id;
        $saveArr['stock'] = 0;
        $this->save($saveArr);
    }

    /**
     * 月库存记录更新
     */
    public function recordUpdate()
    {
        $year = date("Y");
        $month = date("m");

        // 商品规格
        ProductBom::order('id ASC')->chunk(500, function ($list) use ($year, $month) {
            foreach ($list as $product) {
                // 当月月数据
                $monthStartRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('material_uuid', $product->product_package_uuid)
                    ->where('scene', self::MONTH_START)->find();
                $monthEndRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('material_uuid', $product->product_package_uuid)
                    ->where('scene', self::MONTH_END)->find();
                // 当月初数据更新或添加
                if ($monthStartRecord) {
                    $data['stock'] = $product->stock_num ?: 0;
                    $monthStartRecord->save($data);
                } else {
                    $model = new self;
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['scene'] = self::MONTH_START;
                    $data['material_uuid'] = $product->product_package_uuid;
                    $data['stock'] = $product->stock_num ?: 0;
                    $model->save($data);
                }
                // 当月末数据删除
                $monthEndRecord?->delete();
            }
        });

        // 原料
        Material::order('id ASC')->chunk(500, function ($list) use ($year, $month) {
            foreach ($list as $product) {
                // 当月月数据
                $monthStartRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('material_uuid', $product->uuid)
                    ->where('scene', self::MONTH_START)->find();
                $monthEndRecord = self::where('year', $year)
                    ->where('month', $month)
                    ->where('material_uuid', $product->uuid)
                    ->where('scene', self::MONTH_END)->find();
                // 当月初数据更新或添加
                if ($monthStartRecord) {
                    $data['stock'] = $product->stock_num ?: 0;
                    $monthStartRecord->save($data);
                } else {
                    $model = new self;
                    $data['year'] = $year;
                    $data['month'] = $month;
                    $data['scene'] = self::MONTH_START;
                    $data['material_uuid'] = $product->uuid;
                    $data['stock'] = $product->stock_num ?: 0;
                    $model->save($data);
                }
                // 当月末数据删除
                $monthEndRecord?->delete();
            }
        });
    }
}
