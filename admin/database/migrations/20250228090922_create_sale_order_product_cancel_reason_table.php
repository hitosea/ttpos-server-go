<?php

use think\migration\Migrator;


class CreateSaleOrderProductCancelReasonTable extends Migrator
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
        if (!$this->hasTable('sale_order_product_cancel_reason')) {
            $table = $this->table('sale_order_product_cancel_reason', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '销售订单商品退菜原因表',
                'id' => 'id',
                'signed' => false,
            ]);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '自增UUID']);
            $table->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售订单ID']);
            $table->addColumn('sale_order_product_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售订单商品ID']);
            $table->addColumn('return_food_reason_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '退菜原因ID']);
            $table->addColumn('multi_language_name_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '退菜原因-多语言名称ID']);
            $table->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)']);
            $table->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)']);
            $table->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)']);
            $table->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid']);
            $table->create();
        }
    }
}
