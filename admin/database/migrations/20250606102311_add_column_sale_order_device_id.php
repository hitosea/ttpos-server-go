<?php

use think\migration\Migrator;

class AddColumnSaleOrderDeviceId extends Migrator
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
        $table = $this->table('sale_order');
        if (!$table->hasColumn('device_id')) {
            $table->addColumn('device_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '设备ID,用于标识订单来源设备.来源h5时，device_id为h5', 'after' => 'cashier_name']);
        }
        $table->update();

        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('device_id')) {
            $table->addColumn('device_id', 'string', ['limit' => 255, 'default' => '', 'comment' => '设备ID,用于标识订单来源设备.来源h5时，device_id为h5', 'after' => 'is_buffet']);
        }
        $table->update();
    }
}

