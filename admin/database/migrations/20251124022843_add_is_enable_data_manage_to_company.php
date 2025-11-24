<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsEnableDataManageToCompany extends Migrator
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
        $table = $this->table('company');

        // 新增是否启用erp字段（若不存在）
        if (!$table->hasColumn('is_enable_data_manage')) {
            $table->addColumn('is_enable_data_manage', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否启用数据管理: 0不启用, 1启用', 'after' => 'is_enable_erp']);
        }

        $table->update();
    }
}
