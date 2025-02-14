<?php

use think\migration\Migrator;


class CreateTable extends Migrator
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

        if (!$this->hasTable('h5_order')) {
            $table = $this->table('h5_order', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '扫码订单表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '扫码订单ID']);
            $table->addColumn('desk_uuid', 'biginteger', ['default' => 0, 'comment' => '桌台uuid']);
            $table->addColumn('status', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '状态, 0-未下单 1-未接单 2-已接单 3-已拒单']);
            $table->addColumn('create_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)']);
            $table->addColumn('update_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)']);
            $table->addColumn('delete_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)']);
            $table->create();
        }

        if (!$this->hasTable('h5_order_product')) {
            $table = $this->table('h5_order_product', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '扫码订单商品']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '扫码订单商品ID']);
            $table->addColumn('num', 'integer', ['default' => 0, 'comment' => '商品数量']);
            $table->addColumn('h5_order_uuid', 'biginteger', ['default' => 0, 'comment' => '扫码订单uuid']);
            $table->addColumn('create_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '创建时间(时间戳)']);
            $table->addColumn('update_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '更新时间(时间戳)']);
            $table->addColumn('delete_time', 'integer', ['limit' => 10, 'default' => 0, 'comment' => '删除时间(时间戳)']);
            $table->create();
        }
    }
}
