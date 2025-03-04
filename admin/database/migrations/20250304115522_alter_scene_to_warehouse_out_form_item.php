<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterSceneToWarehouseOutFormItem extends Migrator
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
        if ($table->hasColumn('scene')) {
            $table->changeColumn('scene', 'tinyinteger', ['limit' => 2, 'null' => false, 'default' => 1, 'comment' => '场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除', 'after' => 'num']);
        }
        $table->update();
    }
}
