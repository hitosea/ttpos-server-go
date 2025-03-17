<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseFormItem extends Migrator
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
        $table = $this->table('warehouse_form_item', ['id' => false, 'primary_key' => 'id']);
        if (!$table->exists()) {
            $table->addColumn('id', 'integer', ['limit' => 11, 'signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '入库单明细uuid'])
                ->addColumn('num', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 0, 'comment' => '入库数量'])
                ->addColumn('scene', 'integer', ['limit' => 2, 'default' => 0, 'comment' => '场景,0-采购 1-添加入库 2-调整入库 3-退菜入库,这个场景不显示在入库记录页面'])
                ->addColumn('add_stock', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存'])
                ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '材料uuid'])
                ->addColumn('product_bom_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '商品BOM表uuid'])
                ->addColumn('warehouse_form_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '入库单uuid'])
                ->addColumn('sale_order_product_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售订单商品uuid,用于退菜入库'])
                ->addColumn('create_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->setComment('入库单明细表')
                ->create();
        }
    }
}
