<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTableReturnOrderAmount extends Migrator
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
        $table = $this->table('return_order_amount');
        $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '退货金额唯一标识符'])
            ->addColumn('return_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '关联退货单ID'])
            ->addColumn('payment_method_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '关联支付方式ID'])
            ->addColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => '0.00', 'comment' => '退款金额'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->create();
    }
}
