<?php

use think\migration\Migrator;

class AddHeadquarterUuidToFileTables extends Migrator
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
        // 检查 ttpos_file 表是否存在 headquarter_uuid 字段
        if (!$this->table('file')->hasColumn('headquarter_uuid')) {
            $this->table('file')
                ->addColumn('headquarter_uuid', 'biginteger', ['default' => 0, 'comment' => '总部UUID', 'after' => 'group_uuid'])
                ->update();
        }

        // 检查 ttpos_file_group 表是否存在 headquarter_uuid 字段
        if (!$this->table('file_group')->hasColumn('headquarter_uuid')) {
            $this->table('file_group')
                ->addColumn('headquarter_uuid', 'biginteger', ['default' => 0, 'comment' => '总部UUID', 'after' => 'sort'])
                ->update();
        }
    }
}
