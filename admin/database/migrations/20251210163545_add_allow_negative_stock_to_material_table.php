<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAllowNegativeStockToMaterialTable extends Migrator
{
    /**
     * 添加 allow_negative_stock 字段到 ttpos_material 表
     */
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('allow_negative_stock')) {
            $table->addColumn('allow_negative_stock', 'integer', ['limit' => 1, 'null' => false, 'default' => 0, 'comment' => '是否允许负库存：1-允许，0-不允许', 'after' => 'origin_country_code'])->update();
        }

        // 添加索引（如果已存在则忽略）
        try {
            $table->addIndex('allow_negative_stock', ['name' => 'idx_allow_negative_stock'])->update();
        } catch (\Exception $e) {
            // 索引已存在或其他错误，忽略
        }
    }
}

