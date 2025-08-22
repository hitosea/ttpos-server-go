<?php

use think\migration\Migrator;

class AddErpCodeToProductSauceTable extends Migrator
{
    /**
     * 变更方法：为 product_sauce 表新增 erp_code 字段
     */
    public function change()
    {
        $table = $this->table('product_sauce');

        // 新增ERP编码字段（若不存在）
        if (!$table->hasColumn('erp_code')) {
            $table->addColumn('erp_code', 'string', ['limit' => 100, 'default' => '', 'comment' => 'ERP编码', 'after' => 'product_bom_card_uuid']);
        }

        $table->update();
    }
}
