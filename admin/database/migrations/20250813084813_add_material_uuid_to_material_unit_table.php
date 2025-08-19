<?php

use think\migration\Migrator;

class AddMaterialUuidToMaterialUnitTable extends Migrator
{
    /**
     * 变更方法：为 material_unit 表新增 material_uuid 字段
     */
    public function change()
    {
        $table = $this->table('material_unit');

        // 新增原料ID字段（若不存在）
        if (!$table->hasColumn('material_uuid')) {
            $table->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '原料ID', 'after' => 'is_default']);
        }

        $table->update();
    }
}

