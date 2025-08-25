<?php

use think\migration\Migrator;

class AddErpCodeToProductBomTable extends Migrator
{

    /**
     * 变更方法：为 product_bom 表新增 erp_code 字段
     */
    public function change()
    {
        $table = $this->table('product_bom');

        // 新增商品编码字段（若不存在）
        if (!$table->hasColumn('erp_code')) {
            $table->addColumn('erp_code', 'string', ['limit' => 255, 'default' => '', 'comment' => '商品编码', 'after' => 'name']);
        }

        $table->update();
    }
}
