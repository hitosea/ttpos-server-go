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
    protected $name = 'warehouse_monthly_product_bom_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

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
        ProductBom::where('product_flavor_uuid', '>', 0)
            ->chunk(500, function($list) use ($year, $month) {
                foreach ($list as $item) {
                    $record = self::where('year', $year)
                        ->where('month', $month)
                        ->where('scene', self::MONTH_START)
                        ->where('product_bom_uuid', $item['uuid'])
                        ->find();
                    if (!$record) {
                        self::create([
                            'year' => $year,
                            'month' => $month,
                            'scene' => self::MONTH_START,
                            'product_bom_uuid' => $item['uuid'],
                            'stock' => $item['stock_num'],
                        ]);
                    }
                }
            });
        
        // 材料
        Material::chunk(500, function($list) use ($year, $month) {
            foreach ($list as $item) {
                $record = ErpMonthlyMaterialStatistics::where('year', $year)
                    ->where('month', $month)
                    ->where('scene', self::MONTH_START)
                    ->where('material_uuid', $item['uuid'])
                    ->find();
                if (!$record) {
                    ErpMonthlyMaterialStatistics::create([
                        'year' => $year,
                        'month' => $month,
                        'scene' => self::MONTH_START,
                        'material_uuid' => $item['uuid'],
                        'stock' => $item['stock_num'],
                    ]);
                }
            }
        });
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
        if ($currentDate === $lastDayOfMonth) {
            // 商品规格
            ProductBom::where('product_flavor_uuid', '>', 0)
                ->chunk(500, function($list) use ($year, $month) {
                    foreach ($list as $item) {
                        $record = self::where('year', $year)
                            ->where('month', $month)
                            ->where('scene', self::MONTH_END)
                            ->where('product_bom_uuid', $item['uuid'])
                            ->find();
                        if (!$record) {
                            self::create([
                                'year' => $year,
                                'month' => $month,
                                'scene' => self::MONTH_END,
                                'product_bom_uuid' => $item['uuid'],
                                'stock' => $item['stock_num'],
                            ]);
                        }
                    }
                });

            // 材料
            Material::chunk(500, function($list) use ($year, $month) {
                foreach ($list as $item) {
                    $record = ErpMonthlyMaterialStatistics::where('year', $year)
                        ->where('month', $month)
                        ->where('scene', self::MONTH_END)
                        ->where('material_uuid', $item['uuid'])
                        ->find();
                    if (!$record) {
                        ErpMonthlyMaterialStatistics::create([
                            'year' => $year,
                            'month' => $month,
                            'scene' => self::MONTH_END,
                            'material_uuid' => $item['uuid'],
                            'stock' => $item['stock_num'],
                        ]);
                    }
                }
            });
        }
    }

    // 新商品记录
    public static function newProductRecord($productBomUuid)
    {
        $year = date("Y");
        $month = date("m");
        $saveArr['year'] = $year;
        $saveArr['month'] = $month;
        $saveArr['scene'] = self::MONTH_START;
        $saveArr['product_bom_uuid'] = $productBomUuid;
        $saveArr['stock'] = 0;
        self::create($saveArr);
    }
}
