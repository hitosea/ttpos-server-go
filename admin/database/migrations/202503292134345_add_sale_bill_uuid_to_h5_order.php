<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSaleBillUuidToH5Order extends Migrator
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
        $table = $this->table('h5_order');
        if (!$table->hasColumn('sale_bill_uuid')) {
            $table->addColumn('sale_bill_uuid', 'biginteger', [
                'default' => 0,
                'comment' => '销售账单uuid',
                'null' => false,
                'after' => 'sale_order_uuid',
            ]);
        }
        $table->update();
    }
}
