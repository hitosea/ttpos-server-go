<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddInternalCodeToMaterialTable extends Migrator
{
    /**
     * 添加 internal_code 字段到 ttpos_material 表
     */
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('internal_code')) {
            $table->addColumn('internal_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '内部编码', 'after' => 'barcode_value'])
                  ->update();
        }
    }
}
