<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddLocationToMemberSaleOrder extends Migrator
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
        if (!$table->hasColumn('location')) {
            $table->addColumn('location', 'string', ['length' => 255, 'null' => false, 'default' => '', 'comment' => '骑手位置,格式:纬度,经度', 'after' => 'rider_phone']);
            $table->update();
        }
        if (!$table->hasColumn('sale_bill_uuid')) {
            $table->addColumn('sale_bill_uuid', 'biginteger', [ 'signed' => false, 'null' => false, 'default' => 0,'comment' => '销售账单ID', 'after' => 'uuid']);
            $table->update();
        } 
        if (!$table->hasColumn('sale_order_uuid')) {
            $table->addColumn('sale_order_uuid', 'biginteger', [ 'signed' => false, 'null' => false, 'default' => 0,'comment' => '销售订单ID', 'after' => 'sale_bill_uuid']);
            $table->update();
        }
    
        if ($table->hasColumn('rider_latitude')) {
            $table->removeColumn('rider_latitude');     
            $table->update();
        }
        if ($table->hasColumn('rider_longitude')) {
            $table->removeColumn('rider_longitude'); 
            $table->update();
        }
       
    }
}
