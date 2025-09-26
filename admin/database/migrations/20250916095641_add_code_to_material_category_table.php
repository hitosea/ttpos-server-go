<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCodeToMaterialCategoryTable extends Migrator
{
    /**
     * 添加 code 字段到 ttpos_material_category 表
     */
    public function change()
    {
        $table = $this->table('material_category');

        // 检查字段是否已存在
        if (!$table->hasColumn('code')) {
            $table->addColumn('code', 'string', ['limit' => 255, 'default' => '', 'comment' => '原料分类编码', 'after' => 'name'])
                  ->update();
        }
        // 修改name字段为text
        if ($table->hasColumn('name')) {
            $table->changeColumn('name', 'text', ['comment' => '原料分类名称'])
                  ->update();
        }
    }
}
