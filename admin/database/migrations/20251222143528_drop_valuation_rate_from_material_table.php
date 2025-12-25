<?php

use think\migration\Migrator;

class DropValuationRateFromMaterialTable extends Migrator
{
    /**
     * 删除 ttpos_material 表中的估值率字段（如果存在）
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('material')) {
            return;
        }

        $table = $this->table('material');

        // 检查字段是否存在，如果存在则删除
        // 尝试删除可能的字段名：valuation
        if ($table->hasColumn('valuation')) {
            $table->removeColumn('valuation')->update();
        }
    }
}

