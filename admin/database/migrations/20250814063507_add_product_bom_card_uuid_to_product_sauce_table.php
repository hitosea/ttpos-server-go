<?php

use think\migration\Migrator;

class AddProductBomCardUuidToProductSauceTable extends Migrator
{
    /**
     * 变更方法：为 product_sauce 表新增 product_bom_card_uuid 字段
     */
    public function change()
    {
        $table = $this->table('product_sauce');

        // 新增成本卡ID字段（若不存在）
        if (!$table->hasColumn('product_bom_card_uuid')) {
            $table->addColumn('product_bom_card_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '成本卡ID', 'after' => 'sort']);
        }

        $table->update();
    }
} 