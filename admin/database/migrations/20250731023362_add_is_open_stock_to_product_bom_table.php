<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsOpenStockToProductBomTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        $table = $this->table('product_bom');
        
        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('is_open_stock')) {
            $table->addColumn('is_open_stock', 'integer', [
                'null' => false,
                'default' => 1,
                'comment' => '是否开启库存, 0-否 1-是',
                'after' => 'actual_sale_num'
            ])
            ->update();
        }
    }
} 