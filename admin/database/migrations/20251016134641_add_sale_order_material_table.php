<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddSaleOrderMaterialTable extends Migrator
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
        // 创建销售订单原料表，包含以上字段
        if (!$this->hasTable('sale_order_material')) {
            $table = $this->table('sale_order_material',  ['comment' => '销售订单原料表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '销售订单原料ID'])
                ->addColumn('sale_order_uuid', 'biginteger', ['default' => 0, 'comment' => '销售订单ID'])
                ->addColumn('sale_bill_uuid', 'biginteger', ['default' => 0, 'comment' => '销售账单ID'])
                ->addColumn('material_uuid', 'biginteger', ['default' => 0, 'comment' => '原料ID'])
                ->addColumn('num', 'decimal', ['precision' => 22, 'scale' => 4, 'null' => false, 'default' => 0.0000, 'comment' => '数量,原料的实际使用数量'])
                ->addColumn('staff_shift_log_uuid', 'biginteger', ['default' => 0, 'comment' => '员工交班记录ID'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间(时间戳)'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->create();
        }
    }
}
