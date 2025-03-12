<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

class ErpMonthlyMaterialStatistics extends BaseModel
{
    use SoftDelete;

    protected $name = 'warehouse_monthly_material_form';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $autoWriteTimestamp = true;

     /**
     * 月记录类型 0-月初 1-月末
     */
    const MONTH_START = 0;
    const MONTH_END = 1;

    /**
     * 新增材料月初库存记录
     */
    public static function newMaterialRecord($materialUuid)
    {
        $year = date("Y");
        $month = date("m");
        $saveArr['year'] = $year;
        $saveArr['month'] = $month;
        $saveArr['scene'] = self::MONTH_START;
        $saveArr['product_bom_uuid'] = $materialUuid;
        $saveArr['stock'] = 0;
        self::create($saveArr);
    }
}