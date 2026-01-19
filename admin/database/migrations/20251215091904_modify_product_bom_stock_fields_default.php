<?php

use think\migration\Migrator;

class ModifyProductBomStockFieldsDefault extends Migrator
{
    /**
     * 修改 product_bom 表的 use_bom_card_stock 和 is_open_stock 字段默认值为 0
     * 并将 is_open_stock 字段的所有现有值设置为 0
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('product_bom')) {
            return;
        }

        $table = $this->table('product_bom');

        // 修改 use_bom_card_stock 字段默认值为 0
        if ($table->hasColumn('use_bom_card_stock')) {
            $table->changeColumn('use_bom_card_stock', 'integer', [
                'limit' => 10,
                'null' => false,
                'default' => 0,
                'comment' => '是否使用成本卡库存，0-否 1-是'
            ])->update();
        }

        // 先将 is_open_stock 字段的所有现有值设置为 0
        $this->execute("UPDATE `ttpos_product_bom` SET `is_open_stock` = 0 WHERE `delete_time` = 0");

        // 修改 is_open_stock 字段默认值为 0
        if ($table->hasColumn('is_open_stock')) {
            $table->changeColumn('is_open_stock', 'integer', [
                'limit' => 10,
                'null' => false,
                'default' => 0,
                'comment' => '是否开启库存, 0-否 1-是'
            ])->update();
        }
    }
}

