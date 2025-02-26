<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnShowMustPlanAutoAddMustProductToSaleBillTable extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('show_must_plan')) {
            $table->addColumn(Column::tinyInteger('show_must_plan')->setDefault(1)->setComment('是否显示必点方案, 0-不显示 1-显示.点击确认必点商品按钮后改值为0'));
            $table->update();
        }

        if (!$table->hasColumn('auto_add_must_product')) {
            $table->addColumn(Column::tinyInteger('auto_add_must_product')->setDefault(1)->setComment('是否自动加购必点商品, 0-不自动加购 1-自动加购'));
            $table->update();
        }
    }
}
