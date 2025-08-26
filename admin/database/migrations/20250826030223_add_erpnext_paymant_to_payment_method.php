<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpnextPaymantToPaymentMethod extends Migrator
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
        $table = $this->table('payment_method');
        if (!$table->hasColumn('erpnext_payment')) {
            $table->addColumn('erpnext_payment', 'string', ['limit' => 255, 'default' => '', 'comment' => 'ERPNext支付方式，仅用于云平台给商家添加ERPNext支付方式过滤', 'after' => 'default_img']);
        }
        $table->update();
    }
}
