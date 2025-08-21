<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpCodeToProductBomCardTable extends Migrator
{
    /**
     * 迁移
     */
    public function change()
    {
        $table = $this->table('product_bom_card');
        
        // 检查表是否存在
        if (!$this->hasTable('product_bom_card')) {
            return;
        }

        // 检查erp_code字段是否已经存在
        if (!$table->hasColumn('erp_code')) {
            // 在name字段后添加erp_code字段
            $table->addColumn('erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext 成本卡编码', 'after' => 'name'])->update();
        }
    }
}
