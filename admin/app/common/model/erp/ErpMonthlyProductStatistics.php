<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use think\facade\Db;

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
     * 记录月初库存（批量）
     */
    public function recordStart()
    {
        $year = date("Y");
        $month = date("m");
        $now = time();

        $this->batchInsertProductBom($year, $month, self::MONTH_START, $now);
        $this->batchInsertMaterial($year, $month, self::MONTH_START, $now);
    }

    /**
     * 记录月末库存（批量）
     */
    public function recordEnd()
    {
        // 仅月末最后一天执行
        if (date('Y-m-d') !== date('Y-m-t')) {
            return;
        }

        $year = date("Y");
        $month = date("m");
        $now = time();

        $this->batchInsertProductBom($year, $month, self::MONTH_END, $now);
        $this->batchInsertMaterial($year, $month, self::MONTH_END, $now);
    }

    /**
     * 批量插入商品规格库存记录
     */
    private function batchInsertProductBom(string $year, string $month, int $scene, int $now): void
    {
        $db = Db::connect($this->connection);
        $db->execute(
            "INSERT INTO ttpos_warehouse_monthly_product_bom_form
                (year, month, scene, product_bom_uuid, stock, create_time, update_time, delete_time)
            SELECT ?, ?, ?, pb.uuid, pb.stock_num, ?, ?, 0
            FROM ttpos_product_bom pb
            WHERE pb.product_flavor_uuid > 0
              AND pb.delete_time = 0
              AND NOT EXISTS (
                  SELECT 1 FROM ttpos_warehouse_monthly_product_bom_form wmpf
                  WHERE wmpf.year = ?
                    AND wmpf.month = ?
                    AND wmpf.scene = ?
                    AND wmpf.product_bom_uuid = pb.uuid
                    AND wmpf.delete_time = 0
              )",
            [$year, $month, $scene, $now, $now, $year, $month, $scene]
        );
    }

    /**
     * 批量插入材料库存记录
     */
    private function batchInsertMaterial(string $year, string $month, int $scene, int $now): void
    {
        $db = Db::connect($this->connection);
        $db->execute(
            "INSERT INTO ttpos_warehouse_monthly_material_form
                (year, month, scene, material_uuid, stock, create_time, update_time, delete_time)
            SELECT ?, ?, ?, m.uuid, m.stock_num, ?, ?, 0
            FROM ttpos_material m
            WHERE m.delete_time = 0
              AND NOT EXISTS (
                  SELECT 1 FROM ttpos_warehouse_monthly_material_form wmf
                  WHERE wmf.year = ?
                    AND wmf.month = ?
                    AND wmf.scene = ?
                    AND wmf.material_uuid = m.uuid
                    AND wmf.delete_time = 0
              )",
            [$year, $month, $scene, $now, $now, $year, $month, $scene]
        );
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
