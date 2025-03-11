<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddUuidReturnOrderAmount extends Migrator
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
        if (!$table->hasColumn('payment_order_uuid')) {
            $table->addColumn('payment_order_uuid', 'biginteger', ['default' => 0, 'comment' => '关联支付单ID,用于判断支付单的钱还有多少未退', 'after' => 'payment_method_uuid']);
            $table->update();
        }
    }
}
	