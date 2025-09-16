<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSortToMaterialCategoryTable extends Migrator
{
    /**
     * 添加 sort 字段到 ttpos_material_category 表
     */
    public function change()
    {
        $table = $this->table('material_category');

        // 检查字段是否已存在
        if (!$table->hasColumn('sort')) {
            $table->addColumn('sort', 'integer', ['default' => 0, 'comment' => '排序', 'after' => 'multi_language_name_uuid'])
                  ->update();
        }
    }
}
