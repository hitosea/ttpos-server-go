<?php
/**
 * 为外卖订单表添加 erp_stock_deducted 字段
 * 用于标记外卖订单是否已通过 Stock Entry 扣减库存
 */

use think\migration\Migrator;

class AddErpStockDeductedToTakeoutOrder extends Migrator
{
    public function change()
    {
        if ($this->hasTable('takeout_order')) {
            $table = $this->table('takeout_order');
            if (!$table->hasColumn('erp_stock_deducted')) {
                $table->addColumn('erp_stock_deducted', 'integer', [
                    'limit' => 1,
                    'default' => 0,
                    'null' => false,
                    'comment' => '库存是否已通过StockEntry扣减',
                    'after' => 'erp_pos_invoice_resp',
                ])->update();
            }
        }
    }
}
