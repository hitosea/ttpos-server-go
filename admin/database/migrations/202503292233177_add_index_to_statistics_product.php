<?php

use think\facade\Db;
use think\facade\Config;
use think\migration\Migrator;

class AddIndexToStatisticsProduct extends Migrator
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
        $table = $this->table('statistics_product');
        if (!$table->hasIndex('idx_refund_time')) {
            $table->addIndex(['refund_time'], ['name' => 'idx_refund_time']);
        }
        if (!$table->hasIndex('idx_sale_bill_uuid')) {
            $table->addIndex(['sale_bill_uuid'], ['name' => 'idx_sale_bill_uuid']);
        }
        $table->update();
       
    }
}
