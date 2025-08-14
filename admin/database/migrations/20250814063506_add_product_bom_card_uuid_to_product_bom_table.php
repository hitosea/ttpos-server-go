<?php

use think\migration\Migrator;

class AddProductBomCardUuidToProductBomTable extends Migrator
{
    /**
     * 变更方法：为 product_bom 表新增 product_bom_card_uuid 字段
     */
    public function change()
    {
        $table = $this->table('product_bom');

        // 新增成本卡ID字段（若不存在）
        if (!$table->hasColumn('product_bom_card_uuid')) {
            $table->addColumn('product_bom_card_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '成本卡ID', 'after' => 'product_package_uuid']);
        }

        $table->update();
    }
} 