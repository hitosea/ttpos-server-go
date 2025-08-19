<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;
use Phinx\Db\Table\Column;

final class AddUnitNumToSaleOrderProductTable extends AbstractMigration
{
    /**
     * 添加unit_num字段到ttpos_sale_order_product表
     */
    public function up(): void
    {
        $table = $this->table('sale_order_product');
        
        // 检查表是否存在
        if (!$table->exists()) {
            return;
        }
        
        // 检查字段是否已存在
        if (!$table->hasColumn('unit_num')) {
            $table->addColumn('unit_num', 'decimal', [
                'precision' => 12,
                'scale' => 4,
                'default' => 0,
                'null' => false,
                'comment' => '单位数量，用于套餐子商品',
                'after' => 'num'
            ])
            ->update();
        }
    }

    /**
     * 回滚：删除unit_num字段
     */
    public function down(): void
    {
        $table = $this->table('sale_order_product');
        
        if ($table->exists() && $table->hasColumn('unit_num')) {
            $table->removeColumn('unit_num')
                  ->update();
        }
    }
}