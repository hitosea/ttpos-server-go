<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSafetyStockToMaterialTable extends Migrator
{
    /**
     * 添加 safety_stock 字段到 ttpos_material 表
     */
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('safety_stock')) {
            $table->addColumn('safety_stock', 'decimal', ['precision' => 14, 'scale' => 4, 'null' => true, 'default' => null, 'comment' => '安全库存数量', 'after' => 'stock_num'])->update();
        }
    }
}

