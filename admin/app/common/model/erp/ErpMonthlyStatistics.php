<?php

namespace app\common\model\erp;

use think\facade\Env;
use app\common\library\helper;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;
use app\common\model\product\ProductSku;
use app\common\model\erp\ErpWarehouseForm;
use app\common\model\erp\ErpWarehouseOutForm;
/**
 * 月度报表记录模型
 */
class ErpMonthlyStatistics extends BaseModel
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
            ->where('status', '=', 0)  // 状态,0-success已入库 1-canceled已撤销
            ->sum('num');
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
            ->where('wof.status', '=',)  // 状态,0-success已出库 1-canceled已撤销
            ->where('wof.create_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->sum('wofi.num');
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
        $total = self::where('year', $year)->where('month', $month)->where('scene', self::MONTH_START)->value('stock') ?? 0;
        return floatval($total);
    }

    // 获取月末库存
    public static function getMonthEndStock($params)
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
        $total = self::where('year', $year)->where('month', $month)->where('scene', self::MONTH_END)->value('stock') ?? 0;
        if ($total == 0) {
            // 当月为当前商品库存
            $productNum = ProductBom::sum('stock_num');
            $productMaterialNum = Material::sum('stock_num');
            $total = helper::bcadd($productNum, $productMaterialNum, 4);
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
        // 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
        return (new ErpDamagedProductRecord())->alias('dr')
            ->where('dr.delete_time', 0)
            ->where('dr.status', 1)
            ->where('dr.create_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->sum('dr.num');
    }

    // 所有商品月损耗比例
    public static function getMonthDamagedPercent($month_damaged_num, $month_entry_stock, $month_start_stock)
    {
        $monthTotalStock = helper::bcadd($month_entry_stock, $month_start_stock, 4);
        if ($monthTotalStock > 0) {
            $percent = helper::bcdiv($month_damaged_num, $monthTotalStock, 4);
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
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }
        $list =  (new ErpDamagedProductRecord)->alias('dr')
            ->leftJoin('product_bom pb', 'dr.material_uuid = pb.product_package_uuid')
            ->where('dr.delete_time', 0)
            ->where('dr.status', 1)
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('dr.create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->field('pb.product_package_uuid, pb.name, SUM(num) as damage_count')
            ->group('pb.product_package_uuid')
            ->order('damage_count desc')
            ->limit(10)
            ->select()->toArray();
        // 对结果进行处理
        foreach ($list as &$damageCount) {
            $product_package_uuid = $damageCount['product_package_uuid'];
            $damageCount['name'] = Product::getProductNameTextAttr($damageCount['name']);
            // 本月入库数
            // 其他入库
            $entry_other_num = (new ErpWarehouseForm())->alias('eir')
                ->where('eir.product_package_uuid', $product_package_uuid)
                ->where('eir.create_time', '>=', strtotime($start_time))
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', ErpWarehouseForm::STATUS_SUCCESS)  // 0-success已入库 1-canceled已撤销
                ->sum('num');
            // 采购入库
            $entry_purchase_num = (new ErpPurchaseDetail())->alias('epd')
                ->leftJoin('warehouse_form eir', 'epd.purchase_form_uuid = eir.purchase_order_uuid')
                ->where('epd.product_package_uuid', $product_package_uuid)
                ->where('eir.create_time', '>=', strtotime($start_time))
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', ErpWarehouseForm::STATUS_SUCCESS)  // 0-success已入库 1-canceled已撤销
                ->sum('epd.num');
            $entry_num = helper::bcadd($entry_other_num, $entry_purchase_num, 4);
            // 本月初库存
            $start_num = ErpMonthlyProductStatistics::where('year', $year)->where('month', $month)->where('material_uuid', $product_package_uuid)->value('stock') ?? 0;
            $product_stock = helper::bcadd($start_num, $entry_num, 4);
            $damageCount['damage_count'] = floatval($damageCount['damage_count']);
            $damageCount['damage_ratio'] = $product_stock > 0 ? helper::bcdiv($damageCount['damage_count'], $product_stock, 5) : 0;
            $damageCount['damage_ratio'] = ($damageCount['damage_ratio'] * 100);
            $damageCount['damage_ratio'] = round($damageCount['damage_ratio'], 2);
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
        $start_time = strtotime($start_time);
        $end_time = strtotime($end_time);
        //
        $prefix = Env::get('DB_PREFIX');
        return (new Product)->alias('p')
            ->leftJoin("
                (
                    select id, create_time, product_package_uuid, sum(if(op.id,op.total_price,0)) as total_price, sum(if(op.id,op.num,0)) as total_num
                    from {$prefix}sale_order_product op
                    where op.create_time >= {$start_time} And op.create_time <= {$end_time}
                    group by id, product_package_uuid
                ) op
            ", $prefix . "product.product_package_uuid = op.product_package_uuid")
            ->field('p.product_package_uuid, p.name, if(op.id,op.total_price,0) as total_price, if(op.id,op.num,0) as total_num')
            ->group('p.product_package_uuid')
            ->order('total_num,p.product_package_uuid  asc')
            ->limit(10)
            ->select()
            ->toArray();
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
            $data['stock'] = ProductSku::getTotalStock();
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
            $data['stock'] = ProductSku::getTotalStock();
            $model->save($data);
        }
    }
}
