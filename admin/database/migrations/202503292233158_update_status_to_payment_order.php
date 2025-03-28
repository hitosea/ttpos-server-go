<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateStatusToPaymentOrder extends Migrator
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
        $table = $this->table('payment_order');
        if ($table->hasColumn('status')) {
            $table->changeColumn('status', 'integer', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常']);
        }
        if (!$table->hasColumn('status_reason')) {
            $table->addColumn('status_reason', 'text', ['null' => false, 'default' => '', 'comment' => '支付状态原因', 'after' => 'status']);
        }
        $table->update();
    }
}
