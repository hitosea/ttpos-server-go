<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAllowSubstoreVisibleToMaterialTable extends Migrator
{
    /**
     * 添加 allow_substore_visible 字段到 ttpos_material 表
     */
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('allow_substore_visible')) {
            $table->addColumn('allow_substore_visible', 'integer', ['limit' => 1, 'null' => false, 'default' => 1, 'comment' => '允许子店可见：1-允许，0-不允许', 'after' => 'warehouse_uuid'])->update();
        }

        // 检查索引是否已存在
        if (!$table->hasIndex('idx_allow_substore_visible')) {
            $table->addIndex('allow_substore_visible', ['name' => 'idx_allow_substore_visible'])->update();
        }
    }
}

