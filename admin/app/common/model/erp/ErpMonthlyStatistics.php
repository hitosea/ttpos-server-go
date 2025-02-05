<?php

namespace app\common\model\erp;

use think\facade\Env;
use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\product\Product;
use app\common\model\product\ProductSku;
use app\common\model\supplier\Supplier as SupplierModel;

/**
 * 月度报表记录模型
 */
class ErpMonthlyStatistics extends BaseModel
{
    protected $name = 'erp_monthly_statistics';

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
        return (new ErpInventoryRecord())
            ->where('delete_time', 0)
            ->where('create_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->where('status', '=', 10)  // 10-已入库 20-已出库
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
        return (new ErpInventoryRecord())
            ->where('delete_time', 0)
            ->where('out_time', 'between', [strtotime($start_time), strtotime($end_time)])
            ->where('status', '=', 20)  // 10-已入库 20-已出库
            ->sum('num');
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
        $total = self::where('year', $year)->where('month', $month)->where('record_type', self::MONTH_START)->value('stock') ?? 0;
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
        $total = self::where('year', $year)->where('month', $month)->where('record_type', self::MONTH_END)->value('stock') ?? 0;
        if ($total == 0) {
            // 当月为当前商品库存
            $productNum = Product::where('type', 10)->where('is_delete', 0)->sum('product_stock');
            $productMaterialNum = Product::where('type', 20)->where('is_delete', 0)->sum('product_material_stock');
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
        return ErpDamagedProductRecord::where('create_time', 'between', [strtotime($start_time), strtotime($end_time)])->where('review_status', 1)->sum('num');
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
            ->join('product p', 'dr.product_id = p.product_id')
            ->where('dr.delete_time', 0)
            ->where('dr.review_status', 1)
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('dr.create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->field('p.product_id, p.product_name, SUM(num) as damage_count')
            ->group('p.product_id')
            ->order('damage_count desc')
            ->limit(10)
            ->select()->toArray();
        // 对结果进行处理
        foreach ($list as &$damageCount) {
            $product_id = $damageCount['product_id'];
            $damageCount['product_name'] = Product::getProductNameTextAttr($damageCount['product_name']);
            // 本月入库数
            // 其他入库
            $entry_other_num = (new ErpInventoryRecord())->alias('eir')
                ->where('eir.delete_time', 0)
                ->where('eir.product_id', $product_id)
                ->where('eir.create_time', '>=', strtotime($start_time))
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', 10)  // 10-已入库 20-已出库
                ->sum('num');
            // 采购入库
            $entry_purchase_num = (new ErpPurchaseDetail())->alias('epd')
                ->join('erp_inventory_record eir', 'epd.purchase_order_id = eir.purchase_order_id')
                ->where('eir.delete_time', 0)
                ->where('epd.product_id', $product_id)
                ->where('eir.create_time', '>=', strtotime($start_time))
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', 10)  // 10-已入库 20-已出库
                ->sum('epd.actual_purchase_num');
            $entry_num = helper::bcadd($entry_other_num, $entry_purchase_num, 4);
            // 本月初库存
            $start_num = ErpMonthlyProductStatistics::where('year', $year)->where('month', $month)->where('product_id', $product_id)->value('stock') ?? 0;
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
                    select order_product_id, create_time, product_id, sum(if(op.order_product_id,op.total_pay_price,0)) as total_price, sum(if(op.order_product_id,op.total_num,0)) as total_num
                    from {$prefix}order_product op
                    where op.create_time >= {$start_time} And op.create_time <= {$end_time}
                    group by order_product_id, product_id
                ) op
            ", $prefix . "product.product_id = op.product_id")
            ->field('p.product_id, p.product_name, if(op.order_product_id,op.total_price,0) as total_price, if(op.order_product_id,op.total_num,0) as total_num')
            ->where('p.type', Product::TYPE_PRODUCT)
            ->where('p.is_delete', 0)
            ->group('p.product_id')
            ->order('total_num,p.product_id  asc')
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
        if (!self::where('year', $year)->where('month', $month)->where('record_type', self::MONTH_START)->find()) {
            $detail = SupplierModel::where('is_main', '=', 1)->find();
            if (empty($detail)) {
                return false;
            }
            $shop_supplier_id = $detail['shop_supplier_id'];
            $app_id = $detail['app_id'];
            $model = new self;
            //
            $data['year'] = $year;
            $data['month'] = $month;
            $data['record_type'] = self::MONTH_START;
            $data['stock'] = ProductSku::getTotalStock();
            $data['shop_supplier_id'] = $shop_supplier_id;
            $data['app_id'] = $app_id;
            $model->save($data);
        }

        // 判断当前日期是否是月末最后一分钟
        $lastDayOfMonth = date('Y-m-t 23:59');
        $currentDate = date('Y-m-d H:i');
        if (
            $currentDate === $lastDayOfMonth &&
            !self::where('year', $year)->where('month', $month)->where('record_type', self::MONTH_END)->find()
        ) {
            $detail = SupplierModel::where('is_main', '=', 1)->find();
            if (empty($detail)) {
                return false;
            }
            $shop_supplier_id = $detail['shop_supplier_id'];
            $app_id = $detail['app_id'];
            $model = new self;
            //
            $data['year'] = $year;
            $data['month'] = $month;
            $data['record_type'] = self::MONTH_END;
            $data['stock'] = ProductSku::getTotalStock();
            $data['shop_supplier_id'] = $shop_supplier_id;
            $data['app_id'] = $app_id;
            $model->save($data);
        }
    }
}
