<?php

namespace app\common\model\erp;

use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;
use app\common\model\erp\ErpWarehouseForm;
use app\common\model\erp\ErpWarehouseOutForm;
use app\common\model\product\Product;
use think\facade\Db;

/**
 * 月度报表记录模型
 */
class ErpMonthlyStatistics extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_monthly_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;
    protected $pk = 'id';

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

    // 月入库
    public static function getMonthEntry($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }

        return (new ErpWarehouseForm())
            ->where('create_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->whereIn('scene', [0, 1, 2]) // 场景,0-采购入库 1-添加入库 2-调整入库
            ->where('status', '=', 0)  // 状态,0-success已入库
            ->sum('num') ?: 0;
    }

    // 月出库
    public static function getMonthExit($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }

        return (new ErpWarehouseOutForm())->alias('wof')
            ->leftJoin('warehouse_out_form_item wofi', 'wof.uuid = wofi.warehouse_out_form_uuid')
            ->whereIn('wof.scene', [0, 1, 4]) // 场景,0-销售出库 1-调整出库
            ->where('wof.status', 0)  // 状态,0-success已出库 1-canceled已撤销
            ->where('wof.update_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->sum('wofi.num') ?: 0;
    }

    // 获取月初库存
    public static function getMonthStartStock($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
        }
        $total = self::where('year', $year)
            ->where('month', $month)
            ->where('scene', self::MONTH_START)
            ->sum('stock') ?? 0;

        return floatval($total);
    }

    // 获取月末库存
    public static function getMonthEndStock($params)
    {
        $now = time();
        $nowYear = date("Y", $now);
        $nowMonth = date("m", $now);
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
        } else {
            $year = $nowYear;
            $month = $nowMonth;
        }
        // 如果月末数据未生成，则显示当前所有规格和材料的库存
        $total = self::where('year', $year)
            ->where('month', $month)
            ->where('scene', self::MONTH_END)
            ->value('stock') ?? 0;

        if ($total == 0 && $year == $nowYear && $month == $nowMonth) {
            $total = (new self())->getTotalStockNum();
        }

        return floatval($total);
    }

    // 所有商品月损耗数量
    public static function getMonthDamagedNum($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }
        
        return (new ErpDamagedProductRecord())
            ->where('status', 1) // 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
            ->where('create_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->sum('num') ?: 0;
    }

    // 所有商品月损耗比例
    public static function getMonthDamagedPercent($monthDamaged_num, $monthEntry_stock, $monthStart_stock)
    {
        $monthTotalStock = helper::bcadd($monthEntry_stock, $monthStart_stock, 4);
        if ($monthTotalStock > 0) {
            $percent = helper::bcdiv($monthDamaged_num, $monthTotalStock, 4);
            $re = helper::bcmul($percent, 100);
            return floatval($re);
        }
        return 0;
    }

    // 月商品耗损排行榜
    public static function getMonthProductDamagedList($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $startTime = date($year . "-" . $month . "-01");
            $endTime = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $startTime = date($year . "-" . $month . "-01");
            $endTime = date($year . "-" . $month . "-t 23:59:59");
        }
        
        $recordList = ErpDamagedProductRecord::with([
            'sku' => function($q) use ($year, $month, $startTime, $endTime) {
                return $q->with([
                    'product' => function($q) {
                        return $q->withTrashed();
                    },
                    'erpMonthlyProductStatistics' => function($q) use ($year, $month) {
                        return $q->where('year', $year)->where('month', $month)->where('scene', self::MONTH_START);
                    },
                    'erpInventoryRecord' => function($q) use ($startTime, $endTime) {
                        return $q->where('status', 0)
                            ->whereIn('scene', [0, 1, 2])
                            ->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)]);
                    }
                ])->withTrashed();
            },
            'material' => function($q) use ($year, $month, $startTime, $endTime) {
                return $q->with([
                    'erpMonthlyMaterialStatistics' => function($q) use ($year, $month) {
                        return $q->where('year', $year)->where('month', $month)->where('scene', self::MONTH_START);
                    },
                    'erpInventoryRecord' => function($q) use ($startTime, $endTime) {
                        return $q->where('status', 0)
                            ->whereIn('scene', [0, 1, 2])
                            ->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)]);
                    }
                ])->withTrashed();
            },
        ])
            ->where('status', 1)
            ->when($startTime && $endTime, function($q) use ($startTime, $endTime) {
                $q->where('create_time', 'between', [strtotime($startTime), strtotime($endTime)]);
            })
            ->field('product_bom_uuid, material_uuid, SUM(num) as damage_count')
            ->group('product_bom_uuid, material_uuid')
            ->order('num desc')
            ->limit(10)
            ->select();

        $list = [];
        foreach ($recordList as $record) {
            $product = $record->sku ? $record->sku->product : $record->material;
            $productName = $product->product_name_text;
            
            $damageCount = $record->damage_count; // 损耗数量
            $monthStartStock = 0; // 月初库存数
            $monthEntryStock = 0; // 月入库存数
            if ($record->sku) {
                $productName .= ' - ' . $record->sku->spec_name_text;
                foreach ($record->sku->erpMonthlyProductStatistics as $item) {
                    $monthStartStock = helper::bcadd($monthStartStock, $item->stock);
                }
                foreach ($record->sku->erpInventoryRecord as $item) {
                    $monthEntryStock = helper::bcadd($monthEntryStock, $item->num);
                }   
            }
            if ($record->material) {
                foreach ($record->material->erpMonthlyMaterialStatistics as $item) {
                    $monthStartStock = helper::bcadd($monthStartStock, $item->stock);
                }
                foreach ($record->material->erpInventoryRecord as $item) {
                    $monthEntryStock = helper::bcadd($monthEntryStock, $item->num);
                }
            }

            // 损耗数量/（月初原有库存+月入库数量）*100%=损耗比例（保留小数点后两位）
            $damageRatio = 0;
            $stock = helper::bcadd($monthStartStock, $monthEntryStock, 4);
            if ($stock > 0) {
                $damageRatio = helper::bcmul(helper::bcdiv($damageCount, $stock, 5), 100, 2);
            }

            $list[] = [
                'product_id' => $record->product_bom_uuid ?: $record->material_uuid,
                'product_name' => $productName,
                'damage_count' => $damageCount,
                'damage_ratio' => $damageRatio,
            ];
        }

        usort($list, function ($item1, $item2) {
            return $item2['damage_ratio'] <=> $item1['damage_ratio'];
        });

        return $list;
    }

    // 月商品滞销排行榜
    public static function getMonthProductUnsalableList($params)
    {
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }
        $startTime = strtotime($start_time);
        $endTime = strtotime($end_time);

        $bomUuidList = ErpWarehouseOutFormItem::where('scene', 0)
            ->where('update_time', 'between', [$startTime, $endTime])
            ->where('status', 1)
            ->group('product_bom_uuid')
            ->order('update_time desc')
            ->column('product_bom_uuid');

        $rows = ProductBom::alias('bom')
            ->field('p.name as name,bom.create_time as create_time')
            ->leftJoin('product_package p', 'bom.product_package_uuid = p.uuid')
            ->whereNotIn('bom.uuid', $bomUuidList)
            ->where('p.create_time', '<=', $startTime)
            ->group('bom.product_package_uuid')
            ->order('bom.id desc')
            ->limit(10)
            ->select();

        $list = [];
        foreach ($rows as $row) {
            $productNametext = extractLanguage($row['name']);

            $list[] = [
                'product_name_text' => $productNametext,
                'total_num' => 0
            ];
        }

        return $list;
    }

    /**
     * 记录库存
     */
    public function record()
    {
        $year = date("Y");
        $month = date("m");

        // 记录月初数据
        if (!self::where('year', $year)->where('month', $month)->where('scene', self::MONTH_START)->find()) {
            $model = new self;
            $data['year'] = $year;
            $data['month'] = $month;
            $data['scene'] = self::MONTH_START;
            $data['stock'] = $this->getTotalStockNum();
            $model->save($data);
        }

        // 判断当前日期是否是月末最后一分钟
        $lastDayOfMonth = date('Y-m-t 23:59');
        $currentDate = date('Y-m-d H:i');
        if (
            $currentDate === $lastDayOfMonth &&
            !self::where('year', $year)->where('month', $month)->where('scene', self::MONTH_END)->find()
        ) {
            $model = new self;
            $data['year'] = $year;
            $data['month'] = $month;
            $data['scene'] = self::MONTH_END;
            $data['stock'] = $this->getTotalStockNum();
            $model->save($data);
        }
    }

    /**
     * 获取总库存数
     */
    public function getTotalStockNum()
    {
        // 规格库存
        $productBomStockNum = ProductBom::getTotalStockNum();
        // 材料库存
        $materialStockNum = Material::getTotalStockNum();

        return floatval(helper::bcadd($productBomStockNum, $materialStockNum, 4));
    }
}
