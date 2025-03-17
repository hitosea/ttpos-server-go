<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTimeToWarehouseOutFormItem extends Migrator
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
        if (!$table->hasColumn('revoke_time')) {
            $table->addColumn('revoke_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '撤销时间(时间戳)', 'after' => 'reduce_stock'])->update();
        }

        $table = $this->table('warehouse_form_item');
        if (!$table->hasColumn('scene')) {
            $table->updateColumn('scene', 'tinyinteger', ['signed' => false, 'default' => 0, 'comment' => '场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面', 'after' => 'num'])->update();
        }
        $table->update();
    }
}
