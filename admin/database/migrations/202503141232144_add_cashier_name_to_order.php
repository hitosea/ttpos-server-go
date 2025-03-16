<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCashierNameToOrder extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('cashier_name')) {
                $table->addColumn('cashier_name', 'string', [
                    'null' => false,
                    'limit' => 255,
                    'default' => '',
                    'comment' => '收银员名称',
                    'after' => 'cashier_uuid',
                ]);
        }
        $table->update();

        // sale_order
        $table = $this->table('sale_order');
        if (!$table->hasColumn('cashier_name')) {
                $table->addColumn('cashier_name', 'string', [
                    'null' => false,
                    'limit' => 255,
                    'default' => '',
                    'comment' => '收银员名称',
                    'after' => 'cashier_uuid',
                ]);
        }
        $table->update();
    }
}
