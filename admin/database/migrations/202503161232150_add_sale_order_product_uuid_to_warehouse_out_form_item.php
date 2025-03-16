<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSaleOrderProductUuidToWarehouseOutFormItem extends Migrator
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
        $table = $this->table('warehouse_out_form_item');
        if (!$table->hasColumn('sale_order_product_uuid')) {
                $table->addColumn('sale_order_product_uuid', 'biginteger', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
                    'after' => 'uuid',
                ]);
        }
        if (!$table->hasColumn('sale_order_uuid')) {
                $table->addColumn('sale_order_uuid', 'biginteger', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
                    'after' => 'uuid',
                ]);
        }
        if (!$table->hasColumn('sale_bill_uuid')) {
                $table->addColumn('sale_bill_uuid', 'biginteger', [
                    'null' => false,
                    'default' => 0,
                    'comment' => '销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录',
                    'after' => 'uuid',
                ]);
        }
        $table->update();
    }
}
