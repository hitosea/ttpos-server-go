<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnServiceFeeSaleOrderBuffetCustomerType extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if (!$table->hasColumn('service_fee')) {
            $table->addColumn('service_fee', 'decimal', ['precision' => 10, 'scale' => 2, 'default' => 0, 'comment' => '服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例', 'after' => 'tax_fee'])->update();
        }
    }
}
