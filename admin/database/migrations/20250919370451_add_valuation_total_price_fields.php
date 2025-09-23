<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddValuationTotalPriceFields extends Migrator
{
    /**
     * 为采购相关表添加估值和总价字段
     */
    public function change()
    {
        // 为 ttpos_purchase_order_item 表添加字段
        if ($this->hasTable('purchase_order_item')) {
            $table = $this->table('purchase_order_item');
            if (!$table->hasColumn('valuation')) {
                $table->addColumn('valuation', 'decimal', ['precision' => 14, 'scale' => 8, 'default' => 1, 'comment' => '估值单价'])
                      ->update();
            }
            if (!$table->hasColumn('total_price')) {
                $table->addColumn('total_price', 'decimal', ['precision' => 14, 'scale' => 8, 'default' => 0, 'comment' => '总价'])
                      ->update();
            }
        }

        // 为 ttpos_purchase_receipt_order_item 表添加字段
        if ($this->hasTable('purchase_receipt_order_item')) {
            $table = $this->table('purchase_receipt_order_item');
            if (!$table->hasColumn('valuation')) {
                $table->addColumn('valuation', 'decimal', ['precision' => 14, 'scale' => 8, 'default' => 1, 'comment' => '估值单价'])
                      ->update();
            }
            if (!$table->hasColumn('total_price')) {
                $table->addColumn('total_price', 'decimal', ['precision' => 14, 'scale' => 8, 'default' => 0, 'comment' => '总价'])
                      ->update();
            }
        }

        // 为 ttpos_warehouse_item 表添加字段
        if ($this->hasTable('warehouse_item')) {
            $table = $this->table('warehouse_item');
            if (!$table->hasColumn('valuation')) {
                $table->addColumn('valuation', 'decimal', ['precision' => 14, 'scale' => 8, 'default' => 1, 'comment' => '估值单价'])
                      ->update();
            }
        }
    }
}
