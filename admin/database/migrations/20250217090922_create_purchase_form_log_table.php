<?php

use think\migration\Migrator;


class CreatePurchaseFormLogTable extends Migrator
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
        if (!$this->hasTable('purchase_form_log')) {
            $table = $this->table('purchase_form_log', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '采购单日志表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购单日志UUID']);
            $table->addColumn('purchase_form_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '采购单ID']);
            $table->addColumn('operator_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '操作人ID']);
            $table->addColumn('username', 'string', ['limit' => 255, 'default' => '', 'comment' => '操作人员']);
            $table->addColumn('status', 'integer', ['default' => 0, 'comment' => '操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库']);
            $table->addColumn('operation', 'string', ['limit' => 255, 'default' => '', 'comment' => '操作动作']);
            $table->addColumn('remark', 'string', ['limit' => 2000, 'default' => '', 'comment' => '备注']);
            $table->addColumn('create_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)']);
            $table->addColumn('update_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)']);
            $table->addColumn('delete_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)']);
            $table->addIndex('uuid', ['unique' => true]);
            $table->create();
        }
    }
}
