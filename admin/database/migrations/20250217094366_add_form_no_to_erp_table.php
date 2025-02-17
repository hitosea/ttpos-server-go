<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFormNoToErpTable extends Migrator
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
        // 库存交易表
        $table = $this->table('warehouse_form');
        if (!$table->hasColumn('form_no')) {
            $table->addColumn(Column::string('form_no', 255)->setDefault('')->setComment('编号')->setAfter('uuid'));
        }
        $table->update();

        // 采购单表
        $table = $this->table('purchase_form');
        if (!$table->hasColumn('form_no')) {
            $table->addColumn(Column::string('form_no', 255)->setDefault('')->setComment('编号')->setAfter('uuid'));
        }
        if (!$table->hasColumn('num')) {
            $table->addColumn(Column::integer('num')->setDefault(0)->setComment('总数量')->setAfter('amount'));
        }
        if (!$table->hasColumn('status')) {
            $table->addColumn(Column::tinyInteger('status')->setDefault(0)->setComment('状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库')->setAfter('arrival_time'));
        }
        $table->update();

        // 采购单明细表
        $table = $this->table('purchase_form_item');
        if (!$table->hasColumn('estimate_num')) {
            $table->addColumn(Column::integer('estimate_num')->setDefault(0)->setComment('预计数量')->setAfter('material_uuid'));
        }
        if (!$table->hasColumn('estimate_price')) {
            $table->addColumn(Column::decimal('estimate_price', 12, 2)->setDefault(0)->setComment('预计单价')->setAfter('material_uuid'));
        }
        if (!$table->hasColumn('estimate_amount')) {
            $table->addColumn(Column::decimal('estimate_amount', 12, 2)->setDefault(0)->setComment('预计金额')->setAfter('material_uuid'));
        }
        $table->update();

        // 出库单表
        $table = $this->table('warehouse_out_form');
        if (!$table->hasColumn('form_no')) {
            $table->addColumn(Column::string('form_no', 255)->setDefault('')->setComment('编号')->setAfter('uuid'));
        }
        $table->update();

        // 报损单表
        $table = $this->table('loss_report_form');
        if (!$table->hasColumn('form_no')) {
            $table->addColumn(Column::string('form_no', 255)->setDefault('')->setComment('编号')->setAfter('uuid'));
        }
        $table->update();
    }
}
