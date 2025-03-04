<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnRelatedOrderTypeToReturnOrder extends Migrator
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
        $table = $this->table('return_order');
        if (!$table->hasColumn('related_order_type')) {
            $table->addColumn('related_order_type', 'integer', ['signed' => false, 'default' => 0, 'comment' => '关联订单类型：0-销售订单；1-充值订单', 'after' => 'uuid']);
        }
        if ($table->hasColumn('sale_order_no')) {
            $table->renameColumn('sale_order_no', 'related_order_uuid');
        }
        if ($table->hasColumn('sale_order_uuid')) {
            $table->renameColumn('sale_order_uuid', 'related_order_no');
        }
        if (!$table->hasColumn('is_reverse_settlement')) {
            $table->addColumn('is_reverse_settlement', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否反结账：0-否；1-是', 'after' => 'related_order_no']);
        }
        $table->update();
    }
}
