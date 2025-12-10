<?php
use think\migration\Migrator;

class ModifyProductBomCardLogFieldsToText extends Migrator
{
    /**
     * 修改 product_bom_card_log 表的 related_name 和 product_bom_card_name 字段类型为 TEXT
     * Purpose: 确保字段类型为 TEXT，支持存储更长的 JSON 数据
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('product_bom_card_log')) {
            $table = $this->table('product_bom_card_log');

            // 检查并修改 product_bom_card_name 字段
            if ($table->hasColumn('product_bom_card_name')) {
                $table->changeColumn('product_bom_card_name', 'text', [
                    'comment' => '成本卡名称JSON'
                ])->update();
            }

            // 检查并修改 related_name 字段
            if ($table->hasColumn('related_name')) {
                $table->changeColumn('related_name', 'text', [
                    'comment' => '关联名称JSON,商品名称、加料名称'
                ])->update();
            }
        }
    }
}

