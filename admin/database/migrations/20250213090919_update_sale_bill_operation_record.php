<?php

use think\migration\Migrator;


class UpdateSaleBillOperationRecord extends Migrator
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
        $table = $this->table('sale_bill_operation_record');
        if ($table->hasColumn('message')) {
            $table->renameColumn('message', 'data');
            $table->update();
        }
        // 修改数据结构类型
        $table->changeColumn('data', 'text', ['null' => true, 'default' => null])
              ->update();
        
    }
}