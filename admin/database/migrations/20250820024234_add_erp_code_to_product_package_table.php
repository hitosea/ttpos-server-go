<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpCodeToProductPackageTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        $table = $this->table('product_package');
        
        // 检查表是否存在
        if (!$this->hasTable('product_package')) {
            return;
        }

        // 检查erp_code字段是否已经存在
        if (!$table->hasColumn('erp_code')) {
            // 在name字段后添加erp_code字段
            $table->addColumn('erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext 商品编码，每个商品都有一个模版物品编码', 'after' => 'name'])->update();
        }
    }
}
