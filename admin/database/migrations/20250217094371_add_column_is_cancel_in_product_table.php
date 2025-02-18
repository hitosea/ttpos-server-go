<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnIsCancelInProductTable extends Migrator
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
        // 销售订单表
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('is_cancel')) {
            $table->addColumn('is_cancel', 'tinyinteger', ['default' => 0, 'comment' => '是否退菜, 0-否 1-是', 'after' => 'is_gift'])
                  ->update();
        }
        $table->update();
    }
}
