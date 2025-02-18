<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMaterialTypeToPurchaseFormItemTable extends Migrator
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
        // 采购单明细表
        $table = $this->table('purchase_form_item');
        if (!$table->hasColumn('material_type')) {
            $table->addColumn(Column::tinyInteger('material_type')->setDefault(0)->setComment('物料类型,0-商品 1-原料')->setAfter('purchase_form_uuid'));
        }
        $table->update();
    }
}
