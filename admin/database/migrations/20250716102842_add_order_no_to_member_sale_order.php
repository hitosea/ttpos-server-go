<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOrderNoToMemberSaleOrder extends Migrator
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
        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('order_no')) {
            $table->addColumn('order_no', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '订单号', 'after' => 'serial_number']);
            $table->update();
        }
        if (!$table->hasColumn('refund_amount')) {
            $table->addColumn('refund_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '退款金额', 'after' => 'amount']);
            $table->update();
        }
    }
}
