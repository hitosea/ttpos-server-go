<?php

use think\migration\Migrator;

class DropSellableQuantityFromProductBomTable extends Migrator
{
    /**
     * 删除 ttpos_product_bom 表中的 sellable_quantity 字段（如果存在）
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('product_bom')) {
            return;
        }

        $table = $this->table('product_bom');

        // 检查字段是否存在，如果存在则删除
        if ($table->hasColumn('sellable_quantity')) {
            $table->removeColumn('sellable_quantity')->update();
        }
    }
}

