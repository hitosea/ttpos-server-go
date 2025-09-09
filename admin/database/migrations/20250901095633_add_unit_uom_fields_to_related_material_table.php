<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddUnitUomFieldsToRelatedMaterialTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        // 修改关联材料表，添加ERPNext UOM字段
        $this->updateRelatedMaterialTable();
    }

    /**
     * 修改关联材料表，添加ERPNext UOM字段
     */
    private function updateRelatedMaterialTable()
    {
        $table = $this->table('related_material');

        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('unit_uom')) {
            $table->addColumn('unit_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => '单位ERPNext UOM', 'after' => 'unit_name']);
        }

        if (!$table->hasColumn('base_unit_uom')) {
            $table->addColumn('base_unit_uom', 'string', ['limit' => 255, 'default' => '', 'comment' => '基准单位ERPNext UOM', 'after' => 'base_unit_name']);
        }

        $table->update();
    }
}

