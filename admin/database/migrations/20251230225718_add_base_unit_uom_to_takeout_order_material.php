<?php

use think\migration\Migrator;

/**
 * 外卖订单原料表添加基准单位UOM字段
 * @version v2.12.0
 * @spec story-erp-grab-invoice-sync
 */
class AddBaseUnitUomToTakeoutOrderMaterial extends Migrator
{
    /**
     * 执行迁移
     */
    public function change()
    {
        if (!$this->hasTable('takeout_order_material')) {
            return;
        }

        $table = $this->table('takeout_order_material');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('base_unit_uom')) {
            $table->addColumn('base_unit_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => '基准单位ERPNext UOM(来自RelatedMaterial.BaseUnitUom)', 'after' => 'erp_code'])
                  ->update();
        }
    }
}

